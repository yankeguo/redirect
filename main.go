package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	envListen    string
	envTarget    string
	envPrefix    string
	envPermanent bool
	envVerbose   bool

	OK = []byte("OK")
)

func exit(err *error) {
	if *err != nil {
		log.Printf("exited with error: %s", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	var err error
	defer exit(&err)

	log.SetOutput(os.Stdout)

	envListen = strings.TrimSpace(os.Getenv("REDIRECT_LISTEN"))
	if envListen == "" {
		envListen = ":80"
	}
	envTarget = strings.TrimSpace(os.Getenv("REDIRECT_TARGET"))
	if envTarget == "" {
		err = errors.New("missing environment $REDIRECT_TARGET")
		return
	}
	envPrefix = strings.TrimSpace(os.Getenv("REDIRECT_PREFIX"))
	envPermanent, _ = strconv.ParseBool(strings.TrimSpace(os.Getenv("REDIRECT_PERMANENT")))
	envVerbose, _ = strconv.ParseBool(strings.TrimSpace(os.Getenv("REDIRECT_VERBOSE")))

	code := http.StatusFound
	if envPermanent {
		code = http.StatusPermanentRedirect
	}

	m := http.NewServeMux()
	m.HandleFunc("/healthz", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.Header().Set("Content-Length", strconv.Itoa(len(OK)))
		_, _ = rw.Write(OK)
	})

	if strings.HasSuffix(envTarget, "/") {
		m.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
			u := &url.URL{RawQuery: req.URL.RawQuery}
			if envPrefix != "" {
				u.Path = strings.TrimPrefix(req.URL.Path, envPrefix)
			} else {
				u.Path = req.URL.Path
			}
			newURL := envTarget + strings.TrimPrefix(u.String(), "/")
			if envVerbose {
				log.Println(req.Method, req.URL.String(), code, newURL)
			}
			http.Redirect(rw, req, newURL, code)
		})
	} else {
		m.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
			if envVerbose {
				log.Println(req.Method, req.URL.String(), code, envTarget)
			}
			http.Redirect(rw, req, envTarget, code)
		})
	}

	s := &http.Server{
		Addr:    envListen,
		Handler: m,
	}

	chErr := make(chan error, 1)
	chSig := make(chan os.Signal, 1)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		chErr <- s.ListenAndServe()
	}()

	log.Println("listening at:", envListen)

	select {
	case err = <-chErr:
	case sig := <-chSig:
		log.Printf("signal caught: %s", sig.String())
		time.Sleep(time.Second * 3)
		err = s.Shutdown(context.Background())
	}
}
