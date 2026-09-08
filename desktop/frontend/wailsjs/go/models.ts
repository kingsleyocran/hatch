export namespace main {
	
	export class ConfigInfo {
	    default_tld: string;
	    auto_https: boolean;
	    daemon_port: number;
	    https_port: number;
	    dns_port: number;
	
	    static createFrom(source: any = {}) {
	        return new ConfigInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.default_tld = source["default_tld"];
	        this.auto_https = source["auto_https"];
	        this.daemon_port = source["daemon_port"];
	        this.https_port = source["https_port"];
	        this.dns_port = source["dns_port"];
	    }
	}
	export class DomainInfo {
	    domain: string;
	    port: number;
	    dir: string;
	    alive: boolean;
	    https: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DomainInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.port = source["port"];
	        this.dir = source["dir"];
	        this.alive = source["alive"];
	        this.https = source["https"];
	    }
	}
	export class PortInfo {
	    port: number;
	    name: string;
	    process: string;
	    type: string;
	    dir: string;
	
	    static createFrom(source: any = {}) {
	        return new PortInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.name = source["name"];
	        this.process = source["process"];
	        this.type = source["type"];
	        this.dir = source["dir"];
	    }
	}
	export class StatusInfo {
	    running: boolean;
	    uptime: string;
	    domain_count: number;
	    active_count: number;
	
	    static createFrom(source: any = {}) {
	        return new StatusInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.uptime = source["uptime"];
	        this.domain_count = source["domain_count"];
	        this.active_count = source["active_count"];
	    }
	}

}

