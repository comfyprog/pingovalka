package main

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/comfyprog/pingovalka/frontend"
	"github.com/gorilla/websocket"
)

const wsUrl = "/ws"
const version = "0.0.1"

type IndexPageData struct {
	Url     template.JS
	Version string
	Title   string
}

func makeIndexData(title string, websocketUrl string) IndexPageData {
	data := IndexPageData{}
	data.Url = template.JS(websocketUrl)
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

	frontendFs, err := fs.Sub(frontend.FrontendFs, "dist")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(frontendFs)))

	stopChan := make(chan struct{})
	pingChan := pingHosts(config.Hosts, stopChan)

	pingMux := NewPingMux(config.Hosts, pingChan)
	go pingMux.TransmitStatuses()

	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	wsHandler := makeWebsocketHandler(&upgrader, pingMux)

	mux.HandleFunc(wsUrl, wsHandler)

	indexPageTemplate := template.Must(template.ParseFS(frontendFs, "index.html"))
	indexPageData := makeIndexData(config.PageTitle, config.MakeFullPath(wsUrl, "ws"))
	switchMiddleware := switchIndexMiddleware(frontendFs, indexPageTemplate, indexPageData)

	basicAuthMiddleware := makeBasicAuthMiddleware(config.BasicAuth)

	muxWithMiddlewares := switchMiddleware(mux)

	if config.HasBasicAuthConfigured() {
		muxWithMiddlewares = basicAuthMiddleware(muxWithMiddlewares)
	}

	ctx, cancel := context.WithCancel(context.Background())

	httpServer := &http.Server{
		Addr:        config.ListenAddr(),
		Handler:     muxWithMiddlewares,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server returned error: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

	<-signalChan
	log.Println("shutting down...")

	go func() {
		<-signalChan
		log.Fatalln("terminating on repeated shut down signal")
	}()

	stopChan <- struct{}{}

	gracefulCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(gracefulCtx); err != nil {
		log.Printf("shutdown error: %v\n", err)
		defer os.Exit(1)
		return
	} else {
		log.Println("server stopped")
	}

	cancel()
	defer os.Exit(0)
	return
}