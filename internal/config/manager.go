package config

import (
	"github.com/youfun/fastcaddy/internal/api"
	"strings"
)

// Manager 配置管理器 - 提供配置操作的高级接口
type Manager struct {
	client *api.Client
}

// NewManager 创建新的配置管理器
func NewManager() *Manager {
	return &Manager{
		client: api.NewClient(),
	}
}

// NestedSetDict 在嵌套字典中设置值 - 对应 Python 的 nested_setdict(sd, value, *keys) 函数
// 返回更新后的字典，其中在指定键路径处设置了值
func NestedSetDict(dict map[string]interface{}, value interface{}, keys ...string) map[string]interface{} {
	if len(keys) == 0 {
		return dict
	}

	// 确保字典不为 nil
	if dict == nil {
		dict = make(map[string]interface{})
	}

	// 遍历除最后一个键外的所有键，创建嵌套路径
	current := dict
	for _, key := range keys[:len(keys)-1] {
		if current[key] == nil {
			current[key] = make(map[string]interface{})
		}
		// 类型断言，确保是 map 类型
		if nested, ok := current[key].(map[string]interface{}); ok {
			current = nested
		} else {
			// 如果不是 map 类型，创建新的 map
			current[key] = make(map[string]interface{})
			current = current[key].(map[string]interface{})
		}
	}

	// 设置最后一个键的值
	current[keys[len(keys)-1]] = value
	return dict
}

// PathToKeys 将路径分割为键列表 - 对应 Python 的 path2keys(path) 函数
// 按 '/' 分割路径并返回键的切片
func PathToKeys(path string) []string {
	// 去除首尾的斜杠并按 '/' 分割
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

// KeysToPath 将键列表连接为路径 - 对应 Python 的 keys2path(*keys) 函数
// 将键列表用 '/' 连接成路径字符串
func KeysToPath(keys ...string) string {
	if len(keys) == 0 {
		return "/"
	}
	return "/" + strings.Join(keys, "/")
}

// NestedSetConfig 在配置中设置嵌套值 - 对应 Python 的 nested_setcfg(value, *keys) 函数
// 获取当前配置，更新嵌套值，然后保存回去
func (m *Manager) NestedSetConfig(value interface{}, keys ...string) error {
	// 获取当前配置
	config, err := m.client.GetConfig("/")
	if err != nil {
		return err
	}

	// 在配置中设置嵌套值
	updatedConfig := NestedSetDict(config, value, keys...)

	// 保存更新后的配置
	return m.client.PutConfig(updatedConfig, "/", "POST")
}

// InitPath 初始化配置路径 - 对应 Python 的 init_path(path, skip) 函数
// 逐步创建路径中的每个层级，跳过指定数量的初始层级
func (m *Manager) InitPath(path string, skip int) error {
	keys := PathToKeys(path)
	var currentKeys []string

	// 遍历路径中的每个部分
	for i, key := range keys {
		currentKeys = append(currentKeys, key)
		
		// 如果当前索引小于跳过数量，则继续下一个
		if i < skip {
			continue
		}

		// 为当前路径创建空配置
		currentPath := KeysToPath(currentKeys...)
		emptyConfig := make(map[string]interface{})
		
		if err := m.client.PutConfig(emptyConfig, currentPath, "POST"); err != nil {
			return err
		}
	}

	return nil
}

// GetClient 获取底层 API 客户端 - 提供对原始 API 的访问
func (m *Manager) GetClient() *api.Client {
	return m.client
}