package cli

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
)

var completionShells = []string{"bash", "zsh", "fish", "powershell"}

func newCompletionCommand(opts *options, out *output) *cobra.Command {
	command := &cobra.Command{
		Use:       "completion <bash|zsh|fish|powershell>",
		Short:     "Generate a shell completion script",
		Args:      cobra.ExactArgs(1),
		ValidArgs: completionShells,
		RunE: func(command *cobra.Command, args []string) error {
			var script bytes.Buffer
			root := command.Root()
			var err error
			switch args[0] {
			case "bash":
				err = root.GenBashCompletionV2(&script, true)
			case "zsh":
				err = root.GenZshCompletion(&script)
			case "fish":
				err = root.GenFishCompletion(&script, true)
			case "powershell":
				err = root.GenPowerShellCompletion(&script)
			default:
				return invocationError("unsupported shell %q", args[0])
			}
			if err != nil {
				return fmt.Errorf("generate %s completion: %w", args[0], err)
			}
			return out.completion(args[0], script.String(), opts.json)
		},
	}
	command.InitDefaultHelpFlag()
	command.Flags().Lookup("help").Shorthand = ""
	return command
}

func fixedCompletions(values ...string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}
