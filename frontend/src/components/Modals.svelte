<script lang="ts">
    import { AddWallet, RenameWallet, DeleteWallet, UpdateWalletKey, ImportWalletJSON, GetWalletJSONSecure, GetWalletKey, GenerateKeyPair } from '../../wailsjs/go/main/App';
    import { ClipboardSetText } from '../../wailsjs/runtime/runtime';
    import { notifications } from '../stores';
    import type { types } from '../../wailsjs/go/models';
    
    export let type: string = ''; 
    export let wallet: types.WalletStats | null = null;
    export let onClose: () => void;
    
    let name = '';
    let address = '';
    let privateKey = '';
    let password = '';
    let jsonContent = '';

    let addTab: 'create' | 'import' = 'create';
    let importTab: 'manual' | 'json' = 'manual';
    
    if (wallet && type === 'edit') {
        name = wallet.name;
    }
    
    async function handleSubmit() {
        if (type === 'add') {
            if (addTab === 'create') {
                try {
                    const keys = await GenerateKeyPair();
                    const pub = keys["public"];
                    const priv = keys["private"];

                    const err = await AddWallet(name, pub, priv);
                    if (err) notifications.error(err);
                    else {
                        notifications.success("Гаманець створено успішно");
                        onClose();
                    }
                } catch (e) {
                    notifications.error("Помилка генерації ключів: " + e);
                }
            } else {
                if (importTab === 'manual') {
                    const err = await AddWallet(name, address, privateKey);
                    if (err) notifications.error(err); 
                    else {
                        notifications.success("Гаманець імпортовано успішно");
                        onClose();
                    }
                } else {
                    const err = await ImportWalletJSON(jsonContent);
                    if (err) notifications.error(err);
                    else {
                        notifications.success("Гаманець імпортовано з JSON");
                        onClose();
                    }
                }
            }
        } else if (type === 'edit') {
            const err = await RenameWallet(wallet.address, name);
            if (err) notifications.error(err);
            else {
                notifications.success("Гаманець перейменовано");
                onClose();
            }
        } else if (type === 'del') {
            const err = await DeleteWallet(wallet.address, password);
            if (err) notifications.error(err); 
            else {
                notifications.success("Гаманець видалено");
                onClose();
            }
        } else if (type === 'key') {
            const err = await UpdateWalletKey(wallet.address, privateKey, password);
            if (err) notifications.error(err); 
            else {
                notifications.success("Приватний ключ оновлено");
                onClose();
            }
        }
    }

    async function handleExportKey() {
        try {
            const key = await GetWalletKey(wallet.address, password);
            ClipboardSetText(key);
            notifications.success("Приватний ключ скопійовано!");
        } catch (e) {
            notifications.error(e as string);
        }
    }

    async function handleExportJSON() {
        try {
            const json = await GetWalletJSONSecure(wallet.address, password);
            ClipboardSetText(json);
            notifications.success("JSON гаманця скопійовано!");
        } catch (e) {
            notifications.error(e as string);
        }
    }
</script>

<div class="modal-wrap open">
    <div class="glass-card modal no-hover">
        {#if type === 'add'}
            <h2>Додати гаманець</h2>
            
            <div class="tabs">
                <button class="tab-btn" class:active={addTab === 'create'} on:click={() => addTab = 'create'}>Створити новий</button>
                <button class="tab-btn" class:active={addTab === 'import'} on:click={() => addTab = 'import'}>Імпортувати</button>
            </div>

            <form on:submit|preventDefault={handleSubmit}>
                {#if addTab === 'create'}
                    <div class="tab-content">
                        <label class="field-label">Назва</label>
                        <input class="field" placeholder="Ім'я воркера" bind:value={name} required maxlength="50">
                        <p class="hint">Буде згенеровано нову пару ключів (ECDSA P-256).</p>
                    </div>
                {:else}
                    <div class="sub-tabs">
                        <label class="radio-label">
                            <input type="radio" bind:group={importTab} value="manual"> Вручну
                        </label>
                        <label class="radio-label">
                            <input type="radio" bind:group={importTab} value="json"> JSON
                        </label>
                    </div>

                    {#if importTab === 'manual'}
                        <label class="field-label">Назва</label>
                        <input class="field" placeholder="Ім'я воркера" bind:value={name} required maxlength="50">
                        
                        <label class="field-label">Адреса</label>
                        <input class="field" placeholder="hash адреса гаманця" bind:value={address} required>
                        
                        <label class="field-label">
                            Приватний ключ <span class="optional">(Необов'язково)</span>
                        </label>
                        <input type="password" class="field" placeholder="Для транзакцій" bind:value={privateKey}>
                    {:else}
                        <label class="field-label">JSON дані</label>
                        <textarea class="field" rows="5" placeholder={'{"name":"...", "pub":"...", "priv":"..."}'} bind:value={jsonContent} required></textarea>
                    {/if}
                {/if}

                <button type="submit" class="btn btn-primary btn-xl" style="width: 100%; margin-top: 20px;">
                    <i class="fas fa-{addTab === 'create' ? 'magic' : 'plus'}"></i> 
                    {addTab === 'create' ? 'Створити' : 'Додати'}
                </button>
            </form>

        {:else if type === 'edit'}
            <h2>Редагувати гаманець</h2>
            <form on:submit|preventDefault={handleSubmit}>
                 <label class="field-label">Нова назва</label>
                <input class="field" placeholder="Назва" bind:value={name} required maxlength="50">
                <button type="submit" class="btn btn-primary btn-xl" style="width: 100%;">
                    <i class="fas fa-save"></i> Зберегти
                </button>
            </form>

        {:else if type === 'del'}
             <h2 style="color: var(--danger);">Відключити гаманець?</h2>
            <p style="color: #94a3b8; margin-bottom: 30px;">Майнінг зупиниться для цього гаманця.</p>
            <form on:submit|preventDefault={handleSubmit}>
                <label class="field-label">Пароль адміністратора</label>
                <input type="password" class="field" placeholder="Підтвердіть пароль" bind:value={password} required>
                <button type="submit" class="btn btn-xl" style="width: 100%; background: var(--danger); color: white;">
                    <i class="fas fa-trash"></i> Відключити
                </button>
            </form>

        {:else if type === 'key'}
            <h2>Керування ключами</h2>
            
            <div class="key-actions-section">
                <label class="field-label">Експорт</label>
                <div class="btn-group">
                    <button type="button" class="btn btn-secondary" on:click={handleExportKey} disabled={!password}>
                        <i class="fas fa-copy"></i> Копіювати ключ
                    </button>
                    <button type="button" class="btn btn-secondary" on:click={handleExportJSON} disabled={!password}>
                        <i class="fas fa-file-code"></i> Експорт JSON
                    </button>
                </div>
            </div>

            <hr style="border-color: rgba(255,255,255,0.1); margin: 20px 0;">

            <form on:submit|preventDefault={handleSubmit}>
                <label class="field-label">Оновити приватний ключ</label>
                <input type="password" class="field" placeholder="Вставте новий ключ" bind:value={privateKey}>
                
                <label class="field-label" style="margin-top: 15px;">Пароль адміністратора (для всіх дій)</label>
                <input type="password" class="field" placeholder="Підтвердіть пароль" bind:value={password} required>
                
                <button type="submit" class="btn btn-primary btn-xl" style="width: 100%; margin-top: 15px;" disabled={!privateKey}>
                    <i class="fas fa-sync"></i> Оновити ключ
                </button>
            </form>
        {/if}
        
        <button type="button" class="btn" style="width: 100%; background: transparent; margin-top: 10px; color: #94a3b8;" on:click={onClose}>
            Скасувати
        </button>
    </div>
</div>

<style>
    .tabs {
        display: flex;
        background: rgba(0,0,0,0.2);
        border-radius: 8px;
        padding: 4px;
        margin-bottom: 20px;
    }
    .tab-btn {
        flex: 1;
        background: transparent;
        border: none;
        color: #94a3b8;
        padding: 10px;
        cursor: pointer;
        border-radius: 6px;
        transition: all 0.2s;
    }
    .tab-btn.active {
        background: rgba(255,255,255,0.1);
        color: white;
        font-weight: 600;
    }

    .sub-tabs {
        display: flex;
        gap: 20px;
        margin-bottom: 15px;
    }
    .radio-label {
        color: white;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .hint {
        font-size: 0.85em;
        color: #94a3b8;
        margin-top: 8px;
    }

    .btn-group {
        display: flex;
        gap: 10px;
    }
    .btn-secondary {
        flex: 1;
        background: rgba(255,255,255,0.1);
        color: white;
        border: 1px solid rgba(255,255,255,0.1);
    }
    .btn-secondary:hover:not(:disabled) {
        background: rgba(255,255,255,0.2);
    }
    .btn-secondary:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
</style>