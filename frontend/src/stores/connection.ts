import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { ConnectionProfile } from '../types/api';

export const useConnectionStore = defineStore('connection', () => {
  const currentConnection = ref<ConnectionProfile | null>(null);
  const isConnected = ref(false);
  const profiles = ref<ConnectionProfile[]>([]);

  function setCurrentConnection(profile: ConnectionProfile | null) {
    currentConnection.value = profile;
    isConnected.value = profile !== null;
  }

  function setProfiles(newProfiles: ConnectionProfile[]) {
    profiles.value = newProfiles;
  }

  function addProfile(profile: ConnectionProfile) {
    profiles.value.push(profile);
  }

  function updateProfile(id: string, profile: ConnectionProfile) {
    const index = profiles.value.findIndex(p => p.id === id);
    if (index !== -1) {
      profiles.value[index] = profile;
    }
  }

  function removeProfile(id: string) {
    profiles.value = profiles.value.filter(p => p.id !== id);
  }

  return {
    currentConnection,
    isConnected,
    profiles,
    setCurrentConnection,
    setProfiles,
    addProfile,
    updateProfile,
    removeProfile,
  };
});
