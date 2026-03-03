import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Database, Table } from '../types/api';

export const useDatabaseStore = defineStore('database', () => {
  const currentDatabase = ref<string | null>(null);
  const currentTable = ref<string | null>(null);
  const databases = ref<Database[]>([]);
  const tables = ref<Table[]>([]);

  function setCurrentDatabase(database: string | null) {
    currentDatabase.value = database;
    if (!database) {
      currentTable.value = null;
      tables.value = [];
    }
  }

  function setCurrentTable(table: string | null) {
    currentTable.value = table;
  }

  function setDatabases(newDatabases: Database[]) {
    databases.value = newDatabases;
  }

  function setTables(newTables: Table[]) {
    tables.value = newTables;
  }

  return {
    currentDatabase,
    currentTable,
    databases,
    tables,
    setCurrentDatabase,
    setCurrentTable,
    setDatabases,
    setTables,
  };
});
