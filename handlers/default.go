package handlers

import (
	"context"
	"net/http"

	"github.com/bricktsre/tftcalculator/components"
	"github.com/bricktsre/tftcalculator/services"
	"github.com/bricktsre/tftcalculator/session"
	"golang.org/x/exp/slog"
)

type CalculatorService interface {
	Increment(ctx context.Context, sessionID string) (counts services.Counts, err error)
	Get(ctx context.Context, sessionID string) (counts services.Counts, err error)
}

func New(log *slog.Logger, cs CalculatorService) *DefaultHandler {
	return &DefaultHandler{
		Log:               log,
		CalculatorService: cs,
	}
}

type DefaultHandler struct {
	Log               *slog.Logger
	CalculatorService CalculatorService
}

func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.Post(w, r)
		return
	}
	h.Get(w, r)
}

func (h *DefaultHandler) Get(w http.ResponseWriter, r *http.Request) {
	var props ViewProps
	var err error
	props.Counts, err = h.CalculatorService.Get(r.Context(), session.ID(r))
	if err != nil {
		h.Log.Error("failed to get counts", slog.Any("error", err))
		http.Error(w, "failed to get counts", http.StatusInternalServerError)
		return
	}
	h.View(w, r, props)
}

func (h *DefaultHandler) Post(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	counts, err := h.CalculatorService.Increment(r.Context(), session.ID(r))
	if err != nil {
		h.Log.Error("failed to calculate", slog.Any("error", err))
		http.Error(w, "failed to calculate", http.StatusInternalServerError)
		return
	}

	// Display the view.
	h.View(w, r, ViewProps{
		Counts:       counts,
		LevelOptions: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
	})
}

type ViewProps struct {
	Counts       services.Counts
	LevelOptions []string
}

func (h *DefaultHandler) View(w http.ResponseWriter, r *http.Request, props ViewProps) {
	components.Page(props.Counts.Global, props.Counts.Session, props.LevelOptions).Render(r.Context(), w)
}
