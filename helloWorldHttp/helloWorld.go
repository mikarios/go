package main

import (
	"context"
	"flag"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

type Logger struct {
	*log.Logger
}

var (
	logger             = &Logger{log.New(os.Stdout, "http: ", log.LstdFlags)}
	ListeningPort      = flag.Int("port", 8080, "port number")
	Username           = "user"
	Password           = "pass"
	UsernameQueryParam = flag.String("usernameQueryParam", "u", "What the username query param will be.")
	UsernameHeadParam  = flag.String("usernameHeadParam", "u", "What the username header will be.")
	PasswordQueryParam = flag.String("passwordQueryParam", "p", "What the password query param will be.")
	PasswordHeadParam  = flag.String("passwordHeadParam", "p", "What the password header will be.")
	OutputMode         = flag.String("outputMode", "debug", "Accepted values: debug, info, warning, error")
)

const (
	OUTPUT_DEBUG   = "debug"
	OUTPUT_INFO    = "info"
	OUTPUT_WARNING = "warning"
	OUTPUT_ERROR   = "error"
)

func isDebugModeOn() bool {
	return *OutputMode == OUTPUT_DEBUG
}

func isInfoModeOn() bool {
	return *OutputMode == OUTPUT_DEBUG || *OutputMode == OUTPUT_INFO
}

func isWarningModeOn() bool {
	return *OutputMode != OUTPUT_ERROR
}

func (logger *Logger) debug(a ...interface{}) {
	if isDebugModeOn() {
		logger.Println(Brown("Debug:"), Brown(a))
	}
}

func (logger *Logger) success(a ...interface{}) {
	if isInfoModeOn() {
		logger.Println(Green("Success:"), Green(a))
	}
}

func (logger *Logger) info(a ...interface{}) {
	if isInfoModeOn() {
		logger.Println("Info:", a)
	}
}

func (logger *Logger) warn(a ...interface{}) {
	if isWarningModeOn() {
		logger.Println(Magenta("Warning"), Magenta(a))
	}
}

func (logger *Logger) error(a ...interface{}) {
	if isWarningModeOn() {
		logger.Print(Red("Error"), Red(a))
		os.Exit(-1)
	}
}

func helloWorldServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "The page you requested ("+r.URL.Path[1:]+") says Hello!")
}

// Used for logging each request
func logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer logger.debug("Request " + r.URL.String() + " answered")
			logger.debug(r.RemoteAddr + " requested " + r.URL.String())
			next.ServeHTTP(w, r)
		})
	}
}

func authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Just show the favicon, who cares?
		if strings.Contains(r.URL.Path, "favicon.ico") {
			h.ServeHTTP(w, r)
			return
		}
		logger.debug("Checking if user is authenticated")
		if !(r.Header.Get(*PasswordHeadParam) == Password || r.URL.Query().Get(*PasswordQueryParam) == Password) ||
			!(r.Header.Get(*UsernameHeadParam) == Username || r.URL.Query().Get(*UsernameQueryParam) == Username) {
			w.WriteHeader(401)
			fmt.Fprintln(w, "You are not authorized to be here!")
			logger.warn("Unauthorized request", r.Header, r.Body, r.URL)
			return
		}
		logger.debug("Authenticated successfully")
		h.ServeHTTP(w, r)
	})
}

//If we didn't want to dynamically call the next handler we could:
//Don't forget to set server.Handler to : authenticate(router)
//func logging(h http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		defer logger.Println("Request " + r.URL.String() + " answered")
//		logger.Println(r.RemoteAddr + " requested " + r.URL.String())
//		h.ServeHTTP(w, r)
//	})
//}
//
//func authenticate(h http.Handler) http.Handler {
//	authenticate := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		logger.Println("CheckingforPron")
//		if strings.Contains(r.URL.String(), "pron") {
//			w.WriteHeader(404)
//			fmt.Fprintln(w, "You will not find pron here!")
//			return
//		}
//		h.ServeHTTP(w, r)
//	})
//	return logging(checkforpron)
//}

func main() {
	flag.Parse()

	*OutputMode = strings.ToLower(*OutputMode)

	if *OutputMode != OUTPUT_DEBUG && *OutputMode != OUTPUT_INFO && *OutputMode != OUTPUT_WARNING && *OutputMode != OUTPUT_ERROR {
		logger.error("No valid output mode was defined:", *OutputMode)
	}

	logger.info("Server is starting...")
	//Set up all the routing we need
	router := http.NewServeMux()
	router.HandleFunc("/", helloWorldServer)
	fs := http.FileServer(http.Dir("helloWorldHttp/public/"))
	router.Handle("/favicon.ico", fs)
	//Adding our Middleware
	//h1 := authenticate(router)
	//hhs := logging()(h1)

	//Defining custom attributes for the server
	//If we wanted the default ones we could go with http.ListenAndServe(":"+ListeningPort,h1)
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(*ListeningPort),
		Handler:      logging()(authenticate(router)),
		ErrorLog:     logger.Logger,
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
		logger.info("Server is shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.error("Could not gracefully shutdown the server: ", err.Error())
		}
		close(done)
	}()
	logger.info("Server is ready to handle requests at", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.error("Could not listen on :", *ListeningPort, "Error: ", err.Error())
	}
	<-done
	logger.info("Server stopped")
}
