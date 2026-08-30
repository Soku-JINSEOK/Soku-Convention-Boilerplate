package cicd

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,63}$`)

var (
	allowedWorkloads = map[string]bool{
		"api": true, "container-service": true, "desktop-app": true,
		"iac": true, "library": true, "mobile-app": true,
		"static-site": true, "template": true,
	}
	allowedArtifacts = map[string]bool{
		"archive": true, "container-digest": true, "installer": true,
		"none": true, "package": true,
	}
	allowedHosts = map[string]bool{
		"github": true, "github.com": true, "gitlab": true, "gitlab.com": true,
	}
	allowedNetworks = map[string]bool{
		"gcp-private": true, "isolated": true, "none": true,
		"on-premises": true, "private-vpc": true, "public": true,
	}
	allowedClouds = map[string]bool{"aws": true, "azure": true, "gcp": true, "none": true}
	allowedOS     = map[string]bool{"darwin": true, "linux": true, "windows": true}
)

// LoadConfig reads and strictly validates a decision document. It never writes
// the source file or any derived state.
func LoadConfig(path string) (Decision, error) {
	if strings.TrimSpace(path) == "" {
		return Decision{}, invalid("ci-cd.schema.invalid", "decision configuration path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Decision{}, invalid("ci-cd.schema.invalid", "read decision configuration: %v", err)
	}
	return DecodeConfig(data)
}

// DecodeConfig validates one YAML document against the strict v1 input shape.
func DecodeConfig(data []byte) (Decision, error) {
	if len(data) == 0 {
		return Decision{}, invalid("ci-cd.schema.invalid", "decision configuration is empty")
	}
	if err := validateDocumentShape(data); err != nil {
		return Decision{}, err
	}
	var decision Decision
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&decision); err != nil {
		return Decision{}, invalid("ci-cd.schema.invalid", "decode decision configuration: %v", err)
	}
	if err := validateDecision(decision); err != nil {
		return Decision{}, err
	}
	return decision, nil
}

func validateDocumentShape(data []byte) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return invalid("ci-cd.schema.invalid", "decode decision configuration: %v", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return invalid("ci-cd.schema.invalid", "decision configuration must contain exactly one YAML document")
		}
		return invalid("ci-cd.schema.invalid", "decode decision configuration: %v", err)
	}
	if len(document.Content) != 1 {
		return invalid("ci-cd.schema.invalid", "decision configuration must contain one YAML document")
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return invalid("ci-cd.schema.invalid", "decision configuration root must be a mapping")
	}
	if err := validateMapping(root, []string{"schema_version", "mode", "source_host", "workload", "artifact", "required_os", "network_scope", "cloud_authority", "operations_owner", "capabilities", "verification", "delivery"}, "root"); err != nil {
		return err
	}
	verification, ok := mappingValue(root, "verification")
	if !ok || verification.Kind != yaml.MappingNode {
		return invalid("ci-cd.schema.invalid", "verification must be a mapping")
	}
	if err := validateMapping(verification, []string{"pr", "full"}, "verification"); err != nil {
		return err
	}
	delivery, ok := mappingValue(root, "delivery")
	if !ok || delivery.Kind != yaml.MappingNode {
		return invalid("ci-cd.schema.invalid", "delivery must be a mapping")
	}
	if err := validateMapping(delivery, []string{"enabled"}, "delivery"); err != nil {
		return err
	}
	return nil
}

func validateMapping(node *yaml.Node, expected []string, name string) error {
	allowed := make(map[string]bool, len(expected))
	for _, key := range expected {
		allowed[key] = true
	}
	seen := make(map[string]bool, len(node.Content)/2)
	for index := 0; index < len(node.Content); index += 2 {
		key := node.Content[index]
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || !allowed[key.Value] {
			return invalid("ci-cd.schema.invalid", "%s contains an unsupported field %q", name, key.Value)
		}
		if seen[key.Value] {
			return invalid("ci-cd.schema.invalid", "%s contains a duplicate field %q", name, key.Value)
		}
		seen[key.Value] = true
	}
	for _, key := range expected {
		if !seen[key] {
			return invalid("ci-cd.schema.invalid", "%s is missing required field %q", name, key)
		}
	}
	return nil
}

func mappingValue(node *yaml.Node, wanted string) (*yaml.Node, bool) {
	for index := 0; index < len(node.Content); index += 2 {
		if node.Content[index].Value == wanted {
			return node.Content[index+1], true
		}
	}
	return nil, false
}

func validateDecision(decision Decision) error {
	if decision.SchemaVersion != DecisionSchemaVersion {
		return invalid("ci-cd.input.invalid", "schema_version must be %d", DecisionSchemaVersion)
	}
	if decision.Mode != "ci-only" {
		return invalid("ci-cd.input.invalid", "mode must be ci-only")
	}
	if !allowedHosts[decision.SourceHost] || strings.ToLower(decision.SourceHost) != decision.SourceHost {
		return invalid("ci-cd.input.invalid", "source_host is not a supported portable host")
	}
	if !allowedWorkloads[decision.Workload] {
		return invalid("ci-cd.input.invalid", "workload %q is not supported", decision.Workload)
	}
	if !allowedArtifacts[decision.Artifact] {
		return invalid("ci-cd.input.invalid", "artifact %q is not supported", decision.Artifact)
	}
	if len(decision.RequiredOS) == 0 {
		return invalid("ci-cd.input.invalid", "required_os must contain at least one operating system")
	}
	if err := validateSortedUniqueIDs(decision.RequiredOS, allowedOS, "required_os", false); err != nil {
		return err
	}
	if !allowedNetworks[decision.NetworkScope] {
		return invalid("ci-cd.input.invalid", "network_scope %q is not supported", decision.NetworkScope)
	}
	if !allowedClouds[decision.CloudAuthority] {
		return invalid("ci-cd.input.invalid", "cloud_authority %q is not supported", decision.CloudAuthority)
	}
	if decision.OperationsOwner != "declared" && decision.OperationsOwner != "absent" {
		return invalid("ci-cd.input.invalid", "operations_owner must be declared or absent")
	}
	if err := validateSortedUniqueIDs(decision.Capabilities, nil, "capabilities", true); err != nil {
		return err
	}
	if decision.Verification.PR != "ci-quick" || decision.Verification.Full != "full" {
		return invalid("ci-cd.input.invalid", "verification must bind pr=ci-quick and full=full")
	}
	if decision.Delivery.Enabled {
		return invalid("ci-cd.input.invalid", "delivery.enabled must be false")
	}
	return nil
}

func validateSortedUniqueIDs(values []string, allowed map[string]bool, name string, requireSorted bool) error {
	seen := make(map[string]bool, len(values))
	previous := ""
	for _, value := range values {
		if !identifierPattern.MatchString(value) {
			return invalid("ci-cd.input.invalid", "%s contains invalid identifier %q", name, value)
		}
		if allowed != nil && !allowed[value] {
			return invalid("ci-cd.input.invalid", "%s contains unsupported value %q", name, value)
		}
		if seen[value] {
			return invalid("ci-cd.input.invalid", "%s contains duplicate value %q", name, value)
		}
		if requireSorted && previous != "" && value < previous {
			return invalid("ci-cd.input.invalid", "%s must be sorted", name)
		}
		seen[value] = true
		previous = value
	}
	return nil
}
