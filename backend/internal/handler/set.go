package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/google/uuid"
)

type SetService interface {
	CreateSet(ctx context.Context, sessionID uuid.UUID, exerciseID uuid.UUID, weight float64, reps int) (queries.WorkoutSet, error)
	UpdateSet(ctx context.Context, setID uuid.UUID, weight float64, reps int) (queries.WorkoutSet, error)
	DeleteSet(ctx context.Context, setID uuid.UUID) error
}

type SetHandler struct {
	svc SetService
}

func NewSet(svc SetService) *SetHandler {
	return &SetHandler{svc: svc}
}

type CreateSetRequest struct {
	SessionID  string  `json:"session_id"`
	ExerciseID string  `json:"exercise_id"`
	Weight     float64 `json:"weight"`
	Reps       int     `json:"reps"`
}

type UpdateSetRequest struct {
	Weight float64 `json:"weight"`
	Reps   int     `json:"reps"`
}

func setToResponse(s queries.WorkoutSet) SetResponse {
	w, _ := s.Weight.Float64Value()
	return SetResponse{
		ID:         s.ID,
		SessionID:  s.SessionID,
		ExerciseID: s.ExerciseID,
		SetNumber:  s.SetNumber,
		Weight:     w.Float64,
		Reps:       s.Reps,
		CreatedAt:  s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *SetHandler) Create(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid session id", "")
		return
	}

	var req CreateSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
		return
	}

	exerciseID, err := uuid.Parse(req.ExerciseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid exercise_id", "")
		return
	}

	workoutSet, err := h.svc.CreateSet(r.Context(), sessionID, exerciseID, req.Weight, req.Reps)
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

	writeJSON(w, http.StatusCreated, setToResponse(workoutSet))
}

func (h *SetHandler) Update(w http.ResponseWriter, r *http.Request) {
	setID, err := parseUUIDParam(r, "setId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid set id", "")
		return
	}

	var req UpdateSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
		return
	}

	workoutSet, err := h.svc.UpdateSet(r.Context(), setID, req.Weight, req.Reps)
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

	writeJSON(w, http.StatusOK, setToResponse(workoutSet))
}

func (h *SetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	setID, err := parseUUIDParam(r, "setId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid set id", "")
		return
	}

	if err := h.svc.DeleteSet(r.Context(), setID); err != nil {
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