package cli

import "fmt"

// ExitCode is one of the lifecycle contract's complete process outcomes.
type ExitCode int

const (
	ExitSuccess ExitCode = iota
	ExitInternalError
	ExitValidationFailure
	ExitChangesFound
	ExitSafetyRefusal
	ExitCompatibilityFailure
	ExitSourceFailure
	ExitApplyRolledBack
	ExitRollbackFailure
)

// ExitError is a stable error that maps directly to the public CLI contract.
type ExitError struct {
	Code    ExitCode
	Key     string
	Message string
	Cause   error
}

type resultExit struct {
	Code ExitCode
}

func (e *resultExit) Error() string {
	return fmt.Sprintf("command completed with exit code %d", e.Code)
}

func (e *ExitError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return fmt.Sprintf("soku failed with exit code %d", e.Code)
}

func (e *ExitError) Unwrap() error {
	return e.Cause
}

func invocationError(format string, args ...any) *ExitError {
	return &ExitError{
		Code:    ExitValidationFailure,
		Key:     "invocation.invalid",
		Message: fmt.Sprintf(format, args...),
	}
}

func unavailableError(command string) *ExitError {
	return &ExitError{
		Code:    ExitCompatibilityFailure,
		Key:     "feature.unavailable",
		Message: fmt.Sprintf("%s is not available in this release", command),
	}
}
