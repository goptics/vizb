package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ServeSuite struct {
	suite.Suite
}

func (s *ServeSuite) SetupTest() {
	ResetTestState()
}

func (s *ServeSuite) TestDefaultsAndHTTPPolicy() {
	s.Equal(defaultServeHost, serveOpts.Host)
	s.Equal(defaultServePort, serveOpts.Port)
	s.NotNil(serveCmd.Flags().Lookup("host"))
	s.NotNil(serveCmd.Flags().Lookup("port"))

	server := newHTTPServer("127.0.0.1:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		s.Error(err)
		var maxErr *http.MaxBytesError
		s.True(errors.As(err, &maxErr))
	}))
	s.Equal(readHeaderTimeout, server.ReadHeaderTimeout)
	s.Equal(readTimeout, server.ReadTimeout)
	s.Equal(writeTimeout, server.WriteTimeout)
	s.Equal(idleTimeout, server.IdleTimeout)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, maxRequestBodyBytes+1)))
	server.Handler.ServeHTTP(httptest.NewRecorder(), req)
}

func (s *ServeSuite) TestRoutesRegisterOnlyTheThreePOSTOperations() {
	called := make(map[string]bool)
	routes := composeRESTRoutes(restHandlers{
		convert: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called["convert"] = true
			w.WriteHeader(http.StatusNoContent)
		}),
		merge: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called["merge"] = true
			w.WriteHeader(http.StatusNoContent)
		}),
		ui: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called["ui"] = true; w.WriteHeader(http.StatusNoContent) }),
	})

	for _, test := range []struct {
		path string
		name string
	}{
		{path: "/", name: "convert"},
		{path: "/merge", name: "merge"},
		{path: "/ui", name: "ui"},
	} {
		recorder := httptest.NewRecorder()
		routes.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, test.path, nil))
		s.Equal(http.StatusNoContent, recorder.Code, test.path)
		s.True(called[test.name])
	}

	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/", nil),
		httptest.NewRequest(http.MethodPost, "/unknown", nil),
	} {
		recorder := httptest.NewRecorder()
		routes.ServeHTTP(recorder, request)
		s.Contains([]int{http.StatusMethodNotAllowed, http.StatusNotFound}, recorder.Code)
	}
}

func (s *ServeSuite) TestInvalidConfigurationReturnsCommandError() {
	options := serveOptions{Host: defaultServeHost, Port: defaultServePort}
	var listened atomic.Bool
	command := newServeCommand(&options, serveDependencies{
		newHandler: http.NotFoundHandler,
		listen: func(string, string) (net.Listener, error) {
			listened.Store(true)
			return nil, errors.New("must not listen")
		},
		signalContext: directSignalContext,
	})
	command.SetArgs([]string{"--port", "0"})
	err := command.Execute()
	s.EqualError(err, "port must be between 1 and 65535")
	s.False(listened.Load())
}

func (s *ServeSuite) TestListenerFailureReturnsCommandError() {
	err := runServer(context.Background(), serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
		newHandler: http.NotFoundHandler,
		listen: func(string, string) (net.Listener, error) {
			return nil, errors.New("address already in use")
		},
	})
	s.EqualError(err, "listen on 127.0.0.1:8080: address already in use")
}

func (s *ServeSuite) TestCancellationTriggersGracefulShutdown() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	defer listener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
			newHandler: http.NotFoundHandler,
			listen:     func(string, string) (net.Listener, error) { return listener, nil },
		})
	}()

	s.Eventually(func() bool {
		connection, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
		if err != nil {
			return false
		}
		_ = connection.Close()
		return true
	}, time.Second, 10*time.Millisecond)
	cancel()
	s.Require().NoError(<-result)
}

func (s *ServeSuite) TestServeAddress() {
	for _, test := range []struct {
		options serveOptions
		want    string
	}{
		{options: serveOptions{Host: "", Port: 8080}, want: "host must not be empty"},
		{options: serveOptions{Host: "127.0.0.1", Port: -1}, want: "port must be between 1 and 65535"},
		{options: serveOptions{Host: "127.0.0.1", Port: 65536}, want: "port must be between 1 and 65535"},
	} {
		s.Run(test.want, func() {
			_, err := serveAddress(test.options)
			s.EqualError(err, test.want)
		})
	}
	address, err := serveAddress(serveOptions{Host: "::1", Port: 8080})
	s.Require().NoError(err)
	s.Equal("[::1]:8080", address)
}

func directSignalContext(ctx context.Context, _ ...os.Signal) (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx)
}

func TestServeSuite(t *testing.T) {
	suite.Run(t, new(ServeSuite))
}
