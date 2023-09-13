package main

import (
	"AvitoTest/internal/config"
	"AvitoTest/internal/database/models"
	"AvitoTest/internal/http-server/handlers/segment"
	"AvitoTest/internal/http-server/handlers/userSegment"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	log.Info("startup AvitoTest", slog.String("env", cfg.Env))

	db, err := models.New(&cfg.DB)
	if err != nil {
		log.Error("connect to database failed: ", err)
		os.Exit(1)
	}
	defer db.Close()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/segment", segment.Create(log, db))
	router.Delete("/segment/{slug}", segment.Delete(log, db))
	router.Get("/segments/user/{userId}", userSegment.Get(log, db))
	router.Put("/segments/user/{userId}", userSegment.Update(log, db))

	if cfg.Env == "local" {
		log.Info("starting server", slog.String("address", "localhost:8080"))
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("server closed")
			} else {
				log.Error("error during server shutdown", err)
			}
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", err)
		return
	}

}
