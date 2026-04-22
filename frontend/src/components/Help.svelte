<div class="content-wrapper">
    <div class="dash-header" style="flex-shrink: 0;">
        <div class="page-title">Інструкція користувача</div>
    </div>

    <div class="glass-card" style="padding: 40px; max-width: 900px; margin: 0 auto; width: 100%; overflow-y: auto;">

        <div class="help-section">
            <h3><i class="fas fa-rocket text-primary"></i> Початок роботи</h3>
            <p>Вітаємо в <strong>S-UAH Miner</strong>. Ця програма дозволяє добувати криптовалюту S-UAH, використовуючи потужність вашого процесора. Щоб почати:</p>
            <ol>
                <li>При першому запуску встановіть пароль адміністратора, <strong>або залиште поле пустим</strong> — тоді сховище буде відкриватися автоматично.</li>
                <li>Перейдіть у вкладку <strong>Гаманці</strong>.</li>
                <li>Натисніть кнопку <strong>"Додати гаманець"</strong>.</li>
                <li>Виберіть <strong>"Створити новий"</strong> для автоматичної генерації ключів або <strong>"Імпортувати"</strong> (вручну чи JSON), якщо у вас вже є гаманець.</li>
                <li>Майнінг розпочнеться автоматично. Дивитися статус можна на <strong>Головній</strong>.</li>
            </ol>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-gauge-high text-primary"></i> Головна панель</h3>
            <p>На вкладці <strong>Головна</strong> відображаються ключові метрики майнінгу в реальному часі:</p>
            <ul>
                <li><strong>Хешрейт мережі</strong> — середня швидкість обчислень у MH/s.</li>
                <li><strong>Загальний баланс</strong> — сума балансів усіх ваших гаманців на сервері.</li>
                <li><strong>Аптайм</strong> — скільки часу нода працює з моменту запуску.</li>
                <li><strong>Зараховано</strong> — блоки, підтверджені сервером у цій сесії, та поточна швидкість у хвилину.</li>
                <li><strong>Намайнено</strong> — блоки, які ваш процесор знайшов (до відправки), та поточна швидкість у хвилину.</li>
                <li><strong>У черзі</strong> — скільки блоків зараз чекають у буфері відправки на сервер.</li>
                <li><strong>Активні гаманці</strong> — скільки воркерів зараз у стані майнінгу (з усіх доданих).</li>
            </ul>
            <p>Поруч є <strong>Журнал подій</strong> з фільтрами: показати всі записи, лише зараховані блоки або лише помилки. Кнопка <i class="fas fa-expand"></i> у верхньому правому куті графіка відкриває <strong>Режим фокусу</strong> (повний екран з основними метриками).</p>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-wallet text-success"></i> Управління гаманцями</h3>
            <p>У вкладці <strong>Гаманці</strong> кожен рядок — це окремий воркер:</p>
            <ul>
                <li><i class="fas fa-pause text-warning"></i> <strong>Пауза / Старт:</strong> призупинити або відновити майнінг на конкретному гаманці.</li>
                <li><i class="fas fa-key text-primary"></i> <strong>Ключі:</strong>
                    <ul>
                        <li>Скопіювати приватний ключ у буфер обміну.</li>
                        <li>Експортувати гаманець у JSON (для перенесення на інший пристрій).</li>
                        <li>Оновити приватний ключ для гаманця, доданого без нього.</li>
                    </ul>
                </li>
                <li><i class="fas fa-pen"></i> <strong>Редагувати:</strong> змінити назву воркера.</li>
                <li><i class="fas fa-trash text-danger"></i> <strong>Видалити:</strong> прибрати гаманець зі списку.</li>
            </ul>
            <p class="note">Додати гаманець можна трьома способами: згенерувати нову пару ключів (найшвидше), ввести адресу + ключ вручну, або імпортувати JSON-файл, експортований раніше. Якщо ви вкажете лише адресу без приватного ключа, гаманець буде тільки для перегляду — з нього не можна буде надсилати транзакції.</p>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-exchange-alt text-cyan"></i> Транзакції</h3>
            <p>Щоб надіслати кошти:</p>
            <ol>
                <li>Перейдіть у вкладку <strong>Транзакції</strong>.</li>
                <li>Оберіть гаманець-відправник. <strong>У списку з'являться лише ті гаманці, для яких додано приватний ключ</strong> (інакше підписати транзакцію неможливо).</li>
                <li>Введіть адресу отримувача та суму (кнопка <strong>МАКС</strong> підставляє весь доступний баланс).</li>
                <li>Якщо у вас встановлений пароль адміністратора — введіть його для підтвердження. <strong>Якщо паролю немає, поле не з'явиться.</strong></li>
            </ol>
            <p class="note">Транзакції підписуються <strong>локально</strong> через ECDSA (secp256k1). Ваш приватний ключ не залишає пристрій — на сервер передається лише підписаний об'єкт транзакції.</p>
        </div>

        <div class="help-section">
            <h3><i class="fab fa-telegram text-cyan"></i> Автоматизація та Telegram-бот</h3>
            <p>У вкладці <strong>Автоматизація</strong> налаштовується бот, який керує майнінгом за вас:</p>
            <ol>
                <li>Створіть власного бота через <a href="https://t.me/BotFather" target="_blank">@BotFather</a> і вставте <strong>Bot Token</strong>.</li>
                <li>Напишіть <strong>/start</strong> своєму боту зі свого акаунта, щоб отримати <strong>Chat ID</strong>. Якщо маєте публічний канал — можна вставити <code>@username</code> і натиснути <i class="fas fa-magnifying-glass"></i>, щоб розпізнати ID автоматично.</li>
                <li>Натисніть <strong>Тест</strong> — має прийти повідомлення у Telegram.</li>
            </ol>
            <p>Після збереження налаштувань та запуску майнінгу бот показує у Telegram <strong>постійну клавіатуру</strong> з кнопками:</p>
            <ul>
                <li><strong>📊 Статус</strong> (або <code>/status</code>) — статистика сесії: зараховано, намайнено, швидкості, черга, активні гаманці.</li>
                <li><strong>💳 Гаманці</strong> (або <code>/wallets</code>) — список гаманців; для кожного рядка доступні інлайн-дії: клік по назві — увімкнути/вимкнути воркер, <strong>✏ Rename</strong> — перейменувати, <strong>🗑 Delete</strong> — видалити (з підтвердженням), <strong>📤 Export</strong> — вивантажити JSON з приватним ключем прямо в чат.</li>
                <li><strong>▶️ / ⏸ Майнінг</strong> (або <code>/mine</code> / <code>/pause</code>) — перемикає глобальний стан майнінгу. Стан на Головній вкладці програми змінюється синхронно.</li>
                <li><strong>➕ Гаманець</strong> (або <code>/addwallet</code>) — покрокове додавання гаманця з вибору: <em>створити новий</em> (бот згенерує ключі й поверне адресу), <em>імпортувати JSON</em> (формат, як при експорті з програми) або <em>спостережний</em> (тільки адреса).</li>
                <li><strong>⚙️ Налаштування</strong> (або <code>/settings</code>) — редагувати правила автоматизації прямо з Telegram: сповіщення, розклад, ціль по блоках, таймер сесії, крок прогресу. Зміни зберігаються у той самий файл, що й UI-вкладка «Автоматизація».</li>
                <li><strong>🔄 Reset</strong> (або <code>/reset</code>) — скидає сесійні лічильники (зараховано/намайнено) і час сесії, не зупиняючи майнінг.</li>
            </ul>
            <p>Правила автоматизації:</p>
            <ul>
                <li><strong>Ціль по блоках</strong> — коли зараховано вказану кількість блоків, майнінг зупиняється автоматично.</li>
                <li><strong>Таймер сесії</strong> — зупинити через N хвилин після старту.</li>
                <li><strong>Розклад</strong> — автоматично вмикати і вимикати майнінг у заданому вікні (підтримується перехід через північ).</li>
                <li><strong>Повідомляти про прогрес</strong> — без зупинки майнінгу надсилати повідомлення кожного разу, коли сесійний лічильник зарахованих блоків перетинає чергове кратне число (напр. 200, 1000). <code>0</code> вимикає.</li>
                <li><strong>Сповіщення</strong> — приходять у Telegram для стартів, зупинок і досягнення цілі.</li>
            </ul>
            <p class="note">Бот відповідає тільки на повідомлення з акаунта, чий Chat ID збережено — ніхто сторонній не зможе керувати вашим майнером.</p>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-globe text-accent"></i> Веб-моніторинг</h3>
            <p>Ви можете слідкувати за статусом майнінгу з телефону або іншого пристрою в тій же локальній мережі:</p>
            <ol>
                <li>Натисніть кнопку <strong>Веб-моніторинг</strong> у боковому меню.</li>
                <li>Відкриється сторінка з LAN-адресою вашого комп'ютера (вона автоматично копіюється в буфер обміну).</li>
                <li>Відкрийте цю адресу в браузері на іншому пристрої — побачите той самий дашборд.</li>
            </ol>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-tachometer-alt text-warning"></i> Продуктивність</h3>
            <p>У вкладці <strong>Налаштування</strong> можна керувати навантаженням:</p>
            <ul>
                <li><strong>Потоки CPU:</strong> кількість ядер, які задіяні для майнінгу.</li>
                <li><strong>AUTO (0):</strong> використовуються всі доступні ядра для максимальної швидкості.</li>
                <li><strong>Ручний режим:</strong> зменшіть кількість потоків, якщо ПК почав гальмувати під час інших задач.</li>
                <li><strong>Складність</strong> і <strong>Синхр. блоків</strong> — тонкі налаштування мережевої взаємодії. Без потреби не чіпайте.</li>
                <li><strong>Розмір буфера submit-черги</strong> — скільки знайдених блоків можуть одночасно чекати відправки. Більший буфер згладжує сплески (майнер не блокується, поки попередні блоки ще відправляються), але витрачає більше пам'яті. <code>0</code> — значення за замовчуванням. Нове значення застосовується після рестарту майнінгу (або при перемиканні в UI — автоматично).</li>
            </ul>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-shield-halved text-warning"></i> Пароль і безпека</h3>
            <p>Усі дані зберігаються у файлі <code>SHMinerSettings.bin</code>, зашифрованому алгоритмом <strong>AES-GCM</strong> (ключ — Argon2id від паролю). Шлях до файлу та кнопка «Відкрити папку» доступні у вкладці <strong>Налаштування</strong>.</p>
            <ul>
                <li><strong>Пароль — необов'язковий.</strong> Якщо при першому запуску ви залишили поле пустим, сховище шифрується порожнім паролем і відкривається автоматично. Програма нічого додатково не питатиме при видаленні гаманця, зміні ключа, експорті чи транзакції.</li>
                <li>У вкладці <strong>Налаштування → Пароль адміністратора</strong> ви завжди можете встановити пароль (якщо зараз його немає) або змінити його.</li>
                <li><strong>Не загубіть пароль!</strong> Відновити доступ без нього неможливо — потрібно буде імпортувати гаманці заново з раніше збереженого JSON.</li>
            </ul>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-shopping-cart text-accent"></i> Магазин та покупки</h3>
            <p>Ця програма сфокусована на швидкому майнінгу. Для купівлі курсів та інших товарів за зароблені S-UAH використовуйте офіційний бот:</p>
            <ul>
                <li><i class="fab fa-telegram"></i> <strong>Магазин:</strong> <a href="https://t.me/s_hryvnia_bot" target="_blank">@s_hryvnia_bot</a></li>
            </ul>
        </div>

        <div class="help-section">
            <h3><i class="fas fa-link text-primary"></i> Корисні посилання</h3>
            <p>Зв'язатися з розробником і переглянути вихідний код:</p>
            <ul>
                <li><i class="fab fa-telegram text-cyan"></i> <strong>Telegram:</strong> <a href="https://t.me/NeShawyha">@NeShawyha</a></li>
                <li><i class="fab fa-github"></i> <strong>GitHub:</strong> <a href="https://github.com/OlexiyOdarchuk">OlexiyOdarchuk</a></li>
            </ul>
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

    a {
        color: var(--primary);
        text-decoration: none;
        font-weight: 600;
        transition: all 0.3s ease;
        border-bottom: 1px solid transparent;
    }

    a:hover {
        color: var(--neon-cyan);
        text-shadow: 0 0 10px rgba(6, 182, 212, 0.5);
        border-bottom-color: var(--neon-cyan);
    }
</style>
