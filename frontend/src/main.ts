import './style.css'
import { mount } from 'svelte'
import App from './App.svelte'
import { notifications } from './stores';

// Global error handler — surface a generic message to the user and keep the
// full detail in the console (avoids leaking raw backend error internals).
window.onerror = function(message, source, lineno, colno, error) {
    notifications.error("Сталася системна помилка");
    console.error(error ?? message);
};

window.onunhandledrejection = function(event) {
    // Suppress non-critical chart errors
    if (typeof event.reason === 'string' && event.reason.includes('Chart')) return;

    notifications.error("Сталася непередбачена помилка");
    console.error(event.reason);
};

const app = mount(App, {
  target: document.getElementById('app')!
})

export default app
