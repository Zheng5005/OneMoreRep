package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockQuoteStore struct {
	getRandomFunc        func(ctx context.Context) (queries.Quote, error)
	listPaginatedFunc     func(ctx context.Context, arg queries.ListQuotesPaginatedParams) ([]queries.Quote, error)
	countFunc             func(ctx context.Context, category string) (int64, error)
}

func (m *mockQuoteStore) GetRandomQuote(ctx context.Context) (queries.Quote, error) {
	return m.getRandomFunc(ctx)
}

func (m *mockQuoteStore) ListQuotesPaginated(ctx context.Context, arg queries.ListQuotesPaginatedParams) ([]queries.Quote, error) {
	return m.listPaginatedFunc(ctx, arg)
}

func (m *mockQuoteStore) CountQuotes(ctx context.Context, category string) (int64, error) {
	return m.countFunc(ctx, category)
}

func TestQuoteServiceGetRandomQuote(t *testing.T) {
	tests := []struct {
		name     string
		setupMock func() *mockQuoteStore
		wantErr  bool
		wantText string
		wantAuth string
	}{
		{
			name: "returns quote from database",
			setupMock: func() *mockQuoteStore {
				return &mockQuoteStore{
					getRandomFunc: func(_ context.Context) (queries.Quote, error) {
						return queries.Quote{
							Text:   "The last three reps",
							Author: pgtype.Text{String: "Arnold", Valid: true},
						}, nil
					},
				}
			},
			wantErr:  false,
			wantText: "The last three reps",
			wantAuth: "Arnold",
		},
		{
			name: "returns fallback when no rows",
			setupMock: func() *mockQuoteStore {
				return &mockQuoteStore{
					getRandomFunc: func(_ context.Context) (queries.Quote, error) {
						return queries.Quote{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  false,
			wantText: "The only bad workout is the one that didn't happen.",
			wantAuth: "Unknown",
		},
		{
			name: "forwards error from store",
			setupMock: func() *mockQuoteStore {
				return &mockQuoteStore{
					getRandomFunc: func(_ context.Context) (queries.Quote, error) {
						return queries.Quote{}, errors.New("connection failed")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := NewQuoteService(mock)

			quote, err := svc.GetRandomQuote(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if quote.Text != tt.wantText {
					t.Errorf("text = %q, want %q", quote.Text, tt.wantText)
				}
				if quote.Author.String != tt.wantAuth {
					t.Errorf("author = %q, want %q", quote.Author.String, tt.wantAuth)
				}
			}
		})
	}
}

func TestQuoteServiceListQuotes(t *testing.T) {
	tests := []struct {
		name        string
		limit       int32
		offset      int32
		category    string
		setupMock   func() *mockQuoteStore
		wantErr     bool
		wantCount   int
		wantTotal   int64
		wantLimit   int32
		wantOffset  int32
		wantCatFilter bool
	}{
		{
			name:     "returns paginated quotes",
			limit:    10,
			offset:   0,
			category: "",
			setupMock: func() *mockQuoteStore {
				return &mockQuoteStore{
					listPaginatedFunc: func(_ context.Context, arg queries.ListQuotesPaginatedParams) ([]queries.Quote, error) {
						return []queries.Quote{
							{Text: "Quote 1", Author: pgtype.Text{String: "Author 1", Valid: true}},
							{Text: "Quote 2", Author: pgtype.Text{String: "Author 2", Valid: true}},
						}, nil
					},
					countFunc: func(_ context.Context, _ string) (int64, error) {
						return 20, nil
					},
				}
			},
			wantErr:    false,
			wantCount:  2,
			wantTotal:  20,
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:     "filters by category",
			limit:    20,
			offset:   0,
			category: "strength",
			setupMock: func() *mockQuoteStore {
				return &mockQuoteStore{
					listPaginatedFunc: func(_ context.Context, arg queries.ListQuotesPaginatedParams) ([]queries.Quote, error) {
						if arg.Column1 != "strength" {
							t.Errorf("expected category=strength, got %q", arg.Column1)
						}
						return []queries.Quote{
							{Text: "Strength quote", Author: pgtype.Text{String: "Coach", Valid: true}},
						}, nil
					},
					countFunc: func(_ context.Context, cat string) (int64, error) {
						if cat != "strength" {
							t.Errorf("expected category=strength, got %q", cat)
						}
						return 5, nil
					},
				}
			},
			wantErr:      false,
			wantCount:    1,
			wantTotal:    5,
			wantLimit:    20,
			wantOffset:   0,
			wantCatFilter: true,
		},
		{
			name:     "applies default limit and offset",
			limit:    0,
			offset:   -1,
			category: "",
			setupMock: func() *mockQuoteStore {
				return &mockQuoteStore{
					listPaginatedFunc: func(_ context.Context, arg queries.ListQuotesPaginatedParams) ([]queries.Quote, error) {
						if arg.Limit != 20 {
							t.Errorf("expected limit=20, got %d", arg.Limit)
						}
						if arg.Offset != 0 {
							t.Errorf("expected offset=0, got %d", arg.Offset)
						}
						return []queries.Quote{}, nil
					},
					countFunc: func(_ context.Context, _ string) (int64, error) {
						return 0, nil
					},
				}
			},
			wantErr:    false,
			wantCount:  0,
			wantTotal:  0,
			wantLimit:  20,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := NewQuoteService(mock)

			result, err := svc.ListQuotes(context.Background(), tt.limit, tt.offset, tt.category)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(result.Data) != tt.wantCount {
					t.Errorf("got %d quotes, want %d", len(result.Data), tt.wantCount)
				}
				if result.Total != tt.wantTotal {
					t.Errorf("total = %d, want %d", result.Total, tt.wantTotal)
				}
				if result.Limit != tt.wantLimit {
					t.Errorf("limit = %d, want %d", result.Limit, tt.wantLimit)
				}
				if result.Offset != tt.wantOffset {
					t.Errorf("offset = %d, want %d", result.Offset, tt.wantOffset)
				}
			}
		})
	}
}
