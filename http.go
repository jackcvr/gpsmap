package main

import (
	_ "embed"
	"github.com/goccy/go-json"
	"github.com/goji/httpauth"
	"github.com/jackcvr/gpsmap/orm"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"golang.org/x/net/websocket"
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

func ServeHTTP(config HTTPConfig, db *gorm.DB, pubsub *PubSub, debug bool) {
	if !debug {
		const (
			TextHTML = "text/html"
			TextJS   = "text/js"
			TextCSS  = "text/css"
		)
		m := minify.New()
		m.Add(TextHTML, &html.Minifier{
			KeepDocumentTags: true,
			KeepEndTags:      true,
			KeepQuotes:       true,
		})
		m.AddFunc(TextJS, js.Minify)
		m.AddFunc(TextCSS, css.Minify)
		var err error
		if indexHTML, err = m.Bytes(TextHTML, indexHTML); err != nil {
			panic(err)
		}
		if mainJS, err = m.Bytes(TextJS, mainJS); err != nil {
			panic(err)
		}
		if mainCSS, err = m.Bytes(TextCSS, mainCSS); err != nil {
			panic(err)
		}
	}

	http.Handle("GET /records", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		from := req.FormValue("from")
		to := req.FormValue("to")

		var err error
		var start, end time.Time
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
		} else {
			start = toDayLocal(time.Now())
			end = start.Add(time.Hour * 24)
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
	}))

	http.Handle("GET /ws", websocket.Handler(func(ws *websocket.Conn) {
		sub := pubsub.Subscribe(ws)
		defer func() {
			pubsub.Unsubscribe(ws)
		}()
		t := time.Tick(time.Second)
		for {
			select {
			case <-t:
				if err := websocket.Message.Send(ws, []byte("ping")); err != nil {
					errLog.Print(err)
					return
				}
			case r := <-sub:
				if err := websocket.Message.Send(ws, asJSON(r)); err != nil {
					errLog.Print(err)
					return
				}
			}
		}
	}))

	http.Handle("GET /main.js", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		if _, err := w.Write(mainJS); err != nil {
			errLog.Print(err)
		}
	}))

	http.Handle("GET /main.css", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/css")
		if _, err := w.Write(mainCSS); err != nil {
			errLog.Print(err)
		}
	}))

	http.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, err := w.Write(indexHTML); err != nil {
			errLog.Print(err)
		}
	}))

	basicAuth := httpauth.SimpleBasicAuth(config.Username, config.Password)

	log.Printf("HTTP listening on %s", config.Bind)
	log.Fatal(http.ListenAndServeTLS(config.Bind, config.CertFile, config.KeyFile, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.RequestURI, time.Since(start))
		}()
		basicAuth(http.DefaultServeMux).ServeHTTP(w, r)
	})))
}
