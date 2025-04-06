package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)



func trimQuotes(s string) string {
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) && len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	return s
}

// parse 解析TSV文件并返回解析后的记录
func parse(dbName string, filePath string, lineCount ...int) error {
	// 检查提供的文件路径是否存在
	if _, err := os.Stat(filePath); err != nil {
		fmt.Println("找不到TSV文件，请检查文件路径")
		return nil
	}

	fmt.Printf("开始解析文件: %s\n\n", filePath)

	// 处理
	err := doParse(dbName, filePath, lineCount...)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return nil
	}

	return nil
}

// doParse 读取TSV文件并返回标题和行数据
func doParse(dbName string, filePath string, lineCount ...int) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 连接数据库
	collection := DB.Collection(dbName)

	// 创建带缓冲的读取器
	scanner := bufio.NewScanner(file)

	// 设置较大的缓冲区以处理长行
	buffer := make([]byte, 64*1024)    // 64KB缓冲区
	scanner.Buffer(buffer, 10240*1024) // 最大行长度10MB

	// 读取并解析数据
	var headers []string
	
	var rowCount = 0

	// 确定要解析的行数
	maxLines := -1 // -1表示解析全部行
	if len(lineCount) > 0 && lineCount[0] > 0 {
		maxLines = lineCount[0]
	}

	alreadyInserted, err := checkAlreadyInserted(collection)
	if err!= nil {
		return fmt.Errorf("检查数据是否已插入失败: %v", err)
	}
	if alreadyInserted {
		fmt.Printf("本次数据已在先前执行插入")
		return nil
	}

	var counter int64 = 0

    // 启动打印监控协程
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            current := atomic.LoadInt64(&counter)
            fmt.Println("已处理: ", current)
        }
    }()

	var lines [][]string
	for scanner.Scan() && (maxLines == -1 || rowCount < maxLines) {
		line := scanner.Text()

		// 处理第一行（标题行）
		if rowCount == 0 {
			headers = strings.Split(line, "\t")
			fmt.Printf("列标题: %v\n\n", headers)
		} else {
			// 处理数据行
			fields := strings.Split(line, "\t")
			for i, field := range fields {
				fields[i] = trimQuotes(field)
			}
			lines = append(lines, fields)
		}

		rowCount++
		atomic.StoreInt64(&counter, int64(rowCount - 1))

		if len(lines) == 10000 {
			// 将行数据转换为记录结构体
			records := linesToRecords(headers, lines, filePath)
			// printRecords(records)

			// 插入记录到数据库
			err := insertRecords(collection, records)
			if err!= nil {
				return fmt.Errorf("插入数据失败: %v", err)	
			}

			// 清空行数据
			lines = lines[:0]
		}
	}

	if len(lines) > 0 {
		records := linesToRecords(headers, lines, filePath)
		// printRecords(records)

		// 插入记录到数据库
		err := insertRecords(collection, records)
		if err!= nil {
			return fmt.Errorf("插入数据失败: %v", err)	
		}
	}

	// 检查是否有扫描错误
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时出错: %v", err)
	}

	fmt.Printf("共解析了 %d 行数据\n", rowCount - 1)

	return nil
}

// linesToRecords 将行数据转换为记录结构体
func linesToRecords(headers []string, lines [][]string, source string) []*Record {
	var records []*Record

	for _, fields := range lines {
		// 创建记录
		record := &Record{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    source,
		}

		// 填充字段
		for j, field := range fields {
			if j < len(headers) {
				// 根据标题行的字段名称，将数据映射到对应的结构体字段
				switch headers[j] {
				case "?pub":
					record.Pub = field
				case "?title":
					record.Title = field
				case "?page":
					record.Page = field
					record.PageCount = calculatePageCount(field)
				case "?author":
					record.Author = field
				case "?creator":
					record.Creator = field
				case "?author_name":
					record.AuthorName = field
				case "?ordinal":
					record.Ordinal = field
				case "?stream":
					record.Stream = field
				case "?stream_name":
					record.StreamName = field
				case "?affiliation":
					record.Affiliation = field
				}
			}
		}

		records = append(records, record)
	}

	return records
}

// printRecords 打印记录信息
func printRecords(records []*Record) {
	for i, record := range records {
		fmt.Printf("Record %d:\n%+v\n", i, record)
		fmt.Println()
	}
}

func checkAlreadyInserted(collection *mongo.Collection) (bool, error) {
	count, err := collection.CountDocuments(context.TODO(), bson.D{})
	if err!= nil {
		fmt.Printf("获取文档数量失败: %v\n", err)
		return false, err
	}

	if count > 0 {
		fmt.Println("数据库中已存在数据, Abort.")
		return true, nil
	}

	return false, nil
}

func insertRecords(collection *mongo.Collection, records []*Record) error {
	// 插入记录到数据库
	var data []interface{}
	for _, record := range records {
		data = append(data, record)
	}
	_, err := collection.InsertMany(context.TODO(), data)
	if err!= nil {
		fmt.Printf("插入数据失败: %v\n", err)
		return err
	}

	return nil
}


func calculatePageCount(pageStr string) int {
	if strings.Contains(pageStr, "-") {
		// 形如361-363
		parts := strings.Split(pageStr, "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil {
				return end - start + 1
			}
		}
	} else if strings.Contains(pageStr, ":") {
		// 形如88:1-88:24
		parts := strings.Split(pageStr, "-")
		if len(parts) == 2 {
			startParts := strings.Split(parts[0], ":")
			endParts := strings.Split(parts[1], ":")
			if len(startParts) == 2 && len(endParts) == 2 {
				start, err1 := strconv.Atoi(startParts[1])
				end, err2 := strconv.Atoi(endParts[1])
				if err1 == nil && err2 == nil {
					return end - start + 1
				}
			}
		}
	} else {
		// 形如50
		if _, err := strconv.Atoi(pageStr); err == nil {
			return 1
		}
	}
	return 0
}