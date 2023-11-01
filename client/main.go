package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	var num int
	flag.IntVar(&num, "n", 5, "number of concurrent requests")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	slog.SetDefault(log)

	var wg sync.WaitGroup
	for i := 1; i <= num; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(500 * time.Millisecond)
			doRequest(id)
		}(i)
	}
	wg.Wait()
	log.Info("client completed")
}

func doRequest(id int) {
	rid := fmt.Sprintf("req-%d", id)
	log := slog.Default().With("request_id", rid)

	req, err := http.NewRequest("GET", "http://localhost:8080/slow-process", nil)
	if err != nil {
		log.Error("could not create request", "error", err)
		return
	}

	req.Header.Set("X-Request-Id", rid)

	log.Info("sending request")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("could not send request", "error", err)
		return
	}
	defer res.Body.Close()
	log.Info("response received", "status", res.StatusCode)
}
