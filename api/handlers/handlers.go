package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anthonybliss1/ufc-api/api/db"
	"github.com/anthonybliss1/ufc-api/api/pkg"
	"github.com/anthonybliss1/ufc-api/scrape/data"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ListFights(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	and := bson.A{}

	if v := q.Get("event_id"); v != "" {
		and = append(and, bson.M{"event_id": v})
	}
	if v := q.Get("referee"); v != "" {
		and = append(and, bson.M{"referee": v})
	}
	if v := q.Get("method"); v != "" {
		and = append(and, bson.M{"method": v})
	}
	if v := q.Get("fighter_id"); v != "" {
		and = append(and, bson.M{"participants.fighter_id": v})
	}
	if names := q["fighter_name"]; len(names) > 0 {
		for _, n := range names {
			and = append(and, bson.M{
				"participants.fighter_name": bson.M{"$regex": n, "$options": "i"},
			})
		}
	}

	filter := bson.M{}
	if len(and) > 0 {
		filter["$and"] = and
	}

	limit := db.LimitFromQuery(r, 50, 50)
	if after := db.AfterFromQuery(r); after != "" {
		filter["_id"] = bson.M{"$gt": after}
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "_id", Value: 1}})

	cur, err := db.MongoDB.Collection("fights").Find(r.Context(), filter, opts)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, "db error")
		return
	}
	defer cur.Close(r.Context())

	var items []data.Fight
	for cur.Next(r.Context()) {
		var f data.Fight
		if err := cur.Decode(&f); err != nil {
			render.Status(r, 500)
			render.PlainText(w, r, "decode error")
			return
		}
		items = append(items, f)
	}

	if n := len(items); n > 0 {
		w.Header().Set("X-Next-After", items[n-1].ID)
	}

	db.CacheFor(w, 30*time.Second)

	db.RenderJSON(w, r, data.Fights{Items: items})
}

func SearchFights(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	and := bson.A{}

	if v := q.Get("q"); v != "" {
		and = append(and, bson.M{"$or": bson.A{
			bson.M{"fight_detail": bson.M{"$regex": v, "$options": "i"}},
			bson.M{"method": bson.M{"$regex": v, "$options": "i"}},
			bson.M{"method_detail": bson.M{"$regex": v, "$options": "i"}},
			bson.M{"referee": bson.M{"$regex": v, "$options": "i"}},
		}})
	}

	filter := bson.M{}
	if len(and) > 0 {
		filter["$and"] = and
	}

	limit := db.LimitFromQuery(r, 50, 50)
	if after := db.AfterFromQuery(r); after != "" {
		filter["_id"] = bson.M{"$gt": after}
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "_id", Value: 1}})

	cur, err := db.MongoDB.Collection("fights").Find(r.Context(), filter, opts)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, "db error")
		return
	}
	defer cur.Close(r.Context())

	var items []data.Fight
	for cur.Next(r.Context()) {
		var f data.Fight
		if err := cur.Decode(&f); err != nil {
			render.Status(r, 500)
			render.PlainText(w, r, "decode error")
			return
		}
		items = append(items, f)
	}

	if n := len(items); n > 0 {
		w.Header().Set("X-Next-After", items[n-1].ID)
	}

	db.CacheFor(w, 30*time.Second)

	db.RenderJSON(w, r, data.Fights{Items: items})
}

func GetFight(w http.ResponseWriter, r *http.Request) {
	f, _ := r.Context().Value(pkg.CtxFightKey).(*data.Fight)
	if f == nil {
		render.Status(r, 404)
		render.PlainText(w, r, "fight not found")
		return
	}
	db.RenderJSON(w, r, f)
}

func ListFighters(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{}
	q := r.URL.Query()

	if v := q.Get("name"); v != "" {
		filter["name"] = bson.M{"$regex": v, "$options": "i"}
	}
	if v := q.Get("stance"); v != "" {
		filter["stance"] = v
	}
	// ?min_slpm=3.0
	if v := q.Get("min_slpm"); v != "" {
		filter["career_stats.slpm"] = bson.M{"$gte": parseFloat32(v)}
	}
	// query param for start and end date range for *Fighter.DOB
	if v := q.Get("start"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter["dob"] = bson.M{"$gte": t}
		}
	}
	if v := q.Get("end"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			if m, ok := filter["dob"].(bson.M); ok {
				m["$lte"] = t
			} else {
				filter["dob"] = bson.M{"$lte": t}
			}
		}
	}

	limit := db.LimitFromQuery(r, 50, 50)
	if after := db.AfterFromQuery(r); after != "" {
		filter["_id"] = bson.M{"$gt": after}
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "_id", Value: 1}})

	cur, err := db.MongoDB.Collection("fighters").Find(r.Context(), filter, opts)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, "db error")
		return
	}
	defer cur.Close(r.Context())

	var items []data.Fighter
	for cur.Next(r.Context()) {
		var f data.Fighter
		if err := cur.Decode(&f); err != nil {
			render.Status(r, 500)
			render.PlainText(w, r, "decode error")
			return
		}
		items = append(items, f)
	}

	if n := len(items); n > 0 {
		w.Header().Set("X-Next-After", items[n-1].ID)
	}

	db.CacheFor(w, 30*time.Second)

	db.RenderJSON(w, r, data.Fighters{Items: items})
}

func SearchFighters(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	and := bson.A{}

	if v := q.Get("q"); v != "" {
		and = append(and, bson.M{"$or": bson.A{
			bson.M{"name": bson.M{"$regex": v, "$options": "i"}},
			bson.M{"nickname": bson.M{"$regex": v, "$options": "i"}},
		}})
	}

	filter := bson.M{}
	if len(and) > 0 {
		filter["$and"] = and
	}
	limit := db.LimitFromQuery(r, 50, 50)
	if after := db.AfterFromQuery(r); after != "" {
		filter["_id"] = bson.M{"$gt": after}
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "_id", Value: 1}})

	cur, err := db.MongoDB.Collection("fighters").Find(r.Context(), filter, opts)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, "db error")
		return
	}
	defer cur.Close(r.Context())

	var items []data.Fighter
	for cur.Next(r.Context()) {
		var f data.Fighter
		if err := cur.Decode(&f); err != nil {
			render.Status(r, 500)
			render.PlainText(w, r, "decode error")
			return
		}
		items = append(items, f)
	}

	if n := len(items); n > 0 {
		w.Header().Set("X-Next-After", items[n-1].ID)
	}

	db.CacheFor(w, 30*time.Second)

	db.RenderJSON(w, r, data.Fighters{Items: items})
}

func GetFighter(w http.ResponseWriter, r *http.Request) {
	f, _ := r.Context().Value(pkg.CtxFighterKey).(*data.Fighter)
	if f == nil {
		render.Status(r, 404)
		render.PlainText(w, r, "fighter not found")
		return
	}
	db.RenderJSON(w, r, f)
}

func parseFloat32(s string) float32 {
	if f, err := strconv.ParseFloat(s, 32); err == nil {
		return float32(f)
	}
	return 0
}

func ListEvents(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{}
	q := r.URL.Query()

	if v := q.Get("name"); v != "" {
		filter["name"] = bson.M{"$regex": v, "$options": "i"}
	}
	// optional date range: ?start=2023-01-01&end=2023-12-31
	if v := q.Get("start"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter["date"] = bson.M{"$gte": t}
		}
	}
	if v := q.Get("end"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			if m, ok := filter["date"].(bson.M); ok {
				m["$lte"] = t
			} else {
				filter["date"] = bson.M{"$lte": t}
			}
		}
	}

	limit := db.LimitFromQuery(r, 50, 50)
	if after := db.AfterFromQuery(r); after != "" {
		filter["_id"] = bson.M{"$gt": after}
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "date", Value: -1}})

	cur, err := db.MongoDB.Collection("events").Find(r.Context(), filter, opts)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, "db error")
		return
	}
	defer cur.Close(r.Context())

	var items []data.Event
	for cur.Next(r.Context()) {
		var e data.Event
		if err := cur.Decode(&e); err != nil {
			render.Status(r, 500)
			render.PlainText(w, r, "decode error")
			return
		}
		items = append(items, e)
	}

	if n := len(items); n > 0 {
		w.Header().Set("X-Next-After", items[n-1].ID)
	}

	db.CacheFor(w, 30*time.Second)

	db.RenderJSON(w, r, data.Events{Items: items})
}

func SearchEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	and := bson.A{}

	if v := q.Get("q"); v != "" {
		and = append(and, bson.M{"$or": bson.A{
			bson.M{"name": bson.M{"$regex": v, "$options": "i"}},
			bson.M{"location": bson.M{"$regex": v, "$options": "i"}},
		}})
	}

	filter := bson.M{}
	if len(and) > 0 {
		filter["$and"] = and
	}

	limit := db.LimitFromQuery(r, 50, 50)
	if after := db.AfterFromQuery(r); after != "" {
		filter["_id"] = bson.M{"$gt": after}
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "date", Value: -1}})

	cur, err := db.MongoDB.Collection("events").Find(r.Context(), filter, opts)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, "db error")
		return
	}
	defer cur.Close(r.Context())

	var items []data.Event
	for cur.Next(r.Context()) {
		var e data.Event
		if err := cur.Decode(&e); err != nil {
			render.Status(r, 500)
			render.PlainText(w, r, "decode error")
			return
		}
		items = append(items, e)
	}

	if n := len(items); n > 0 {
		w.Header().Set("X-Next-After", items[n-1].ID)
	}

	db.CacheFor(w, 30*time.Second)

	db.RenderJSON(w, r, data.Events{Items: items})
}

func GetEvent(w http.ResponseWriter, r *http.Request) {
	e, _ := r.Context().Value(pkg.CtxEventKey).(*data.Event)
	if e == nil {
		render.Status(r, 404)
		render.PlainText(w, r, "event not found")
		return
	}
	db.RenderJSON(w, r, e)
}

func ListUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{}
	q := r.URL.Query()

	if v := q.Get("name"); v != "" {
		filter["name"] = bson.M{"$regex": v, "$options": "i"}
	}
	// optional date range: ?start=2023-01-01&end=2023-12-31
	if v := q.Get("start"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter["date"] = bson.M{"$gte": t}
		}
	}
	if v := q.Get("end"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			if m, ok := filter["date"].(bson.M); ok {
				m["$lte"] = t
			} else {
				filter["date"] = bson.M{"$lte": t}
			}
		}
	}

	limit := db.LimitFromQuery(r, 50, 50)
	if after := db.AfterFromQuery(r); after != "" {
		filter["_id"] = bson.M{"$gt": after}
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "date", Value: -1}})

	cur, err := db.MongoDB.Collection("upcomingEvents").Find(r.Context(), filter, opts)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, "db error")
		return
	}
	defer cur.Close(r.Context())

	var items []data.UpcomingEvent
	for cur.Next(r.Context()) {
		var e data.UpcomingEvent
		if err := cur.Decode(&e); err != nil {
			render.Status(r, 500)
			render.PlainText(w, r, "decode error")
			return
		}
		items = append(items, e)
	}

	if n := len(items); n > 0 {
		w.Header().Set("X-Next-After", items[n-1].ID)
	}

	db.CacheFor(w, 30*time.Second)

	db.RenderJSON(w, r, data.UpcomingEvents{Items: items})
}

func SearchUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	and := bson.A{}

	if v := q.Get("q"); v != "" {
		and = append(and, bson.M{"$or": bson.A{
			bson.M{"name": bson.M{"$regex": v, "$options": "i"}},
			bson.M{"location": bson.M{"$regex": v, "$options": "i"}},
		}})
	}

	filter := bson.M{}
	if len(and) > 0 {
		filter["$and"] = and
	}

	limit := db.LimitFromQuery(r, 50, 50)
	if after := db.AfterFromQuery(r); after != "" {
		filter["_id"] = bson.M{"$gt": after}
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "date", Value: -1}})

	cur, err := db.MongoDB.Collection("upcomingEvents").Find(r.Context(), filter, opts)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, "db error")
		return
	}
	defer cur.Close(r.Context())

	var items []data.UpcomingEvent
	for cur.Next(r.Context()) {
		var e data.UpcomingEvent
		if err := cur.Decode(&e); err != nil {
			render.Status(r, 500)
			render.PlainText(w, r, "decode error")
			return
		}
		items = append(items, e)
	}

	if n := len(items); n > 0 {
		w.Header().Set("X-Next-After", items[n-1].ID)
	}

	db.CacheFor(w, 30*time.Second)

	db.RenderJSON(w, r, data.UpcomingEvents{Items: items})
}

func GetUpcomingEvent(w http.ResponseWriter, r *http.Request) {
	e, _ := r.Context().Value(pkg.CtxUpcomingEventKey).(*data.UpcomingEvent)
	if e == nil {
		render.Status(r, 404)
		render.PlainText(w, r, "event not found")
		return
	}
	db.RenderJSON(w, r, e)
}

func ListUpcomingFights(w http.ResponseWriter, r *http.Request) {
	and := bson.A{}
	q := r.URL.Query()

	if v := q.Get("upcoming_event_id"); v != "" {
		and = append(and, bson.M{"upcoming_event_id": v})
	}

	if names := q["fighter_name"]; len(names) > 0 {
		for _, n := range names {
			and = append(and, bson.M{
				"tale_of_the_tape.name": bson.M{"$regex": n, "$options": "i"},
			})
		}
	}

	filter := bson.M{}
	if len(and) > 0 {
		filter["$and"] = and
	}

	limit := db.LimitFromQuery(r, 50, 50)
	if after := db.AfterFromQuery(r); after != "" {
		filter["_id"] = bson.M{"$gt": after}
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "_id", Value: 1}})

	cur, err := db.MongoDB.Collection("upcomingFights").Find(r.Context(), filter, opts)
	if err != nil {
		render.Status(r, 500)
		render.PlainText(w, r, "db error")
		return
	}
	defer cur.Close(r.Context())

	var items []data.UpcomingFight
	for cur.Next(r.Context()) {
		var f data.UpcomingFight
		if err := cur.Decode(&f); err != nil {
			render.Status(r, 500)
			render.PlainText(w, r, "decode error")
			return
		}
		items = append(items, f)
	}

	if n := len(items); n > 0 {
		w.Header().Set("X-Next-After", items[n-1].ID)
	}

	db.CacheFor(w, 30*time.Second)

	db.RenderJSON(w, r, data.UpcomingFights{Items: items})
}

func GetUpcomingFight(w http.ResponseWriter, r *http.Request) {
	f, _ := r.Context().Value(pkg.CtxUpcomingFightKey).(*data.UpcomingFight)
	if f == nil {
		render.Status(r, 404)
		render.PlainText(w, r, "fight not found")
		return
	}
	db.RenderJSON(w, r, f)
}
