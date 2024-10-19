package command

import "github.com/Superdanda/hade/framework/cobra"

func AddKernelCommands(root *cobra.Command) {
	root.AddCommand(DemoCommand)
	// 挂载AppCommand命令
	root.AddCommand(initAppCommand())
	root.AddCommand(initCronCommand())
}
