use std::ptr;
use std::sync::{
    Arc, Condvar, Mutex,
    atomic::{AtomicBool, AtomicU64, Ordering},
};
use std::thread::{self, JoinHandle};

use itoa::Buffer as ItoaBuffer;
use sha2::{Digest, Sha256};

const REWARD_PART: &[u8] = b"1";
const HASH_BUFFER_SIZE: usize = 256;
const BATCH_SIZE: u64 = 128;
const HASH_FLUSH_THRESHOLD: u64 = 1024;
const MAX_DIFFICULTY_BITS: u32 = 256;

#[derive(Clone, Copy)]
pub struct Difficulty {
    bytes: usize,
    nibble_mask: Option<u8>,
}

impl Difficulty {
    pub fn compile(bits: u32) -> Self {
        let bits = bits.min(MAX_DIFFICULTY_BITS);
        let bytes = (bits / 8) as usize;
        let rem = bits % 8;

        let nibble_mask = if rem == 0 {
            None
        } else {
            Some(((0xFF_u8) << (8 - rem)) as u8)
        };

        Self { bytes, nibble_mask }
    }

    fn matches(&self, hash: &[u8; 32]) -> bool {
        for byte in 0..self.bytes.min(hash.len()) {
            if hash[byte] != 0 {
                return false;
            }
        }

        if let Some(mask) = self.nibble_mask {
            if self.bytes < hash.len() {
                if hash[self.bytes] & mask != 0 {
                    return false;
                }
            }
        }
        true
    }
}

#[derive(Clone, Copy, Debug)]
pub struct MineResult {
    pub nonce: u64,
    pub hash: [u8; 32],
    pub hash_count: u64,
}

struct SessionResources {
    before_nonce: Arc<[u8]>,
    after_nonce: Arc<[u8]>,
    before_len: usize,
    after_len: usize,
}

impl SessionResources {
    fn new(prev_hash: Vec<u8>, wallet: Vec<u8>, timestamp: i64) -> Self {
        let before = prev_hash;
        let before_len = before.len();
        let mut after = wallet;
        after.extend_from_slice(REWARD_PART);
        let mut itoa = ItoaBuffer::new();
        let timestamp_bytes = itoa.format(timestamp);
        after.extend_from_slice(timestamp_bytes.as_bytes());
        let after_len = after.len();

        Self {
            before_nonce: Arc::from(before),
            after_nonce: Arc::from(after),
            before_len,
            after_len,
        }
    }
}

pub(crate) struct SessionState {
    difficulty: Difficulty,
    resources: SessionResources,
    nonce_cursor: AtomicU64,
    hash_count: AtomicU64,
    found: AtomicBool,
    stop: AtomicBool,
    result: Mutex<Option<MineResult>>,
    condvar: Condvar,
}

impl SessionState {
    fn new(resources: SessionResources, difficulty_bits: u32) -> Self {
        Self {
            difficulty: Difficulty::compile(difficulty_bits),
            resources,
            nonce_cursor: AtomicU64::new(0),
            hash_count: AtomicU64::new(0),
            found: AtomicBool::new(false),
            stop: AtomicBool::new(false),
            result: Mutex::new(None),
            condvar: Condvar::new(),
        }
    }

    pub(crate) fn wait_for_result(&self) -> Option<MineResult> {
        let mut guard = self.result.lock().unwrap();
        while guard.is_none() && !self.stop.load(Ordering::Acquire) {
            guard = self.condvar.wait(guard).unwrap();
        }
        guard.take()
    }

    pub(crate) fn snapshot_hash_count(&self) -> u64 {
        self.hash_count.swap(0, Ordering::Relaxed)
    }
}

pub struct SessionHandle {
    state: Arc<SessionState>,
    handles: Vec<JoinHandle<()>>,
}

impl SessionHandle {
    pub(crate) fn state(&self) -> Arc<SessionState> {
        Arc::clone(&self.state)
    }

    pub fn wait_result(&self) -> Option<MineResult> {
        self.state.wait_for_result()
    }

    pub fn stop(self) {
        self.state.stop.store(true, Ordering::Release);
        self.state.condvar.notify_all();
        for handle in self.handles {
            let _ = handle.join();
        }
    }
}

pub struct SessionBuilder {
    resources: SessionResources,
    difficulty_bits: u32,
    threads: usize,
}

impl SessionBuilder {
    pub fn new(
        prev_hash: Vec<u8>,
        wallet: Vec<u8>,
        timestamp: i64,
        difficulty_bits: u32,
        threads: usize,
    ) -> Self {
        let resources = SessionResources::new(prev_hash, wallet, timestamp);
        let thread_count = threads.clamp(1, num_cpus::get());
        let clamped_diff = difficulty_bits.min(MAX_DIFFICULTY_BITS);

        Self {
            resources,
            difficulty_bits: clamped_diff,
            threads: thread_count,
        }
    }

    pub fn spawn(self) -> SessionHandle {
        let state = Arc::new(SessionState::new(self.resources, self.difficulty_bits));
        let mut handles = Vec::with_capacity(self.threads);
        for _ in 0..self.threads {
            let thread_state = Arc::clone(&state);
            handles.push(thread::spawn(move || worker_loop(thread_state)));
        }
        SessionHandle { state, handles }
    }
}

fn worker_loop(state: Arc<SessionState>) {
    let before_ptr = state.resources.before_nonce.as_ptr();
    let after_ptr = state.resources.after_nonce.as_ptr();
    let before_len = state.resources.before_len;
    let after_len = state.resources.after_len;
    let difficulty = state.difficulty;

    let mut buffer = [0u8; HASH_BUFFER_SIZE];
    let mut itoa = ItoaBuffer::new();
    let mut hasher = Sha256::new();
    let mut local_count = 0u64;

    while !state.stop.load(Ordering::Relaxed) && !state.found.load(Ordering::Acquire) {
        let base = state.nonce_cursor.fetch_add(BATCH_SIZE, Ordering::Relaxed);
        for offset in 0..BATCH_SIZE {
            if state.stop.load(Ordering::Relaxed) || state.found.load(Ordering::Acquire) {
                break;
            }

            let nonce = base.wrapping_add(offset);
            let mut pos = 0;

            unsafe {
                ptr::copy_nonoverlapping(before_ptr, buffer.as_mut_ptr(), before_len);
            }
            pos += before_len;
            let nonce_bytes = itoa.format(nonce).as_bytes();
            unsafe {
                ptr::copy_nonoverlapping(
                    nonce_bytes.as_ptr(),
                    buffer.as_mut_ptr().add(pos),
                    nonce_bytes.len(),
                );
            }
            pos += nonce_bytes.len();
            unsafe {
                ptr::copy_nonoverlapping(after_ptr, buffer.as_mut_ptr().add(pos), after_len);
            }
            pos += after_len;

            hasher.update(&buffer[..pos]);
            let hash = hasher.finalize_reset();
            local_count += 1;
            let hash_bytes: [u8; 32] = hash.into();

            if difficulty.matches(&hash_bytes) {
                let total_hashes =
                    state.hash_count.fetch_add(local_count, Ordering::Relaxed) + local_count;
                local_count = 0;
                state.found.store(true, Ordering::Release);
                let mut guard = state.result.lock().unwrap();
                if guard.is_none() {
                    *guard = Some(MineResult {
                        nonce,
                        hash: hash_bytes,
                        hash_count: total_hashes,
                    });
                    state.condvar.notify_all();
                }
                break;
            }

            if local_count >= HASH_FLUSH_THRESHOLD {
                state.hash_count.fetch_add(local_count, Ordering::Relaxed);
                local_count = 0;
            }
        }
    }

    if local_count > 0 {
        state.hash_count.fetch_add(local_count, Ordering::Relaxed);
    }
}
