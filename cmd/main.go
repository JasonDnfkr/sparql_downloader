package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("未找到.env文件")
		return
	}

	// 初始化数据库连接
	if err := initDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	filePath := "dblp/20250406_hci.tsv"
	download(filePath)

	// 调用parseTsv函数解析
	parse("dblp_hci_records", filePath)

	fmt.Println("数据已存储至 MongoDB")
}
