<script lang="ts">
    import { onMount } from "svelte";
    import {
        GetConfig,
        UpdateConfig,
        HasPassword,
        SendTestTelegramMessage,
        ResolveTelegramChatID,
    } from "../../wailsjs/go/main/App";
    import { notifications } from "../stores";

    let loaded = false;
    let saving = false;
    let testing = false;
    let resolving = false;
    let password = "";
    let passwordRequired = true;

    let config: any = null;

    let auto = {
        telegram_bot_token: "",
        telegram_chat_id: "",
        schedule_start: "",
        schedule_stop: "",
        block_target: 0,
        session_minutes: 0,
        schedule_enabled: false,
        notify_on_start: false,
        notify_on_stop: false,
        notify_on_target: false,
        notify_on_error: false,
    };

    onMount(async () => {
        try {
            passwordRequired = await HasPassword();
        } catch {
            passwordRequired = true;
        }
        const cfg = await GetConfig();
        config = cfg;
        if (cfg.automation) {
            auto = {
                telegram_bot_token: cfg.automation.telegram_bot_token || "",
                telegram_chat_id: cfg.automation.telegram_chat_id || "",
                schedule_start: cfg.automation.schedule_start || "",
                schedule_stop: cfg.automation.schedule_stop || "",
                block_target: cfg.automation.block_target || 0,
                session_minutes: cfg.automation.session_minutes || 0,
                schedule_enabled: !!cfg.automation.schedule_enabled,
                notify_on_start: !!cfg.automation.notify_on_start,
                notify_on_stop: !!cfg.automation.notify_on_stop,
                notify_on_target: !!cfg.automation.notify_on_target,
                notify_on_error: !!cfg.automation.notify_on_error,
            };
        }
        loaded = true;
    });

    async function save() {
        if (!config) return;
        saving = true;
        try {
            const merged = { ...config, automation: auto };
            const err = await UpdateConfig(merged, password);
            if (err) {
                notifications.error("Помилка: " + err);
            } else {
                notifications.success("Автоматизацію збережено");
                password = "";
            }
        } finally {
            saving = false;
        }
    }

    async function sendTest() {
        if (!auto.telegram_bot_token || !auto.telegram_chat_id) {
            notifications.error("Вкажіть токен та Chat ID");
            return;
        }
        testing = true;
        try {
            const err = await SendTestTelegramMessage(
                auto.telegram_bot_token,
                auto.telegram_chat_id,
            );
            if (err) {
                notifications.error("Telegram: " + err);
            } else {
                notifications.success("Тест надіслано в Telegram");
            }
        } finally {
            testing = false;
        }
    }

    async function resolveId() {
        if (!auto.telegram_bot_token || !auto.telegram_chat_id) {
            notifications.error("Вкажіть токен та Chat ID / username");
            return;
        }
        resolving = true;
        try {
            const id = await ResolveTelegramChatID(
                auto.telegram_bot_token,
                auto.telegram_chat_id,
            );
            auto.telegram_chat_id = id;
            notifications.success("Chat ID розпізнано");
        } catch (e: any) {
            notifications.error(
                "Не вдалося: " + (e?.message || e?.toString() || "помилка"),
            );
        } finally {
            resolving = false;
        }
    }
</script>

<div class="content-wrapper">
    <div class="dash-header" style="flex-shrink: 0;">
        <div class="page-title">
            <i class="fab fa-telegram" style="color: #26a5e4"></i> Автоматизація
        </div>
    </div>

    {#if !loaded}
        <div class="glass-card" style="padding: 30px; text-align: center;">
            <i class="fas fa-spinner fa-spin"></i> Завантаження...
        </div>
    {:else}
        <div
            class="glass-card"
            style="padding: 24px; margin-bottom: 18px; max-width: 900px;"
        >
            <div class="sec-title">
                <i class="fas fa-key"></i> Підключення бота
            </div>
            <p class="hint">
                Створіть бота через <b>@BotFather</b> і візьміть токен. Потім
                напишіть будь-що своєму боту, щоб він отримав ваш Chat ID —
                його можна отримати автоматично, якщо замість ID вказати
                <b>@username</b> публічного чату.
            </p>

            <div class="form-row">
                <label>Telegram Bot Token</label>
                <input
                    type="text"
                    class="input"
                    bind:value={auto.telegram_bot_token}
                    placeholder="123456:ABC-..."
                />
            </div>

            <div class="form-row">
                <label>Chat ID або @username</label>
                <div style="display:flex; gap:8px;">
                    <input
                        type="text"
                        class="input"
                        bind:value={auto.telegram_chat_id}
                        placeholder="123456789 або @my_channel"
                        style="flex:1"
                    />
                    <button
                        class="btn btn-ghost"
                        on:click={resolveId}
                        disabled={resolving}
                        title="Розпізнати Chat ID за username"
                    >
                        {#if resolving}
                            <i class="fas fa-spinner fa-spin"></i>
                        {:else}
                            <i class="fas fa-magnifying-glass"></i>
                        {/if}
                    </button>
                    <button
                        class="btn btn-primary"
                        on:click={sendTest}
                        disabled={testing}
                    >
                        {#if testing}
                            <i class="fas fa-spinner fa-spin"></i> Тест
                        {:else}
                            <i class="fas fa-paper-plane"></i> Тест
                        {/if}
                    </button>
                </div>
            </div>
        </div>

        <div
            class="glass-card"
            style="padding: 24px; margin-bottom: 18px; max-width: 900px;"
        >
            <div class="sec-title">
                <i class="fas fa-bell"></i> Сповіщення
            </div>
            <div class="toggle-grid">
                <label class="toggle-row">
                    <input
                        type="checkbox"
                        bind:checked={auto.notify_on_start}
                    />
                    <span>Старт майнінгу</span>
                </label>
                <label class="toggle-row">
                    <input type="checkbox" bind:checked={auto.notify_on_stop} />
                    <span>Зупинка майнінгу</span>
                </label>
                <label class="toggle-row">
                    <input
                        type="checkbox"
                        bind:checked={auto.notify_on_target}
                    />
                    <span>Досягнення цілі по блоках</span>
                </label>
                <label class="toggle-row">
                    <input
                        type="checkbox"
                        bind:checked={auto.notify_on_error}
                    />
                    <span>Помилки (1 раз на 60с)</span>
                </label>
            </div>
        </div>

        <div
            class="glass-card"
            style="padding: 24px; margin-bottom: 18px; max-width: 900px;"
        >
            <div class="sec-title">
                <i class="fas fa-bullseye"></i> Ціль по блоках
            </div>
            <p class="hint">
                Коли кількість зарахованих блоків у сесії досягне цього
                значення — майнінг буде призупинено. <code>0</code> вимикає правило.
            </p>
            <div class="form-row">
                <label>Зупинити після N зарахованих блоків</label>
                <input
                    type="number"
                    min="0"
                    class="input"
                    bind:value={auto.block_target}
                />
            </div>
        </div>

        <div
            class="glass-card"
            style="padding: 24px; margin-bottom: 18px; max-width: 900px;"
        >
            <div class="sec-title">
                <i class="fas fa-stopwatch"></i> Таймер сесії
            </div>
            <p class="hint">
                Через скільки хвилин з моменту запуску автоматично зупинити.
                <code>0</code> — без таймера.
            </p>
            <div class="form-row">
                <label>Тривалість сесії, хв</label>
                <input
                    type="number"
                    min="0"
                    class="input"
                    bind:value={auto.session_minutes}
                />
            </div>
        </div>

        <div
            class="glass-card"
            style="padding: 24px; margin-bottom: 18px; max-width: 900px;"
        >
            <div class="sec-title">
                <i class="fas fa-calendar-days"></i> Розклад
            </div>
            <p class="hint">
                Автоматично запускати та зупиняти майнінг у заданому вікні.
                Вікно може переходити через північ (напр. 22:00 → 07:00).
            </p>
            <label class="toggle-row" style="margin-bottom: 10px;">
                <input type="checkbox" bind:checked={auto.schedule_enabled} />
                <span>Увімкнути розклад</span>
            </label>

            <div class="sched-grid">
                <div class="form-row">
                    <label>Старт (HH:MM)</label>
                    <input
                        type="time"
                        class="input"
                        bind:value={auto.schedule_start}
                        disabled={!auto.schedule_enabled}
                    />
                </div>
                <div class="form-row">
                    <label>Стоп (HH:MM)</label>
                    <input
                        type="time"
                        class="input"
                        bind:value={auto.schedule_stop}
                        disabled={!auto.schedule_enabled}
                    />
                </div>
            </div>
        </div>

        <div
            class="glass-card"
            style="padding: 24px; max-width: 900px;"
        >
            <div class="sec-title">
                <i class="fas fa-floppy-disk"></i> Збереження
            </div>
            {#if passwordRequired}
                <div class="form-row">
                    <label>Пароль для збереження</label>
                    <input
                        type="password"
                        class="input"
                        bind:value={password}
                        placeholder="••••••••"
                    />
                </div>
            {/if}
            <button
                class="btn btn-primary"
                on:click={save}
                disabled={saving || (passwordRequired && !password)}
            >
                {#if saving}
                    <i class="fas fa-spinner fa-spin"></i> Збереження...
                {:else}
                    <i class="fas fa-check"></i> Зберегти автоматизацію
                {/if}
            </button>
        </div>
    {/if}
</div>

<style>
    .sec-title {
        font-size: 1.05rem;
        font-weight: 700;
        margin-bottom: 10px;
        display: flex;
        align-items: center;
        gap: 10px;
        color: #e2e8f0;
    }
    .hint {
        color: #94a3b8;
        font-size: 0.85rem;
        margin-bottom: 14px;
        line-height: 1.5;
    }
    .hint code {
        background: rgba(255, 255, 255, 0.06);
        padding: 1px 6px;
        border-radius: 4px;
        font-family: var(--font-mono);
    }
    .form-row {
        margin-bottom: 12px;
    }
    .form-row label {
        display: block;
        font-size: 0.78rem;
        text-transform: uppercase;
        letter-spacing: 0.6px;
        color: #94a3b8;
        margin-bottom: 6px;
        font-weight: 600;
    }
    .input {
        width: 100%;
        background: rgba(255, 255, 255, 0.04);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 8px;
        padding: 10px 12px;
        color: #e2e8f0;
        font-family: var(--font-mono);
    }
    .input:focus {
        outline: none;
        border-color: var(--primary);
        background: rgba(255, 255, 255, 0.06);
    }
    .toggle-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
        gap: 10px;
    }
    .toggle-row {
        display: flex;
        align-items: center;
        gap: 10px;
        background: rgba(255, 255, 255, 0.03);
        padding: 10px 12px;
        border-radius: 8px;
        border: 1px solid rgba(255, 255, 255, 0.06);
        cursor: pointer;
        font-size: 0.9rem;
        color: #cbd5e1;
    }
    .toggle-row input {
        accent-color: var(--primary);
    }
    .sched-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 14px;
    }
    .btn-ghost {
        background: rgba(255, 255, 255, 0.04);
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: #cbd5e1;
    }
    .btn-ghost:hover {
        background: rgba(255, 255, 255, 0.08);
    }
</style>
