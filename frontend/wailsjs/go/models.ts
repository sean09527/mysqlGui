export namespace backend {
	
	export class FileDialogFilter {
	    displayName: string;
	    pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new FileDialogFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.displayName = source["displayName"];
	        this.pattern = source["pattern"];
	    }
	}

}

export namespace importexport {
	
	export class ColumnMapping {
	    FileColumns: string[];
	    TableColumns: string[];
	
	    static createFrom(source: any = {}) {
	        return new ColumnMapping(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.FileColumns = source["FileColumns"];
	        this.TableColumns = source["TableColumns"];
	    }
	}
	export class ImportError {
	    Row: number;
	    Message: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Row = source["Row"];
	        this.Message = source["Message"];
	    }
	}
	export class ImportResult {
	    TotalRows: number;
	    SuccessRows: number;
	    FailedRows: number;
	    Errors: ImportError[];
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TotalRows = source["TotalRows"];
	        this.SuccessRows = source["SuccessRows"];
	        this.FailedRows = source["FailedRows"];
	        this.Errors = this.convertValues(source["Errors"], ImportError);
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

export namespace query {
	
	export class QueryError {
	    code: number;
	    message: string;
	    position: number;
	
	    static createFrom(source: any = {}) {
	        return new QueryError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.position = source["position"];
	    }
	}
	export class QueryResult {
	    id: string;
	    type: string;
	    columns?: string[];
	    rows?: any[][];
	    rowsAffected: number;
	    executionTime: number;
	    error?: QueryError;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	        this.rowsAffected = source["rowsAffected"];
	        this.executionTime = source["executionTime"];
	        this.error = this.convertValues(source["error"], QueryError);
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

export namespace repository {
	
	export class Column {
	    name: string;
	    type: string;
	    nullable: boolean;
	    defaultValue?: string;
	    autoIncrement: boolean;
	    comment: string;
	
	    static createFrom(source: any = {}) {
	        return new Column(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.nullable = source["nullable"];
	        this.defaultValue = source["defaultValue"];
	        this.autoIncrement = source["autoIncrement"];
	        this.comment = source["comment"];
	    }
	}
	export class OrderBy {
	    Column: string;
	    Direction: string;
	
	    static createFrom(source: any = {}) {
	        return new OrderBy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Column = source["Column"];
	        this.Direction = source["Direction"];
	    }
	}
	export class Filter {
	    Column: string;
	    Operator: string;
	    Value: any;
	
	    static createFrom(source: any = {}) {
	        return new Filter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Column = source["Column"];
	        this.Operator = source["Operator"];
	        this.Value = source["Value"];
	    }
	}
	export class DataQuery {
	    Database: string;
	    Table: string;
	    Columns: string[];
	    Filters: Filter[];
	    OrderBy: OrderBy[];
	    Limit: number;
	    Offset: number;
	
	    static createFrom(source: any = {}) {
	        return new DataQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Database = source["Database"];
	        this.Table = source["Table"];
	        this.Columns = source["Columns"];
	        this.Filters = this.convertValues(source["Filters"], Filter);
	        this.OrderBy = this.convertValues(source["OrderBy"], OrderBy);
	        this.Limit = source["Limit"];
	        this.Offset = source["Offset"];
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
	export class DataResult {
	    Columns: string[];
	    Rows: any[][];
	    Total: number;
	
	    static createFrom(source: any = {}) {
	        return new DataResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Columns = source["Columns"];
	        this.Rows = source["Rows"];
	        this.Total = source["Total"];
	    }
	}
	export class Database {
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Database(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	    }
	}
	
	export class ForeignKey {
	    name: string;
	    columns: string[];
	    referencedTable: string;
	    referencedColumns: string[];
	    onDelete: string;
	    onUpdate: string;
	
	    static createFrom(source: any = {}) {
	        return new ForeignKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.columns = source["columns"];
	        this.referencedTable = source["referencedTable"];
	        this.referencedColumns = source["referencedColumns"];
	        this.onDelete = source["onDelete"];
	        this.onUpdate = source["onUpdate"];
	    }
	}
	export class Index {
	    name: string;
	    type: string;
	    columns: string[];
	    nonUnique: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Index(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.columns = source["columns"];
	        this.nonUnique = source["nonUnique"];
	    }
	}
	
	export class Table {
	    name: string;
	    rows: number;
	    engine: string;
	    comment: string;
	
	    static createFrom(source: any = {}) {
	        return new Table(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.rows = source["rows"];
	        this.engine = source["engine"];
	        this.comment = source["comment"];
	    }
	}

}

export namespace schema {
	
	export class PrimaryKey {
	    columns: string[];
	
	    static createFrom(source: any = {}) {
	        return new PrimaryKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	    }
	}
	export class TableSchema {
	    name: string;
	    columns: repository.Column[];
	    primaryKey?: PrimaryKey;
	    indexes: repository.Index[];
	    foreignKeys: repository.ForeignKey[];
	    engine: string;
	    charset: string;
	    comment: string;
	
	    static createFrom(source: any = {}) {
	        return new TableSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.columns = this.convertValues(source["columns"], repository.Column);
	        this.primaryKey = this.convertValues(source["primaryKey"], PrimaryKey);
	        this.indexes = this.convertValues(source["indexes"], repository.Index);
	        this.foreignKeys = this.convertValues(source["foreignKeys"], repository.ForeignKey);
	        this.engine = source["engine"];
	        this.charset = source["charset"];
	        this.comment = source["comment"];
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

export namespace storage {
	
	export class QueryHistoryEntry {
	    id: number;
	    // Go type: time
	    timestamp: any;
	    connectionId: string;
	    database: string;
	    sql: string;
	    executionTime: number;
	    rowsAffected: number;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QueryHistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.connectionId = source["connectionId"];
	        this.database = source["database"];
	        this.sql = source["sql"];
	        this.executionTime = source["executionTime"];
	        this.rowsAffected = source["rowsAffected"];
	        this.success = source["success"];
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
	export class SavedQuery {
	    id: number;
	    name: string;
	    sql: string;
	    description: string;
	    connectionId: string;
	    database: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new SavedQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sql = source["sql"];
	        this.description = source["description"];
	        this.connectionId = source["connectionId"];
	        this.database = source["database"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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

export namespace types {
	
	export class Column {
	    name: string;
	    type: string;
	    nullable: boolean;
	    defaultValue?: string;
	    autoIncrement: boolean;
	    comment: string;
	
	    static createFrom(source: any = {}) {
	        return new Column(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.nullable = source["nullable"];
	        this.defaultValue = source["defaultValue"];
	        this.autoIncrement = source["autoIncrement"];
	        this.comment = source["comment"];
	    }
	}
	export class ColumnDiff {
	    columnName: string;
	    oldColumn: Column;
	    newColumn: Column;
	
	    static createFrom(source: any = {}) {
	        return new ColumnDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columnName = source["columnName"];
	        this.oldColumn = this.convertValues(source["oldColumn"], Column);
	        this.newColumn = this.convertValues(source["newColumn"], Column);
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
	export class ConnectionProfile {
	    id: string;
	    name: string;
	    host: string;
	    port: number;
	    username: string;
	    password: string;
	    database: string;
	    charset: string;
	    timeout: number;
	    sshEnabled: boolean;
	    sshHost?: string;
	    sshPort?: number;
	    sshUsername?: string;
	    sshPassword?: string;
	    sshKeyPath?: string;
	    // Go type: time
	    createdAt?: any;
	    // Go type: time
	    updatedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.database = source["database"];
	        this.charset = source["charset"];
	        this.timeout = source["timeout"];
	        this.sshEnabled = source["sshEnabled"];
	        this.sshHost = source["sshHost"];
	        this.sshPort = source["sshPort"];
	        this.sshUsername = source["sshUsername"];
	        this.sshPassword = source["sshPassword"];
	        this.sshKeyPath = source["sshKeyPath"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class ForeignKey {
	    name: string;
	    columns: string[];
	    referencedTable: string;
	    referencedColumns: string[];
	    onDelete: string;
	    onUpdate: string;
	
	    static createFrom(source: any = {}) {
	        return new ForeignKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.columns = source["columns"];
	        this.referencedTable = source["referencedTable"];
	        this.referencedColumns = source["referencedColumns"];
	        this.onDelete = source["onDelete"];
	        this.onUpdate = source["onUpdate"];
	    }
	}
	export class Index {
	    name: string;
	    type: string;
	    columns: string[];
	
	    static createFrom(source: any = {}) {
	        return new Index(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.columns = source["columns"];
	    }
	}
	export class SQLStatement {
	    sql: string;
	    type: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new SQLStatement(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sql = source["sql"];
	        this.type = source["type"];
	        this.description = source["description"];
	    }
	}
	export class TableDiff {
	    tableName: string;
	    columnsAdded: Column[];
	    columnsRemoved: Column[];
	    columnsModified: ColumnDiff[];
	    indexesAdded: Index[];
	    indexesRemoved: Index[];
	    foreignKeysAdded: ForeignKey[];
	    foreignKeysRemoved: ForeignKey[];
	
	    static createFrom(source: any = {}) {
	        return new TableDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tableName = source["tableName"];
	        this.columnsAdded = this.convertValues(source["columnsAdded"], Column);
	        this.columnsRemoved = this.convertValues(source["columnsRemoved"], Column);
	        this.columnsModified = this.convertValues(source["columnsModified"], ColumnDiff);
	        this.indexesAdded = this.convertValues(source["indexesAdded"], Index);
	        this.indexesRemoved = this.convertValues(source["indexesRemoved"], Index);
	        this.foreignKeysAdded = this.convertValues(source["foreignKeysAdded"], ForeignKey);
	        this.foreignKeysRemoved = this.convertValues(source["foreignKeysRemoved"], ForeignKey);
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
	export class SchemaDiff {
	    tablesOnlyInSource: string[];
	    tablesOnlyInTarget: string[];
	    tableDifferences: TableDiff[];
	
	    static createFrom(source: any = {}) {
	        return new SchemaDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tablesOnlyInSource = source["tablesOnlyInSource"];
	        this.tablesOnlyInTarget = source["tablesOnlyInTarget"];
	        this.tableDifferences = this.convertValues(source["tableDifferences"], TableDiff);
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
	export class SyncScript {
	    statements: SQLStatement[];
	
	    static createFrom(source: any = {}) {
	        return new SyncScript(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statements = this.convertValues(source["statements"], SQLStatement);
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

