package command

import (
	"fmt"
	"github.com/Superdanda/hade/framework/cobra"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/util"
)

func initEnvCommand() *cobra.Command {
	envCommand.AddCommand(envListCommand)
	return envCommand
}

var envCommand = &cobra.Command{
	Use:   "env",
	Short: "获取当前的App环境",
	Run: func(c *cobra.Command, args []string) {
		container := c.GetContainer()
		env := container.MustMake(contract.EnvKey).(contract.Env)
		fmt.Println("environment:", env.AppEnv())
	},
}

// envListCommand 获取所有的App环境变量
var envListCommand = &cobra.Command{
	Use:   "list",
	Short: "获取所有的环境变量",
	Run: func(c *cobra.Command, args []string) {
		// 获取env环境
		container := c.GetContainer()
		envService := container.MustMake(contract.EnvKey).(contract.Env)
		envs := envService.All()
		outs := [][]string{}
		for k, v := range envs {
			outs = append(outs, []string{k, v})
		}
		util.PrettyPrint(outs)
	},
}
