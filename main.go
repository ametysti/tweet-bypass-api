package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
)

func getToken(id string) string {
	idNum, _ := strconv.ParseInt(id, 10, 64)
	token := (float64(idNum) / 1e15) * math.Pi
	tokenStr := fmt.Sprintf("%x", token)
	return removeZeroesAndDot(tokenStr)
}

func removeZeroesAndDot(s string) string {
	return removeRegex(s, `(0+|\\.)`)
}

func removeRegex(s string, regex string) string {
	return regexp.MustCompile(regex).ReplaceAllString(s, "")
}

func main() {
	if _, err := os.Stat("/.dockerenv"); err != nil {
		err := godotenv.Load()

		if err != nil {
			println("Failed to load .env:", err.Error())
		}
	}

	r := chi.NewRouter()

	if os.Getenv("ENABLE_LOGGING") == "true" {
		println("Enabled logging")
		r.Use(middleware.Logger)
	}

	ratelimit, _ := strconv.Atoi(os.Getenv("RATELIMIT"))

	if ratelimit > 0 {
		fmt.Printf("Ratelimiting by %d", ratelimit)
		r.Use(httprate.LimitByIP(ratelimit, 1*time.Minute))
	}

	r.Get("/api/retvieveTweet", func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{}

		id := r.URL.Query().Get("id")

		if id == "" {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		features := "tfw_timeline_list:;tfw_follower_count_sunset:true;tfw_tweet_edit_backend:on;tfw_refsrc_session:on;tfw_fosnr_soft_interventions_enabled:on;tfw_mixed_media_15897:treatment;tfw_experiments_cookie_expiration:1209600;tfw_show_birdwatch_pivots_enabled:on;tfw_duplicate_scribes_to_settings:on;tfw_use_profile_image_shape_enabled:on;tfw_video_hls_dynamic_manifests_15082:true_bitrate;tfw_legacy_timeline_sunset:true;tfw_tweet_edit_frontend:on"

		requestUrl := fmt.Sprintf("https://cdn.syndication.twimg.com/tweet-result?id=%s&lang=en&token=%s&features=%s", id, getToken(id), features)

		req, _ := http.NewRequest("GET", requestUrl, nil)

		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 14.5; rv:127.0) Gecko/20100101 Firefox/127.0")

		resp, _ := client.Do(req)

		if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			http.Error(w, "Tweet not found (API supposed to have Content-Type application/json)", 404)
			return
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Twitter API returned unknown Content-Type", 400)
			return
		}

		body, _ := io.ReadAll(resp.Body)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		w.WriteHeader(resp.StatusCode)
		fmt.Fprint(w, string(body))
	})

	http.ListenAndServe("127.0.0.1:3000", r)
}
