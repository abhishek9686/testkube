package commands

import (
	"github.com/kubeshop/testkube/cmd/kubectl-testkube/commands/common/validator"
	"github.com/kubeshop/testkube/cmd/kubectl-testkube/commands/executors"
	"github.com/kubeshop/testkube/cmd/kubectl-testkube/commands/tests"
	"github.com/kubeshop/testkube/cmd/kubectl-testkube/commands/testsuites"
	"github.com/kubeshop/testkube/cmd/kubectl-testkube/commands/webhooks"
	"github.com/kubeshop/testkube/pkg/ui"
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <resourceName>",
		Aliases: []string{"c"},
		Short:   "Create resource",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Logo()
			cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			validator.PersistentPreRunVersionCheck(cmd, Version)
		}}

	cmd.AddCommand(tests.NewCreateTestsCmd())
	cmd.AddCommand(testsuites.NewCreateTestSuitesCmd())
	cmd.AddCommand(webhooks.NewCreateWebhookCmd())
	cmd.AddCommand(executors.NewCreateExecutorCmd())

	return cmd
}
