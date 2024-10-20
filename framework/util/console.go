package util

import (
	"fmt"
	"strings"
)

// PrettyPrint 美观输出数组
func PrettyPrint(arr [][]string) {
	if len(arr) == 0 {
		return
	}

	rows := len(arr)
	cols := len(arr[0])

	// 计算每列的最大宽度
	colMaxs := make([]int, cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			cleanedStr := sanitizeString(arr[i][j])
			if len(cleanedStr) > colMaxs[j] {
				colMaxs[j] = len(cleanedStr)
			}
		}
	}

	// 打印数组，按列对齐
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			cleanedStr := sanitizeString(arr[i][j])
			fmt.Print(cleanedStr)
			// 添加适当的空格以对齐列
			padding := colMaxs[j] - len(cleanedStr) + 2
			fmt.Print(strings.Repeat(" ", padding))
		}
		fmt.Print("\n")
	}
}

// sanitizeString 清理字符串，去除多余空格和换行符
func sanitizeString(s string) string {
	return strings.TrimSpace(s)
}
