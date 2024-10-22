package command

import (
	"log"
	"os/exec"

	"github.com/Superdanda/hade/framework/cobra"
)

// go just run local go bin
var goCommand = &cobra.Command{
	Use:   "go",
	Short: "运行go命令，要求go必须安装",
	RunE: func(c *cobra.Command, args []string) error {
		path, err := exec.LookPath("go")
		if err != nil {
			log.Fatalln("请在PATH路径中安装go")
		}
		return runCommand(path, args)
	},
}
