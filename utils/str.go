package utils

import (
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

func RandString(len int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

func GetConfigKeyFromFilename(filePath string) string {
	base := filepath.Base(filePath)                      // 获取 logConsumer.go
	name := strings.TrimSuffix(base, filepath.Ext(base)) // 去除扩展名 -> logConsumer
	return strings.TrimRight(name, "Consumer")           // 提取 log
}
