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
	"sync"
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

func (s *ServeSuite) TestServeFailureIsReturned() {
	err := runServer(context.Background(), serveOptions{Host: defaultServeHost, Port: defaultServePort}, serveDependencies{
		newHandler: http.NotFoundHandler,
		listen: func(string, string) (net.Listener, error) {
			return errorListener{err: errors.New("accept failed")}, nil
		},
	})
	s.EqualError(err, "serve HTTP: accept failed")
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
	<-listener.acceptStarted
	cancel()
	s.Require().NoError(<-result)
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
}

func (s *ServeSuite) TestConvertEndpointRejectsInvalidRequests() {
	handler := newRESTHandler()
	for _, test := range []struct {
		name        string
		body        string
		contentType string
		accept      string
		wantStatus  int
	}{
		{name: "missing input", body: `{}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "invalid input kind", body: `{"input":42}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "unknown chart", body: `{"input":"x,y\na,1\n","charts":{"types":["unknown"]}}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "unsupported parser", body: `{"input":"x,y\na,1\n","parser":"go"}`, contentType: "application/json", wantStatus: http.StatusUnprocessableEntity},
		{name: "html not accepted", body: `{"input":"x,y\na,1\n","output":{"format":"html"}}`, contentType: "application/json", accept: "application/json", wantStatus: http.StatusNotAcceptable},
		{name: "invalid output format", body: `{"input":"x,y\na,1\n","output":{"format":"csv"}}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "unknown field", body: `{"input":"x,y\na,1\n","extra":true}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "wrong content type", body: `{}`, contentType: "text/plain", wantStatus: http.StatusUnsupportedMediaType},
		{name: "multiple JSON values", body: `{} {}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
	} {
		s.Run(test.name, func() {
			recorder := s.apiRequest(handler, "/", test.body, test.contentType, test.accept)
			s.Equal(test.wantStatus, recorder.Code)
			s.Contains(recorder.Header().Get("Content-Type"), "application/problem+json")
			s.Equal(float64(test.wantStatus), s.problemStatus(recorder))
		})
	}
}

func (s *ServeSuite) TestMergeEndpoint() {
	handler := newRESTHandler()
	recorder := s.apiRequest(handler, "/merge", `{"datasets":[{"name":"one"}]}`, "application/json", "")
	s.Equal(http.StatusBadRequest, recorder.Code)

	recorder = s.apiRequest(handler, "/merge", `{"datasets":[{"name":"Bench","tag":"v1","data":[{"name":"case","yAxis":"1"}]},{"name":"Bench","tag":"v2","data":[{"name":"case","yAxis":"2"}]}],"tagAxis":"x"}`, "application/json", "")
	s.Equal(http.StatusOK, recorder.Code)
	var merged []map[string]any
	s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &merged))
	s.Len(merged, 1)

	recorder = s.apiRequest(handler, "/merge", `{"datasets":[{"name":"one"},{"name":"two"}],"tagAxis":"invalid"}`, "application/json", "")
	s.Equal(http.StatusUnprocessableEntity, recorder.Code)
}

func (s *ServeSuite) TestUIEndpoint() {
	handler := newRESTHandler()
	valid := `{"datasets":{"name":"Bench","data":[{"name":"case","yAxis":"1"}]}}`
	recorder := s.apiRequest(handler, "/ui", valid, "application/json", "text/html")
	s.Equal(http.StatusOK, recorder.Code)
	s.Contains(recorder.Header().Get("Content-Type"), "text/html")
	s.Contains(recorder.Body.String(), "VIZB_DATA")

	for _, test := range []struct {
		name       string
		body       string
		accept     string
		wantStatus int
	}{
		{name: "not accepted", body: valid, accept: "application/json", wantStatus: http.StatusNotAcceptable},
		{name: "missing datasets", body: `{}`, accept: "text/html", wantStatus: http.StatusBadRequest},
		{name: "empty datasets", body: `{"datasets":[]}`, accept: "text/html", wantStatus: http.StatusBadRequest},
		{name: "invalid datasets", body: `{"datasets":true}`, accept: "text/html", wantStatus: http.StatusBadRequest},
	} {
		s.Run(test.name, func() {
			recorder := s.apiRequest(handler, "/ui", test.body, "application/json", test.accept)
			s.Equal(test.wantStatus, recorder.Code)
		})
	}
}

func (s *ServeSuite) TestRequestHelpers() {
	s.Equal([]byte("text"), s.inlineInput(`"text"`))
	s.Equal([]byte(`{"value":1}`), s.inlineInput(`{"value":1}`))
	_, err := inlineInput(json.RawMessage(`42`))
	s.ErrorContains(err, "input must be")
	_, err = inlineInput(nil)
	s.ErrorContains(err, "input must be")

	datasets, err := decodeDatasets(json.RawMessage(`{"name":"Bench"}`))
	s.Require().NoError(err)
	s.Len(datasets, 1)
	_, err = decodeDatasets(nil)
	s.ErrorContains(err, "datasets is required")

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	s.True(accepts(request, "text/html"))
	request.Header.Set("Accept", "*/*")
	s.True(accepts(request, "text/html"))
	request.Header.Set("Accept", "application/json")
	s.False(accepts(request, "text/html"))
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

type errorListener struct{ err error }

func (l errorListener) Accept() (net.Conn, error) { return nil, l.err }
func (l errorListener) Close() error              { return nil }
func (l errorListener) Addr() net.Addr            { return testAddr("in-memory") }

type blockingListener struct {
	acceptStarted chan struct{}
	closed        chan struct{}
	once          sync.Once
}

func newBlockingListener() *blockingListener {
	return &blockingListener{acceptStarted: make(chan struct{}), closed: make(chan struct{})}
}

func (l *blockingListener) Accept() (net.Conn, error) {
	select {
	case <-l.acceptStarted:
	default:
		close(l.acceptStarted)
	}
	<-l.closed
	return nil, net.ErrClosed
}

func (l *blockingListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *blockingListener) Addr() net.Addr { return testAddr("in-memory") }

func directSignalContext(ctx context.Context, _ ...os.Signal) (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx)
}

func TestServeSuite(t *testing.T) {
	suite.Run(t, new(ServeSuite))
}
