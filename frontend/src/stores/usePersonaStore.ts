import { create } from 'zustand';
import { fetchPersonas } from '../api/client';
import type { PersonaSummary } from '../types';

export interface PersonaState {
  personas: PersonaSummary[];
  loading: boolean;
  error: string | null;
  fetchPersonas: () => Promise<void>;
  getPersonaById: (id: string) => PersonaSummary | undefined;
  getPersonaNames: (ids: string[]) => string;
}

export const usePersonaStore = create<PersonaState>((set, get) => ({
  personas: [],
  loading: false,
  error: null,

  fetchPersonas: async () => {
    set({ loading: true, error: null });
    try {
      const data = await fetchPersonas();
      set({ personas: data, loading: false });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to load personas', loading: false });
    }
  },

  getPersonaById: (id: string) => {
    return get().personas.find((p) => p.id === id);
  },

  getPersonaNames: (ids: string[]) => {
    const { personas } = get();
    return ids
      .map((id) => personas.find((p) => p.id === id)?.name ?? id)
      .join('、');
  },
}));
