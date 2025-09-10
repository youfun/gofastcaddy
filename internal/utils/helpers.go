package utils

import (
	"os"
	"strings"
)

// 常量定义 - 环境变量名
const (
	CloudflareTokenEnv = "CADDY_CF_TOKEN"    // Cloudflare API 令牌环境变量
	CloudflareAltEnv   = "CLOUDFLARE_API_TOKEN" // 备用 Cloudflare 令牌环境变量
)

// GetCloudflareToken 获取 Cloudflare API 令牌
// 从环境变量中获取 Cloudflare API 令牌，支持多个环境变量名
func GetCloudflareToken() string {
	// 首先尝试 CADDY_CF_TOKEN
	if token := os.Getenv(CloudflareTokenEnv); token != "" {
		return token
	}
	
	// 其次尝试 CLOUDFLARE_API_TOKEN
	if token := os.Getenv(CloudflareAltEnv); token != "" {
		return token
	}
	
	return ""
}

// NormalizePath 规范化路径格式
// 确保路径以 '/' 开头和结尾
func NormalizePath(path string) string {
	if path == "" {
		return "/"
	}
	
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	
	return path
}

// CleanPath 清理路径格式
// 确保路径以 '/' 开头但不以 '/' 结尾（除非是根路径）
func CleanPath(path string) string {
	if path == "" || path == "/" {
		return "/"
	}
	
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	path = strings.TrimSuffix(path, "/")
	
	return path
}

// SplitPath 分割路径为组件
// 将路径按 '/' 分割，忽略空组件
func SplitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

// JoinPath 连接路径组件
// 将路径组件用 '/' 连接成完整路径
func JoinPath(components ...string) string {
	if len(components) == 0 {
		return "/"
	}
	
	var validComponents []string
	for _, comp := range components {
		comp = strings.Trim(comp, "/")
		if comp != "" {
			validComponents = append(validComponents, comp)
		}
	}
	
	if len(validComponents) == 0 {
		return "/"
	}
	
	return "/" + strings.Join(validComponents, "/")
}

// ValidateHost 验证主机名格式
// 检查主机名是否符合基本格式要求
func ValidateHost(host string) bool {
	if host == "" {
		return false
	}
	
	// 基本验证：不能包含空格、斜杠等
	if strings.ContainsAny(host, " /\\") {
		return false
	}
	
	return true
}

// ValidateURL 验证 URL 格式
// 检查 URL 是否包含主机和端口信息
func ValidateURL(url string) bool {
	if url == "" {
		return false
	}
	
	// 基本验证：应该包含 ':' 分隔符
	if !strings.Contains(url, ":") {
		return false
	}
	
	return true
}

// DefaultIfEmpty 如果值为空则返回默认值
func DefaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// StringSliceContains 检查字符串切片是否包含指定值
func StringSliceContains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// MergeStringMaps 合并多个字符串映射
// 后面的映射会覆盖前面映射中的相同键
func MergeStringMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}