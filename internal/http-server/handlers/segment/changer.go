package segment

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
	"net/http"
)

type Request struct {
	Slug string `json:"slug" validate:"required"`
}

type Response struct{}

type SegmentChanger interface {
	CreateSegment(slug string) error
	DeleteSegment(slug string) error
}

func Create(log *slog.Logger, changer SegmentChanger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("decoding request failed", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{})
			return
		}

		log.Info("request body decoded", slog.Any("request", req))
		err = changer.CreateSegment(req.Slug)
		if err != nil {
			log.Error("creating segment failed", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{})
			return
		}

		log.Info("segment created", slog.Any("slug", req.Slug))
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{})
	}
}

func Delete(log *slog.Logger, changer SegmentChanger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		slug := chi.URLParam(r, "slug")

		err := changer.DeleteSegment(slug)
		if err != nil {
			log.Error("deleting segment failed", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{})
			return
		}

		log.Info("segment deleted", slog.Any("slug", slug))
		render.JSON(w, r, Response{})
	}
}
