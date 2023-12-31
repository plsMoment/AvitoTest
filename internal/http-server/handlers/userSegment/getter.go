package userSegment

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"net/http"
)

type Response struct {
	Slugs []string `json:"slugs,omitempty"`
}

type UserSegmentsGetter interface {
	GetUserSegments(userId uuid.UUID) ([]string, error)
}

func Get(log *slog.Logger, getter UserSegmentsGetter) http.HandlerFunc {
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

		slugs, err := getter.GetUserSegments(userId)
		if err != nil {
			log.Error("getting user's segment failed", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{})
			return
		}

		render.JSON(w, r, Response{Slugs: slugs})
	}
}
