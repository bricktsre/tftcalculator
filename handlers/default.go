package handlers

import (
	"log/slog"
	"net/http"

	"github.com/bricktsre/dash-tft/components"
	"github.com/bricktsre/dash-tft/services"
	"github.com/bricktsre/dash-tft/session"
)

type XService interface {
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
		Counts: counts,
	})
}

type ViewProps struct {
	Counts services.Counts
}

func (h *DefaultHandler) View(w http.ResponseWriter, r *http.Request, props ViewProps) {
	components.Page(props.Counts.Global, props.Counts.Session).Render(r.Context(), w)
}
