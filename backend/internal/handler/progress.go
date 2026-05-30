package handler

import (
	"context"
	"net/http"

	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/google/uuid"
)

type ProgressService interface {
	GetExerciseLastValues(ctx context.Context, exerciseID uuid.UUID) (*service.LastValues, error)
	GetSessionSummary(ctx context.Context, sessionID uuid.UUID) (*service.SessionSummary, error)
	GetExerciseHistory(ctx context.Context, exerciseID uuid.UUID, filter string) (*service.ExerciseHistory, error)
	GetVolumeAggregation(ctx context.Context, groupBy string, exerciseID *uuid.UUID) ([]service.VolumePeriod, error)
}

type Progress struct {
	svc ProgressService
}

func NewProgress(svc ProgressService) *Progress {
	return &Progress{svc: svc}
}

type LastValuesResponse struct {
	Weight *float64 `json:"weight,omitempty"`
	Reps   *float64 `json:"reps,omitempty"`
}

func lastValuesToResponse(lv *service.LastValues) LastValuesResponse {
	if lv == nil {
		return LastValuesResponse{}
	}
	return LastValuesResponse{
		Weight: lv.Weight,
		Reps:   lv.Reps,
	}
}

func (h *Progress) GetLastValues(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid exercise id", "")
		return
	}

	result, err := h.svc.GetExerciseLastValues(r.Context(), id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			writeError(w, http.StatusInternalServerError, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusOK, lastValuesToResponse(result))
}

func (h *Progress) GetSessionSummary(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid session id", "")
		return
	}

	result, err := h.svc.GetSessionSummary(r.Context(), id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			if appErr.Code == "NOT_FOUND" {
				writeError(w, http.StatusNotFound, appErr.Code, appErr.Message, "")
				return
			}
			writeError(w, http.StatusInternalServerError, appErr.Code, appErr.Message, "")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Progress) GetExerciseHistory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid exercise id", "")
		return
	}

	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	result, err := h.svc.GetExerciseHistory(r.Context(), id, filter)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			if appErr.Code == "BAD_REQUEST" {
				writeError(w, http.StatusBadRequest, appErr.Code, appErr.Message, "")
				return
			}
			writeError(w, http.StatusInternalServerError, appErr.Code, appErr.Message, "")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Progress) GetVolumeAggregation(w http.ResponseWriter, r *http.Request) {
	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "group_by is required", "")
		return
	}

	var exerciseID *uuid.UUID
	if exerciseIDStr := r.URL.Query().Get("exercise_id"); exerciseIDStr != "" {
		id, err := uuid.Parse(exerciseIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid exercise_id", "")
			return
		}
		exerciseID = &id
	}

	result, err := h.svc.GetVolumeAggregation(r.Context(), groupBy, exerciseID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			if appErr.Code == "BAD_REQUEST" {
				writeError(w, http.StatusBadRequest, appErr.Code, appErr.Message, "")
				return
			}
			writeError(w, http.StatusInternalServerError, appErr.Code, appErr.Message, "")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusOK, result)
}