package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gin-web/global"
	//"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	_ "net/url"
	"reflect"
	_ "reflect"
	"strings"
	_ "strings"
	"time"
)

// Method 请求方法枚举
type Method string

const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	DELETE Method = "DELETE"
)

// ParamType 参数类型枚举
type ParamType int

const (
	Query ParamType = iota
	JSON
	Form
)

//// Response 统一响应接口 //先不统一了
//type Response interface {
//	IsSuccess() bool
//	GetCode() int
//	GetMessage() string
//	GetData() interface{}
//}

// HttpClient 抽象HTTP客户端接口
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// BaseClient 请求基类
type BaseClient struct {
	client    HttpClient
	baseURL   string
	method    Method
	uri       string
	timeout   time.Duration
	paramType ParamType
	headers   map[string]string
	unpacker  Unpacker
	processor ResponseProcessor
	//dataWrapper   DataWrapper
	errorHandler  ErrorHandler
	requestIDFunc func() string
}

// Unpacker 响应解码器类型
type Unpacker func([]byte) (interface{}, error)

// ResponseProcessor 响应处理器类型
type ResponseProcessor func(interface{}) interface{}

// DataWrapper 数据包装器类型
type DataWrapper func(interface{}) interface{}

// ErrorHandler 错误处理器类型
type ErrorHandler func(error) error

//// Logger 日志记录接口
//type Logger interface {
//	LogRequest(duration time.Duration, url string, start, end time.Time)
//}

// NewBaseClient 构造函数
func NewBaseClient(baseURL string) *BaseClient {
	return &BaseClient{
		client:    &http.Client{},
		baseURL:   baseURL,
		method:    GET,
		timeout:   3 * time.Second,
		paramType: Query,
		headers:   make(map[string]string),
		unpacker:  defaultUnpacker,
		processor: defaultProcessor,
		//dataWrapper:   defaultDataWrapper,
		errorHandler:  defaultErrorHandler,
		requestIDFunc: defaultRequestID,
	}
}

// 默认配置项
func defaultUnpacker(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func defaultProcessor(data interface{}) interface{} {
	return data
}

//func defaultDataWrapper(data interface{}) interface{} {
//	return map[string]interface{}{
//		"code":    0,
//		"message": "success",
//		"data":    data,
//	}
//}

func defaultErrorHandler(err error) error {
	return fmt.Errorf("request error: %w", err)
}

func defaultRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Exec 执行请求
func (c *BaseClient) Exec(ctx context.Context, params interface{}) (interface{}, error) {
	//start := time.Now()
	defer func() {
		//c.logger.LogRequest(time.Since(start), c.fullURL(), start, time.Now())
		global.App.Log.Info("Request end!!!!!:")
	}()

	req, err := c.buildRequest(ctx, params)
	if err != nil {
		return nil, c.errorHandler(err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, c.errorHandler(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, c.errorHandler(err)
	}

	data, err := c.unpacker(body)
	if err != nil {
		return nil, c.errorHandler(err)
	}

	processed := c.processor(data)
	return processed, nil
}

// 构建请求对象
func (c *BaseClient) buildRequest(ctx context.Context, params interface{}) (*http.Request, error) {
	url := c.fullURL()

	var req *http.Request
	var err error

	switch c.method {
	case GET:
		req, err = http.NewRequestWithContext(ctx, string(c.method), url, nil)
		if err == nil {
			err := c.addQueryParams(req, params)
			if err != nil {
				return nil, err
			}
		}
	default:
		body, err := c.encodeParams(params)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequestWithContext(ctx, string(c.method), url, body)
	}

	if err != nil {
		return nil, err
	}

	c.addHeaders(req)
	return req, nil
}

// 添加请求头
func (c *BaseClient) addHeaders(req *http.Request) {
	req.Header.Add("X-Request-ID", c.requestIDFunc())
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}
}

// 编码参数
func (c *BaseClient) encodeParams(params interface{}) (io.Reader, error) {
	switch c.paramType {
	case JSON:
		data, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(data), nil
	// 其他参数类型处理...
	default:
		return nil, fmt.Errorf("unsupported param type")
	}
}

// 完整URL
func (c *BaseClient) fullURL() string {
	return c.baseURL + c.uri
}

// 链式调用配置方法
func (c *BaseClient) WithMethod(method Method) *BaseClient {
	c.method = method
	return c
}

func (c *BaseClient) WithURI(uri string) *BaseClient {
	c.uri = uri
	return c
}

func (c *BaseClient) WithTimeout(timeout time.Duration) *BaseClient {
	c.timeout = timeout
	return c
}

func (c *BaseClient) WithParamType(paramType ParamType) *BaseClient {
	c.paramType = paramType
	return c
}

func (c *BaseClient) WithHeader(key, value string) *BaseClient {
	c.headers[key] = value
	return c
}

func (c *BaseClient) WithUnpacker(fn Unpacker) *BaseClient {
	c.unpacker = fn
	return c
}

func (c *BaseClient) WithProcessor(fn ResponseProcessor) *BaseClient {
	c.processor = fn
	return c
}

// addQueryParams 添加URL查询参数（GET请求专用）
func (c *BaseClient) addQueryParams(req *http.Request, params interface{}) error {
	query := req.URL.Query()

	// 处理不同类型的参数
	switch v := params.(type) {
	case url.Values:
		for key, values := range v {
			for _, value := range values {
				query.Add(key, value)
			}
		}
	case map[string]string:
		for key, value := range v {
			query.Add(key, value)
		}
	case map[string][]string:
		for key, values := range v {
			for _, value := range values {
				query.Add(key, value)
			}
		}
	case map[string]interface{}:
		for key, value := range v {
			query.Add(key, fmt.Sprintf("%v", value))
		}
	default:
		// 通过反射处理结构体
		val := reflect.ValueOf(params)
		if val.Kind() == reflect.Struct {
			typ := val.Type()
			for i := 0; i < val.NumField(); i++ {
				field := typ.Field(i)
				// 优先读取query tag，其次form tag，最后字段名
				key := field.Tag.Get("query")
				if key == "" {
					key = field.Tag.Get("form")
				}
				if key == "" {
					key = strings.ToLower(field.Name)
				}
				value := val.Field(i).Interface()
				query.Add(key, fmt.Sprintf("%v", value))
			}
		} else {
			return fmt.Errorf("unsupported params type: %T", params)
		}
	}

	req.URL.RawQuery = query.Encode()
	return nil
}
