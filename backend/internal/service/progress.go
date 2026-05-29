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