<script>
    import { WindowMinimise, WindowToggleMaximise, Quit } from '../../wailsjs/runtime/runtime';
</script>


<div class="title-bar" on:dblclick={WindowToggleMaximise}>
    <div class="app-title">
        <i class="fas fa-cube"></i> S-UAH MINER
    </div>
    <div class="window-controls">
        <button class="control-btn minimize" on:click={WindowMinimise}>
            <i class="fas fa-minus"></i>
        </button>
        <button class="control-btn maximize" on:click={WindowToggleMaximise}>
            <i class="fas fa-expand"></i>
        </button>
        <button class="control-btn close" on:click={Quit}>
            <i class="fas fa-times"></i>
        </button>
    </div>
</div>

<style>
    .title-bar {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        height: 32px;
        background: rgba(15, 23, 42, 0.8);
        backdrop-filter: blur(10px);
        display: flex;
        justify-content: space-between;
        align-items: center;
        z-index: 9999;
        --wails-draggable: drag;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    }

    .app-title {
        padding-left: 15px;
        font-size: 0.8rem;
        font-weight: 700;
        color: #94a3b8;
        display: flex;
        align-items: center;
        gap: 8px;
        letter-spacing: 0.5px;
    }

    .window-controls {
        display: flex;
        height: 100%;
        --wails-draggable: no-drag; 
    }

    .control-btn {
        background: transparent;
        border: none;
        color: #94a3b8;
        width: 46px;
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        transition: all 0.2s;
        font-size: 0.85rem;
    }

    .control-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        color: white;
    }

    .control-btn.close:hover {
        background: #ef4444;
        color: white;
    }
</style>