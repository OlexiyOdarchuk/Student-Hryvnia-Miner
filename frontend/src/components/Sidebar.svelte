<script lang="ts">
    import { activeTab } from '../stores';
    import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
    import { GetConfig } from '../../wailsjs/go/main/App';

    function nav(tab) {
        activeTab.set(tab);
    }

    async function openWebMonitor() {
        const conf = await GetConfig();
        let port = conf.server_port || ":8080";
        if (port.startsWith(":")) port = port.substring(1);
        BrowserOpenURL(`http://localhost:${port}`);
    }
</script>

<nav class="sidebar">
    <div class="brand" style="--wails-draggable:drag">
        <i class="fas fa-cube" style="color: var(--primary)"></i> S-UAH MINER
    </div>
    
    <div class="nav-btn" class:active={$activeTab === 'dash'} on:click={() => nav('dash')}>
        <i class="fas fa-chart-pie"></i> 
        <span>Головна</span>
    </div>
    <div class="nav-btn" class:active={$activeTab === 'wallets'} on:click={() => nav('wallets')}>
        <i class="fas fa-wallet"></i> 
        <span>Гаманці</span>
    </div>
    <div class="nav-btn" class:active={$activeTab === 'transactions'} on:click={() => nav('transactions')}>
        <i class="fas fa-exchange-alt"></i> 
        <span>Транзакції</span>
    </div>
    <div class="nav-btn" class:active={$activeTab === 'stats'} on:click={() => nav('stats')}>
        <i class="fas fa-chart-bar"></i> 
        <span>Статистика</span>
    </div>
    <div class="nav-btn" class:active={$activeTab === 'logs'} on:click={() => nav('logs')}>
        <i class="fas fa-terminal"></i> 
        <span>Термінал</span>
    </div>
    <div class="nav-btn" class:active={$activeTab === 'settings'} on:click={() => nav('settings')}>
        <i class="fas fa-cog"></i> 
        <span>Налаштування</span>
    </div>

    <div class="nav-btn" class:active={$activeTab === 'help'} on:click={() => nav('help')}>
        <i class="fas fa-question-circle"></i> 
        <span>Інструкція</span>
    </div>

    <div class="nav-btn" class:active={$activeTab === 'contact'} on:click={() => nav('contact')}>
        <i class="fas fa-envelope"></i> 
        <span>Зв'язок</span>
    </div>

    
    <div style="height: 1px; background: rgba(255,255,255,0.1); margin: 10px 20px;"></div>

    <div class="nav-btn" on:click={openWebMonitor}>
        <i class="fas fa-globe"></i> 
        <span>Вебмоніторинг</span>
    </div>

     <div class="nav-btn" on:click={() => document.dispatchEvent(new CustomEvent('toggle-focus'))}>
        <i class="fas fa-expand"></i> 
        <span>Режим фокусу</span>
    </div>
    
    <div style="margin-top: auto; padding-top: 20px; border-top: 1px solid rgba(255,255,255,0.1);">
        <button class="btn btn-primary btn-xl" style="width: 100%;" on:click={() => document.dispatchEvent(new CustomEvent('open-modal', { detail: 'add' }))}>
            <i class="fas fa-plus"></i> Додати гаманець
        </button>
        
        <div style="margin-top: 30px; display: flex; flex-direction: column; gap: 12px; align-items: center;">
            <div style="font-size: 0.85rem; color: #64748b; text-align: center;">
                <div style="margin-bottom: 10px;">S-UAH Miner // by iShawyha</div>
                <div style="display: flex; justify-content: center; gap: 15px; font-size: 1.1rem;">
                    <i class="fab fa-telegram social-icon" style="cursor: pointer; transition: color 0.2s;" on:click={() => BrowserOpenURL('https://t.me/NeShawyha')}></i>
                    <i class="fab fa-github social-icon" style="cursor: pointer; transition: color 0.2s;" on:click={() => BrowserOpenURL('https://github.com/OlexiyOdarchuk')}></i>
                </div>
            </div>
        </div>
    </div>
</nav>

<style>
    .social-icon:hover {
        color: var(--primary);
    }
</style>