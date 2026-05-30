import { type StateCreator } from 'zustand';
import { sessionApi } from '../api/session';
import type { Session, SetResponse, SessionSummary, HistorySession } from '../types';
import type { AllSlices } from './index';

export interface SessionSlice {
  sessions: Session[];
  currentSession: Session | null;
  sessionHistory: HistorySession[];
  sessionSummary: SessionSummary | null;
  loading: boolean;
  error: string | null;
  fetchSessions: (params?: { limit?: number; offset?: number }) => Promise<void>;
  fetchSessionHistory: (params?: { limit?: number; offset?: number }) => Promise<void>;
  startSession: (routineId?: string) => Promise<Session>;
  endSession: (id: string) => Promise<Session>;
  addSet: (sessionId: string, data: { exercise_id: string; weight: number; reps: number }) => Promise<SetResponse>;
  updateSet: (sessionId: string, setId: string, data: { weight?: number; reps?: number }) => Promise<SetResponse>;
  deleteSet: (sessionId: string, setId: string) => Promise<void>;
  fetchSessionSummary: (id: string) => Promise<SessionSummary>;
  setCurrentSession: (session: Session | null) => void;
}

export const createSessionSlice: StateCreator<AllSlices, [], [], SessionSlice> = (set) => ({
  sessions: [],
  currentSession: null,
  sessionHistory: [],
  sessionSummary: null,
  loading: false,
  error: null,

  fetchSessions: async (params) => {
    set({ loading: true, error: null });
    try {
      const response = await sessionApi.list(params);
      set({ sessions: response.data });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to fetch sessions' });
    } finally {
      set({ loading: false });
    }
  },

  fetchSessionHistory: async (params) => {
    set({ loading: true, error: null });
    try {
      const response = await sessionApi.history(params);
      set({ sessionHistory: response.data });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to fetch session history' });
    } finally {
      set({ loading: false });
    }
  },

  startSession: async (routineId) => {
    const session = await sessionApi.start(routineId);
    set((state) => ({ sessions: [session, ...state.sessions], currentSession: session }));
    return session;
  },

  endSession: async (id) => {
    const session = await sessionApi.end(id);
    set((state) => ({
      sessions: state.sessions.map((s) => (s.id === id ? session : s)),
      currentSession: state.currentSession?.id === id ? null : state.currentSession,
    }));
    return session;
  },

  addSet: async (sessionId, data) => {
    const newSet = await sessionApi.addSet(sessionId, data);
    set((state) => {
      const sessions = state.sessions.map((s) => {
        if (s.id !== sessionId) return s;
        return { ...s, sets: [...(s.sets || []), newSet] };
      });
      const currentSession = state.currentSession?.id === sessionId
        ? { ...state.currentSession, sets: [...(state.currentSession.sets || []), newSet] }
        : state.currentSession;
      return { sessions, currentSession };
    });
    return newSet;
  },

  updateSet: async (sessionId, setId, data) => {
    const updatedSet = await sessionApi.updateSet(sessionId, setId, data);
    set((state) => {
      const applyUpdate = (s: Session): Session => ({
        ...s,
        sets: s.sets?.map((st): SetResponse => (st.id === setId ? { ...st, ...updatedSet } : st)),
      });
      return {
        sessions: state.sessions.map((s) => (s.id === sessionId ? applyUpdate(s) : s)),
        currentSession: state.currentSession?.id === sessionId ? applyUpdate(state.currentSession) : state.currentSession,
      };
    });
    return updatedSet;
  },

  deleteSet: async (sessionId, setId) => {
    await sessionApi.deleteSet(sessionId, setId);
    set((state) => {
      const removeSet = (s: Session) => ({
        ...s,
        sets: s.sets?.filter((st) => st.id !== setId),
      });
      return {
        sessions: state.sessions.map((s) => (s.id === sessionId ? removeSet(s) : s)),
        currentSession: state.currentSession?.id === sessionId ? removeSet(state.currentSession) : state.currentSession,
      };
    });
  },

  fetchSessionSummary: async (id) => {
    const summary = await sessionApi.summary(id);
    set({ sessionSummary: summary });
    return summary;
  },

  setCurrentSession: (session) => {
    set({ currentSession: session });
  },
});
