package db

import (
	"net/http"
	"strconv"
	"time"
)

func CacheFor(w http.ResponseWriter, d time.Duration) {
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(int(d.Seconds())))
}
