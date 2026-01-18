<div class="content-wrapper">
    <div class="dash-header" style="flex-shrink: 0;">
        <div class="page-title">Інструкція користувача</div>
    </div>

    <div class="glass-card" style="padding: 40px; max-width: 900px; margin: 0 auto; width: 100%; overflow-y: auto;">
        
        <div class="help-section">
            <h3><i class="fas fa-rocket text-primary"></i> Початок роботи</h3>
            <p>Вітаємо в <strong>S-UAH Miner</strong>. Ця програма дозволяє добувати криптовалюту S-UAH, використовуючи потужність вашого процесора. Щоб почати:</p>
            <ol>
                <li>Перейдіть у вкладку <strong>Гаманці</strong>.</li>
                <li>Натисніть кнопку <strong>"Додати гаманець"</strong>.</li>
                <li>Введіть довільну назву (наприклад, "Мій комп'ютер") та адресу гаманця. Якщо у вас немає адреси, згенеруйте її на офіційному сайті.</li>
                <li>Майнінг розпочнеться автоматично.</li>
            </ol>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-wallet text-success"></i> Управління гаманцями</h3>
            <p>У вкладці "Гаманці" ви можете керувати своїми воркерами:</p>
            <ul>
                <li><i class="fas fa-pause text-warning"></i> <strong>Пауза/Старт:</strong> Тимчасово зупинити майнінг на конкретному гаманці.</li>
                <li><i class="fas fa-key text-primary"></i> <strong>Ключ:</strong> Додати приватний ключ для можливості здійснювати транзакції з цього гаманця.</li>
                <li><i class="fas fa-pen"></i> <strong>Редагувати:</strong> Змінити назву воркера.</li>
                <li><i class="fas fa-trash text-danger"></i> <strong>Видалити:</strong> Видалити гаманець зі списку.</li>
            </ul>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-exchange-alt text-cyan"></i> Транзакції</h3>
            <p>Для відправки коштів:</p>
            <ol>
                <li>Перейдіть у вкладку <strong>Транзакції</strong>.</li>
                <li>Оберіть гаманець, з якого хочете надіслати кошти (на ньому має бути доданий приватний ключ).</li>
                <li>Введіть адресу отримувача та суму.</li>
                <li>Введіть ваш <strong>пароль адміністратора</strong> для підтвердження.</li>
            </ol>
            <p class="note"><strong>Примітка:</strong> Транзакції підписуються локально. Ваш приватний ключ зберігається в зашифрованому вигляді.</p>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-globe text-accent"></i> Веб-моніторинг</h3>
            <p>Ви можете слідкувати за статусом майнінгу з телефону або іншого пристрою в тій же мережі:</p>
            <ol>
                <li>Натисніть кнопку <strong>Веб-моніторинг</strong> в меню зліва.</li>
                <li>Відкриється браузер з адресою (зазвичай <code>http://localhost:8080</code>).</li>
                <li>Щоб відкрити з телефону, дізнайтеся IP-адресу комп'ютера (наприклад, <code>192.168.1.5:8080</code>).</li>
            </ol>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-cog text-warning"></i> Безпека</h3>
            <p>Всі ваші дані зберігаються у файлі <code>storage.bin</code>, зашифрованому алгоритмом AES-GCM. Пароль, який ви створили при першому вході, є ключем до цього файлу. <strong>Не загубіть його!</strong> Відновити доступ без пароля неможливо.</p>
        </div>

    </div>
</div>

<style>
    .help-section {
        margin-bottom: 40px;
    }
    
    h3 {
        font-size: 1.4rem;
        margin-bottom: 15px;
        color: white;
        display: flex;
        align-items: center;
        gap: 10px;
    }
    
    p, li {
        color: #cbd5e1;
        line-height: 1.6;
        font-size: 1rem;
    }
    
    ol, ul {
        margin-left: 20px;
        margin-top: 10px;
    }
    
    li {
        margin-bottom: 8px;
    }
    
    .note {
        background: rgba(129, 140, 248, 0.1);
        padding: 15px;
        border-radius: 10px;
        border-left: 3px solid var(--primary);
        margin-top: 15px;
        font-size: 0.9rem;
    }
    
    .text-primary { color: var(--primary); }
    .text-success { color: var(--success); }
    .text-warning { color: var(--warning); }
    .text-danger { color: var(--danger); }
    .text-cyan { color: var(--neon-cyan); }
    .text-accent { color: var(--accent); }
</style>
