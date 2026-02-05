package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"gin-web/config"
)

// 使用示例
type MyAPIResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	//Message string      `json:"message"`
}

// 主结构体
type LogParams struct {
	Data struct {
		ReqID    string `json:"req_id"`
		Name     string `json:"name"`
		LogLevel int    `json:"log_level"`
		Detail   string `json:"detail"`
		Custom1  string `json:"custom1,omitempty"` // omitempty 表示字段为空时不输出
		Custom2  string `json:"custom2,omitempty"`
	} `json:"data"`
	TableName string `json:"table_name"`
}

func LogClient(baseURL string) *BaseClient {
	return &BaseClient{
		client:    &http.Client{},
		baseURL:   baseURL,
		method:    GET,
		timeout:   3 * time.Second,
		paramType: Query,
		headers:   make(map[string]string),
		unpacker:  logUnpacker,
		processor: logProcessor,
		//dataWrapper:   defaultDataWrapper,
		errorHandler:  logErrorHandler,
		requestIDFunc: logRequestID,
	}
}

// 默认配置项
func logUnpacker(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func logProcessor(data interface{}) interface{} {
	return data
}

func logErrorHandler(err error) error {
	return fmt.Errorf("request error: %w", err)
}

func logRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// SendTableStoreLog 发送日志到 TableStore（需要通过依赖注入获取 cfg 和 log）
func SendTableStoreLog(cfg *config.Configuration, log *zap.Logger, params any) {
	client := LogClient(GetApiUrl(cfg, "log_url")).
		WithMethod(POST).
		WithURI("insert").
		WithParamType(JSON).
		WithHeader("Content-Type", "application/json").
		WithTimeout(5 * time.Second).
		WithLogger(log)

	if client == nil {
		log.Error("Failed to create client")
		fmt.Println("Failed to create client")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	response, err := client.Exec(ctx, params)
	// 处理错误
	if err != nil {
		log.Error("Failed to send request", zap.Error(err),
			zap.Any("params", params))
	}
	// 处理响应
	if response != nil {
		log.Info("SendTableStoreLog response", zap.Any("data", response))
	}
}
