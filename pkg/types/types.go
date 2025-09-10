package types

// Caddy 配置结构 - 表示整个 Caddy 配置的顶层结构
type CaddyConfig struct {
	Apps map[string]interface{} `json:"apps"`
}

// 路由规则结构 - 定义单个路由规则
type Route struct {
	ID       string        `json:"@id,omitempty"`       // 路由唯一标识符
	Match    []RouteMatch  `json:"match"`               // 匹配条件列表
	Handle   []Handler     `json:"handle"`              // 处理器列表
	Terminal bool          `json:"terminal"`            // 是否为终端路由
}

// 路由匹配规则 - 定义路由匹配条件
type RouteMatch struct {
	Host []string `json:"host,omitempty"` // 主机名匹配列表
	Path []string `json:"path,omitempty"` // 路径匹配列表
}

// 处理器结构 - 定义路由处理逻辑
type Handler struct {
	Handler   string     `json:"handler"`              // 处理器类型 (如 "reverse_proxy", "subroute")
	Upstreams []Upstream `json:"upstreams,omitempty"`  // 上游服务器列表 (用于反向代理)
	Routes    []Route    `json:"routes,omitempty"`     // 子路由列表 (用于子路由处理器)
}

// 上游服务器 - 定义反向代理的目标服务器
type Upstream struct {
	Dial string `json:"dial"` // 目标服务器地址 (如 "localhost:8080")
}

// HTTP 服务器配置 - 定义 HTTP 服务器的配置
type HTTPServer struct {
	Listen    []string `json:"listen"`              // 监听地址列表
	Routes    []Route  `json:"routes"`              // 路由列表
	Protocols []string `json:"protocols,omitempty"` // 支持的协议列表
}

// TLS 自动化策略 - 定义 TLS 证书自动化策略
type TLSAutomationPolicy struct {
	Issuers []TLSIssuer `json:"issuers"` // 证书颁发者列表
}

// TLS 证书颁发者 - 定义证书颁发者配置
type TLSIssuer struct {
	Module     string                 `json:"module"`               // 颁发者模块类型 (如 "acme", "internal")
	Challenges map[string]interface{} `json:"challenges,omitempty"` // ACME 挑战配置
}

// ACME DNS 提供商配置 - 定义 DNS 挑战提供商
type ACMEProvider struct {
	Name     string `json:"name"`     // 提供商名称 (如 "cloudflare")
	APIToken string `json:"api_token"` // API 令牌
}

// PKI 配置 - 定义 PKI 证书颁发机构配置
type PKIConfig struct {
	InstallTrust bool `json:"install_trust"` // 是否安装信任根证书
}