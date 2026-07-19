package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"sort"
	"strings"

	internalcharts "github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/pkg/core"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/shared"
)

type groupingOptions struct {
	Pattern *string  `json:"pattern"`
	Regex   string   `json:"regex"`
	Columns []string `json:"columns"`
	Filter  string   `json:"filter"`
}

type unitOptions struct {
	Memory *string `json:"memory"`
	Time   *string `json:"time"`
	Number *string `json:"number"`
}

type chartSelection struct {
	Types   []string          `json:"types"`
	Configs []json.RawMessage `json:"configs"`
}

type statisticsOptions struct {
	Enabled *bool    `json:"enabled"`
	Math    []string `json:"math"`
}

func (o *statisticsOptions) UnmarshalJSON(data []byte) error {
	if err := rejectNullFields(data, "/statistics", map[string]string{
		"enabled": "/statistics/enabled", "math": "/statistics/math",
	}); err != nil {
		return err
	}
	type wire statisticsOptions
	var decoded wire
	if err := strictDecodeRequestObject(data, &decoded, "/statistics"); err != nil {
		return err
	}
	*o = statisticsOptions(decoded)
	return nil
}

type convertRequest struct {
	Input       json.RawMessage  `json:"input"`
	ID          *string          `json:"id"`
	Name        *string          `json:"name"`
	Theme       *string          `json:"theme"`
	Description *string          `json:"description"`
	Tag         *string          `json:"tag"`
	Parser      *string          `json:"parser"`
	Grouping    *groupingOptions `json:"grouping"`
	Units       *unitOptions     `json:"units"`
	Select      []string         `json:"select"`
	JSONPath    string           `json:"jsonPath"`
	Charts      chartSelection   `json:"charts"`
	Output      *convertOutput   `json:"output"`
}

type convertOutput struct {
	Format *string `json:"format"`
}

func (r *convertRequest) UnmarshalJSON(data []byte) error {
	if err := rejectNullFields(data, "/", map[string]string{
		"id": "/id", "name": "/name", "theme": "/theme", "description": "/description",
		"tag": "/tag", "parser": "/parser", "grouping": "/grouping", "units": "/units",
		"select": "/select", "jsonPath": "/jsonPath", "charts": "/charts", "output": "/output",
	}); err != nil {
		return err
	}
	type wire convertRequest
	var decoded wire
	if err := strictDecodeRequestObject(data, &decoded, ""); err != nil {
		return err
	}
	*r = convertRequest(decoded)
	return nil
}

func (o *groupingOptions) UnmarshalJSON(data []byte) error {
	if err := rejectNullFields(data, "/grouping", map[string]string{
		"pattern": "/grouping/pattern", "regex": "/grouping/regex",
		"columns": "/grouping/columns", "filter": "/grouping/filter",
	}); err != nil {
		return err
	}
	type wire groupingOptions
	var decoded wire
	if err := strictDecodeRequestObject(data, &decoded, "/grouping"); err != nil {
		return err
	}
	*o = groupingOptions(decoded)
	return nil
}

func (o *unitOptions) UnmarshalJSON(data []byte) error {
	if err := rejectNullFields(data, "/units", map[string]string{
		"memory": "/units/memory", "time": "/units/time", "number": "/units/number",
	}); err != nil {
		return err
	}
	type wire unitOptions
	var decoded wire
	if err := strictDecodeRequestObject(data, &decoded, "/units"); err != nil {
		return err
	}
	*o = unitOptions(decoded)
	return nil
}

func (s *chartSelection) UnmarshalJSON(data []byte) error {
	if err := rejectNullFields(data, "/charts", map[string]string{
		"types": "/charts/types", "configs": "/charts/configs",
	}); err != nil {
		return err
	}
	type wire chartSelection
	var decoded wire
	if err := strictDecodeRequestObject(data, &decoded, "/charts"); err != nil {
		return err
	}
	*s = chartSelection(decoded)
	return nil
}

func (o *convertOutput) UnmarshalJSON(data []byte) error {
	if err := rejectNullFields(data, "/output", map[string]string{"format": "/output/format"}); err != nil {
		return err
	}
	type wire convertOutput
	var decoded wire
	if err := strictDecodeRequestObject(data, &decoded, "/output"); err != nil {
		return err
	}
	*o = convertOutput(decoded)
	return nil
}

func rejectNullFields(data []byte, nullPath string, paths map[string]string) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return bodyValidationError(nullPath, "invalid_type", nullPath+" must not be null")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	keys := make([]string, 0, len(paths))
	for field := range paths {
		keys = append(keys, field)
	}
	sort.Strings(keys)
	for _, field := range keys {
		path := paths[field]
		if raw, ok := fields[field]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return bodyValidationError(path, "invalid_type", path+" must not be null")
		}
	}
	return nil
}

func strictDecodeRequestObject(data []byte, target any, prefix string) error {
	err := strictDecode(data, target)
	if err == nil {
		return nil
	}
	const unknownFieldPrefix = `json: unknown field "`
	message := err.Error()
	if strings.HasPrefix(message, unknownFieldPrefix) && strings.HasSuffix(message, `"`) {
		field := strings.TrimSuffix(strings.TrimPrefix(message, unknownFieldPrefix), `"`)
		path := prefix + "/" + strings.ReplaceAll(strings.ReplaceAll(field, "~", "~0"), "/", "~1")
		return bodyValidationError(path, "unknown_field", "unknown request field "+field)
	}
	return err
}

type mergeRequest struct {
	Datasets []json.RawMessage `json:"datasets"`
	TagAxis  string            `json:"tagAxis"`
}

type uiRequest struct {
	Datasets   json.RawMessage    `json:"datasets"`
	Charts     chartSelection     `json:"charts"`
	Statistics *statisticsOptions `json:"statistics"`
}

func (r *uiRequest) UnmarshalJSON(data []byte) error {
	if err := rejectNullFields(data, "/", map[string]string{"statistics": "/statistics"}); err != nil {
		return err
	}
	type wire uiRequest
	var decoded wire
	if err := strictDecodeRequestObject(data, &decoded, ""); err != nil {
		return err
	}
	*r = uiRequest(decoded)
	return nil
}

type apiValidationError struct {
	Location string `json:"location"`
	Path     string `json:"path"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

func (e apiValidationError) Error() string {
	return e.Message
}

func bodyValidationError(path, code, message string) apiValidationError {
	return apiValidationError{Location: "body", Path: path, Code: code, Message: message}
}

func conversionOptionValidationError(selection chartSelection, optionErr *core.OptionError) apiValidationError {
	var path string
	switch optionErr.Name {
	case "filter":
		path = "/grouping/filter"
	case "grouping":
		path = "/grouping"
	case "jsonPath":
		path = "/jsonPath"
	case "select":
		path = "/select"
	case "swap":
		path = chartConfigFieldPath(selection, "swap", 0)
	default:
		path = "/charts/configs"
	}
	return bodyValidationError(path, "inapplicable_option", optionErr.Error())
}

func conversionWarningValidationErrors(selection chartSelection, warnings []string) []apiValidationError {
	result := make([]apiValidationError, 0, len(warnings))
	occurrences := map[string]int{}
	for _, warning := range warnings {
		flagName := warningFlagName(warning)
		jsonKey := chartFlagJSONKey(flagName)
		path := "/charts/configs"
		if jsonKey != "" {
			path = chartConfigFieldPath(selection, jsonKey, occurrences[jsonKey])
			occurrences[jsonKey]++
		}
		result = append(result, bodyValidationError(path, "inapplicable_option", warning))
	}
	return result
}

func warningFlagName(warning string) string {
	const prefix = `flag "`
	if !strings.HasPrefix(warning, prefix) {
		return ""
	}
	name, _, ok := strings.Cut(strings.TrimPrefix(warning, prefix), `"`)
	if !ok {
		return ""
	}
	return name
}

func chartFlagJSONKey(flagName string) string {
	for _, chartType := range internalcharts.Registered() {
		for _, flag := range internalcharts.FlagsFor(chartType) {
			if flag.Name == flagName {
				return flag.JSONKey
			}
		}
	}
	return ""
}

func chartConfigFieldPath(selection chartSelection, field string, occurrence int) string {
	matches := make([]string, 0, len(selection.Configs))
	for i, raw := range selection.Configs {
		var values map[string]json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			continue
		}
		if _, ok := values[field]; ok {
			matches = append(matches, fmt.Sprintf("/charts/configs/%d/%s", i, field))
		}
	}
	if occurrence < len(matches) {
		return matches[occurrence]
	}
	return "/charts/configs"
}

func buildConvertInput(request convertRequest, input []byte) (core.ConvertInput, []string, *apiValidationError) {
	key := "auto"
	if request.Parser != nil {
		key = *request.Parser
	}
	if !slices.Contains([]string{"auto", "csv", "json", "go", "javascript", "rust"}, key) {
		err := bodyValidationError("/parser", "invalid_enum", "parser must be one of auto, csv, json, go, javascript, or rust")
		return core.ConvertInput{}, nil, &err
	}

	cfg, validationErr := buildParserConfig(request, key)
	if validationErr != nil {
		return core.ConvertInput{}, nil, validationErr
	}
	configs, types, validationErr := materialiseConversionCharts(request.Charts)
	if validationErr != nil {
		return core.ConvertInput{}, nil, validationErr
	}
	metadata, validationErr := buildConvertMetadata(request)
	if validationErr != nil {
		return core.ConvertInput{}, nil, validationErr
	}

	return core.ConvertInput{
		Input:    input,
		Parser:   key,
		Config:   cfg,
		Metadata: metadata,
		Charts:   configs,
	}, types, nil
}

func buildConvertMetadata(request convertRequest) (core.Metadata, *apiValidationError) {
	metadata := core.Metadata{Name: "Comparisons", Theme: "default"}
	if request.ID != nil {
		metadata.ID = *request.ID
	}
	if request.Name != nil {
		metadata.Name = *request.Name
	}
	if request.Description != nil {
		metadata.Description = *request.Description
	}
	if request.Tag != nil {
		metadata.Tag = *request.Tag
	}
	if request.Theme == nil {
		return metadata, nil
	}
	metadata.Theme = style.NormalizeTheme(*request.Theme)
	if err := style.ValidateTheme(metadata.Theme); err != nil {
		validationErr := bodyValidationError("/theme", "invalid_value", err.Error())
		return core.Metadata{}, &validationErr
	}
	return metadata, nil
}

func buildParserConfig(request convertRequest, key string) (parser.Config, *apiValidationError) {
	cfg := parser.Config{GroupPattern: "x", MemUnit: "B", TimeUnit: "ns", JSONPath: request.JSONPath}
	if request.Grouping != nil {
		if request.Grouping.Pattern != nil {
			cfg.GroupPattern = *request.Grouping.Pattern
		}
		cfg.GroupRegex = request.Grouping.Regex
		cfg.Group = slices.Clone(request.Grouping.Columns)
		cfg.Filter = request.Grouping.Filter
	}
	if err := parser.ValidateGroupPattern(cfg.GroupPattern); err != nil {
		validationErr := bodyValidationError("/grouping/pattern", "invalid_value", err.Error())
		return cfg, &validationErr
	}
	for i, column := range cfg.Group {
		if strings.TrimSpace(column) == "" {
			validationErr := bodyValidationError(fmt.Sprintf("/grouping/columns/%d", i), "min_length", "grouping columns must not be empty")
			return cfg, &validationErr
		}
	}
	if cfg.GroupRegex != "" {
		if _, err := regexp.Compile(cfg.GroupRegex); err != nil {
			validationErr := bodyValidationError("/grouping/regex", "invalid_regex", err.Error())
			return cfg, &validationErr
		}
	}
	if cfg.Filter != "" {
		if _, err := regexp.Compile(cfg.Filter); err != nil {
			validationErr := bodyValidationError("/grouping/filter", "invalid_regex", err.Error())
			return cfg, &validationErr
		}
	}

	if request.Units != nil {
		if request.Units.Memory != nil && !slices.Contains([]string{"b", "B", "KB", "MB", "GB"}, *request.Units.Memory) {
			validationErr := bodyValidationError("/units/memory", "invalid_enum", "memory unit must be one of b, B, KB, MB, or GB")
			return cfg, &validationErr
		}
		if request.Units.Time != nil && !slices.Contains([]string{"ns", "us", "ms", "s"}, *request.Units.Time) {
			validationErr := bodyValidationError("/units/time", "invalid_enum", "time unit must be one of ns, us, ms, or s")
			return cfg, &validationErr
		}
		if request.Units.Number != nil && !slices.Contains([]string{"K", "M", "B", "T"}, *request.Units.Number) {
			validationErr := bodyValidationError("/units/number", "invalid_enum", "number unit must be one of K, M, B, or T")
			return cfg, &validationErr
		}
		if request.Units.Memory != nil {
			cfg.MemUnit = *request.Units.Memory
		}
		if request.Units.Time != nil {
			cfg.TimeUnit = *request.Units.Time
		}
		if request.Units.Number != nil {
			cfg.NumberUnit = *request.Units.Number
		}
	}

	var err error
	cfg, err = parser.ResolveGroupConfig(cfg)
	if err != nil {
		validationErr := bodyValidationError("/grouping", "invalid_grouping", err.Error())
		return cfg, &validationErr
	}
	if validationErr := applySelectOptions(&cfg, request.Select); validationErr != nil {
		return cfg, validationErr
	}
	cfg.Mode = parser.ResolveMode(cfg)
	return cfg, nil
}

func applySelectOptions(cfg *parser.Config, rawSelect []string) *apiValidationError {
	if len(rawSelect) == 0 {
		return nil
	}
	if parser.IsExplicitGrouping(*cfg) {
		seen := map[string]bool{}
		for i, raw := range rawSelect {
			if strings.TrimSpace(raw) == "" {
				err := bodyValidationError(fmt.Sprintf("/select/%d", i), "min_length", "select expressions must not be empty")
				return &err
			}
			selected, parseErr := parser.ParseSelectFlag(raw)
			if parseErr != nil {
				err := bodyValidationError(fmt.Sprintf("/select/%d", i), "invalid_select", parseErr.Error())
				return &err
			}
			for _, column := range selected {
				if seen[column.Source] {
					err := bodyValidationError(fmt.Sprintf("/select/%d", i), "duplicate_value", "duplicate selected column "+column.Source)
					return &err
				}
				seen[column.Source] = true
				cfg.Select = append(cfg.Select, column)
			}
		}
		groupSet := map[string]bool{}
		for _, column := range parser.EffectiveGroupColumns(*cfg) {
			groupSet[column] = true
		}
		for _, column := range cfg.Select {
			if groupSet[column.Source] {
				err := bodyValidationError("/select", "conflicting_option", "column "+column.Source+" cannot be in both select and grouping.columns")
				return &err
			}
		}
		return nil
	}

	for i, raw := range rawSelect {
		if strings.TrimSpace(raw) == "" {
			err := bodyValidationError(fmt.Sprintf("/select/%d", i), "min_length", "select expressions must not be empty")
			return &err
		}
		view, parseErr := parser.ParseSelectViewFlag(raw)
		if parseErr != nil {
			err := bodyValidationError(fmt.Sprintf("/select/%d", i), "invalid_select", parseErr.Error())
			return &err
		}
		cfg.SelectViews = append(cfg.SelectViews, view)
	}
	if len(cfg.SelectViews) > 1 {
		if err := parser.ValidateMultiSelectStatViews(cfg.SelectViews); err != nil {
			validationErr := bodyValidationError("/select", "invalid_select", err.Error())
			return &validationErr
		}
	}
	return nil
}

func materialiseConversionCharts(selection chartSelection) ([]internalcharts.ChartConfig, []string, *apiValidationError) {
	types, validationErr := validateChartTypes(selection.Types, true)
	if validationErr != nil {
		return nil, nil, validationErr
	}
	overrides, _, validationErr := decodeChartOverrides(selection.Configs, types)
	if validationErr != nil {
		return nil, nil, validationErr
	}
	configs := make([]internalcharts.ChartConfig, 0, len(types))
	for _, chartType := range types {
		// chartType and overrides are validated above, so Materialise cannot fail.
		config, _ := internalcharts.Materialise(chartType, nil, overrides[chartType])
		configs = append(configs, config)
	}
	return configs, types, nil
}

func validateChartTypes(input []string, conversionDefaults bool) ([]string, *apiValidationError) {
	if input != nil && len(input) == 0 {
		err := bodyValidationError("/charts/types", "min_items", "charts.types must contain at least one chart type")
		return nil, &err
	}
	types := slices.Clone(input)
	if types == nil && conversionDefaults {
		types = slices.Clone(shared.DefaultChartTypes)
	}
	seen := map[string]bool{}
	for i, chartType := range types {
		if _, ok := internalcharts.Get(chartType); !ok {
			err := bodyValidationError(fmt.Sprintf("/charts/types/%d", i), "invalid_enum", "unknown chart type "+chartType)
			return nil, &err
		}
		if seen[chartType] {
			err := bodyValidationError(fmt.Sprintf("/charts/types/%d", i), "duplicate_value", "chart types must be unique")
			return nil, &err
		}
		seen[chartType] = true
	}
	return types, nil
}

func decodeChartOverrides(rawConfigs []json.RawMessage, selected []string) (map[string]internalcharts.ChartConfig, []string, *apiValidationError) {
	overrides := map[string]internalcharts.ChartConfig{}
	order := make([]string, 0, len(rawConfigs))
	for i, raw := range rawConfigs {
		config, validationErr := decodeChartConfig(raw, fmt.Sprintf("/charts/configs/%d", i))
		if validationErr != nil {
			return nil, nil, validationErr
		}
		chartType := config.ChartType()
		if _, exists := overrides[chartType]; exists {
			err := bodyValidationError(fmt.Sprintf("/charts/configs/%d/type", i), "duplicate_value", "only one config is allowed per chart type")
			return nil, nil, &err
		}
		if selected != nil && !slices.Contains(selected, chartType) {
			err := bodyValidationError(fmt.Sprintf("/charts/configs/%d/type", i), "inapplicable_option", "chart config type must also appear in charts.types")
			return nil, nil, &err
		}
		overrides[chartType] = config
		order = append(order, chartType)
	}
	return overrides, order, nil
}

func decodeChartConfig(raw json.RawMessage, path string) (internalcharts.ChartConfig, *apiValidationError) {
	var discriminator struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &discriminator); err != nil {
		validationErr := bodyValidationError(path, "invalid_chart_config", err.Error())
		return nil, &validationErr
	}
	if discriminator.Type == "" {
		validationErr := bodyValidationError(path+"/type", "required", "chart config type is required")
		return nil, &validationErr
	}
	config, err := internalcharts.New(discriminator.Type)
	if err != nil {
		validationErr := bodyValidationError(path+"/type", "invalid_enum", err.Error())
		return nil, &validationErr
	}
	if err := strictDecodeRequestObject(raw, config, path); err != nil {
		var validationErr apiValidationError
		if errors.As(err, &validationErr) {
			return nil, &validationErr
		}
		validationErr = bodyValidationError(path, "invalid_chart_config", err.Error())
		return nil, &validationErr
	}
	if validationErr := validateChartConfigValues(raw, path); validationErr != nil {
		return nil, validationErr
	}
	return config, nil
}

func validateChartConfigValues(raw json.RawMessage, path string) *apiValidationError {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		validationErr := bodyValidationError(path, "invalid_chart_config", err.Error())
		return &validationErr
	}
	if validationErr := rejectNullObjectValues(values, path); validationErr != nil {
		return validationErr
	}
	if scaleRaw, ok := values["scale"]; ok {
		var scale string
		if err := json.Unmarshal(scaleRaw, &scale); err != nil || (scale != "linear" && scale != "log") {
			validationErr := bodyValidationError(path+"/scale", "invalid_enum", "scale must be linear or log")
			return &validationErr
		}
	}
	if symbolRaw, ok := values["symbol"]; ok {
		var symbol string
		if err := json.Unmarshal(symbolRaw, &symbol); err != nil {
			validationErr := bodyValidationError(path+"/symbol", "invalid_type", "symbol must be a string")
			return &validationErr
		}
		if err := internalcharts.ValidateSymbolValue(symbol); err != nil {
			validationErr := bodyValidationError(path+"/symbol", "invalid_value", err.Error())
			return &validationErr
		}
	}
	if sizeRaw, ok := values["symbolSize"]; ok {
		var size float64
		if err := json.Unmarshal(sizeRaw, &size); err != nil || size <= 0 {
			validationErr := bodyValidationError(path+"/symbolSize", "exclusive_minimum", "symbolSize must be greater than zero")
			return &validationErr
		}
	}
	if sortRaw, ok := values["sort"]; ok {
		var sortFields map[string]json.RawMessage
		if err := json.Unmarshal(sortRaw, &sortFields); err != nil {
			validationErr := bodyValidationError(path+"/sort", "invalid_value", "sort must be an object")
			return &validationErr
		}
		if validationErr := rejectNullObjectValues(sortFields, path+"/sort"); validationErr != nil {
			return validationErr
		}
		if _, ok := sortFields["enabled"]; !ok {
			validationErr := bodyValidationError(path+"/sort/enabled", "required", "sort.enabled is required")
			return &validationErr
		}
		orderRaw, ok := sortFields["order"]
		if !ok {
			validationErr := bodyValidationError(path+"/sort/order", "required", "sort.order is required")
			return &validationErr
		}
		var order string
		if err := json.Unmarshal(orderRaw, &order); err != nil || (order != "asc" && order != "desc") {
			validationErr := bodyValidationError(path+"/sort/order", "invalid_enum", "sort.order must be asc or desc")
			return &validationErr
		}
	}
	if statRaw, ok := values["stat"]; ok {
		var statFields map[string]json.RawMessage
		if err := json.Unmarshal(statRaw, &statFields); err != nil {
			validationErr := bodyValidationError(path+"/stat", "invalid_value", "stat must be an object")
			return &validationErr
		}
		if validationErr := rejectNullObjectValues(statFields, path+"/stat"); validationErr != nil {
			return validationErr
		}
		if _, ok := statFields["enabled"]; !ok {
			validationErr := bodyValidationError(path+"/stat/enabled", "required", "stat.enabled is required")
			return &validationErr
		}
		if _, ok := statFields["math"]; !ok {
			validationErr := bodyValidationError(path+"/stat/math", "required", "stat.math is required")
			return &validationErr
		}
		var stat shared.StatConfig
		if err := json.Unmarshal(statRaw, &stat); err != nil {
			validationErr := bodyValidationError(path+"/stat", "invalid_value", err.Error())
			return &validationErr
		}
		if validationErr := validateStatMath(stat.Math, path+"/stat/math"); validationErr != nil {
			return validationErr
		}
	}
	return nil
}

func rejectNullObjectValues(values map[string]json.RawMessage, path string) *apiValidationError {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if bytes.Equal(bytes.TrimSpace(values[key]), []byte("null")) {
			validationErr := bodyValidationError(path+"/"+key, "invalid_type", path+"/"+key+" must not be null")
			return &validationErr
		}
	}
	return nil
}

func validateStatMath(math []string, path string) *apiValidationError {
	seen := map[string]bool{}
	for i, value := range math {
		if !slices.Contains(shared.ValidStatMath, value) {
			err := bodyValidationError(fmt.Sprintf("%s/%d", path, i), "invalid_enum", "unknown statistics category "+value)
			return &err
		}
		if seen[value] {
			err := bodyValidationError(fmt.Sprintf("%s/%d", path, i), "duplicate_value", "statistics categories must be unique")
			return &err
		}
		seen[value] = true
	}
	return nil
}

func applyUIOptions(datasets []shared.Dataset, selection chartSelection, statistics *statisticsOptions) ([]shared.Dataset, []string, *apiValidationError) {
	types, validationErr := validateChartTypes(selection.Types, false)
	if validationErr != nil {
		return nil, nil, validationErr
	}
	overrides, overrideOrder, validationErr := decodeChartOverrides(selection.Configs, types)
	if validationErr != nil {
		return nil, nil, validationErr
	}
	var stat *shared.StatConfig
	if statistics != nil {
		if validationErr := validateStatMath(statistics.Math, "/statistics/math"); validationErr != nil {
			return nil, nil, validationErr
		}
		enabled := true
		if statistics.Enabled != nil {
			enabled = *statistics.Enabled
		}
		stat = &shared.StatConfig{Enabled: enabled, Math: slices.Clone(statistics.Math)}
	}

	result := slices.Clone(datasets)
	for i := range result {
		existing := map[string]internalcharts.ChartConfig{}
		existingOrder := make([]string, 0, len(result[i].Settings))
		for _, config := range result[i].Settings {
			existing[config.ChartType()] = config
			existingOrder = append(existingOrder, config.ChartType())
		}
		targetTypes := types
		if targetTypes == nil {
			targetTypes = slices.Clone(existingOrder)
			for _, chartType := range overrideOrder {
				if !slices.Contains(targetTypes, chartType) {
					targetTypes = append(targetTypes, chartType)
				}
			}
		}
		configs := make([]internalcharts.ChartConfig, 0, len(targetTypes))
		for _, chartType := range targetTypes {
			base, hasBase := existing[chartType]
			override := overrides[chartType]
			seed := map[string]any{}
			if hasBase {
				raw, err := json.Marshal(base)
				if err != nil {
					validationErr := bodyValidationError("/datasets", "invalid_dataset", err.Error())
					return nil, nil, &validationErr
				}
				_ = json.Unmarshal(raw, &seed) // json.Marshal always produces valid JSON.
			}
			if stat != nil {
				seed["stat"] = stat
			}
			// chartType and override were validated before assembling the seed.
			config, _ := internalcharts.Materialise(chartType, seed, override)
			configs = append(configs, config)
		}
		result[i].Settings = configs
	}
	return result, types, nil
}

type datasetWire struct {
	ID           string              `json:"id"`
	Tag          string              `json:"tag"`
	Timestamp    string              `json:"timestamp"`
	Name         *string             `json:"name"`
	Theme        string              `json:"theme"`
	History      []historyWire       `json:"history"`
	Description  string              `json:"description"`
	Meta         *shared.Meta        `json:"meta"`
	Axes         *[]axisWire         `json:"axes"`
	Settings     *[]json.RawMessage  `json:"settings"`
	Data         *[]shared.DataPoint `json:"data"`
	PreserveRows bool                `json:"preserveRows"`
}

type historyWire struct {
	Tag       *string      `json:"tag"`
	Timestamp *string      `json:"timestamp"`
	Meta      *shared.Meta `json:"meta"`
}

type axisWire struct {
	Key   *string `json:"key"`
	Label string  `json:"label"`
	Type  string  `json:"type"`
}

func decodeStrictDataset(raw json.RawMessage, path string) (shared.Dataset, *apiValidationError) {
	var wire datasetWire
	if err := strictDecode(raw, &wire); err != nil {
		validationErr := bodyValidationError(path, "invalid_dataset", err.Error())
		return shared.Dataset{}, &validationErr
	}
	switch {
	case wire.Name == nil:
		validationErr := bodyValidationError(path+"/name", "required", "dataset name is required")
		return shared.Dataset{}, &validationErr
	case wire.Axes == nil:
		validationErr := bodyValidationError(path+"/axes", "required", "dataset axes are required")
		return shared.Dataset{}, &validationErr
	case wire.Settings == nil:
		validationErr := bodyValidationError(path+"/settings", "required", "dataset settings are required")
		return shared.Dataset{}, &validationErr
	case wire.Data == nil:
		validationErr := bodyValidationError(path+"/data", "required", "dataset data is required")
		return shared.Dataset{}, &validationErr
	}

	axes := make([]shared.Axis, 0, len(*wire.Axes))
	for i, axis := range *wire.Axes {
		axisPath := fmt.Sprintf("%s/axes/%d", path, i)
		if axis.Key == nil {
			validationErr := bodyValidationError(axisPath+"/key", "required", "axis key is required")
			return shared.Dataset{}, &validationErr
		}
		if !slices.Contains([]string{"name", "x", "y", "z"}, *axis.Key) {
			validationErr := bodyValidationError(axisPath+"/key", "invalid_enum", "axis key must be one of name, x, y, or z")
			return shared.Dataset{}, &validationErr
		}
		if axis.Type != "" && axis.Type != "value" {
			validationErr := bodyValidationError(axisPath+"/type", "invalid_enum", "axis type must be empty or value")
			return shared.Dataset{}, &validationErr
		}
		axes = append(axes, shared.Axis{Key: *axis.Key, Label: axis.Label, Type: axis.Type})
	}

	settings := make([]internalcharts.ChartConfig, 0, len(*wire.Settings))
	settingTypes := make(map[string]bool, len(*wire.Settings))
	for i, rawConfig := range *wire.Settings {
		configPath := fmt.Sprintf("%s/settings/%d", path, i)
		config, validationErr := decodeChartConfig(rawConfig, configPath)
		if validationErr != nil {
			return shared.Dataset{}, validationErr
		}
		if settingTypes[config.ChartType()] {
			validationErr := bodyValidationError(configPath+"/type", "duplicate_value", "dataset setting types must be unique")
			return shared.Dataset{}, &validationErr
		}
		settingTypes[config.ChartType()] = true
		settings = append(settings, config)
	}

	history := make([]shared.HistoryEntry, 0, len(wire.History))
	for i, entry := range wire.History {
		entryPath := fmt.Sprintf("%s/history/%d", path, i)
		if entry.Tag == nil {
			validationErr := bodyValidationError(entryPath+"/tag", "required", "history tag is required")
			return shared.Dataset{}, &validationErr
		}
		if entry.Timestamp == nil {
			validationErr := bodyValidationError(entryPath+"/timestamp", "required", "history timestamp is required")
			return shared.Dataset{}, &validationErr
		}
		history = append(history, shared.HistoryEntry{Tag: *entry.Tag, Timestamp: *entry.Timestamp, Meta: entry.Meta})
	}

	return shared.Dataset{
		ID:           wire.ID,
		Tag:          wire.Tag,
		Timestamp:    wire.Timestamp,
		Name:         *wire.Name,
		Theme:        wire.Theme,
		History:      history,
		Description:  wire.Description,
		Meta:         wire.Meta,
		Axes:         axes,
		Settings:     settings,
		Data:         slices.Clone(*wire.Data),
		PreserveRows: wire.PreserveRows,
	}, nil
}

func decodeDatasetArray(rawDatasets []json.RawMessage, path string) ([]shared.Dataset, *apiValidationError) {
	datasets := make([]shared.Dataset, 0, len(rawDatasets))
	for i, raw := range rawDatasets {
		dataset, validationErr := decodeStrictDataset(raw, fmt.Sprintf("%s/%d", path, i))
		if validationErr != nil {
			return nil, validationErr
		}
		datasets = append(datasets, dataset)
	}
	return datasets, nil
}

func decodeDatasets(raw json.RawMessage) ([]shared.Dataset, *apiValidationError) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		validationErr := bodyValidationError("/datasets", "required", "datasets is required")
		return nil, &validationErr
	}
	if trimmed[0] == '[' {
		var rawDatasets []json.RawMessage
		if err := strictDecode(trimmed, &rawDatasets); err != nil {
			validationErr := bodyValidationError("/datasets", "invalid_dataset", err.Error())
			return nil, &validationErr
		}
		if len(rawDatasets) == 0 {
			validationErr := bodyValidationError("/datasets", "min_items", "at least one dataset is required")
			return nil, &validationErr
		}
		return decodeDatasetArray(rawDatasets, "/datasets")
	}
	if trimmed[0] != '{' {
		validationErr := bodyValidationError("/datasets", "invalid_type", "datasets must be a Dataset object or array")
		return nil, &validationErr
	}
	dataset, validationErr := decodeStrictDataset(trimmed, "/datasets")
	if validationErr != nil {
		return nil, validationErr
	}
	return []shared.Dataset{dataset}, nil
}

func strictDecode(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("must contain one JSON value")
	}
	return nil
}
