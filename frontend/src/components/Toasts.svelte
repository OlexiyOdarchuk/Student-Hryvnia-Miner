<script lang="ts">
    import { notifications } from '../stores';
    import { fly } from 'svelte/transition';
</script>

<div class="toast-container">
    {#each $notifications as toast (toast.id)}
        <div class="toast" class:error={toast.type === 'error'} transition:fly={{ x: 200, duration: 300 }}>
            <i class="fas {toast.type === 'error' ? 'fa-exclamation-circle' : toast.type === 'success' ? 'fa-check-circle' : 'fa-info-circle'}"></i>
            <span>{toast.message}</span>
        </div>
    {/each}
</div>
