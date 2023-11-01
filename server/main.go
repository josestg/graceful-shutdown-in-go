package main

import (
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func slowHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("X-Request-Id")
	log := slog.Default().With("id", id, "method", r.Method, "path", r.URL.Path)

	// simulate slow process.
	delay := time.Duration(5+rand.Intn(5)) * time.Second // 5-10 seconds.

	startedAt := time.Now()
	log.Info("req received", "delay", delay)
	defer func() {
		log.Info("req completed", "latency", time.Since(startedAt))
	}()

	time.Sleep(delay)
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, id)
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	slog.SetDefault(log)

	mux := http.NewServeMux()
	mux.HandleFunc("/slow-process", slowHandler)

	// create http server.
	srv := &http.Server{Addr: ":8080", Handler: mux}
	log.Info("server started", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("could not start server", "error", err)
		os.Exit(1)
	}

	watchedSignals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}

	shutdownListener := make(chan os.Signal, 1)
	signal.Notify(shutdownListener, watchedSignals...)

	log.Info("listen for shutdown request", "watched_signals", watchedSignals)
	select {
	case sig := <-shutdownListener:
		log.Info("received shutdown request", "signal", sig)
	}
}
