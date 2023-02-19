package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"log"
	"net/http"
	"untitled_go/serve/path"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	port := "8080"

	mux := http.NewServeMux()
	mux.HandleFunc("/", path.GetHomepage)
	mux.HandleFunc("/hello", path.GetHello)
	mux.HandleFunc("/test", path.GetTest)
	mux.HandleFunc("/chat", path.ChatWithGpt)
	done := make(chan bool)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
		if err != nil {
			log.Fatalf("failed start server at port %s", port)
		}
	}()

	log.Printf("server started at port %s", port)
	<-done
	close(done)
}
