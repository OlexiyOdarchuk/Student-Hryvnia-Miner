<script lang="ts">
    import { stats, notifications } from "../stores";
    import { ec as EC } from "elliptic";
    import { GetWalletKey } from "../../wailsjs/go/main/App";

    let selectedAddress: string = "";
    let toAddress: string = "";
    let amount: number | null = null; // Починаємо з порожнього
    let password: string = "";
    let sending: boolean = false;

    const ec = new EC("secp256k1");

    async function sha256(message: string): Promise<string> {
        const msgBuffer = new TextEncoder().encode(message);
        const hashBuffer = await crypto.subtle.digest("SHA-256", msgBuffer);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        return hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
    }

    function selectWallet(wallet: any) {
        selectedAddress = wallet.address;
    }

    function setMax() {
        if (!selectedAddress) return;
        const wallet = ($stats.wallets || []).find(
            (w: any) => w.address === selectedAddress,
        );
        if (wallet) {
            amount = wallet.server_balance;
        }
    }

    async function handleSend() {
        // Валідація
        if (!selectedAddress) {
            notifications.error("Оберіть гаманець для списання коштів");
            return;
        }
        if (!toAddress || toAddress.length < 20) {
            notifications.error("Введіть коректну адресу отримувача");
            return;
        }
        if (!amount || amount <= 0) {
            notifications.error("Сума повинна бути більшою за 0");
            return;
        }
        if (!password) {
            notifications.error("Введіть пароль для підпису");
            return;
        }

        sending = true;

        try {
            const privateKey = await GetWalletKey(selectedAddress, password);

            if (!privateKey) {
                throw new Error("Невірний пароль або ключ не знайдено");
            }

            // КРОК 2: Готуємо об'єкт (Важливо: amount як число!)
            const txObject = {
                from: selectedAddress,
                to: toAddress.trim(),
                amount: Number(amount),
                fee: 0,
            };

            // КРОК 3: Хешуємо JSON
            const jsonString = JSON.stringify(txObject);
            const hashHex = await sha256(jsonString);

            // КРОК 4: Підписуємо
            const keyPair = ec.keyFromPrivate(privateKey);
            const signature = keyPair.sign(hashHex, "hex").toDER("hex");

            // КРОК 5: Відправляємо
            const payload = { ...txObject, signature };

            const response = await fetch(
                "https://s-hryvnia-1.onrender.com/transaction",
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(payload),
                },
            );

            // Обробка відповіді (іноді текст, іноді JSON)
            const textResp = await response.text();
            let result: any = {};
            try {
                result = JSON.parse(textResp);
            } catch (e) {}

            if (response.ok) {
                notifications.success("Транзакція успішно відправлена! 🚀");
                // Очистка
                toAddress = "";
                amount = null;
                password = "";
            } else {
                const errorMsg = result.message || result.error || textResp;
                throw new Error(errorMsg || "Сервер відхилив транзакцію");
            }
        } catch (e: any) {
            console.error(e);
            notifications.error("Помилка: " + e.message || e);
        } finally {
            sending = false;
        }
    }
</script>

<div class="content-wrapper">
    <div class="dash-header">
        <div class="page-title">Відправити транзакцію</div>
    </div>

    <div class="tx-layout">
        <div class="glass-card tx-form-card">
            <h3 class="card-title">
                <i class="fas fa-paper-plane text-primary"></i>
                Деталі платежу
            </h3>

            <form on:submit|preventDefault={handleSend}>
                <div class="field-group">
                    <label class="field-label"
                        ><i class="fas fa-wallet"></i> Гаманець відправника</label
                    >
                    <div class="wallet-select-grid">
                        {#each $stats.wallets || [] as wallet}
                            <div
                                class="wallet-select-card"
                                class:selected={selectedAddress ===
                                    wallet.address}
                                on:click={() => selectWallet(wallet)}
                                role="button"
                                tabindex="0"
                            >
                                <div class="wallet-header">
                                    <div class="wallet-icon">
                                        <i class="fas fa-wallet"></i>
                                    </div>
                                    <div class="wallet-name">{wallet.name}</div>
                                </div>
                                <div class="wallet-balance">
                                    {wallet.server_balance.toFixed(2)} S-UAH
                                </div>
                                <div class="wallet-addr">
                                    {wallet.address.substring(0, 10)}...
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>

                <div class="field-group">
                    <label class="field-label"
                        ><i class="fas fa-map-marker-alt"></i> Адреса отримувача</label
                    >
                    <input
                        type="text"
                        class="field"
                        placeholder="hash гаманця"
                        bind:value={toAddress}
                    />
                </div>

                <div class="field-group">
                    <label class="field-label"
                        ><i class="fas fa-coins"></i> Сума (S-UAH)</label
                    >
                    <div class="input-wrapper">
                        <input
                            type="number"
                            class="field no-spin"
                            placeholder="0.00"
                            step="any"
                            min="0"
                            bind:value={amount}
                        />
                        <button type="button" class="btn-max" on:click={setMax}
                            >МАКС</button
                        >
                    </div>
                </div>

                <div class="field-group">
                    <label class="field-label"
                        ><i class="fas fa-lock"></i> Пароль адміністратора</label
                    >
                    <input
                        type="password"
                        class="field"
                        placeholder="Для безпеки переказів"
                        bind:value={password}
                    />
                </div>

                <button
                    type="submit"
                    class="btn btn-primary btn-xl submit-btn"
                    disabled={sending}
                >
                    {#if sending}
                        <i class="fas fa-circle-notch fa-spin"></i> Обробка...
                    {:else}
                        <i class="fas fa-paper-plane"></i> Надіслати кошти
                    {/if}
                </button>
            </form>
        </div>

        <div class="glass-card tx-sidebar">
            <h3 class="card-title">
                <i class="fas fa-rocket text-primary"></i> Поради
            </h3>

            <div class="tips-container">
                <div class="tip-card tip-primary">
                    <div class="tip-icon"><i class="fas fa-lightbulb"></i></div>
                    <div>
                        <div class="tip-title">Швидка вставка</div>
                        <div class="tip-text">
                            Використовуйте кнопку <b>МАКС</b>, щоб швидко
                            вибрати весь доступний баланс.
                        </div>
                    </div>
                </div>
                <div class="tip-card tip-success">
                    <div class="tip-icon">
                        <i class="fas fa-shield-alt"></i>
                    </div>
                    <div>
                        <div class="tip-title">Безпека</div>
                        <div class="tip-text">
                            Транзакції підписуються локально. Ваш ключ ніколи не
                            залишає цей пристрій.
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<style>
    /* LAYOUT */
    .content-wrapper {
        display: flex;
        flex-direction: column;
        height: 100%;
    }
    .dash-header {
        padding: 0 0 20px 0;
    }
    .page-title {
        font-size: 1.8rem;
        font-weight: 700;
    }

    .tx-layout {
        display: grid;
        grid-template-columns: 2fr 1fr;
        gap: 30px;
        align-items: start;
        padding-bottom: 40px;
    }

    /* CARDS */
    .glass-card {
        background: rgba(30, 41, 59, 0.6);
        backdrop-filter: blur(20px);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 20px;
        padding: 30px;
    }
    .card-title {
        margin-bottom: 25px;
        font-size: 1.4rem;
        display: flex;
        align-items: center;
        gap: 10px;
    }
    .text-primary {
        color: #818cf8;
    }

    /* FORM FIELDS */
    .field-group {
        margin-bottom: 20px;
    }
    .field-label {
        display: block;
        margin-bottom: 8px;
        color: #94a3b8;
        font-size: 0.9rem;
        font-weight: 600;
    }
    .field-label i {
        width: 20px;
        text-align: center;
        margin-right: 5px;
    }

    .field {
        width: 100%;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 12px;
        padding: 14px 16px;
        color: white;
        font-size: 1rem;
        outline: none;
        transition: 0.3s;
    }
    .field:focus {
        border-color: #818cf8;
        box-shadow: 0 0 15px rgba(129, 140, 248, 0.15);
    }

    /* INPUT WITH BUTTON (FIXED) */
    .input-wrapper {
        position: relative;
        display: flex;
        align-items: center;
    }
    .input-wrapper .field {
        padding-right: 80px;
    } /* Space for button */

    .btn-max {
        position: absolute;
        right: 8px;
        top: 35%;
        transform: translateY(-50%);
        background: rgba(129, 140, 248, 0.15);
        color: #818cf8;
        border: 1px solid rgba(129, 140, 248, 0.3);
        border-radius: 8px;
        padding: 6px 12px;
        font-size: 0.75rem;
        font-weight: 700;
        cursor: pointer;
        transition: 0.2s;
    }
    .btn-max:hover {
        background: #818cf8;
        color: white;
    }

    /* WALLET SELECTION */
    .wallet-select-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
        gap: 15px;
    }
    .wallet-select-card {
        background: rgba(255, 255, 255, 0.03);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 12px;
        padding: 15px;
        cursor: pointer;
        transition: 0.2s;
    }
    .wallet-select-card:hover {
        background: rgba(255, 255, 255, 0.06);
    }
    .wallet-select-card.selected {
        border-color: #818cf8;
        background: rgba(129, 140, 248, 0.1);
        box-shadow: 0 4px 20px rgba(129, 140, 248, 0.15);
    }
    .wallet-header {
        display: flex;
        align-items: center;
        gap: 10px;
        margin-bottom: 10px;
    }
    .wallet-icon {
        width: 32px;
        height: 32px;
        border-radius: 8px;
        background: rgba(129, 140, 248, 0.2);
        color: #818cf8;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    .wallet-balance {
        font-size: 1.2rem;
        font-weight: 700;
        color: #34d399;
        font-family: monospace;
    }
    .wallet-addr {
        font-size: 0.8rem;
        color: #64748b;
        margin-top: 5px;
    }

    /* BUTTONS */
    .btn {
        border: none;
        border-radius: 12px;
        font-weight: 600;
        cursor: pointer;
        transition: 0.2s;
    }
    .btn-primary {
        background: linear-gradient(135deg, #818cf8, #4f46e5);
        color: white;
        box-shadow: 0 4px 15px rgba(79, 70, 229, 0.4);
    }
    .btn-primary:hover {
        transform: translateY(-2px);
        box-shadow: 0 6px 20px rgba(79, 70, 229, 0.5);
    }
    .btn-primary:disabled {
        opacity: 0.7;
        transform: none;
        cursor: not-allowed;
    }
    .btn-xl {
        padding: 16px;
        font-size: 1rem;
    }
    .submit-btn {
        width: 100%;
        margin-top: 10px;
    }

    /* TIPS */
    .tip-card {
        padding: 15px;
        border-radius: 12px;
        margin-bottom: 15px;
        display: flex;
        gap: 15px;
        align-items: flex-start;
    }
    .tip-primary {
        background: rgba(129, 140, 248, 0.1);
        border-left: 3px solid #818cf8;
    }
    .tip-success {
        background: rgba(52, 211, 153, 0.1);
        border-left: 3px solid #34d399;
    }
    .tip-icon {
        font-size: 1.2rem;
        padding-top: 2px;
    }
    .tip-primary .tip-icon {
        color: #fbbf24;
    }
    .tip-success .tip-icon {
        color: #34d399;
    }
    .tip-title {
        font-weight: 700;
        margin-bottom: 4px;
        color: white;
        font-size: 0.95rem;
    }
    .tip-text {
        font-size: 0.85rem;
        color: #cbd5e1;
        line-height: 1.5;
    }

    /* UTILS */
    .no-spin::-webkit-outer-spin-button,
    .no-spin::-webkit-inner-spin-button {
        -webkit-appearance: none;
        margin: 0;
    }

    @media (max-width: 1024px) {
        .tx-layout {
            grid-template-columns: 1fr;
        }
        .tx-sidebar {
            display: none;
        } /* Hide tips on mobile to save space */
    }
</style>
