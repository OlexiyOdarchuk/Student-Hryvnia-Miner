#ifndef SHMINER_H
#define SHMINER_H

#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    const char* prev_hash;
    size_t prev_hash_len;
    const char* wallet;
    size_t wallet_len;
    uint32_t difficulty_bits;
    uint32_t threads;
    int64_t timestamp;
} ShMinerWorkRequest;

typedef struct {
    uint64_t nonce;
    uint8_t hash[32];
    uint64_t hash_count;
} ShMinerResult;

uint32_t create_session(const ShMinerWorkRequest* work);
bool step_session(uint32_t session_id, ShMinerResult* result);
void stop_session(uint32_t session_id);

#ifdef __cplusplus
}
#endif

#endif // SHMINER_H
