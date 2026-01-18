<script>
    import { onMount, createEventDispatcher } from 'svelte';
    import { IsStorageInitialized, InitStorage, UnlockStorage } from '../../wailsjs/go/main/App';

    const dispatch = createEventDispatcher();

    let isSetup = false;
    let password = '';
    let confirmPassword = '';
    let error = '';
    let loading = true;

    onMount(async () => {
        const initialized = await IsStorageInitialized();
        isSetup = !initialized;
        loading = false;
    });

    async function handleSubmit() {
        error = '';
        
        if (isSetup) {
            if (password.length < 4) {
                error = "Пароль надто короткий";
                return;
            }
            if (password !== confirmPassword) {
                error = "Паролі не співпадають";
                return;
            }
            
            const err = await InitStorage(password);
            if (err) error = err;
            else dispatch('login');
            
        } else {
            const err = await UnlockStorage(password);
            if (err) error = err;
            else dispatch('login');
        }
    }
</script>

<div class="auth-container">
    <div class="glass-card auth-card">
        <div class="brand" style="justify-content: center; margin-bottom: 30px;">
            <i class="fas fa-cube" style="color: var(--primary)"></i> S-UAH MINER
        </div>

        {#if loading}
            <div style="text-align: center; color: #94a3b8;">Перевірка сховища...</div>
        {:else}
            <h2 style="text-align: center; margin-bottom: 20px;">
                {isSetup ? 'Початкове налаштування' : 'Вхід адміністратора'}
            </h2>
            
            <form on:submit|preventDefault={handleSubmit}>
                <label class="field-label">Пароль адміністратора</label>
                <input type="password" class="field" bind:value={password} placeholder="Введіть пароль" required autofocus>
                
                {#if isSetup}
                    <label class="field-label">Підтвердження паролю</label>
                    <input type="password" class="field" bind:value={confirmPassword} placeholder="Повторіть пароль" required>
                {/if}
                
                {#if error}
                    <div style="color: var(--danger); font-size: 0.9rem; margin-bottom: 15px; text-align: center;">
                        {error}
                    </div>
                {/if}
                
                <button type="submit" class="btn btn-primary btn-xl" style="width: 100%;">
                    {isSetup ? 'Створити сховище' : 'Розблокувати'}
                </button>
            </form>
            
            {#if isSetup}
                <div style="margin-top: 20px; font-size: 0.8rem; color: #64748b; text-align: center;">
                    Цей пароль зашифрує ваші гаманці та налаштування. Не загубіть його.
                </div>
            {/if}
        {/if}
    </div>
</div>

<style>
    .auth-container {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 100vh;
        width: 100%;
    }
    
    .auth-card {
        width: 100%;
        max-width: 400px;
        padding: 40px;
    }
</style>