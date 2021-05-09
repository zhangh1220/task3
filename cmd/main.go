package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"golang.org/x/sync/errgroup"
	"github.com/pkg/errors"
)

func main() {
	g,ctx := errgroup.WithContext(context.Background())

	quit := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter,r *http.Request){
		quit <- struct{}{}
	})

	server:=http.Server{
		Addr:     ":8000",
		Handler:mux,
	}

	g.Go(func()error {
		return server.ListenAndServe()
	})

	g.Go(func() error{
		select {
			case <- ctx.Done():
				log.Println("errgroup exit")
			case <-quit:
				log.Println("http server shutting down")
		}

		tCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		return server.Shutdown(tCtx)
	})

	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		select {
			case s := <-c:
				return errors.Errorf("get os signal: %v", s)
			case <-ctx.Done():
				return ctx.Err()
		}
	})

	g.Wait()
}

