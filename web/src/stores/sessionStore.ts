import { type StateCreator } from 'zustand';
import { sessionApi } from '../api/session';
import type { Session, SetResponse, SessionSummary, HistorySession } from '../types';
import type { AllSlices } from './index';

const LS_ACTIVE_SESSION_KEY = 'omr_active_session';

export interface SessionWithSets extends Session {
  sets?: SetResponse[];
}

export interface SessionSlice {
  sessions: Session[];
  currentSession: SessionWithSets | null;
  sessionHistory: HistorySession[];
  sessionSummary: SessionSummary | null;
  activeSessionId: string | null;
  loading: boolean;
  error: string | null;
  fetchSessions: (params?: { limit?: number; offset?: number }) => Promise<void>;
  fetchSessionHistory: (params?: { limit?: number; offset?: number }) => Promise<void>;
  startSession: (routineId?: string) => Promise<SessionWithSets>;
  endSession: (id: string) => Promise<Session>;
  addSet: (sessionId: string, data: { exercise_id: string; weight: number; reps: number }) => Promise<SetResponse>;
  updateSet: (sessionId: string, setId: string, data: { weight?: number; reps?: number }) => Promise<SetResponse>;
  deleteSet: (sessionId: string, setId: string) => Promise<void>;
  fetchSessionSummary: (id: string) => Promise<SessionSummary>;
  setCurrentSession: (session: SessionWithSets | null) => void;
  hydrateActiveSession: () => Promise<void>;
  clearActiveSession: () => void;
}

export const createSessionSlice: StateCreator<AllSlices, [], [], SessionSlice> = (set) => ({
  sessions: [],
  currentSession: null,
  sessionHistory: [],
  sessionSummary: null,
  activeSessionId: null,
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
    const sessionWithSets: SessionWithSets = { ...session, sets: [] };
    localStorage.setItem(LS_ACTIVE_SESSION_KEY, session.id);
    set((state) => ({
      sessions: [session, ...state.sessions],
      currentSession: sessionWithSets,
      activeSessionId: session.id,
    }));
    return sessionWithSets;
  },

  endSession: async (id) => {
    const session = await sessionApi.end(id);
    localStorage.removeItem(LS_ACTIVE_SESSION_KEY);
    set((state) => ({
      sessions: state.sessions.map((s) => (s.id === id ? session : s)),
      currentSession: state.currentSession?.id === id ? null : state.currentSession,
      activeSessionId: state.activeSessionId === id ? null : state.activeSessionId,
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

  hydrateActiveSession: async () => {
    const savedId = localStorage.getItem(LS_ACTIVE_SESSION_KEY);
    if (!savedId) return;

    set({ loading: true, error: null });
    try {
      const sessionWithSets = await sessionApi.getWithSets(savedId);
      if (sessionWithSets.ended_at) {
        localStorage.removeItem(LS_ACTIVE_SESSION_KEY);
        set({ activeSessionId: null });
      } else {
        set({ currentSession: sessionWithSets, activeSessionId: savedId });
      }
    } catch {
      localStorage.removeItem(LS_ACTIVE_SESSION_KEY);
      set({ activeSessionId: null });
    } finally {
      set({ loading: false });
    }
  },

  clearActiveSession: () => {
    localStorage.removeItem(LS_ACTIVE_SESSION_KEY);
    set({ currentSession: null, activeSessionId: null });
  },
});
