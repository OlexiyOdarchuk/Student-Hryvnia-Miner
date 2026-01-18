export namespace backend {
	
	export class AppConfig {
	    base_url: string;
	    server_port: string;
	    difficulty: number;
	    http_timeout: number;
	    max_retries: number;
	    retry_delay_ms: number;
	    balance_freq_s: number;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.base_url = source["base_url"];
	        this.server_port = source["server_port"];
	        this.difficulty = source["difficulty"];
	        this.http_timeout = source["http_timeout"];
	        this.max_retries = source["max_retries"];
	        this.retry_delay_ms = source["retry_delay_ms"];
	        this.balance_freq_s = source["balance_freq_s"];
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
	export class WalletStats {
	    address: string;
	    private_key?: string;
	    name: string;
	    short: string;
	    session_mined: number;
	    total_mined: number;
	    server_balance: number;
	    status: string;
	    working: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WalletStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.private_key = source["private_key"];
	        this.name = source["name"];
	        this.short = source["short"];
	        this.session_mined = source["session_mined"];
	        this.total_mined = source["total_mined"];
	        this.server_balance = source["server_balance"];
	        this.status = source["status"];
	        this.working = source["working"];
	    }
	}
	export class DashboardData {
	    hashrate: number;
	    session_blocks: number;
	    lifetime_blocks: number;
	    uptime: string;
	    total_balance: number;
	    wallets: WalletStats[];
	    new_logs: LogEntry[];
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hashrate = source["hashrate"];
	        this.session_blocks = source["session_blocks"];
	        this.lifetime_blocks = source["lifetime_blocks"];
	        this.uptime = source["uptime"];
	        this.total_balance = source["total_balance"];
	        this.wallets = this.convertValues(source["wallets"], WalletStats);
	        this.new_logs = this.convertValues(source["new_logs"], LogEntry);
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

