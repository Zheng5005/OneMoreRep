export interface Exercise {
  id: string;
  name: string;
  target_muscle: string;
  notes: string | null;
  created_at: string;
}

export interface RoutineExercise {
  id: string;
  routine_id: string;
  exercise_id: string;
  order: number;
  exercise_name?: string;
  target_muscle?: string;
}

export interface Routine {
  id: string;
  name: string;
  created_at: string;
  exercises?: RoutineExercise[];
}

export interface SetResponse {
  id: string;
  session_id: string;
  exercise_id: string;
  set_number: number;
  weight: number;
  reps: number;
  created_at: string;
  exercise_name?: string;
}

export interface Session {
  id: string;
  routine_id: string | null;
  started_at: string;
  ended_at: string | null;
  sets?: SetResponse[];
}

export interface LastValues {
  weight: number | null;
  reps: number | null;
}

export interface SessionExerciseSummary {
  exercise_id: string;
  exercise_name: string;
  sets_count: number;
  best_volume: number;
  best_weight: number;
  best_reps: number;
}

export interface SessionSummary {
  session_id: string;
  started_at: string;
  ended_at: string | null;
  duration_seconds: number;
  total_volume: number;
  exercise_count: number;
  total_sets: number;
  exercises: SessionExerciseSummary[];
}

export interface HistorySet {
  set_id: string;
  set_number: number;
  weight: number;
  reps: number;
  volume: number;
  is_pr: boolean;
}

export interface HistorySession {
  session_id: string;
  started_at: string;
  ended_at: string | null;
  sets: HistorySet[];
}

export interface VolumePoint {
  period?: string;
  session_id?: string;
  started_at?: string;
  total_volume: number;
}

export interface Quote {
  id: string;
  text: string;
  author: string;
  category?: string | null;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface ApiErrorBody {
  code: string;
  message: string;
  field?: string;
}

export interface ApiErrorResponse {
  error: ApiErrorBody;
}
