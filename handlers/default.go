package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bricktsre/tftcalculator/components"
	"github.com/bricktsre/tftcalculator/services"
	"github.com/bricktsre/tftcalculator/session"
	"golang.org/x/exp/slog"
)

type CalculatorService interface {
	Calculate(ctx context.Context, sessionID string, level, tier, goal, copiesOwned, tierOwned int) (counts services.Counts, err error)
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
	props.LevelOptions = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

	h.View(w, r, props)
}

func (h *DefaultHandler) Post(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var level, tier, goal, copiesOwned, tierOwned int
	var err error
	errorOccurred := false

	if r.Form.Has("levelDropdown") {
		level, err = strconv.Atoi(r.Form.Get("levelDropdown"))
		if err != nil {
			h.Log.Error("Error parsing level dropdown value")
			errorOccurred = true
		}
	}

	if r.Form.Has("tierDropdown") {
		tier, err = strconv.Atoi(r.Form.Get("tierDropdown"))
		if err != nil {
			h.Log.Error("Error parsing tier dropdown value")
			errorOccurred = true
		}
	}

	if r.Form.Has("goalDropdown") {
		goal, err = strconv.Atoi(r.Form.Get("goalDropdown"))
		if err != nil {
			h.Log.Error("Error parsing goal dropdown value")
			errorOccurred = true
		}
	}

	if r.Form.Has("copiesOwned") {
		copiesOwned, err = strconv.Atoi(r.Form.Get(("copiesOwned")))
		if err != nil {
			h.Log.Error("Error parsing copies owned input")
			errorOccurred = true
		}
	}

	if r.Form.Has("tierOwned") {
		tierOwned, err = strconv.Atoi(r.Form.Get(("tierOwned")))
		if err != nil {
			h.Log.Error("Error parsing tier owned input")
			errorOccurred = true
		}
	}

	var counts services.Counts
	if !errorOccurred {
		counts, err = h.CalculatorService.Calculate(r.Context(), session.ID(r), level, tier, goal, copiesOwned, tierOwned)
		if err != nil {
			h.Log.Error("failed to calculate", slog.Any("error", err))
			http.Error(w, "failed to calculate", http.StatusInternalServerError)
			return
		}
	} else {
		counts.ExpectedRolls = 0
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
	components.Page(props.Counts.ExpectedRolls, props.LevelOptions).Render(r.Context(), w)
}
