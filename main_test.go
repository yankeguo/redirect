package main

import (
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestCreateServer(t *testing.T) {
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t.Run("standard", func(t *testing.T) {
		ms := createServer(Options{
			Listen:  ":1122",
			Target:  "https://b.example.com",
			Verbose: true,
		})
		require.Equal(t, ":1122", ms.Addr)

		s := httptest.NewServer(ms.Handler)
		defer s.Close()

		t.Run("health-check", func(t *testing.T) {
			res, err := c.Get(s.URL + "/healthz")
			require.NoError(t, err)

			defer res.Body.Close()

			require.Equal(t, http.StatusOK, res.StatusCode)

			buf, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, "OK", string(buf))
		})

		t.Run("redirect", func(t *testing.T) {
			res, err := c.Get(s.URL + "/bbb/ccc")
			require.NoError(t, err)

			defer res.Body.Close()

			require.Equal(t, http.StatusFound, res.StatusCode)
			require.Equal(t, "https://b.example.com", res.Header.Get("Location"))
		})

	})

	t.Run("rewrite", func(t *testing.T) {
		ms := createServer(Options{
			Listen:    ":1122",
			Target:    "https://b.example.com/ddd/",
			Verbose:   true,
			Permanent: true,
		})
		require.Equal(t, ":1122", ms.Addr)

		s := httptest.NewServer(ms.Handler)
		defer s.Close()

		t.Run("redirect", func(t *testing.T) {
			res, err := c.Get(s.URL + "/bbb/ccc")
			require.NoError(t, err)

			defer res.Body.Close()

			require.Equal(t, http.StatusMovedPermanently, res.StatusCode)
			require.Equal(t, "https://b.example.com/ddd/bbb/ccc", res.Header.Get("Location"))
		})

	})

	t.Run("rewrite-prefix", func(t *testing.T) {
		ms := createServer(Options{
			Listen:    ":1122",
			Target:    "https://b.example.com/ddd/",
			Prefix:    "/bbb",
			Verbose:   true,
			Permanent: true,
		})
		require.Equal(t, ":1122", ms.Addr)

		s := httptest.NewServer(ms.Handler)
		defer s.Close()

		t.Run("redirect", func(t *testing.T) {
			res, err := c.Get(s.URL + "/bbb/ccc")
			require.NoError(t, err)

			defer res.Body.Close()

			require.Equal(t, http.StatusMovedPermanently, res.StatusCode)
			require.Equal(t, "https://b.example.com/ddd/ccc", res.Header.Get("Location"))
		})

	})
}

func TestCMDMain(t *testing.T) {
	os.Setenv("REDIRECT_LISTEN", ":8080")
	os.Setenv("REDIRECT_TARGET", "https://bbb.example.com")

	chDone := make(chan struct{})

	go func() {
		main()
		close(chDone)
	}()

	time.Sleep(time.Second)

	self, _ := os.FindProcess(os.Getpid())
	self.Signal(syscall.SIGINT)

	<-chDone
}
