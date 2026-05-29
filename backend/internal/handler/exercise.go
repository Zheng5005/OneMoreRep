package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ExerciseService defines the operations the exercise handler needs.
type ExerciseService interface {
	CreateExercise(ctx context.Context, name, targetMuscle, notes string) (queries.Exercise, error)
	ListExercises(ctx context.Context, limit, offset int32, search string) (service.ExerciseListResult, error)
	GetExercise(ctx context.Context, id uuid.UUID) (queries.Exercise, error)
	UpdateExercise(ctx context.Context, id uuid.UUID, name, targetMuscle, notes string) (queries.Exercise, error)
	DeleteExercise(ctx context.Context, id uuid.UUID) error
}

// ExerciseResponse is the response for single exercise endpoints.
type ExerciseResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	TargetMuscle string    `json:"target_muscle"`
	Notes        string    `json:"notes"`
	CreatedAt    string    `json:"created_at"`
}

// ExerciseListResponse is the response for list exercises.
type ExerciseListResponse struct {
	Data       []ExerciseResponse `json:"data"`
	Pagination Pagination         `json:"pagination"`
}

// Pagination holds pagination metadata.
type Pagination struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
	Total  int64 `json:"total"`
}

// ErrorResponse is the standard error response shape.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail holds error details.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Exercise handles exercise HTTP requests.
type Exercise struct {
	svc ExerciseService
}

// NewExercise creates a new Exercise handler.
func NewExercise(svc ExerciseService) *Exercise {
	return &Exercise{svc: svc}
}

func exerciseToResponse(ex queries.Exercise) ExerciseResponse {
	notes := ""
	if ex.Notes.Valid {
		notes = ex.Notes.String
	}
	tm := ""
	if ex.TargetMuscle.Valid {
		tm = ex.TargetMuscle.String
	}
	return ExerciseResponse{
		ID:           ex.ID,
		Name:         ex.Name,
		TargetMuscle: tm,
		Notes:        notes,
		CreatedAt:    ex.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message, field string) {
	writeJSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Field:   field,
		},
	})
}

func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	raw := chi.URLParam(r, key)
	return uuid.Parse(raw)
}

func parseIntQuery(r *http.Request, key string, defaultVal int32) int32 {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return defaultVal
	}
	return int32(v)
}

// Create handles POST /api/v1/exercises.
func (h *Exercise) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		TargetMuscle string `json:"target_muscle"`
		Notes        string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
		return
	}

	ex, err := h.svc.CreateExercise(r.Context(), req.Name, req.TargetMuscle, req.Notes)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			status := http.StatusInternalServerError
			switch appErr.Code {
			case "VALIDATION_ERROR":
				status = http.StatusUnprocessableEntity
			case "CONFLICT":
				status = http.StatusConflict
			}
			writeError(w, status, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusCreated, exerciseToResponse(ex))
}

// List handles GET /api/v1/exercises.
func (h *Exercise) List(w http.ResponseWriter, r *http.Request) {
	limit := parseIntQuery(r, "limit", 20)
	offset := parseIntQuery(r, "offset", 0)
	search := r.URL.Query().Get("search")

	result, err := h.svc.ListExercises(r.Context(), limit, offset, search)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			writeError(w, http.StatusBadRequest, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	resp := ExerciseListResponse{
		Data: make([]ExerciseResponse, len(result.Data)),
		Pagination: Pagination{
			Limit:  result.Limit,
			Offset: result.Offset,
			Total:  result.Total,
		},
	}
	for i, ex := range result.Data {
		resp.Data[i] = exerciseToResponse(ex)
	}

	writeJSON(w, http.StatusOK, resp)
}

// Get handles GET /api/v1/exercises/:id.
func (h *Exercise) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid exercise id", "")
		return
	}

	ex, err := h.svc.GetExercise(r.Context(), id)
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

	writeJSON(w, http.StatusOK, exerciseToResponse(ex))
}

// Update handles PUT /api/v1/exercises/:id.
func (h *Exercise) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid exercise id", "")
		return
	}

	var req struct {
		Name         string `json:"name"`
		TargetMuscle string `json:"target_muscle"`
		Notes        string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
		return
	}

	ex, err := h.svc.UpdateExercise(r.Context(), id, req.Name, req.TargetMuscle, req.Notes)
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

	writeJSON(w, http.StatusOK, exerciseToResponse(ex))
}

// Delete handles DELETE /api/v1/exercises/:id.
func (h *Exercise) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid exercise id", "")
		return
	}

	if err := h.svc.DeleteExercise(r.Context(), id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			status := http.StatusInternalServerError
			switch appErr.Code {
			case "NOT_FOUND":
				status = http.StatusNotFound
			case "REFERENCED_RESOURCE":
				status = http.StatusConflict
			}
			writeError(w, status, appErr.Code, appErr.Message, "")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
