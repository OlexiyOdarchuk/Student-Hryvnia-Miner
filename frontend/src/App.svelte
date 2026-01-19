<script lang="ts">
    import { onMount } from 'svelte';
    import { activeTab, stats, logs, connected } from './stores';
    import { EventsOn } from '../wailsjs/runtime/runtime';
    
    import Auth from './components/Auth.svelte';
    import Sidebar from './components/Sidebar.svelte';
    import Dashboard from './components/Dashboard.svelte';
    import Wallets from './components/Wallets.svelte';
    import Transactions from './components/Transactions.svelte';
    import Statistics from './components/Statistics.svelte';
    import Logs from './components/Logs.svelte';
    import Settings from './components/Settings.svelte';
    import Help from './components/Help.svelte';
    import Modals from './components/Modals.svelte';
    import FocusMode from './components/FocusMode.svelte';
    import Toasts from './components/Toasts.svelte';
    
    let modalType = null;
    let modalWallet = null;
    let showFocus = false;
    let authenticated = false;

    onMount(() => {
        // Listen for stats
        EventsOn("stats", (data) => {
            stats.set(data);
            connected.set(true);
        });

        // Listen for logs
        EventsOn("log", (log) => {
            logs.update(l => {
                const newLogs = [...l, log];
                if (newLogs.length > 100) return newLogs.slice(newLogs.length - 100);
                return newLogs;
            });
        });
        
        // Listeners for UI actions
        document.addEventListener('open-modal', (e: any) => {
            modalType = e.detail.type || e.detail;
            modalWallet = e.detail.wallet || null;
        });
        
        document.addEventListener('toggle-focus', () => {
            showFocus = !showFocus;
        });
        
        // Keyboard shortcuts
        window.addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
                e.preventDefault();
                showFocus = !showFocus;
            }
            if (e.key === 'Escape') {
                if (showFocus) showFocus = false;
                if (modalType) closeModal();
            }
        });
    });
    
    function closeModal() {
        modalType = null;
        modalWallet = null;
    }

    function handleLogin() {
        authenticated = true;
    }
</script>

{#if !authenticated}
    <Toasts />
    <Auth on:login={handleLogin} />
{:else}
    <div class="app">
        <Toasts />
        <Sidebar />
        <main class="main">
            {#if $activeTab === 'dash'}
                <Dashboard />
            {:else if $activeTab === 'wallets'}
                <Wallets />
            {:else if $activeTab === 'transactions'}
                <Transactions />
            {:else if $activeTab === 'stats'}
                <Statistics />
            {:else if $activeTab === 'logs'}
                <Logs />
            {:else if $activeTab === 'settings'}
                <Settings />
            {:else if $activeTab === 'help'}
                <Help />
            {/if}
        </main>
    </div>

    {#if showFocus}
        <FocusMode on:close={() => showFocus = false} />
    {/if}

    {#if modalType}
        <Modals type={modalType} wallet={modalWallet} onClose={closeModal} />
    {/if}
{/if}