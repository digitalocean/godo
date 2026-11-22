package sandbox

import (
	"context"
	"errors"
	"iter"
)

// Pages reads a list one page at a time with NextItems, or all of it with All.
// Nothing is fetched until you read a page.
type Pages[T any] struct {
	fetch   func(ctx context.Context, token string) (items []T, next string, err error)
	token   string
	started bool
}

// HasNext reports whether another page may exist: true until a page came back
// without a next-page token.
func (p *Pages[T]) HasNext() bool { return !p.started || p.token != "" }

// NextToken is the token that resumes after the last page read. Empty on the
// last page.
func (p *Pages[T]) NextToken() string { return p.token }

// NextItems reads the next page.
func (p *Pages[T]) NextItems(ctx context.Context) ([]T, error) {
	if !p.HasNext() {
		return nil, errors.New("sandbox: no more items to list")
	}
	items, next, err := p.fetch(ctx, p.token)
	if err != nil {
		return nil, err
	}
	p.started = true
	p.token = next
	return items, nil
}

// All yields every remaining item, fetching pages as it goes. A failed page is
// yielded as the error, and the sequence ends.
func (p *Pages[T]) All(ctx context.Context) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for p.HasNext() {
			page, err := p.NextItems(ctx)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range page {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
