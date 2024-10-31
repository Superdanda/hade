package request

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// DefaultClient 默认定义一个60秒超时的请求
var DefaultClient = NewClient(60 * time.Second)

// Client 定义了一个包含 http.Client 的结构体，可以设置超时时间等参数
type Client struct {
	httpClient *http.Client
}

type Headers = map[string]string

// NewClient 返回一个新的 Client 实例，可以设置请求超时时间
func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Request 封装了 HTTP 请求，返回响应的字节数组
func (c *Client) Request(method, url string, headers map[string]string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	if headers != nil && len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 发起请求
	resp, err := DefaultClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 读取响应体
	return io.ReadAll(resp.Body)
}

// requestJSON 使用泛型，将响应解析为指定的类型 T
func requestJSON[T any](method, url string, headers map[string]string, body io.Reader) (T, error) {
	var result T
	// 发起请求，获取字节数组
	data, err := DefaultClient.Request(method, url, headers, body)
	if err != nil {
		return result, err
	}
	// 解析 JSON 到指定类型
	err = json.Unmarshal(data, &result)
	return result, err
}

// Get 快捷 GET 请求，返回字节数组
func Get(url string, headers Headers) ([]byte, error) {
	return DefaultClient.Request(http.MethodGet, url, headers, nil)
}

// GetJSON 快捷 GET 请求，使用泛型返回指定类型
func GetJSON[T any](url string, headers Headers) (T, error) {
	return requestJSON[T](http.MethodGet, url, headers, nil)
}

// Post 快捷 POST 请求，返回字节数组
func Post(url string, headers Headers, body interface{}) ([]byte, error) {
	jsonBody, err := JSONBody(body)
	if err != nil {
		return nil, err
	}
	return DefaultClient.Request(http.MethodPost, url, headers, jsonBody)
}

// PostJSON 快捷 POST 请求，使用泛型返回指定类型
func PostJSON[T any](url string, headers Headers, body interface{}) (T, error) {
	var result T
	jsonBody, err := JSONBody(body)
	if err != nil {
		return result, err
	}
	return requestJSON[T](http.MethodPost, url, headers, jsonBody)
}

// JSONBody 接受任意结构体数据，将其序列化为 JSON 格式并作为请求体
func JSONBody(data interface{}) (*bytes.Buffer, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(jsonBytes), nil
}

// RequestBuilder 定义了一个请求构建器，支持链式调用
type RequestBuilder struct {
	client  *Client
	method  string
	url     string
	headers Headers
	body    io.Reader
}

// NewRequest 创建一个新的 RequestBuilder 实例
func NewRequest() *RequestBuilder {
	return &RequestBuilder{
		client:  DefaultClient,
		headers: make(map[string]string),
	}
}

// WithClient 设置自定义的 Client
func (rb *RequestBuilder) WithClient(client *Client) *RequestBuilder {
	rb.client = client
	return rb
}

// Get 设置请求方法为 GET，并指定 URL
func (rb *RequestBuilder) Get(url string) *RequestBuilder {
	rb.method = http.MethodGet
	rb.url = url
	return rb
}

// Post 设置请求方法为 POST，并指定 URL
func (rb *RequestBuilder) Post(url string) *RequestBuilder {
	rb.method = http.MethodPost
	rb.url = url
	return rb
}

// Header 添加一个请求头
func (rb *RequestBuilder) Header(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

// Headers 批量添加请求头，接受任意数量的键值对
func (rb *RequestBuilder) Headers(headers map[string]string) *RequestBuilder {
	for k, v := range headers {
		rb.headers[k] = v
	}
	return rb
}

// JSONBody 设置请求体为 JSON 格式，接受任意可被 json.Marshal 的类型
func (rb *RequestBuilder) JSONBody(body interface{}) *RequestBuilder {
	data, _ := json.Marshal(body)
	rb.body = bytes.NewBuffer(data)
	rb.Header("Content-Type", "application/json")
	return rb
}

// Body 设置请求体，接受一个 io.Reader
func (rb *RequestBuilder) Body(body io.Reader) *RequestBuilder {
	rb.body = body
	return rb
}

// Do 发送请求，返回响应的字节数组
func (rb *RequestBuilder) Do() ([]byte, error) {
	req, err := http.NewRequest(rb.method, rb.url, rb.body)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for key, value := range rb.headers {
		req.Header.Set(key, value)
	}

	// 发起请求
	resp, err := rb.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	return io.ReadAll(resp.Body)
}
