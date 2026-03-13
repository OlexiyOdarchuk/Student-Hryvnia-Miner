import './style.css'
import { mount } from 'svelte'
import App from './App.svelte'
import { notifications } from './stores';

// Global error handler
window.onerror = function(message, source, lineno, colno, error) {
    notifications.error("System Error: " + message);
    console.error(error);
};

window.onunhandledrejection = function(event) {
    // Suppress non-critical chart errors
    if (typeof event.reason === 'string' && event.reason.includes('Chart')) return;
    
    notifications.error("Async Error: " + event.reason);
    console.error(event.reason);
};

const app = mount(App, {
  target: document.getElementById('app')!
})

export default app
