import { writable } from 'svelte/store';

export const activeTab = writable('dash');
export const stats = writable({
    hashrate: 0,
    session_blocks: 0,
    lifetime_blocks: 0,
    uptime: "00:00:00",
    total_balance: 0,
    wallets: [],
    new_logs: []
});
export const logs = writable([]);
export const connected = writable(false);

function createNotificationStore() {
    const { subscribe, update } = writable([]);

    return {
        subscribe,
        show: (message, type = 'info') => {
            const id = Date.now();
            update(n => [...n, { id, message, type }]);
            setTimeout(() => {
                update(n => n.filter(t => t.id !== id));
            }, 3000);
        },
        success: (msg) => notifications.show(msg, 'success'),
        error: (msg) => notifications.show(msg, 'error'),
        info: (msg) => notifications.show(msg, 'info')
    };
}

export const notifications = createNotificationStore();