package main

import (
	_ "embed"
	"github.com/goccy/go-json"
	"github.com/goji/httpauth"
	"github.com/jackcvr/gpsmap/orm"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

const DateFormat = "2006-01-02"

//go:embed public/index.html
var indexHTML []byte

//go:embed public/main.js
var mainJS []byte

//go:embed public/main.css
var mainCSS []byte

func asJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func toDayLocal(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func logRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.RequestURI, time.Since(start))
	})
}

func StartHTTPServer(db *gorm.DB, config HTTPConfig, debug bool) {
	if !debug {
		m := minify.New()
		m.Add("text/html", &html.Minifier{
			KeepDocumentTags: true,
			KeepEndTags:      true,
			KeepQuotes:       true,
		})
		m.AddFunc("text/js", js.Minify)
		m.AddFunc("text/css", css.Minify)
		var err error
		if indexHTML, err = m.Bytes("text/html", indexHTML); err != nil {
			panic(err)
		}
		if mainJS, err = m.Bytes("text/js", mainJS); err != nil {
			panic(err)
		}
		if mainCSS, err = m.Bytes("text/css", mainCSS); err != nil {
			panic(err)
		}
	}

	basicAuth := httpauth.SimpleBasicAuth(config.Username, config.Password)
	Handler := func(h http.Handler) http.Handler {
		return logRequest(basicAuth(h))
	}

	http.Handle("GET /records", Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		from := req.FormValue("from")
		to := req.FormValue("to")

		var err error
		now := time.Now()
		start := toDayLocal(now)
		end := start.Add(time.Hour * 24)

		if from != "" && to != "" {
			if start, err = time.Parse(DateFormat, from); err == nil {
				start = toDayLocal(start)
				if end, err = time.Parse(DateFormat, to); err == nil {
					end = toDayLocal(end)
				}
			}
			if err != nil {
				errLog.Print(err)
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}

		var recs []orm.Record
		if err = db.Where("created_at >= ? AND created_at < ?", start, end).Find(&recs).Error; err != nil {
			errLog.Print(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(asJSON(recs))
		if err != nil {
			errLog.Print(err)
		}
	})))

	http.Handle("GET /main.js", Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		if _, err := w.Write(mainJS); err != nil {
			errLog.Print(err)
		}
	})))

	http.Handle("GET /main.css", Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/css")
		if _, err := w.Write(mainCSS); err != nil {
			errLog.Print(err)
		}
	})))

	http.Handle("GET /", Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, err := w.Write(indexHTML); err != nil {
			errLog.Print(err)
		}
	})))

	log.Printf("HTTP listening on %s", config.Bind)
	log.Fatal(http.ListenAndServeTLS(config.Bind, config.CertFile, config.KeyFile, nil))
}
