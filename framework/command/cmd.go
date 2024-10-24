package command

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Superdanda/hade/framework/cobra"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/util"
	"github.com/jianfengye/collection"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"html/template"
	"os"
	"path/filepath"
)

func initCmdCommand() *cobra.Command {
	CmdCommand.AddCommand(cmdListCommand)
	CmdCommand.AddCommand(cmdNewCommand)
	return CmdCommand
}

var CmdCommand = &cobra.Command{
	Use:   "command",
	Short: "显示帮助信息",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			c.Help()
		}
		return nil
	},
}

var cmdListCommand = &cobra.Command{
	Use:   "list",
	Short: "列出所有控制台命令",
	RunE: func(c *cobra.Command, args []string) error {
		cmds := c.Root().Commands()
		ps := [][]string{}
		for _, cmd := range cmds {
			line := []string{cmd.Name(), cmd.Short}
			ps = append(ps, line)
		}
		util.PrettyPrint(ps)
		return nil
	},
}

var cmdNewCommand = &cobra.Command{
	Use:     "new",
	Aliases: []string{"create", "init"}, // 设置别名为 create init
	Short:   "创建一个控制台命令",
	RunE: func(c *cobra.Command, args []string) error {
		container := c.GetContainer()

		fmt.Println("开始创建控制台命令...")
		var name string
		var folder string
		{
			prompt := &survey.Input{
				Message: "请输入控制台命令名称:",
			}
			err := survey.AskOne(prompt, &name)
			if err != nil {
				return err
			}
		}

		{
			prompt := &survey.Input{
				Message: "请输入文件夹名称(默认: 同控制台命令):",
			}
			err := survey.AskOne(prompt, &folder)
			if err != nil {
				return err
			}
		}

		if folder == "" {
			folder = name
		}

		app := container.MustMake(contract.AppKey).(contract.App)

		pFolder := app.CommandFolder()
		subFolders, err := util.SubDir(pFolder)
		if err != nil {
			return err
		}

		if collection.NewStrCollection(subFolders).Contains(folder) {
			fmt.Println("目录名称已经存在")
			return nil
		}

		if err = os.Mkdir(filepath.Join(pFolder, folder), 0700); err != nil {
			return err
		}

		titleCase := cases.Title(language.English)
		funcMap := template.FuncMap{
			"title": func(s string) string {
				return titleCase.String(s)
			},
		}
		{
			//  创建name.go
			file := filepath.Join(pFolder, folder, name+".go")
			f, err := os.Create(file)
			if err != nil {
				return errors.Cause(err)
			}

			// 使用contractTmp模版来初始化template，并且让这个模版支持title方法，即支持{{.|title}}
			t := template.Must(template.New("cmd").Funcs(funcMap).Parse(cmdTmpl))
			// 将name传递进入到template中渲染，并且输出到contract.go 中
			if err := t.Execute(f, name); err != nil {
				return errors.Cause(err)
			}
		}

		fmt.Println("创建新命令行工具成功，路径:", filepath.Join(pFolder, folder))
		fmt.Println("请记得开发完成后将命令行工具挂载到 console/kernel.go")
		return nil
	},
}

// 命令行工具模版
var cmdTmpl = `package {{.}}

import (
	"fmt"

	"github.com/Superdanda/hade/framework/cobra"
)

var {{.|title}}Command = &cobra.Command{
	Use:   "{{.}}",
	Short: "{{.}}",
	RunE: func(c *cobra.Command, args []string) error {
        container := c.GetContainer()
		fmt.Println(container)
		return nil
	},
}

`
