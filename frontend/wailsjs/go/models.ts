export namespace config {
	
	export class AppConfig {
	    miner_id: string;
	    telegram_handle: string;
	    base_url: string;
	    server_port: string;
	    difficulty: number;
	    http_timeout: number;
	    retry_delay_ms: number;
	    balance_freq_s: number;
	    block_check_freq_ms: number;
	    max_retries: number;
	    threads: number;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.miner_id = source["miner_id"];
	        this.telegram_handle = source["telegram_handle"];
	        this.base_url = source["base_url"];
	        this.server_port = source["server_port"];
	        this.difficulty = source["difficulty"];
	        this.http_timeout = source["http_timeout"];
	        this.retry_delay_ms = source["retry_delay_ms"];
	        this.balance_freq_s = source["balance_freq_s"];
	        this.block_check_freq_ms = source["block_check_freq_ms"];
	        this.max_retries = source["max_retries"];
	        this.threads = source["threads"];
	    }
	}

}

export namespace types {
	
	export class WalletStats {
	    server_balance: number;
	    address: string;
	    private_key?: string;
	    name: string;
	    session_mined: number;
	    total_mined: number;
	    working: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WalletStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server_balance = source["server_balance"];
	        this.address = source["address"];
	        this.private_key = source["private_key"];
	        this.name = source["name"];
	        this.session_mined = source["session_mined"];
	        this.total_mined = source["total_mined"];
	        this.working = source["working"];
	    }
	}
	export class LogEntry {
	    id: number;
	    time: string;
	    message: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.time = source["time"];
	        this.message = source["message"];
	        this.type = source["type"];
	    }
	}
	export class DashboardData {
	    new_logs: LogEntry[];
	    wallets: WalletStats[];
	    hashrate: number;
	    total_balance: number;
	    session_blocks: number;
	    lifetime_blocks: number;
	    uptime: string;
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.new_logs = this.convertValues(source["new_logs"], LogEntry);
	        this.wallets = this.convertValues(source["wallets"], WalletStats);
	        this.hashrate = source["hashrate"];
	        this.total_balance = source["total_balance"];
	        this.session_blocks = source["session_blocks"];
	        this.lifetime_blocks = source["lifetime_blocks"];
	        this.uptime = source["uptime"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

