package userSegment

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"net/http"
)

type Request struct {
	AddSlugs    []string `json:"add_slugs" validate:"required"`
	DeleteSlugs []string `json:"delete_slugs" validate:"required"`
}

type UserSegmentChanger interface {
	ChangeUserSegments(userId uuid.UUID, addSlugs []string, deleteSlugs []string) error
}

func Update(log *slog.Logger, changer UserSegmentChanger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userId, err := uuid.Parse(chi.URLParam(r, "userId"))
		if err != nil {
			log.Error("parsing URL parameter failed", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{})
			return
		}

		var req Request
		err = render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("decoding request failed", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{})
			return
		}

		log.Info("request body decoded", slog.Any("request", req))
		err = changer.ChangeUserSegments(userId, req.AddSlugs, req.DeleteSlugs)
		if err != nil {
			log.Error("changing user's segments failed", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{})
			return
		}

		render.JSON(w, r, Response{})
	}
}
