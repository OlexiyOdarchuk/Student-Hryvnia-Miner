//! SHA-256 bench-кернел для живого демо на ishawyha.dev, зібраний у WASM.
//!
//! Перебір робить `rust_miner_core::scan` — та сама функція, що лінкується
//! в десктопний SHMiner. Різниця лише в параметрі `STOP_ON_HIT`: майнер
//! зупиняється на першій знахідці, бенчмарк проходить увесь діапазон і
//! рахує влучання.
//!
//! Прообраз дослівно повторює `cmd/wasm-bench` (версію на Go), щоб гілки
//! бенчмарку лишалися порівнюваними:
//!
//!   prevHash + nonce + miner + reward + timestamp,  складність 20 біт
//!
//! JS керує циклом порціями через `bench_batch(n)`: у WASM немає потоків,
//! і довгий цикл всередині модуля заморозив би воркеру event loop разом зі
//! звітами про прогрес і кнопкою «стоп».

use std::cell::RefCell;

use rust_miner_core::scan;

const PREV_HASH: &[u8; 64] = b"0000000000000000000000000000000000000000000000000000000000000000";
const MINER: &[u8] = b"ishawyha_bench_wallet_pubkey_for_demo_only_xxxxxxxxxxxxxxxx";
const REWARD: &[u8] = b"1";

/// 20 біт == 5 нульових hex-символів на початку — рівно те, що перевіряє
/// офіційний MINERJS.js, і дефолт бекенду SHMiner.
const DIFF_BITS: u32 = 20;

#[derive(Default)]
struct State {
    suffix: Vec<u8>,
    nonce: i64,
    hashes: u64,
    hits: u64,
}

thread_local! {
    static STATE: RefCell<State> = RefCell::new(State::default());
}

/// Скидає лічильники й задає стартовий nonce.
///
/// `offset` розводить воркерів пулу по різних ділянках простору nonce, щоб
/// вони не хешували ті самі прообрази. `ts_ms` приходить з JS, бо в
/// `wasm32-unknown-unknown` немає годинника.
#[no_mangle]
pub extern "C" fn bench_reset(offset: f64, ts_ms: f64) {
    let offset = if offset.is_finite() && offset >= 0.0 {
        offset as i64
    } else {
        0
    };
    let ts = if ts_ms.is_finite() && ts_ms >= 0.0 {
        ts_ms as u64
    } else {
        0
    };

    let mut suffix = Vec::with_capacity(MINER.len() + REWARD.len() + 20);
    suffix.extend_from_slice(MINER);
    suffix.extend_from_slice(REWARD);
    suffix.extend_from_slice(ts.to_string().as_bytes());

    STATE.with(|state| {
        *state.borrow_mut() = State {
            suffix,
            nonce: offset,
            hashes: 0,
            hits: 0,
        };
    });
}

/// Проганяє `n` хешів і оновлює лічильники.
#[no_mangle]
pub extern "C" fn bench_batch(n: u32) {
    if n == 0 {
        return;
    }

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let start = state.nonce;
        let result = scan::<false>(PREV_HASH, &state.suffix, DIFF_BITS, start, 1, i64::from(n));

        state.hashes += result.hashes;
        state.hits += result.hits;
        state.nonce = start.saturating_add(i64::from(n));
    });
}

/// Скільки хешів пораховано від останнього `bench_reset`.
#[no_mangle]
pub extern "C" fn bench_count() -> f64 {
    STATE.with(|state| state.borrow().hashes as f64)
}

/// Скільки з них склали блок потрібної складності.
#[no_mangle]
pub extern "C" fn bench_tokens() -> f64 {
    STATE.with(|state| state.borrow().hits as f64)
}
