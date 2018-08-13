package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func helloWorldServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

var (
	logger        = log.New(os.Stdout, "http: ", log.LstdFlags)
	ListeningPort = "8080"
)

func loggerServer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Println(r.RemoteAddr + " requested " + r.URL.String())
		h.ServeHTTP(w, r)
	})
}

func main() {
	logger.Println("Server is starting...")
	//Set up all the routing we need
	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(helloWorldServer))
	fs := http.FileServer(http.Dir("helloWorldHttp/public/"))
	router.Handle("/favicon.ico", fs)
	//Adding our Middleware
	h1 := loggerServer(router)

	//Defining custom attributes for the server
	//If we wanted the default ones we could go with http.ListenAndServe(":"+ListeningPort,h1)
	server := &http.Server{
		Addr:         ":" + ListeningPort,
		Handler:      h1,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		//Shutting down the server
		<-quit
		logger.Println("Server is shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatal("Could not gracefully shutdown the server: " + err.Error() + "\n")
		}
		close(done)
	}()
	logger.Println("Server is ready to handle requests at", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Could not listen on :" + ListeningPort + ". Error: " + err.Error())
	}
	<-done
	logger.Println("Server stopped")
}
