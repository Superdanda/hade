package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// GetExecDirectory 获取当前执行程序目录
func GetExecDirectory() string {
	file, err := os.Getwd()
	if err == nil {
		return file + "/"
	}
	return ""
}

// CheckProcessExist Will return true if the process with PID exists.
func CheckProcessExist(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false
	}
	return true
}

func KillProcess(pid int) error {
	process, _ := os.FindProcess(pid)
	if process != nil {
		if runtime.GOOS == "windows" {
			err := process.Kill()
			if err != nil {
				return err
			}
			fmt.Println("成功杀死进程，PID：" + strconv.Itoa(pid))
		} else {
			err := process.Signal(syscall.SIGTERM)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func FindProcessByPortAndKill(port string) {
	byPort, err := findProcessByPort(port)
	if err == nil {
		killProcess(byPort)
	}
}

// 查找占用指定端口的进程ID
func findProcessByPort(port string) (string, error) {
	var cmd *exec.Cmd

	// 根据操作系统选择命令
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", fmt.Sprintf("netstat -ano | findstr :%s", port))
	} else {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("lsof -i :%s -t", port))
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v", err)
	}

	// 获取 PID 并进行处理
	pid := strings.TrimSpace(out.String())
	if pid == "" {
		return "", fmt.Errorf("no process found on port %s", port)
	}

	// 对于 Windows，netstat 输出可能有多行，取第一行的 PID
	if runtime.GOOS == "windows" {
		lines := strings.Split(pid, "\n")
		fields := strings.Fields(lines[0])
		if len(fields) >= 5 {
			pid = fields[4]
		}
	}

	return pid, nil
}

// 杀死指定 PID 的进程
func killProcess(pid string) error {
	var cmd *exec.Cmd
	// 根据操作系统选择命令
	if runtime.GOOS == "windows" {
		cmd = exec.Command("taskkill", "/PID", pid, "/F")
	} else {
		cmd = exec.Command("kill", "-9", pid)
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to kill process %s: %v", pid, err)
	}
	return nil
}
