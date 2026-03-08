# Refactor Plan

## `nodeclient` (без зміни поведінки)
- Перенести HTTP-логіку з `backend/api.go` в `backend/internal/nodeclient`.
- Залишити тимчасові обгортки у старому пакеті, щоб нічого не зламати.
- Критерій: майнінг і баланс працюють як до рефактору.

## `storage` модуль
- Перенести `backend/storage.go` в `backend/internal/storage`.
- Винести інтерфейс `Storage` (`Load/Save/ChangePassword/Exists`).
- Критерій: `InitStorage`, `UnlockStorage`, зміна пароля працюють без змін у UI.

## `wallet` сервіс
- Перенести `backend/wallet.go` в `backend/internal/wallet`.
- Замість глобалів `Wallets/walletDataMap` зробити `WalletService` зі `struct` state + mutex.
- Критерій: add/delete/rename/toggle/import/export без регресій.

## `stats` сервіс
- Перенести `backend/stats.go` в `backend/internal/stats`.
- Створити `StatsService` (hashrate, uptime, dashboard, logs), прибрати глобали.
- Критерій: події `log/stats` у Wails працюють стабільно.

## `mining` сервіс + dependency injection
- Перенести `backend/miner.go` в `backend/internal/mining`.
- Інжектнути залежності інтерфейсами: `NodeClient`, `WalletReader`, `StatsWriter`, `StorageSync`.
- Критерій: `StartMiningLoop` працює без доступу до глобального state.

## `webui` модуль
- Перенести `backend/server_ui.go` в `backend/internal/webui`.
- `webui` читає дані лише через інтерфейси сервісів, не напряму з map/slice.
- Критерій: `/api/stats` і `/ws` віддають ті самі поля, без `private_key`.

## composition root + cleanup
- Додати `backend/internal/app` для збірки всіх сервісів.
- `app.go` лишити thin-адаптером Wails.
- Видалити старі глобали й дублюючі файли в `backend/`.
- Критерій: збірка, запуск, ручний smoke-test усіх основних сценаріїв.

## Мінімум тестів по етапах
1. `nodeclient`: retry/backoff + обробка не-200.
2. `wallet`: унікальність `name/address`, toggle, delete.
3. `storage`: encrypt/decrypt, wrong password.
4. `mining`: difficulty check + stop по context cancel.

## Фінальний чек перед merge
- Повний smoke-test UI сценаріїв.
- Прогон unit-тестів.
- Перевірка, що в публічні DTO немає `private_key`.
- Прибрані тимчасові обгортки/адаптери після міграції.

## Що обов'язково по тестах має бути

### Must-have (писати обов'язково)
1. `backend/internal/storage` (або поточний `storage.go`)
- encrypt/decrypt roundtrip;
- wrong password;
- пошкоджений файл / короткий ciphertext;
- `ChangePassword` зберігає доступ до даних.

2. `backend/internal/wallet` (`wallet.go`)
- унікальність `name` і `address`;
- `Add/Delete/Rename/Toggle`;
- `SetAllMining`;
- `syncStorage` консистентність.

3. `backend/internal/nodeclient` (`api.go`)
- retry/backoff: успіх після N фейлів;
- stop після `MaxRetries`;
- обробка non-200;
- таймаут/мережеві помилки.

4. `backend/internal/mining` (`miner.go`)
- `compileDifficultyBits` + `checkDifficultyFast` (table-driven);
- зупинка по `context cancel`;
- при знайденому nonce викликається submit (через mock інтерфейс).

### Nice-to-have (писати вибірково)
1. `config`
- валідація/дефолти;
- конвертація `seconds/ms -> time.Duration`.

2. `stats`
- агрегування `GetDashboardData`;
- uptime format;
- hashrate history rollover.

### Практичне правило для старту
1. На кожен модуль спочатку: happy path + 1 error path.
2. Далі додавати edge cases по мірі рефактору.
3. Тестувати поведінку, а не внутрішню реалізацію.
4. Мокати тільки зовнішні залежності: HTTP, filesystem, час, random.
