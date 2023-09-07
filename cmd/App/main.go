package main

import (
	"AvitoTest/internal/config"
	"AvitoTest/internal/database/models"
	"AvitoTest/internal/http-server/handlers/segment"
	"AvitoTest/internal/http-server/handlers/userSegment"
	"context"
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

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/segment", segment.Create(log, db))
	router.Delete("/segment", segment.Delete(log, db))
	router.Get("/segments/{userId}", userSegment.Get(log, db))
	router.Put("/segments/user", userSegment.Update(log, db))

	log.Info("starting server", slog.String("address", cfg.Address))

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
		srv.ListenAndServe()
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

	log.Info("server stopped")
}
