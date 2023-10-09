export namespace model {
	
	export class EventData {
	    status: string;
	    data: any;
	
	    static createFrom(source: any = {}) {
	        return new EventData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.data = source["data"];
	    }
	}
	export class Feature {
	    id: string;
	    name: string;
	    icon: string;
	    publicKey: string;
	    md5: string;
	    platform: string;
	
	    static createFrom(source: any = {}) {
	        return new Feature(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.icon = source["icon"];
	        this.publicKey = source["publicKey"];
	        this.md5 = source["md5"];
	        this.platform = source["platform"];
	    }
	}

}

