package cicd

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var immutableSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

const (
	platformGitHubHosted     = "github-hosted"
	platformGCPManaged       = "gcp-managed"
	platformGitHubSelfHosted = "github-self-hosted"
	platformUndecided        = "undecided"
)

// BuildPlan reads the portable decision and repository metadata and returns a
// deterministic read-only plan. It does not create directories, write files,
// inspect a runner inventory, or contact a provider.
func BuildPlan(root, configPath string) (Plan, error) {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return Plan{}, invalid("ci-cd.schema.invalid", "read decision configuration: %v", err)
	}
	decision, err := DecodeConfig(configData)
	if err != nil {
		return Plan{}, err
	}
	identity, err := readRepositoryIdentity(root, configData)
	if err != nil {
		return Plan{}, err
	}

	choice := choosePlatform(decision)
	plan := Plan{
		SchemaVersion:  PlanSchemaID,
		Input:          canonicalInput(decision),
		Repository:     identity,
		Mode:           decision.Mode,
		Platform:       choice.platform,
		Installability: false,
		Requirements:   requirements(decision),
		Verification: VerificationPlan{
			PR: ProfileInvocation{
				Profile: decision.Verification.PR,
				Argv:    []string{"scripts/verify.sh", "--profile", "ci-quick", "--group", "<group-id>", "--base", "<base-sha>", "--head", "<head-sha>"},
			},
			Full: ProfileInvocation{
				Profile: decision.Verification.Full,
				Argv:    []string{"scripts/verify.sh", "--profile", "full"},
			},
		},
		AdapterResolution: AdapterResolution{
			Status: "unpublished",
			Reason: "no reviewed ci-cd-adapter-mapping-v1 catalog is present",
		},
		Reasons:              append([]string{}, choice.reasons...),
		RejectedAlternatives: append([]RejectedAlternative{}, choice.rejected...),
		MissingInputs:        append([]string{}, choice.missing...),
	}
	if choice.platform == platformUndecided {
		plan.AdapterResolution = AdapterResolution{
			Status: "undecided",
			Reason: "requirements do not match an approved validation mapping",
		}
	} else {
		plan.Reasons = append(plan.Reasons, "selected platform has no trusted adapter mapping in this release")
	}
	sort.Strings(plan.Reasons)
	sort.Strings(plan.MissingInputs)
	sort.Slice(plan.RejectedAlternatives, func(i, j int) bool {
		if plan.RejectedAlternatives[i].Platform != plan.RejectedAlternatives[j].Platform {
			return plan.RejectedAlternatives[i].Platform < plan.RejectedAlternatives[j].Platform
		}
		return plan.RejectedAlternatives[i].Reason < plan.RejectedAlternatives[j].Reason
	})
	return withDigest(plan)
}

type platformChoice struct {
	platform string
	reasons  []string
	rejected []RejectedAlternative
	missing  []string
}

func choosePlatform(decision Decision) platformChoice {
	choice := platformChoice{platform: platformUndecided, reasons: []string{}, rejected: []RejectedAlternative{}, missing: []string{}}
	github := decision.SourceHost == "github" || decision.SourceHost == "github.com"
	linuxOnly := len(decision.RequiredOS) == 1 && decision.RequiredOS[0] == "linux"
	specialCapability := hasSpecialCapability(decision.Capabilities)
	signingCapability := hasSigningCapability(decision.Capabilities)

	if signingCapability {
		choice.reasons = append(choice.reasons, "signing or notarization credentials require an unavailable approved credential boundary")
		choice.missing = append(choice.missing, "approved signing credential")
		return rejectAll(choice, "credential-bound validation is not installable")
	}
	if decision.NetworkScope == "isolated" || decision.NetworkScope == "on-premises" {
		choice.reasons = append(choice.reasons, "isolated or on-premises validation has no approved portable mapping")
		choice.missing = append(choice.missing, "approved isolated validation mapping")
		return rejectAll(choice, "isolated and on-premises execution is not mapped")
	}
	if decision.CloudAuthority == "aws" || decision.CloudAuthority == "azure" {
		choice.reasons = append(choice.reasons, "AWS and Azure authority are outside the reviewed validation mapping")
		return rejectAll(choice, "native AWS and Azure validation is not mapped")
	}
	if !github {
		choice.reasons = append(choice.reasons, "source_host is not GitHub and no reviewed source collaboration mapping is available")
		return rejectAll(choice, "only GitHub source collaboration is mapped")
	}

	if decision.CloudAuthority == "gcp" || decision.NetworkScope == "gcp-private" {
		if !linuxOnly {
			choice.reasons = append(choice.reasons, "GCP managed validation requires Linux-only requirements")
			choice.missing = append(choice.missing, "Linux-only required_os for gcp-managed validation")
			return rejectAll(choice, "required operating systems do not match gcp-managed")
		}
		if decision.OperationsOwner != "declared" {
			choice.reasons = append(choice.reasons, "GCP managed validation requires a declared operations owner")
			choice.missing = append(choice.missing, "declared operations owner")
			return rejectAll(choice, "operations owner is absent for gcp-managed validation")
		}
		choice.platform = platformGCPManaged
		choice.reasons = append(choice.reasons, "GCP authority or gcp-private scope selects managed GCP validation")
		choice.rejected = append(choice.rejected,
			RejectedAlternative{Platform: platformGitHubHosted, Reason: "GitHub-hosted validation does not own the declared GCP private authority"},
			RejectedAlternative{Platform: platformGitHubSelfHosted, Reason: "self-hosted validation is not the authoritative GCP mapping"},
		)
		return choice
	}

	if specialCapability || !linuxOnly || decision.NetworkScope == "private-vpc" {
		if decision.OperationsOwner != "declared" {
			choice.reasons = append(choice.reasons, "specialized or private validation requires a declared operations owner")
			choice.missing = append(choice.missing, "declared operations owner")
			return rejectAll(choice, "operations owner is absent for github-self-hosted validation")
		}
		choice.platform = platformGitHubSelfHosted
		choice.reasons = append(choice.reasons, "GitHub UX with named specialized validation ownership selects self-hosted validation")
		choice.rejected = append(choice.rejected,
			RejectedAlternative{Platform: platformGitHubHosted, Reason: "hosted runners do not satisfy the specialized OS, capability, or private-network requirement"},
			RejectedAlternative{Platform: platformGCPManaged, Reason: "requirements do not declare GCP authority"},
		)
		return choice
	}

	if decision.NetworkScope == "public" || decision.NetworkScope == "none" {
		choice.platform = platformGitHubHosted
		choice.reasons = append(choice.reasons, "standard GitHub public validation selects GitHub-hosted validation")
		choice.rejected = append(choice.rejected,
			RejectedAlternative{Platform: platformGCPManaged, Reason: "requirements do not declare GCP authority or gcp-private scope"},
			RejectedAlternative{Platform: platformGitHubSelfHosted, Reason: "standard public Linux validation has no specialized owner requirement"},
		)
		return choice
	}

	choice.reasons = append(choice.reasons, "network scope has no approved validation mapping")
	return rejectAll(choice, "network scope is not mapped")
}

func rejectAll(choice platformChoice, reason string) platformChoice {
	choice.platform = platformUndecided
	choice.rejected = append(choice.rejected,
		RejectedAlternative{Platform: platformGitHubHosted, Reason: reason},
		RejectedAlternative{Platform: platformGCPManaged, Reason: reason},
		RejectedAlternative{Platform: platformGitHubSelfHosted, Reason: reason},
	)
	return choice
}

func hasSpecialCapability(values []string) bool {
	for _, value := range values {
		switch value {
		case "darwin", "desktop", "gpu", "hardware", "macos", "mobile", "native", "specialized":
			return true
		}
	}
	return false
}

func hasSigningCapability(values []string) bool {
	for _, value := range values {
		if strings.Contains(value, "credential") || strings.Contains(value, "signing") || strings.Contains(value, "notarization") || value == "secret" || value == "signing-key" {
			return true
		}
	}
	return false
}

func readRepositoryIdentity(root string, configData []byte) (RepositoryIdentity, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return RepositoryIdentity{}, invalid("ci-cd.repository.invalid", "resolve repository root: %v", err)
	}
	info, err := os.Stat(absoluteRoot)
	if err != nil || !info.IsDir() {
		if err == nil {
			err = errors.New("repository root is not a directory")
		}
		return RepositoryIdentity{}, invalid("ci-cd.repository.invalid", "repository root is unavailable")
	}
	head, err := gitValue(absoluteRoot, "rev-parse", "--verify", "HEAD")
	if err != nil || !immutableSHA.MatchString(head) {
		return RepositoryIdentity{}, invalid("ci-cd.repository.invalid", "repository HEAD is not an immutable commit")
	}
	tree, err := gitValue(absoluteRoot, "rev-parse", "--verify", "HEAD^{tree}")
	if err != nil || !immutableSHA.MatchString(tree) {
		return RepositoryIdentity{}, invalid("ci-cd.repository.invalid", "repository tree is not an immutable identity")
	}
	remote, err := gitValue(absoluteRoot, "config", "--get", "remote.origin.url")
	if err != nil {
		return RepositoryIdentity{}, invalid("ci-cd.repository.invalid", "repository origin host is unavailable")
	}
	host := remoteHost(remote)
	if host == "" {
		return RepositoryIdentity{}, invalid("ci-cd.repository.invalid", "repository origin host is not portable")
	}
	profile, err := os.ReadFile(filepath.Join(absoluteRoot, filepath.FromSlash(ProfileFile)))
	if err != nil {
		return RepositoryIdentity{}, invalid("ci-cd.repository.invalid", "verification profile catalog is unavailable")
	}
	catalog, err := os.ReadFile(filepath.Join(absoluteRoot, filepath.FromSlash(CoreCatalogFile)))
	if err != nil {
		return RepositoryIdentity{}, invalid("ci-cd.repository.invalid", "core catalog is unavailable")
	}
	return RepositoryIdentity{
		RemoteHost:    host,
		HeadSHA:       head,
		TreeSHA:       tree,
		ConfigSHA256:  sha256Hex(configData),
		ProfileSHA256: sha256Hex(profile),
		CatalogSHA256: sha256Hex(catalog),
	}, nil
}

func gitValue(root string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", root}, args...)
	output, err := exec.Command("git", commandArgs...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func remoteHost(remote string) string {
	remote = strings.TrimSpace(remote)
	if strings.HasPrefix(remote, "git@") {
		value := strings.TrimPrefix(remote, "git@")
		if index := strings.IndexByte(value, ':'); index >= 0 {
			return strings.ToLower(value[:index])
		}
	}
	parsed, err := url.Parse(remote)
	if err != nil {
		return ""
	}
	host := parsed.Hostname()
	return strings.ToLower(strings.TrimPrefix(host, "www."))
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// HumanPlan renders the same canonical model used by JSON output.
func HumanPlan(plan Plan) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Soku CI/CD plan: %s\nPlatform: %s\nInstallability: %t\nMode: %s\n",
		plan.SchemaVersion, plan.Platform, plan.Installability, plan.Mode)
	fmt.Fprintf(&builder, "Repository host: %s\nHEAD: %s\nTree: %s\n",
		plan.Repository.RemoteHost, plan.Repository.HeadSHA, plan.Repository.TreeSHA)
	fmt.Fprintf(&builder, "Config SHA-256: %s\nProfile SHA-256: %s\nCatalog SHA-256: %s\n",
		plan.Repository.ConfigSHA256, plan.Repository.ProfileSHA256, plan.Repository.CatalogSHA256)
	fmt.Fprintf(&builder, "PR profile: %s (%s)\nFull profile: %s (%s)\n",
		plan.Verification.PR.Profile, strings.Join(plan.Verification.PR.Argv, " "),
		plan.Verification.Full.Profile, strings.Join(plan.Verification.Full.Argv, " "))
	fmt.Fprintf(&builder, "Adapter: %s\nPlan digest: %s\n", plan.AdapterResolution.Status, plan.PlanDigest)
	if len(plan.Reasons) > 0 {
		builder.WriteString("Reasons:\n")
		for _, reason := range plan.Reasons {
			fmt.Fprintf(&builder, "  - %s\n", reason)
		}
	}
	if len(plan.MissingInputs) > 0 {
		builder.WriteString("Missing inputs:\n")
		for _, missing := range plan.MissingInputs {
			fmt.Fprintf(&builder, "  - %s\n", missing)
		}
	}
	return builder.String()
}
