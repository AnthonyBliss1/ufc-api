package pkg

import (
	"context"
	"net/http"

	apiErrors "github.com/anthonybliss1/ufc-api/api/api_errors"
	"github.com/anthonybliss1/ufc-api/api/db"
	"github.com/anthonybliss1/ufc-api/scrape/data"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ctxKey string

const (
	CtxFightKey         ctxKey = "fight"
	CtxFighterKey       ctxKey = "fighter"
	CtxEventKey         ctxKey = "event"
	CtxUpcomingEventKey ctxKey = "upcomingEvent"
	CtxUpcomingFightKey ctxKey = "upcomingFight"
)

func FightCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "fightID")
		if id == "" {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}

		var f data.Fight
		err := db.MongoDB.Collection("fights").FindOne(r.Context(), bson.M{"_id": id}).Decode(&f)
		if err != nil {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), CtxFightKey, &f)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FighterCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "fighterID")
		if id == "" {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}
		var f data.Fighter
		if err := db.MongoDB.Collection("fighters").FindOne(r.Context(), bson.M{"_id": id}).Decode(&f); err != nil {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), CtxFighterKey, &f)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func EventCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "eventID")
		if id == "" {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}
		var e data.Event
		if err := db.MongoDB.Collection("events").FindOne(r.Context(), bson.M{"_id": id}).Decode(&e); err != nil {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), CtxEventKey, &e)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UpcomingEventCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "upcomingEventID")
		if id == "" {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}
		var e data.UpcomingEvent
		if err := db.MongoDB.Collection("upcomingEvents").FindOne(r.Context(), bson.M{"_id": id}).Decode(&e); err != nil {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), CtxUpcomingEventKey, &e)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UpcomingFightCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "upcomingFightID")
		if id == "" {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}

		var f data.UpcomingFight
		err := db.MongoDB.Collection("upcomingFights").FindOne(r.Context(), bson.M{"_id": id}).Decode(&f)
		if err != nil {
			render.Render(w, r, apiErrors.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), CtxUpcomingFightKey, &f)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
