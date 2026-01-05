export namespace core {
	
	export class FilterConfig {
	    includePaths: string[];
	    excludePaths: string[];
	    includeNames: string[];
	    excludeNames: string[];
	    // Go type: time
	    newerThan?: any;
	    // Go type: time
	    olderThan?: any;
	    minSize: number;
	    maxSize: number;
	
	    static createFrom(source: any = {}) {
	        return new FilterConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.includePaths = source["includePaths"];
	        this.excludePaths = source["excludePaths"];
	        this.includeNames = source["includeNames"];
	        this.excludeNames = source["excludeNames"];
	        this.newerThan = this.convertValues(source["newerThan"], null);
	        this.olderThan = this.convertValues(source["olderThan"], null);
	        this.minSize = source["minSize"];
	        this.maxSize = source["maxSize"];
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
	export class TaskConfig {
	    sourcePaths: string[];
	    destinationDir: string;
	    filters: FilterConfig;
	    useCompression: boolean;
	    useEncryption: boolean;
	    algorithm: number;
	    password: string;
	    incremental: boolean;
	    watchDebounceMs: number;
	    cronExpr: string;
	    watchPaths: string[];
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    lastBackupPath: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourcePaths = source["sourcePaths"];
	        this.destinationDir = source["destinationDir"];
	        this.filters = this.convertValues(source["filters"], FilterConfig);
	        this.useCompression = source["useCompression"];
	        this.useEncryption = source["useEncryption"];
	        this.algorithm = source["algorithm"];
	        this.password = source["password"];
	        this.incremental = source["incremental"];
	        this.watchDebounceMs = source["watchDebounceMs"];
	        this.cronExpr = source["cronExpr"];
	        this.watchPaths = source["watchPaths"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.lastBackupPath = source["lastBackupPath"];
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
	export class BackupTask {
	    id: string;
	    name: string;
	    type: string;
	    enabled: boolean;
	    config: TaskConfig;
	
	    static createFrom(source: any = {}) {
	        return new BackupTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	        this.config = this.convertValues(source["config"], TaskConfig);
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

export namespace main {
	
	export class BackupConfig {
	    sourcePaths: string[];
	    destinationDir: string;
	    filters: core.FilterConfig;
	    useCompression: boolean;
	    useEncryption: boolean;
	    encryptionAlgorithm: string;
	    encryptionPassword: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourcePaths = source["sourcePaths"];
	        this.destinationDir = source["destinationDir"];
	        this.filters = this.convertValues(source["filters"], core.FilterConfig);
	        this.useCompression = source["useCompression"];
	        this.useEncryption = source["useEncryption"];
	        this.encryptionAlgorithm = source["encryptionAlgorithm"];
	        this.encryptionPassword = source["encryptionPassword"];
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
	export class BackupRecord {
	    ID: number;
	    FileName: string;
	    BackupPath: string;
	    // Go type: time
	    CreatedAt: any;
	    SourcePaths: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.FileName = source["FileName"];
	        this.BackupPath = source["BackupPath"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.SourcePaths = source["SourcePaths"];
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
	export class FileInfo {
	    path: string;
	    name: string;
	    size: number;
	    mode: string;
	    // Go type: time
	    modTime: any;
	    isDir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.mode = source["mode"];
	        this.modTime = this.convertValues(source["modTime"], null);
	        this.isDir = source["isDir"];
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
	export class Profile {
	    id: number;
	    name: string;
	    paths: string;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.paths = source["paths"];
	    }
	}
	export class RestoreConfig {
	    backupFile: string;
	    restoreDir: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new RestoreConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.backupFile = source["backupFile"];
	        this.restoreDir = source["restoreDir"];
	        this.password = source["password"];
	    }
	}

}

