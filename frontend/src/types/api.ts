// Type definitions for frontend-backend communication

export interface ConnectionProfile {
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
  createdAt: string;
  updatedAt: string;
}

export interface TestResult {
  success: boolean;
  message: string;
  error?: string;
}

export interface Database {
  name: string;
}

export interface Table {
  name: string;
  rows: number;
  engine: string;
  comment: string;
}

export interface Column {
  name: string;
  type: string;
  nullable: boolean;
  defaultValue?: string;
  autoIncrement: boolean;
  comment: string;
}

export interface PrimaryKey {
  columns: string[];
}

export interface Index {
  name: string;
  type: string;
  columns: string[];
}

export interface ForeignKey {
  name: string;
  columns: string[];
  referencedTable: string;
  referencedColumns: string[];
  onDelete: string;
  onUpdate: string;
}

export interface TableSchema {
  name: string;
  columns: Column[];
  primaryKey?: PrimaryKey;
  indexes: Index[];
  foreignKeys: ForeignKey[];
  engine: string;
  charset: string;
  comment: string;
}

export interface SchemaChange {
  type: string; // ADD_COLUMN, DROP_COLUMN, MODIFY_COLUMN, ADD_INDEX, DROP_INDEX, etc.
  target: string; // column name, index name, etc.
  definition?: any; // new definition for the change
}

export interface DataQuery {
  database: string;
  table: string;
  columns: string[];
  filters: Filter[];
  orderBy: OrderBy[];
  limit: number;
  offset: number;
}

export interface Filter {
  column: string;
  operator: string;
  value: any;
}

export interface OrderBy {
  column: string;
  direction: string;
}

export interface DataResult {
  columns: string[];
  rows: any[][];
  total: number;
}

export interface QueryResult {
  type: string;
  columns: string[];
  rows: any[][];
  rowsAffected: number;
  executionTime: number;
  error?: QueryError;
}

export interface QueryError {
  code: number;
  message: string;
  position: number;
}

export interface SchemaDiff {
  tablesOnlyInSource: string[];
  tablesOnlyInTarget: string[];
  tableDifferences: TableDiff[];
}

export interface TableDiff {
  tableName: string;
  columnsAdded: Column[];
  columnsRemoved: Column[];
  columnsModified: ColumnDiff[];
  indexesAdded: Index[];
  indexesRemoved: Index[];
  foreignKeysAdded: ForeignKey[];
  foreignKeysRemoved: ForeignKey[];
}

export interface ColumnDiff {
  columnName: string;
  oldColumn: Column;
  newColumn: Column;
}

export interface SyncScript {
  statements: SQLStatement[];
}

export interface SQLStatement {
  sql: string;
  type: string;
  description: string;
}

export interface ImportResult {
  totalRows: number;
  successRows: number;
  failedRows: number;
  errors: ImportError[];
}

export interface ImportError {
  row: number;
  message: string;
}

export interface ColumnMapping {
  fileColumns: string[];
  tableColumns: string[];
}

export interface LogEntry {
  id: number;
  timestamp: string;
  level: string;
  operation: string;
  message: string;
  details?: Record<string, any>;
  connectionId?: string;
}

export interface LogFilter {
  level?: string;
  startTime?: string;
  endTime?: string;
  keyword?: string;
  connectionId?: string;
  limit: number;
  offset: number;
}
