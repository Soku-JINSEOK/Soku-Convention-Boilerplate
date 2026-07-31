package manual

import (
	"fmt"
	"sort"
	"strings"
)

// CapturePlan is one stable manual-section-to-asset relation.
type CapturePlan struct {
	ScenarioID   string `json:"scenario_id"`
	StepIndex    int    `json:"step_index"`
	CaptureID    string `json:"capture_id"`
	OutputPath   string `json:"output_path"`
	Mode         string `json:"mode"`
	Caption      string `json:"caption"`
	Authenticity string `json:"authenticity"`
}

// PlanReport is the deterministic read-only planning result.
type PlanReport struct {
	State             string        `json:"state"`
	SchemaVersion     int           `json:"schema_version"`
	ConfigurationPath string        `json:"configuration_path"`
	ConfigurationHash string        `json:"configuration_hash"`
	RuntimeAdapter    string        `json:"runtime_adapter"`
	BackendMode       string        `json:"backend_mode"`
	MapProvider       string        `json:"map_provider"`
	ExecutionMode     string        `json:"execution_mode"`
	Authenticity      string        `json:"authenticity"`
	AllowHosts        []string      `json:"allow_hosts"`
	Captures          []CapturePlan `json:"captures"`
}

// BuildPlan loads and deterministically renders a capture plan.
func BuildPlan(root, configPath string) (PlanReport, error) {
	loaded, err := LoadConfig(root, configPath)
	if err != nil {
		return PlanReport{}, err
	}
	authenticity := "runtime-authentic"
	if loaded.Config.Backend.Mode != "real-local" && loaded.Config.Backend.Mode != "none" ||
		loaded.Config.Map.SourceRelation == "declared-adapter" ||
		loaded.Config.Runtime.Adapter == "gas-html-service" {
		authenticity = "runtime-authentic-with-adapters"
	}
	report := PlanReport{
		State: "planned", SchemaVersion: 1,
		ConfigurationPath: loaded.Path, ConfigurationHash: loaded.Hash,
		RuntimeAdapter: loaded.Config.Runtime.Adapter, BackendMode: loaded.Config.Backend.Mode,
		MapProvider: loaded.Config.Map.Provider, ExecutionMode: loaded.Config.Execution.Mode,
		Authenticity: authenticity,
		AllowHosts:   append([]string{}, loaded.Config.Execution.AllowHosts...),
		Captures:     []CapturePlan{},
	}
	for _, scenario := range loaded.Config.Scenarios {
		for index, step := range scenario.Steps {
			if step.Action != "capture" {
				continue
			}
			report.Captures = append(report.Captures, CapturePlan{
				ScenarioID: scenario.ID, StepIndex: index, CaptureID: step.ID,
				OutputPath: strings.TrimSuffix(loaded.Config.Output.Directory, "/") + "/" + step.ID + ".png",
				Mode:       step.Capture.Mode, Caption: step.Capture.Caption, Authenticity: authenticity,
			})
		}
	}
	sort.Slice(report.Captures, func(i, j int) bool {
		if report.Captures[i].ScenarioID != report.Captures[j].ScenarioID {
			return report.Captures[i].ScenarioID < report.Captures[j].ScenarioID
		}
		return report.Captures[i].StepIndex < report.Captures[j].StepIndex
	})
	return report, nil
}

// HumanPlan renders stable review-oriented text.
func HumanPlan(report PlanReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Soku docs manual plan: %s\nConfig: %s (%s)\nRuntime: %s\nBackend: %s\nMap: %s\nAuthenticity: %s\n",
		report.State, report.ConfigurationPath, report.ConfigurationHash, report.RuntimeAdapter,
		report.BackendMode, report.MapProvider, report.Authenticity)
	builder.WriteString("Captures:\n")
	for _, capture := range report.Captures {
		fmt.Fprintf(&builder, "  %s -> %s (%s)\n", capture.CaptureID, capture.OutputPath, capture.Mode)
	}
	return builder.String()
}
