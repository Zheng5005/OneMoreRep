package handler

import (
	"context"
	"net/http"

	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
)

type QuoteService interface {
	GetRandomQuote(ctx context.Context) (queries.Quote, error)
	ListQuotes(ctx context.Context, limit, offset int32, category string) (service.QuoteListResult, error)
}

type QuoteResponse struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Author   string `json:"author"`
	Category string `json:"category,omitempty"`
}

type QuoteListResponse struct {
	Data       []QuoteResponse `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

type Quote struct {
	svc QuoteService
}

func NewQuote(svc QuoteService) *Quote {
	return &Quote{svc: svc}
}

func quoteToResponse(q queries.Quote) QuoteResponse {
	author := ""
	if q.Author.Valid {
		author = q.Author.String
	}
	category := ""
	if q.Category.Valid {
		category = q.Category.String
	}
	return QuoteResponse{
		ID:       q.ID.String(),
		Text:     q.Text,
		Author:   author,
		Category: category,
	}
}

func (h *Quote) Random(w http.ResponseWriter, r *http.Request) {
	quote, err := h.svc.GetRandomQuote(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}
	writeJSON(w, http.StatusOK, quoteToResponse(quote))
}

func (h *Quote) List(w http.ResponseWriter, r *http.Request) {
	limit := parseIntQuery(r, "limit", 20)
	offset := parseIntQuery(r, "offset", 0)
	category := r.URL.Query().Get("category")

	result, err := h.svc.ListQuotes(r.Context(), limit, offset, category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
		return
	}

	resp := QuoteListResponse{
		Data: make([]QuoteResponse, len(result.Data)),
		Pagination: Pagination{
			Limit:  result.Limit,
			Offset: result.Offset,
			Total:  result.Total,
		},
	}
	for i, q := range result.Data {
		resp.Data[i] = quoteToResponse(q)
	}

	writeJSON(w, http.StatusOK, resp)
}
