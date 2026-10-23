// Command openapi-gen generates godo-style package v2 service code from a
// bundled DigitalOcean OpenAPI document (full spec).
//
// Supported common CRUD shapes (exceptions are skipped with a reason):
//   - GET  + array envelope     → List(ctx, [pathParams...], *ListOptions) ([]T, *Response, error)
//   - GET  + object envelope    → Get(ctx, [pathParams...]) (*T, *Response, error)
//   - POST + *_create           → Create(ctx, [pathParams...], *Req) (*T, *Response, error)
//   - PUT/PATCH + *_update      → Update(ctx, [pathParams...], *Req) (*T, *Response, error)
//   - DELETE (*_delete|_destroy)→ Delete(ctx, [pathParams...]) (*Response, error)
//
// oneOf/anyOf request bodies: one method per $ref variant (e.g. Create + CreateMultiple).
// oneOf/anyOf responses are paired by index when possible.
//
// Skipped examples: non-CRUD POSTs (actions), non-JSON bodies, odd envelopes.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"gopkg.in/yaml.v3"
)

type config struct {
	Package   string            `yaml:"package"`
	OutputDir string            `yaml:"output_dir"`
	Spec      string            `yaml:"spec"`
	Specs     []string          `yaml:"specs"`
	ModelMap  map[string]string `yaml:"model_map"`
	Tags      []string          `yaml:"tags"`
}

type openAPI struct {
	Paths      map[string]pathItem `yaml:"paths"`
	Components *components         `yaml:"components"`
}

type components struct {
	Schemas    map[string]*schema    `yaml:"schemas"`
	Responses  map[string]*response  `yaml:"responses"`
	Parameters map[string]*parameter `yaml:"parameters"`
}

type pathItem struct {
	Get    *operation `yaml:"get"`
	Post   *operation `yaml:"post"`
	Put    *operation `yaml:"put"`
	Patch  *operation `yaml:"patch"`
	Delete *operation `yaml:"delete"`
}

type operation struct {
	OperationID string              `yaml:"operationId"`
	Tags        []string            `yaml:"tags"`
	Summary     string              `yaml:"summary"`
	Description string              `yaml:"description"`
	Parameters  []parameter         `yaml:"parameters"`
	RequestBody *requestBody        `yaml:"requestBody"`
	Responses   map[string]response `yaml:"responses"`
}

type parameter struct {
	Ref      string  `yaml:"$ref"`
	Name     string  `yaml:"name"`
	In       string  `yaml:"in"`
	Required bool    `yaml:"required"`
	Schema   *schema `yaml:"schema"`
}

type requestBody struct {
	Required bool                 `yaml:"required"`
	Content  map[string]mediaType `yaml:"content"`
}

type response struct {
	Ref     string               `yaml:"$ref"`
	Content map[string]mediaType `yaml:"content"`
}

type mediaType struct {
	Schema schema `yaml:"schema"`
}

type schema struct {
	Type                 string             `yaml:"type"`
	Format               string             `yaml:"format"`
	Description          string             `yaml:"description"`
	Required             []string           `yaml:"required"`
	Properties           map[string]*schema `yaml:"properties"`
	Items                *schema            `yaml:"items"`
	Ref                  string             `yaml:"$ref"`
	AllOf                []*schema          `yaml:"allOf"`
	OneOf                []*schema          `yaml:"oneOf"`
	AnyOf                []*schema          `yaml:"anyOf"`
	AdditionalProperties any                `yaml:"additionalProperties"`
}

type pathParam struct {
	Name   string // openapi name, e.g. droplet_id
	GoName string // Go arg name, e.g. dropletID
	GoType string // int or string
	Verb   string // %d or %s for fmt.Sprintf
}

type serviceMethod struct {
	Service     string
	MethodName  string
	HTTPMethod  string
	Path        string
	Summary     string
	OperationID string

	IsList   bool
	IsGet    bool
	IsCreate bool
	IsUpdate bool
	IsDelete bool

	ItemType  string
	RootKey   string
	RootType  string
	SchemaRef string // response model schema name

	PathParams  []pathParam
	RequestType string // Go type for body (* omitted in template)
	RequestRef  string // schema name for body model
	HasBody     bool
	HasRoot     bool // response has JSON envelope to decode
	NeedsFmt    bool
	ReturnsList bool // create/update that returns []T (e.g. multi-create)

	// Precomputed snippets for templates.
	FuncParams string // excluding ctx, e.g. `dropletID int, req *Foo`
	PathExpr   string // Go expr yielding path string
	FmtArgs    string // ", dropletID" for fmt.Sprintf
}

type serviceDef struct {
	Name     string
	FileBase string
	Methods  []serviceMethod
}

type modelField struct {
	Name string
	Type string
	Tag  string
}

type modelDef struct {
	Name   string
	Fields []modelField
}

func main() {
	cfgPath := flag.String("config", "v2/gen/godo-gen.yaml", "path to godo-gen config")
	repoRoot := flag.String("root", ".", "repository root")
	flag.Parse()

	if err := run(*repoRoot, *cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "openapi-gen: %v\n", err)
		os.Exit(1)
	}
}

func run(root, cfgPath string) error {
	cfgBytes, err := os.ReadFile(filepath.Join(root, cfgPath))
	if err != nil {
		return err
	}
	var cfg config
	if err := yaml.Unmarshal(cfgBytes, &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	if cfg.Package == "" {
		cfg.Package = "v2"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "v2"
	}
	specs := cfg.Specs
	if len(specs) == 0 && cfg.Spec != "" {
		specs = []string{cfg.Spec}
	}
	if len(specs) == 0 {
		return fmt.Errorf("config.spec (or specs) must be set")
	}

	wantTags := map[string]bool{}
	for _, t := range cfg.Tags {
		wantTags[t] = true
	}
	allTags := len(wantTags) == 0

	var methods []serviceMethod
	var skipped []string
	neededSchemas := map[string]bool{}
	var doc openAPI

	for _, specRel := range specs {
		specBytes, err := os.ReadFile(filepath.Join(root, specRel))
		if err != nil {
			return err
		}
		doc = openAPI{}
		if err := yaml.Unmarshal(specBytes, &doc); err != nil {
			return fmt.Errorf("parse spec %s: %w", specRel, err)
		}
		if doc.Components == nil {
			doc.Components = &components{}
		}

		paths := make([]string, 0, len(doc.Paths))
		for p := range doc.Paths {
			paths = append(paths, p)
		}
		sort.Strings(paths)

		for _, p := range paths {
			item := doc.Paths[p]
			for _, op := range []struct {
				m string
				o *operation
			}{
				{"GET", item.Get},
				{"POST", item.Post},
				{"PUT", item.Put},
				{"PATCH", item.Patch},
				{"DELETE", item.Delete},
			} {
				if op.o == nil {
					continue
				}
				tag := pickTag(op.o.Tags, wantTags, allTags)
				if tag == "" {
					continue
				}
				built, skipReason, err := buildMethods(tag, op.m, strings.TrimPrefix(p, "/"), op.o, &doc, cfg.ModelMap)
				if err != nil {
					return fmt.Errorf("%s %s: %w", op.m, p, err)
				}
				if skipReason != "" {
					skipped = append(skipped, fmt.Sprintf("%s %s (%s): %s", op.m, p, op.o.OperationID, skipReason))
					continue
				}
				for _, sm := range built {
					methods = append(methods, sm)
					if sm.SchemaRef != "" {
						collectSchemaDeps(sm.SchemaRef, doc.Components.Schemas, neededSchemas)
					}
					if sm.RequestRef != "" {
						collectSchemaDeps(sm.RequestRef, doc.Components.Schemas, neededSchemas)
					}
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "openapi-gen: %d operations generated, %d skipped\n", len(methods), len(skipped))
	if len(methods) == 0 {
		return fmt.Errorf("no supported operations found")
	}

	uniquifyMethodNames(methods)
	services := groupServices(methods)
	outDir := filepath.Join(root, cfg.OutputDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	if err := clearGenerated(outDir); err != nil {
		return err
	}

	models := buildModels(doc.Components.Schemas, neededSchemas, cfg.ModelMap)
	if err := writeModels(outDir, cfg.Package, models); err != nil {
		return err
	}
	specLabel := strings.Join(specs, ", ")
	if err := writeServices(outDir, cfg.Package, specLabel, services); err != nil {
		return err
	}
	return writeClient(outDir, cfg.Package, specLabel, services)
}

func clearGenerated(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".gen.go") {
			if err := os.Remove(filepath.Join(outDir, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func pickTag(tags []string, want map[string]bool, all bool) string {
	for _, t := range tags {
		if t == "" {
			continue
		}
		if strings.Contains(strings.ToLower(t), "introduction") {
			continue
		}
		if all || want[t] {
			return t
		}
	}
	return ""
}

func buildMethods(tag, httpMethod, path string, op *operation, doc *openAPI, modelMap map[string]string) ([]serviceMethod, string, error) {
	base := serviceMethod{
		Service:     goIdent(tag),
		HTTPMethod:  httpMethod,
		Path:        path,
		Summary:     oneLine(op.Summary),
		OperationID: op.OperationID,
	}

	params, err := extractPathParams(path, op, doc)
	if err != nil {
		return nil, err.Error(), nil
	}
	base.PathParams = params
	base.NeedsFmt = len(params) > 0
	base.PathExpr, base.FmtArgs = pathExpression(path, params)

	kind, skip := classifyOperation(httpMethod, op.OperationID)
	if skip != "" {
		return nil, skip, nil
	}

	baseName := crudMethodName(op.OperationID, tag, kind)

	switch kind {
	case "list":
		sm := base
		sm.IsList = true
		sm.MethodName = baseName
		rootKey, schemaName, err := listResponseShape(op, doc)
		if err != nil {
			return nil, err.Error(), nil
		}
		sm.RootKey = rootKey
		sm.SchemaRef = schemaName
		sm.ItemType = mapModel(modelMap, schemaName)
		sm.RootType = goIdent(op.OperationID) + "Root"
		sm.HasRoot = true
		sm.FuncParams = buildFuncParams(sm)
		return []serviceMethod{sm}, "", nil

	case "get":
		sm := base
		sm.MethodName = baseName
		rootKey, schemaName, err := objectResponseShape(op, doc)
		if err != nil {
			rootKey, schemaName, err = listResponseShape(op, doc)
			if err != nil {
				return nil, err.Error(), nil
			}
			sm.IsList = true
		} else {
			sm.IsGet = true
		}
		sm.RootKey = rootKey
		sm.SchemaRef = schemaName
		sm.ItemType = mapModel(modelMap, schemaName)
		sm.RootType = goIdent(op.OperationID) + "Root"
		sm.HasRoot = true
		sm.FuncParams = buildFuncParams(sm)
		return []serviceMethod{sm}, "", nil

	case "create", "update":
		reqVariants, reqSkip := requestBodyVariants(op, doc)
		if reqSkip != "" {
			return nil, reqSkip, nil
		}
		respVariants, respErr := responseShapeVariants(op, doc)
		if respErr != nil && len(reqVariants) <= 1 {
			return nil, respErr.Error(), nil
		}

		out := make([]serviceMethod, 0, len(reqVariants))
		for i, reqName := range reqVariants {
			sm := base
			if kind == "create" {
				sm.IsCreate = true
			} else {
				sm.IsUpdate = true
			}
			sm.MethodName = variantMethodName(baseName, reqName, i, len(reqVariants))
			if reqName != "" {
				sm.HasBody = true
				sm.RequestRef = reqName
				sm.RequestType = mapModel(modelMap, reqName)
			}

			var shape respShape
			var ok bool
			switch {
			case len(respVariants) == len(reqVariants):
				shape, ok = respVariants[i], true
			case len(respVariants) == 1:
				shape, ok = respVariants[0], true
			case len(respVariants) > 0 && i < len(respVariants):
				shape, ok = respVariants[i], true
			}
			if !ok {
				return nil, fmt.Sprintf("no response shape for request variant %q", reqName), nil
			}
			sm.RootKey = shape.RootKey
			sm.SchemaRef = shape.SchemaName
			sm.ItemType = mapModel(modelMap, shape.SchemaName)
			sm.ReturnsList = shape.IsList
			sm.RootType = goIdent(op.OperationID) + sm.MethodName + "Root"
			sm.HasRoot = true
			sm.FuncParams = buildFuncParams(sm)
			out = append(out, sm)
		}
		if len(out) == 0 {
			return nil, "no request body variants", nil
		}
		return out, "", nil

	case "delete":
		sm := base
		sm.IsDelete = true
		sm.MethodName = baseName
		sm.HasRoot = false
		sm.FuncParams = buildFuncParams(sm)
		return []serviceMethod{sm}, "", nil

	default:
		return nil, "unsupported operation kind", nil
	}
}

type respShape struct {
	RootKey    string
	SchemaName string
	IsList     bool
}

// variantMethodName turns oneOf branches into Create / CreateMultiple / CreateFoo.
func variantMethodName(base, schemaName string, index, total int) string {
	if total <= 1 {
		return base
	}
	lower := strings.ToLower(schemaName)
	switch {
	case strings.Contains(lower, "multi"):
		if base == "Create" {
			return "CreateMultiple"
		}
		if base == "Update" {
			return "UpdateMultiple"
		}
		return base + "Multiple"
	case index == 0 && (strings.Contains(lower, "single") || strings.Contains(lower, "one")):
		return base
	case index == 0:
		return base
	default:
		suffix := goIdent(schemaName)
		// Avoid CreateDropletSingleCreate — strip trailing Create/Update/Request.
		for _, trim := range []string{"Create", "Update", "Request", "Input", "Body"} {
			if strings.HasSuffix(suffix, trim) && suffix != trim {
				suffix = strings.TrimSuffix(suffix, trim)
			}
		}
		if suffix == "" || suffix == base {
			return fmt.Sprintf("%s%d", base, index+1)
		}
		return base + suffix
	}
}

func classifyOperation(httpMethod, operationID string) (kind, skip string) {
	id := strings.ToLower(operationID)
	switch strings.ToUpper(httpMethod) {
	case "GET":
		if strings.Contains(id, "_list") {
			return "list", ""
		}
		if strings.Contains(id, "_get") {
			return "get", ""
		}
		// Fallback: allow plain GETs to be classified by caller via response shape attempts.
		// Prefer get; listResponseShape will fail if not an array.
		return "get", ""
	case "POST":
		if strings.Contains(id, "_create") {
			return "create", ""
		}
		return "", "non-CRUD POST (not *_create) — skipped as exception"
	case "PUT", "PATCH":
		if strings.Contains(id, "_update") {
			return "update", ""
		}
		return "", "non-CRUD PUT/PATCH (not *_update) — skipped as exception"
	case "DELETE":
		if strings.Contains(id, "_delete") || strings.Contains(id, "_destroy") {
			return "delete", ""
		}
		return "", "non-CRUD DELETE (not *_delete/*_destroy) — skipped as exception"
	default:
		return "", "unsupported HTTP method"
	}
}

func crudMethodName(operationID, tag, kind string) string {
	name := methodNameFromOperation(operationID, tag)
	// Map OpenAPI "destroy" naming to godo-style Delete.
	if name == "Destroy" {
		return "Delete"
	}
	if strings.HasPrefix(name, "Destroy") {
		return "Delete" + strings.TrimPrefix(name, "Destroy")
	}
	// Normalize bare verb names when operationId is just resource_create etc.
	switch kind {
	case "list":
		if name == "" || name == goIdent(tag) {
			return "List"
		}
	case "get":
		if name == "" || name == goIdent(tag) {
			return "Get"
		}
	case "create":
		if name == "" || name == goIdent(tag) {
			return "Create"
		}
	case "update":
		if name == "" || name == goIdent(tag) {
			return "Update"
		}
	case "delete":
		if name == "" || name == goIdent(tag) || name == "Delete" {
			return "Delete"
		}
	}
	return name
}

func buildFuncParams(sm serviceMethod) string {
	var parts []string
	for _, p := range sm.PathParams {
		parts = append(parts, p.GoName+" "+p.GoType)
	}
	if sm.HasBody {
		parts = append(parts, "req *"+sm.RequestType)
	}
	if sm.IsList {
		parts = append(parts, "opt *ListOptions")
	}
	return strings.Join(parts, ", ")
}

func pathExpression(path string, params []pathParam) (expr, fmtArgs string) {
	if len(params) == 0 {
		return strconvQuote(path), ""
	}
	format := path
	var args []string
	for _, p := range params {
		token := "{" + p.Name + "}"
		format = strings.ReplaceAll(format, token, p.Verb)
		args = append(args, p.GoName)
	}
	return strconvQuote(format), ", " + strings.Join(args, ", ")
}

func strconvQuote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

var pathParamRe = regexp.MustCompile(`\{([^}]+)\}`)

func extractPathParams(path string, op *operation, doc *openAPI) ([]pathParam, error) {
	matches := pathParamRe.FindAllStringSubmatch(path, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	byName := map[string]*parameter{}
	for i := range op.Parameters {
		p, err := resolveParameter(&op.Parameters[i], doc)
		if err != nil {
			return nil, err
		}
		if p != nil && p.In == "path" && p.Name != "" {
			byName[p.Name] = p
		}
	}

	out := make([]pathParam, 0, len(matches))
	for _, m := range matches {
		name := m[1]
		pp := pathParam{
			Name:   name,
			GoName: goArgName(name),
			GoType: "string",
			Verb:   "%s",
		}
		if p, ok := byName[name]; ok && p.Schema != nil {
			sch := resolveSchema(p.Schema, doc)
			if sch != nil && sch.Type == "integer" {
				pp.GoType = "int"
				pp.Verb = "%d"
			}
		}
		out = append(out, pp)
	}
	return out, nil
}

func resolveParameter(p *parameter, doc *openAPI) (*parameter, error) {
	if p == nil {
		return nil, nil
	}
	if p.Ref == "" {
		return p, nil
	}
	name := strings.TrimPrefix(p.Ref, "#/components/parameters/")
	// Also accept bare names after bundle.
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}
	if doc.Components == nil || doc.Components.Parameters == nil {
		return nil, fmt.Errorf("missing components.parameters for %s", p.Ref)
	}
	ref, ok := doc.Components.Parameters[name]
	if !ok || ref == nil {
		// Redocly often inlines parameters; if unresolved, fall back to path token as string.
		return p, nil
	}
	return ref, nil
}

func goArgName(s string) string {
	ident := goIdent(s)
	if ident == "" {
		return "id"
	}
	// lowerCamelCase
	r := []rune(ident)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func mapModel(modelMap map[string]string, schemaName string) string {
	if mapped, ok := modelMap[schemaName]; ok {
		return mapped
	}
	return goIdent(schemaName)
}

// oneLine collapses whitespace so OpenAPI summaries are safe in Go // comments.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.Join(strings.Fields(s), " ")
}

func resolveResponse(resp response, doc *openAPI) (response, error) {
	if resp.Ref == "" {
		return resp, nil
	}
	name := strings.TrimPrefix(resp.Ref, "#/components/responses/")
	if doc.Components == nil || doc.Components.Responses == nil {
		return resp, fmt.Errorf("missing components.responses for %s", resp.Ref)
	}
	ref, ok := doc.Components.Responses[name]
	if !ok || ref == nil {
		return resp, fmt.Errorf("unknown response ref %s", resp.Ref)
	}
	return *ref, nil
}

func resolveSchema(s *schema, doc *openAPI) *schema {
	if s == nil {
		return nil
	}
	if s.Ref != "" {
		name := strings.TrimPrefix(s.Ref, "#/components/schemas/")
		if doc.Components != nil && doc.Components.Schemas != nil {
			if ref, ok := doc.Components.Schemas[name]; ok && ref != nil {
				return resolveSchema(ref, doc)
			}
		}
		return s
	}
	if len(s.AllOf) > 0 {
		merged := &schema{Properties: map[string]*schema{}}
		for _, part := range s.AllOf {
			p := resolveSchema(part, doc)
			if p == nil {
				continue
			}
			if p.Type != "" {
				merged.Type = p.Type
			}
			for k, v := range p.Properties {
				merged.Properties[k] = v
			}
			if p.Ref != "" && merged.Ref == "" {
				merged.Ref = p.Ref
			}
		}
		return merged
	}
	return s
}

func successResponse(op *operation, doc *openAPI) (response, error) {
	for _, code := range []string{"200", "201", "202"} {
		raw, ok := op.Responses[code]
		if !ok {
			continue
		}
		resp, err := resolveResponse(raw, doc)
		if err != nil {
			return response{}, err
		}
		return resp, nil
	}
	return response{}, fmt.Errorf("missing 200/201/202 response")
}

func responseProperties(op *operation, doc *openAPI) (map[string]*schema, error) {
	resp, err := successResponse(op, doc)
	if err != nil {
		return nil, err
	}
	mt, ok := resp.Content["application/json"]
	if !ok {
		return nil, fmt.Errorf("missing application/json content")
	}
	sch := resolveSchema(&mt.Schema, doc)
	if sch == nil {
		return nil, fmt.Errorf("empty success schema")
	}
	if len(sch.Properties) == 0 && sch.Ref != "" {
		sch = resolveSchema(sch, doc)
	}
	return sch.Properties, nil
}

func listResponseShape(op *operation, doc *openAPI) (rootKey, schemaName string, err error) {
	props, err := responseProperties(op, doc)
	if err != nil {
		return "", "", err
	}
	for key, prop := range props {
		p := resolveSchema(prop, doc)
		if p == nil {
			continue
		}
		if p.Type == "array" && p.Items != nil {
			item := p.Items
			if item.Ref != "" {
				return key, strings.TrimPrefix(item.Ref, "#/components/schemas/"), nil
			}
			return "", "", fmt.Errorf("inline array items not supported")
		}
	}
	return "", "", fmt.Errorf("no array property in success schema")
}

func objectResponseShape(op *operation, doc *openAPI) (rootKey, schemaName string, err error) {
	props, err := responseProperties(op, doc)
	if err != nil {
		return "", "", err
	}
	for key, prop := range props {
		if key == "links" || key == "meta" {
			continue
		}
		if prop != nil && prop.Ref != "" {
			return key, strings.TrimPrefix(prop.Ref, "#/components/schemas/"), nil
		}
		p := resolveSchema(prop, doc)
		if p != nil && p.Ref != "" {
			return key, strings.TrimPrefix(p.Ref, "#/components/schemas/"), nil
		}
	}
	return "", "", fmt.Errorf("no object $ref property in success schema")
}

// requestBodyVariants returns one schema name per request body shape.
// oneOf/anyOf with $ref members → one entry each; plain $ref/object → single entry.
func requestBodyVariants(op *operation, doc *openAPI) (names []string, skip string) {
	if op.RequestBody == nil {
		return []string{""}, ""
	}
	mt, ok := op.RequestBody.Content["application/json"]
	if !ok {
		return nil, "request body is not application/json"
	}
	sch := &mt.Schema
	branches := sch.OneOf
	if len(branches) == 0 {
		branches = sch.AnyOf
	}
	if len(branches) > 0 {
		for i, b := range branches {
			name, err := schemaRefOrRegister(b, doc, fmt.Sprintf("%sRequest%d", goIdent(op.OperationID), i+1))
			if err != "" {
				return nil, err
			}
			names = append(names, name)
		}
		if len(names) == 0 {
			return nil, "oneOf/anyOf request body has no usable $ref variants"
		}
		return names, ""
	}
	name, err := schemaRefOrRegister(sch, doc, goIdent(op.OperationID)+"Request")
	if err != "" {
		return nil, err
	}
	return []string{name}, ""
}

func schemaRefOrRegister(sch *schema, doc *openAPI, inlineName string) (name, skip string) {
	if sch == nil {
		return "", "empty schema"
	}
	if sch.Ref != "" {
		return strings.TrimPrefix(sch.Ref, "#/components/schemas/"), ""
	}
	resolved := resolveSchema(sch, doc)
	if resolved != nil && resolved.Ref != "" {
		return strings.TrimPrefix(resolved.Ref, "#/components/schemas/"), ""
	}
	if resolved != nil && (resolved.Type == "object" || len(resolved.Properties) > 0) {
		if doc.Components.Schemas == nil {
			doc.Components.Schemas = map[string]*schema{}
		}
		doc.Components.Schemas[inlineName] = resolved
		return inlineName, ""
	}
	return "", "unsupported request body schema"
}

// responseShapeVariants returns one response shape per oneOf/anyOf branch, or a single shape.
func responseShapeVariants(op *operation, doc *openAPI) ([]respShape, error) {
	resp, err := successResponse(op, doc)
	if err != nil {
		return nil, err
	}
	mt, ok := resp.Content["application/json"]
	if !ok {
		return nil, fmt.Errorf("missing application/json content")
	}
	sch := &mt.Schema
	branches := sch.OneOf
	if len(branches) == 0 {
		branches = sch.AnyOf
	}
	if len(branches) > 0 {
		var out []respShape
		for _, b := range branches {
			shape, err := shapeFromProperties(schemaProperties(b, doc))
			if err != nil {
				return nil, err
			}
			out = append(out, shape)
		}
		return out, nil
	}
	// Single envelope (or allOf-merged).
	props, err := responseProperties(op, doc)
	if err != nil {
		return nil, err
	}
	shape, err := shapeFromProperties(props)
	if err != nil {
		return nil, err
	}
	return []respShape{shape}, nil
}

func schemaProperties(s *schema, doc *openAPI) map[string]*schema {
	if s == nil {
		return nil
	}
	resolved := resolveSchema(s, doc)
	if resolved == nil {
		return nil
	}
	if len(resolved.Properties) > 0 {
		return resolved.Properties
	}
	return nil
}

func shapeFromProperties(props map[string]*schema) (respShape, error) {
	if props == nil {
		return respShape{}, fmt.Errorf("empty response properties")
	}
	// Prefer array envelopes (multi-create) then object $ref.
	for key, prop := range props {
		if key == "links" || key == "meta" {
			continue
		}
		if prop == nil {
			continue
		}
		if prop.Type == "array" && prop.Items != nil && prop.Items.Ref != "" {
			return respShape{
				RootKey:    key,
				SchemaName: strings.TrimPrefix(prop.Items.Ref, "#/components/schemas/"),
				IsList:     true,
			}, nil
		}
	}
	for key, prop := range props {
		if key == "links" || key == "meta" {
			continue
		}
		if prop != nil && prop.Ref != "" {
			return respShape{
				RootKey:    key,
				SchemaName: strings.TrimPrefix(prop.Ref, "#/components/schemas/"),
				IsList:     false,
			}, nil
		}
	}
	return respShape{}, fmt.Errorf("no object/array $ref property in success schema")
}

func methodNameFromOperation(operationID, tag string) string {
	name := operationID
	prefix := toSnake(tag) + "_"
	if strings.HasPrefix(strings.ToLower(name), prefix) {
		// Case-insensitive strip of tag_snake_ prefix.
		name = name[len(prefix):]
	} else if i := strings.Index(name, "_"); i > 0 {
		// sshKeys_list / vpcPeerings_create → drop resource prefix before first "_".
		rest := name[i+1:]
		lower := strings.ToLower(rest)
		for _, v := range []string{"list", "get", "create", "update", "delete", "destroy"} {
			if lower == v || strings.HasPrefix(lower, v+"_") {
				name = rest
				break
			}
		}
	}
	return goIdent(name)
}

func toSnake(s string) string {
	s = strings.TrimSpace(s)
	s = nonIdent.ReplaceAllString(s, "_")
	return strings.ToLower(strings.Trim(s, "_"))
}

func collectSchemaDeps(name string, schemas map[string]*schema, out map[string]bool) {
	if name == "" || out[name] {
		return
	}
	out[name] = true
	sch, ok := schemas[name]
	if !ok || sch == nil {
		return
	}
	walkSchemaRefs(sch, schemas, out)
}

func walkSchemaRefs(s *schema, schemas map[string]*schema, out map[string]bool) {
	if s == nil {
		return
	}
	if s.Ref != "" {
		collectSchemaDeps(strings.TrimPrefix(s.Ref, "#/components/schemas/"), schemas, out)
		return
	}
	for _, part := range s.AllOf {
		walkSchemaRefs(part, schemas, out)
	}
	for _, part := range s.OneOf {
		walkSchemaRefs(part, schemas, out)
	}
	for _, part := range s.AnyOf {
		walkSchemaRefs(part, schemas, out)
	}
	for _, p := range s.Properties {
		walkSchemaRefs(p, schemas, out)
	}
	if s.Items != nil {
		walkSchemaRefs(s.Items, schemas, out)
	}
}

func buildModels(schemas map[string]*schema, needed map[string]bool, modelMap map[string]string) []modelDef {
	names := make([]string, 0, len(needed))
	for n := range needed {
		names = append(names, n)
	}
	sort.Strings(names)

	var out []modelDef
	for _, name := range names {
		goName := mapModel(modelMap, name)
		sch, ok := schemas[name]
		if !ok || sch == nil {
			out = append(out, modelDef{Name: goName})
			continue
		}
		sch = resolveSchema(sch, &openAPI{Components: &components{Schemas: schemas}})
		fields := make([]modelField, 0, len(sch.Properties))
		propNames := make([]string, 0, len(sch.Properties))
		for p := range sch.Properties {
			propNames = append(propNames, p)
		}
		sort.Strings(propNames)
		for _, p := range propNames {
			fields = append(fields, modelField{
				Name: goIdent(p),
				Type: goType(sch.Properties[p], schemas),
				Tag:  p,
			})
		}
		out = append(out, modelDef{Name: goName, Fields: fields})
	}
	return out
}

func goType(s *schema, schemas map[string]*schema) string {
	if s == nil {
		return "json.RawMessage"
	}
	if s.Ref != "" {
		return "*" + goIdent(strings.TrimPrefix(s.Ref, "#/components/schemas/"))
	}
	switch s.Type {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		if s.Items == nil {
			return "[]any"
		}
		if s.Items.Ref != "" {
			return "[]" + goIdent(strings.TrimPrefix(s.Items.Ref, "#/components/schemas/"))
		}
		inner := goType(s.Items, schemas)
		if strings.HasPrefix(inner, "*") {
			return "[]" + strings.TrimPrefix(inner, "*")
		}
		return "[]" + inner
	case "object":
		return "json.RawMessage"
	default:
		if len(s.AllOf) > 0 || len(s.Properties) > 0 {
			return "json.RawMessage"
		}
		return "json.RawMessage"
	}
}

var nonIdent = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func goIdent(s string) string {
	s = strings.TrimSpace(s)
	s = nonIdent.ReplaceAllString(s, "_")
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if unicode.IsDigit(rune(p[0])) {
			p = "N" + p
		}
		b.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			b.WriteString(p[1:])
		}
	}
	out := b.String()
	if out == "" {
		return "X"
	}
	return out
}

// uniquifyMethodNames ensures each service has unique Go method names.
// Collisions (e.g. two ops both named Get) are expanded from the full operationId.
func uniquifyMethodNames(methods []serviceMethod) {
	type sk struct{ svc, name string }
	counts := map[sk]int{}
	for _, m := range methods {
		counts[sk{m.Service, m.MethodName}]++
	}
	for i := range methods {
		k := sk{methods[i].Service, methods[i].MethodName}
		if counts[k] <= 1 {
			continue
		}
		methods[i].MethodName = goIdent(methods[i].OperationID)
	}
	seen := map[string]bool{}
	for i := range methods {
		base := methods[i].MethodName
		name := methods[i].Service + "." + base
		n := 2
		for seen[name] {
			methods[i].MethodName = fmt.Sprintf("%s%d", base, n)
			name = methods[i].Service + "." + methods[i].MethodName
			n++
		}
		seen[name] = true
	}
}

func groupServices(methods []serviceMethod) []serviceDef {
	byName := map[string]*serviceDef{}
	var order []string
	for _, m := range methods {
		if _, ok := byName[m.Service]; !ok {
			byName[m.Service] = &serviceDef{
				Name:     m.Service,
				FileBase: strings.ToLower(m.Service),
			}
			order = append(order, m.Service)
		}
		byName[m.Service].Methods = append(byName[m.Service].Methods, m)
	}
	sort.Strings(order)
	out := make([]serviceDef, 0, len(order))
	for _, name := range order {
		out = append(out, *byName[name])
	}
	return out
}

func writeModels(outDir, pkg string, models []modelDef) error {
	path := filepath.Join(outDir, "models.gen.go")
	var buf bytes.Buffer
	if err := modelsTmpl.Execute(&buf, map[string]any{
		"Package": pkg,
		"Models":  models,
	}); err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format %s: %w\n%s", path, err, buf.String())
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d models)\n", path, len(models))
	return nil
}

func writeServices(outDir, pkg, specLabel string, services []serviceDef) error {
	for _, svc := range services {
		needsFmt := false
		for _, m := range svc.Methods {
			if m.NeedsFmt {
				needsFmt = true
				break
			}
		}
		path := filepath.Join(outDir, fmt.Sprintf("%s.gen.go", svc.FileBase))
		var buf bytes.Buffer
		if err := serviceTmpl.Execute(&buf, map[string]any{
			"Package":  pkg,
			"Service":  svc,
			"Spec":     specLabel,
			"NeedsFmt": needsFmt,
		}); err != nil {
			return err
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("format %s: %w\n%s", path, err, buf.String())
		}
		if err := os.WriteFile(path, formatted, 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", path)
	}
	return nil
}

func writeClient(outDir, pkg, specLabel string, services []serviceDef) error {
	path := filepath.Join(outDir, "client.gen.go")
	var buf bytes.Buffer
	if err := clientTmpl.Execute(&buf, map[string]any{
		"Package":  pkg,
		"Services": services,
		"Spec":     specLabel,
	}); err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format %s: %w\n%s", path, err, buf.String())
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d services)\n", path, len(services))
	return nil
}

var tmplFuncs = template.FuncMap{
	"lower": strings.ToLower,
	"export": func(s string) string {
		return goIdent(s)
	},
	"httpMethodConst": func(m string) string {
		switch strings.ToUpper(m) {
		case "GET":
			return "Get"
		case "POST":
			return "Post"
		case "PUT":
			return "Put"
		case "PATCH":
			return "Patch"
		case "DELETE":
			return "Delete"
		default:
			return "Get"
		}
	},
	"sigParams": func(p string) string {
		if p == "" {
			return ""
		}
		return ", " + p
	},
}

var modelsTmpl = template.Must(template.New("models").Parse(`// Code generated by tools/openapi-gen from the bundled OpenAPI spec. DO NOT EDIT.

package {{.Package}}

import "encoding/json"

var _ = json.RawMessage{}

{{range .Models}}
// {{.Name}} is generated from OpenAPI components/schemas.
type {{.Name}} struct {
{{- range .Fields}}
	{{.Name}} {{.Type}} ` + "`json:\"{{.Tag}},omitempty\"`" + `
{{- end}}
}
{{end}}
`))

var serviceTmpl = template.Must(template.New("service").Funcs(tmplFuncs).Parse(`// Code generated by tools/openapi-gen from {{.Spec}}. DO NOT EDIT.

package {{.Package}}

import (
	"context"
{{- if .NeedsFmt}}
	"fmt"
{{- end}}
	"net/http"
)

// {{.Service.Name}}Service is generated from the OpenAPI "{{.Service.Name}}" tag.
type {{.Service.Name}}Service interface {
{{- range .Service.Methods}}
{{- if or .IsList .ReturnsList}}
	{{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) ([]{{.ItemType}}, *Response, error)
{{- else if .IsGet}}
	{{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) (*{{.ItemType}}, *Response, error)
{{- else if .IsCreate}}
	{{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) (*{{.ItemType}}, *Response, error)
{{- else if .IsUpdate}}
	{{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) (*{{.ItemType}}, *Response, error)
{{- else if .IsDelete}}
	{{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) (*Response, error)
{{- end}}
{{- end}}
}

type {{.Service.Name | lower}}ServiceOp struct {
	invoker Invoker
}

var _ {{.Service.Name}}Service = &{{.Service.Name | lower}}ServiceOp{}

func new{{.Service.Name}}Service(inv Invoker) {{.Service.Name}}Service {
	return &{{.Service.Name | lower}}ServiceOp{invoker: inv}
}

{{- range .Service.Methods}}

{{- if .HasRoot}}
type {{.RootType}} struct {
{{- if or .IsList .ReturnsList}}
	{{export .RootKey}} []{{.ItemType}} ` + "`json:\"{{.RootKey}}\"`" + `
	Links               *Links           ` + "`json:\"links\"`" + `
	Meta                *Meta            ` + "`json:\"meta\"`" + `
{{- else}}
	{{export .RootKey}} *{{.ItemType}} ` + "`json:\"{{.RootKey}}\"`" + `
{{- end}}
}
{{- end}}

{{if .Summary}}// {{.MethodName}} {{.Summary}}.
{{else}}// {{.MethodName}} calls {{.OperationID}}.
{{end}}// OperationID: {{.OperationID}}
{{- if .IsList}}
func (s *{{.Service | lower}}ServiceOp) {{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) ([]{{.ItemType}}, *Response, error) {
{{- if .NeedsFmt}}
	path := fmt.Sprintf({{.PathExpr}}{{.FmtArgs}})
{{- else}}
	path := {{.PathExpr}}
{{- end}}
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.invoker.NewRequest(ctx, http.Method{{httpMethodConst .HTTPMethod}}, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new({{.RootType}})
	resp, err := s.invoker.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}
	if m := root.Meta; m != nil {
		resp.Meta = m
	}
	return root.{{export .RootKey}}, resp, err
}
{{- else if .IsGet}}
func (s *{{.Service | lower}}ServiceOp) {{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) (*{{.ItemType}}, *Response, error) {
{{- if .NeedsFmt}}
	path := fmt.Sprintf({{.PathExpr}}{{.FmtArgs}})
{{- else}}
	path := {{.PathExpr}}
{{- end}}
	req, err := s.invoker.NewRequest(ctx, http.Method{{httpMethodConst .HTTPMethod}}, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new({{.RootType}})
	resp, err := s.invoker.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.{{export .RootKey}}, resp, err
}
{{- else if and (or .IsCreate .IsUpdate) .ReturnsList}}
func (s *{{.Service | lower}}ServiceOp) {{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) ([]{{.ItemType}}, *Response, error) {
{{- if .NeedsFmt}}
	path := fmt.Sprintf({{.PathExpr}}{{.FmtArgs}})
{{- else}}
	path := {{.PathExpr}}
{{- end}}
{{- if .HasBody}}
	httpReq, err := s.invoker.NewRequest(ctx, http.Method{{httpMethodConst .HTTPMethod}}, path, req)
{{- else}}
	httpReq, err := s.invoker.NewRequest(ctx, http.Method{{httpMethodConst .HTTPMethod}}, path, nil)
{{- end}}
	if err != nil {
		return nil, nil, err
	}
	root := new({{.RootType}})
	resp, err := s.invoker.Do(ctx, httpReq, root)
	if err != nil {
		return nil, resp, err
	}
	return root.{{export .RootKey}}, resp, err
}
{{- else if or .IsCreate .IsUpdate}}
func (s *{{.Service | lower}}ServiceOp) {{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) (*{{.ItemType}}, *Response, error) {
{{- if .NeedsFmt}}
	path := fmt.Sprintf({{.PathExpr}}{{.FmtArgs}})
{{- else}}
	path := {{.PathExpr}}
{{- end}}
{{- if .HasBody}}
	httpReq, err := s.invoker.NewRequest(ctx, http.Method{{httpMethodConst .HTTPMethod}}, path, req)
{{- else}}
	httpReq, err := s.invoker.NewRequest(ctx, http.Method{{httpMethodConst .HTTPMethod}}, path, nil)
{{- end}}
	if err != nil {
		return nil, nil, err
	}
	root := new({{.RootType}})
	resp, err := s.invoker.Do(ctx, httpReq, root)
	if err != nil {
		return nil, resp, err
	}
	return root.{{export .RootKey}}, resp, err
}
{{- else if .IsDelete}}
func (s *{{.Service | lower}}ServiceOp) {{.MethodName}}(ctx context.Context{{sigParams .FuncParams}}) (*Response, error) {
{{- if .NeedsFmt}}
	path := fmt.Sprintf({{.PathExpr}}{{.FmtArgs}})
{{- else}}
	path := {{.PathExpr}}
{{- end}}
	req, err := s.invoker.NewRequest(ctx, http.Method{{httpMethodConst .HTTPMethod}}, path, nil)
	if err != nil {
		return nil, err
	}
	return s.invoker.Do(ctx, req, nil)
}
{{- end}}
{{- end}}
`))

var clientTmpl = template.Must(template.New("client").Parse(`// Code generated by tools/openapi-gen from {{.Spec}}. DO NOT EDIT.

package {{.Package}}

// Client is the experimental OpenAPI-generated API surface.
// Access it via godo.Client.V2.
type Client struct {
{{- range .Services}}
	{{.Name}} {{.Name}}Service
{{- end}}
}

// New returns a v2 Client that performs requests through inv.
func New(inv Invoker) *Client {
	return &Client{
{{- range .Services}}
		{{.Name}}: new{{.Name}}Service(inv),
{{- end}}
	}
}
`))
