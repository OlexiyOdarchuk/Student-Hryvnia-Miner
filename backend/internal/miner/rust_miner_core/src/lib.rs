//! Ядро перебору nonce для SHA-256 PoW.
//!
//! Повідомлення блоку має вигляд `prefix || nonce || suffix`, де змінюється
//! лише `nonce`. Замість того щоб щоразу проганяти все повідомлення через
//! високорівневий `Digest` API, ми:
//!
//! 1. один раз стискаємо всі повні 64-байтові блоки `prefix` у midstate;
//! 2. один раз збираємо «хвіст» (залишок prefix + nonce + suffix + padding)
//!    із уже проставленою довжиною повідомлення;
//! 3. у гарячому циклі лише перезаписуємо цифри nonce і викликаємо
//!    `compress256` напряму — без буферизації, без padding-логіки,
//!    без алокацій і без серіалізації результату.
//!
//! Перебудова хвоста потрібна тільки коли змінюється кількість цифр nonce,
//! тобто раз на десятки тисяч ітерацій.

use sha2::block_api::compress256;

const BLOCK: usize = 64;

/// Початковий стан SHA-256 (FIPS 180-4).
const SHA256_IV: [u32; 8] = [
    0x6a09_e667,
    0xbb67_ae85,
    0x3c6e_f372,
    0xa54f_f53a,
    0x510e_527f,
    0x9b05_688c,
    0x1f83_d9ab,
    0x5be0_cd19,
];

/// Версія ABI. Go-сторона звіряє її, щоб не злінкуватися зі старим `.a`.
pub const CORE_ABI_VERSION: u32 = 2;

#[no_mangle]
pub extern "C" fn shminer_core_abi_version() -> u32 {
    CORE_ABI_VERSION
}

#[inline(always)]
fn as_blocks(buf: &[u8]) -> &[[u8; BLOCK]] {
    debug_assert_eq!(buf.len() % BLOCK, 0);
    // SAFETY: `[u8; BLOCK]` має розмір BLOCK і вирівнювання 1, а довжина buf
    // кратна BLOCK, тож розбиття на масиви коректне.
    unsafe { core::slice::from_raw_parts(buf.as_ptr().cast::<[u8; BLOCK]>(), buf.len() / BLOCK) }
}

/// Чи має хеш `bits` нульових старших бітів.
#[inline(always)]
fn meets_difficulty(state: &[u32; 8], bits: u32) -> bool {
    if bits == 0 {
        return true;
    }
    if bits <= 64 {
        let head = ((state[0] as u64) << 32) | (state[1] as u64);
        return head >> (64 - bits) == 0;
    }
    // Складність >64 біт на практиці не трапляється, тож тут простий шлях.
    let mut left = bits;
    for word in state.iter() {
        if left >= 32 {
            if *word != 0 {
                return false;
            }
            left -= 32;
        } else {
            return word >> (32 - left) == 0;
        }
        if left == 0 {
            return true;
        }
    }
    true
}

#[inline(always)]
fn digit_count(n: i64) -> usize {
    let mut digits = 1;
    let mut rest = n / 10;
    while rest > 0 {
        digits += 1;
        rest /= 10;
    }
    digits
}

/// Готовий до стиснення хвіст повідомлення разом із padding.
struct Tail {
    buf: Vec<u8>,
    nonce_at: usize,
    digits: usize,
}

impl Tail {
    fn build(head: &[u8], suffix: &[u8], prefix_len: usize, nonce: i64) -> Self {
        let digits = digit_count(nonce);
        let msg_len = head.len() + digits + suffix.len();
        let padded = (msg_len + 9).div_ceil(BLOCK) * BLOCK;

        let mut buf = vec![0u8; padded];
        buf[..head.len()].copy_from_slice(head);
        let suffix_at = head.len() + digits;
        buf[suffix_at..suffix_at + suffix.len()].copy_from_slice(suffix);
        buf[msg_len] = 0x80;

        let bit_len = ((prefix_len + digits + suffix.len()) as u64).wrapping_mul(8);
        buf[padded - 8..].copy_from_slice(&bit_len.to_be_bytes());

        let mut tail = Self {
            buf,
            nonce_at: head.len(),
            digits,
        };
        tail.write_nonce(nonce);
        tail
    }

    /// Записує nonce у поле цифр (право-вирівняно, без ведучих нулів,
    /// бо `digits == digit_count(nonce)` за побудовою).
    fn write_nonce(&mut self, nonce: i64) {
        let mut rest = nonce;
        for slot in self.buf[self.nonce_at..self.nonce_at + self.digits]
            .iter_mut()
            .rev()
        {
            *slot = b'0' + (rest % 10) as u8;
            rest /= 10;
        }
    }

    /// Додає `step` просто в десятковому записі — це дешевше за повторну
    /// конвертацію числа (ланцюжок ділень) на кожній ітерації.
    ///
    /// Повертає `false`, якщо результат більше не вміщається в поточну
    /// кількість цифр: тоді хвіст треба перебудувати.
    #[inline(always)]
    fn bump_nonce(&mut self, step: i64) -> bool {
        let mut carry = step;
        let mut at = self.nonce_at + self.digits;
        while carry != 0 {
            if at == self.nonce_at {
                return false;
            }
            at -= 1;
            let sum = match i64::from(self.buf[at] - b'0').checked_add(carry) {
                Some(sum) => sum,
                None => return false,
            };
            self.buf[at] = b'0' + (sum % 10) as u8;
            carry = sum / 10;
        }
        true
    }
}

/// Перебирає `chunk_size` значень nonce з кроком `step`.
///
/// Повертає знайдений nonce або `-1`, якщо в цьому чанку розв'язку немає.
///
/// # Safety
///
/// `prefix`/`suffix` мають вказувати на щонайменше `prefix_len`/`suffix_len`
/// читабельних байтів і залишатися валідними на час виклику. Нульова довжина
/// допускає нульовий вказівник.
#[no_mangle]
pub unsafe extern "C" fn shminer_mine_chunk(
    prefix: *const u8,
    prefix_len: usize,
    suffix: *const u8,
    suffix_len: usize,
    diff_bits: u32,
    start_nonce: i64,
    step: i64,
    chunk_size: i64,
) -> i64 {
    if start_nonce < 0 || step <= 0 || chunk_size <= 0 {
        return -1;
    }
    if (prefix.is_null() && prefix_len != 0) || (suffix.is_null() && suffix_len != 0) {
        return -1;
    }

    // SAFETY: контракт функції гарантує валідність вказівників і довжин.
    let prefix = unsafe { slice_from_raw(prefix, prefix_len) };
    let suffix = unsafe { slice_from_raw(suffix, suffix_len) };

    // Крок 1: усі повні блоки prefix стискаємо один раз.
    let whole = prefix_len - prefix_len % BLOCK;
    let mut midstate = SHA256_IV;
    if whole > 0 {
        compress256(&mut midstate, as_blocks(&prefix[..whole]));
    }
    let head = &prefix[whole..];

    let mut nonce = start_nonce;
    let mut tail = Tail::build(head, suffix, prefix_len, nonce);
    let mut remaining = chunk_size;

    loop {
        let mut state = midstate;
        compress256(&mut state, as_blocks(&tail.buf));

        if meets_difficulty(&state, diff_bits) {
            return nonce;
        }

        remaining -= 1;
        if remaining == 0 {
            return -1;
        }

        nonce = match nonce.checked_add(step) {
            Some(next) => next,
            None => return -1,
        };

        if !tail.bump_nonce(step) {
            tail = Tail::build(head, suffix, prefix_len, nonce);
        }
    }
}

#[inline(always)]
unsafe fn slice_from_raw<'a>(ptr: *const u8, len: usize) -> &'a [u8] {
    if len == 0 {
        &[]
    } else {
        // SAFETY: перевірено вище, що ptr не нульовий при len != 0.
        unsafe { core::slice::from_raw_parts(ptr, len) }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use sha2::{Digest, Sha256};

    fn reference(prefix: &[u8], nonce: i64, suffix: &[u8]) -> [u8; 32] {
        let mut hasher = Sha256::new();
        hasher.update(prefix);
        hasher.update(nonce.to_string().as_bytes());
        hasher.update(suffix);
        hasher.finalize().into()
    }

    fn leading_zero_bits(hash: &[u8; 32]) -> u32 {
        let mut bits = 0;
        for byte in hash {
            bits += byte.leading_zeros();
            if *byte != 0 {
                break;
            }
        }
        bits
    }

    fn mine(prefix: &[u8], suffix: &[u8], bits: u32, start: i64, step: i64, n: i64) -> i64 {
        unsafe {
            shminer_mine_chunk(
                prefix.as_ptr(),
                prefix.len(),
                suffix.as_ptr(),
                suffix.len(),
                bits,
                start,
                step,
                n,
            )
        }
    }

    /// Ядро має знаходити рівно те, що знайшов би наївний перебір.
    #[test]
    fn matches_reference_across_prefix_lengths() {
        let suffix = b"04b22cebe3c00859251";
        // 0, 63, 64, 65, 128 байтів prefix — усі варіанти вирівнювання.
        for prefix_len in [0usize, 1, 63, 64, 65, 128] {
            let prefix: Vec<u8> = (0..prefix_len).map(|i| b'a' + (i % 26) as u8).collect();
            for bits in [0u32, 1, 8, 12] {
                let expected =
                    (0..5000).find(|n| leading_zero_bits(&reference(&prefix, *n, suffix)) >= bits);
                let got = mine(&prefix, suffix, bits, 0, 1, 5000);
                assert_eq!(
                    got,
                    expected.unwrap_or(-1),
                    "prefix_len={prefix_len} bits={bits}"
                );
            }
        }
    }

    /// Перехід кількості цифр nonce (999 -> 1000) не має ламати layout.
    #[test]
    fn survives_digit_rollover() {
        let prefix = vec![b'0'; 64];
        let suffix = b"wallet1";
        for start in [7i64, 95, 990, 9_995, 99_990] {
            for offset in 0..12i64 {
                let nonce = start + offset;
                let expect = leading_zero_bits(&reference(&prefix, nonce, suffix));
                let got = mine(&prefix, suffix, expect, nonce, 1, 1);
                assert_eq!(got, nonce, "nonce={nonce}");
            }
        }
    }

    /// Крок > 1 має давати ту саму розкладку по потоках, що й суцільний перебір.
    #[test]
    fn stepping_partitions_search_space() {
        let prefix = vec![b'f'; 64];
        let suffix = b"s";
        let bits = 12;
        let single = mine(&prefix, suffix, bits, 0, 1, 20_000);
        assert!(single >= 0);

        let step = 4;
        let best = (0..step)
            .map(|worker| mine(&prefix, suffix, bits, worker, step, 20_000 / step))
            .filter(|n| *n >= 0)
            .min()
            .unwrap();
        assert_eq!(best, single);
    }

    /// Інкрементний лічильник має давати той самий десятковий запис, що й
    /// звичайна конвертація — на довгому пробігу і на різних кроках.
    #[test]
    fn incremental_counter_matches_full_scan() {
        let prefix = vec![b'c'; 64];
        let suffix = b"wallet-suffix-1";
        for step in [1i64, 2, 7, 16, 97, 1000] {
            for start in [0i64, 9, 99, 12_345] {
                // Складність, якої точно не буде — змушує пройти весь чанк
                // і тим самим прогнати всі перебудови розрядності.
                assert_eq!(mine(&prefix, suffix, 250, start, step, 3000), -1);

                // А тепер перевіряємо, що знайдений nonce збігається з еталоном.
                let expected = (0..3000)
                    .map(|i| start + i * step)
                    .find(|n| leading_zero_bits(&reference(&prefix, *n, suffix)) >= 10);
                let got = mine(&prefix, suffix, 10, start, step, 3000);
                assert_eq!(got, expected.unwrap_or(-1), "step={step} start={start}");
            }
        }
    }

    /// Величезний крок не має переповнювати лічильник — має спрацювати
    /// відкат на повну перебудову хвоста.
    #[test]
    fn handles_extreme_step() {
        let prefix = vec![b'0'; 64];
        let suffix = b"s";
        for step in [i64::MAX, i64::MAX / 3, 1 << 60] {
            // Другий nonce або переповниться (-1), або буде порахований чесно.
            let got = mine(&prefix, suffix, 250, 1, step, 4);
            assert_eq!(got, -1);
        }
        // Крок, що перевищує i64 при додаванні до нонсу -> checked_add -> -1
        assert_eq!(mine(&prefix, suffix, 0, i64::MAX, 1, 4), i64::MAX);
    }

    #[test]
    fn rejects_bad_arguments() {
        let prefix = vec![b'0'; 64];
        assert_eq!(mine(&prefix, b"s", 8, -1, 1, 10), -1);
        assert_eq!(mine(&prefix, b"s", 8, 0, 0, 10), -1);
        assert_eq!(mine(&prefix, b"s", 8, 0, 1, 0), -1);
    }

    #[test]
    fn difficulty_boundaries() {
        let mut state = [0u32; 8];
        assert!(meets_difficulty(&state, 0));
        assert!(meets_difficulty(&state, 64));
        assert!(meets_difficulty(&state, 200));
        state[1] = 1;
        assert!(meets_difficulty(&state, 63));
        assert!(!meets_difficulty(&state, 64));
        state[0] = 0x0000_0001;
        assert!(meets_difficulty(&state, 31));
        assert!(!meets_difficulty(&state, 32));
    }
}
