package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const rootHelp = `Manage Soku conventions safely and reproducibly.

Usage:
  soku <command> [flags]

Commands:
  init        Initialize managed convention state
  status      Inspect lifecycle state without changes
  diff        Compare current and desired managed state
  upgrade     Upgrade managed convention state

Flags:
      --config string       explicit portable YAML configuration file
      --help                help for soku
      --json                emit one machine-readable JSON envelope
      --non-interactive     forbid interactive prompts
      --quiet               suppress non-essential human output
      --version             print version information
`

type options struct {
	configPath     string
	json           bool
	quiet          bool
	nonInteractive bool
	version        bool
}

type dependencies struct {
	runtime  Runtime
	handlers Handlers
	metadata BuildMetadata
}

// Run connects the process streams to the injectable command implementation.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	file, _ := stdin.(*os.File)
	return runWith(args, stdout, stderr, dependencies{
		runtime:  osRuntime{stdin: file},
		handlers: defaultHandlers(),
		metadata: resolveBuildMetadata(),
	})
}

func runWith(args []string, stdout, stderr io.Writer, deps dependencies) int {
	jsonMode := hasJSONFlag(args)
	out := &output{stdout: stdout, stderr: stderr, json: jsonMode}
	invokedCommand := commandFromArgs(args)
	if invokedCommand == cobra.ShellCompRequestCmd || invokedCommand == cobra.ShellCompNoDescRequestCmd {
		exitError := invocationError("shell completion is not available")
		out.failure(invokedCommand, exitError)
		return int(exitError.Code)
	}
	root := newRootCommand(deps, out)
	root.SetArgs(args)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	_, err := root.ExecuteC()
	if err == nil {
		return 0
	}
	var completed *resultExit
	if errors.As(err, &completed) {
		return int(completed.Code)
	}

	exitError := normalizeError(err)
	out.failure(invokedCommand, exitError)
	return int(exitError.Code)
}

func newRootCommand(deps dependencies, out *output) *cobra.Command {
	var opts options
	root := &cobra.Command{
		Use:           "soku",
		Short:         "Manage Soku conventions safely and reproducibly",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return invocationError("unexpected argument %q", args[0])
			}
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			if opts.version {
				return out.version(deps.metadata)
			}
			return out.help("help", rootHelp)
		},
	}

	root.CompletionOptions.DisableDefaultCmd = true
	root.DisableSuggestions = true
	root.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Hidden: true,
		RunE: func(*cobra.Command, []string) error {
			return invocationError("help is available only through --help")
		},
	})
	root.SetHelpFunc(func(command *cobra.Command, _ []string) {
		_ = out.help(commandName(command), helpFor(command))
	})
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return invocationError("%s", err)
	})
	root.InitDefaultHelpFlag()
	root.Flags().Lookup("help").Shorthand = ""

	flags := root.PersistentFlags()
	flags.StringVar(&opts.configPath, "config", "", "explicit portable YAML configuration file")
	flags.BoolVar(&opts.json, "json", false, "emit one machine-readable JSON envelope")
	flags.BoolVar(&opts.quiet, "quiet", false, "suppress non-essential human output")
	flags.BoolVar(&opts.nonInteractive, "non-interactive", false, "forbid interactive prompts")
	flags.BoolVar(&opts.version, "version", false, "print version information")

	root.AddCommand(
		newLifecycleCommand("init", true, &opts, deps, out, deps.handlers.Init),
		newLifecycleCommand("status", false, &opts, deps, out, deps.handlers.Status),
		newLifecycleCommand("diff", false, &opts, deps, out, deps.handlers.Diff),
		newLifecycleCommand("upgrade", true, &opts, deps, out, deps.handlers.Upgrade),
	)
	return root
}

func newLifecycleCommand(
	name string,
	mutation bool,
	opts *options,
	deps dependencies,
	out *output,
	handler Handler,
) *cobra.Command {
	var dryRun bool
	var yes bool
	var command *cobra.Command
	command = &cobra.Command{
		Use:   name,
		Short: lifecycleDescription(name),
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return invocationError("%s does not accept arguments", name)
			}
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			if opts.version {
				return out.version(deps.metadata)
			}
			if err := validateConfig(deps.runtime, opts.configPath); err != nil {
				return err
			}

			terminal := deps.runtime.IsTerminal()
			interactive := !opts.nonInteractive && terminal
			if mutation && dryRun && yes {
				return invocationError("--dry-run and --yes cannot be used together")
			}
			if mutation && !dryRun && !yes && !interactive {
				return invocationError("non-interactive mutation requires --dry-run or --yes")
			}

			request := Request{
				Command:        name,
				ConfigPath:     opts.configPath,
				JSON:           out.json,
				Quiet:          opts.quiet,
				NonInteractive: opts.nonInteractive || !terminal,
				DryRun:         dryRun,
				Yes:            yes,
				Interactive:    interactive,
			}
			result, err := invokeHandler(command.Context(), handler, request)
			if err != nil {
				return err
			}
			if err := out.result(name, result, opts.quiet); err != nil {
				return err
			}
			if result.Code != ExitSuccess {
				return &resultExit{Code: result.Code}
			}
			return nil
		},
	}
	if mutation {
		command.Flags().BoolVar(&dryRun, "dry-run", false, "produce a complete plan without writing")
		command.Flags().BoolVar(&yes, "yes", false, "approve an already validated mutation plan")
	}
	command.InitDefaultHelpFlag()
	command.Flags().Lookup("help").Shorthand = ""
	return command
}

func invokeHandler(ctx context.Context, handler Handler, request Request) (result Result, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = Result{}
			err = &ExitError{
				Code:    ExitInternalError,
				Key:     "internal.error",
				Message: "internal command failure",
			}
		}
	}()
	if handler == nil {
		return Result{}, &ExitError{
			Code:    ExitInternalError,
			Key:     "internal.error",
			Message: "internal command failure",
		}
	}
	result, err = handler.Handle(ctx, request)
	if err != nil {
		var exitError *ExitError
		if errors.As(err, &exitError) {
			return Result{}, err
		}
		return Result{}, &ExitError{
			Code:    ExitInternalError,
			Key:     "internal.error",
			Message: "internal command failure",
			Cause:   err,
		}
	}
	if result.Code != ExitSuccess && result.Code != ExitChangesFound && result.Code != ExitCompatibilityFailure {
		return Result{}, &ExitError{
			Code: ExitInternalError, Key: "internal.error", Message: "internal command failure",
		}
	}
	return result, nil
}

func validateConfig(runtime Runtime, path string) error {
	if path == "" {
		return nil
	}
	info, err := runtime.Stat(path)
	if err != nil {
		return configError(path, err)
	}
	if !info.Mode().IsRegular() {
		return configError(path, fmt.Errorf("not a regular file"))
	}
	reader, err := runtime.Open(path)
	if err != nil {
		return configError(path, err)
	}
	if err := reader.Close(); err != nil {
		return configError(path, err)
	}
	return nil
}

func configError(path string, cause error) *ExitError {
	return &ExitError{
		Code:    ExitValidationFailure,
		Key:     "configuration.invalid",
		Message: fmt.Sprintf("configuration file %q is not a readable regular file", path),
		Cause:   cause,
	}
}

func normalizeError(err error) *ExitError {
	var exitError *ExitError
	if errors.As(err, &exitError) {
		return exitError
	}
	return invocationError("%s", err)
}

func hasJSONFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--json" || arg == "--json=true" {
			return true
		}
	}
	return false
}

func commandFromArgs(args []string) string {
	valueFlags := map[string]bool{"--config": true}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if valueFlags[arg] {
			index++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return arg
	}
	return "soku"
}

func commandName(command *cobra.Command) string {
	if command == nil || command.Name() == "soku" {
		return "help"
	}
	return command.Name()
}

func lifecycleDescription(name string) string {
	switch name {
	case "init":
		return "Initialize managed convention state"
	case "status":
		return "Inspect lifecycle state without changes"
	case "diff":
		return "Compare current and desired managed state"
	case "upgrade":
		return "Upgrade managed convention state"
	default:
		return "Manage convention state"
	}
}

func helpFor(command *cobra.Command) string {
	if command == nil || command.Name() == "soku" {
		return rootHelp
	}
	mutationFlags := ""
	if command.Name() == "init" || command.Name() == "upgrade" {
		mutationFlags = "      --dry-run            produce a complete plan without writing\n" +
			"      --yes                approve an already validated mutation plan\n"
	}
	return fmt.Sprintf(`%s.

Usage:
  soku %s [flags]

Flags:
      --config string       explicit portable YAML configuration file
%s      --help                help for %s
      --json                emit one machine-readable JSON envelope
      --non-interactive     forbid interactive prompts
      --quiet               suppress non-essential human output
      --version             print version information
`, lifecycleDescription(command.Name()), command.Name(), mutationFlags, command.Name())
}
