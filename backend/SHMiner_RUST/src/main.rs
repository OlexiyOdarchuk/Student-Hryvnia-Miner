use std::{
    env, thread,
    time::{Duration, SystemTime, UNIX_EPOCH},
};

use anyhow::{Context, Result};
use log::{error, info, warn};
use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};
use shminer::SessionBuilder;

const DEFAULT_BASE_URL: &str = "https://s-hryvnia-1.onrender.com";
const DEFAULT_DIFFICULTY_BITS: u32 = 20;
const DEFAULT_HTTP_TIMEOUT_SECS: u64 = 5;
const DEFAULT_WALLETS: [&str; 1] = [
    "04b22cebe3c0085925e016647ba96e54282763dbcbcc149db52baa3aaef1b76826edcc3feee1eb0ac26acc09d6bc4f3f956ab91f14d2caca25c3402bee8712ab61",
];

#[derive(Clone)]
struct MinerConfig {
    base_url: String,
    difficulty_bits: u32,
    threads: usize,
    wallets: Vec<String>,
    http_timeout: Duration,
}

impl MinerConfig {
    fn from_env() -> Self {
        let base_url = env::var("SH_MINER_SERVER").unwrap_or_else(|_| DEFAULT_BASE_URL.to_string());

        let difficulty_bits = env::var("SH_MINER_DIFFICULTY")
            .ok()
            .and_then(|v| v.parse::<u32>().ok())
            .filter(|v| *v > 0)
            .unwrap_or(DEFAULT_DIFFICULTY_BITS);

        let threads = env::var("SH_MINER_THREADS")
            .ok()
            .and_then(|v| v.parse::<usize>().ok())
            .filter(|v| *v > 0)
            .unwrap_or_else(num_cpus::get)
            .max(1);

        let wallets = env::var("SH_MINER_WALLETS")
            .ok()
            .map(|list| {
                list.split(',')
                    .filter_map(|w| {
                        let trimmed = w.trim();
                        if trimmed.is_empty() {
                            None
                        } else {
                            Some(trimmed.to_owned())
                        }
                    })
                    .collect()
            })
            .filter(|list: &Vec<String>| !list.is_empty())
            .unwrap_or_else(|| DEFAULT_WALLETS.iter().map(|w| w.to_string()).collect());

        let http_timeout = env::var("SH_MINER_HTTP_TIMEOUT")
            .ok()
            .and_then(|v| v.parse::<u64>().ok())
            .map(Duration::from_secs)
            .unwrap_or(Duration::from_secs(DEFAULT_HTTP_TIMEOUT_SECS));

        Self {
            base_url,
            difficulty_bits,
            threads,
            wallets,
            http_timeout,
        }
    }
}

#[derive(Deserialize)]
struct ChainBlock {
    hash: String,
}

#[derive(Serialize)]
struct Transaction {
    from: String,
    to: String,
    amount: u32,
}

#[derive(Serialize)]
struct BlockSubmission {
    #[serde(rename = "prevHash")]
    prev_hash: String,
    transactions: Vec<Transaction>,
    nonce: u64,
    miner: String,
    reward: u32,
    timestamp: i64,
    hash: String,
}

#[derive(Serialize)]
struct SubmitBlockRequest {
    block: BlockSubmission,
}

fn current_timestamp_ms() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_millis() as i64
}

fn main() -> Result<()> {
    let config = MinerConfig::from_env();

    env_logger::Builder::from_default_env()
        .format_timestamp_secs()
        .init();

    info!(
        "Rust miner using {} threads, {}-bit difficulty, {} wallet(s)",
        config.threads,
        config.difficulty_bits,
        config.wallets.len()
    );

    let client = Client::builder()
        .timeout(config.http_timeout)
        .build()
        .context("failed to build HTTP client")?;

    let mut wallet_index = 0usize;

    loop {
        let prev_hash = match get_chain_head(&client, &config.base_url) {
            Ok(hash) => hash,
            Err(err) => {
                warn!("Cannot fetch latest block: {err}");
                thread::sleep(Duration::from_secs(2));
                continue;
            }
        };

        if config.wallets.is_empty() {
            warn!("Wallet list is empty, waiting for configuration");
            thread::sleep(Duration::from_secs(3));
            continue;
        }

        let wallet = config.wallets[wallet_index % config.wallets.len()].clone();
        wallet_index = wallet_index.wrapping_add(1);

        info!("Mining for wallet {}", wallet);

        let timestamp = current_timestamp_ms();
        let session = SessionBuilder::new(
            prev_hash.as_bytes().to_vec(),
            wallet.as_bytes().to_vec(),
            timestamp,
            config.difficulty_bits,
            config.threads,
        )
        .spawn();

        let outcome = session.wait_result();
        session.stop();

        if let Some(result) = outcome {
            let hash_hex = hex::encode(result.hash);
            info!(
                "✅ Block found {} (nonce {}, {} hashes total)",
                hash_hex, result.nonce, result.hash_count
            );

            if let Err(err) = submit_block(
                &client,
                &config.base_url,
                &prev_hash,
                &wallet,
                result.nonce,
                timestamp,
                &hash_hex,
            ) {
                error!("Block submission failed: {err}");
            } else {
                info!("Block credited for wallet {}", wallet);
            }
        } else {
            warn!("Mining session stopped before a valid hash was produced");
        }
    }
}

fn get_chain_head(client: &Client, base_url: &str) -> Result<String> {
    let url = format!("{}/chain", base_url.trim_end_matches('/'));

    let blocks: Vec<ChainBlock> = client
        .get(url)
        .send()
        .context("chain request failed")?
        .error_for_status()
        .context("chain request returned an error status")?
        .json()
        .context("failed to parse chain response")?;

    blocks
        .last()
        .map(|block| block.hash.clone())
        .context("chain response contains no blocks")
}

fn submit_block(
    client: &Client,
    base_url: &str,
    prev_hash: &str,
    wallet: &str,
    nonce: u64,
    timestamp: i64,
    hash_hex: &str,
) -> Result<()> {
    let url = format!("{}/submit-block", base_url.trim_end_matches('/'));
    let payload = SubmitBlockRequest {
        block: BlockSubmission {
            prev_hash: prev_hash.to_string(),
            transactions: vec![Transaction {
                from: String::new(),
                to: wallet.to_string(),
                amount: 1,
            }],
            nonce,
            miner: wallet.to_string(),
            reward: 1,
            timestamp,
            hash: hash_hex.to_string(),
        },
    };

    client
        .post(url)
        .json(&payload)
        .send()
        .context("submit block request failed")?
        .error_for_status()
        .context("server rejected the submitted block")?;

    Ok(())
}
