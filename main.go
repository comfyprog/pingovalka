package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
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

var version string

func main() {
	configFilename, showVersion, output, err := getFlags(os.Args[1:])

	if err == flag.ErrHelp {
		fmt.Println(output)
		os.Exit(2)
	} else if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if showVersion {
		fmt.Println(getVersionString(version))
		os.Exit(0)
	}

	if configFilename == "" {
		fmt.Println("Error: no config file provided")
		os.Exit(1)
	}

	rawConfig, err := getRawConfig(configFilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config, err := parseConfig(rawConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	stopChan := make(chan struct{})
	pingChan := make(chan Host, 5)
	defer close(pingChan)
	pingHosts(config.Hosts, pingChan, stopChan, config.ConstantUpdates)

	pingMux := NewPingMux(config.Hosts, pingChan)

	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	wsHandler := makeWebsocketHandler(&upgrader, pingMux, time.Second*10)

	indexPageTemplate := template.Must(template.ParseFS(frontend.FrontendFs, "index.html"))
	indexPageData := makeIndexData(config.PageTitle, wsUrl)

	switchMiddleware := switchIndexMiddleware(frontend.FrontendFs, indexPageTemplate, indexPageData)
	logRequests := requestLogMiddleware()

	basicAuthMiddleware := makeBasicAuthMiddleware(config.BasicAuth)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(frontend.FrontendFs)))
	mux.HandleFunc(wsUrl, wsHandler)

	muxWithMiddlewares := logRequests(switchMiddleware(mux))

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

	go pingMux.TransmitStatuses()

	log.Printf("server started on %s\n", config.ListenAddr())

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
