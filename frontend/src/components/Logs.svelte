<script lang="ts">
    import { logs } from '../stores';
    
    function clearLogs() {
        logs.set([]);
    }
</script>

<div class="content-wrapper">
    <div class="dash-header" style="flex-shrink: 0;">
        <div class="page-title">Системний термінал</div>
        <button class="btn btn-icon" on:click={clearLogs} style="background: rgba(239, 68, 68, 0.1); color: var(--danger);">
            <i class="fas fa-trash"></i>
        </button>
    </div>

    
    <div class="term-layout">
        <div class="glass-card no-hover term-window">
            <div id="terminal">
                <div style="color: var(--neon-cyan); opacity: 0.8; margin-bottom: 10px;">
                    <i class="fas fa-terminal"></i> [СИСТЕМА] Квантовий інтерфейс ініціалізовано...
                </div>
                
                {#each [...$logs].reverse() as log}
                    <div class="log-line">
                        <span class="log-time">[{log.time}]</span>
                        <span style={log.type === 'ERROR' ? 'color: var(--danger)' : log.type === 'INFO' ? 'color: var(--success)' : ''}>{log.message}</span>
                    </div>
                {/each}
            </div>
        </div>
        
        <div class="term-sidebar">
            <div class="glass-card no-hover" style="padding: 25px;">
                <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 15px;">
                    <i class="fas fa-info-circle" style="font-size: 1.5rem; color: var(--primary);"></i>
                    <div>
                        <div style="font-weight: 700; color: white; font-size: 1.1rem;">Інфо</div>
                        <div style="font-size: 0.85rem; color: #94a3b8;">Статус системи</div>
                    </div>
                </div>
                <div style="font-size: 0.9rem; color: #cbd5e1; line-height: 1.6;">
                    Моніторинг активний. <br>
                    Логи надходять у реальному часі.
                </div>
            </div>
        </div>
    </div>
</div>

<style>
    .term-layout {
        display: grid;
        grid-template-columns: 3fr 1fr;
        gap: 30px;
        flex: 1;
        min-height: 0;
    }

    .term-window {
        display: flex; 
        flex-direction: column; 
        overflow: hidden; 
    }

    #terminal {
        flex: 1; 
        padding: 30px; 
        overflow-y: auto; 
        font-family: var(--font-mono); 
        color: #cbd5e1; 
        font-size: 0.95rem; 
        background: rgba(0,0,0,0.3);
    }

    .log-line {
        margin-top: 5px;
        word-break: break-all;
    }

    .log-time {
        opacity: 0.5; 
        margin-right: 10px;
        color: #94a3b8;
    }

    .term-sidebar {
        display: flex; 
        flex-direction: column; 
        gap: 20px;
    }

    @media (max-width: 900px) {
        .term-layout {
            grid-template-columns: 1fr; 
        }
        .term-sidebar {
            display: none; 
        }
    }
</style>
