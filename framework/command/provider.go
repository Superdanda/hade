package command

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/cobra"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/util"

	"github.com/jianfengye/collection"
	"github.com/pkg/errors"
)

func initProviderCommand() *cobra.Command {
	providerCommand.AddCommand(providerCreateCommand)
	providerCommand.AddCommand(providerListCommand)
	return providerCommand
}

var providerCommand = &cobra.Command{
	Use:   "provider",
	Short: "服务相关命令",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			c.Help()
		}
		return nil
	},
}

var providerListCommand = &cobra.Command{
	Use:   "list",
	Short: "列出容器内的所有服务，列出它们的字符串凭证",
	RunE: func(c *cobra.Command, args []string) error {
		container := c.GetContainer()
		nameList := container.NameList()
		// 打印
		for _, line := range nameList {
			println(line)
		}
		return nil
	},
}

var providerCreateCommand = &cobra.Command{
	Use:     "create",
	Aliases: []string{"create", "init"},
	Short:   "创建服务",
	RunE: func(c *cobra.Command, args []string) error {
		container := c.GetContainer()
		fmt.Println("创建一个服务")
		var name, folder string
		var interfaceNames []string
		{
			prompt := &survey.Input{
				Message: "请输入服务名称(服务凭证)：",
			}
			err := survey.AskOne(prompt, &name)
			if err != nil {
				return err
			}
		}
		{
			prompt := &survey.Input{
				Message: "请输入服务所在目录名称(默认: 同服务名称):",
			}
			err := survey.AskOne(prompt, &folder)
			if err != nil {
				return err
			}
		}

		// 检查服务是否存在
		providers := container.(*framework.HadeContainer).NameList()
		providerColl := collection.NewStrCollection(providers)
		if providerColl.Contains(name) {
			fmt.Println("服务名称已经存在")
			return nil
		}

		if folder == "" {
			folder = name
		}

		app := container.MustMake(contract.AppKey).(contract.App)

		pFolder := app.ProviderFolder()
		subFolders, err := util.SubDir(pFolder)
		if err != nil {
			return err
		}
		subColl := collection.NewStrCollection(subFolders)
		if subColl.Contains(folder) {
			fmt.Println("目录名称已经存在")
			return nil
		}

		// Keep asking for interface names until the user presses Enter to stop
		for {
			prompt := &survey.Input{
				Message: "请输入接口名称（直接按回车结束输入）：",
			}

			var input string
			err := survey.AskOne(prompt, &input)
			if err != nil {
				return err
			}

			// If the user presses Enter without input, break the loop
			if strings.TrimSpace(input) == "" {
				fmt.Println("接口输入结束")
				break
			}

			// Add the valid interface name to the slice
			interfaceNames = append(interfaceNames, input)
			fmt.Printf("已添加接口：%s\n", input)
		}

		// Print all the interface names added by the user
		if len(interfaceNames) > 0 {
			fmt.Println("已添加的接口名称：", interfaceNames)
		} else {
			fmt.Println("未添加任何接口")
		}

		// 开始创建文件
		if err := os.Mkdir(filepath.Join(pFolder, folder), 0700); err != nil {
			return err
		}
		// 模板数据
		config := container.MustMake(contract.ConfigKey).(contract.Config)
		data := map[string]interface{}{
			"appName":     config.GetAppName(),
			"packageName": name,
			"interfaces":  interfaceNames,
		}
		// 创建title这个模版方法
		funcs := template.FuncMap{"title": strings.Title}
		{
			//  创建contract.go
			file := filepath.Join(pFolder, folder, "contract.go")
			f, err := os.Create(file)
			if err != nil {
				return errors.Cause(err)
			}

			// 使用contractTmp模版来初始化template，并且让这个模版支持title方法，即支持{{.packageName | title}}
			t := template.Must(template.New("contract").Funcs(funcs).Parse(contractTmp))
			// 将name传递进入到template中渲染，并且输出到contract.go 中
			if err := t.Execute(f, data); err != nil {
				return errors.Cause(err)
			}
		}
		{
			// 创建provider.go
			file := filepath.Join(pFolder, folder, "provider.go")
			f, err := os.Create(file)
			if err != nil {
				return err
			}
			t := template.Must(template.New("provider").Funcs(funcs).Parse(providerTmp))
			if err := t.Execute(f, data); err != nil {
				return err
			}
		}
		{
			//  创建service.go
			file := filepath.Join(pFolder, folder, "service.go")
			f, err := os.Create(file)
			if err != nil {
				return err
			}
			t := template.Must(template.New("service").Funcs(funcs).Parse(serviceTmp))
			if err := t.Execute(f, data); err != nil {
				return err
			}
		}

		moduleFolder := app.HttpModuleFolder()
		pModuleFolder := filepath.Join(moduleFolder, name)
		util.EnsureDir(pModuleFolder)
		{
			// module 目录下 创建 服务包

			// 创建api 我呢见
			{
				//  创建api.go
				file := filepath.Join(pModuleFolder, "api.go")
				f, err := os.Create(file)
				if err != nil {
					return err
				}
				t := template.Must(template.New("api").Funcs(funcs).Parse(apiTmp))
				if err := t.Execute(f, data); err != nil {
					return err
				}
			}

			// 创建api_controller文件
			{
				tmpl, err := template.New("controller").Funcs(funcs).Parse(apiControllerTmp)
				if err != nil {
					fmt.Println("Failed to parse template:", err)
					return err
				}
				// Generate a file for each interface
				for _, interfaceName := range interfaceNames {
					data := map[string]interface{}{
						"packageName":   name,
						"appName":       config.GetAppName(),
						"interfaceName": interfaceName,
					}

					// Create the output file
					filePath := filepath.Join(pModuleFolder, "api_"+interfaceName+".go")
					file, err := os.Create(filePath)
					if err != nil {
						fmt.Println("Failed to create file:", err)
						return errors.New("Failed to create file: " + filePath)
					}
					defer file.Close()

					// Execute the template with the data
					if err := tmpl.Execute(file, data); err != nil {
						fmt.Println("Failed to execute template:", err)
						return errors.New("Failed to execute template: " + filePath)
					}
					fmt.Printf("Generated file: %s\n", filePath)
				}
			}

			//创建 dto 文件
			{
				//  创建dto.go
				file := filepath.Join(pModuleFolder, "dto.go")
				f, err := os.Create(file)
				if err != nil {
					return err
				}
				t := template.Must(template.New("dto").Funcs(funcs).Parse(dtoTmp))
				if err := t.Execute(f, data); err != nil {
					return err
				}
			}

			//创建 mapper 文件
			{
				//  创建mapper.go
				file := filepath.Join(pModuleFolder, "mapper.go")
				f, err := os.Create(file)
				if err != nil {
					return err
				}
				t := template.Must(template.New("mapper").Funcs(funcs).Parse(mapperTmp))
				if err := t.Execute(f, data); err != nil {
					return err
				}
			}
		}
		fmt.Println("创建服务成功, 文件夹地址:", filepath.Join(pFolder, folder))
		fmt.Println("请不要忘记挂载新创建的服务")
		return nil
	},
}

var contractTmp = `package {{.packageName}}

const {{.packageName | title}}Key = "{{.appName}}:{{.packageName}}"

type Service interface {
	// 请在这里定义你的方法
    Foo() string
}
`

var providerTmp = `package {{.packageName}}

import (
	"github.com/Superdanda/hade/framework"
)

type {{.packageName | title}}Provider struct {
	framework.ServiceProvider

	c framework.Container
}

func (sp *{{.packageName | title}}Provider) Name() string {
	return {{.packageName | title}}Key
}

func (sp *{{.packageName | title}}Provider) Register(c framework.Container) framework.NewInstance {
	return New{{.packageName | title}}Service
}

func (sp *{{.packageName | title}}Provider) IsDefer() bool {
	return false
}

func (sp *{{.packageName | title}}Provider) Params(c framework.Container) []interface{} {
	return []interface{}{c}
}

func (sp *{{.packageName | title}}Provider) Boot(c framework.Container) error {
	return nil
}

`

var serviceTmp = `package {{.packageName}}

import "github.com/Superdanda/hade/framework"

type {{.packageName | title}}Service struct {
	container framework.Container
}

func New{{.packageName | title}}Service(params ...interface{}) (interface{}, error) {
	container := params[0].(framework.Container)
	return &{{.packageName | title}}Service{container: container}, nil
}

func (s *{{.packageName | title}}Service) Foo() string {
    return ""
}
`

var apiTmp = `package {{.packageName}}
import (
	"github.com/Superdanda/hade/framework/gin"
)

type {{.packageName | title}}Api struct{}

func RegisterRoutes(r *gin.Engine) error {
	api := {{.packageName | title}}Api{}

	if !r.IsBind({{.packageName}}.{{.packageName | title}}Key) {
		r.Bind(&{{.packageName}}.{{.packageName | title}}Provider{})
	}

	{{.packageName}}Group := r.Group("/{{.packageName}}")
	{
        {{range .interfaces}}
		// {{.}} route
		{{$.packageName}}Group.POST("/{{.}}", api.{{. | title}})
		{{end}}
	}
	return nil
}
`

var apiControllerTmp = `package {{.packageName}}
import (
	"github.com/Superdanda/hade/framework/gin"
)

// {{.interfaceName | title}} handler
func (api *{{.packageName | title}}Api) {{.interfaceName | title}}(c *gin.Context) {
	// TODO: Add logic for {{.interfaceName}}
}
`

var dtoTmp = `package {{.packageName}}

type {{.packageName | title}}DTO struct {} 
`
var mapperTmp = `package {{.packageName}}

func Convert{{.packageName | title}}ToDTO({{.packageName}} *{{.packageName}}.{{.packageName | title}}) *{{.packageName | title}}DTO {
	if {{.packageName}} == nil {
		return nil
	}
	return &{{.packageName | title}}DTO{}
}
`
