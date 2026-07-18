package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/pkg/core"
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
	serveDeps = serveDependencies{
		newHandler:    newRESTHandler,
		listen:        net.Listen,
		signalContext: signal.NotifyContext,
		shutdown:      (*http.Server).Shutdown,
	}
	serveBag = cli.NewFlagBag(serveFlags())
)

var serveCmd = newServeCommand(serveBag, serveDeps)

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

type restRouter struct {
	routes map[string]http.Handler
}

func newRESTHandler() http.Handler {
	return composeRESTRoutes(restHandlers{
		convert: http.HandlerFunc(handleConvert),
		merge:   http.HandlerFunc(handleMerge),
		ui:      http.HandlerFunc(handleUI),
	})
}

func composeRESTRoutes(handlers restHandlers) http.Handler {
	return restRouter{routes: map[string]http.Handler{
		"/":      handlers.convert,
		"/merge": handlers.merge,
		"/ui":    handlers.ui,
	}}
}

func (router restRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, ok := router.routes[r.URL.Path]
	if !ok {
		writeAPIProblem(w, r, http.StatusNotFound, "Not found", "The requested operation does not exist.")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeAPIProblem(w, r, http.StatusMethodNotAllowed, "Method not allowed", "This operation only supports POST.")
		return
	}
	handler.ServeHTTP(w, r)
}

// newServeCommand wires a configured HTTP server to Cobra. Its dependencies
// are explicit so command and lifecycle behavior can be exercised without a
// real listener or operating-system signal.
func newServeCommand(bag *cli.FlagBag, deps serveDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "serve",
		Short:        "Run Vizb's local REST API server",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := deps.signalContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runServer(ctx, serveOptions{
				Host: bag.String("host"),
				Port: bag.Int("port"),
			}, deps)
		},
	}
	bag.Bind(cmd.Flags())
	return cmd
}

func serveFlags() []flags.Flag {
	return []flags.Flag{
		{Name: "host", Default: defaultServeHost, Usage: "Host interface to listen on", Kind: flags.KindString},
		{Name: "port", Shorthand: "p", Default: defaultServePort, Usage: "TCP port to listen on", Kind: flags.KindInt},
	}
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
			if closeErr := server.Close(); closeErr != nil && !errors.Is(closeErr, http.ErrServerClosed) {
				return fmt.Errorf("shutdown HTTP server: %w", errors.Join(err, fmt.Errorf("close HTTP server: %w", closeErr)))
			}
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
	if len(request.Input) == 0 {
		writeValidationProblem(w, r, bodyValidationError("/input", "required", "input is required"))
		return
	}
	if string(request.Input) == "null" {
		writeValidationProblem(w, r, bodyValidationError("/input", "invalid_type", "input must be a string, JSON object, or JSON array"))
		return
	}
	input, err := inlineInput(request.Input)
	if err != nil {
		writeValidationProblem(w, r, bodyValidationError("/input", "invalid_type", err.Error()))
		return
	}
	format := "dataset"
	if request.Output != nil && request.Output.Format != nil {
		format = *request.Output.Format
	}
	if format != "dataset" && format != "html" {
		writeValidationProblem(w, r, bodyValidationError("/output/format", "invalid_enum", "output.format must be dataset or html"))
		return
	}
	responseType := "application/json"
	if format == "html" {
		responseType = "text/html"
	}
	if !accepts(r, responseType) {
		writeAPIProblem(w, r, http.StatusNotAcceptable, "Not acceptable", "Accept must allow "+responseType)
		return
	}

	convertInput, types, validationErr := buildConvertInput(request, input)
	if validationErr != nil {
		writeValidationProblem(w, r, *validationErr)
		return
	}
	result, err := core.Convert(convertInput)
	if err != nil {
		var optionErr *core.OptionError
		if errors.As(err, &optionErr) {
			writeValidationProblem(w, r, conversionOptionValidationError(request.Charts, optionErr))
			return
		}
		writeAPIProblem(w, r, http.StatusUnprocessableEntity, "Input processing failed", err.Error())
		return
	}
	if len(result.Warnings) > 0 {
		writeValidationProblem(w, r, conversionWarningValidationErrors(request.Charts, result.Warnings)...)
		return
	}
	if format == "html" {
		html, err := generateUI([]shared.Dataset{*result.Dataset}, types)
		if err != nil {
			writeInternalServerError(w, r, "generate convert HTML", err)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, html)
		return
	}
	writeAPIJSON(w, http.StatusOK, result.Dataset)
}

func handleMerge(w http.ResponseWriter, r *http.Request) {
	var request mergeRequest
	if !decodeAPIRequest(w, r, &request) {
		return
	}
	if !accepts(r, "application/json") {
		writeAPIProblem(w, r, http.StatusNotAcceptable, "Not acceptable", "Accept must allow application/json")
		return
	}
	if len(request.Datasets) < 2 {
		writeValidationProblem(w, r, bodyValidationError("/datasets", "min_items", "at least two datasets are required"))
		return
	}
	tagAxis := request.TagAxis
	datasets, validationErr := decodeDatasetArray(request.Datasets, "/datasets")
	if validationErr != nil {
		writeValidationProblem(w, r, *validationErr)
		return
	}
	merged, err := core.Merge(datasets, shared.Dimension(tagAxis))
	if err != nil {
		writeValidationProblem(w, r, bodyValidationError("/tagAxis", "invalid_enum", "tagAxis must be one of name, x, y, or z"))
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
		writeValidationProblem(w, r, *err)
		return
	}
	datasets, types, validationErr := applyUIOptions(datasets, request.Charts, request.Statistics)
	if validationErr != nil {
		writeValidationProblem(w, r, *validationErr)
		return
	}
	html, generationErr := generateUI(datasets, types)
	if generationErr != nil {
		writeInternalServerError(w, r, "generate UI HTML", generationErr)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, html)
}

func decodeAPIRequest(w http.ResponseWriter, r *http.Request, target any) bool {
	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if contentType == "" || err != nil || mediaType != "application/json" {
		writeAPIProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported media type", "Content-Type must be application/json")
		return false
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeRequestDecodeProblem(w, r, err)
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			writeMalformedJSONProblem(w, r, bodyValidationError("/", "multiple_values", "request body must contain one JSON value"))
		} else {
			writeRequestDecodeProblem(w, r, err)
		}
		return false
	}
	return true
}

func writeRequestDecodeProblem(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr apiValidationError
	if errors.As(err, &validationErr) {
		writeValidationProblem(w, r, validationErr)
		return
	}
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		writeAPIProblem(w, r, http.StatusRequestEntityTooLarge, "Content too large", fmt.Sprintf("Request body must not exceed %d bytes.", maxBytesErr.Limit))
		return
	}
	if isMalformedJSON(err) {
		writeMalformedJSONProblem(w, r, bodyValidationError("/", "invalid_json", err.Error()))
		return
	}
	writeValidationProblem(w, r, bodyValidationError("/", "invalid_json", err.Error()))
}

func isMalformedJSON(err error) bool {
	var syntaxErr *json.SyntaxError
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &syntaxErr)
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

func accepts(r *http.Request, mediaType string) bool {
	values := r.Header.Values("Accept")
	if len(values) == 0 {
		return true
	}
	targetType, targetSubtype, ok := strings.Cut(mediaType, "/")
	if !ok {
		return false
	}
	bestSpecificity, bestQuality := -1, -1.0
	for _, value := range values {
		for _, mediaRange := range strings.Split(value, ",") {
			acceptedType, params, err := mime.ParseMediaType(strings.TrimSpace(mediaRange))
			if err != nil {
				continue
			}
			quality := 1.0
			if rawQuality, exists := params["q"]; exists {
				quality, err = strconv.ParseFloat(rawQuality, 64)
				if err != nil || quality < 0 || quality > 1 {
					continue
				}
			}
			rangeType, rangeSubtype, ok := strings.Cut(acceptedType, "/")
			if !ok {
				continue
			}
			specificity := -1
			switch {
			case rangeType == targetType && rangeSubtype == targetSubtype:
				specificity = 2
			case rangeType == targetType && rangeSubtype == "*":
				specificity = 1
			case rangeType == "*" && rangeSubtype == "*":
				specificity = 0
			}
			if specificity > bestSpecificity || (specificity == bestSpecificity && quality > bestQuality) {
				bestSpecificity, bestQuality = specificity, quality
			}
		}
	}
	return bestSpecificity >= 0 && bestQuality > 0
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
		"type":     problemType(status),
		"title":    title,
		"status":   status,
		"detail":   detail,
		"instance": r.URL.Path,
	})
}

func writeInternalServerError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	log.Printf("serve: %s: %v", operation, err)
	writeAPIProblem(w, r, http.StatusInternalServerError, "Internal server error", "The server could not generate the response.")
}

func writeValidationProblem(w http.ResponseWriter, r *http.Request, validationErrors ...apiValidationError) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":     "https://vizb.goptics.org/problems/validation",
		"title":    "Unprocessable content",
		"status":   http.StatusUnprocessableEntity,
		"detail":   "Request validation failed.",
		"instance": r.URL.Path,
		"errors":   validationErrors,
	})
}

func writeMalformedJSONProblem(w http.ResponseWriter, r *http.Request, validationErrors ...apiValidationError) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":     problemType(http.StatusBadRequest),
		"title":    "Malformed JSON",
		"status":   http.StatusBadRequest,
		"detail":   "Request body must contain valid JSON.",
		"instance": r.URL.Path,
		"errors":   validationErrors,
	})
}

func problemType(status int) string {
	slug := "internal"
	switch status {
	case http.StatusBadRequest:
		slug = "malformed-json"
	case http.StatusNotFound:
		slug = "not-found"
	case http.StatusMethodNotAllowed:
		slug = "method-not-allowed"
	case http.StatusNotAcceptable:
		slug = "not-acceptable"
	case http.StatusRequestEntityTooLarge:
		slug = "content-too-large"
	case http.StatusUnsupportedMediaType:
		slug = "unsupported-media-type"
	case http.StatusUnprocessableEntity:
		slug = "processing"
	}
	return "https://vizb.goptics.org/problems/" + slug
}
