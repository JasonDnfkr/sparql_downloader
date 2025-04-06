package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadTSV 读取并解析TSV文件，如果filePath为空，则使用默认路径
// 参数:
//   - filePath: TSV文件的路径，如果为空字符串则使用默认路径
//
// 返回:
//   - 无
func ReadTSV(filePath string) {
	// 如果未提供文件路径，则使用默认路径
	if filePath == "" {
		// 获取TSV文件路径 - 使用绝对路径
		execDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("无法获取当前工作目录: %v\n", err)
			return
		}

		// 构建文件路径
		filePath = filepath.Join(execDir, "dblp", "20250406_mini.tsv")
		fmt.Printf("尝试读取文件: %s\n", filePath)

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// 如果文件不存在，尝试向上一级目录查找
			parentDir := filepath.Dir(execDir)
			filePath = filepath.Join(parentDir, "dblp", "20250406_mini.tsv")
			fmt.Printf("尝试备选路径: %s\n", filePath)
		}
	} else {
		fmt.Printf("使用指定文件路径: %s\n", filePath)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		return
	}
	defer file.Close()

	// 创建一个带缓冲的读取器，提高读取效率
	reader := bufio.NewReader(file)

	// 读取并解析前100行
	lineCount := 0
	headers := []string{}

	fmt.Println("开始解析TSV文件...")

	for lineCount < 100 {
		// 读取一行
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				// 处理最后一行（可能没有换行符）
				if line != "" {
					processLine(line, lineCount, headers)
					lineCount++
				}
				break
			}
			fmt.Printf("读取行时出错: %v\n", err)
			break
		}

		// 去除行尾的换行符
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		// 处理行
		processLine(line, lineCount, headers)

		// 如果是第一行，保存列标题
		if lineCount == 0 {
			headers = strings.Split(line, "\t")
		}

		lineCount++
	}

	fmt.Printf("共解析了 %d 行数据\n", lineCount)
}

// 处理单行数据
func processLine(line string, lineNum int, headers []string) {
	// 按制表符分割字段
	fields := strings.Split(line, "\t")

	// 打印行号和内容
	fmt.Printf("行 %d:\n", lineNum+1)

	// 如果是第一行（标题行），直接打印
	if lineNum == 0 {
		fmt.Println(line)
		return
	}

	// 对于数据行，打印每个字段及其对应的标题
	for i, field := range fields {
		if i < len(headers) {
			fmt.Printf("  %s: %s\n", headers[i], field)
		} else {
			fmt.Printf("  字段%d: %s\n", i+1, field)
		}
	}
	fmt.Println()
}
