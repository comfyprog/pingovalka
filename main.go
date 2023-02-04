package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/comfyprog/pingovalka/frontend"
)

const wsUrl = "/ws"
const version = "0.0.1"

type IndexPageData struct {
	Url     template.JS
	Version string
	Title   string
}

func makeIndexData(title string) IndexPageData {
	data := IndexPageData{}
	data.Url = wsUrl
	data.Version = version
	data.Title = title
	return data
}

func switchIndexMiddleware(frontendFs fs.FS, indexPageTemplate *template.Template, data IndexPageData) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				indexPageTemplate.Execute(w, data)
				return
			} else {
				next.ServeHTTP(w, r)
			}
		}
	}
}

func main() {
	configFilename, err := getConfigFileName()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rawConfig, err := getRawConfig(configFilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	config, err := getConfig(rawConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(config)

	frontendFs, err := fs.Sub(frontend.FrontendFs, "dist")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(frontendFs)))

	indexPageTemplate := template.Must(template.ParseFS(frontendFs, "index.html"))
	indexPageData := makeIndexData(config.PageTitle)
	switchMiddleware := switchIndexMiddleware(frontendFs, indexPageTemplate, indexPageData)
	muxWithCustomIndex := switchMiddleware(mux)

	log.Fatal(http.ListenAndServe(config.ListenAddr(), muxWithCustomIndex))
}