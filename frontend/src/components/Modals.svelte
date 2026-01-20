<script lang="ts">
    import { AddWallet, RenameWallet, DeleteWallet, UpdateWalletKey } from '../../wailsjs/go/main/App';
    import { notifications } from '../stores';
    import type { backend } from '../../wailsjs/go/models';
    
    export let type: string = ''; 
    export let wallet: backend.WalletStats | null = null;
    export let onClose: () => void;
    
    let name = '';
    let address = '';
    let privateKey = '';
    let password = '';
    
    if (wallet && type === 'edit') {
        name = wallet.name;
    }
    
    async function handleSubmit() {
        if (type === 'add') {
            const err = await AddWallet(name, address, privateKey);
            if (err) notifications.error(err); 
            else {
                notifications.success("Гаманець додано успішно");
                onClose();
            }
        } else if (type === 'edit') {
            await RenameWallet(wallet.address, name);
            notifications.success("Гаманець перейменовано");
            onClose();
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
</script>

<div class="modal-wrap open">
    <div class="glass-card modal no-hover">
        {#if type === 'add'}
            <h2>Додати гаманець</h2>
            <form on:submit|preventDefault={handleSubmit}>
                <label class="field-label">Назва</label>
                <input class="field" placeholder="Ім'я воркера" bind:value={name} required maxlength="50">
                
                <label class="field-label">Адреса</label>
                <input class="field" placeholder="hash адреса гаманця" bind:value={address} required>
                
                <label class="field-label">
                    Приватний ключ <span class="optional">(Необов'язково)</span>
                </label>
                <input type="password" class="field" placeholder="Для транзакцій" bind:value={privateKey}>
                
                <button type="submit" class="btn btn-primary btn-xl" style="width: 100%;">
                    <i class="fas fa-plus"></i> Додати
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
            <h2>Оновити ключ</h2>
            <form on:submit|preventDefault={handleSubmit}>
                <label class="field-label">Приватний ключ</label>
                <input type="password" class="field" placeholder="Вставте ключ" bind:value={privateKey} required>
                
                <label class="field-label" style="margin-top: 15px;">Пароль адміністратора</label>
                <input type="password" class="field" placeholder="Підтвердіть пароль" bind:value={password} required>
                
                <button type="submit" class="btn btn-primary btn-xl" style="width: 100%;">
                    <i class="fas fa-key"></i> Оновити
                </button>
            </form>
        {/if}
        
        <button type="button" class="btn" style="width: 100%; background: transparent; margin-top: 10px; color: #94a3b8;" on:click={onClose}>
            Скасувати
        </button>
    </div>
</div>
