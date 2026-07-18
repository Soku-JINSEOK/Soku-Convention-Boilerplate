package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/initcmd"
	lifecyclestatus "github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/status"
)

// Request contains the stable invocation state passed to lifecycle behavior.
type Request struct {
	Command            string
	ConfigPath         string
	JSON               bool
	Quiet              bool
	NonInteractive     bool
	DryRun             bool
	Yes                bool
	Interactive        bool
	BoilerplateSource  string
	BoilerplateRelease string
	Stacks             []string
	Profile            string
	ProjectName        string
	ModulePath         string
	JavaGroup          string
	ServiceName        string
	Verify             bool
	SourceSet          bool
	ReleaseSet         bool
	StacksSet          bool
	ProfileSet         bool
	ProjectNameSet     bool
	ModulePathSet      bool
	JavaGroupSet       bool
	ServiceNameSet     bool
	VerifySet          bool
	Input              io.Reader
	PromptOutput       io.Writer
	SokuVersion        string
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
	Init    Handler
	Status  Handler
	Diff    Handler
	Upgrade Handler
}

func defaultHandlers() Handlers {
	return Handlers{
		Init:    initHandler(),
		Status:  statusHandler(),
		Diff:    unavailableHandler("diff"),
		Upgrade: unavailableHandler("upgrade"),
	}
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
		report, err := initcmd.Run(ctx, initcmd.Options{Root: root, ConfigPath: request.ConfigPath, Explicit: initcmd.Explicit{Source: request.BoilerplateSource, Release: request.BoilerplateRelease, Stacks: request.Stacks, Profile: request.Profile, ProjectName: request.ProjectName, ModulePath: request.ModulePath, JavaGroup: request.JavaGroup, ServiceName: request.ServiceName, Verify: request.Verify, SourceSet: request.SourceSet, ReleaseSet: request.ReleaseSet, StacksSet: request.StacksSet, ProfileSet: request.ProfileSet, ProjectNameSet: request.ProjectNameSet, ModulePathSet: request.ModulePathSet, JavaGroupSet: request.JavaGroupSet, ServiceNameSet: request.ServiceNameSet, VerifySet: request.VerifySet}, DryRun: request.DryRun, Yes: request.Yes, Interactive: request.Interactive, Confirm: confirm, SokuVersion: request.SokuVersion}, nil)
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

func unavailableHandler(command string) Handler {
	return HandlerFunc(func(context.Context, Request) error {
		return unavailableError(command)
	})
}
