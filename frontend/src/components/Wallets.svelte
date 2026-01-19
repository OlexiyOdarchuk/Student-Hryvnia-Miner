<script lang="ts">
    import { stats } from '../stores';
    import { ToggleWallet } from '../../wailsjs/go/main/App';
    import { ClipboardSetText } from '../../wailsjs/runtime/runtime';
    import { notifications } from '../stores'; // Assuming you use this for feedback
    import type { backend } from '../../wailsjs/go/models';

    function openModal(type: string, wallet?: backend.WalletStats) {
        document.dispatchEvent(new CustomEvent('open-modal', { detail: { type, wallet } }));
    }

    async function toggle(wallet: backend.WalletStats) {
        await ToggleWallet(wallet.address);
    }

    function copyAddress(addr: string) {
        ClipboardSetText(addr);
        notifications.info("Адреса скопійована: " + addr.substring(0, 10) + "...");
    }

    // Row Pulse Logic
    let pulsingWallets = {}; // map address -> boolean
    let lastMined = {}; // map address -> count

    $: if ($stats && $stats.wallets) {
        $stats.wallets.forEach(w => {
            const last = lastMined[w.address] || 0;
            if (w.total_mined > last) {
                // Only pulse if not first load (or if you want pulse on load, remove check)
                if (last !== 0 || w.total_mined === 1) {
                    pulsingWallets[w.address] = true;
                    setTimeout(() => {
                        pulsingWallets[w.address] = false;
                    }, 1000);
                }
            }
            lastMined[w.address] = w.total_mined;
        });
    }
</script>

<div class="content-wrapper">
    
    <div class="dash-header" style="flex-shrink: 0; align-items: center;">
        <div class="page-title">Гаманці</div>
        <button class="btn btn-primary" on:click={() => openModal('add')}>
            <i class="fas fa-plus"></i> Додати гаманець
        </button>
    </div>

    <div class="glass-card table-container">
        <div class="table-wrap">
            <table class="lux-table">
                <thead>
                    <tr>
                        <th style="width: 20%;">Назва</th>
                        <th style="width: 25%;">Адреса</th>
                        <th style="width: 10%;">Статус</th>
                        <th class="hide-sm" style="width: 10%;">С. Блоки</th>
                        <th class="hide-sm" style="width: 10%;">З. Блоки</th>
                        <th style="width: 15%;">Баланс</th>
                        <th style="text-align: right; width: 10%;">Дії</th>
                    </tr>
                </thead>
                <tbody>
                    {#each ($stats.wallets || []) as wallet}
                    <tr class:row-pulse={pulsingWallets[wallet.address]}>
                        <td style="font-weight: 700; color: white;">{wallet.name}</td>
                        <td>
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <div class="addr-chip" title="Натисніть щоб скопіювати" on:click={() => copyAddress(wallet.address)} role="button" tabindex="0">
                                {wallet.address.substring(0, 10)}...{wallet.address.substring(wallet.address.length - 8)}
                                <i class="fas fa-copy" style="margin-left: 8px; font-size: 0.8em; opacity: 0.7;"></i>
                            </div>
                        </td>
                        <td>
                            {#if wallet.working}
                            <span class="status-badge status-active">АКТИВНИЙ</span>
                            {:else}
                            <span class="status-badge status-paused">ПАУЗА</span>
                            {/if}
                        </td>
                        <td class="hide-sm">{wallet.session_mined}</td>
                        <td class="hide-sm">{wallet.total_mined}</td>
                        <td style="font-family: var(--font-mono); color: var(--success); font-weight: 700;">
                            {wallet.server_balance.toFixed(2)} S-UAH
                        </td>
                        <td style="text-align: right;">
                            <div class="wallet-actions">
                                 <button class="btn btn-icon" on:click={() => toggle(wallet)} title={wallet.working ? "Призупинити" : "Відновити"}>
                                    <i class="fas {wallet.working ? 'fa-pause' : 'fa-play'}" style="color: {wallet.working ? 'var(--warning)' : 'var(--success)'}"></i>
                                </button>
                                <button class="btn btn-icon" on:click={() => openModal('key', wallet)} title="Оновити ключ">
                                    <i class="fas fa-key"></i>
                                </button>
                                <button class="btn btn-icon" on:click={() => openModal('edit', wallet)} title="Редагувати">
                                    <i class="fas fa-pen"></i>
                                </button>
                                <button class="btn btn-icon danger" on:click={() => openModal('del', wallet)} title="Видалити">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                        </td>
                    </tr>
                    {/each}
                    {#if ($stats.wallets || []).length === 0}
                    <tr>
                        <td colspan="7" style="text-align: center; padding: 40px; color: #64748b;">
                            Немає гаманців. Натисніть "Додати гаманець" вище.
                        </td>
                    </tr>
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>

<style>
    .table-container {
        flex: 1;
        overflow: hidden; /* Hide overflow on card */
        display: flex;
        flex-direction: column;
    }
    
    .table-wrap {
        flex: 1;
        overflow-y: auto; /* Scroll table internally */
        padding: 20px;
    }

    .addr-chip {
        cursor: pointer;
        transition: background 0.2s;
    }
    .addr-chip:hover {
        background: rgba(129, 140, 248, 0.2);
        color: white;
    }

    @keyframes pulse-row {
        0% { background: rgba(52, 211, 153, 0.15); }
        100% { background: transparent; }
    }
    .row-pulse {
        animation: pulse-row 1s ease-out;
    }
</style>