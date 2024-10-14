package main

import (
	"encoding/json"
	"github.com/goji/httpauth"
	"github.com/jackcvr/gpstrack/orm"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const PublicDir = "./public"

var (
	AbsPublicDir string
	CertFile     = os.Getenv("HTTP_CERT_FILE")
	KeyFile      = os.Getenv("HTTP_KEY_FILE")
	Username     = os.Getenv("HTTP_USERNAME")
	Password     = os.Getenv("HTTP_PASSWORD")
)

func init() {
	var err error
	AbsPublicDir, err = filepath.Abs(PublicDir)
	if err != nil {
		panic(err)
	}
}

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

func StartHTTPServer(addr string) {
	basicAuth := httpauth.SimpleBasicAuth(Username, Password)

	http.Handle("GET /records", logRequest(basicAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		from := req.FormValue("from")
		to := req.FormValue("to")

		var err error
		now := time.Now()
		start := toDayLocal(now)
		end := start.Add(time.Hour * 24)

		if from != "" && to != "" {
			start, err = time.Parse("2006-1-2", from)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			start = toDayLocal(start)
			end, err = time.Parse("2006-1-2", to)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			end = toDayLocal(end)
		}

		var recs []orm.Record
		if err = orm.Client.Where("created_at >= ? AND created_at < ?", start, end).Find(&recs).Error; err != nil {
			errLog.Print(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(asJSON(recs))
		if err != nil {
			errLog.Print(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}))))

	http.Handle("GET /", logRequest(basicAuth(http.FileServer(http.Dir(AbsPublicDir)))))

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServeTLS(addr, CertFile, KeyFile, nil))
}
