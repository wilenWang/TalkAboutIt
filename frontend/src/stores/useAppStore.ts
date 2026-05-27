import { create } from 'zustand';
import type { StreamMessage } from '../components/MessageStream';

export type AppStatus = 'idle' | 'creating' | 'streaming' | 'completed';

interface AppState {
  // Topic & settings
  topic: string;
  rounds: number;
  selectedPersonas: string[];

  // Streaming state
  status: AppStatus;
  messages: StreamMessage[];
  currentSpeaker: { name: string; avatar: string } | null;
  error: string | null;
  rtId: string | null;
  currentRound: number;

  // Actions
  setTopic: (topic: string) => void;
  setRounds: (rounds: number) => void;
  togglePersona: (id: string) => void;
  setSelectedPersonas: (ids: string[]) => void;

  startCreating: () => void;
  startStreaming: () => void;
  complete: () => void;
  reset: () => void;

  setCurrentSpeaker: (speaker: { name: string; avatar: string } | null) => void;
  setError: (error: string | null) => void;
  setRtId: (id: string | null) => void;
  setCurrentRound: (round: number) => void;

  addMessageChunk: (data: { chunk: string; persona_id: string; round: number; persona_name?: string; avatar?: string }) => void;
  finalizeMessage: (data: { message_id: string; persona_id: string; persona_name: string; avatar: string; round: number; content: string }) => void;
  abortMessage: (personaId: string) => void;
  setMessages: (messages: StreamMessage[]) => void;
  clearMessages: () => void;
}

const initialTopic = 'AI 会取代程序员吗？';

export const useAppStore = create<AppState>((set, get) => ({
  topic: initialTopic,
  rounds: 3,
  selectedPersonas: [],
  status: 'idle',
  messages: [],
  currentSpeaker: null,
  error: null,
  rtId: null,
  currentRound: 1,

  setTopic: (topic) => set({ topic }),
  setRounds: (rounds) => set({ rounds }),

  togglePersona: (id) => {
    const { selectedPersonas } = get();
    if (selectedPersonas.includes(id)) {
      set({ selectedPersonas: selectedPersonas.filter((s) => s !== id) });
    } else if (selectedPersonas.length < 4) {
      set({ selectedPersonas: [...selectedPersonas, id] });
    }
  },

  setSelectedPersonas: (ids) => set({ selectedPersonas: ids }),

  startCreating: () => set({ status: 'creating', error: null }),
  startStreaming: () => set({ status: 'streaming', error: null }),
  complete: () => set({ status: 'completed', currentSpeaker: null }),

  reset: () => set({
    status: 'idle',
    messages: [],
    currentSpeaker: null,
    error: null,
    rtId: null,
    currentRound: 1,
  }),

  setCurrentSpeaker: (speaker) => set({ currentSpeaker: speaker }),
  setError: (error) => set({ error }),
  setRtId: (rtId) => set({ rtId }),
  setCurrentRound: (currentRound) => set({ currentRound }),

  addMessageChunk: (data) => {
    const { currentSpeaker } = get();
    set((state) => {
      const prev = state.messages;
      const last = prev[prev.length - 1];
      if (last && last.status === 'streaming' && last.personaId === data.persona_id) {
        const updated = [...prev];
        updated[updated.length - 1] = {
          ...last,
          content: last.content + data.chunk,
        };
        return { messages: updated };
      }
      const newMsg: StreamMessage = {
        id: `temp_${data.persona_id}_${Date.now()}`,
        avatar: data.avatar ?? currentSpeaker?.avatar ?? '🤖',
        author: data.persona_name ?? currentSpeaker?.name ?? 'AI',
        personaId: data.persona_id,
        round: data.round,
        content: data.chunk,
        status: 'streaming',
      };
      return { messages: [...prev, newMsg] };
    });
  },

  finalizeMessage: (data) => {
    set((state) => {
      const filtered = state.messages.filter(
        (m) => m.id !== data.message_id && !(m.status === 'streaming' && m.personaId === data.persona_id)
      );
      const newMsg: StreamMessage = {
        id: data.message_id,
        avatar: data.avatar,
        author: data.persona_name,
        personaId: data.persona_id,
        round: data.round,
        content: data.content,
        status: 'done',
      };
      return { messages: [...filtered, newMsg], currentSpeaker: null };
    });
  },

  abortMessage: (personaId) => {
    set((state) => ({
      messages: state.messages.filter((m) => !(m.status === 'streaming' && m.personaId === personaId)),
      currentSpeaker: null,
    }));
  },

  setMessages: (messages) => set({ messages }),
  clearMessages: () => set({ messages: [] }),
}));
