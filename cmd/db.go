package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func initDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建 MongoDB 连接 URI
	dbName := getEnv("DB_NAME", "venuelens")
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", "your_password"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "27017"),
		dbName,
	)

	// 连接 MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	Client = client
	DB = client.Database(getEnv("DB_NAME", "venuelens"))
	log.Println("MongoDB 连接成功")

	return nil
}

func CloseDB() error {
	if Client != nil {
		return Client.Disconnect(context.Background())
	}
	return nil
}

// 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

