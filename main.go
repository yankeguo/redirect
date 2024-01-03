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
	OK = []byte("OK")
)

type Options struct {
	Listen    string
	Target    string
	Prefix    string
	Permanent bool
	Verbose   bool
}

func createServer(opts Options) *http.Server {
	code := http.StatusFound
	if opts.Permanent {
		code = http.StatusMovedPermanently
	}

	m := http.NewServeMux()
	m.HandleFunc("/healthz", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.Header().Set("Content-Length", strconv.Itoa(len(OK)))
		_, _ = rw.Write(OK)
	})

	if strings.HasSuffix(opts.Target, "/") {
		m.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
			u := &url.URL{RawQuery: req.URL.RawQuery}
			if opts.Prefix == "" {
				u.Path = req.URL.Path
			} else {
				u.Path = strings.TrimPrefix(req.URL.Path, opts.Prefix)
			}
			newURL := opts.Target + strings.TrimPrefix(u.String(), "/")
			if opts.Verbose {
				log.Println(req.Method, req.URL.String(), code, newURL)
			}
			http.Redirect(rw, req, newURL, code)
		})
	} else {
		m.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
			if opts.Verbose {
				log.Println(req.Method, req.URL.String(), code, opts.Target)
			}
			http.Redirect(rw, req, opts.Target, code)
		})
	}

	return &http.Server{
		Addr:    opts.Listen,
		Handler: m,
	}
}

func envStr(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func envBool(key string) bool {
	v, _ := strconv.ParseBool(envStr(key))
	return v
}

func main() {
	var err error
	defer func() {
		if err == nil {
			log.Println("exited")
			return
		}
		log.Println("exited with error:", err.Error())
		os.Exit(1)
	}()

	log.SetOutput(os.Stdout)

	opts := Options{
		Listen:    envStr("REDIRECT_LISTEN"),
		Target:    envStr("REDIRECT_TARGET"),
		Prefix:    envStr("REDIRECT_PREFIX"),
		Permanent: envBool("REDIRECT_PERMANENT"),
		Verbose:   envBool("REDIRECT_VERBOSE"),
	}

	if opts.Listen == "" {
		opts.Listen = ":80"
	}

	if opts.Target == "" {
		err = errors.New("missing environment $REDIRECT_TARGET")
		return
	}

	s := createServer(opts)

	chErr := make(chan error, 1)
	chSig := make(chan os.Signal, 1)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		chErr <- s.ListenAndServe()
	}()

	log.Println("listening at:", s.Addr)

	select {
	case err = <-chErr:
		return
	case sig := <-chSig:
		log.Println("signal caught:", sig.String())
	}

	time.Sleep(time.Second * 3)

	err = s.Shutdown(context.Background())
}
