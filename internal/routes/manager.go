package routes

import (
	"fmt"
	"strconv"

	"github.com/youfun/fastcaddy/internal/api"
	"github.com/youfun/fastcaddy/internal/config"
	"github.com/youfun/fastcaddy/pkg/types"
)

// 常量定义 - 服务器和路由配置路径
const (
	ServersPath = "/apps/http/servers"
	RoutesPath  = ServersPath + "/srv0/routes"
)

// Manager 路由管理器 - 处理路由相关配置
type Manager struct {
	client        *api.Client
	configManager *config.Manager
}

// NewManager 创建新的路由管理器
func NewManager() *Manager {
	return &Manager{
		client:        api.NewClient(),
		configManager: config.NewManager(),
	}
}

// InitRoutes 初始化 HTTP 路由配置 - 对应 Python 的 init_routes(srv_name, skip) 函数
// 创建基础的 HTTP 服务器和路由配置
func (m *Manager) InitRoutes(serverName string, skip int) error {
	// 如果服务器路径已存在，直接返回
	if m.client.HasPath(ServersPath) {
		return nil
	}

	// 初始化服务器路径
	if err := m.configManager.InitPath(ServersPath, skip); err != nil {
		return err
	}

	// 创建基础 HTTP 服务器配置
	serverConfig := types.HTTPServer{
		Listen:    []string{":80", ":443"},           // 监听 HTTP 和 HTTPS 端口
		Routes:    []types.Route{},                   // 空路由列表
		Protocols: []string{"h1", "h2"},              // 支持 HTTP/1.1 和 HTTP/2
	}

	// 设置服务器配置
	serverPath := fmt.Sprintf("%s/%s", ServersPath, serverName)
	return m.client.PutConfig(serverConfig, serverPath, "POST")
}

// AddRoute 添加路由规则 - 对应 Python 的 add_route(route) 函数
// 将路由配置添加到 Caddy 服务器
func (m *Manager) AddRoute(route types.Route) error {
	return m.client.PutConfig(route, RoutesPath, "POST")
}

// DeleteByID 删除指定 ID 的路由 - 对应 Python 的 del_id(id) 函数
// 通过路由 ID 删除特定路由
func (m *Manager) DeleteByID(id string) error {
	return m.client.DeleteByID(id)
}

// AddReverseProxy 添加反向代理路由 - 对应 Python 的 add_reverse_proxy(from_host, to_url) 函数
// 创建从指定主机到目标 URL 的反向代理
func (m *Manager) AddReverseProxy(fromHost, toURL string) error {
	// 如果已存在相同主机的路由，先删除
	if m.client.HasID(fromHost) {
		if err := m.client.DeleteByID(fromHost); err != nil {
			return fmt.Errorf("删除现有路由失败: %w", err)
		}
	}

	// 创建反向代理路由配置
	route := types.Route{
		ID: fromHost,
		Handle: []types.Handler{
			{
				Handler: "reverse_proxy",
				Upstreams: []types.Upstream{
					{
						Dial: toURL,
					},
				},
			},
		},
		Match: []types.RouteMatch{
			{
				Host: []string{fromHost},
			},
		},
		Terminal: true, // 设置为终端路由
	}

	// 添加路由
	return m.AddRoute(route)
}

// AddWildcardRoute 添加通配符子域名路由 - 对应 Python 的 add_wildcard_route(domain) 函数
// 为指定域名创建通配符子域名路由
func (m *Manager) AddWildcardRoute(domain string) error {
	// 创建通配符路由配置
	route := types.Route{
		ID: fmt.Sprintf("wildcard-%s", domain),
		Match: []types.RouteMatch{
			{
				Host: []string{fmt.Sprintf("*.%s", domain)}, // 通配符匹配
			},
		},
		Handle: []types.Handler{
			{
				Handler: "subroute", // 使用子路由处理器
				Routes:  []types.Route{},
			},
		},
		Terminal: true,
	}

	// 添加路由
	return m.AddRoute(route)
}

// AddSubReverseProxy 添加子域名反向代理 - 对应 Python 的 add_sub_reverse_proxy 函数
// 为通配符域名下的特定子域名添加反向代理，支持多端口
func (m *Manager) AddSubReverseProxy(domain, subdomain string, ports []string, host string) error {
	wildcardID := fmt.Sprintf("wildcard-%s", domain)
	routeID := fmt.Sprintf("%s.%s", subdomain, domain)

	// 如果 host 为空，默认使用 localhost
	if host == "" {
		host = "localhost"
	}

	// 构建上游服务器列表
	var upstreams []types.Upstream
	for _, port := range ports {
		upstreams = append(upstreams, types.Upstream{
			Dial: fmt.Sprintf("%s:%s", host, port),
		})
	}

	// 创建子路由配置
	newRoute := types.Route{
		ID: routeID,
		Match: []types.RouteMatch{
			{
				Host: []string{routeID},
			},
		},
		Handle: []types.Handler{
			{
				Handler:   "reverse_proxy",
				Upstreams: upstreams,
			},
		},
	}

	// 将子路由添加到通配符路由的处理器中
	// 这里使用 "..." 语法来追加到现有路由列表
	subroutePath := fmt.Sprintf("%s/handle/0/routes/...", wildcardID)
	return m.client.PutByID([]types.Route{newRoute}, subroutePath, "POST")
}

// AddSubReverseProxyWithPorts 添加子域名反向代理（支持单个端口或端口列表）
// 这是一个便利方法，可以接受不同类型的端口参数
func (m *Manager) AddSubReverseProxyWithPorts(domain, subdomain string, ports interface{}, host string) error {
	var portList []string

	// 处理不同类型的端口参数
	switch v := ports.(type) {
	case string:
		portList = []string{v}
	case int:
		portList = []string{strconv.Itoa(v)}
	case []string:
		portList = v
	case []int:
		for _, port := range v {
			portList = append(portList, strconv.Itoa(port))
		}
	case []interface{}:
		for _, port := range v {
			switch p := port.(type) {
			case string:
				portList = append(portList, p)
			case int:
				portList = append(portList, strconv.Itoa(p))
			case float64: // JSON 数字默认解析为 float64
				portList = append(portList, strconv.Itoa(int(p)))
			}
		}
	default:
		return fmt.Errorf("不支持的端口类型: %T", ports)
	}

	return m.AddSubReverseProxy(domain, subdomain, portList, host)
}