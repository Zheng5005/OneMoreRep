package service

import (
	"context"

	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type FallbackQuote struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

var DefaultQuote = FallbackQuote{
	Text:   "The only bad workout is the one that didn't happen.",
	Author: "Unknown",
}

type QuoteStore interface {
	GetRandomQuote(ctx context.Context) (queries.Quote, error)
	ListQuotesPaginated(ctx context.Context, arg queries.ListQuotesPaginatedParams) ([]queries.Quote, error)
	CountQuotes(ctx context.Context, category string) (int64, error)
}

type QuoteListResult struct {
	Data   []queries.Quote
	Limit  int32
	Offset int32
	Total  int64
}

type QuoteService struct {
	store QuoteStore
}

func NewQuoteService(store QuoteStore) *QuoteService {
	return &QuoteService{store: store}
}

func (s *QuoteService) GetRandomQuote(ctx context.Context) (queries.Quote, error) {
	quote, err := s.store.GetRandomQuote(ctx)
	if err == pgx.ErrNoRows {
		return queries.Quote{
			Text:   DefaultQuote.Text,
			Author: pgtype.Text{String: DefaultQuote.Author, Valid: true},
		}, nil
	}
	return quote, err
}

func (s *QuoteService) ListQuotes(ctx context.Context, limit, offset int32, category string) (QuoteListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	quotes, err := s.store.ListQuotesPaginated(ctx, queries.ListQuotesPaginatedParams{
		Column1: category,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return QuoteListResult{}, err
	}

	total, err := s.store.CountQuotes(ctx, category)
	if err != nil {
		return QuoteListResult{}, err
	}

	return QuoteListResult{
		Data:   quotes,
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}, nil
}
