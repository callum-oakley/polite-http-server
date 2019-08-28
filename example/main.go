package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/callum-oakley/polite-http-server"
)

func main() {
	s := polite.New(http.Server{
		Addr: ":5000",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for i := 0; i < 10; i++ {
				if _, err := io.WriteString(w, fmt.Sprintf("tick %v\n", i)); err != nil {
					log.Printf("io.WriteString: %v", err)
					return
				}
				w.(http.Flusher).Flush()
				time.Sleep(time.Second)
			}
		}),
	})

	shutdownFinished := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		startedShutdown := time.Now()

		log.Println("shutting down")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("s.Shutdown: %v", err)
		}

		log.Printf(
			"shut down complete after %vms",
			math.Round(time.Since(startedShutdown).Seconds()*1000),
		)

		close(shutdownFinished)
	}()

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatalf("net.Listen: %v", err)
	}
	defer l.Close()

	if err := s.ServeTLS(l, "example/cert.pem", "example/key.pem"); err != nil {
		log.Printf("s.ServeTLS: %v", err)
	}

	<-shutdownFinished
}
