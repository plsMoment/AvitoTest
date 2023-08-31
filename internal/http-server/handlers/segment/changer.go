package segment

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
	"net/http"
)

type Request struct {
	Slug string `json:"slug" validate:"required"`
}

type Response struct {
	Status int `json:"status"`
}

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
			log.Error("failed decoding request", err)
			render.JSON(w, r, Response{Status: http.StatusInternalServerError})
			return
		}

		log.Info("request body decoded", slog.Any("request", req))
		err = changer.CreateSegment(req.Slug)
		if err != nil {
			log.Error("creating segment failed:", err)
			render.JSON(w, r, Response{Status: http.StatusInternalServerError})
			return
		}

		log.Info("segment created, slug: ", req.Slug)
		render.JSON(w, r, Response{Status: http.StatusOK})
	}
}

func Delete(log *slog.Logger, changer SegmentChanger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed decoding request", err)
			render.JSON(w, r, Response{Status: http.StatusInternalServerError})
			return
		}

		log.Info("request body decoded", slog.Any("request", req))
		err = changer.DeleteSegment(req.Slug)
		if err != nil {
			log.Error("deleting segment failed: ", err)
			render.JSON(w, r, Response{Status: http.StatusInternalServerError})
			return
		}

		log.Info("segment deleted, slug: ", req.Slug)
		render.JSON(w, r, Response{Status: http.StatusOK})
	}
}
