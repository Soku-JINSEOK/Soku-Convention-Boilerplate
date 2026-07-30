// Package manual implements the opt-in real-runtime documentation component.
package manual

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
	"gopkg.in/yaml.v3"
)

var (
	idPattern      = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	envPattern     = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	hostPattern    = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	secretLiteral  = regexp.MustCompile(`(?i)(AIza[0-9A-Za-z_-]{20,}|(?:api[_-]?key|token|secret|signature|password)\s*:\s*[^$\s][^\r\n#]*)`)
	credentialPart = regexp.MustCompile(`(?i)(key|token|signature|secret|password)=([^&\s]+)`)
)

// Error is a stable CLI-facing manual component failure.
type Error struct {
	Code    int
	Key     string
	Message string
	Cause   error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Cause }

func failure(code int, key, format string, values ...any) *Error {
	return &Error{Code: code, Key: key, Message: fmt.Sprintf(format, values...)}
}

// Config is capture configuration schema v1.
type Config struct {
	SchemaVersion int        `yaml:"schema_version" json:"schema_version"`
	Execution     Execution  `yaml:"execution" json:"execution"`
	Runtime       Runtime    `yaml:"runtime" json:"runtime"`
	Browser       Browser    `yaml:"browser" json:"browser"`
	Backend       Backend    `yaml:"backend" json:"backend"`
	Map           Map        `yaml:"map" json:"map"`
	Output        Output     `yaml:"output" json:"output"`
	Scenarios     []Scenario `yaml:"scenarios" json:"scenarios"`
	PDF           *PDFReview `yaml:"pdf,omitempty" json:"pdf,omitempty"`
	Hooks         []Hook     `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	Metadata      *Metadata  `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Execution struct {
	Mode          string   `yaml:"mode" json:"mode"`
	AllowHosts    []string `yaml:"allow_hosts" json:"allow_hosts"`
	RequestBudget int      `yaml:"request_budget" json:"request_budget"`
}

type Runtime struct {
	Adapter         string   `yaml:"adapter" json:"adapter"`
	Command         []string `yaml:"command,omitempty" json:"command,omitempty"`
	HealthURL       string   `yaml:"health_url,omitempty" json:"health_url,omitempty"`
	StaticDirectory string   `yaml:"static_directory,omitempty" json:"static_directory,omitempty"`
	SourceFiles     []string `yaml:"source_files,omitempty" json:"source_files,omitempty"`
	SourceFragments []string `yaml:"source_fragments,omitempty" json:"source_fragments,omitempty"`
	Serve           Serve    `yaml:"serve" json:"serve"`
}

type Serve struct {
	Origin string `yaml:"origin" json:"origin"`
}

type Browser struct {
	Engine            string   `yaml:"engine" json:"engine"`
	Viewport          Viewport `yaml:"viewport" json:"viewport"`
	DeviceScaleFactor float64  `yaml:"device_scale_factor" json:"device_scale_factor"`
	Locale            string   `yaml:"locale" json:"locale"`
	Timezone          string   `yaml:"timezone" json:"timezone"`
	Fonts             []string `yaml:"fonts" json:"fonts"`
}

type Viewport struct {
	Width  int `yaml:"width" json:"width"`
	Height int `yaml:"height" json:"height"`
}

type Backend struct {
	Mode    string  `yaml:"mode" json:"mode"`
	Adapter string  `yaml:"adapter,omitempty" json:"adapter,omitempty"`
	Fixture string  `yaml:"fixture,omitempty" json:"fixture,omitempty"`
	HAR     string  `yaml:"har,omitempty" json:"har,omitempty"`
	Routes  []Route `yaml:"routes,omitempty" json:"routes,omitempty"`
}

type Route struct {
	Method  string `yaml:"method" json:"method"`
	URL     string `yaml:"url" json:"url"`
	Status  int    `yaml:"status" json:"status"`
	Fixture string `yaml:"fixture,omitempty" json:"fixture,omitempty"`
	Body    string `yaml:"body,omitempty" json:"body,omitempty"`
}

type Map struct {
	Provider            string      `yaml:"provider" json:"provider"`
	SourceRelation      string      `yaml:"source_relation" json:"source_relation"`
	APIKeyEnv           string      `yaml:"api_key_env,omitempty" json:"api_key_env,omitempty"`
	BillingOwner        string      `yaml:"billing_owner,omitempty" json:"billing_owner,omitempty"`
	RestrictionReviewed bool        `yaml:"restriction_reviewed,omitempty" json:"restriction_reviewed,omitempty"`
	ExecutionMode       string      `yaml:"execution_mode" json:"execution_mode"`
	MapLoadBudget       int         `yaml:"map_load_budget" json:"map_load_budget"`
	RequestBudget       int         `yaml:"request_budget" json:"request_budget"`
	Readiness           Readiness   `yaml:"readiness" json:"readiness"`
	Attribution         Attribution `yaml:"attribution" json:"attribution"`
}

type Readiness struct {
	Type string `yaml:"type" json:"type"`
	Name string `yaml:"name" json:"name"`
}

type Attribution struct {
	Required bool   `yaml:"required" json:"required"`
	Locator  string `yaml:"locator,omitempty" json:"locator,omitempty"`
}

type Output struct {
	Directory      string `yaml:"directory" json:"directory"`
	Report         string `yaml:"report" json:"report"`
	GeneratedIndex string `yaml:"generated_index" json:"generated_index"`
}

type Scenario struct {
	ID    string `yaml:"id" json:"id"`
	Route string `yaml:"route" json:"route"`
	Steps []Step `yaml:"steps" json:"steps"`
}

type Step struct {
	Action   string            `yaml:"action" json:"action"`
	ID       string            `yaml:"id,omitempty" json:"id,omitempty"`
	Locator  *Locator          `yaml:"locator,omitempty" json:"locator,omitempty"`
	Value    string            `yaml:"value,omitempty" json:"value,omitempty"`
	Wait     *Wait             `yaml:"wait,omitempty" json:"wait,omitempty"`
	Capture  *Capture          `yaml:"capture,omitempty" json:"capture,omitempty"`
	Dialog   *Dialog           `yaml:"dialog,omitempty" json:"dialog,omitempty"`
	Metadata map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Locator struct {
	By    string `yaml:"by" json:"by"`
	Value string `yaml:"value" json:"value"`
}

type Wait struct {
	State  string `yaml:"state" json:"state"`
	Value  string `yaml:"value,omitempty" json:"value,omitempty"`
	HoldMS int    `yaml:"hold_ms,omitempty" json:"hold_ms,omitempty"`
}

type Capture struct {
	Mode     string   `yaml:"mode" json:"mode"`
	Locator  *Locator `yaml:"locator,omitempty" json:"locator,omitempty"`
	Region   *Region  `yaml:"region,omitempty" json:"region,omitempty"`
	Padding  int      `yaml:"padding,omitempty" json:"padding,omitempty"`
	FullPage bool     `yaml:"full_page,omitempty" json:"full_page,omitempty"`
	Caption  string   `yaml:"caption" json:"caption"`
}

type Region struct {
	X      int `yaml:"x" json:"x"`
	Y      int `yaml:"y" json:"y"`
	Width  int `yaml:"width" json:"width"`
	Height int `yaml:"height" json:"height"`
}

type Dialog struct {
	Mode    string `yaml:"mode" json:"mode"`
	Action  string `yaml:"action" json:"action"`
	Message string `yaml:"message,omitempty" json:"message,omitempty"`
}

type PDFReview struct {
	Path          string `yaml:"path" json:"path"`
	ExpectedPages int    `yaml:"expected_pages" json:"expected_pages"`
	RasterOutput  string `yaml:"raster_output" json:"raster_output"`
}

type Hook struct {
	Name string `yaml:"name" json:"name"`
	Path string `yaml:"path" json:"path"`
}

type Metadata struct {
	ManualTitle string `yaml:"manual_title" json:"manual_title"`
}

// LoadedConfig preserves the canonical source identity needed for plans.
type LoadedConfig struct {
	Config Config
	Path   string
	Hash   string
}

// LoadConfig strictly decodes exactly one YAML document and validates all
// semantic safety boundaries before returning it.
func LoadConfig(root, configPath string) (LoadedConfig, error) {
	if err := manifest.ValidatePath(configPath); err != nil {
		return LoadedConfig{}, failure(2, "manual.configuration.invalid", "config path is not portable: %v", err)
	}
	fullPath := pathFromRoot(root, configPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return LoadedConfig{}, failure(2, "manual.configuration.invalid", "read capture configuration %q: %v", configPath, err)
	}
	if secretLiteral.Match(data) {
		return LoadedConfig{}, failure(2, "manual.configuration.secret", "capture configuration contains a credential-like literal; use an environment-variable name")
	}
	var config Config
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return LoadedConfig{}, failure(2, "manual.configuration.invalid", "decode capture configuration: %v", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return LoadedConfig{}, failure(2, "manual.configuration.invalid", "capture configuration must contain exactly one YAML document")
	}
	if err := validateConfig(config); err != nil {
		return LoadedConfig{}, err
	}
	sum := sha256.Sum256(data)
	return LoadedConfig{Config: config, Path: configPath, Hash: hex.EncodeToString(sum[:])}, nil
}

func validateConfig(config Config) error {
	if config.SchemaVersion != 1 {
		return failure(2, "manual.configuration.invalid", "schema_version must be 1")
	}
	if config.Execution.Mode != "local-manual" {
		return failure(4, "manual.execution.refused", "execution.mode must be local-manual")
	}
	if config.Execution.RequestBudget < 0 || config.Execution.RequestBudget > 100 {
		return failure(2, "manual.configuration.invalid", "execution.request_budget must be between 0 and 100")
	}
	hosts := append([]string(nil), config.Execution.AllowHosts...)
	sort.Strings(hosts)
	if strings.Join(hosts, "\x00") != strings.Join(config.Execution.AllowHosts, "\x00") {
		return failure(2, "manual.configuration.invalid", "execution.allow_hosts must be sorted")
	}
	for index, host := range hosts {
		if host == "" || strings.ContainsAny(host, "/:@?#") || !hostPattern.MatchString(host) {
			return failure(2, "manual.configuration.invalid", "allowlisted host %q is invalid", host)
		}
		if index > 0 && host == hosts[index-1] {
			return failure(2, "manual.configuration.invalid", "execution.allow_hosts contains duplicate host %q", host)
		}
	}
	if !oneOf(config.Runtime.Adapter, "dev-server", "static-build", "gas-html-service") {
		return failure(2, "manual.configuration.invalid", "runtime.adapter is unsupported")
	}
	if config.Runtime.Serve.Origin != "loopback-http" {
		return failure(4, "manual.execution.refused", "runtime.serve.origin must be loopback-http")
	}
	switch config.Runtime.Adapter {
	case "dev-server":
		if len(config.Runtime.Command) == 0 || !loopbackURL(config.Runtime.HealthURL) || len(config.Runtime.SourceFiles) == 0 {
			return failure(2, "manual.configuration.invalid", "dev-server requires command argv, source_files, and a loopback health_url")
		}
	case "static-build":
		if err := validateProjectPath("runtime.static_directory", config.Runtime.StaticDirectory); err != nil {
			return err
		}
	case "gas-html-service":
		if len(config.Runtime.SourceFragments) == 0 {
			return failure(2, "manual.configuration.invalid", "gas-html-service requires source_fragments")
		}
		for _, fragment := range config.Runtime.SourceFragments {
			if err := validateProjectPath("runtime.source_fragments", fragment); err != nil {
				return err
			}
		}
	}
	for _, source := range config.Runtime.SourceFiles {
		if err := validateProjectPath("runtime.source_files", source); err != nil {
			return err
		}
	}
	if config.Browser.Engine != "chromium" || config.Browser.Viewport.Width < 320 || config.Browser.Viewport.Width > 3840 ||
		config.Browser.Viewport.Height < 320 || config.Browser.Viewport.Height > 2160 ||
		config.Browser.DeviceScaleFactor < 1 || config.Browser.DeviceScaleFactor > 3 ||
		config.Browser.Locale == "" || config.Browser.Timezone == "" {
		return failure(2, "manual.configuration.invalid", "browser settings are outside the supported deterministic bounds")
	}
	if len(config.Browser.Fonts) == 0 {
		return failure(2, "manual.configuration.invalid", "browser.fonts must declare at least one reviewed font family")
	}
	if !oneOf(config.Backend.Mode, "real-local", "http-route", "har-replay", "browser-bridge", "none") {
		return failure(2, "manual.configuration.invalid", "backend.mode is unsupported")
	}
	if config.Backend.Fixture != "" {
		if err := validateProjectPath("backend.fixture", config.Backend.Fixture); err != nil {
			return err
		}
		if !strings.Contains(config.Backend.Fixture, ".synthetic.") {
			return failure(4, "manual.fixture.refused", "backend fixture filename must disclose synthetic data")
		}
	}
	if config.Backend.Mode == "har-replay" {
		if err := validateProjectPath("backend.har", config.Backend.HAR); err != nil {
			return err
		}
		if !strings.HasSuffix(strings.ToLower(config.Backend.HAR), ".sanitized.har") {
			return failure(4, "manual.har.refused", "HAR replay requires a *.sanitized.har file")
		}
	}
	for _, route := range config.Backend.Routes {
		if route.Method == "" || route.URL == "" || route.Status < 100 || route.Status > 599 || !strings.HasPrefix(route.URL, "/") {
			return failure(2, "manual.configuration.invalid", "HTTP route mocks require a method, relative URL, and valid status")
		}
		if route.Fixture != "" {
			if err := validateProjectPath("backend.routes.fixture", route.Fixture); err != nil {
				return err
			}
			if !strings.Contains(route.Fixture, ".synthetic.") {
				return failure(4, "manual.fixture.refused", "HTTP route fixture filename must disclose synthetic data")
			}
		}
	}
	if err := validateMap(config); err != nil {
		return err
	}
	for _, item := range []struct{ name, value string }{
		{"output.directory", config.Output.Directory},
		{"output.report", config.Output.Report},
		{"output.generated_index", config.Output.GeneratedIndex},
	} {
		if err := validateProjectPath(item.name, item.value); err != nil {
			return err
		}
	}
	if len(config.Scenarios) == 0 {
		return failure(2, "manual.configuration.invalid", "at least one scenario is required")
	}
	scenarios := map[string]bool{}
	captures := map[string]bool{}
	for _, scenario := range config.Scenarios {
		if !idPattern.MatchString(scenario.ID) || scenarios[scenario.ID] || !strings.HasPrefix(scenario.Route, "/") || len(scenario.Steps) == 0 {
			return failure(2, "manual.configuration.invalid", "scenario ids must be unique and routes must be relative")
		}
		scenarios[scenario.ID] = true
		overlayPending := false
		for _, step := range scenario.Steps {
			if err := validateStep(step, captures); err != nil {
				return err
			}
			if step.Action == "dialog" && step.Dialog != nil && step.Dialog.Mode == "documentation-overlay" {
				overlayPending = true
			}
			if step.Action == "capture" && overlayPending {
				caption := strings.ToLower(step.Capture.Caption)
				if !strings.Contains(caption, "documentation") || !strings.Contains(caption, "overlay") {
					return failure(4, "manual.disclosure.refused", "documentation-overlay capture caption must disclose the documentation overlay")
				}
				overlayPending = false
			}
		}
	}
	if config.PDF != nil {
		if err := validateProjectPath("pdf.path", config.PDF.Path); err != nil {
			return err
		}
		if err := validateProjectPath("pdf.raster_output", config.PDF.RasterOutput); err != nil {
			return err
		}
		if config.PDF.ExpectedPages <= 0 {
			return failure(2, "manual.configuration.invalid", "pdf.expected_pages must be positive")
		}
	}
	hooks := map[string]bool{}
	for _, hook := range config.Hooks {
		if !idPattern.MatchString(hook.Name) || hooks[hook.Name] {
			return failure(2, "manual.configuration.invalid", "hook names must be portable and unique")
		}
		hooks[hook.Name] = true
		if err := validateProjectPath("hooks.path", hook.Path); err != nil {
			return err
		}
	}
	return nil
}

func validateMap(config Config) error {
	value := config.Map
	if !oneOf(value.Provider, "none", "local-deterministic", "leaflet-osm", "google-maps-javascript") {
		return failure(2, "manual.configuration.invalid", "map.provider is unsupported")
	}
	if !oneOf(value.SourceRelation, "original-application", "declared-adapter", "not-applicable") ||
		value.ExecutionMode != "local-manual" || value.MapLoadBudget < 0 || value.RequestBudget < 0 {
		return failure(2, "manual.configuration.invalid", "map source relation, execution mode, or budgets are invalid")
	}
	if value.Provider == "none" {
		if value.SourceRelation != "not-applicable" || value.MapLoadBudget != 0 ||
			value.RequestBudget != 0 || !sameStrings(config.Execution.AllowHosts, []string{}) {
			return failure(2, "manual.configuration.invalid", "map provider none requires not-applicable relation and zero budgets")
		}
		return nil
	}
	if value.Readiness.Name == "" || !oneOf(value.Readiness.Type, "hook", "event", "leaflet-load", "google-idle", "deterministic") {
		return failure(4, "manual.map.refused", "map provider requires named state-based readiness")
	}
	if !value.Attribution.Required || value.Attribution.Locator == "" {
		return failure(4, "manual.map.refused", "map provider requires visible attribution and a locator")
	}
	switch value.Provider {
	case "local-deterministic":
		if value.MapLoadBudget > 1 || value.RequestBudget != 0 ||
			!sameStrings(config.Execution.AllowHosts, []string{}) {
			return failure(2, "manual.configuration.invalid", "local deterministic map permits one load and no external requests")
		}
	case "leaflet-osm":
		if value.MapLoadBudget != 1 || value.RequestBudget < 1 || value.RequestBudget > 32 ||
			!sameStrings(config.Execution.AllowHosts, []string{"tile.openstreetmap.org"}) {
			return failure(4, "manual.map.refused", "leaflet-osm requires one map load, a bounded request budget, and the reviewed OSM tile host")
		}
	case "google-maps-javascript":
		if !envPattern.MatchString(value.APIKeyEnv) || value.APIKeyEnv != "GOOGLE_MAPS_API_KEY" ||
			strings.TrimSpace(value.BillingOwner) == "" || !value.RestrictionReviewed ||
			value.MapLoadBudget != 1 || value.RequestBudget < 1 || value.RequestBudget > 32 {
			return failure(4, "manual.map.refused", "Google Maps requires GOOGLE_MAPS_API_KEY, billing owner, restriction review, one load, and a bounded request budget")
		}
		required := []string{"maps.googleapis.com", "maps.gstatic.com"}
		if !sameStrings(config.Execution.AllowHosts, required) {
			return failure(4, "manual.map.refused", "Google Maps egress must exactly match provider profile v1")
		}
	}
	if value.SourceRelation == "declared-adapter" {
		for _, scenario := range config.Scenarios {
			for _, step := range scenario.Steps {
				if step.Action == "capture" && (step.Capture == nil || !strings.Contains(strings.ToLower(step.Capture.Caption), "adapter")) {
					return failure(4, "manual.disclosure.refused", "provider substitution requires adapter disclosure in every capture caption")
				}
			}
		}
	}
	return nil
}

func sameStrings(left, right []string) bool {
	return strings.Join(left, "\x00") == strings.Join(right, "\x00")
}

func validateStep(step Step, captures map[string]bool) error {
	if !oneOf(step.Action, "goto", "click", "fill", "select", "wait", "capture", "dialog") {
		return failure(2, "manual.configuration.invalid", "unsupported scenario action %q", step.Action)
	}
	if step.Locator != nil {
		if !oneOf(step.Locator.By, "role", "label", "test-id", "text", "css") || step.Locator.Value == "" {
			return failure(2, "manual.configuration.invalid", "locator must use role, label, test-id, text, or explicit css")
		}
	}
	switch step.Action {
	case "click", "fill", "select":
		if step.Locator == nil {
			return failure(2, "manual.configuration.invalid", "%s action requires a locator", step.Action)
		}
	case "wait":
		if step.Wait == nil || !oneOf(step.Wait.State, "visible", "hidden", "enabled", "text", "attribute", "event", "hook", "fonts-ready") ||
			step.Wait.HoldMS < 0 || step.Wait.HoldMS > 5000 {
			return failure(2, "manual.configuration.invalid", "wait action is invalid")
		}
		if step.Wait.State == "attribute" {
			name, _, found := strings.Cut(step.Wait.Value, "=")
			if !found || name == "" || strings.ContainsAny(name, " \t\r\n") {
				return failure(2, "manual.configuration.invalid", "attribute wait value must use name=value")
			}
		}
	case "capture":
		if !idPattern.MatchString(step.ID) || captures[step.ID] || step.Capture == nil || step.Capture.Caption == "" {
			return failure(2, "manual.configuration.invalid", "capture ids must be unique and captions are required")
		}
		captures[step.ID] = true
		if !oneOf(step.Capture.Mode, "viewport", "full-page", "locator", "region", "sequence") ||
			step.Capture.Padding < 0 || step.Capture.Padding > 256 {
			return failure(2, "manual.configuration.invalid", "capture mode or padding is invalid")
		}
		if step.Capture.Mode == "locator" && step.Capture.Locator == nil && step.Locator == nil {
			return failure(2, "manual.configuration.invalid", "locator capture requires a locator")
		}
		if step.Capture.Mode == "region" && (step.Capture.Region == nil || step.Capture.Region.Width <= 0 || step.Capture.Region.Height <= 0) {
			return failure(2, "manual.configuration.invalid", "region capture requires positive bounds")
		}
	case "dialog":
		if step.Dialog == nil || !oneOf(step.Dialog.Mode, "native", "app-owned", "documentation-overlay") ||
			!oneOf(step.Dialog.Action, "accept", "dismiss") {
			return failure(2, "manual.configuration.invalid", "dialog action is invalid")
		}
	}
	return nil
}

func validateProjectPath(field, value string) error {
	if err := manifest.ValidatePath(value); err != nil {
		return failure(2, "manual.configuration.invalid", "%s: %v", field, err)
	}
	return nil
}

func loopbackURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "http" || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	host := parsed.Hostname()
	return host == "localhost" || net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
}

func pathFromRoot(root, value string) string {
	return filepath.Join(root, filepath.FromSlash(value))
}

func oneOf(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}
	return false
}

func redact(value string) string {
	return credentialPart.ReplaceAllString(value, "$1=[REDACTED]")
}
