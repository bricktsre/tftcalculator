package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bricktsre/tftcalculator/handlers"
	"github.com/bricktsre/tftcalculator/services"
	"github.com/bricktsre/tftcalculator/session"
	"golang.org/x/exp/slog"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cs := services.NewCalculator(log)
	h := handlers.New(log, cs)

	var secureFlag = true
	if os.Getenv("SECURE_FLAG") == "false" {
		secureFlag = false
	}

	sh := session.NewMiddleware(h, session.WithSecure(secureFlag))

	server := &http.Server{
		Addr:         "localhost:9000",
		Handler:      sh,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	fmt.Printf("Listening on %v\n", server.Addr)
	server.ListenAndServe()
}
