package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

const wsUrl = "/ws"

type IndexPageData struct {
	Url     template.JSStr
	Version string
	Title   string
}

func makeIndexData(title string, websocketUrl string) IndexPageData {
	data := IndexPageData{}
	data.Url = template.JSStr(websocketUrl)
	data.Version = version
	data.Title = title
	return data
}

func switchIndexMiddleware(frontendFs fs.FS, indexPageTemplate *template.Template, data IndexPageData) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" || r.URL.Path == "/index.html" {
				err := indexPageTemplate.Execute(w, data)
				if err != nil {
					log.Println(err)
				}
				return
			} else {
				next.ServeHTTP(w, r)
			}
		}
	}
}

func makeBasicAuthMiddleware(credentials []BasicAuthCredentials) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Add("WWW-Authenticate", `Basic realm="Authorization:"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			for _, c := range credentials {
				if c.Username == username && c.Password == password {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Add("WWW-Authenticate", `Basic realm="Authorization:"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
}

func requestLogMiddleware() func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			log.Printf("%s\t%s%s", r.Method, r.Host, r.URL)
		}
	}
}
