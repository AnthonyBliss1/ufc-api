package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/anthonybliss1/ufc-api/api/db"
	"github.com/anthonybliss1/ufc-api/api/handlers"
	"github.com/anthonybliss1/ufc-api/api/pkg"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/joho/godotenv"
)

//go:embed .env
var embeddedEnv string

func init() {
	godotenv.Load(".env")
	envMap, err := godotenv.Unmarshal(embeddedEnv)
	if err != nil {
		log.Fatalf("[Failed to unmarshal embeddedEnv file: %v]\n", err)
	}

	for key, value := range envMap {
		os.Setenv(key, value)
	}

	fmt.Print("\n[Environment Variables Set!]\n\n")
}

func main() {
	db.InitMongo()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Timeout(60 * time.Second))

	// defining /fights route with subroute for /fights/{id} and /search endpoint
	r.Route("/fights", func(r chi.Router) {
		r.Get("/", handlers.ListFights)
		r.Get("/search", handlers.SearchFights) // GET /fights/search

		r.Route("/{fightID}", func(r chi.Router) {
			r.Use(pkg.FightCtx)           // Load the *Fight on the request context
			r.Get("/", handlers.GetFight) // GET /fights/123
		})
	})

	// defining /fighters route with subroute for /fighters/{id} and /search
	r.Route("/fighters", func(r chi.Router) {
		r.Get("/", handlers.ListFighters)
		r.Get("/search", handlers.SearchFighters) // GET /fights/search

		r.Route("/{fighterID}", func(r chi.Router) {
			r.Use(pkg.FighterCtx)           // Load the *Fighter on the request context
			r.Get("/", handlers.GetFighter) // GET /fighters/123
		})
	})

	// defining /events route with subroute for /events/{id} and /search
	r.Route("/events", func(r chi.Router) {
		r.Get("/", handlers.ListEvents)
		r.Get("/search", handlers.SearchEvents) // GET /events/search

		r.Route("/{eventID}", func(r chi.Router) {
			r.Use(pkg.EventCtx)           // Load the *Event on the request context
			r.Get("/", handlers.GetEvent) // GET /events/123
		})
	})

	// defining /upcomingEvents route with subroute for /UpcomingEvents/{id} and /search
	r.Route("/upcomingEvents", func(r chi.Router) {
		r.Get("/", handlers.ListUpcomingEvents)
		r.Get("/search", handlers.SearchUpcomingEvents) // GET /upcomingEvents/search

		r.Route("/{upcomingEventID}", func(r chi.Router) {
			r.Use(pkg.UpcomingEventCtx)           // Load the *UpcomingEvent on the request context
			r.Get("/", handlers.GetUpcomingEvent) // GET /upcomingEvents/123
		})
	})

	// defining /upcomingFights route with subroute for /upcomingFights/{id}
	r.Route("/upcomingFights", func(r chi.Router) {
		r.Get("/", handlers.ListUpcomingFights)

		r.Route("/{upcomingFightID}", func(r chi.Router) {
			r.Use(pkg.UpcomingFightCtx)           // Load the *UpcomingFight on the request context
			r.Get("/", handlers.GetUpcomingFight) // GET /upcomingFights/123
		})
	})

	log.Fatal(http.ListenAndServe("0.0.0.0:8000", r))
}
