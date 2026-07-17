package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	internalcharts "github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/pkg/core"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

const (
	defaultServeHost = "127.0.0.1"
	defaultServePort = 8080

	maxRequestBodyBytes = 10 << 20 // 10 MiB
	readHeaderTimeout   = 5 * time.Second
	readTimeout         = 15 * time.Second
	writeTimeout        = 60 * time.Second
	idleTimeout         = 60 * time.Second
	shutdownTimeout     = 10 * time.Second
)

type serveOptions struct {
	Host string
	Port int
}

type serveDependencies struct {
	newHandler    func() http.Handler
	listen        func(network, address string) (net.Listener, error)
	signalContext func(context.Context, ...os.Signal) (context.Context, context.CancelFunc)
	shutdown      func(*http.Server, context.Context) error
}

var (
	serveOpts = serveOptions{Host: defaultServeHost, Port: defaultServePort}
	serveDeps = serveDependencies{
		newHandler:    newRESTHandler,
		listen:        net.Listen,
		signalContext: signal.NotifyContext,
		shutdown:      (*http.Server).Shutdown,
	}
)

var serveCmd = newServeCommand(&serveOpts, serveDeps)

func init() {
	rootCmd.AddCommand(serveCmd)
}

// restHandlers is the composition point for the REST API. Keeping the three
// operations as explicit handlers makes the route map testable and lets future
// transports reuse the request-scoped core without changing server lifecycle
// code.
type restHandlers struct {
	convert http.Handler
	merge   http.Handler
	ui      http.Handler
}

func newRESTHandler() http.Handler {
	return composeRESTRoutes(restHandlers{
		convert: http.HandlerFunc(handleConvert),
		merge:   http.HandlerFunc(handleMerge),
		ui:      http.HandlerFunc(handleUI),
	})
}

func composeRESTRoutes(handlers restHandlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("POST /{$}", handlers.convert)
	mux.Handle("POST /merge", handlers.merge)
	mux.Handle("POST /ui", handlers.ui)
	return mux
}

// newServeCommand wires a configured HTTP server to Cobra. Its dependencies
// are explicit so command and lifecycle behavior can be exercised without a
// real listener or operating-system signal.
func newServeCommand(opts *serveOptions, deps serveDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "serve",
		Short:        "Run Vizb's local REST API server",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := deps.signalContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runServer(ctx, *opts, deps)
		},
	}
	cmd.Flags().StringVar(&opts.Host, "host", defaultServeHost, "Host interface to listen on")
	cmd.Flags().IntVarP(&opts.Port, "port", "p", defaultServePort, "TCP port to listen on")
	return cmd
}

func runServer(ctx context.Context, opts serveOptions, deps serveDependencies) error {
	address, err := serveAddress(opts)
	if err != nil {
		return err
	}

	listener, err := deps.listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}

	server := newHTTPServer(address, deps.newHandler())
	serveResult := make(chan error, 1)
	go func() { serveResult <- server.Serve(listener) }()

	select {
	case err := <-serveResult:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		shutdown := deps.shutdown
		if shutdown == nil {
			shutdown = (*http.Server).Shutdown
		}
		if err := shutdown(server, shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		if err := <-serveResult; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}
}

func newHTTPServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           http.MaxBytesHandler(handler, maxRequestBodyBytes),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func serveAddress(opts serveOptions) (string, error) {
	if opts.Host == "" {
		return "", fmt.Errorf("host must not be empty")
	}
	if opts.Port < 1 || opts.Port > 65535 {
		return "", fmt.Errorf("port must be between 1 and 65535")
	}
	return net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port)), nil
}

type convertRequest struct {
	Input  json.RawMessage `json:"input"`
	Parser string          `json:"parser"`
	Charts struct {
		Types []string `json:"types"`
	} `json:"charts"`
	Output struct {
		Format string `json:"format"`
	} `json:"output"`
}

type mergeRequest struct {
	Datasets []shared.Dataset `json:"datasets"`
	TagAxis  string           `json:"tagAxis"`
}

type uiRequest struct {
	Datasets json.RawMessage `json:"datasets"`
	Charts   struct {
		Types []string `json:"types"`
	} `json:"charts"`
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	handleConvertWithGenerator(w, r, core.GenerateUI)
}

func handleConvertWithGenerator(
	w http.ResponseWriter,
	r *http.Request,
	generateUI func([]shared.Dataset, []string) (string, error),
) {
	var request convertRequest
	if !decodeAPIRequest(w, r, &request) {
		return
	}
	if len(request.Input) == 0 || string(request.Input) == "null" {
		writeAPIProblem(w, r, http.StatusBadRequest, "Invalid request", "input is required")
		return
	}
	input, err := inlineInput(request.Input)
	if err != nil {
		writeAPIProblem(w, r, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}
	types := request.Charts.Types
	if len(types) == 0 {
		types = shared.DefaultChartTypes
	}
	configs := make([]internalcharts.ChartConfig, 0, len(types))
	for _, chartType := range types {
		config, err := internalcharts.Materialise(chartType, nil, nil)
		if err != nil {
			writeAPIProblem(w, r, http.StatusBadRequest, "Invalid request", err.Error())
			return
		}
		configs = append(configs, config)
	}
	result, err := core.Convert(core.ConvertInput{
		Input: input, Parser: request.Parser, Config: parser.Config{}, Charts: configs,
	})
	if err != nil {
		writeAPIProblem(w, r, http.StatusUnprocessableEntity, "Input processing failed", err.Error())
		return
	}
	if request.Output.Format == "html" {
		if !accepts(r, "text/html") {
			writeAPIProblem(w, r, http.StatusNotAcceptable, "Not acceptable", "Accept must allow text/html")
			return
		}
		html, err := generateUI([]shared.Dataset{*result.Dataset}, types)
		if err != nil {
			writeAPIProblem(w, r, http.StatusInternalServerError, "Internal server error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, html)
		return
	}
	if request.Output.Format != "" && request.Output.Format != "dataset" {
		writeAPIProblem(w, r, http.StatusBadRequest, "Invalid request", "output.format must be dataset or html")
		return
	}
	writeAPIJSON(w, http.StatusOK, result.Dataset)
}

func handleMerge(w http.ResponseWriter, r *http.Request) {
	var request mergeRequest
	if !decodeAPIRequest(w, r, &request) {
		return
	}
	if len(request.Datasets) < 2 {
		writeAPIProblem(w, r, http.StatusBadRequest, "Invalid request", "at least two datasets are required")
		return
	}
	merged, err := core.Merge(request.Datasets, shared.Dimension(request.TagAxis))
	if err != nil {
		writeAPIProblem(w, r, http.StatusUnprocessableEntity, "Input processing failed", err.Error())
		return
	}
	writeAPIJSON(w, http.StatusOK, merged)
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	handleUIWithGenerator(w, r, core.GenerateUI)
}

func handleUIWithGenerator(
	w http.ResponseWriter,
	r *http.Request,
	generateUI func([]shared.Dataset, []string) (string, error),
) {
	var request uiRequest
	if !decodeAPIRequest(w, r, &request) {
		return
	}
	if !accepts(r, "text/html") {
		writeAPIProblem(w, r, http.StatusNotAcceptable, "Not acceptable", "Accept must allow text/html")
		return
	}
	datasets, err := decodeDatasets(request.Datasets)
	if err != nil {
		writeAPIProblem(w, r, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}
	html, err := generateUI(datasets, request.Charts.Types)
	if err != nil {
		writeAPIProblem(w, r, http.StatusUnprocessableEntity, "Input processing failed", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, html)
}

func decodeAPIRequest(w http.ResponseWriter, r *http.Request, target any) bool {
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			writeAPIProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported media type", "Content-Type must be application/json")
			return false
		}
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeAPIProblem(w, r, http.StatusBadRequest, "Invalid request", err.Error())
		return false
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		writeAPIProblem(w, r, http.StatusBadRequest, "Invalid request", "request body must contain one JSON value")
		return false
	}
	return true
}

func inlineInput(raw json.RawMessage) ([]byte, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return []byte(text), nil
	}
	if len(raw) == 0 || (raw[0] != '{' && raw[0] != '[') {
		return nil, fmt.Errorf("input must be a string, JSON object, or JSON array")
	}
	return raw, nil
}

func decodeDatasets(raw json.RawMessage) ([]shared.Dataset, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("datasets is required")
	}
	var datasets []shared.Dataset
	if raw[0] == '[' {
		if err := json.Unmarshal(raw, &datasets); err != nil {
			return nil, err
		}
	} else {
		var dataset shared.Dataset
		if err := json.Unmarshal(raw, &dataset); err != nil {
			return nil, err
		}
		datasets = []shared.Dataset{dataset}
	}
	if len(datasets) == 0 {
		return nil, fmt.Errorf("at least one dataset is required")
	}
	return datasets, nil
}

func accepts(r *http.Request, mediaType string) bool {
	accept := r.Header.Get("Accept")
	return accept == "" || accept == "*/*" || accept == mediaType
}

func writeAPIJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAPIProblem(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":     "https://vizb.goptics.org/problems/request",
		"title":    title,
		"status":   status,
		"detail":   detail,
		"instance": r.URL.Path,
	})
}
