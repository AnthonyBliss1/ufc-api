package db

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var MongoDB *mongo.Database

func InitMongo() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping mongo: %v", err)
	}

	MongoDB = client.Database("ufc")
}

func Paginator(r *http.Request, defaultLimit int64) (skip, limit int64) {
	q := r.URL.Query()
	page := int64(1)
	limit = defaultLimit

	if v := q.Get("page"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			page = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= 50 { // set max to 50 entries
			limit = n
		}
	}

	skip = (page - 1) * limit
	return
}

func RenderJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	render.Status(r, 200)
	render.JSON(w, r, v)
}
