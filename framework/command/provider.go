package command

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/cobra"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/util"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/jianfengye/collection"
	"github.com/pkg/errors"
)

func initProviderCommand() *cobra.Command {
	providerCommand.AddCommand(providerCreateCommand)
	providerCommand.AddCommand(providerListCommand)
	providerCommand.AddCommand(providerRepositoryCommand)
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
		interfaceNames := &RouteNode{}
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

		// 收集用户输入并填充嵌套映射
		for {
			prompt := &survey.Input{
				Message: "请输入接口路径（格式：/user/login，直接按回车结束输入）：",
			}

			var input string
			err := survey.AskOne(prompt, &input)
			if err != nil {
				return err
			}

			// 如果用户直接按回车，结束输入
			if strings.TrimSpace(input) == "" {
				fmt.Println("接口输入结束")
				break
			}

			// 解析输入的路径为路径部分
			pathParts := strings.Split(strings.TrimPrefix(input, "/"), "/")
			if len(pathParts) == 0 {
				fmt.Println("路径格式错误，请输入正确格式：/节点/接口")
				continue
			}

			// 将路径部分插入到路由树中
			insertIntoRouteTree(pathParts, []string{}, interfaceNames)
			fmt.Printf("已添加接口：%s\n", input)
		}
		interfaceNames.NeedExtra = false

		// 打印所有添加的接口（测试用）
		printRouteTree(interfaceNames, 0)

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
			"structName":  name,
		}
		// 创建title这个模版方法
		funcs := template.FuncMap{
			"title": strings.Title,
			"dict": func(values ...interface{}) (map[string]interface{}, error) {
				if len(values)%2 != 0 {
					return nil, fmt.Errorf("invalid dict call: missing key or value")
				}
				dict := make(map[string]interface{}, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil, fmt.Errorf("dict keys must be strings")
					}
					dict[key] = values[i+1]
				}
				return dict, nil
			},
			"len": func(v interface{}) int {
				return reflect.ValueOf(v).Len()
			},
		}

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
			file := filepath.Join(pFolder, folder, "services.go")
			f, err := os.Create(file)
			if err != nil {
				return err
			}
			t := template.Must(template.New("services").Funcs(funcs).Parse(serviceTmp))
			if err := t.Execute(f, data); err != nil {
				return err
			}
		}

		if interfaceNames.Children == nil || len(interfaceNames.Children) == 0 {
			return nil
		}

		moduleFolder := app.HttpModuleFolder()
		pModuleFolder := filepath.Join(moduleFolder, name)
		util.EnsureDir(pModuleFolder)
		{
			// module 目录下 创建 服务包

			// 创建api 我呢见
			{
				// 创建 api.go
				file := filepath.Join(pModuleFolder, "api.go")
				f, err := os.Create(file)
				if err != nil {
					return err
				}
				data["interfaces"] = interfaceNames // 传递嵌套的接口名称映射
				t := template.Must(template.New("api").Funcs(funcs).Parse(apiTmp))
				if err := t.Execute(f, data); err != nil {
					return err
				}
			}

			// 创建api_controller文件
			{
				tmpl := template.Must(template.New("controller").Funcs(funcs).Parse(apiControllerTmp))
				data["packageName"] = name
				data["structName"] = name

				// 递归生成控制器文件
				if err := generateControllers(interfaceNames, []string{}, tmpl, data, pModuleFolder); err != nil {
					fmt.Println("生成控制器失败:", err)
					return err
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

var providerRepositoryCommand = &cobra.Command{
	Use:   "repository",
	Short: "创建仓储层实现",
	RunE: func(c *cobra.Command, args []string) error {
		container := c.GetContainer()
		fmt.Println("创建一个仓储层实现")
		var name string
		var idType string
		{
			prompt := &survey.Input{
				Message: "请输入仓储层实现名称（例如：user）：",
			}
			err := survey.AskOne(prompt, &name)
			if err != nil {
				return err
			}
		}

		name = strings.TrimSpace(name)
		if name == "" {
			fmt.Println("服务名称不能为空")
			return nil
		}

		{
			prompt := &survey.Input{
				Message: "请输入仓储层存储模型的ID类型（默认为：int64）：",
			}
			err := survey.AskOne(prompt, &idType)
			if err != nil {
				return err
			}
		}
		idType = strings.TrimSpace(idType)
		if idType == "" {
			idType = "int64"
		}

		// 检查服务是否存在
		// 这里可以添加检查逻辑，防止重复创建

		app := container.MustMake(contract.AppKey).(contract.App)
		config := container.MustMake(contract.ConfigKey).(contract.Config)
		appName := config.GetAppName()
		infrastructureDir := app.InfrastructureFolder()

		// 准备模板数据
		data := map[string]interface{}{
			"ModuleAlias":  fmt.Sprintf("%sModule", name),
			"ModulePath":   fmt.Sprintf("%s/app/provider/%v", appName, name),
			"StructName":   strings.Title(name),
			"EntityName":   strings.Title(name),
			"EntityKey":    fmt.Sprintf("%sKey", strings.Title(name)),
			"VariableName": name,
			"AppName":      appName,
			"IDType":       idType,
		}

		// 定义模板函数
		funcs := template.FuncMap{
			"title": strings.Title,
			"lower": strings.ToLower,
		}

		// 解析模板文件
		//tmplPath := filepath.Join(app.TemplateFolder(), "repository_template.go.tmpl")
		tmpl, err := template.New("repository").Funcs(funcs).Parse(repositoryTmp)
		if err != nil {
			return err
		}

		// 确定生成文件的路径
		infrastructurePath := filepath.Join(infrastructureDir, fmt.Sprintf("%s.go", name))
		if err := os.MkdirAll(infrastructureDir, 0755); err != nil {
			return err
		}

		// 创建并写入文件
		file, err := os.Create(infrastructurePath)
		if err != nil {
			return err
		}
		defer file.Close()

		err = tmpl.Execute(file, data)
		if err != nil {
			return err
		}

		fmt.Printf("成功创建服务：%s，文件位于：%s\n", name, infrastructurePath)
		return nil
	},
}

func generateControllers(node *RouteNode, pathParts []string, tmpl *template.Template, data map[string]interface{}, moduleFolder string) error {
	// 更新路径部分
	newPathParts := append(pathParts, node.Path)

	// 如果当前节点需要生成处理函数
	if node.NeedExtra {
		data["interfaceName"] = strings.Join(newPathParts, "_")
		data["methodName"] = node.HandlerName

		// 生成文件名，例如 api_user_edit_auto.go
		fileName := "api" + strings.Join(newPathParts, "_") + ".go"
		filePath := filepath.Join(moduleFolder, fileName)

		// 创建控制器文件
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %s", filePath)
		}
		defer file.Close()

		// 渲染模板
		if err := tmpl.Execute(file, data); err != nil {
			return fmt.Errorf("渲染模板失败: %s", filePath)
		}
		fmt.Printf("生成接口控制器文件: %s\n", filePath)
	}

	// 递归处理子节点
	for _, child := range node.Children {
		if err := generateControllers(child, newPathParts, tmpl, data, moduleFolder); err != nil {
			return err
		}
	}
	return nil
}

func insertIntoRouteTree(pathParts []string, fullPathParts []string, currentNode *RouteNode) {
	if len(pathParts) == 0 {
		currentNode.NeedExtra = true
		return
	}

	// 获取当前路径部分
	part := pathParts[0]
	fullPathParts = append(fullPathParts, part)

	// 查找是否已存在该路径部分的子节点
	var child *RouteNode
	for _, node := range currentNode.Children {
		if node.Path == part {
			child = node
			break
		}
	}

	// 如果子节点不存在，则创建新的子节点
	if child == nil {
		handlerName := ""
		for _, v := range fullPathParts {
			handlerName += strings.Title(v)
		}
		child = &RouteNode{
			Path:        part,
			HandlerName: handlerName,
		}
		currentNode.Children = append(currentNode.Children, child)
	}

	// 递归处理剩余路径部分
	insertIntoRouteTree(pathParts[1:], fullPathParts, child)
}

func printRouteTree(node *RouteNode, level int) {
	indent := strings.Repeat("  ", level)
	fmt.Printf("%s- %s\n", indent, node.Path)
	for _, child := range node.Children {
		printRouteTree(child, level+1)
	}
}

type RouteNode struct {
	Path        string       // 路由路径
	Children    []*RouteNode // 子路由节点列表
	NeedExtra   bool         // 是否需要额外生成接口
	HandlerName string
}

var contractTmp = `package {{.packageName}}

const {{.packageName | title}}Key = "{{.appName}}:{{.packageName}}"

type Service interface {
	// 请在这里定义你的方法
    Foo() string
}

type {{.packageName | title}} struct {}
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

// 注册路由
func RegisterRoutes(r *gin.Engine) error {

    api := {{.packageName | title}}Api{}

	if !r.IsBind({{.packageName}}.{{.packageName | title}}Key) {
		r.Bind(&{{.packageName}}.{{.packageName | title}}Provider{})
	}

    {{template "registerRoutes" dict "node" .interfaces "groupVar" "r" "apiVar" "api"}}

    return nil
}

{{- define "registerRoutes"}}
{{- $node := .node -}}
{{- $groupVar := .groupVar -}}
{{- $apiVar := .apiVar -}}

{{- if ne $node.Path "root" -}}
    {{- $hasChildren := gt (len $node.Children) 0 -}}
    {{- $groupName := (printf "%sGroup" $node.Path) -}}
    {{- if $hasChildren -}}
        {{$groupName}} := {{$groupVar}}.Group("/{{$node.Path}}")
        {
            {{- if $node.NeedExtra -}}
            {{$groupName}}.POST("/", {{$apiVar}}.{{ $node.HandlerName }})
            {{- end}}
            {{range $child := $node.Children}}
                {{template "registerRoutes" dict "node" $child "groupVar" $groupName "apiVar" $apiVar}}
            {{- end}}
        }
    {{- else}}
        {{- if $node.NeedExtra}}
        {{$groupVar}}.POST("/{{$node.Path}}", {{$apiVar}}.{{ $node.HandlerName }})
        {{- end}}
    {{end}}
{{- else}}
    {{range $child := $node.Children}}
        {{template "registerRoutes" dict "node" $child "groupVar" $groupVar "apiVar" $apiVar}}
    {{- end}}
{{- end}}
{{- end}}
`

var apiControllerTmp = `package {{.packageName}}
import (
    "github.com/Superdanda/hade/framework/gin"
)

// {{.methodName}} handler
func (api *{{.structName | title}}Api) {{.methodName}}(c *gin.Context) {
    // TODO: Implement {{.methodName}}
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

var repositoryTmp = `package infrastructure
import (
	"{{.AppName}}/app/provider/database_connect"
	{{.ModuleAlias}} "{{.ModulePath}}"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/provider/repository"
	"gorm.io/gorm"
)

type {{.StructName}}Repository struct {
	container framework.Container
	db        *gorm.DB
	contract.OrmRepository[{{.ModuleAlias}}.{{.EntityName}}, {{.IDType}}]
	{{.ModuleAlias}}.Repository
}

func NewOrm{{.StructName}}RepositoryAndRegister(container framework.Container) {
	// 获取必要的服务对象
	connectService := container.MustMake(database_connect.DatabaseConnectKey).(database_connect.Service)
	infrastructureService := container.MustMake(contract.InfrastructureKey).(contract.InfrastructureService)
	repositoryService := container.MustMake(contract.RepositoryKey).(contract.RepositoryService)

	connect := connectService.DefaultDatabaseConnect()
	{{.VariableName}}OrmService := &{{.StructName}}Repository{container: container, db: connect}
	infrastructureService.RegisterOrmRepository({{.ModuleAlias}}.{{.EntityKey}}, {{.VariableName}}OrmService)

	// 注册通用仓储对象
	repository.RegisterRepository[{{.ModuleAlias}}.{{.EntityName}}, {{.IDType}}](repositoryService, {{.ModuleAlias}}.{{.EntityKey}}, {{.VariableName}}OrmService)
}

func (u *{{.StructName}}Repository) SaveToDB(entity *{{.ModuleAlias}}.{{.EntityName}}) error {
	return u.db.Save(entity).Error
}

func (u *{{.StructName}}Repository) FindByIDFromDB(id {{.IDType}}) (*{{.ModuleAlias}}.{{.EntityName}}, error) {
	entity := &{{.ModuleAlias}}.{{.EntityName}}{}
	err := u.db.First(entity, id).Error
	return entity, err
}

func (u *{{.StructName}}Repository) FindByIDsFromDB(ids []{{.IDType}}) ([]*{{.ModuleAlias}}.{{.EntityName}}, error) {
	var entities []*{{.ModuleAlias}}.{{.EntityName}}
	err := u.db.Where("id IN ?", ids).Find(&entities).Error
	return entities, err
}

func (u *{{.StructName}}Repository) GetPrimaryKey(entity *{{.ModuleAlias}}.{{.EntityName}}) {{.IDType}} {
	return entity.ID
}

func (u *{{.StructName}}Repository) GetBaseField() string {
	return {{.ModuleAlias}}.{{.EntityKey}}
}

func (u *{{.StructName}}Repository) GetFieldQueryFunc(fieldName string) (func(value string) ([]*{{.ModuleAlias}}.{{.EntityName}}, error), bool) {
	switch fieldName {
	// 根据您的实际情况添加字段查询函数
	default:
		return nil, false
	}
}

func (u *{{.StructName}}Repository) GetFieldInQueryFunc(fieldName string) (func(values []string) ([]*{{.ModuleAlias}}.{{.EntityName}}, error), bool) {
	switch fieldName {
	// 根据您的实际情况添加字段批量查询函数
	default:
		return nil, false
	}
}

func (u *{{.StructName}}Repository) GetFieldValueFunc(fieldName string) (func(entity *{{.ModuleAlias}}.{{.EntityName}}) string, bool) {
	switch fieldName {
	// 根据您的实际情况添加获取字段值的函数
	default:
		return nil, false
	}
}
`
