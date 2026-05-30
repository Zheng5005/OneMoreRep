package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
)

type mockQuoteService struct {
	getRandomFunc func(ctx context.Context) (queries.Quote, error)
	listFunc      func(ctx context.Context, limit, offset int32, category string) (service.QuoteListResult, error)
}

func (m *mockQuoteService) GetRandomQuote(ctx context.Context) (queries.Quote, error) {
	return m.getRandomFunc(ctx)
}

func (m *mockQuoteService) ListQuotes(ctx context.Context, limit, offset int32, category string) (service.QuoteListResult, error) {
	return m.listFunc(ctx, limit, offset, category)
}

func TestQuoteHandlerRandom(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func() *mockQuoteService
		wantStatus int
		wantText   string
		wantAuthor string
	}{
		{
			name: "happy path",
			setupMock: func() *mockQuoteService {
				return &mockQuoteService{
					getRandomFunc: func(_ context.Context) (queries.Quote, error) {
						return queries.Quote{
							ID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
							Text:   "The last three reps",
							Author: pgtype.Text{String: "Arnold", Valid: true},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
			wantText:   "The last three reps",
			wantAuthor: "Arnold",
		},
		{
			name: "internal error",
			setupMock: func() *mockQuoteService {
				return &mockQuoteService{
					getRandomFunc: func(_ context.Context) (queries.Quote, error) {
						return queries.Quote{}, errors.New("connection failed")
					},
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewQuote(mock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/quotes/random", nil)
			rec := httptest.NewRecorder()

			h.Random(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantText != "" {
				var resp QuoteResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Text != tt.wantText {
					t.Errorf("text = %q, want %q", resp.Text, tt.wantText)
				}
				if resp.Author != tt.wantAuthor {
					t.Errorf("author = %q, want %q", resp.Author, tt.wantAuthor)
				}
			}
		})
	}
}

func TestQuoteHandlerList(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		setupMock    func() *mockQuoteService
		wantStatus   int
		wantCount    int
		wantTotal    int64
		wantLimit    int32
		wantOffset   int32
		wantCategory string
	}{
		{
			name: "happy path",
			url:  "/api/v1/quotes?limit=10&offset=0",
			setupMock: func() *mockQuoteService {
				return &mockQuoteService{
					listFunc: func(_ context.Context, limit, offset int32, category string) (service.QuoteListResult, error) {
						return service.QuoteListResult{
							Data: []queries.Quote{
								{Text: "Quote 1", Author: pgtype.Text{String: "Author 1", Valid: true}},
								{Text: "Quote 2", Author: pgtype.Text{String: "Author 2", Valid: true}},
							},
							Limit:  10,
							Offset: 0,
							Total:  20,
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
			wantTotal:  20,
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name: "filter by category",
			url:  "/api/v1/quotes?category=strength",
			setupMock: func() *mockQuoteService {
				return &mockQuoteService{
					listFunc: func(_ context.Context, limit, offset int32, category string) (service.QuoteListResult, error) {
						if category != "strength" {
							t.Errorf("expected category=strength, got %q", category)
						}
						return service.QuoteListResult{
							Data: []queries.Quote{
								{Text: "Strength quote", Author: pgtype.Text{String: "Coach", Valid: true}},
							},
							Limit:  20,
							Offset: 0,
							Total:  5,
						}, nil
					},
				}
			},
			wantStatus:   http.StatusOK,
			wantCount:    1,
			wantTotal:    5,
			wantLimit:    20,
			wantOffset:   0,
			wantCategory: "strength",
		},
		{
			name: "default pagination",
			url:  "/api/v1/quotes",
			setupMock: func() *mockQuoteService {
				return &mockQuoteService{
					listFunc: func(_ context.Context, limit, offset int32, category string) (service.QuoteListResult, error) {
						if limit != 20 {
							t.Errorf("expected limit=20, got %d", limit)
						}
						if offset != 0 {
							t.Errorf("expected offset=0, got %d", offset)
						}
						return service.QuoteListResult{
							Data:   []queries.Quote{},
							Limit:  20,
							Offset: 0,
							Total:  0,
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
			wantLimit: 20,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewQuote(mock)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			h.List(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var resp QuoteListResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if len(resp.Data) != tt.wantCount {
				t.Errorf("got %d quotes, want %d", len(resp.Data), tt.wantCount)
			}
			if resp.Pagination.Total != tt.wantTotal {
				t.Errorf("total = %d, want %d", resp.Pagination.Total, tt.wantTotal)
			}
			if resp.Pagination.Limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", resp.Pagination.Limit, tt.wantLimit)
			}
			if resp.Pagination.Offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", resp.Pagination.Offset, tt.wantOffset)
			}
		})
	}
}
