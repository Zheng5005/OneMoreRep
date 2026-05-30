package service

import (
	"context"
	"math"
	"time"

	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProgressStore interface {
	GetExerciseLastValues(ctx context.Context, exerciseID uuid.UUID) (queries.GetExerciseLastValuesRow, error)
	GetSessionSummary(ctx context.Context, id uuid.UUID) (queries.GetSessionSummaryRow, error)
	GetSessionExerciseBreakdown(ctx context.Context, sessionID uuid.UUID) ([]queries.GetSessionExerciseBreakdownRow, error)
	GetWorkoutSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
	GetExerciseHistory(ctx context.Context, arg queries.GetExerciseHistoryParams) ([]queries.GetExerciseHistoryRow, error)
	GetVolumeBySession(ctx context.Context, exerciseID uuid.UUID) ([]queries.GetVolumeBySessionRow, error)
	GetVolumeByWeek(ctx context.Context, exerciseID uuid.UUID) ([]queries.GetVolumeByWeekRow, error)
	GetVolumeByMonth(ctx context.Context, exerciseID uuid.UUID) ([]queries.GetVolumeByMonthRow, error)
}

type LastValues struct {
	Weight *float64
	Reps   *float64
}

type ExerciseBreakdown struct {
	ExerciseID   uuid.UUID `json:"exercise_id"`
	ExerciseName string    `json:"exercise_name"`
	SetsCount    int64     `json:"sets_count"`
	BestVolume   float64   `json:"best_volume"`
	BestWeight   float64   `json:"best_weight"`
	BestReps     float64   `json:"best_reps"`
}

type SessionSummary struct {
	SessionID     uuid.UUID           `json:"session_id"`
	StartedAt     string              `json:"started_at"`
	EndedAt       *string             `json:"ended_at,omitempty"`
	DurationSecs  int64               `json:"duration_secs"`
	TotalVolume   float64             `json:"total_volume"`
	ExerciseCount int64               `json:"exercise_count"`
	TotalSets     int64               `json:"total_sets"`
	Exercises     []ExerciseBreakdown `json:"exercises"`
}

type ExerciseHistorySet struct {
	SetID           uuid.UUID `json:"set_id"`
	SetNumber       int32     `json:"set_number"`
	Weight          float64   `json:"weight"`
	Reps            int32     `json:"reps"`
	Volume          float64   `json:"volume"`
	IsPR            bool      `json:"is_pr"`
	SetCreatedAt    string    `json:"set_created_at"`
	SessionID       uuid.UUID `json:"session_id"`
	SessionStartedAt string   `json:"session_started_at"`
	SessionEndedAt   *string   `json:"session_ended_at,omitempty"`
}

type ExerciseHistory struct {
	Sessions []ExerciseHistorySession `json:"sessions"`
}

type ExerciseHistorySession struct {
	SessionID       uuid.UUID             `json:"session_id"`
	StartedAt       string                `json:"started_at"`
	EndedAt         *string               `json:"ended_at,omitempty"`
	Sets            []ExerciseHistorySet  `json:"sets"`
}

type VolumePeriod struct {
	Period      string  `json:"period"`
	TotalVolume float64 `json:"total_volume"`
}

type ProgressService struct {
	store  ProgressStore
	withTx func(ctx context.Context, fn func(ProgressStore) error) error
}

func NewProgressService(store ProgressStore) *ProgressService {
	return &ProgressService{store: store}
}

func numericToFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	case int:
		return float64(n)
	default:
		return 0
	}
}

func (s *ProgressService) GetExerciseHistory(ctx context.Context, exerciseID uuid.UUID, filter string) (*ExerciseHistory, error) {
	if filter != "all" && filter != "30d" && filter != "6m" {
		return nil, NewBadRequestError("invalid filter value: must be all, 30d, or 6m")
	}

	rows, err := s.store.GetExerciseHistory(ctx, queries.GetExerciseHistoryParams{
		ExerciseID: exerciseID,
		Column2:    filter,
	})
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return &ExerciseHistory{Sessions: []ExerciseHistorySession{}}, nil
	}

	var maxVolume float64
	for _, r := range rows {
		wf, _ := r.Weight.Float64Value()
		weight := wf.Float64
		volume := weight * float64(r.Reps)
		if volume > maxVolume {
			maxVolume = volume
		}
	}

	sessionMap := make(map[uuid.UUID]*ExerciseHistorySession)
	for _, r := range rows {
		wf, _ := r.Weight.Float64Value()
		weight := wf.Float64
		volume := weight * float64(r.Reps)

		session, exists := sessionMap[r.SessionID]
		if !exists {
			sessionStartedAt := r.SessionStartedAt.Format("2006-01-02T15:04:05Z07:00")
			session = &ExerciseHistorySession{
				SessionID:    r.SessionID,
				StartedAt:    sessionStartedAt,
				Sets:         []ExerciseHistorySet{},
			}
			if r.SessionEndedAt.Valid {
				t := r.SessionEndedAt.Time.Format("2006-01-02T15:04:05Z07:00")
				session.EndedAt = &t
			}
			sessionMap[r.SessionID] = session
		}

		set := ExerciseHistorySet{
			SetID:            r.SetID,
			SetNumber:        r.SetNumber,
			Weight:           weight,
			Reps:             r.Reps,
			Volume:           volume,
			IsPR:             volume > 0 && volume == maxVolume,
			SetCreatedAt:     r.SetCreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			SessionID:        r.SessionID,
			SessionStartedAt: r.SessionStartedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if r.SessionEndedAt.Valid {
			t := r.SessionEndedAt.Time.Format("2006-01-02T15:04:05Z07:00")
			set.SessionEndedAt = &t
		}
		session.Sets = append(session.Sets, set)
	}

	sessions := make([]ExerciseHistorySession, 0, len(sessionMap))
	for _, session := range sessionMap {
		sessions = append(sessions, *session)
	}

	return &ExerciseHistory{Sessions: sessions}, nil
}

func (s *ProgressService) GetVolumeAggregation(ctx context.Context, groupBy string, exerciseID *uuid.UUID) ([]VolumePeriod, error) {
	var exerciseUUID uuid.UUID
	if exerciseID != nil {
		exerciseUUID = *exerciseID
	}

	switch groupBy {
	case "session":
		rows, err := s.store.GetVolumeBySession(ctx, exerciseUUID)
		if err != nil {
			return nil, err
		}
		result := make([]VolumePeriod, 0, len(rows))
		for _, r := range rows {
			result = append(result, VolumePeriod{
				Period:      r.SessionID.String(),
				TotalVolume: numericToFloat64(r.TotalVolume),
			})
		}
		return result, nil
	case "week":
		rows, err := s.store.GetVolumeByWeek(ctx, exerciseUUID)
		if err != nil {
			return nil, err
		}
		result := make([]VolumePeriod, 0, len(rows))
		for _, r := range rows {
			result = append(result, VolumePeriod{
				Period:      r.Period,
				TotalVolume: numericToFloat64(r.TotalVolume),
			})
		}
		return result, nil
	case "month":
		rows, err := s.store.GetVolumeByMonth(ctx, exerciseUUID)
		if err != nil {
			return nil, err
		}
		result := make([]VolumePeriod, 0, len(rows))
		for _, r := range rows {
			result = append(result, VolumePeriod{
				Period:      r.Period,
				TotalVolume: numericToFloat64(r.TotalVolume),
			})
		}
		return result, nil
	default:
		return nil, NewBadRequestError("invalid group_by value: must be session, week, or month")
	}
}

func (s *ProgressService) GetExerciseLastValues(ctx context.Context, exerciseID uuid.UUID) (*LastValues, error) {
	row, err := s.store.GetExerciseLastValues(ctx, exerciseID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	result := &LastValues{}

	wf, _ := row.Weight.Float64Value()
	if row.Weight.Valid && !math.IsNaN(wf.Float64) {
		weight := wf.Float64
		result.Weight = &weight
	}

	if row.Reps > 0 {
		reps := float64(row.Reps)
		result.Reps = &reps
	}

	return result, nil
}

func (s *ProgressService) GetSessionSummary(ctx context.Context, sessionID uuid.UUID) (*SessionSummary, error) {
	session, err := s.store.GetWorkoutSession(ctx, sessionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, NewNotFoundError("session not found")
		}
		return nil, err
	}

	summary, err := s.store.GetSessionSummary(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	breakdown, err := s.store.GetSessionExerciseBreakdown(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var durationSecs int64
	if session.EndedAt.Valid {
		durationSecs = int64(session.EndedAt.Time.Sub(session.StartedAt).Seconds())
	} else {
		durationSecs = int64(time.Now().Sub(session.StartedAt).Seconds())
	}

	exercises := make([]ExerciseBreakdown, 0, len(breakdown))
	for _, b := range breakdown {
		exercises = append(exercises, ExerciseBreakdown{
			ExerciseID:   b.ExerciseID,
			ExerciseName: b.ExerciseName,
			SetsCount:    b.SetsCount,
			BestVolume:   numericToFloat64(b.BestVolume),
			BestWeight:   numericToFloat64(b.BestWeight),
			BestReps:     numericToFloat64(b.BestReps),
		})
	}

	result := &SessionSummary{
		SessionID:     summary.SessionID,
		StartedAt:     session.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		DurationSecs:  durationSecs,
		TotalVolume:   numericToFloat64(summary.TotalVolume),
		ExerciseCount: summary.ExerciseCount,
		TotalSets:     summary.TotalSets,
		Exercises:     exercises,
	}

	if session.EndedAt.Valid {
		t := session.EndedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		result.EndedAt = &t
	}

	return result, nil
}