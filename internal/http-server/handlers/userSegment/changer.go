package userSegment

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"net/http"
)

type Request struct {
	UserId      uuid.UUID `json:"user_id" validate:"required,uuid"`
	AddSlugs    []string  `json:"add_slugs" validate:"required"`
	DeleteSlugs []string  `json:"delete_slugs" validate:"required"`
}

type UserSegmentChanger interface {
	ChangeUserSegments(userId uuid.UUID, addSlugs []string, deleteSlugs []string) error
}

func Update(log *slog.Logger, changer UserSegmentChanger) http.HandlerFunc {
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
		err = changer.ChangeUserSegments(req.UserId, req.AddSlugs, req.DeleteSlugs)
		if err != nil {
			log.Error("creating segment failed: ", err)
			render.JSON(w, r, Response{Status: http.StatusInternalServerError})
			return
		}

		render.JSON(w, r, Response{Status: http.StatusOK})
	}
}
