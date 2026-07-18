package cli

import (
	"context"
	"errors"
	"os"

	lifecyclestatus "github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/status"
)

// Request contains the stable invocation state passed to lifecycle behavior.
type Request struct {
	Command        string
	ConfigPath     string
	JSON           bool
	Quiet          bool
	NonInteractive bool
	DryRun         bool
	Yes            bool
	Interactive    bool
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
		Init:    unavailableHandler("init"),
		Status:  statusHandler(),
		Diff:    unavailableHandler("diff"),
		Upgrade: unavailableHandler("upgrade"),
	}
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
