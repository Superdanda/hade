package command

import (
	"github.com/Superdanda/hade/framework/cobra"
	"log"
	"os/exec"
)

var npmCommand = &cobra.Command{
	Use:   "npm",
	Short: "运行npm命令，要求npm必须安装",
	RunE: func(c *cobra.Command, args []string) error {
		path, err := exec.LookPath("npm")
		if err != nil {
			log.Fatalln("请在PATH路径中安装npm")
		}
		return runCommand(path, args)
	},
}
