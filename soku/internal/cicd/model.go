// Package cicd implements the portable CI/CD decision contract.
package cicd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

const (
	DecisionSchemaVersion = 1
	DecisionSchemaID      = "ci-cd-decision-v1"
	PlanSchemaID          = "ci-cd-plan-v1"
	ProfileFile           = "verification/profiles.yml"
	CoreCatalogFile       = "soku/catalog/index-v2.json"
	AdapterCatalogFile    = "soku/catalog/ci-cd-adapter-mapping-v1.json"
)

// Error is a stable, user-facing planning error. Code is intentionally
// independent from Go error text so callers can safely automate it.
type Error struct {
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Cause }

func invalid(code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Decision is the strict, portable input contract accepted by plan.
type Decision struct {
	SchemaVersion   int          `yaml:"schema_version" json:"schema_version"`
	Mode            string       `yaml:"mode" json:"mode"`
	SourceHost      string       `yaml:"source_host" json:"source_host"`
	Workload        string       `yaml:"workload" json:"workload"`
	Artifact        string       `yaml:"artifact" json:"artifact"`
	RequiredOS      []string     `yaml:"required_os" json:"required_os"`
	NetworkScope    string       `yaml:"network_scope" json:"network_scope"`
	CloudAuthority  string       `yaml:"cloud_authority" json:"cloud_authority"`
	OperationsOwner string       `yaml:"operations_owner" json:"operations_owner"`
	Capabilities    []string     `yaml:"capabilities" json:"capabilities"`
	Verification    Verification `yaml:"verification" json:"verification"`
	Delivery        Delivery     `yaml:"delivery" json:"delivery"`
}

type Verification struct {
	PR   string `yaml:"pr" json:"pr"`
	Full string `yaml:"full" json:"full"`
}

type Delivery struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// Plan is the canonical read-only result. It deliberately contains no
// repository name, absolute path, timestamp, environment inventory, or raw
// configuration.
type Plan struct {
	SchemaVersion        string                `json:"schema_version"`
	Input                InputSummary          `json:"input"`
	Repository           RepositoryIdentity    `json:"repository"`
	Mode                 string                `json:"mode"`
	Platform             string                `json:"platform"`
	Installability       bool                  `json:"installability"`
	Requirements         RequirementsSummary   `json:"requirements"`
	Verification         VerificationPlan      `json:"verification"`
	AdapterResolution    AdapterResolution     `json:"adapter_resolution"`
	Reasons              []string              `json:"reasons"`
	RejectedAlternatives []RejectedAlternative `json:"rejected_alternatives"`
	MissingInputs        []string              `json:"missing_inputs"`
	PlanDigest           string                `json:"plan_digest"`
}

type InputSummary struct {
	SchemaVersion   int          `json:"schema_version"`
	Mode            string       `json:"mode"`
	SourceHost      string       `json:"source_host"`
	Workload        string       `json:"workload"`
	Artifact        string       `json:"artifact"`
	RequiredOS      []string     `json:"required_os"`
	NetworkScope    string       `json:"network_scope"`
	CloudAuthority  string       `json:"cloud_authority"`
	OperationsOwner string       `json:"operations_owner"`
	Capabilities    []string     `json:"capabilities"`
	Verification    Verification `json:"verification"`
	Delivery        Delivery     `json:"delivery"`
}

type RepositoryIdentity struct {
	RemoteHost           string `json:"remote_host"`
	HeadSHA              string `json:"head_sha"`
	TreeSHA              string `json:"tree_sha"`
	ConfigSHA256         string `json:"config_sha256"`
	ProfileSHA256        string `json:"profile_sha256"`
	CatalogSHA256        string `json:"catalog_sha256"`
	AdapterCatalogSHA256 string `json:"adapter_catalog_sha256"`
}

type RequirementsSummary struct {
	Workload        string   `json:"workload"`
	Artifact        string   `json:"artifact"`
	RequiredOS      []string `json:"required_os"`
	NetworkScope    string   `json:"network_scope"`
	CloudAuthority  string   `json:"cloud_authority"`
	OperationsOwner string   `json:"operations_owner"`
	Capabilities    []string `json:"capabilities"`
}

type VerificationPlan struct {
	PR   ProfileInvocation `json:"pr"`
	Full ProfileInvocation `json:"full"`
}

type ProfileInvocation struct {
	Profile string   `json:"profile"`
	Argv    []string `json:"argv"`
}

type AdapterResolution struct {
	Status         string `json:"status"`
	MappingID      string `json:"mapping_id,omitempty"`
	AdapterID      string `json:"adapter_id,omitempty"`
	AdapterRef     string `json:"adapter_ref,omitempty"`
	AdapterSHA256  string `json:"adapter_sha256,omitempty"`
	RendererID     string `json:"renderer_id,omitempty"`
	RendererSHA256 string `json:"renderer_sha256,omitempty"`
	Reason         string `json:"reason"`
}

type RejectedAlternative struct {
	Platform string `json:"platform"`
	Reason   string `json:"reason"`
}

// CanonicalJSON returns the byte-stable JSON form used for plan_digest. The
// digest field is included as an empty value when calculating the digest and
// populated only in the returned plan.
func CanonicalJSON(plan Plan) ([]byte, error) {
	return json.Marshal(plan)
}

func withDigest(plan Plan) (Plan, error) {
	plan.PlanDigest = ""
	data, err := CanonicalJSON(plan)
	if err != nil {
		return Plan{}, err
	}
	digest := sha256.Sum256(data)
	plan.PlanDigest = hex.EncodeToString(digest[:])
	return plan, nil
}

func canonicalInput(decision Decision) InputSummary {
	requiredOS := append([]string{}, decision.RequiredOS...)
	sort.Strings(requiredOS)
	capabilities := append([]string{}, decision.Capabilities...)
	return InputSummary{
		SchemaVersion:   decision.SchemaVersion,
		Mode:            decision.Mode,
		SourceHost:      decision.SourceHost,
		Workload:        decision.Workload,
		Artifact:        decision.Artifact,
		RequiredOS:      requiredOS,
		NetworkScope:    decision.NetworkScope,
		CloudAuthority:  decision.CloudAuthority,
		OperationsOwner: decision.OperationsOwner,
		Capabilities:    capabilities,
		Verification:    decision.Verification,
		Delivery:        decision.Delivery,
	}
}

func requirements(decision Decision) RequirementsSummary {
	requiredOS := append([]string{}, decision.RequiredOS...)
	sort.Strings(requiredOS)
	capabilities := append([]string{}, decision.Capabilities...)
	return RequirementsSummary{
		Workload:        decision.Workload,
		Artifact:        decision.Artifact,
		RequiredOS:      requiredOS,
		NetworkScope:    decision.NetworkScope,
		CloudAuthority:  decision.CloudAuthority,
		OperationsOwner: decision.OperationsOwner,
		Capabilities:    capabilities,
	}
}
