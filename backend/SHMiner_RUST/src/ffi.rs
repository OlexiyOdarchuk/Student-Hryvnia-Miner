use std::{
    collections::HashMap,
    ffi::c_char,
    ptr,
    sync::{Arc, Mutex},
};

use once_cell::sync::Lazy;

use crate::mining::{SessionBuilder, SessionHandle, SessionState};

#[repr(C)]
pub struct ShMinerWorkRequest {
    pub prev_hash: *const c_char,
    pub prev_hash_len: usize,
    pub wallet: *const c_char,
    pub wallet_len: usize,
    pub difficulty_bits: u32,
    pub threads: u32,
    pub timestamp: i64,
}

#[repr(C)]
pub struct ShMinerResult {
    pub nonce: u64,
    pub hash: [u8; 32],
    pub hash_count: u64,
}

struct SessionManager {
    next_id: u32,
    sessions: HashMap<u32, SessionHandle>,
}

impl SessionManager {
    fn new() -> Self {
        Self {
            next_id: 1,
            sessions: HashMap::new(),
        }
    }

    fn insert(&mut self, handle: SessionHandle) -> u32 {
        let id = self.next_id;
        self.sessions.insert(id, handle);
        self.next_id = if self.next_id == u32::MAX {
            1
        } else {
            self.next_id + 1
        };
        id
    }

    fn get_state(&self, session_id: u32) -> Option<Arc<SessionState>> {
        self.sessions.get(&session_id).map(|handle| handle.state())
    }

    fn remove(&mut self, session_id: u32) -> Option<SessionHandle> {
        self.sessions.remove(&session_id)
    }
}

static SESSION_MANAGER: Lazy<Mutex<SessionManager>> =
    Lazy::new(|| Mutex::new(SessionManager::new()));

fn ptr_to_vec(ptr: *const c_char, len: usize) -> Option<Vec<u8>> {
    if ptr.is_null() || len == 0 {
        return None;
    }
    let bytes = ptr::slice_from_raw_parts(ptr as *const u8, len);
    Some(unsafe { (*bytes).to_vec() })
}

#[unsafe(no_mangle)]
pub extern "C" fn create_session(work: *const ShMinerWorkRequest) -> u32 {
    let work = match unsafe { work.as_ref() } {
        Some(w) => w,
        None => return 0,
    };

    let prev_hash = match ptr_to_vec(work.prev_hash, work.prev_hash_len) {
        Some(v) => v,
        None => return 0,
    };
    let wallet = match ptr_to_vec(work.wallet, work.wallet_len) {
        Some(v) => v,
        None => return 0,
    };

    let handle = SessionBuilder::new(
        prev_hash,
        wallet,
        work.timestamp,
        work.difficulty_bits,
        work.threads as usize,
    )
    .spawn();

    let mut manager = SESSION_MANAGER.lock().unwrap();
    manager.insert(handle)
}

#[unsafe(no_mangle)]
pub extern "C" fn step_session(session_id: u32, out: *mut ShMinerResult) -> bool {
    if out.is_null() {
        return false;
    }

    let state = {
        let manager = SESSION_MANAGER.lock().unwrap();
        manager.get_state(session_id)
    };

    let state = match state {
        Some(s) => s,
        None => return false,
    };

    if let Some(result) = state.wait_for_result() {
        unsafe {
            (*out).nonce = result.nonce;
            (*out).hash = result.hash;
            (*out).hash_count = result.hash_count;
        }
        return true;
    }

    let snapshot = state.snapshot_hash_count();
    unsafe {
        (*out).nonce = 0;
        (*out).hash = [0u8; 32];
        (*out).hash_count = snapshot;
    }
    false
}

#[unsafe(no_mangle)]
pub extern "C" fn stop_session(session_id: u32) {
    let handle = {
        let mut manager = SESSION_MANAGER.lock().unwrap();
        manager.remove(session_id)
    };

    if let Some(session) = handle {
        session.stop();
    }
}
