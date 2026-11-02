export namespace main {
	
	export class AppConfig {
	    host: string;
	    port: string;
	    proxyAddr: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.proxyAddr = source["proxyAddr"];
	    }
	}

}

