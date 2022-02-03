package tests

import (
	"github.com/kubeshop/testkube/cmd/kubectl-testkube/commands/common"
	"github.com/kubeshop/testkube/cmd/kubectl-testkube/commands/common/validator"
	"github.com/kubeshop/testkube/pkg/ui"
	"github.com/spf13/cobra"
)

func NewDeleteTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <testName>",
		Short: "Delete tests",
		Long:  `Delete tests by name`,
		Args:  validator.TestName,
		Run: func(cmd *cobra.Command, args []string) {
			ui.Logo()

			client, namespace := common.GetClient(cmd)

			name := args[0]
			err := client.DeleteTest(name, namespace)
			ui.ExitOnError("delete test "+name+" from namespace "+namespace, err)
			ui.Success("Succesfully deleted", name)
		},
	}

	return cmd
}