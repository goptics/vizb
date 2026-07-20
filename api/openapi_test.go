package api

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	bar "github.com/goptics/vizb/internal/charts/bar"
	heatmap "github.com/goptics/vizb/internal/charts/heatmap"
	line "github.com/goptics/vizb/internal/charts/line"
	pie "github.com/goptics/vizb/internal/charts/pie"
	radar "github.com/goptics/vizb/internal/charts/radar"
	scatter "github.com/goptics/vizb/internal/charts/scatter"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

type OpenAPISuite struct {
	suite.Suite
}

func (s *OpenAPISuite) TestOpenAPIContract() {
	t := s.T()
	contract := readContract(t)
	if got := contract["openapi"]; got != "3.1.1" {
		t.Fatalf("openapi version = %#v, want 3.1.1", got)
	}

	paths := mustMap(t, contract["paths"], "paths")
	if len(paths) != 3 {
		t.Fatalf("paths has %d entries, want exactly three", len(paths))
	}
	for _, path := range []string{"/", "/merge", "/ui"} {
		pathItem := mustMap(t, paths[path], "paths."+path)
		if len(pathItem) != 1 || pathItem["post"] == nil {
			t.Fatalf("%s must expose only POST, got %#v", path, pathItem)
		}
	}

	verifyReferences(t, contract, contract, "#")
	verifyOperationExamples(t, contract)
}

func (s *OpenAPISuite) TestReusableSchemasMatchGoWireTypes() {
	t := s.T()
	contract := readContract(t)
	schemas := mustMap(t, mustMap(t, contract["components"], "components")["schemas"], "components.schemas")

	for schemaName, value := range map[string]any{
		"Dataset":            shared.Dataset{},
		"HistoryEntry":       shared.HistoryEntry{},
		"Meta":               shared.Meta{},
		"CPUInfo":            shared.CPUInfo{},
		"Axis":               shared.Axis{},
		"DataPoint":          shared.DataPoint{},
		"Stat":               shared.Stat{},
		"Sort":               shared.Sort{},
		"StatisticsConfig":   shared.StatConfig{},
		"BarChartConfig":     bar.Config{},
		"LineChartConfig":    line.Config{},
		"ScatterChartConfig": scatter.Config{},
		"PieChartConfig":     pie.Config{},
		"HeatmapChartConfig": heatmap.Config{},
		"RadarChartConfig":   radar.Config{},
	} {
		schema := mustMap(t, schemas[schemaName], "components.schemas."+schemaName)
		got := propertyNames(t, schema, schemaName)
		want := jsonFieldNames(value)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s properties = %v, want Go wire fields %v", schemaName, got, want)
		}
	}

	for schemaName, required := range map[string][]string{
		"Dataset":            {"name", "axes", "settings", "data"},
		"HistoryEntry":       {"tag", "timestamp"},
		"Axis":               {"key"},
		"Sort":               {"enabled", "order"},
		"StatisticsConfig":   {"enabled", "math"},
		"BarChartConfig":     {"type"},
		"LineChartConfig":    {"type"},
		"ScatterChartConfig": {"type"},
		"PieChartConfig":     {"type"},
		"HeatmapChartConfig": {"type"},
		"RadarChartConfig":   {"type"},
	} {
		schema := mustMap(t, schemas[schemaName], "components.schemas."+schemaName)
		if got := stringSliceValue(schema["required"]); !reflect.DeepEqual(got, sorted(required)) {
			t.Errorf("%s required = %v, want %v", schemaName, got, sorted(required))
		}
	}
}

func (s *OpenAPISuite) TestMergeDeclaresNotAcceptableResponse() {
	contract := readContract(s.T())
	paths := mustMap(s.T(), contract["paths"], "paths")
	merge := mustMap(s.T(), mustMap(s.T(), paths["/merge"], "paths./merge")["post"], "paths./merge.post")
	responses := mustMap(s.T(), merge["responses"], "paths./merge.post.responses")
	notAcceptable := mustMap(s.T(), responses["406"], "paths./merge.post.responses.406")
	s.Equal("#/components/responses/NotAcceptableProblem", notAcceptable["$ref"])
}

func (s *OpenAPISuite) TestOperationsDeclareRequestProblemResponses() {
	contract := readContract(s.T())
	paths := mustMap(s.T(), contract["paths"], "paths")
	for _, path := range []string{"/", "/merge", "/ui"} {
		s.Run(path, func() {
			operation := mustMap(s.T(), mustMap(s.T(), paths[path], "paths."+path)["post"], "paths."+path+".post")
			responses := mustMap(s.T(), operation["responses"], "paths."+path+".post.responses")
			for status, responseName := range map[string]string{
				"400": "MalformedJSONProblem",
				"413": "ContentTooLargeProblem",
				"422": "UnprocessableContentProblem",
			} {
				response := mustMap(s.T(), responses[status], "paths."+path+".post.responses."+status)
				s.Equal("#/components/responses/"+responseName, response["$ref"])
			}
		})
	}
}

func (s *OpenAPISuite) TestDatasetSettingsRequireUniqueChartTypes() {
	contract := readContract(s.T())
	components := mustMap(s.T(), contract["components"], "components")
	schemas := mustMap(s.T(), components["schemas"], "components.schemas")
	datasetSchema := schemas["Dataset"]
	dataset := map[string]any{
		"name": "Bench",
		"axes": []any{},
		"settings": []any{
			map[string]any{"type": "bar"},
			map[string]any{"type": "line"},
		},
		"data": []any{},
	}
	s.NoError(validateSchema(contract, datasetSchema, dataset, "dataset"))

	dataset["settings"] = []any{
		map[string]any{"type": "bar"},
		map[string]any{"type": "bar", "showLabels": true},
	}
	s.ErrorContains(validateSchema(contract, datasetSchema, dataset, "dataset"), "want at most 1")
}

func (s *OpenAPISuite) TestUIContractRejectsRemoteInputAndReturnsHTML() {
	contract := readContract(s.T())
	paths := mustMap(s.T(), contract["paths"], "paths")
	operation := mustMap(s.T(), mustMap(s.T(), paths["/ui"], "paths./ui")["post"], "paths./ui.post")

	requestBody, err := dereference(contract, operation["requestBody"])
	s.Require().NoError(err)
	requestContent := mustMap(s.T(), mustMap(s.T(), requestBody, "UI request body")["content"], "UI request content")
	requestSchema := mustMap(s.T(), requestContent["application/json"], "UI JSON request")["schema"]
	request := map[string]any{
		"datasets": map[string]any{
			"name":     "Bench",
			"axes":     []any{},
			"settings": []any{},
			"data":     []any{},
		},
		"dataUrl": "https://example.com/data.json",
	}
	s.ErrorContains(validateSchema(contract, requestSchema, request, "UI request"), `unknown property "dataUrl"`)

	responses := mustMap(s.T(), operation["responses"], "paths./ui.post.responses")
	success, err := dereference(contract, responses["200"])
	s.Require().NoError(err)
	successContent := mustMap(s.T(), mustMap(s.T(), success, "UI success response")["content"], "UI success content")
	s.Len(successContent, 1)
	s.NotNil(successContent["text/html"])
}

func (s *OpenAPISuite) TestLineAndScatterSymbolsMatchServerValidation() {
	contract := readContract(s.T())
	schemas := mustMap(s.T(), mustMap(s.T(), contract["components"], "components")["schemas"], "components.schemas")
	for _, schemaName := range []string{"LineChartConfig", "ScatterChartConfig"} {
		s.Run(schemaName, func() {
			schema := schemas[schemaName]
			for _, symbol := range []string{"circle", "CIRCLE", "image://marker.svg", "path://M0 0", "M0 0"} {
				s.NoError(validateSchema(contract, schema, map[string]any{"type": strings.TrimSuffix(strings.ToLower(schemaName), "chartconfig"), "symbol": symbol}, schemaName))
			}
			s.Error(validateSchema(contract, schema, map[string]any{"type": strings.TrimSuffix(strings.ToLower(schemaName), "chartconfig"), "symbol": "star"}, schemaName))
		})
	}
}

func (s *OpenAPISuite) TestRootConversionContract() {
	t := s.T()
	contract := readContract(t)
	components := mustMap(t, contract["components"], "components")
	schemas := mustMap(t, components["schemas"], "components.schemas")
	request := mustMap(t, schemas["ConvertRequest"], "components.schemas.ConvertRequest")
	s.Equal(
		[]string{"charts", "description", "grouping", "id", "input", "jsonPath", "name", "output", "parser", "select", "tag", "theme", "units"},
		propertyNames(t, request, "ConvertRequest"),
	)
	s.Equal([]string{"input"}, stringSliceValue(request["required"]))
	s.NotContains(propertyNames(t, request, "ConvertRequest"), "metadata")

	paths := mustMap(t, contract["paths"], "paths")
	root := mustMap(t, mustMap(t, paths["/"], "paths./")["post"], "paths./.post")
	responses := mustMap(t, root["responses"], "paths./.post.responses")
	s.Equal([]string{"200", "400", "406", "413", "415", "422", "500"}, sortedMapKeys(responses))

	allResponses := mustMap(t, components["responses"], "components.responses")
	success := mustMap(t, allResponses["ConvertSuccess"], "components.responses.ConvertSuccess")
	content := mustMap(t, success["content"], "components.responses.ConvertSuccess.content")
	s.Equal([]string{"application/json", "text/html"}, sortedMapKeys(content))
	jsonResponse := mustMap(t, content["application/json"], "ConvertSuccess.application/json")
	jsonSchema := mustMap(t, jsonResponse["schema"], "ConvertSuccess.application/json.schema")
	s.Equal("#/components/schemas/Dataset", jsonSchema["$ref"])
}

func TestOpenAPISuite(t *testing.T) {
	suite.Run(t, new(OpenAPISuite))
}

func readContract(t *testing.T) map[string]any {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locating openapi_test.go")
	}
	content, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var contract map[string]any
	if err := yaml.Unmarshal(content, &contract); err != nil {
		t.Fatalf("parse openapi.yaml: %v", err)
	}
	return contract
}

func verifyReferences(t *testing.T, root map[string]any, value any, location string) {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		if ref, ok := typed["$ref"].(string); ok {
			if !strings.HasPrefix(ref, "#/") {
				t.Errorf("%s has unsupported reference %q", location, ref)
			} else if _, err := resolveReference(root, ref); err != nil {
				t.Errorf("%s: %v", location, err)
			}
		}
		for key, child := range typed {
			verifyReferences(t, root, child, location+"/"+key)
		}
	case []any:
		for i, child := range typed {
			verifyReferences(t, root, child, location+"/"+strconv.Itoa(i))
		}
	}
}

func verifyOperationExamples(t *testing.T, root map[string]any) {
	t.Helper()
	seen := map[string]bool{}
	paths := mustMap(t, root["paths"], "paths")
	for path, rawPathItem := range paths {
		operation := mustMap(t, mustMap(t, rawPathItem, "paths."+path)["post"], "paths."+path+".post")
		verifyContentExamples(t, root, operation["requestBody"], "request "+path, seen)
		for status, response := range mustMap(t, operation["responses"], path+".responses") {
			verifyContentExamples(t, root, response, "response "+path+" "+status, seen)
		}
	}
	for _, name := range []string{
		"csvConversion", "jsonHTMLConversion", "convertedDataset", "convertedHTML", "taggedDatasets",
		"mergedDatasets", "datasetUI", "selfContainedHTML", "unknownOption", "invalidCSV",
	} {
		if !seen[name] {
			t.Errorf("required contract example %q is missing", name)
		}
	}
}

func verifyContentExamples(t *testing.T, root map[string]any, raw any, location string, seen map[string]bool) {
	t.Helper()
	if raw == nil {
		return
	}
	item, err := dereference(root, raw)
	if err != nil {
		t.Errorf("%s: %v", location, err)
		return
	}
	content, ok := item.(map[string]any)["content"]
	if !ok {
		return
	}
	for mediaType, rawMedia := range mustMap(t, content, location+".content") {
		media := mustMap(t, rawMedia, location+".content."+mediaType)
		schema := media["schema"]
		for name, rawExample := range mustMapOrEmpty(t, media["examples"], location+".examples") {
			example := mustMap(t, rawExample, location+".examples."+name)
			if err := validateSchema(root, schema, example["value"], "example "+name); err != nil {
				t.Errorf("%s (%s): %v", location, mediaType, err)
			}
			seen[name] = true
		}
	}
}

func validateSchema(root map[string]any, rawSchema, value any, location string) error {
	dereferenced, err := dereference(root, rawSchema)
	if err != nil {
		return err
	}
	schema := mustMapValue(dereferenced, location+" schema")
	if constant, exists := schema["const"]; exists && !reflect.DeepEqual(constant, value) {
		return fmt.Errorf("%s = %#v, want constant %#v", location, value, constant)
	}
	if enum, exists := schema["enum"]; exists {
		matched := false
		for _, candidate := range mustSliceValue(enum, location+" enum") {
			matched = matched || reflect.DeepEqual(candidate, value)
		}
		if !matched {
			return fmt.Errorf("%s = %#v is not in enum", location, value)
		}
	}
	if allOf, exists := schema["allOf"]; exists {
		for _, child := range mustSliceValue(allOf, location+" allOf") {
			if err := validateSchema(root, child, value, location); err != nil {
				return err
			}
		}
	}
	if oneOf, exists := schema["oneOf"]; exists {
		matches := 0
		for _, child := range mustSliceValue(oneOf, location+" oneOf") {
			if validateSchema(root, child, value, location) == nil {
				matches++
			}
		}
		if matches != 1 {
			return fmt.Errorf("%s matches %d oneOf schemas, want 1", location, matches)
		}
	}
	if contains, exists := schema["contains"]; exists {
		array, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s must be an array for contains, got %T", location, value)
		}
		matches := 0
		for i, childValue := range array {
			if validateSchema(root, contains, childValue, location+"/"+strconv.Itoa(i)) == nil {
				matches++
			}
		}
		minMatches := 1
		if min, ok := schema["minContains"].(int); ok {
			minMatches = min
		}
		if matches < minMatches {
			return fmt.Errorf("%s matches %d items, want at least %d", location, matches, minMatches)
		}
		if maxMatches, ok := schema["maxContains"].(int); ok && matches > maxMatches {
			return fmt.Errorf("%s matches %d items, want at most %d", location, matches, maxMatches)
		}
	}

	typeName, _ := schema["type"].(string)
	switch typeName {
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s must be an object, got %T", location, value)
		}
		properties := mustMapOrEmptyValue(schema["properties"], location+" properties")
		for _, key := range stringSliceValue(schema["required"]) {
			if _, exists := object[key]; !exists {
				return fmt.Errorf("%s is missing required property %q", location, key)
			}
		}
		if strict, declared := schema["additionalProperties"].(bool); declared && !strict {
			for key := range object {
				if _, known := properties[key]; !known {
					return fmt.Errorf("%s has unknown property %q", location, key)
				}
			}
		}
		for key, child := range properties {
			if childValue, exists := object[key]; exists {
				if err := validateSchema(root, child, childValue, location+"/"+key); err != nil {
					return err
				}
			}
		}
	case "array":
		array, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s must be an array, got %T", location, value)
		}
		if min, ok := schema["minItems"].(int); ok && len(array) < min {
			return fmt.Errorf("%s has %d items, want at least %d", location, len(array), min)
		}
		if child, exists := schema["items"]; exists {
			for i, childValue := range array {
				if err := validateSchema(root, child, childValue, location+"/"+strconv.Itoa(i)); err != nil {
					return err
				}
			}
		}
	case "string":
		stringValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s must be a string, got %T", location, value)
		}
		if min, ok := schema["minLength"].(int); ok && len(stringValue) < min {
			return fmt.Errorf("%s has length %d, want at least %d", location, len(stringValue), min)
		}
		if pattern, ok := schema["pattern"].(string); ok {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("%s has invalid pattern %q: %w", location, pattern, err)
			}
			if !re.MatchString(stringValue) {
				return fmt.Errorf("%s = %#v does not match pattern %q", location, stringValue, pattern)
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s must be a boolean, got %T", location, value)
		}
	case "integer":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		default:
			return fmt.Errorf("%s must be an integer, got %T", location, value)
		}
	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		default:
			return fmt.Errorf("%s must be a number, got %T", location, value)
		}
	}
	return nil
}

func dereference(root map[string]any, raw any) (any, error) {
	current := raw
	for {
		object, ok := current.(map[string]any)
		if !ok {
			return current, nil
		}
		ref, ok := object["$ref"].(string)
		if !ok {
			return current, nil
		}
		resolved, err := resolveReference(root, ref)
		if err != nil {
			return nil, err
		}
		current = resolved
	}
}

func resolveReference(root map[string]any, ref string) (any, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("unsupported reference %q", ref)
	}
	var current any = root
	for _, token := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		key := strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		object, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("reference %q traverses non-object at %q", ref, key)
		}
		var exists bool
		current, exists = object[key]
		if !exists {
			return nil, fmt.Errorf("reference %q does not exist", ref)
		}
	}
	return current, nil
}

func propertyNames(t *testing.T, schema map[string]any, location string) []string {
	t.Helper()
	properties := mustMap(t, schema["properties"], location+".properties")
	result := make([]string, 0, len(properties))
	for name := range properties {
		result = append(result, name)
	}
	return sorted(result)
}

func jsonFieldNames(value any) []string {
	typ := reflect.TypeOf(value)
	result := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name != "" && name != "-" {
			result = append(result, name)
		}
	}
	return sorted(result)
}

func mustMap(t *testing.T, value any, location string) map[string]any {
	t.Helper()
	return mustMapValue(value, location)
}

func mustMapValue(value any, location string) map[string]any {
	object, ok := value.(map[string]any)
	if !ok {
		panic(fmt.Sprintf("%s must be an object, got %T", location, value))
	}
	return object
}

func mustMapOrEmpty(t *testing.T, value any, location string) map[string]any {
	t.Helper()
	if value == nil {
		return map[string]any{}
	}
	return mustMap(t, value, location)
}

func mustMapOrEmptyValue(value any, location string) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return mustMapValue(value, location)
}

func mustSliceValue(value any, location string) []any {
	items, ok := value.([]any)
	if !ok {
		panic(fmt.Sprintf("%s must be an array, got %T", location, value))
	}
	return items
}

func stringSliceValue(value any) []string {
	if value == nil {
		return nil
	}
	items := mustSliceValue(value, "string array")
	result := make([]string, 0, len(items))
	for _, value := range items {
		if stringValue, ok := value.(string); ok {
			result = append(result, stringValue)
		}
	}
	return sorted(result)
}

func sorted(values []string) []string {
	result := append([]string(nil), values...)
	for i := range result {
		for j := i + 1; j < len(result); j++ {
			if result[j] < result[i] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

func sortedMapKeys(values map[string]any) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return sorted(result)
}
