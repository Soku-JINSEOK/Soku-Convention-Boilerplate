package cicd

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	engineSource = "github:Soku-JINSEOK/ci-cd-control-plane-engine"
	engineRef    = "56c89877446817112819cc43a77d027d4bedb292"
)

// rendererAssets is deliberately embedded. The renderer never downloads or
// executes an adapter, provider, action, image, or remote template.
//
//go:embed templates/*
var rendererAssets embed.FS

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// AdapterCatalog is the repository-owned, immutable mapping boundary between
// the portable decision contract and validation-only renderers.
type AdapterCatalog struct {
	SchemaVersion string           `json:"schema_version"`
	Engine        EngineProvenance `json:"engine"`
	Mappings      []AdapterMapping `json:"mappings"`
}

type EngineProvenance struct {
	Source string           `json:"source"`
	Ref    string           `json:"ref"`
	Files  []ProvenanceFile `json:"files"`
}

type ProvenanceFile struct {
	Path                 string `json:"path"`
	SHA256               string `json:"sha256"`
	ImplementationSHA256 string `json:"implementation_sha256,omitempty"`
}

type AdapterMapping struct {
	MappingID              string              `json:"mapping_id"`
	Platform               string              `json:"platform"`
	AdapterID              string              `json:"adapter_id"`
	AdapterRef             string              `json:"adapter_ref"`
	EngineAdapterID        string              `json:"engine_adapter_id"`
	EngineDescriptorPath   string              `json:"engine_descriptor_path"`
	EngineDescriptorSHA256 string              `json:"engine_descriptor_sha256"`
	ImplementationSHA256   string              `json:"implementation_sha256"`
	RendererID             string              `json:"renderer_id"`
	RendererTemplate       string              `json:"renderer_template"`
	RendererSHA256         string              `json:"renderer_sha256"`
	Verification           MappingVerification `json:"verification"`
	Runner                 RunnerRequirements  `json:"runner"`
	DeliveryAuthority      string              `json:"delivery_authority"`
	OutputPath             string              `json:"output_path"`
}

type MappingVerification struct {
	PR       string   `json:"pr"`
	PRArgv   []string `json:"pr_argv"`
	Full     string   `json:"full"`
	FullArgv []string `json:"full_argv"`
}

type RunnerRequirements struct {
	OS            []string `json:"os"`
	Architectures []string `json:"architectures"`
	Capabilities  []string `json:"capabilities"`
	Networks      []string `json:"networks"`
}

type RenderedMapping struct {
	Path    string
	Mapping AdapterMapping
	Content []byte
}

var expectedUpstreamFiles = map[string]string{
	"execution/adapters/gcp-cloud-build-v1.synthetic.json":   "beff940bc5eb93d7e69cdd7a4801003e3cda3dd1b8c9b9a636a6e22efd2be1fb",
	"execution/adapters/jenkins-hybrid-v1.synthetic.json":    "1be9ac0467098a25cfcabf7c3cf9d8db2c54c333da58901e604d7a9e3ea00eca",
	"execution/adapters/local-reference-v1.json":             "5d5e6cfadc8d0bb2d31686f4ad9954e320a2012b6eaa71b6c729149a8d5f4a54",
	"execution/fixtures/portable-local-success.json":         "adf82c61c074cd927ad7e274e2741fa766d7d613df601ae74a8c2b2824d038ac",
	"execution/fixtures/portable-reviewed-contract.json":     "c9ee836f5c1f4046772df78c263d913abfaa53e3156e43faade92ced86f1c35a",
	"execution/schema/adapter-descriptor-v1.schema.json":     "6c0dd7ea93fa029a00a45a7b34b3ebc355643c888f262d82963fd352cf080345",
	"execution/schema/execution-requirements-v1.schema.json": "e355c414d72e9f0b2092f6a12a5a42960933f180cf80ea190f89e18d475c2e4f",
	"execution/schema/normalized-evidence-v1.schema.json":    "e514a42bb17418a0fc8f9bec882006d9e80664d162e2cb0611b12bdb707c3d57",
}

var expectedUpstreamImplementations = map[string]string{
	"execution/adapters/gcp-cloud-build-v1.synthetic.json": "15e2b0d3c33891ebb115c06ea0e24e8db76162125a0188686ee864fe068c6a4e",
	"execution/adapters/jenkins-hybrid-v1.synthetic.json":  "33b82051ca56d41f3ce885264880101a079893acef4367334d0cbad221f93423",
	"execution/adapters/local-reference-v1.json":           "25a0a928b72e6286bca1fe6b8811b2037475e6711fb7b84aaa92bee5e6dea8a7",
}

var expectedMappingByPlatform = map[string]struct {
	mappingID            string
	adapterID            string
	engineAdapterID      string
	engineDescriptorPath string
	outputPath           string
	rendererID           string
	rendererTemplate     string
}{
	platformGitHubHosted: {
		mappingID: "ci-only-github-hosted-v1", adapterID: "github-hosted-validation-v1",
		engineAdapterID: "jenkins-hybrid-v1", engineDescriptorPath: "execution/adapters/jenkins-hybrid-v1.synthetic.json",
		outputPath: ".github/workflows/soku-ci.yml", rendererID: "github-hosted-validation-v1", rendererTemplate: "github-hosted-validation-v1.yml",
	},
	platformGCPManaged: {
		mappingID: "ci-only-gcp-managed-v1", adapterID: "gcp-managed-validation-v1",
		engineAdapterID: "gcp-cloud-build-v1", engineDescriptorPath: "execution/adapters/gcp-cloud-build-v1.synthetic.json",
		outputPath: "cloudbuild/soku-ci-validation.yaml", rendererID: "gcp-managed-validation-v1", rendererTemplate: "gcp-managed-validation-v1.yaml",
	},
	platformGitHubSelfHosted: {
		mappingID: "ci-only-github-self-hosted-v1", adapterID: "github-self-hosted-validation-v1",
		engineAdapterID: "jenkins-hybrid-v1", engineDescriptorPath: "execution/adapters/jenkins-hybrid-v1.synthetic.json",
		outputPath: ".soku/ci-cd/github-self-hosted-validation.yml", rendererID: "github-self-hosted-validation-v1", rendererTemplate: "github-self-hosted-validation-v1.yml",
	},
}

func fixedPRArgv() []string {
	return []string{"scripts/verify.sh", "--profile", "ci-quick", "--group", "<group-id>", "--base", "<base-sha>", "--head", "<head-sha>"}
}

func fixedFullArgv() []string {
	return []string{"scripts/verify.sh", "--profile", "full"}
}

// LoadAdapterCatalog reads and validates the repository-owned mapping catalog.
// It performs no network access and never executes the vendored material.
func LoadAdapterCatalog(root string) (AdapterCatalog, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(AdapterCatalogFile)))
	if err != nil {
		return AdapterCatalog{}, invalid("ci-cd.catalog.missing", "read adapter mapping catalog: %v", err)
	}
	catalog, err := DecodeAdapterCatalog(data)
	if err != nil {
		return AdapterCatalog{}, err
	}
	if err := validateAdapterCatalog(root, catalog); err != nil {
		return AdapterCatalog{}, err
	}
	return catalog, nil
}

// DecodeAdapterCatalog validates the strict JSON shape without consulting the
// repository. ValidateAdapterCatalog is the complete provenance check.
func DecodeAdapterCatalog(data []byte) (AdapterCatalog, error) {
	var catalog AdapterCatalog
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		return AdapterCatalog{}, invalid("ci-cd.catalog.schema", "decode adapter mapping catalog: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return AdapterCatalog{}, invalid("ci-cd.catalog.schema", "adapter mapping catalog contains multiple JSON values")
	}
	if catalog.SchemaVersion != "ci-cd-adapter-mapping-v1" {
		return AdapterCatalog{}, invalid("ci-cd.catalog.schema", "unsupported adapter mapping catalog version")
	}
	return catalog, nil
}

func validateAdapterCatalog(root string, catalog AdapterCatalog) error {
	if catalog.Engine.Source != engineSource || catalog.Engine.Ref != engineRef || !immutableSHA.MatchString(catalog.Engine.Ref) {
		return invalid("ci-cd.catalog.provenance", "adapter catalog engine provenance is not the reviewed immutable source")
	}
	if len(catalog.Engine.Files) != len(expectedUpstreamFiles) {
		return invalid("ci-cd.catalog.provenance", "adapter catalog must enumerate every vendored upstream file exactly once")
	}
	seenFiles := make(map[string]bool, len(catalog.Engine.Files))
	previousFile := ""
	for _, file := range catalog.Engine.Files {
		if previousFile != "" && file.Path <= previousFile {
			return invalid("ci-cd.catalog.provenance", "adapter catalog upstream files must be sorted")
		}
		previousFile = file.Path
		if seenFiles[file.Path] || file.Path == "" || !sha256Pattern.MatchString(file.SHA256) {
			return invalid("ci-cd.catalog.provenance", "adapter catalog contains an invalid or repeated upstream file")
		}
		if expectedImplementation, isAdapter := expectedUpstreamImplementations[file.Path]; isAdapter {
			if file.ImplementationSHA256 != expectedImplementation {
				return invalid("ci-cd.catalog.provenance", "upstream implementation provenance is stale for %q", file.Path)
			}
		} else if file.ImplementationSHA256 != "" {
			return invalid("ci-cd.catalog.provenance", "non-adapter upstream file %q contains an implementation hash", file.Path)
		}
		seenFiles[file.Path] = true
		expected, ok := expectedUpstreamFiles[file.Path]
		if !ok || file.SHA256 != expected {
			return invalid("ci-cd.catalog.provenance", "upstream provenance is stale for %q", file.Path)
		}
		localPath := filepath.Join(root, "soku", "internal", "cicd", "testdata", "upstream", filepath.FromSlash(strings.TrimPrefix(file.Path, "execution/")))
		data, err := os.ReadFile(localPath)
		if err != nil || sha256Hex(data) != file.SHA256 {
			return invalid("ci-cd.catalog.provenance", "vendored upstream file does not match its recorded hash: %q", file.Path)
		}
	}
	if len(seenFiles) != len(expectedUpstreamFiles) {
		return invalid("ci-cd.catalog.provenance", "adapter catalog upstream file set is incomplete")
	}
	if len(catalog.Mappings) != len(expectedMappingByPlatform) {
		return invalid("ci-cd.catalog.mapping", "adapter catalog must define exactly three validation mappings")
	}
	seenMappings := map[string]bool{}
	seenPlatforms := map[string]bool{}
	previousMapping := ""
	for _, mapping := range catalog.Mappings {
		if previousMapping != "" && mapping.MappingID <= previousMapping {
			return invalid("ci-cd.catalog.mapping", "adapter mappings must be sorted by mapping_id")
		}
		previousMapping = mapping.MappingID
		if seenMappings[mapping.MappingID] || seenPlatforms[mapping.Platform] {
			return invalid("ci-cd.catalog.mapping", "adapter mapping IDs and platforms must be unique")
		}
		seenMappings[mapping.MappingID] = true
		seenPlatforms[mapping.Platform] = true
		if err := validateAdapterMapping(root, catalog.Engine, mapping); err != nil {
			return err
		}
	}
	for platform := range expectedMappingByPlatform {
		if !seenPlatforms[platform] {
			return invalid("ci-cd.catalog.mapping", "adapter catalog is missing the %s mapping", platform)
		}
	}
	return nil
}

// ValidateAdapterCatalog is the explicit conformance entry point used by
// lifecycle code and tests. It checks schema shape, upstream byte hashes,
// immutable descriptors, renderer semantics, and the exact three mappings.
func ValidateAdapterCatalog(root string) error {
	_, err := LoadAdapterCatalog(root)
	return err
}

// ValidateRendered checks a rendered caller against the semantic safety
// boundary. It is intentionally independent from writing or executing it.
func ValidateRendered(mapping AdapterMapping, content []byte) error {
	return validateRenderedSemantics(mapping, content)
}

func validateAdapterMapping(root string, engine EngineProvenance, mapping AdapterMapping) error {
	expected, ok := expectedMappingByPlatform[mapping.Platform]
	if !ok || mapping.MappingID != expected.mappingID || mapping.AdapterID != expected.adapterID ||
		mapping.EngineAdapterID != expected.engineAdapterID || mapping.EngineDescriptorPath != expected.engineDescriptorPath ||
		mapping.OutputPath != expected.outputPath || mapping.RendererID != expected.rendererID || mapping.RendererTemplate != expected.rendererTemplate {
		return invalid("ci-cd.catalog.mapping", "adapter mapping %q is outside the reviewed three-mapping set", mapping.MappingID)
	}
	if mapping.AdapterRef != engine.Ref || !immutableSHA.MatchString(mapping.AdapterRef) || mapping.DeliveryAuthority != "none" {
		return invalid("ci-cd.catalog.provenance", "mapping %q does not bind an immutable engine ref and no delivery authority", mapping.MappingID)
	}
	if !sha256Pattern.MatchString(mapping.EngineDescriptorSHA256) || !sha256Pattern.MatchString(mapping.ImplementationSHA256) || !sha256Pattern.MatchString(mapping.RendererSHA256) {
		return invalid("ci-cd.catalog.provenance", "mapping %q contains an invalid provenance hash", mapping.MappingID)
	}
	expectedDescriptorHash, ok := expectedUpstreamFiles[mapping.EngineDescriptorPath]
	if !ok || mapping.EngineDescriptorSHA256 != expectedDescriptorHash {
		return invalid("ci-cd.catalog.provenance", "mapping %q references a stale descriptor hash", mapping.MappingID)
	}
	descriptorPath := filepath.Join(root, "soku", "internal", "cicd", "testdata", "upstream", filepath.FromSlash(strings.TrimPrefix(mapping.EngineDescriptorPath, "execution/")))
	descriptorData, err := os.ReadFile(descriptorPath)
	if err != nil || sha256Hex(descriptorData) != mapping.EngineDescriptorSHA256 {
		return invalid("ci-cd.catalog.provenance", "mapping %q descriptor bytes do not match provenance", mapping.MappingID)
	}
	var descriptor upstreamAdapterDescriptor
	if err := decodeStrictJSON(descriptorData, &descriptor); err != nil {
		return invalid("ci-cd.catalog.provenance", "mapping %q descriptor is invalid: %v", mapping.MappingID, err)
	}
	if descriptor.SchemaVersion != "adapter-descriptor-v1" || descriptor.AdapterID != mapping.EngineAdapterID ||
		descriptor.ImplementationSHA256 != mapping.ImplementationSHA256 || descriptor.SourceBinding != "exact-clean-git-sha" ||
		descriptor.DeliveryAuthority != "none" {
		return invalid("ci-cd.catalog.provenance", "mapping %q descriptor identity is stale or unsafe", mapping.MappingID)
	}
	if err := validateDescriptorLifecycle(mapping.MappingID, descriptor.Lifecycle); err != nil {
		return err
	}
	if mapping.Verification.PR != "ci-quick" || mapping.Verification.Full != "full" ||
		!equalStrings(mapping.Verification.PRArgv, fixedPRArgv()) || !equalStrings(mapping.Verification.FullArgv, fixedFullArgv()) {
		return invalid("ci-cd.catalog.profile", "mapping %q does not bind the fixed ci-quick/full argv contract", mapping.MappingID)
	}
	if err := validateRunner(mapping.Runner); err != nil {
		return invalid("ci-cd.catalog.capability", "mapping %q: %v", mapping.MappingID, err)
	}
	if !runnerMatchesPlatform(mapping.Platform, mapping.Runner) {
		return invalid("ci-cd.catalog.capability", "mapping %q runner capability does not match its platform", mapping.MappingID)
	}
	content, err := rendererBytes(mapping.RendererTemplate)
	if err != nil || sha256Hex(content) != mapping.RendererSHA256 {
		return invalid("ci-cd.catalog.renderer", "mapping %q renderer template hash is stale", mapping.MappingID)
	}
	if err := validateRenderedSemantics(mapping, content); err != nil {
		return err
	}
	return nil
}

func validateDescriptorLifecycle(mappingID, lifecycle string) error {
	if lifecycle != "experimental" && lifecycle != "stable" {
		return invalid("ci-cd.catalog.lifecycle", "mapping %q uses a deprecated or disabled adapter lifecycle", mappingID)
	}
	return nil
}

func resolveAdapter(root, platform string, decision Decision) (AdapterResolution, bool, error) {
	catalog, err := LoadAdapterCatalog(root)
	if err != nil {
		return AdapterResolution{}, false, err
	}
	mapping, ok := mappingForPlatform(catalog, platform)
	if !ok {
		return AdapterResolution{Status: "unpublished", Reason: "selected platform has no trusted adapter mapping"}, false, nil
	}
	if missing := missingRunnerRequirements(mapping.Runner, decision); len(missing) > 0 {
		return AdapterResolution{
			Status:         "incompatible",
			MappingID:      mapping.MappingID,
			AdapterID:      mapping.AdapterID,
			AdapterRef:     mapping.AdapterRef,
			AdapterSHA256:  mapping.ImplementationSHA256,
			RendererID:     mapping.RendererID,
			RendererSHA256: mapping.RendererSHA256,
			Reason:         "selected mapping does not satisfy: " + strings.Join(missing, ", "),
		}, false, nil
	}
	return AdapterResolution{
		Status:         "resolved",
		MappingID:      mapping.MappingID,
		AdapterID:      mapping.AdapterID,
		AdapterRef:     mapping.AdapterRef,
		AdapterSHA256:  mapping.ImplementationSHA256,
		RendererID:     mapping.RendererID,
		RendererSHA256: mapping.RendererSHA256,
		Reason:         "trusted validation-only mapping is available",
	}, true, nil
}

func missingRunnerRequirements(runner RunnerRequirements, decision Decision) []string {
	missing := []string{}
	for _, required := range decision.RequiredOS {
		if !containsString(runner.OS, required) {
			missing = append(missing, "required operating system "+required)
		}
	}
	for _, required := range decision.Capabilities {
		if !containsString(runner.Capabilities, required) {
			missing = append(missing, "capability "+required)
		}
	}
	if !containsString(runner.Networks, decision.NetworkScope) {
		missing = append(missing, "network scope "+decision.NetworkScope)
	}
	return missing
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

type upstreamAdapterDescriptor struct {
	SchemaVersion        string `json:"schema_version"`
	AdapterID            string `json:"adapter_id"`
	ImplementationSHA256 string `json:"implementation_sha256"`
	Lifecycle            string `json:"lifecycle"`
	ExecutionMode        string `json:"execution_mode"`
	Supported            struct {
		RequirementsVersions []string `json:"requirements_versions"`
		EvidenceVersions     []string `json:"evidence_versions"`
		OS                   []string `json:"os"`
		Architectures        []string `json:"architectures"`
		Shells               []string `json:"shells"`
		Networks             []string `json:"networks"`
		Capabilities         []string `json:"capabilities"`
	} `json:"supported"`
	SourceBinding     string            `json:"source_binding"`
	FailureSemantics  map[string]string `json:"failure_semantics"`
	Conclusions       []string          `json:"conclusions"`
	DeliveryAuthority string            `json:"delivery_authority"`
}

func validateRunner(runner RunnerRequirements) error {
	for name, values := range map[string][]string{
		"os": runner.OS, "architectures": runner.Architectures, "capabilities": runner.Capabilities, "networks": runner.Networks,
	} {
		if len(values) == 0 || !sort.StringsAreSorted(values) || hasDuplicates(values) {
			return fmt.Errorf("runner %s must be non-empty, sorted, and unique", name)
		}
	}
	return nil
}

func runnerMatchesPlatform(platform string, runner RunnerRequirements) bool {
	want, ok := map[string]RunnerRequirements{
		platformGitHubHosted: {
			OS: []string{"linux"}, Architectures: []string{"amd64"},
			Capabilities: []string{"process-execution"}, Networks: []string{"none", "public"},
		},
		platformGCPManaged: {
			OS: []string{"linux"}, Architectures: []string{"amd64"},
			Capabilities: []string{"container", "process-execution"}, Networks: []string{"gcp-private", "public"},
		},
		platformGitHubSelfHosted: {
			OS: []string{"darwin", "linux", "windows"}, Architectures: []string{"amd64", "arm64"},
			Capabilities: []string{"hardware", "native", "process-execution"}, Networks: []string{"private-vpc", "public"},
		},
	}[platform]
	return ok && equalStrings(runner.OS, want.OS) && equalStrings(runner.Architectures, want.Architectures) &&
		equalStrings(runner.Capabilities, want.Capabilities) && equalStrings(runner.Networks, want.Networks)
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func rendererBytes(name string) ([]byte, error) {
	return rendererAssets.ReadFile(filepath.ToSlash(filepath.Join("templates", name)))
}

// RenderMapping returns a validated, static validation caller. It is intended
// for PR C to consume; this function itself never writes the returned bytes.
func RenderMapping(root, mappingID string) (RenderedMapping, error) {
	catalog, err := LoadAdapterCatalog(root)
	if err != nil {
		return RenderedMapping{}, err
	}
	for _, mapping := range catalog.Mappings {
		if mapping.MappingID == mappingID {
			content, readErr := rendererBytes(mapping.RendererTemplate)
			if readErr != nil {
				return RenderedMapping{}, invalid("ci-cd.catalog.renderer", "read renderer %q: %v", mapping.RendererTemplate, readErr)
			}
			return RenderedMapping{Path: mapping.OutputPath, Mapping: mapping, Content: append([]byte(nil), content...)}, nil
		}
	}
	return RenderedMapping{}, invalid("ci-cd.catalog.mapping", "mapping %q is not published", mappingID)
}

func validateRenderedSemantics(mapping AdapterMapping, content []byte) error {
	var document yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&document); err != nil {
		return invalid("ci-cd.semantic.schema", "renderer %q is not valid YAML: %v", mapping.MappingID, err)
	}
	if document.Kind == 0 {
		return invalid("ci-cd.semantic.schema", "renderer %q is empty", mapping.MappingID)
	}
	lower := strings.ToLower(string(content))
	if containsDownloadAndExecute(lower) {
		return invalid("ci-cd.semantic.download-and-execute", "renderer %q contains download-and-execute behavior", mapping.MappingID)
	}
	if containsDeliveryBehavior(lower) {
		return invalid("ci-cd.semantic.delivery", "renderer %q contains delivery behavior", mapping.MappingID)
	}
	if hasUndeclaredEnvironment(&document, lower) {
		return invalid("ci-cd.semantic.undeclared-secret", "renderer %q declares secrets or an unreviewed environment", mapping.MappingID)
	}
	if hasBroadPermissions(&document) {
		return invalid("ci-cd.semantic.broad-permission", "renderer %q requests a broad or write permission", mapping.MappingID)
	}
	if err := validateSourceReferences(mapping, &document); err != nil {
		return err
	}
	if hasMutableReference(&document) {
		return invalid("ci-cd.semantic.mutable-reference", "renderer %q contains a mutable action or image reference", mapping.MappingID)
	}
	if err := validateCommands(mapping, &document); err != nil {
		return err
	}
	if sha256Hex(content) != mapping.RendererSHA256 {
		return invalid("ci-cd.semantic.stale-template", "renderer %q does not match its trusted template hash", mapping.MappingID)
	}
	return nil
}

func validateCommands(mapping AdapterMapping, document *yaml.Node) error {
	runs := scalarValuesForKey(document, "run")
	entrypoints := scalarValuesForKey(document, "entrypoint")
	args := sequenceValuesForKey(document, "args")
	if mapping.Platform == platformGCPManaged {
		if len(runs) != 0 {
			return invalid("ci-cd.semantic.arbitrary-run", "GCP validation mapping cannot contain arbitrary run commands")
		}
		if len(entrypoints) != 2 || len(args) != 2 {
			return invalid("ci-cd.semantic.profile", "GCP validation mapping must expose exactly ci-quick and full commands")
		}
		for _, entrypoint := range entrypoints {
			if strings.TrimSpace(entrypoint) != "./scripts/verify.sh" {
				return invalid("ci-cd.semantic.arbitrary-run", "renderer contains an unapproved entrypoint")
			}
		}
		if !equalStrings(args[0], fixedPRArgv()[1:]) || !equalStrings(args[1], fixedFullArgv()[1:]) {
			return invalid("ci-cd.semantic.profile", "renderer does not preserve fixed profile argv ordering")
		}
		return nil
	}
	if len(entrypoints) != 0 || len(args) != 0 {
		return invalid("ci-cd.semantic.arbitrary-run", "GitHub validation mapping contains an unapproved command form")
	}
	if len(runs) > 2 {
		return invalid("ci-cd.semantic.arbitrary-run", "GitHub validation mapping contains an extra run command")
	}
	if len(runs) != 2 {
		return invalid("ci-cd.semantic.profile", "GitHub validation mapping must contain two fixed profile commands")
	}
	quick := normalizeCommand(runs[0])
	full := normalizeCommand(runs[1])
	if quick != normalizeCommand("./scripts/verify.sh --profile ci-quick --group '${{ matrix.group }}' --base '${{ github.event.pull_request.base.sha || github.event.before }}' --head '${{ github.event.pull_request.head.sha || github.sha }}'") {
		if strings.Contains(quick, "--profile") {
			return invalid("ci-cd.semantic.profile", "renderer changed the fixed ci-quick argv")
		}
		return invalid("ci-cd.semantic.arbitrary-run", "renderer contains an unapproved run command")
	}
	if full != normalizeCommand("./scripts/verify.sh --profile full") {
		if strings.Contains(full, "--profile") {
			return invalid("ci-cd.semantic.profile", "renderer changed the fixed full argv")
		}
		return invalid("ci-cd.semantic.arbitrary-run", "renderer contains an unapproved run command")
	}
	return nil
}

func normalizeCommand(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func containsDownloadAndExecute(value string) bool {
	for _, token := range []string{"curl ", "wget ", "invoke-webrequest", "git clone", "download", "| sh", "| bash", "bash -c", "sh -c", "go run", "python -c"} {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func containsDeliveryBehavior(value string) bool {
	for _, token := range []string{"deploy", "publish", "release", "terraform apply", "docker push", "gcloud builds submit", "artifactregistry"} {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func hasUndeclaredEnvironment(document *yaml.Node, lower string) bool {
	if strings.Contains(lower, "secrets.") || strings.Contains(lower, "${{ secrets") || strings.Contains(lower, "password") || strings.Contains(lower, "private_key") {
		return true
	}
	return len(mappingValuesForKey(document, "env")) > 0 || len(mappingValuesForKey(document, "secrets")) > 0
}

func hasBroadPermissions(document *yaml.Node) bool {
	for _, value := range mappingValuesForKey(document, "permissions") {
		if value.Kind == yaml.MappingNode {
			for i := 1; i < len(value.Content); i += 2 {
				permission := strings.ToLower(strings.TrimSpace(value.Content[i].Value))
				if permission == "write" || permission == "*" || permission == "read-write" {
					return true
				}
			}
		} else if strings.ToLower(strings.TrimSpace(value.Value)) != "read" && strings.ToLower(strings.TrimSpace(value.Value)) != "none" {
			return true
		}
	}
	return false
}

func hasMutableReference(document *yaml.Node) bool {
	for _, value := range scalarValuesForKey(document, "uses") {
		parts := strings.SplitN(value, "@", 2)
		if len(parts) != 2 || !immutableSHA.MatchString(strings.TrimSpace(parts[1])) {
			return true
		}
	}
	for _, value := range scalarValuesForKey(document, "name") {
		value = strings.TrimSpace(value)
		if strings.Contains(value, "/") && strings.Contains(value, ":") && !strings.Contains(value, "@sha256:") {
			return true
		}
	}
	for _, key := range []string{"image", "provider", "action"} {
		for _, value := range scalarValuesForKey(document, key) {
			if !strings.Contains(value, "@sha256:") {
				return true
			}
		}
	}
	return false
}

func validateSourceReferences(mapping AdapterMapping, document *yaml.Node) error {
	for _, value := range scalarValuesForKey(document, "uses") {
		parts := strings.SplitN(strings.TrimSpace(value), "@", 2)
		if len(parts) != 2 {
			return invalid("ci-cd.semantic.source-mismatch", "renderer %q has an invalid action source", mapping.MappingID)
		}
		if parts[0] != "actions/checkout" {
			return invalid("ci-cd.semantic.source-mismatch", "renderer %q uses an unapproved action source", mapping.MappingID)
		}
		if parts[1] != "3d3c42e5aac5ba805825da76410c181273ba90b1" {
			if immutableSHA.MatchString(parts[1]) {
				return invalid("ci-cd.semantic.source-mismatch", "renderer %q changed its approved action source ref", mapping.MappingID)
			}
		}
	}
	return nil
}

func scalarValuesForKey(document *yaml.Node, key string) []string {
	var result []string
	for _, value := range mappingValuesForKey(document, key) {
		if value.Kind == yaml.ScalarNode {
			result = append(result, value.Value)
		}
	}
	return result
}

func sequenceValuesForKey(document *yaml.Node, key string) [][]string {
	var result [][]string
	for _, value := range mappingValuesForKey(document, key) {
		if value.Kind != yaml.SequenceNode {
			continue
		}
		var values []string
		for _, item := range value.Content {
			if item.Kind != yaml.ScalarNode {
				values = nil
				break
			}
			values = append(values, item.Value)
		}
		if values != nil {
			result = append(result, values)
		}
	}
	return result
}

func mappingValuesForKey(document *yaml.Node, key string) []*yaml.Node {
	var result []*yaml.Node
	var visit func(*yaml.Node)
	visit = func(node *yaml.Node) {
		if node == nil {
			return
		}
		if node.Kind == yaml.DocumentNode {
			for _, child := range node.Content {
				visit(child)
			}
			return
		}
		if node.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(node.Content); i += 2 {
				if node.Content[i].Value == key {
					result = append(result, node.Content[i+1])
				}
				visit(node.Content[i+1])
			}
			return
		}
		for _, child := range node.Content {
			visit(child)
		}
	}
	visit(document)
	return result
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func hasDuplicates(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func mappingForPlatform(catalog AdapterCatalog, platform string) (AdapterMapping, bool) {
	for _, mapping := range catalog.Mappings {
		if mapping.Platform == platform {
			return mapping, true
		}
	}
	return AdapterMapping{}, false
}
