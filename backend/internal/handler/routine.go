package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/google/uuid"
)

// RoutineService defines the operations the routine handler needs.
type RoutineService interface {
	CreateRoutine(ctx context.Context, name string) (queries.Routine, error)
	ListRoutines(ctx context.Context, limit, offset int32) (service.RoutineListResult, error)
	GetRoutine(ctx context.Context, id uuid.UUID) (service.RoutineDetail, error)
	UpdateRoutine(ctx context.Context, id uuid.UUID, name string) (queries.Routine, error)
	DeleteRoutine(ctx context.Context, id uuid.UUID) error
	AddRoutineExercise(ctx context.Context, routineID uuid.UUID, exerciseIDStr string, order *int32) (queries.RoutineExercise, error)
	UpdateRoutineExerciseOrder(ctx context.Context, routineID uuid.UUID, routineExerciseID uuid.UUID, newOrder int32) (queries.RoutineExercise, error)
	DeleteRoutineExercise(ctx context.Context, routineID uuid.UUID, routineExerciseID uuid.UUID) error
}

// RoutineExerciseResponse represents a routine exercise in JSON.
type RoutineExerciseResponse struct {
	ID           uuid.UUID `json:"id"`
	RoutineID    uuid.UUID `json:"routine_id"`
	ExerciseID   uuid.UUID `json:"exercise_id"`
	Order        int32     `json:"order"`
	ExerciseName string    `json:"exercise_name"`
	TargetMuscle string    `json:"target_muscle"`
}

// RoutineResponse represents a routine in JSON.
type RoutineResponse struct {
	ID        uuid.UUID               `json:"id"`
	Name      string                  `json:"name"`
	CreatedAt string                  `json:"created_at"`
	Exercises []RoutineExerciseResponse `json:"exercises"`
}

// RoutineListResponse is the response for list routines.
type RoutineListResponse struct {
	Data       []RoutineResponse `json:"data"`
	Pagination Pagination        `json:"pagination"`
}

// Routine handles routine HTTP requests.
type Routine struct {
	svc RoutineService
}

// NewRoutine creates a new Routine handler.
func NewRoutine(svc RoutineService) *Routine {
	return &Routine{svc: svc}
}

func routineToResponse(r queries.Routine) RoutineResponse {
	return RoutineResponse{
		ID:        r.ID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Exercises: []RoutineExerciseResponse{},
	}
}

func routineExerciseToResponse(re queries.ListRoutineExercisesRow) RoutineExerciseResponse {
	tm := ""
	if re.TargetMuscle.Valid {
		tm = re.TargetMuscle.String
	}
	return RoutineExerciseResponse{
		ID:           re.ID,
		RoutineID:    re.RoutineID,
		ExerciseID:   re.ExerciseID,
		Order:        re.Order,
		ExerciseName: re.ExerciseName,
		TargetMuscle: tm,
	}
}

func routineDetailToResponse(d service.RoutineDetail) RoutineResponse {
	resp := routineToResponse(d.Routine)
	resp.Exercises = make([]RoutineExerciseResponse, len(d.Exercises))
	for i, re := range d.Exercises {
		resp.Exercises[i] = routineExerciseToResponse(re)
	}
	return resp
}

func routineExerciseModelToResponse(re queries.RoutineExercise) RoutineExerciseResponse {
	return RoutineExerciseResponse{
		ID:         re.ID,
		RoutineID:  re.RoutineID,
		ExerciseID: re.ExerciseID,
		Order:      re.Order,
	}
}

// Create handles POST /api/v1/routines.
func (h *Routine) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
		return
	}

	routine, err := h.svc.CreateRoutine(r.Context(), req.Name)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			status := http.StatusInternalServerError
			switch appErr.Code {
			case "VALIDATION_ERROR":
				status = http.StatusUnprocessableEntity
			}
			writeError(w, status, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusCreated, routineToResponse(routine))
}

// List handles GET /api/v1/routines.
func (h *Routine) List(w http.ResponseWriter, r *http.Request) {
	limit := parseIntQuery(r, "limit", 20)
	offset := parseIntQuery(r, "offset", 0)

	result, err := h.svc.ListRoutines(r.Context(), limit, offset)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			writeError(w, http.StatusBadRequest, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	resp := RoutineListResponse{
		Data: make([]RoutineResponse, len(result.Data)),
		Pagination: Pagination{
			Limit:  result.Limit,
			Offset: result.Offset,
			Total:  result.Total,
		},
	}
	for i, routine := range result.Data {
		resp.Data[i] = routineToResponse(routine)
	}

	writeJSON(w, http.StatusOK, resp)
}

// Get handles GET /api/v1/routines/:id.
func (h *Routine) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routine id", "")
		return
	}

	detail, err := h.svc.GetRoutine(r.Context(), id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			if appErr.Code == "NOT_FOUND" {
				writeError(w, http.StatusNotFound, appErr.Code, appErr.Message, "")
				return
			}
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusOK, routineDetailToResponse(detail))
}

// Update handles PUT /api/v1/routines/:id.
func (h *Routine) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routine id", "")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
		return
	}

	routine, err := h.svc.UpdateRoutine(r.Context(), id, req.Name)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			status := http.StatusInternalServerError
			switch appErr.Code {
			case "VALIDATION_ERROR":
				status = http.StatusUnprocessableEntity
			case "NOT_FOUND":
				status = http.StatusNotFound
			}
			writeError(w, status, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusOK, routineToResponse(routine))
}

// Delete handles DELETE /api/v1/routines/:id.
func (h *Routine) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routine id", "")
		return
	}

	if err := h.svc.DeleteRoutine(r.Context(), id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			status := http.StatusInternalServerError
			if appErr.Code == "NOT_FOUND" {
				status = http.StatusNotFound
			}
			writeError(w, status, appErr.Code, appErr.Message, "")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddExercise handles POST /api/v1/routines/:id/exercises.
func (h *Routine) AddExercise(w http.ResponseWriter, r *http.Request) {
	routineID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routine id", "")
		return
	}

	var req struct {
		ExerciseID string `json:"exercise_id"`
		Order      *int32 `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
		return
	}

	re, err := h.svc.AddRoutineExercise(r.Context(), routineID, req.ExerciseID, req.Order)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			status := http.StatusInternalServerError
			switch appErr.Code {
			case "VALIDATION_ERROR":
				status = http.StatusUnprocessableEntity
			case "NOT_FOUND":
				status = http.StatusNotFound
			case "CONFLICT":
				status = http.StatusConflict
			}
			writeError(w, status, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusCreated, routineExerciseModelToResponse(re))
}

// UpdateExerciseOrder handles PUT /api/v1/routines/:id/exercises/:routineExerciseId.
func (h *Routine) UpdateExerciseOrder(w http.ResponseWriter, r *http.Request) {
	routineID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routine id", "")
		return
	}

	routineExerciseID, err := parseUUIDParam(r, "routineExerciseId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routine exercise id", "")
		return
	}

	var req struct {
		Order int32 `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
		return
	}

	re, err := h.svc.UpdateRoutineExerciseOrder(r.Context(), routineID, routineExerciseID, req.Order)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			status := http.StatusInternalServerError
			switch appErr.Code {
			case "VALIDATION_ERROR":
				status = http.StatusUnprocessableEntity
			case "NOT_FOUND":
				status = http.StatusNotFound
			}
			writeError(w, status, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusOK, routineExerciseModelToResponse(re))
}

// DeleteExercise handles DELETE /api/v1/routines/:id/exercises/:routineExerciseId.
func (h *Routine) DeleteExercise(w http.ResponseWriter, r *http.Request) {
	routineID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routine id", "")
		return
	}

	routineExerciseID, err := parseUUIDParam(r, "routineExerciseId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routine exercise id", "")
		return
	}

	if err := h.svc.DeleteRoutineExercise(r.Context(), routineID, routineExerciseID); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			status := http.StatusInternalServerError
			if appErr.Code == "NOT_FOUND" {
				status = http.StatusNotFound
			}
			writeError(w, status, appErr.Code, appErr.Message, "")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
