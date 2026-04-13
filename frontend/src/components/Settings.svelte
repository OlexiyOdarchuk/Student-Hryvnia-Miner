<script lang="ts">
    import { onMount } from 'svelte';
    import { GetConfig, UpdateConfig, ChangePassword, GetSystemInfo, GetConfigFilePath, OpenConfigFolder } from '../../wailsjs/go/main/App';
    let config = {
        miner_id: '',
        telegram_handle: '',
        base_url: '',
        server_port: '',
        difficulty: 0,
        max_retries: 0,
        retry_delay_ms: 0,
        balance_freq_s: 0,
        block_check_freq_ms: 0,
        http_timeout: 0,
        threads: 0
    };
    
    let maxCores = 4; 
    
    let password = '';
    let message = '';
    let isError = false;
    
    
    let oldPass = '';
    let newPass = '';
    let confirmPass = '';
    let passMsg = '';
    let isPassError = false;
    let configPath = "Завантаження..."; 

    onMount(async () => {
        config = await GetConfig();
        configPath = await GetConfigFilePath(); 
        try {
            const info = await GetSystemInfo();
            if (info && info.cpu_cores) {
                maxCores = info.cpu_cores;
            }
        } catch(e) { console.error(e); }
    });

    async function save() {
        message = '';
        isError = false;
        
        const err = await UpdateConfig(config, password);
        if (err) {
            isError = true;
            message = "Помилка: " + err;
        } else {
            message = "Налаштування збережено успішно!";
            password = '';
        }
    }
    
    async function changePass() {
        passMsg = '';
        isPassError = false;
        
        if (newPass !== confirmPass) {
            isPassError = true;
            passMsg = "Паролі не співпадають";
            return;
        }
        
        const err = await ChangePassword(oldPass, newPass);
        if (err) {
            isPassError = true;
            passMsg = "Помилка: " + err;
        } else {
            passMsg = "Пароль змінено успішно!";
            oldPass = ''; newPass = ''; confirmPass = '';
        }
    }

    function handleOpenFolder() {
        OpenConfigFolder();
    }

</script>

<div class="content-wrapper">
    <div class="dash-header" style="flex-shrink: 0;">
        <div class="page-title">Налаштування</div>
    </div>

    <div class="glass-card" style="padding: 40px; max-width: 800px; margin: 0 auto; width: 100%; overflow-y: auto;">
        
        
        <h3 style="margin-bottom: 20px; color: white;">Конфігурація майнера</h3>
        <div class="settings-card">
            <h3>Дані програми</h3>
                
            <div class="path-container">
                <div class="path-info">
                    <span class="label">Шлях до конфігурації:</span>
                    <code style="user-select: all;">{configPath}</code>
                </div>
                    
                <button class="btn btn-secondary btn-sm" on:click={handleOpenFolder}>
                    <i class="fas fa-folder-open"></i> Відкрити папку
                </button>
            </div>
        </div>
        <form on:submit|preventDefault={save} style="display: grid; gap: 25px;">
            <div style="padding: 15px; background: rgba(129, 140, 248, 0.05); border-radius: 12px; border: 1px solid rgba(129, 140, 248, 0.1);">
                <label class="field-label" style="color: var(--primary);">Ваш унікальний Miner ID</label>
                <code style="font-family: var(--font-mono); font-size: 0.9rem; color: #94a3b8; display: block; word-break: break-all;">{config.miner_id}</code>
                <div style="font-size: 0.7rem; color: #64748b; margin-top: 5px;">Цей ID використовується для анонімної статистики та ідентифікації вашої ноди.</div>
            </div>

            <div>
                <label class="field-label">Ваш Telegram (@username або посилання)</label>
                <input class="field" placeholder="@username (необов'язково)" bind:value={config.telegram_handle}>
                <div style="font-size: 0.75rem; color: #64748b; margin-top: -15px;">Для зв'язку з розробником та персоналізації статистики.</div>
            </div>

            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                <div>
                    <label class="field-label">Базовий URL сервера</label>
                    <input class="field" bind:value={config.base_url} required>
                </div>
                <div>
                    <label class="field-label">Локальний порт монітору</label>
                    <input class="field" bind:value={config.server_port} required>
                </div>
            </div>
            
            <div style="margin-top: 10px; padding: 15px; background: rgba(255,255,255,0.03); border-radius: 12px; border: 1px solid rgba(255,255,255,0.05);">
                 <label class="field-label" style="display:flex; justify-content:space-between; margin-bottom: 10px;">
                    <span>Потоки CPU</span>
                    <span style="color: var(--primary); font-weight: 700; font-family: var(--font-mono);">{config.threads === 0 ? 'AUTO (' + maxCores + ')' : config.threads}</span>
                 </label>
                 <input type="range" class="range-slider" min="0" max={maxCores} step="1" bind:value={config.threads} style="width: 100%;">
                 <div style="font-size: 0.75rem; color: #64748b; margin-top: 8px;">
                    0 = Використовувати всі доступні ядра ({maxCores}). Зменшіть, якщо ПК гальмує.
                 </div>
            </div>

            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                <div>
                    <label class="field-label">Складність</label>
                    <input type="number" class="field" bind:value={config.difficulty} required>
                </div>
                <div>
                    <label class="field-label">Макс. спроб (API)</label>
                    <input type="number" class="field" bind:value={config.max_retries} required>
                </div>
            </div>
            
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                <div>
                    <label class="field-label">Затримка повторних запитів API (мс)</label>
                    <input type="number" class="field" bind:value={config.retry_delay_ms} required>
                </div>
                <div>
                    <label class="field-label">Синхр. блоків (мс)</label>
                    <input type="number" class="field" bind:value={config.block_check_freq_ms} required>
                    <div style="font-size: 0.75rem; color: #64748b; margin-top: -15px; margin-bottom: 20px;">Частота перевірки статусу мережі</div>
                </div>
            </div>
            
            <div style="margin-top: 10px; padding-top: 10px; border-top: 1px solid rgba(255,255,255,0.1);">
                <label class="field-label" style="color: var(--primary);">Підтвердіть поточним паролем</label>
                <input type="password" class="field" bind:value={password} placeholder="Поточний пароль">
            </div>
            
            {#if message}
                <div style="padding: 10px; border-radius: 8px; text-align: center; {isError ? 'background: rgba(239, 68, 68, 0.2); color: var(--danger);' : 'background: rgba(16, 185, 129, 0.2); color: var(--success);'}">
                    {message}
                </div>
            {/if}
            
            <button type="submit" class="btn btn-primary btn-xl">
                Зберегти конфігурацію
            </button>
        </form>
        
        
        <h3 style="margin-bottom: 20px; color: white; border-top: 1px solid rgba(255,255,255,0.1); padding-top: 30px;">Зміна паролю адміністратора</h3>
        <form on:submit|preventDefault={changePass} style="display: grid; gap: 20px;">
            <div>
                <label class="field-label">Поточний пароль</label>
                <input type="password" class="field" bind:value={oldPass}>
            </div>
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                <div>
                    <label class="field-label">Новий пароль</label>
                    <input type="password" class="field" bind:value={newPass} minlength="4">
                </div>
                <div>
                    <label class="field-label">Підтвердіть новий</label>
                    <input type="password" class="field" bind:value={confirmPass} minlength="4">
                </div>
            </div>
            
            {#if passMsg}
                <div style="padding: 10px; border-radius: 8px; text-align: center; {isPassError ? 'background: rgba(239, 68, 68, 0.2); color: var(--danger);' : 'background: rgba(16, 185, 129, 0.2); color: var(--success);'}">
                    {passMsg}
                </div>
            {/if}
            
            <button type="submit" class="btn btn-secondary btn-xl">
                Оновити пароль
            </button>
        </form>
    </div>
</div>