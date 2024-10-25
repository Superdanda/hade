package util

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Exists  判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// EnsureDir 如果文件夹不存在，就创建文件夹
func EnsureDir(path string) error {
	if !Exists(path) {
		fmt.Printf("Directory does not exist, creating: %s\n", path)
		// 使用 MkdirAll 递归创建路径，支持多层级路径
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// IsHiddenDirectory 路径是否是隐藏路径
func IsHiddenDirectory(path string) bool {
	return len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".")
}

// SubDir 输出所有子目录，目录名
func SubDir(folder string) ([]string, error) {
	subs, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for _, sub := range subs {
		if sub.IsDir() {
			ret = append(ret, sub.Name())
		}
	}
	return ret, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// 复制单个文件
func CopyFile(srcFile, dstFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// 递归复制整个目录
func CopyDir(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 目标路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dstDir, relPath)

		// 如果是目录，创建目标目录
		if info.IsDir() {
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				return err
			}
		} else {
			// 如果是文件，复制文件
			if err := CopyFile(path, destPath); err != nil {
				return err
			}
		}

		return nil
	})
}

func ReadFileToString(file string) (string, error) {
	// 读取 pid 文件内容
	data, err := os.ReadFile(file)
	if err != nil {
		return "", errors.Wrap(err, "读取文件失败: "+file)
	}
	return string(data), nil
}

func ReadFileToInt(file string) (int, error) {
	toString, _ := ReadFileToString(file)
	return strconv.Atoi(toString)
}
