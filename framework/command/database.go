package command

import (
	"errors"
	"fmt"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/cobra"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/provider/orm"
	"go/ast"
	"go/parser"
	"go/token"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"reflect"
)

func initDatabaseCommand() *cobra.Command {
	databaseCommand.AddCommand(databaseModelSyncCommand)
	return databaseCommand
}

var databaseCommand = &cobra.Command{
	Use:   "database",
	Short: "数据库相关命令",
	Long:  "操控数据库相关命令，同步数据模型到数据库",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			c.Help()
		}
		return nil
	},
}

var databaseModelSyncCommand = &cobra.Command{
	Use:   "modelSync",
	Short: "同步数据模型到数据库",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println("请输入数据库连接名称")
		}
		container := c.GetContainer()
		appService := container.MustMake(contract.AppKey).(contract.App)
		configService := container.MustMake(contract.ConfigKey).(contract.Config)
		syncPathFolder := filepath.Join(appService.BaseFolder(), configService.GetString("database.sync.filePath"))

		ormService := container.MustMake(contract.ORMKey).(contract.ORMService)
		db, err := ormService.GetDB(orm.WithConfigPath("database." + args[0]))

		if err != nil {
			return err
		}
		err = AutoMigrateStructs(db, syncPathFolder, container)
		if err != nil {
			return err
		}
		return nil
	},
}

// ScanStructsInPackage 扫描给定包路径下的所有结构体
func ScanStructsInPackage(pkgPath string, container framework.Container) ([]interface{}, error) {
	var structs []interface{}

	// 遍历包路径下的所有 Go 文件
	err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理 .go 文件
		if filepath.Ext(path) != ".go" {
			return nil
		}

		// 解析 Go 文件，构建 AST
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return err
		}

		// 遍历 AST，找到所有类型声明
		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			// 遍历所有类型声明，找到结构体类型
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				// 检查是否为结构体类型
				_, ok = typeSpec.Type.(*ast.StructType)
				if ok {
					// 创建结构体的零值并添加到结果中
					structName := typeSpec.Name.Name
					instance, err := createStructInstance(structName, container)
					if err != nil {
						fmt.Println(err.Error())
					}
					if instance != nil {
						structs = append(structs, instance)
					}
				}
			}
		}
		return nil
	})
	return structs, err
}

// createStructInstance 使用反射创建结构体的实例
func createStructInstance(structName string, container framework.Container) (interface{}, error) {
	typeRegister := container.MustMake(contract.TypeRegisterKey).(contract.TypeRegisterService)
	// 假设结构体已导入，使用反射获取类型
	structType, b := typeRegister.GetType(structName)
	if b {
		instance := reflect.New(structType).Interface()
		return instance, nil
	}
	return nil, errors.New(structName + "类型没有注册")
}

func AutoMigrateStructs(db *gorm.DB, pkgPath string, container framework.Container) error {
	// 扫描包路径下的所有结构体
	structs, err := ScanStructsInPackage(pkgPath, container)
	if err != nil {
		return err
	}
	// 使用 GORM 进行自动迁移
	for _, model := range structs {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("自动迁移失败: %v", err)
		}
		fmt.Printf("成功迁移模型：%T\n", model)
	}
	return nil
}
