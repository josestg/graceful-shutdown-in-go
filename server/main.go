package main

import (
	"context"
	"errors"
	"flag"
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
	var waitTolerance time.Duration
	flag.DurationVar(&waitTolerance, "wait", 5*time.Second, "wait tolerance for graceful shutdown")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	slog.SetDefault(log)

	mux := http.NewServeMux()
	mux.HandleFunc("/slow-process", slowHandler)

	// create http server.
	srv := &http.Server{Addr: ":8080", Handler: mux}

	serverError := make(chan error, 1)
	go func() {
		log.Info("server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			// only capture error if it's not server closed error.
			if !errors.Is(err, http.ErrServerClosed) {
				serverError <- err
			}
		}
	}()

	watchedSignals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}

	shutdownListener := make(chan os.Signal, 1)
	signal.Notify(shutdownListener, watchedSignals...)

	log.Info("listen for shutdown request", "watched_signals", watchedSignals)
	select {
	case err := <-serverError:
		log.Error("listen and serve failed", "error", err)

	case sig := <-shutdownListener:
		log.Info("received shutdown request", "signal", sig)

		// shutdown process.
		log.Info("shutting down server", "wait_tolerance", waitTolerance)
		defer log.Info("server shutdown gracefully")

		// we don't want to wait forever for connections to close.
		ctx, cancel := context.WithTimeout(context.Background(), waitTolerance)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("server shutdown failed", "error", err)
		}
	}
}
