package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/google/uuid"
)

type SessionService interface {
	CreateSession(ctx context.Context, routineID *uuid.UUID) (queries.WorkoutSession, error)
	GetSession(ctx context.Context, id uuid.UUID) (service.SessionDetail, error)
	GetActiveSession(ctx context.Context) (*service.SessionDetail, error)
	EndSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
}

type SetResponse struct {
	ID           uuid.UUID `json:"id"`
	SessionID    uuid.UUID `json:"session_id"`
	ExerciseID   uuid.UUID `json:"exercise_id"`
	SetNumber    int32     `json:"set_number"`
	Weight       float64   `json:"weight"`
	Reps         int32     `json:"reps"`
	CreatedAt    string    `json:"created_at"`
	ExerciseName string    `json:"exercise_name,omitempty"`
}

type SessionResponse struct {
	ID        uuid.UUID    `json:"id"`
	RoutineID *uuid.UUID   `json:"routine_id,omitempty"`
	StartedAt string       `json:"started_at"`
	EndedAt   *string      `json:"ended_at,omitempty"`
	Sets      []SetResponse `json:"sets"`
}

type Session struct {
	svc SessionService
}

func NewSession(svc SessionService) *Session {
	return &Session{svc: svc}
}

func sessionToResponse(s queries.WorkoutSession) SessionResponse {
	var routineID *uuid.UUID
	if s.RoutineID.Valid {
		id := uuid.UUID(s.RoutineID.Bytes)
		routineID = &id
	}
	var endedAt *string
	if s.EndedAt.Valid {
		t := s.EndedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		endedAt = &t
	}
	return SessionResponse{
		ID:        s.ID,
		RoutineID: routineID,
		StartedAt: s.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		EndedAt:   endedAt,
		Sets:      []SetResponse{},
	}
}

func sessionDetailToResponse(d service.SessionDetail) SessionResponse {
	resp := sessionToResponse(queries.WorkoutSession{
		ID:        d.Session.ID,
		RoutineID: d.Session.RoutineID,
		StartedAt: d.Session.StartedAt,
		EndedAt:   d.Session.EndedAt,
	})

	setsRaw, ok := d.Session.Sets.([]interface{})
	if !ok {
		return resp
	}

	sets := make([]SetResponse, 0, len(setsRaw))
	for _, s := range setsRaw {
		setMap, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		setResp := SetResponse{}
		if id, ok := setMap["id"].(string); ok {
			setResp.ID, _ = uuid.Parse(id)
		}
		if sid, ok := setMap["session_id"].(string); ok {
			setResp.SessionID, _ = uuid.Parse(sid)
		}
		if eid, ok := setMap["exercise_id"].(string); ok {
			setResp.ExerciseID, _ = uuid.Parse(eid)
		}
		if sn, ok := setMap["set_number"].(float64); ok {
			setResp.SetNumber = int32(sn)
		}
		if w, ok := setMap["weight"].(float64); ok {
			setResp.Weight = w
		}
		if rp, ok := setMap["reps"].(float64); ok {
			setResp.Reps = int32(rp)
		}
		if ca, ok := setMap["created_at"].(string); ok {
			setResp.CreatedAt = ca
		}
		if en, ok := setMap["exercise_name"].(string); ok {
			setResp.ExerciseName = en
		}
		sets = append(sets, setResp)
	}
	resp.Sets = sets
	return resp
}

func (h *Session) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RoutineID *string `json:"routine_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
		return
	}

	var routineID *uuid.UUID
	if req.RoutineID != nil && *req.RoutineID != "" {
		id, err := uuid.Parse(*req.RoutineID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routine_id", "")
			return
		}
		routineID = &id
	}

	session, err := h.svc.CreateSession(r.Context(), routineID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			writeError(w, http.StatusInternalServerError, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusCreated, sessionToResponse(session))
}

func (h *Session) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid session id", "")
		return
	}

	detail, err := h.svc.GetSession(r.Context(), id)
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

	writeJSON(w, http.StatusOK, sessionDetailToResponse(detail))
}

func (h *Session) GetActive(w http.ResponseWriter, r *http.Request) {
	detail, err := h.svc.GetActiveSession(r.Context())
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			writeError(w, http.StatusInternalServerError, appErr.Code, appErr.Message, appErr.Field)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	if detail == nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}

	writeJSON(w, http.StatusOK, sessionDetailToResponse(*detail))
}

func (h *Session) End(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid session id", "")
		return
	}

	session, err := h.svc.EndSession(r.Context(), id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			status := http.StatusInternalServerError
			switch appErr.Code {
			case "NOT_FOUND":
				status = http.StatusNotFound
			case "CONFLICT":
				status = http.StatusConflict
			}
			writeError(w, status, appErr.Code, appErr.Message, "")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	writeJSON(w, http.StatusOK, sessionToResponse(session))
}