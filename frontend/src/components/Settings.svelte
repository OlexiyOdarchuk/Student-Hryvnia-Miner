<script lang="ts">
    import { onMount } from 'svelte';
    import { GetConfig, UpdateConfig, ChangePassword } from '../../wailsjs/go/main/App';

    let config = {
        base_url: '',
        server_port: '',
        difficulty: 0,
        max_retries: 0,
        retry_delay_ms: 0,
        balance_freq_s: 0,
        http_timeout: 0
    };
    
    let password = '';
    let message = '';
    let isError = false;
    
    // Password change vars
    let oldPass = '';
    let newPass = '';
    let confirmPass = '';
    let passMsg = '';
    let isPassError = false;

    onMount(async () => {
        config = await GetConfig();
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
</script>

<div class="content-wrapper">
    <div class="dash-header" style="flex-shrink: 0;">
        <div class="page-title">Налаштування</div>
    </div>

    <div class="glass-card" style="padding: 40px; max-width: 800px; margin: 0 auto; width: 100%; overflow-y: auto;">
        
        <!-- Config Form -->
        <h3 style="margin-bottom: 20px; color: white;">Конфігурація майнера</h3>
        <form on:submit|preventDefault={save} style="display: grid; gap: 20px; margin-bottom: 40px;">
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                <div>
                    <label class="field-label">Базовий URL сервера</label>
                    <input class="field" bind:value={config.base_url} required>
                </div>
                <div>
                    <label class="field-label">Локальний API порт</label>
                    <input class="field" bind:value={config.server_port} required>
                </div>
            </div>
            
            <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 20px;">
                <div>
                    <label class="field-label">Складність</label>
                    <input type="number" class="field" bind:value={config.difficulty} required>
                </div>
                <div>
                    <label class="field-label">Макс. спроб</label>
                    <input type="number" class="field" bind:value={config.max_retries} required>
                </div>
                <div>
                    <label class="field-label">Затримка (мс)</label>
                    <input type="number" class="field" bind:value={config.retry_delay_ms} required>
                </div>
            </div>
            
            <div style="margin-top: 10px; padding-top: 10px; border-top: 1px solid rgba(255,255,255,0.1);">
                <label class="field-label" style="color: var(--primary);">Підтвердіть поточним паролем</label>
                <input type="password" class="field" bind:value={password} placeholder="Поточний пароль" required>
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
        
        <!-- Password Change Form -->
        <h3 style="margin-bottom: 20px; color: white; border-top: 1px solid rgba(255,255,255,0.1); padding-top: 30px;">Зміна паролю адміністратора</h3>
        <form on:submit|preventDefault={changePass} style="display: grid; gap: 20px;">
            <div>
                <label class="field-label">Поточний пароль</label>
                <input type="password" class="field" bind:value={oldPass} required>
            </div>
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                <div>
                    <label class="field-label">Новий пароль</label>
                    <input type="password" class="field" bind:value={newPass} required minlength="4">
                </div>
                <div>
                    <label class="field-label">Підтвердіть новий</label>
                    <input type="password" class="field" bind:value={confirmPass} required minlength="4">
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
