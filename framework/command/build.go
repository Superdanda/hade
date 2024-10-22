package command

import (
	"fmt"
	"github.com/Superdanda/hade/framework/cobra"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func initBuildCommand() *cobra.Command {
	buildCommand.AddCommand(buildSelfCommand)
	buildCommand.AddCommand(buildBackendCommand)
	buildCommand.AddCommand(buildFrontendCommand)
	buildCommand.AddCommand(buildAllCommand)
	return buildCommand
}

var buildSelfCommand = &cobra.Command{
	Use:   "self",
	Short: "编译hade命令",
	RunE: func(c *cobra.Command, args []string) error {
		return goCommand.RunE(c, []string{"build"})
	},
}

var buildBackendCommand = &cobra.Command{
	Use:   "backend",
	Short: "使用go编译后端",
	RunE: func(c *cobra.Command, args []string) error {
		path, err := exec.LookPath("go")
		if err != nil {
			log.Fatalln("hade go: 请在Path路径中先安装go")
		}

		// 根据系统设置输出文件名
		output := "hade"
		if runtime.GOOS == "windows" {
			output += ".exe"
		}

		// 指定输出目录
		outputDir := filepath.Join(".")
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("无法创建输出目录: %v", err)
		}

		// 设置工作目录（避免权限问题）
		tempDir := filepath.Join(outputDir, "go-temp")
		err = os.MkdirAll(tempDir, 0755)
		if err != nil {
			return fmt.Errorf("无法创建临时目录: %v", err)
		}

		// 构建命令，设置环境变量
		cmd := exec.Command(path, "build", "-o", filepath.Join(outputDir, output), "./")
		cmd.Env = append(os.Environ(), "GOTMPDIR="+tempDir) // 设置临时目录

		// 执行编译命令并获取输出
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("go build error:")
			fmt.Println(string(out))
			fmt.Println("--------------")
			return err
		}
		fmt.Println(string(out))
		fmt.Println("编译hade成功")
		return nil
	},
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "编译相关命令",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			c.Help()
		}
		return nil
	},
}

var buildFrontendCommand = &cobra.Command{
	Use:   "frontend",
	Short: "使用npm编译前端",
	RunE: func(c *cobra.Command, args []string) error {
		err := npmCommand.RunE(c, []string{"run", "build"})
		if err != nil {
			fmt.Println("=============  前端编译失败 ============")
			return err
		}
		fmt.Println("=============  前端编译成功 ============")
		return nil
	},
}

var buildAllCommand = &cobra.Command{
	Use:   "all",
	Short: "同时编译前后端",
	RunE: func(c *cobra.Command, args []string) error {
		fmt.Println("=============  开始编译前后端 ============")
		err := buildFrontendCommand.RunE(c, args)
		if err != nil {
			return err
		}

		err = buildBackendCommand.RunE(c, args)
		if err != nil {
			return err
		}

		return nil
	},
}
