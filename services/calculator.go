package services

import (
	"context"
	"errors"

	"golang.org/x/exp/slog"
)

type Counts struct {
	Global  int
	Session int
}

func NewCalculator(log *slog.Logger) Calculator {
	return Calculator{
		Log: log,
	}
}

type Calculator struct {
	Log *slog.Logger
}

func (cs Calculator) Increment(ctx context.Context, sessionID string) (counts Counts, err error) {
	errs := make([]error, 2)
	counts.Global = 5
	counts.Session = 1

	return counts, errors.Join(errs...)
}

func (cs Calculator) Get(ctx context.Context, sessionID string) (counts Counts, err error) {
	counts.Global = 1
	counts.Session = 2
	return
}
