package command

import (
	"fmt"
	"github.com/Superdanda/hade/framework/cobra"
	"os"
	"os/exec"
)

func AddKernelCommands(root *cobra.Command) {
	root.AddCommand(initCronCommand())
	// 挂载AppCommand命令
	root.AddCommand(initAppCommand())
	root.AddCommand(initEnvCommand())
	root.AddCommand(initBuildCommand())
	root.AddCommand(initDevCommand())
	root.AddCommand(initProviderCommand())
	root.AddCommand(initCmdCommand())
	root.AddCommand(initMiddlewareCommand())
	root.AddCommand(initNewCommand())
	root.AddCommand(initSwaggerCommand())
	root.AddCommand(initDeployCommand())
}

// 封装通用的命令执行器
func runCommand(path string, args []string) error {
	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("命令执行失败: %v", err)
	}
	return nil
}
