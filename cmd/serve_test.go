package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/goptics/vizb/cmd/cli"
	internalcharts "github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/pkg/core"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type ServeSuite struct {
	suite.Suite
}

const (
	validDatasetJSON = `{"name":"Bench","axes":[{"key":"name"},{"key":"y"}],"settings":[{"type":"bar"}],"data":[{"name":"case","yAxis":"1"}]}`
	firstMergeJSON   = `{"name":"Bench","tag":"v1","axes":[{"key":"name"},{"key":"y"}],"settings":[{"type":"bar"}],"data":[{"name":"case","yAxis":"1"}]}`
	secondMergeJSON  = `{"name":"Bench","tag":"v2","axes":[{"key":"name"},{"key":"y"}],"settings":[{"type":"bar"}],"data":[{"name":"case","yAxis":"2"}]}`
)

func (s *ServeSuite) SetupTest() {
	ResetTestState()
}

func (s *ServeSuite) TestDefaultsAndHTTPPolicy() {
	s.Equal(defaultServeHost, serveBag.String("host"))
	s.Equal(defaultServePort, serveBag.Int("port"))
	s.NotNil(serveCmd.Flags().Lookup("host"))
	s.NotNil(serveCmd.Flags().Lookup("port"))

	var applicationCalled atomic.Bool
	server := newHTTPServer("127.0.0.1:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request convertRequest
		if !decodeAPIRequest(w, r, &request) {
			return
		}
		applicationCalled.Store(true)
	}))
	s.Equal(readHeaderTimeout, server.ReadHeaderTimeout)
	s.Equal(readTimeout, server.ReadTimeout)
	s.Equal(writeTimeout, server.WriteTimeout)
	s.Equal(idleTimeout, server.IdleTimeout)

	body := append([]byte(`{"input":"`), bytes.Repeat([]byte("a"), maxRequestBodyBytes)...)
	body = append(body, '"', '}')
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(recorder, req)
	s.Equal(http.StatusRequestEntityTooLarge, recorder.Code)
	s.Equal("application/problem+json", recorder.Header().Get("Content-Type"))
	s.JSONEq(`{
		"type":"https://vizb.goptics.org/problems/content-too-large",
		"title":"Content too large",
		"status":413,
		"detail":"Request body must not exceed 10485760 bytes.",
		"instance":"/"
	}`, recorder.Body.String())
	s.False(applicationCalled.Load())
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

	for _, test := range []struct {
		name    string
		request *http.Request
		status  int
		allow   string
		body    string
	}{
		{
			name:    "wrong method",
			request: httptest.NewRequest(http.MethodGet, "/", nil),
			status:  http.StatusMethodNotAllowed,
			allow:   http.MethodPost,
			body: `{
				"type":"https://vizb.goptics.org/problems/method-not-allowed",
				"title":"Method not allowed",
				"status":405,
				"detail":"This operation only supports POST.",
				"instance":"/"
			}`,
		},
		{
			name:    "unknown path",
			request: httptest.NewRequest(http.MethodPost, "/unknown", nil),
			status:  http.StatusNotFound,
			body: `{
				"type":"https://vizb.goptics.org/problems/not-found",
				"title":"Not found",
				"status":404,
				"detail":"The requested operation does not exist.",
				"instance":"/unknown"
			}`,
		},
	} {
		s.Run(test.name, func() {
			recorder := httptest.NewRecorder()
			routes.ServeHTTP(recorder, test.request)
			s.Equal(test.status, recorder.Code)
			s.Equal("application/problem+json", recorder.Header().Get("Content-Type"))
			s.Equal(test.allow, recorder.Header().Get("Allow"))
			s.JSONEq(test.body, recorder.Body.String())
		})
	}
}

func (s *ServeSuite) TestInvalidConfigurationReturnsCommandError() {
	bag := cli.NewFlagBag(serveFlags())
	var listened atomic.Bool
	command := newServeCommand(bag, serveDependencies{
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

func (s *ServeSuite) TestServeFailureIsReturned() {
	err := runServer(context.Background(), serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
		newHandler: http.NotFoundHandler,
		listen: func(string, string) (net.Listener, error) {
			return errorListener{err: errors.New("accept failed")}, nil
		},
	})
	s.EqualError(err, "serve HTTP: accept failed")
}

func (s *ServeSuite) TestServerClosedIsASuccessfulServeResult() {
	err := runServer(context.Background(), serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
		newHandler: http.NotFoundHandler,
		listen: func(string, string) (net.Listener, error) {
			return errorListener{err: http.ErrServerClosed}, nil
		},
	})
	s.NoError(err)
}

func (s *ServeSuite) TestCancellationShutsDownInMemoryListener() {
	listener := newBlockingListener()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
			newHandler: http.NotFoundHandler,
			listen:     func(string, string) (net.Listener, error) { return listener, nil },
		})
	}()
	s.waitForSignal(listener.acceptStarted)
	cancel()
	s.Require().NoError(s.waitForResult(result))
}

func (s *ServeSuite) TestCancellationTriggersGracefulShutdown() {
	listener := newBlockingListener()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var shutdownCalled atomic.Bool
	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
			newHandler: http.NotFoundHandler,
			listen:     func(string, string) (net.Listener, error) { return listener, nil },
			shutdown: func(server *http.Server, ctx context.Context) error {
				shutdownCalled.Store(true)
				return server.Shutdown(ctx)
			},
		})
	}()

	s.waitForSignal(listener.acceptStarted)
	cancel()
	s.Require().NoError(s.waitForResult(result))
	s.True(shutdownCalled.Load())
}

func (s *ServeSuite) TestShutdownFailureIsReturned() {
	listener := newBlockingListener()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
			newHandler: http.NotFoundHandler,
			listen:     func(string, string) (net.Listener, error) { return listener, nil },
			shutdown:   func(*http.Server, context.Context) error { return errors.New("drain failed") },
		})
	}()

	s.waitForSignal(listener.acceptStarted)
	cancel()
	s.EqualError(s.waitForResult(result), "shutdown HTTP server: drain failed")
	s.waitForSignal(listener.closed)
}

func (s *ServeSuite) TestShutdownFailureIgnoresCloseError() {
	listener := closeErrorListener{
		blockingListener: newBlockingListener(),
		err:              errors.New("close failed"),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	drainErr := errors.New("drain failed")
	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
			newHandler: http.NotFoundHandler,
			listen:     func(string, string) (net.Listener, error) { return listener, nil },
			shutdown:   func(*http.Server, context.Context) error { return drainErr },
		})
	}()

	s.waitForSignal(listener.acceptStarted)
	cancel()
	err := s.waitForResult(result)
	s.ErrorIs(err, drainErr)
	s.waitForSignal(listener.closed)
}

func (s *ServeSuite) TestShutdownDeadlineForceClosesActiveConnections() {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()
	listener := newSingleConnectionListener(serverConn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	handlerFinished := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
			newHandler: func() http.Handler {
				return http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					defer close(handlerFinished)
					close(requestStarted)
					select {
					case <-releaseRequest:
					case <-time.After(time.Second):
						s.Fail("timed out waiting to release the active request")
					}
				})
			},
			listen: func(string, string) (net.Listener, error) { return listener, nil },
			shutdown: func(server *http.Server, parent context.Context) error {
				deadline, cancel := context.WithTimeout(parent, 20*time.Millisecond)
				defer cancel()
				return server.Shutdown(deadline)
			},
		})
	}()

	clientResult := make(chan struct {
		response []byte
		err      error
	}, 1)
	go func() {
		_, err := io.WriteString(clientConn, "POST / HTTP/1.1\r\nHost: vizb\r\nContent-Length: 2\r\n\r\n{}")
		if err != nil {
			clientResult <- struct {
				response []byte
				err      error
			}{err: err}
			return
		}
		response, err := io.ReadAll(clientConn)
		clientResult <- struct {
			response []byte
			err      error
		}{response: response, err: err}
	}()
	s.waitForSignal(requestStarted)
	cancel()
	s.EqualError(s.waitForResult(result), "shutdown HTTP server: context deadline exceeded")
	select {
	case client := <-clientResult:
		s.NoError(client.err)
		s.Empty(client.response)
	case <-time.After(time.Second):
		s.Fail("active client connection was not force-closed")
	}
	close(releaseRequest)
	select {
	case <-handlerFinished:
	case <-time.After(time.Second):
		s.Fail("active request handler did not finish")
	}
}

func (s *ServeSuite) TestServeFailureDuringShutdownIsReturned() {
	listener := newBlockingErrorListener(errors.New("accept failed during shutdown"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
			newHandler: http.NotFoundHandler,
			listen:     func(string, string) (net.Listener, error) { return listener, nil },
			shutdown: func(*http.Server, context.Context) error {
				return listener.Close()
			},
		})
	}()

	s.waitForSignal(listener.acceptStarted)
	cancel()
	s.EqualError(s.waitForResult(result), "serve HTTP: accept failed during shutdown")
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

func (s *ServeSuite) TestConvertEndpoint() {
	handler := newRESTHandler()
	request := `{"input":"region,value\nwest,12\n","parser":"csv","charts":{"types":["bar"]}}`
	recorder := s.apiRequest(handler, "/", request, "application/json", "application/json")
	s.Equal(http.StatusOK, recorder.Code)
	s.Contains(recorder.Header().Get("Content-Type"), "application/json")
	var dataset map[string]any
	s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &dataset))
	s.Len(dataset["data"], 1)

	recorder = s.apiRequest(handler, "/", request[:len(request)-1]+`,"output":{"format":"html"}}`, "application/json", "text/html")
	s.Equal(http.StatusOK, recorder.Code)
	s.Contains(recorder.Header().Get("Content-Type"), "text/html")
	s.Contains(recorder.Body.String(), "VIZB_DATA")

	recorder = s.apiRequest(handler, "/", request[:len(request)-1]+`,"output":{"format":"dataset"}}`, "application/json", "application/json")
	s.Equal(http.StatusOK, recorder.Code)

	recorder = s.apiRequest(handler, "/", `{"input":[{"region":"west","value":12}],"charts":{"types":["bar"]}}`, "application/json", "")
	s.Equal(http.StatusOK, recorder.Code)

	recorder = s.apiRequest(
		handler,
		"/",
		`{"input":"region,latency\nwest,12\neast,18\n","parser":"csv","select":["region,latency"],"charts":{"types":["bar"]}}`,
		"application/json",
		"application/json",
	)
	s.Equal(http.StatusOK, recorder.Code, recorder.Body.String())

	documented := `{
		"input":{"data":[{"region":"west","latency":12},{"region":"east","latency":18}]},
		"parser":"auto",
		"grouping":{"pattern":"x","columns":["region"]},
		"units":{"memory":"KB","time":"ms","number":"K"},
		"select":["latency"],
		"jsonPath":".data",
		"charts":{"types":["bar"],"configs":[{"type":"bar","showLabels":true}]},
		"output":{"format":"dataset"}
	}`
	recorder = s.apiRequest(handler, "/", documented, "application/json", "application/json")
	s.Equal(http.StatusOK, recorder.Code)
	s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &dataset))
	s.Equal("Comparisons", dataset["name"])
}

func (s *ServeSuite) TestConvertEndpointSupportsBenchmarkFamilies() {
	handler := newRESTHandler()
	for _, test := range []struct {
		name   string
		parser string
		input  string
	}{
		{name: "go", parser: "go", input: "BenchmarkFoo-8 100 123 ns/op\n"},
		{name: "javascript", parser: "javascript", input: " · foo 1234 0.1 0.2 0.3 0.4 0.5 0.6 0.7 ±1.5% 100\n"},
		{name: "rust", parser: "rust", input: "foo time: [21.234 ns 21.456 ns 21.678 ns]\n"},
	} {
		s.Run(test.name, func() {
			body, err := json.Marshal(map[string]any{
				"input":  test.input,
				"parser": test.parser,
				"charts": map[string]any{"types": []string{"bar"}},
			})
			s.Require().NoError(err)
			recorder := s.apiRequest(handler, "/", string(body), "application/json", "application/json")
			s.Equal(http.StatusOK, recorder.Code, recorder.Body.String())
		})
	}
}

func (s *ServeSuite) TestConvertEndpointRejectsInvalidRequests() {
	handler := newRESTHandler()
	for _, test := range []struct {
		name        string
		body        string
		contentType string
		accept      string
		wantStatus  int
		wantErrors  bool
		wantPath    string
	}{
		{name: "missing input", body: `{}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true},
		{name: "null input", body: `{"input":null}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true},
		{name: "invalid input kind", body: `{"input":42}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true},
		{name: "unknown chart", body: `{"input":"x,y\na,1\n","charts":{"types":["unknown"]}}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true},
		{name: "unsupported parser", body: `{"input":"x,y\na,1\n","parser":"go"}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity},
		{name: "html not accepted", body: `{"input":"x,y\na,1\n","output":{"format":"html"}}`, contentType: "application/json", accept: "application/json", wantStatus: http.StatusNotAcceptable},
		{name: "dataset not accepted", body: `{"input":"x,y\na,1\n"}`, contentType: "application/json", accept: "text/html", wantStatus: http.StatusNotAcceptable},
		{name: "invalid output format", body: `{"input":"x,y\na,1\n","output":{"format":"csv"}}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true},
		{name: "metadata is not supported", body: `{"input":"x,y\na,1\n","metadata":{"name":"example"}}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true},
		{name: "unknown field", body: `{"input":"x,y\na,1\n","extra":true}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true},
		{name: "request not object", body: `"foo"`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true, wantPath: "/"},
		{name: "grouping not object", body: `{"input":"x,y\na,1\n","grouping":"foo"}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true, wantPath: "/grouping"},
		{name: "grouping pattern wrong type", body: `{"input":"x,y\na,1\n","grouping":{"pattern":123}}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity, wantErrors: true, wantPath: "/grouping/pattern"},
		{name: "missing content type", body: `{}`, wantStatus: http.StatusUnsupportedMediaType},
		{name: "wrong content type", body: `{}`, contentType: "text/plain", wantStatus: http.StatusUnsupportedMediaType},
		{name: "malformed content type", body: `{}`, contentType: `application/json; charset="`, wantStatus: http.StatusUnsupportedMediaType},
		{name: "multiple JSON values", body: `{} {}`, contentType: "application/json", wantStatus: http.StatusBadRequest, wantErrors: true},
	} {
		s.Run(test.name, func() {
			recorder := s.apiRequest(handler, "/", test.body, test.contentType, test.accept)
			s.Equal(test.wantStatus, recorder.Code)
			s.Contains(recorder.Header().Get("Content-Type"), "application/problem+json")
			s.Equal(float64(test.wantStatus), s.problemStatus(recorder))
			if test.wantErrors {
				var problem struct {
					Errors []apiValidationError `json:"errors"`
				}
				s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &problem))
				s.NotEmpty(problem.Errors)
				if test.wantPath != "" {
					s.Equal(test.wantPath, problem.Errors[0].Path)
					s.Equal("invalid_type", problem.Errors[0].Code)
				}
			}
		})
	}
}

func (s *ServeSuite) TestConvertEndpointValidatesStructuredOptions() {
	handler := newRESTHandler()
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid parser", body: `{"input":"x,y\na,1\n","parser":"xml"}`},
		{name: "invalid group pattern", body: `{"input":"x,y\na,1\n","parser":"csv","grouping":{"pattern":"["}}`},
		{name: "empty group column", body: `{"input":"x,y\na,1\n","parser":"csv","grouping":{"pattern":"x","columns":[""]}}`},
		{name: "invalid group regex", body: `{"input":"x,y\na,1\n","grouping":{"regex":"["}}`},
		{name: "invalid filter regex", body: `{"input":"x,y\na,1\n","grouping":{"filter":"["}}`},
		{name: "misaligned grouping", body: `{"input":"a,b,c\nx,y,1\n","parser":"csv","grouping":{"pattern":"x/y","columns":["a","b"]}}`},
		{name: "invalid memory unit", body: `{"input":"x,y\na,1\n","units":{"memory":"TB"}}`},
		{name: "invalid time unit", body: `{"input":"x,y\na,1\n","units":{"time":"minute"}}`},
		{name: "invalid number unit", body: `{"input":"x,y\na,1\n","units":{"number":"Q"}}`},
		{name: "json path with csv", body: `{"input":"x,y\na,1\n","parser":"csv","jsonPath":".data"}`},
		{name: "select with benchmark", body: `{"input":"BenchmarkFoo 100 1 ns/op\n","parser":"go","select":["x,y"]}`},
		{name: "empty select", body: `{"input":"x,y\na,1\n","select":[""]}`},
		{name: "empty grouped select", body: `{"input":"region,value\nwest,1\n","grouping":{"columns":["region"]},"select":[""]}`},
		{name: "invalid grouped select", body: `{"input":"region,value\nwest,1\n","grouping":{"columns":["region"]},"select":["value{"]}`},
		{name: "invalid solo select", body: `{"input":"x,y\na,1\n","select":["x"]}`},
		{name: "invalid repeatable select", body: `{"input":"a,b,c\n1,2,3\n","select":["a,b,c","a,b"]}`},
		{name: "duplicate grouped select", body: `{"input":"region,value\nwest,1\n","parser":"csv","grouping":{"pattern":"x","columns":["region"]},"select":["value","value"]}`},
		{name: "group select conflict", body: `{"input":"region,value\nwest,1\n","parser":"csv","grouping":{"pattern":"x","columns":["region"]},"select":["region"]}`},
		{name: "structured grouping separator mismatch", body: `{"input":"a,b,c\nx,y,1\n","grouping":{"pattern":"x,y,z","columns":["a/b/c"]}}`},
		{name: "empty chart types", body: `{"input":"x,y\na,1\n","charts":{"types":[]}}`},
		{name: "duplicate chart types", body: `{"input":"x,y\na,1\n","charts":{"types":["bar","bar"]}}`},
		{name: "missing config type", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"showLabels":true}]}}`},
		{name: "unknown config type", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"unknown"}]}}`},
		{name: "duplicate configs", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar"},{"type":"bar"}]}}`},
		{name: "unselected config", body: `{"input":"x,y\na,1\n","charts":{"types":["bar"],"configs":[{"type":"line"}]}}`},
		{name: "unknown config field", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","bogus":true}]}}`},
		{name: "invalid scale", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","scale":"square"}]}}`},
		{name: "invalid symbol size", body: `{"input":"x,y\na,1\n","charts":{"types":["line"],"configs":[{"type":"line","symbolSize":0}]}}`},
		{name: "sort not object", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","sort":true}]}}`},
		{name: "sort missing enabled", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","sort":{"order":"asc"}}]}}`},
		{name: "sort missing order", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","sort":{"enabled":true}}]}}`},
		{name: "sort invalid order", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","sort":{"enabled":true,"order":"sideways"}}]}}`},
		{name: "stat not object", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","stat":true}]}}`},
		{name: "stat missing enabled", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","stat":{"math":[]}}]}}`},
		{name: "stat missing math", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","stat":{"enabled":true}}]}}`},
		{name: "null stat math", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","stat":{"enabled":true,"math":null}}]}}`},
		{name: "stat invalid math", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","stat":{"enabled":true,"math":["bogus"]}}]}}`},
		{name: "stat duplicate math", body: `{"input":"x,y\na,1\n","charts":{"configs":[{"type":"bar","stat":{"enabled":true,"math":["counts","counts"]}}]}}`},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			recorder := s.apiRequest(handler, "/", test.body, "application/json", "application/json")
			s.Equal(http.StatusUnprocessableEntity, recorder.Code, recorder.Body.String())
			var problem struct {
				Errors []apiValidationError `json:"errors"`
			}
			s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &problem))
			s.NotEmpty(problem.Errors)
		})
	}
}

func (s *ServeSuite) TestConvertEndpointReportsUIGenerationFailure() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleConvertWithGenerator(w, r, func([]shared.Dataset, []string) (string, error) {
			return "", errors.New("template failed")
		})
	})
	body := `{"input":"region,value\nwest,12\n","parser":"csv","charts":{"types":["bar"]},"output":{"format":"html"}}`
	recorder := s.apiRequest(handler, "/", body, "application/json", "text/html")
	s.Equal(http.StatusInternalServerError, recorder.Code)
	s.Equal(float64(http.StatusInternalServerError), s.problemStatus(recorder))
	s.NotContains(recorder.Body.String(), "template failed")
	s.Contains(recorder.Body.String(), "could not generate the response")
}

func (s *ServeSuite) TestConvertEndpointReportsInapplicableChartOptions() {
	handler := newRESTHandler()
	recorder := s.apiRequest(
		handler,
		"/",
		`{"input":"region,value\nwest,12\n","parser":"csv","charts":{"types":["bar"],"configs":[{"type":"bar","threeDRotate":true}]}}`,
		"application/json",
		"application/json",
	)
	s.Equal(http.StatusUnprocessableEntity, recorder.Code, recorder.Body.String())

	var problem struct {
		Errors []apiValidationError `json:"errors"`
	}
	s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &problem))
	s.Require().Len(problem.Errors, 1)
	s.Equal("/charts/configs/0/threeDRotate", problem.Errors[0].Path)
	s.Equal("inapplicable_option", problem.Errors[0].Code)
}

func (s *ServeSuite) TestMergeEndpoint() {
	handler := newRESTHandler()
	recorder := s.apiRequest(handler, "/merge", `{`, "application/json", "")
	s.Equal(http.StatusBadRequest, recorder.Code)

	recorder = s.apiRequest(handler, "/merge", `{"datasets":[],"extra":true}`, "application/json", "")
	s.Equal(http.StatusUnprocessableEntity, recorder.Code)
	s.Contains(recorder.Body.String(), `"path":"/"`)

	recorder = s.apiRequest(handler, "/merge", `{"datasets":[{"name":"one"}]}`, "application/json", "")
	s.Equal(http.StatusUnprocessableEntity, recorder.Code)

	recorder = s.apiRequest(handler, "/merge", `{"datasets":[`+firstMergeJSON+`,`+secondMergeJSON+`],"tagAxis":"x"}`, "application/json", "")
	s.Equal(http.StatusOK, recorder.Code)
	var merged []map[string]any
	s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &merged))
	s.Len(merged, 1)

	recorder = s.apiRequest(handler, "/merge", `{"datasets":[`+firstMergeJSON+`,`+secondMergeJSON+`],"tagAxis":"invalid"}`, "application/json", "")
	s.Equal(http.StatusUnprocessableEntity, recorder.Code)
	s.Contains(recorder.Body.String(), `"/tagAxis"`)

	recorder = s.apiRequest(handler, "/merge", `{"datasets":[{"name":"one"},{"name":"two"}]}`, "application/json", "")
	s.Equal(http.StatusUnprocessableEntity, recorder.Code)

	recorder = s.apiRequest(handler, "/merge", `{"datasets":[`+firstMergeJSON[:len(firstMergeJSON)-1]+`,"extra":true},`+secondMergeJSON+`]}`, "application/json", "")
	s.Equal(http.StatusUnprocessableEntity, recorder.Code)

	recorder = s.apiRequest(handler, "/merge", `{"datasets":[`+firstMergeJSON+`,`+secondMergeJSON+`]}`, "application/json", "text/html")
	s.Equal(http.StatusNotAcceptable, recorder.Code)
}

func (s *ServeSuite) TestUIEndpoint() {
	handler := newRESTHandler()
	valid := `{"datasets":` + validDatasetJSON + `}`
	recorder := s.apiRequest(handler, "/ui", valid, "application/json", "text/html")
	s.Equal(http.StatusOK, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Contains(recorder.Body.String(), "VIZB_DATA")

	recorder = s.apiRequest(handler, "/ui", `{"datasets":[`+validDatasetJSON+`]}`, "application/json", "text/html")
	s.Equal(http.StatusOK, recorder.Code)

	for _, test := range []struct {
		name       string
		body       string
		accept     string
		wantStatus int
	}{
		{name: "invalid request", body: `{`, accept: "text/html", wantStatus: http.StatusBadRequest},
		{name: "not accepted", body: valid, accept: "application/json", wantStatus: http.StatusNotAcceptable},
		{name: "missing datasets", body: `{}`, accept: "text/html", wantStatus: http.StatusUnprocessableEntity},
		{name: "empty datasets", body: `{"datasets":[]}`, accept: "text/html", wantStatus: http.StatusUnprocessableEntity},
		{name: "invalid datasets", body: `{"datasets":true}`, accept: "text/html", wantStatus: http.StatusUnprocessableEntity},
		{name: "incomplete dataset", body: `{"datasets":{"name":"Bench"}}`, accept: "text/html", wantStatus: http.StatusUnprocessableEntity},
		{name: "unknown nested field", body: `{"datasets":` + validDatasetJSON[:len(validDatasetJSON)-1] + `,"extra":true}}`, accept: "text/html", wantStatus: http.StatusUnprocessableEntity},
		{name: "remote request input", body: `{"datasets":` + validDatasetJSON + `,"dataUrl":"https://example.com/data.json"}`, accept: "text/html", wantStatus: http.StatusUnprocessableEntity},
		{name: "remote dataset input", body: `{"datasets":` + validDatasetJSON[:len(validDatasetJSON)-1] + `,"dataUrl":"https://example.com/data.json"}}`, accept: "text/html", wantStatus: http.StatusUnprocessableEntity},
		{name: "invalid chart type", body: `{"datasets":` + validDatasetJSON + `,"charts":{"types":["unknown"]}}`, accept: "text/html", wantStatus: http.StatusUnprocessableEntity},
	} {
		s.Run(test.name, func() {
			recorder := s.apiRequest(handler, "/ui", test.body, "application/json", test.accept)
			s.Equal(test.wantStatus, recorder.Code)
			s.Equal("application/problem+json", recorder.Header().Get("Content-Type"))
			s.Equal(float64(test.wantStatus), s.problemStatus(recorder))
		})
	}
}

func (s *ServeSuite) TestUIEndpointAcceptsConfigsStatisticsAndMediaRanges() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleUIWithGenerator(w, r, func(datasets []shared.Dataset, charts []string) (string, error) {
			s.Equal([]string{"bar"}, charts)
			s.Require().Len(datasets, 1)
			s.Require().Len(datasets[0].Settings, 1)
			s.True(datasets[0].Settings[0].StatEnabled())
			s.Equal([]string{"counts"}, datasets[0].Settings[0].StatMath())
			raw, err := json.Marshal(datasets[0].Settings[0])
			s.Require().NoError(err)
			var config map[string]any
			s.Require().NoError(json.Unmarshal(raw, &config))
			s.Equal(true, config["showLabels"])
			return "<html>ok</html>", nil
		})
	})
	body := `{"datasets":` + validDatasetJSON + `,"charts":{"types":["bar"],"configs":[{"type":"bar","showLabels":true}]},"statistics":{"enabled":true,"math":["counts"]}}`
	recorder := s.apiRequest(handler, "/ui", body, "application/json", "text/html;q=0.9,application/xhtml+xml")
	s.Equal(http.StatusOK, recorder.Code, recorder.Body.String())
}

func (s *ServeSuite) TestUIEndpointMaterialisesDefaultsForSelectedChartTypes() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleUIWithGenerator(w, r, func(datasets []shared.Dataset, charts []string) (string, error) {
			s.Equal([]string{"line"}, charts)
			s.Require().Len(datasets, 1)
			s.Require().Len(datasets[0].Settings, 1)
			s.Equal("line", datasets[0].Settings[0].ChartType())
			raw, err := json.Marshal(datasets[0].Settings[0])
			s.Require().NoError(err)
			var config map[string]any
			s.Require().NoError(json.Unmarshal(raw, &config))
			s.Equal("linear", config["scale"])
			return "<html>ok</html>", nil
		})
	})
	body := `{"datasets":` + validDatasetJSON + `,"charts":{"types":["line"]}}`
	recorder := s.apiRequest(handler, "/ui", body, "application/json", "text/html")
	s.Equal(http.StatusOK, recorder.Code, recorder.Body.String())
}

func (s *ServeSuite) TestUIEndpointAddsAnUnselectedOverrideType() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleUIWithGenerator(w, r, func(datasets []shared.Dataset, charts []string) (string, error) {
			s.Nil(charts)
			s.Require().Len(datasets, 1)
			s.Require().Len(datasets[0].Settings, 2)
			s.Equal("bar", datasets[0].Settings[0].ChartType())
			s.Equal("line", datasets[0].Settings[1].ChartType())
			return "<html>ok</html>", nil
		})
	})
	body := `{"datasets":` + validDatasetJSON + `,"charts":{"configs":[{"type":"line","smooth":true}]}}`
	recorder := s.apiRequest(handler, "/ui", body, "application/json", "text/html")
	s.Equal(http.StatusOK, recorder.Code, recorder.Body.String())
}

func (s *ServeSuite) TestUIEndpointValidatesDatasetsAndStatistics() {
	handler := newRESTHandler()
	tests := []struct {
		name     string
		datasets string
		suffix   string
	}{
		{name: "missing name", datasets: `{"axes":[],"settings":[],"data":[]}`},
		{name: "missing axes", datasets: `{"name":"Bench","settings":[],"data":[]}`},
		{name: "missing settings", datasets: `{"name":"Bench","axes":[],"data":[]}`},
		{name: "missing data", datasets: `{"name":"Bench","axes":[],"settings":[]}`},
		{name: "axis missing key", datasets: `{"name":"Bench","axes":[{}],"settings":[],"data":[]}`},
		{name: "invalid axis key", datasets: `{"name":"Bench","axes":[{"key":"metric"}],"settings":[],"data":[]}`},
		{name: "invalid axis type", datasets: `{"name":"Bench","axes":[{"key":"x","type":"category"}],"settings":[],"data":[]}`},
		{name: "invalid setting", datasets: `{"name":"Bench","axes":[],"settings":[{"type":"unknown"}],"data":[]}`},
		{name: "duplicate setting types", datasets: `{"name":"Bench","axes":[],"settings":[{"type":"bar"},{"type":"bar"}],"data":[]}`},
		{name: "history missing tag", datasets: `{"name":"Bench","history":[{"timestamp":"now"}],"axes":[],"settings":[],"data":[]}`},
		{name: "history missing timestamp", datasets: `{"name":"Bench","history":[{"tag":"v1"}],"axes":[],"settings":[],"data":[]}`},
		{name: "invalid statistics", datasets: validDatasetJSON, suffix: `,"statistics":{"math":["bogus"]}`},
		{name: "duplicate statistics", datasets: validDatasetJSON, suffix: `,"statistics":{"math":["counts","counts"]}`},
		{name: "unselected chart config", datasets: validDatasetJSON, suffix: `,"charts":{"types":["bar"],"configs":[{"type":"line"}]}`},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			body := `{"datasets":` + test.datasets + test.suffix + `}`
			recorder := s.apiRequest(handler, "/ui", body, "application/json", "text/html")
			s.Equal(http.StatusUnprocessableEntity, recorder.Code, recorder.Body.String())
		})
	}
}

func (s *ServeSuite) TestUIEndpointReportsGenerationFailure() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleUIWithGenerator(w, r, func([]shared.Dataset, []string) (string, error) {
			return "", errors.New("template failed")
		})
	})
	body := `{"datasets":` + validDatasetJSON + `}`
	recorder := s.apiRequest(handler, "/ui", body, "application/json", "text/html")
	s.Equal(http.StatusInternalServerError, recorder.Code)
	s.Equal("application/problem+json", recorder.Header().Get("Content-Type"))
	s.Equal(float64(http.StatusInternalServerError), s.problemStatus(recorder))
	s.NotContains(recorder.Body.String(), "template failed")
	s.Contains(recorder.Body.String(), "could not generate the response")
}

func (s *ServeSuite) TestRequestHelpers() {
	s.Equal([]byte("text"), s.inlineInput(`"text"`))
	s.Equal([]byte(`{"value":1}`), s.inlineInput(`{"value":1}`))
	s.Equal([]byte(`[{"value":1}]`), s.inlineInput(`[{"value":1}]`))
	_, err := inlineInput(json.RawMessage(`42`))
	s.ErrorContains(err, "input must be")
	_, err = inlineInput(nil)
	s.ErrorContains(err, "input must be")

	datasets, validationErr := decodeDatasets(json.RawMessage(validDatasetJSON))
	s.Nil(validationErr)
	s.Len(datasets, 1)
	datasets, validationErr = decodeDatasets(json.RawMessage(`[` + validDatasetJSON + `,` + validDatasetJSON + `]`))
	s.Nil(validationErr)
	s.Len(datasets, 2)
	_, validationErr = decodeDatasets(json.RawMessage(`[invalid]`))
	s.NotNil(validationErr)
	_, validationErr = decodeDatasets(nil)
	s.ErrorContains(validationErr, "datasets is required")

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	s.True(accepts(request, "text/html"))
	request.Header.Set("Accept", "*/*")
	s.True(accepts(request, "text/html"))
	request.Header.Set("Accept", "application/json")
	s.False(accepts(request, "text/html"))
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	s.True(accepts(request, "text/html"))
	request.Header.Set("Accept", "text/*;q=0.8")
	s.True(accepts(request, "text/html"))
	request.Header.Set("Accept", "text/html;q=0,*/*;q=1")
	s.False(accepts(request, "text/html"))
	request.Header.Set("Accept", "text/html;q=bogus")
	s.False(accepts(request, "text/html"))
	request.Header.Set("Accept", "text/html;q=2")
	s.False(accepts(request, "text/html"))
	request.Header.Set("Accept", "not a media range")
	s.False(accepts(request, "text/html"))
	request.Header.Set("Accept", "text")
	s.False(accepts(request, "text/html"))
	request.Header.Set("Accept", "invalid")
	s.False(accepts(request, "invalid"))

	var target map[string]any
	s.Error(strictDecode([]byte(`{} {}`), &target))

	request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"input":"x,y\\na,1\\n"} trailing`))
	request.Header.Set("Content-Type", "application/json")
	s.False(decodeAPIRequest(httptest.NewRecorder(), request, &convertRequest{}))

	_, validationErr = decodeChartConfig(json.RawMessage(`{`), "/charts/configs/0")
	s.NotNil(validationErr)
	validationErr = validateChartConfigValues(json.RawMessage(`{"stat":"invalid"}`), "/charts/configs/0")
	s.NotNil(validationErr)

	cfg := parser.Config{Group: []string{"region"}, GroupPattern: "x"}
	validationErr = applySelectOptions(&cfg, []string{""})
	s.NotNil(validationErr)
	validationErr = applySelectOptions(&cfg, []string{"region{"})
	s.NotNil(validationErr)
	cfg = parser.Config{}
	validationErr = applySelectOptions(&cfg, []string{"region,value,extra", "region,value"})
	s.NotNil(validationErr)
	cfg = parser.Config{}
	validationErr = applySelectOptions(&cfg, []string{"region,value"})
	s.Nil(validationErr)
	pattern := "x,y"
	_, validationErr = buildParserConfig(convertRequest{
		Grouping: &groupingOptions{Pattern: &pattern, Columns: []string{"region/value"}},
	}, "csv")
	s.NotNil(validationErr)

	validationErr = validateChartConfigValues(json.RawMessage(`[]`), "/charts/configs/0")
	s.NotNil(validationErr)
	validationErr = validateChartConfigValues(json.RawMessage(`{"sort":true}`), "/charts/configs/0")
	s.NotNil(validationErr)
	validationErr = validateChartConfigValues(json.RawMessage(`{"stat":{"enabled":true,"math":"invalid"}}`), "/charts/configs/0")
	s.NotNil(validationErr)
	_, _, validationErr = applyUIOptions(nil, chartSelection{Configs: []json.RawMessage{json.RawMessage(`{`)}}, nil)
	s.NotNil(validationErr)

	datasets, validationErr = decodeDatasets(json.RawMessage(validDatasetJSON))
	s.Nil(validationErr)
	var result []shared.Dataset
	var types []string
	result, types, validationErr = applyUIOptions(datasets, chartSelection{Configs: []json.RawMessage{json.RawMessage(`{"type":"line"}`)}}, nil)
	s.Nil(validationErr)
	s.Nil(types)
	s.Len(result[0].Settings, 2)

	for _, test := range []struct {
		name   string
		config malformedChartConfig
	}{
		{name: "marshal existing config", config: malformedChartConfig{chartType: "bar", marshalErr: errors.New("marshal failed")}},
		{name: "decode existing config", config: malformedChartConfig{chartType: "bar", raw: []byte(`{`)}},
	} {
		s.Run(test.name, func() {
			invalidDataset := shared.Dataset{Settings: []internalcharts.ChartConfig{test.config}}
			_, _, validationErr := applyUIOptions([]shared.Dataset{invalidDataset}, chartSelection{Types: []string{"bar"}}, nil)
			s.NotNil(validationErr)
		})
	}

	_, validationErr = decodeStrictDataset(json.RawMessage(`{"name":"Bench","history":[{"tag":"v1","timestamp":"now"}],"axes":[],"settings":[],"data":[]}`), "/datasets/0")
	s.Nil(validationErr)
}

func (s *ServeSuite) TestRequestContractHelpers() {
	s.Run("nested object decoding", func() {
		for _, test := range []struct {
			name   string
			raw    string
			target any
			path   string
		}{
			{name: "null grouping", raw: `null`, target: new(groupingOptions), path: "/grouping"},
			{name: "null charts", raw: `null`, target: new(chartSelection), path: "/charts"},
			{name: "unknown charts field", raw: `{"bogus":true}`, target: new(chartSelection), path: "/charts/bogus"},
			{name: "null output", raw: `null`, target: new(convertOutput), path: "/output"},
			{name: "unknown output field", raw: `{"bogus":true}`, target: new(convertOutput), path: "/output/bogus"},
		} {
			s.Run(test.name, func() {
				err := json.Unmarshal([]byte(test.raw), test.target)
				var validationErr apiValidationError
				s.Require().ErrorAs(err, &validationErr)
				s.Equal(test.path, validationErr.Path)
			})
		}
	})

	s.Run("statistics decoding", func() {
		for _, test := range []struct {
			name string
			raw  string
			path string
		}{
			{name: "null statistics", raw: `{"statistics":null}`, path: "/statistics"},
			{name: "null enabled", raw: `{"statistics":{"enabled":null}}`, path: "/statistics/enabled"},
			{name: "null math", raw: `{"statistics":{"math":null}}`, path: "/statistics/math"},
			{name: "unknown statistics field", raw: `{"statistics":{"bogus":true}}`, path: "/statistics/bogus"},
		} {
			s.Run(test.name, func() {
				var request uiRequest
				err := json.Unmarshal([]byte(test.raw), &request)
				var validationErr apiValidationError
				s.Require().ErrorAs(err, &validationErr)
				s.Equal(test.path, validationErr.Path)
			})
		}

		var omitted uiRequest
		s.Require().NoError(json.Unmarshal([]byte(`{}`), &omitted))
		s.Nil(omitted.Statistics)

		var populated uiRequest
		s.Require().NoError(json.Unmarshal([]byte(`{"statistics":{"enabled":false,"math":["counts"]}}`), &populated))
		s.Require().NotNil(populated.Statistics)
		s.False(*populated.Statistics.Enabled)
		s.Equal([]string{"counts"}, populated.Statistics.Math)
	})

	s.Run("conversion option paths", func() {
		selection := chartSelection{Configs: []json.RawMessage{
			json.RawMessage(`invalid`),
			json.RawMessage(`{"type":"bar","swap":"yx"}`),
		}}
		for _, test := range []struct {
			name string
			path string
		}{
			{name: "filter", path: "/grouping/filter"},
			{name: "grouping", path: "/grouping"},
			{name: "jsonPath", path: "/jsonPath"},
			{name: "select", path: "/select"},
			{name: "swap", path: "/charts/configs/1/swap"},
			{name: "other", path: "/charts/configs"},
		} {
			got := conversionOptionValidationError(selection, &core.OptionError{Name: test.name, Err: errors.New("not applicable")})
			s.Equal(test.path, got.Path)
		}
		s.Equal("/charts/configs", chartConfigFieldPath(selection, "swap", 1))
	})

	s.Run("warning parsing", func() {
		s.Empty(warningFlagName("unstructured warning"))
		s.Empty(warningFlagName(`flag "unterminated`))
		s.Empty(chartFlagJSONKey("not-a-flag"))
		s.Equal("show-labels", warningFlagName(`flag "show-labels" skipped: unsupported`))
		s.Equal("showLabels", chartFlagJSONKey("show-labels"))

		selection := chartSelection{Configs: []json.RawMessage{
			json.RawMessage(`{"type":"bar","showLabels":true}`),
			json.RawMessage(`{"type":"line","showLabels":false}`),
		}}
		errors := conversionWarningValidationErrors(selection, []string{
			`flag "show-labels" skipped: unsupported`,
			`flag "show-labels" skipped: unsupported`,
			"unstructured warning",
		})
		s.Equal([]string{
			"/charts/configs/0/showLabels",
			"/charts/configs/1/showLabels",
			"/charts/configs",
		}, []string{errors[0].Path, errors[1].Path, errors[2].Path})
	})

	s.Run("conversion request decoding", func() {
		for _, test := range []struct {
			name   string
			raw    string
			target any
			path   string
		}{
			{name: "null convert field", raw: `{"theme":null}`, target: new(convertRequest), path: "/theme"},
			{name: "unknown grouping field", raw: `{"unexpected":true}`, target: new(groupingOptions), path: "/grouping/unexpected"},
			{name: "null unit field", raw: `{"memory":null}`, target: new(unitOptions), path: "/units/memory"},
			{name: "unknown unit field", raw: `{"unexpected":true}`, target: new(unitOptions), path: "/units/unexpected"},
		} {
			s.Run(test.name, func() {
				err := json.Unmarshal([]byte(test.raw), test.target)
				var validationErr apiValidationError
				s.Require().ErrorAs(err, &validationErr)
				s.Equal(test.path, validationErr.Path)
			})
		}

		s.Error(rejectNullFields([]byte(`{`), "/", map[string]string{}))
	})

	s.Run("conversion metadata", func() {
		id, name := "run-1", "Example"
		description, tag, theme := "Description", "release", " Westeros "
		metadata, validationErr := buildConvertMetadata(convertRequest{
			ID:          &id,
			Name:        &name,
			Description: &description,
			Tag:         &tag,
			Theme:       &theme,
		})
		s.Require().Nil(validationErr)
		s.Equal(core.Metadata{ID: id, Name: name, Description: description, Tag: tag, Theme: "westeros"}, metadata)

		invalidTheme := "not-a-theme"
		_, _, validationErr = buildConvertInput(convertRequest{Theme: &invalidTheme}, []byte("x,y\\na,1\\n"))
		s.Require().NotNil(validationErr)
		s.Equal("/theme", validationErr.Path)
	})

	s.Run("chart config decoding", func() {
		_, validationErr := decodeChartConfig(json.RawMessage(`{`), "/config")
		s.Require().NotNil(validationErr)
		s.Equal("/config", validationErr.Path)

		_, validationErr = decodeChartConfig(json.RawMessage(`{"type":"bar","showLabels":"yes"}`), "/config")
		s.Require().NotNil(validationErr)
		s.Equal("/config/showLabels", validationErr.Path)

		for _, raw := range []string{
			`{`,
			`{"scale":null}`,
			`{"symbol":1}`,
			`{"symbol":"not-a-symbol"}`,
			`{"sort":true}`,
			`{"sort":{"enabled":null,"order":"asc"}}`,
			`{"stat":true}`,
			`{"stat":{"enabled":"yes","math":[]}}`,
		} {
			s.NotNil(validateChartConfigValues(json.RawMessage(raw), "/config"), raw)
		}
	})

	s.Run("UI materialisation marshal error", func() {
		for _, test := range []struct {
			name   string
			config internalcharts.ChartConfig
		}{
			{
				name:   "marshal base",
				config: testChartConfig{chartType: "bar", marshalErr: errors.New("marshal failed")},
			},
		} {
			s.Run(test.name, func() {
				_, _, validationErr := applyUIOptions(
					[]shared.Dataset{{Settings: []internalcharts.ChartConfig{test.config}}},
					chartSelection{},
					nil,
				)
				s.Require().NotNil(validationErr)
			})
		}
	})

	s.Run("valid history", func() {
		raw := json.RawMessage(`{"name":"Bench","history":[{"tag":"v1","timestamp":"now"}],"axes":[],"settings":[],"data":[]}`)
		datasets, validationErr := decodeDatasets(raw)
		s.Require().Nil(validationErr)
		s.Require().Len(datasets, 1)
		s.Require().Len(datasets[0].History, 1)
		s.Equal("v1", datasets[0].History[0].Tag)
	})
}

func (s *ServeSuite) waitForSignal(signal <-chan struct{}) {
	s.T().Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		s.FailNow("timed out waiting for signal")
	}
}

func (s *ServeSuite) waitForResult(result <-chan error) error {
	s.T().Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(time.Second):
		s.FailNow("timed out waiting for server result")
		return nil
	}
}

func (s *ServeSuite) apiRequest(handler http.Handler, path, body, contentType, accept string) *httptest.ResponseRecorder {
	s.T().Helper()
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if accept != "" {
		request.Header.Set("Accept", accept)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func (s *ServeSuite) problemStatus(recorder *httptest.ResponseRecorder) float64 {
	s.T().Helper()
	var problem map[string]any
	s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &problem))
	return problem["status"].(float64)
}

func (s *ServeSuite) inlineInput(raw string) []byte {
	s.T().Helper()
	input, err := inlineInput(json.RawMessage(raw))
	s.Require().NoError(err)
	return input
}

type testAddr string

func (a testAddr) Network() string { return "tcp" }
func (a testAddr) String() string  { return string(a) }

type malformedChartConfig struct {
	chartType  string
	raw        []byte
	marshalErr error
}

func (c malformedChartConfig) MarshalJSON() ([]byte, error) { return c.raw, c.marshalErr }
func (c malformedChartConfig) ChartType() string            { return c.chartType }
func (malformedChartConfig) StatEnabled() bool              { return false }
func (malformedChartConfig) StatMath() []string             { return nil }
func (malformedChartConfig) SwapString() string             { return "" }

type errorListener struct{ err error }

func (l errorListener) Accept() (net.Conn, error) { return nil, l.err }
func (l errorListener) Close() error              { return nil }
func (l errorListener) Addr() net.Addr            { return testAddr("in-memory") }

type blockingListener struct {
	acceptStarted chan struct{}
	closed        chan struct{}
	once          sync.Once
	acceptErr     error
}

func newBlockingListener() *blockingListener {
	return newBlockingErrorListener(net.ErrClosed)
}

func newBlockingErrorListener(err error) *blockingListener {
	return &blockingListener{
		acceptStarted: make(chan struct{}),
		closed:        make(chan struct{}),
		acceptErr:     err,
	}
}

func (l *blockingListener) Accept() (net.Conn, error) {
	select {
	case <-l.acceptStarted:
	default:
		close(l.acceptStarted)
	}
	<-l.closed
	return nil, l.acceptErr
}

func (l *blockingListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *blockingListener) Addr() net.Addr { return testAddr("in-memory") }

type closeErrorListener struct {
	*blockingListener
	err error
}

func (l closeErrorListener) Close() error {
	_ = l.blockingListener.Close()
	return l.err
}

type singleConnectionListener struct {
	conn          net.Conn
	acceptStarted chan struct{}
	closed        chan struct{}
	acceptOnce    sync.Once
	closeOnce     sync.Once
}

func newSingleConnectionListener(conn net.Conn) *singleConnectionListener {
	return &singleConnectionListener{
		conn:          conn,
		acceptStarted: make(chan struct{}),
		closed:        make(chan struct{}),
	}
}

func (l *singleConnectionListener) Accept() (net.Conn, error) {
	accepted := false
	l.acceptOnce.Do(func() {
		accepted = true
		close(l.acceptStarted)
	})
	if accepted {
		return l.conn, nil
	}
	<-l.closed
	return nil, net.ErrClosed
}

func (l *singleConnectionListener) Close() error {
	l.closeOnce.Do(func() { close(l.closed) })
	return nil
}

func (l *singleConnectionListener) Addr() net.Addr { return testAddr("in-memory") }

func directSignalContext(ctx context.Context, _ ...os.Signal) (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx)
}

type testChartConfig struct {
	chartType  string
	raw        json.RawMessage
	marshalErr error
}

func (c testChartConfig) ChartType() string { return c.chartType }
func (testChartConfig) StatEnabled() bool   { return false }
func (testChartConfig) StatMath() []string  { return nil }
func (testChartConfig) SwapString() string  { return "" }

func (c testChartConfig) MarshalJSON() ([]byte, error) {
	return c.raw, c.marshalErr
}

func TestServeSuite(t *testing.T) {
	suite.Run(t, new(ServeSuite))
}
