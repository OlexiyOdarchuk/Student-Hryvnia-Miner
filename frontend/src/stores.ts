import { writable } from 'svelte/store';
import { backend } from '../wailsjs/go/models';

export const activeTab = writable<string>('dash');

export const stats = writable<backend.DashboardData>({
    hashrate: 0,
    session_blocks: 0,
    lifetime_blocks: 0,
    uptime: "00:00:00",
    total_balance: 0,
    wallets: [],
    new_logs: []
} as backend.DashboardData);

export const logs = writable<backend.LogEntry[]>([]);
export const connected = writable<boolean>(false);

export interface Notification {
    id: number;
    message: string;
    type: 'success' | 'error' | 'info';
}

function createNotificationStore() {
    const { subscribe, update } = writable<Notification[]>([]);

    return {
        subscribe,
        show: (message: string, type: 'success' | 'error' | 'info' = 'info') => {
            const id = Date.now();
            update(n => [...n, { id, message, type }]);
            setTimeout(() => {
                update(n => n.filter(t => t.id !== id));
            }, 3000);
        },
        success: (msg: string) => notifications.show(msg, 'success'),
        error: (msg: string) => notifications.show(msg, 'error'),
        info: (msg: string) => notifications.show(msg, 'info')
    };
}

export const notifications = createNotificationStore();