package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/cicd"
	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/initcmd"
	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manual"
	lifecyclestatus "github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/status"
)

// Request contains the stable invocation state passed to lifecycle behavior.
type Request struct {
	Command                  string
	ConfigPath               string
	JSON                     bool
	Quiet                    bool
	NonInteractive           bool
	DryRun                   bool
	Yes                      bool
	Interactive              bool
	BoilerplateSource        string
	BoilerplateRelease       string
	Stacks                   []string
	Profile                  string
	ProjectName              string
	ModulePath               string
	JavaGroup                string
	ServiceName              string
	IntegrationSource        string
	IntegrationRef           string
	IntegrationConfig        string
	Verify                   bool
	ProjectSync              bool
	ProjectSyncProjectNumber int
	Path                     string
	ExpectedSHA256           string
	Probe                    bool
	SourceSet                bool
	ReleaseSet               bool
	StacksSet                bool
	ProfileSet               bool
	ProjectNameSet           bool
	ModulePathSet            bool
	JavaGroupSet             bool
	ServiceNameSet           bool
	VerifySet                bool
	Input                    io.Reader
	PromptOutput             io.Writer
	SokuVersion              string
}

// Result contains successful command output, including diagnostic non-zero exits.
type Result struct {
	Human string
	Data  any
	Code  ExitCode
}

// Handler is replaced by later lifecycle issues without changing CLI parsing.
type Handler interface {
	Handle(context.Context, Request) (Result, error)
}

// HandlerFunc adapts the legacy error-only boundary to Handler.
type HandlerFunc func(context.Context, Request) error

func (f HandlerFunc) Handle(ctx context.Context, request Request) (Result, error) {
	return Result{}, f(ctx, request)
}

// ResultHandlerFunc adapts a result-producing function to Handler.
type ResultHandlerFunc func(context.Context, Request) (Result, error)

func (f ResultHandlerFunc) Handle(ctx context.Context, request Request) (Result, error) {
	return f(ctx, request)
}

// Handlers contains one independently replaceable lifecycle boundary.
type Handlers struct {
	Init             Handler
	Status           Handler
	Diff             Handler
	Upgrade          Handler
	CICDPlan         Handler
	CICDInit         Handler
	ManualPlan       Handler
	ManualDoctor     Handler
	ManualInit       Handler
	OwnershipHandoff Handler
}

func defaultHandlers() Handlers {
	return Handlers{
		Init:             initHandler(),
		Status:           statusHandler(),
		Diff:             transitionHandler(false),
		Upgrade:          transitionHandler(true),
		CICDPlan:         cicdPlanHandler(),
		CICDInit:         cicdInitHandler(),
		ManualPlan:       manualPlanHandler(),
		ManualDoctor:     manualDoctorHandler(),
		ManualInit:       manualInitHandler(),
		OwnershipHandoff: ownershipHandoffHandler(),
	}
}

func cicdPlanHandler() Handler {
	return ResultHandlerFunc(func(_ context.Context, request Request) (Result, error) {
		root, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		plan, err := cicd.BuildPlan(root, request.ConfigPath)
		if err != nil {
			var planningError *cicd.Error
			if errors.As(err, &planningError) {
				return Result{}, &ExitError{
					Code: ExitValidationFailure, Key: planningError.Code,
					Message: planningError.Message, Cause: planningError,
				}
			}
			return Result{}, err
		}
		return Result{Human: cicd.HumanPlan(plan), Data: plan, Code: ExitSuccess}, nil
	})
}

func cicdInitHandler() Handler {
	return HandlerFunc(func(_ context.Context, _ Request) error {
		return &ExitError{
			Code:    ExitSafetyRefusal,
			Key:     "ci-cd.init.unavailable",
			Message: "soku ci-cd init is not available in this release",
		}
	})
}

func ownershipHandoffHandler() Handler {
	return ResultHandlerFunc(func(_ context.Context, request Request) (Result, error) {
		root, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		confirm := func(report initcmd.HandoffReport) (bool, error) {
			if request.PromptOutput == nil || request.Input == nil {
				return false, fmt.Errorf("interactive streams are unavailable")
			}
			if _, err := fmt.Fprint(request.PromptOutput, initcmd.HumanHandoff(report)+"Apply this handoff? [y/N] "); err != nil {
				return false, err
			}
			var answer string
			_, err := fmt.Fscanln(request.Input, &answer)
			if err != nil && err != io.EOF {
				return false, err
			}
			answer = strings.ToLower(strings.TrimSpace(answer))
			return answer == "y" || answer == "yes", nil
		}
		report, err := initcmd.HandoffOwnership(initcmd.HandoffOptions{
			Root: root, Path: request.Path, ExpectedSHA256: request.ExpectedSHA256,
			DryRun: request.DryRun, Yes: request.Yes, Interactive: request.Interactive,
			Confirm: confirm, SokuVersion: request.SokuVersion,
		})
		if err != nil {
			var failure *initcmd.Failure
			if errors.As(err, &failure) {
				return Result{}, &ExitError{Code: ExitCode(failure.Code), Key: failure.Key, Message: failure.Message, Cause: failure, Data: failure.Data}
			}
			return Result{}, err
		}
		return Result{Human: initcmd.HumanHandoff(report), Data: report, Code: ExitSuccess}, nil
	})
}

func manualPlanHandler() Handler {
	return ResultHandlerFunc(func(_ context.Context, request Request) (Result, error) {
		root, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		report, err := manual.BuildPlan(root, request.ConfigPath)
		if err != nil {
			return Result{}, manualExitError(err)
		}
		return Result{Human: manual.HumanPlan(report), Data: report, Code: ExitSuccess}, nil
	})
}

func manualDoctorHandler() Handler {
	return ResultHandlerFunc(func(ctx context.Context, request Request) (Result, error) {
		root, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		report, err := manual.Doctor(ctx, root, request.ConfigPath, request.Probe)
		if err != nil {
			return Result{}, manualExitError(err)
		}
		code := ExitSuccess
		if report.State != "ready" {
			code = ExitChangesFound
		}
		return Result{Human: manual.HumanDoctor(report), Data: report, Code: code}, nil
	})
}

func manualInitHandler() Handler {
	return ResultHandlerFunc(func(_ context.Context, request Request) (Result, error) {
		root, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		report, err := manual.Init(manual.InitOptions{
			Root: root, ConfigPath: request.ConfigPath, DryRun: request.DryRun,
			Yes: request.Yes, SokuVersion: request.SokuVersion,
		})
		if err != nil {
			return Result{}, manualExitError(err)
		}
		return Result{Human: manual.HumanInit(report), Data: report, Code: ExitSuccess}, nil
	})
}

func manualExitError(err error) error {
	var manualError *manual.Error
	if errors.As(err, &manualError) {
		return &ExitError{
			Code: ExitCode(manualError.Code), Key: manualError.Key,
			Message: manualError.Message, Cause: manualError,
		}
	}
	return err
}

func initHandler() Handler {
	return ResultHandlerFunc(func(ctx context.Context, request Request) (Result, error) {
		root, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		confirm := func(report initcmd.Report) (bool, error) {
			if request.PromptOutput == nil || request.Input == nil {
				return false, fmt.Errorf("interactive streams are unavailable")
			}
			if _, err := fmt.Fprint(request.PromptOutput, initcmd.Human(report)+"Apply this plan? [y/N] "); err != nil {
				return false, err
			}
			var answer string
			_, err := fmt.Fscanln(request.Input, &answer)
			if err != nil && err != io.EOF {
				return false, err
			}
			answer = strings.ToLower(strings.TrimSpace(answer))
			return answer == "y" || answer == "yes", nil
		}
		projectNumber := request.ProjectSyncProjectNumber
		if request.ProjectSync && projectNumber < 1 && request.Interactive && !request.Yes && !request.NonInteractive {
			if request.PromptOutput == nil || request.Input == nil {
				return Result{}, &ExitError{Code: ExitValidationFailure, Key: "project-sync.project-number.required", Message: "--project-sync-project-number is required when interactive input is unavailable"}
			}
			if _, err := fmt.Fprint(request.PromptOutput, "GitHub Project number (positive integer): "); err != nil {
				return Result{}, err
			}
			var raw string
			if _, err := fmt.Fscanln(request.Input, &raw); err != nil && err != io.EOF {
				return Result{}, &ExitError{Code: ExitValidationFailure, Key: "project-sync.project-number.invalid", Message: "read Project number: " + err.Error(), Cause: err}
			}
			parsed, parseErr := strconv.Atoi(strings.TrimSpace(raw))
			if parseErr != nil || parsed < 1 {
				return Result{}, &ExitError{Code: ExitValidationFailure, Key: "project-sync.project-number.invalid", Message: "GitHub Project number must be a positive integer", Cause: parseErr}
			}
			projectNumber = parsed
		}
		report, err := initcmd.Run(ctx, initcmd.Options{Root: root, ConfigPath: request.ConfigPath, Explicit: initcmd.Explicit{Source: request.BoilerplateSource, Release: request.BoilerplateRelease, Stacks: request.Stacks, Profile: request.Profile, ProjectName: request.ProjectName, ModulePath: request.ModulePath, JavaGroup: request.JavaGroup, ServiceName: request.ServiceName, Verify: request.Verify, SourceSet: request.SourceSet, ReleaseSet: request.ReleaseSet, StacksSet: request.StacksSet, ProfileSet: request.ProfileSet, ProjectNameSet: request.ProjectNameSet, ModulePathSet: request.ModulePathSet, JavaGroupSet: request.JavaGroupSet, ServiceNameSet: request.ServiceNameSet, VerifySet: request.VerifySet}, DryRun: request.DryRun, Yes: request.Yes, Interactive: request.Interactive, Confirm: confirm, SokuVersion: request.SokuVersion, IntegrationSource: request.IntegrationSource, IntegrationRef: request.IntegrationRef, IntegrationConfigPath: request.IntegrationConfig, IntegrationFetcher: initcmd.NewSourceClient(), ProjectSync: request.ProjectSync, ProjectSyncProjectNumber: projectNumber}, nil)
		if err != nil {
			var failure *initcmd.Failure
			if errors.As(err, &failure) {
				return Result{}, &ExitError{Code: ExitCode(failure.Code), Key: failure.Key, Message: failure.Message, Cause: failure, Data: failure.Data}
			}
			return Result{}, err
		}
		return Result{Human: initcmd.Human(report), Data: report, Code: ExitSuccess}, nil
	})
}

func statusHandler() Handler {
	return ResultHandlerFunc(func(context.Context, Request) (Result, error) {
		root, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		result, err := lifecyclestatus.Inspect(root)
		if err != nil {
			var validationError *lifecyclestatus.ValidationError
			if errors.As(err, &validationError) {
				return Result{}, &ExitError{
					Code: ExitValidationFailure, Key: "manifest.invalid", Message: validationError.Error(), Cause: validationError,
				}
			}
			return Result{}, err
		}
		return Result{Human: result.Human, Data: result.Report, Code: ExitCode(result.Code)}, nil
	})
}

func transitionHandler(apply bool) Handler {
	return ResultHandlerFunc(func(ctx context.Context, request Request) (Result, error) {
		root, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		confirm := func(report initcmd.TransitionReport) (bool, error) {
			if request.PromptOutput == nil || request.Input == nil {
				return false, fmt.Errorf("interactive streams are unavailable")
			}
			if _, err := fmt.Fprint(request.PromptOutput, initcmd.HumanTransition("upgrade", report)+"Apply this plan? [y/N] "); err != nil {
				return false, err
			}
			var answer string
			_, err := fmt.Fscanln(request.Input, &answer)
			if err != nil && err != io.EOF {
				return false, err
			}
			answer = strings.ToLower(strings.TrimSpace(answer))
			return answer == "y" || answer == "yes", nil
		}
		report, err := initcmd.RunTransition(ctx, initcmd.TransitionOptions{
			Root: root, ConfigPath: request.ConfigPath, TargetRelease: request.BoilerplateRelease, TargetProfile: request.Profile, DryRun: request.DryRun,
			Yes: request.Yes, Interactive: request.Interactive, Confirm: confirm,
			SokuVersion: request.SokuVersion, IntegrationSource: request.IntegrationSource,
			IntegrationRef: request.IntegrationRef, IntegrationConfigPath: request.IntegrationConfig,
			IntegrationFetcher: initcmd.NewSourceClient(),
		}, nil, apply)
		if err != nil {
			var failure *initcmd.Failure
			if errors.As(err, &failure) {
				return Result{}, &ExitError{Code: ExitCode(failure.Code), Key: failure.Key, Message: failure.Message, Cause: failure, Data: failure.Data}
			}
			return Result{}, err
		}
		code := ExitSuccess
		if !apply && report.HasChanges {
			code = ExitChangesFound
		}
		return Result{Human: initcmd.HumanTransition(request.Command, report), Data: report, Code: code}, nil
	})
}
