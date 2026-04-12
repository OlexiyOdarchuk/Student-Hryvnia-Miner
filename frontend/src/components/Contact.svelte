<script lang="ts">
    import { GetConfig, SendMessageToDeveloper } from '../../wailsjs/go/main/App';
    import { notifications } from '../stores';
    import { onMount } from 'svelte';

    let contact = '';
    let message = '';
    let isSending = false;

    onMount(async () => {
        const config = await GetConfig();
        contact = config.telegram_handle;
    });

    async function sendMessage() {
        if (!message.trim()) return;
        
        isSending = true;
        try {
            await SendMessageToDeveloper(contact, message);
            notifications.success("Повідомлення надіслано розробнику!");
            message = '';
        } catch (e) {
            notifications.error("Не вдалося надіслати повідомлення.");
        } finally {
            isSending = false;
        }
    }
</script>

<div class="content-wrapper">
    <div class="dash-header">
        <div class="page-title">Зворотний зв'язок</div>
    </div>

    <div class="glass-card" style="padding: 40px; max-width: 800px; margin: 0 auto;">
        <div style="text-align: center; margin-bottom: 30px;">
            <i class="fas fa-envelope-open-text" style="font-size: 3rem; color: var(--primary); margin-bottom: 20px;"></i>
            <h2>Написати розробнику</h2>
            <p style="color: #94a3b8; margin-top: 10px;">
                Маєте ідею, знайшли баг або хочете подякувати? <br> 
                Ваше повідомлення прийде мені прямо в Telegram.
            </p>
        </div>

        <form on:submit|preventDefault={sendMessage} style="display: grid; gap: 20px;">
            <div>
                <label class="field-label">Ваш контакт (необов'язково)</label>
                <input class="field" placeholder="@username або email" bind:value={contact}>
            </div>

            <div>
                <label class="field-label">Ваше повідомлення</label>
                <textarea class="field" rows="6" placeholder="Опишіть вашу ідею або проблему..." bind:value={message} required></textarea>
            </div>

            <button type="submit" class="btn btn-primary btn-xl" disabled={isSending}>
                {#if isSending}
                    <i class="fas fa-spinner fa-spin"></i> Надсилаю...
                {:else}
                    <i class="fas fa-paper-plane"></i> Надіслати повідомлення
                {/if}
            </button>
        </form>

        <div style="margin-top: 40px; padding-top: 30px; border-top: 1px solid rgba(255,255,255,0.1); text-align: center;">
            <div style="font-size: 0.85rem; color: #64748b; margin-bottom: 15px;">Також ви можете знайти мене тут:</div>
            <div style="display: flex; justify-content: center; gap: 20px;">
                <a href="https://t.me/NeShawyha" target="_blank" class="social-link">
                    <i class="fab fa-telegram"></i> Telegram
                </a>
                <a href="https://github.com/OlexiyOdarchuk" target="_blank" class="social-link">
                    <i class="fab fa-github"></i> GitHub
                </a>
            </div>
        </div>
    </div>
</div>

<style>
    .social-link {
        color: #94a3b8;
        text-decoration: none;
        display: flex;
        align-items: center;
        gap: 8px;
        transition: 0.3s;
        font-weight: 600;
    }
    .social-link:hover {
        color: var(--primary);
    }
    h2 {
        font-size: 1.8rem;
        font-weight: 800;
        margin: 0;
    }
</style>
