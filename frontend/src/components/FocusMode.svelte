<script lang="ts">
    import { createEventDispatcher, onMount, onDestroy } from 'svelte';
    import { stats } from '../stores';
    import { WindowFullscreen, WindowUnfullscreen } from '../../wailsjs/runtime/runtime';
    
    const dispatch = createEventDispatcher();
    
    function close() {
        dispatch('close');
    }

    onMount(() => {
        WindowFullscreen();
    });

    onDestroy(() => {
        WindowUnfullscreen();
    });
</script>

<div id="focus-layer" class="active">
    <button class="btn btn-icon focus-exit-btn" on:click={close} title="Exit Focus Mode">
        <i class="fas fa-times"></i>
    </button>
    
    <div class="focus-container">
        
        <div class="zen-ring">
            <div id="f-hash" class="zen-val">{$stats.hashrate.toFixed(2)}</div>
            <div class="zen-label">MH/s ШВИДКІСТЬ</div>
        </div>
        
        
        <div class="focus-stats">
            <div class="focus-stat-card">
                <div class="focus-stat-label">Загальний Баланс</div>
                <div id="f-balance" class="focus-stat-value success">{$stats.total_balance.toFixed(2)}</div>
                <div class="focus-sub">S-UAH</div>
            </div>
            
            <div class="focus-stat-card">
                <div class="focus-stat-label">Знайдено Блоків</div>
                <div id="f-blocks" class="focus-stat-value warning">{$stats.session_blocks}</div>
                <div class="focus-sub">За цю сесію</div>
            </div>
            
            <div class="focus-stat-card">
                <div class="focus-stat-label">Аптайм</div>
                <div id="f-uptime" class="focus-stat-value cyan">{$stats.uptime}</div>
                <div class="focus-sub">Час роботи</div>
            </div>
        </div>
    </div>
</div>

<style>
    #focus-layer {
        position: fixed;
        inset: 0;
        width: 100vw;
        height: 100vh;
        z-index: 99999;
        background: linear-gradient(135deg, rgba(10, 14, 26, 0.98) 0%, rgba(15, 23, 42, 0.98) 100%);
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        backdrop-filter: blur(20px);
    }

    .focus-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 8vh; 
        width: 100%;
        height: 100%;
    }

    .zen-ring {
        
        width: 50vmin; 
        height: 50vmin;
        max-width: 600px;
        max-height: 600px;
        min-width: 300px;
        min-height: 300px;
        
        border-radius: 50%;
        background: radial-gradient(circle, rgba(99,102,241,0.15) 0%, rgba(99,102,241,0.05) 50%, transparent 70%);
        display: flex; 
        flex-direction: column; 
        align-items: center; 
        justify-content: center;
        border: 2px solid rgba(129, 140, 248, 0.2); 
        position: relative;
        box-shadow: 
            0 0 100px rgba(99, 102, 241, 0.3),
            inset 0 0 60px rgba(99, 102, 241, 0.1);
        animation: focusRingFloat 3s ease-in-out infinite;
    }

    .zen-val {
        font-size: 14vmin; 
        font-weight: 800; 
        font-family: var(--font-mono); 
        line-height: 1;
        background: linear-gradient(135deg, #fff 0%, #a5b4fc 100%);
        -webkit-background-clip: text;
        background-clip: text;
        -webkit-text-fill-color: transparent;
        text-shadow: 0 0 30px rgba(129, 140, 248, 0.5);
    }
    
    .zen-label {
        letter-spacing: 4px; 
        color: #94a3b8; 
        margin-top: 2vh; 
        font-size: 2vmin; 
        font-weight: 600;
    }

    .focus-stats {
        display: flex;
        justify-content: center;
        gap: 5vw;
        width: 80%;
    }

    .focus-stat-card {
        text-align: center;
        background: rgba(255, 255, 255, 0.03);
        padding: 20px 40px;
        border-radius: 20px;
        border: 1px solid rgba(255,255,255,0.05);
    }

    .focus-stat-value {
        font-size: 4vmin;
        font-weight: 800;
        font-family: var(--font-mono);
        margin: 10px 0;
    }

    .focus-stat-label {
        text-transform: uppercase;
        letter-spacing: 2px;
        font-size: 1.5vmin;
        color: #94a3b8;
    }

    .focus-sub {
        font-size: 1.5vmin;
        color: #64748b;
    }
    
    
</style>
