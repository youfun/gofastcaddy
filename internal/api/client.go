package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client Caddy API 客户端 - 封装与 Caddy REST API 的交互
type Client struct {
	BaseURL    string       // Caddy API 基础 URL (默认: http://localhost:2019)
	HTTPClient *http.Client // HTTP 客户端
}

// NewClient 创建新的 Caddy API 客户端
func NewClient() *Client {
	return &Client{
		BaseURL: "http://localhost:2019",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetIDURL 根据路径生成 ID 的完整 URL - 用于通过 ID 访问配置
// 对应 Python 的 get_id(path) 函数
func (c *Client) GetIDURL(path string) string {
	// 确保路径以 '/' 开头和结尾
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return fmt.Sprintf("%s/id%s", c.BaseURL, path)
}

// GetConfigURL 根据路径生成配置的完整 URL - 用于访问配置路径
// 对应 Python 的 get_path(path) 函数
func (c *Client) GetConfigURL(path string) string {
	// 确保路径以 '/' 开头和结尾
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return fmt.Sprintf("%s/config%s", c.BaseURL, path)
}

// GetByID 通过 ID 获取配置 - 对应 Python 的 gid(path) 函数
func (c *Client) GetByID(path string) (map[string]interface{}, error) {
	url := c.GetIDURL(path)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取 ID 配置失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 ID 配置失败, 状态码: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应 JSON 失败: %w", err)
	}

	return result, nil
}

// GetConfig 获取指定路径的配置 - 对应 Python 的 gcfg(path, method) 函数
func (c *Client) GetConfig(path string) (map[string]interface{}, error) {
	url := c.GetConfigURL(path)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取配置失败, 状态码: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应 JSON 失败: %w", err)
	}

	return result, nil
}

// HasID 检查指定 ID 是否已设置 - 对应 Python 的 has_id(id) 函数
func (c *Client) HasID(id string) bool {
	_, err := c.GetByID(id)
	return err == nil
}

// HasPath 检查指定路径是否已设置 - 对应 Python 的 has_path(path) 函数
func (c *Client) HasPath(path string) bool {
	_, err := c.GetConfig(path)
	return err == nil
}

// PutByID 将配置数据放入指定 ID 路径 - 对应 Python 的 pid(d, path, method) 函数
func (c *Client) PutByID(data interface{}, path, method string) error {
	url := c.GetIDURL(path)
	return c.sendRequest(method, url, data)
}

// PutConfig 将配置数据放入指定配置路径 - 对应 Python 的 pcfg(d, path, method) 函数
func (c *Client) PutConfig(data interface{}, path, method string) error {
	url := c.GetConfigURL(path)
	return c.sendRequest(method, url, data)
}

// DeleteByID 删除指定 ID 的配置 - 对应 Python 的 del_id(id) 函数
func (c *Client) DeleteByID(id string) error {
	url := c.GetIDURL(id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("创建删除请求失败: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送删除请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("删除配置失败, 状态码: %d", resp.StatusCode)
	}

	return nil
}

// sendRequest 发送 HTTP 请求的通用方法 - 内部辅助函数
func (c *Client) sendRequest(method, url string, data interface{}) error {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("序列化请求数据失败: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(strings.ToUpper(method), url, body)
	if err != nil {
		return fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送 HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 尝试读取错误信息
		body, _ := io.ReadAll(resp.Body)
		var errorMsg map[string]interface{}
		if json.Unmarshal(body, &errorMsg) == nil {
			if errStr, ok := errorMsg["error"].(string); ok {
				return fmt.Errorf("请求失败, 状态码: %d, 错误: %s", resp.StatusCode, errStr)
			}
		}
		return fmt.Errorf("请求失败, 状态码: %d", resp.StatusCode)
	}

	return nil
}