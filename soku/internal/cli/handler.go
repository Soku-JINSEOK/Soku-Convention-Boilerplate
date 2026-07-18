package cli

import "context"

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

// Handler is replaced by later lifecycle issues without changing CLI parsing.
type Handler interface {
	Handle(context.Context, Request) error
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(context.Context, Request) error

func (f HandlerFunc) Handle(ctx context.Context, request Request) error {
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
		Status:  unavailableHandler("status"),
		Diff:    unavailableHandler("diff"),
		Upgrade: unavailableHandler("upgrade"),
	}
}

func unavailableHandler(command string) Handler {
	return HandlerFunc(func(context.Context, Request) error {
		return unavailableError(command)
	})
}
