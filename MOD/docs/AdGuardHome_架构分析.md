# AdGuard Home 架构分析文档

## 目录

1. [项目概述](#项目概述)
2. [技术栈](#技术栈)
3. [整体架构](#整体架构)
4. [核心模块](#核心模块)
5. [数据流](#数据流)
6. [关键设计](#关键设计)
7. [部署架构](#部署架构)

---

## 项目概述

AdGuard Home 是一个网络范围的广告和跟踪器拦截 DNS 服务器。它作为 DNS 服务器运行,可以保护家庭网络中的所有设备,无需在每个设备上安装客户端软件。

### 核心功能

- **DNS 过滤**: 基于规则列表拦截广告和跟踪器域名
- **安全浏览**: 阻止恶意和钓鱁网站
- **家长控制**: 过滤不适合儿童的内容
- **安全搜索**: 强制搜索引擎使用安全搜索模式
- **DHCP 服务器**: 内置 DHCP 服务器功能
- **查询日志**: 记录所有 DNS 查询
- **统计分析**: DNS 使用情况统计
- **DNS 重写**: 自定义 DNS 响应
- **多协议支持**: DNS-over-HTTPS (DoH), DNS-over-TLS (DoT), DNS-over-QUIC (DoQ), DNSCrypt

### 项目特点

- 跨平台支持 (Linux, macOS, Windows, FreeBSD, OpenBSD)
- 单一二进制文件部署
- Web 管理界面
- 支持 Docker 和 Snap 部署
- 开源免费

---

## 技术栈

### 后端技术

**编程语言**: Go 1.25.4

**核心依赖库**:

```
DNS 处理:
- github.com/AdguardTeam/dnsproxy v0.77.0        # DNS 代理核心
- github.com/miekg/dns v1.1.68                   # DNS 协议实现
- github.com/AdguardTeam/urlfilter v0.22.1       # URL/域名过滤引擎

网络库:
- github.com/quic-go/quic-go v0.55.0             # QUIC 协议支持
- golang.org/x/net                                # 网络扩展库
- github.com/mdlayher/netlink                     # Linux netlink 接口

DHCP:
- github.com/insomniacslk/dhcp                    # DHCP 协议实现

存储:
- go.etcd.io/bbolt v1.4.3                        # 嵌入式键值数据库

工具库:
- github.com/AdguardTeam/golibs v0.35.2          # AdGuard 通用库
- github.com/google/uuid                          # UUID 生成
- gopkg.in/natefinch/lumberjack.v2               # 日志轮转
```

### 前端技术

**技术栈**:
- React.js - UI 框架
- Tabler - UI 组件库
- Webpack - 构建工具
- TypeScript - 类型支持
- Playwright - E2E 测试

**构建系统**:
- Node.js v24.10.0+
- npm v10.8+

---

## 整体架构

### 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端设备                                │
│  (电脑、手机、智能电视、IoT 设备等)                                │
└────────────────────┬────────────────────────────────────────────┘
                     │ DNS 查询
                     │ (UDP/TCP/DoH/DoT/DoQ/DNSCrypt)
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                    AdGuard Home 服务器                            │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Web 管理界面 (HTTP/HTTPS)                    │   │
│  │  - 配置管理  - 统计查看  - 日志查询  - 过滤规则管理       │   │
│  └─────────────────────────────────────────────────────────┘   │
│                            │                                     │
│  ┌─────────────────────────┴─────────────────────────────┐     │
│  │                   核心控制层 (home)                      │     │
│  │  - 配置管理  - 认证授权  - TLS 管理  - 更新管理         │     │
│  └─────────────────────────┬─────────────────────────────┘     │
│                            │                                     │
│  ┌────────────────┬────────┴────────┬──────────────────┐       │
│  │                │                 │                   │       │
│  ▼                ▼                 ▼                   ▼       │
│ ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐   │
│ │DNS转发器  │  │过滤引擎   │  │DHCP服务器│  │客户端管理     │   │
│ │dnsforward│  │filtering │  │dhcpd/    │  │clients       │   │
│ │          │  │          │  │dhcpsvc   │  │              │   │
│ └────┬─────┘  └────┬─────┘  └──────────┘  └──────────────┘   │
│      │            │                                            │
│      │            ├─ 规则列表 (rulelist)                       │
│      │            ├─ 安全浏览 (hashprefix)                     │
│      │            ├─ 家长控制 (hashprefix)                     │
│      │            ├─ 安全搜索 (safesearch)                     │
│      │            └─ DNS 重写 (rewrite)                        │
│      │                                                          │
│  ┌───┴────────────┴───────────────────────────────────┐       │
│  │              支持服务层                              │       │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────┐     │       │
│  │  │查询日志   │  │统计模块   │  │系统集成       │     │       │
│  │  │querylog  │  │stats     │  │arpdb/rdns/   │     │       │
│  │  │          │  │          │  │whois/ipset   │     │       │
│  │  └──────────┘  └──────────┘  └──────────────┘     │       │
│  └──────────────────────────────────────────────────┘       │
│                                                               │
└───────────────────────────────────────────────────────────────┘
                     │
                     │ 上游 DNS 查询
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    上游 DNS 服务器                            │
│  (Google DNS, Cloudflare, Quad9, 自定义 DNS 等)              │
└─────────────────────────────────────────────────────────────┘
```



### 模块关系图

```
┌──────────────────────────────────────────────────────────────┐
│                         main.go                               │
│                    (应用程序入口)                              │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────────┐
│                    internal/home                              │
│                  (全局上下文管理)                              │
│                                                                │
│  homeContext {                                                │
│    clients    clientsContainer                                │
│    stats      stats.Interface                                 │
│    queryLog   querylog.QueryLog                               │
│    dnsServer  *dnsforward.Server                              │
│    dhcpServer dhcpd.Interface                                 │
│    filters    *filtering.DNSFilter                            │
│    web        *webAPI                                         │
│    etcHosts   *aghnet.HostsContainer                          │
│  }                                                             │
└────────────────────────────────────────────────────────────────┘
```

---

## 核心模块

### 1. DNS 转发模块 (dnsforward)

**位置**: `internal/dnsforward/`

**职责**: DNS 请求处理和转发的核心模块

**核心组件**:

```go
type Server struct {
    // DNS 代理
    dnsProxy *proxy.Proxy
    internalProxy *proxy.Proxy
    
    // 过滤器
    dnsFilter *filtering.DNSFilter
    access *accessManager
    
    // 客户端处理
    addrProc client.AddressProcessor
    dhcpServer DHCP
    
    // 查询日志和统计
    queryLog querylog.QueryLog
    stats stats.Interface
    
    // 配置
    conf ServerConfig
    privateNets netutil.SubnetSet
    dns64Pref netip.Prefix
}
```

**主要功能**:

1. **请求处理流程**:
   - 接收 DNS 请求 (UDP/TCP/DoH/DoT/DoQ/DNSCrypt)
   - 客户端识别和认证
   - 访问控制检查
   - 应用过滤规则
   - 转发到上游 DNS 服务器
   - 处理响应并返回客户端

2. **协议支持**:
   - Plain DNS (UDP/TCP, 端口 53)
   - DNS-over-HTTPS (DoH, 端口 443)
   - DNS-over-TLS (DoT, 端口 853)
   - DNS-over-QUIC (DoQ, 端口 784)
   - DNSCrypt

3. **特殊功能**:
   - DNS64 支持 (IPv6 到 IPv4 映射)
   - EDNS Client Subnet (ECS)
   - DNSSEC 验证
   - 响应缓存
   - 速率限制

**关键文件**:
- `dnsforward.go` - 服务器主逻辑
- `process.go` - 请求处理流程
- `filter.go` - 过滤逻辑集成
- `access.go` - 访问控制
- `upstreams.go` - 上游服务器管理

---

### 2. 过滤引擎 (filtering)

**位置**: `internal/filtering/`

**职责**: DNS 请求和响应的过滤处理

**核心组件**:

```go
type DNSFilter struct {
    // 过滤引擎
    filteringEngine *urlfilter.DNSEngine
    filteringEngineAllow *urlfilter.DNSEngine
    rulesStorage *filterlist.RuleStorage
    rulesStorageAllow *filterlist.RuleStorage
    
    // 安全服务
    safeBrowsingChecker Checker
    parentalControlChecker Checker
    safeSearch SafeSearch
    
    // 配置
    conf *Config
    hostCheckers []hostChecker
}
```

**过滤类型**:

1. **规则列表过滤**:
   - Adblock 风格规则
   - Hosts 文件格式
   - 支持通配符和正则表达式
   - 黑名单和白名单

2. **安全浏览** (Safe Browsing):
   - 基于哈希前缀的检查
   - 阻止恶意和钓鱼网站
   - 使用 AdGuard DNS 服务

3. **家长控制** (Parental Control):
   - 过滤成人内容
   - 基于哈希前缀检查
   - 可配置阻止级别

4. **安全搜索** (Safe Search):
   - 强制搜索引擎安全模式
   - 支持 Google, Bing, Yandex 等
   - CNAME 重写实现

5. **服务阻止** (Blocked Services):
   - 快速阻止流行服务
   - 预定义规则集
   - 全局或按客户端配置

6. **DNS 重写** (Rewrites):
   - 自定义 DNS 响应
   - A/AAAA/CNAME 记录
   - 支持通配符

**过滤流程**:

```
DNS 请求
    │
    ├─> 1. DNS 重写检查
    │       └─> 匹配则返回自定义响应
    │
    ├─> 2. /etc/hosts 检查
    │       └─> 匹配则返回本地解析
    │
    ├─> 3. 规则列表匹配
    │       ├─> 白名单规则 → 允许
    │       └─> 黑名单规则 → 阻止
    │
    ├─> 4. 服务阻止检查
    │       └─> 匹配则阻止
    │
    ├─> 5. 安全浏览检查
    │       └─> 恶意网站 → 阻止
    │
    ├─> 6. 家长控制检查
    │       └─> 不适内容 → 阻止
    │
    └─> 7. 安全搜索检查
            └─> 搜索引擎 → 重写为安全版本
```

**子模块**:

- `rulelist/` - 规则列表管理
- `hashprefix/` - 哈希前缀检查 (安全浏览/家长控制)
- `safesearch/` - 安全搜索实现
- `rewrite/` - DNS 重写功能

---

### 3. DHCP 服务器 (dhcpd/dhcpsvc)

**位置**: `internal/dhcpd/`, `internal/dhcpsvc/`

**职责**: DHCP 服务器功能

**说明**: 项目中存在两个 DHCP 实现:
- `dhcpd/` - 旧版实现 (Unix 系统)
- `dhcpsvc/` - 新版实现 (正在迁移)

**核心功能**:

1. **DHCPv4 支持**:
   - IP 地址分配
   - 静态租约
   - 自定义选项
   - 租约管理

2. **DHCPv6 支持**:
   - IPv6 地址分配
   - 前缀委派
   - 路由器通告 (RA)
   - SLAAC 支持

3. **集成功能**:
   - 与 DNS 集成 (主机名解析)
   - 客户端识别
   - 租约持久化

**关键特性**:

```go
type DHCPServer interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    
    // 租约管理
    AddStaticLease(lease *Lease) error
    RemoveStaticLease(lease *Lease) error
    
    // 查询
    Leases() []*Lease
    FindMACbyIP(ip netip.Addr) (mac net.HardwareAddr)
    
    // 与 DNS 集成
    HostByIP(ip netip.Addr) (host string)
    IPByHost(host string) (ip netip.Addr)
}
```

---

### 4. 客户端管理 (client)

**位置**: `internal/client/`

**职责**: 客户端识别、配置和管理

**核心组件**:

```go
type Storage interface {
    // 客户端 CRUD
    Add(ctx context.Context, cli *Persistent) error
    Update(ctx context.Context, name string, cli *Persistent) error
    Remove(ctx context.Context, name string) error
    Find(name string) (cli *Persistent, ok bool)
    
    // 运行时客户端
    FindByIP(ip netip.Addr) (cli *Runtime, ok bool)
}
```

**客户端类型**:

1. **持久化客户端** (Persistent):
   - 手动配置
   - 自定义设置
   - 标签和分组
   - 上游 DNS 配置

2. **运行时客户端** (Runtime):
   - 自动发现
   - 从 /etc/hosts
   - 从 DHCP 租约
   - 从 rDNS 查询
   - 从 ARP 表

**客户端设置**:

```go
type Settings struct {
    // 识别信息
    ClientName string
    ClientIP   netip.Addr
    ClientTags []string
    
    // 过滤设置
    FilteringEnabled    bool
    SafeSearchEnabled   bool
    SafeBrowsingEnabled bool
    ParentalEnabled     bool
    
    // 服务阻止
    BlockedServices *BlockedServices
    
    // 自定义上游
    Upstreams []string
}
```

**地址处理器** (AddressProcessor):
- rDNS 反向解析
- WHOIS 信息查询
- 客户端信息更新

---

### 5. 查询日志 (querylog)

**位置**: `internal/querylog/`

**职责**: DNS 查询记录和检索

**存储格式**:

```json
{
  "IP": "192.168.1.100",
  "T": "2024-12-04T10:30:45Z",
  "QH": "example.com",
  "QT": "A",
  "QC": "IN",
  "CP": "doh",
  "Answer": "base64...",
  "Result": {
    "IsFiltered": true,
    "Reason": 3,
    "Rule": "||ads.example.com^",
    "FilterID": 1
  },
  "Elapsed": 45,
  "Upstream": "https://dns.google/dns-query"
}
```

**功能特性**:

1. **日志记录**:
   - 所有 DNS 查询
   - 过滤结果
   - 响应时间
   - 上游服务器

2. **查询功能**:
   - 按时间范围
   - 按域名搜索
   - 按客户端 IP
   - 按过滤状态

3. **隐私保护**:
   - IP 地址匿名化
   - 可配置保留时间
   - 自动轮转

4. **性能优化**:
   - 内存缓冲
   - 批量写入
   - 索引优化

---

### 6. 统计模块 (stats)

**位置**: `internal/stats/`

**职责**: DNS 使用统计和分析

**统计数据**:

```go
type Stats struct {
    // 总计数器
    NumDNSQueries           uint64
    NumBlockedFiltering     uint64
    NumReplacedSafebrowsing uint64
    NumReplacedSafesearch   uint64
    NumReplacedParental     uint64
    AvgProcessingTime       float64
    
    // 时间序列数据
    DNSQueries         []uint64
    BlockedFiltering   []uint64
    ReplacedParental   []uint64
    ReplacedSafebrowsing []uint64
    
    // Top 列表
    TopQueriedDomains []map[string]uint64
    TopBlockedDomains []map[string]uint64
    TopClients        []map[string]uint64
}
```

**时间单位**:
- 小时 (24 小时)
- 天 (7/30/90 天)

**存储机制**:
- 内存中的当前单元
- 定期刷新到磁盘
- 使用 bbolt 数据库



---

### 7. Web API 模块 (home/web)

**位置**: `internal/home/`

**职责**: HTTP API 和 Web 界面服务

**核心功能**:

1. **安装向导**:
   - 首次启动配置
   - 网络接口选择
   - 管理员账户创建
   - 端口冲突检测

2. **认证授权**:
   - 基于 Session 的认证
   - bcrypt 密码哈希
   - Cookie 管理
   - 速率限制

3. **配置管理**:
   - YAML 配置读写
   - 热重载
   - 配置迁移
   - 备份恢复

4. **API 端点**:
   - `/control/status` - 全局状态
   - `/control/dns_*` - DNS 配置
   - `/control/filtering/*` - 过滤配置
   - `/control/clients/*` - 客户端管理
   - `/control/querylog` - 查询日志
   - `/control/stats` - 统计数据
   - `/control/dhcp/*` - DHCP 配置

**认证流程**:

```
客户端                    服务器
  │                        │
  ├─ POST /control/login ─>│
  │  {name, password}      │
  │                        ├─ 验证凭据
  │                        ├─ 生成 Session
  │                        ├─ 存储到 DB
  │<─ Set-Cookie: session ─┤
  │                        │
  ├─ GET /control/status ─>│
  │  Cookie: session=...   │
  │                        ├─ 验证 Session
  │<─ 200 OK {data} ───────┤
```

---

### 8. 配置管理 (configmigrate)

**位置**: `internal/configmigrate/`

**职责**: 配置文件版本迁移

**迁移机制**:

```go
type Migrator struct {
    // 当前版本
    currentVersion int
    
    // 迁移函数列表
    migrations []migration
}

type migration func(diskConf yobj) (err error)
```

**版本历史**: 当前支持 v1 到 v31 的迁移

**迁移示例**:
- v1: 初始版本
- v10: 添加 DHCP 配置
- v20: 客户端配置重构
- v28: 统计配置更新
- v31: 最新版本

---

### 9. 系统集成模块

#### 9.1 ARP 数据库 (arpdb)

**位置**: `internal/arpdb/`

**职责**: 从系统 ARP 表获取客户端信息

**平台支持**:
- Linux: 读取 `/proc/net/arp`
- BSD: 执行 `arp -a`
- Windows: 执行 `arp -a`
- OpenBSD: 执行 `arp -an`

#### 9.2 反向 DNS (rdns)

**位置**: `internal/rdns/`

**职责**: IP 地址反向解析为主机名

**功能**:
- PTR 记录查询
- 缓存机制
- 私有地址处理
- 批量解析

#### 9.3 WHOIS 查询 (whois)

**位置**: `internal/whois/`

**职责**: 查询 IP 地址的 WHOIS 信息

**返回信息**:
- 组织名称
- 国家
- 城市
- 网络范围

#### 9.4 IPSet 集成 (ipset)

**位置**: `internal/ipset/`

**职责**: Linux ipset 集成

**功能**:
- 将域名解析的 IP 添加到 ipset
- 支持 IPv4 和 IPv6
- 用于防火墙规则

**配置示例**:
```yaml
ipset:
  - domain1.com,domain2.com/ipset_name
```

---

## 数据流

### DNS 查询处理流程

```
1. 客户端发送 DNS 查询
   │
   ▼
2. dnsProxy 接收请求
   │
   ├─> 协议解析 (UDP/TCP/DoH/DoT/DoQ)
   ├─> 客户端识别 (IP/ClientID)
   └─> 传递给 dnsforward.Server
   │
   ▼
3. BeforeRequestHandler (前置处理)
   │
   ├─> 访问控制检查
   │   ├─> allowlist/blocklist
   │   └─> blocked_hosts
   │
   ├─> 客户端设置应用
   │   ├─> 从持久化配置
   │   ├─> 从 DHCP 租约
   │   └─> 从运行时信息
   │
   └─> 速率限制检查
   │
   ▼
4. 过滤处理 (filtering.DNSFilter)
   │
   ├─> DNS 重写检查
   │   └─> 匹配 → 返回自定义响应
   │
   ├─> /etc/hosts 检查
   │   └─> 匹配 → 返回本地解析
   │
   ├─> 规则列表过滤
   │   ├─> 白名单 → 标记为允许
   │   └─> 黑名单 → 阻止
   │
   ├─> 服务阻止检查
   │   └─> 匹配 → 阻止
   │
   ├─> 安全浏览检查
   │   └─> 恶意 → 阻止
   │
   ├─> 家长控制检查
   │   └─> 不适 → 阻止
   │
   └─> 安全搜索检查
       └─> 搜索引擎 → 重写
   │
   ▼
5. 判断是否需要上游查询
   │
   ├─> 已被阻止 → 构造阻止响应
   │   ├─> NXDOMAIN
   │   ├─> REFUSED
   │   ├─> 0.0.0.0 / ::
   │   └─> 自定义 IP
   │
   └─> 未阻止 → 转发到上游
       │
       ▼
6. 上游 DNS 查询
   │
   ├─> 选择上游服务器
   │   ├─> 默认上游
   │   ├─> 客户端自定义上游
   │   ├─> 私有 rDNS 上游
   │   └─> 并行/最快模式
   │
   ├─> 发送查询
   │   ├─> 支持多种协议
   │   ├─> 超时控制
   │   └─> 重试机制
   │
   └─> 接收响应
   │
   ▼
7. 响应后处理
   │
   ├─> CNAME 过滤
   │   └─> 检查 CNAME 目标
   │
   ├─> IP 地址过滤
   │   └─> 检查 A/AAAA 记录
   │
   ├─> DNS64 处理
   │   └─> IPv6 合成
   │
   └─> IPSet 处理
       └─> 添加 IP 到 ipset
   │
   ▼
8. 记录和统计
   │
   ├─> 查询日志记录
   │   ├─> 请求信息
   │   ├─> 响应信息
   │   ├─> 过滤结果
   │   └─> 处理时间
   │
   └─> 统计更新
       ├─> 查询计数
       ├─> 阻止计数
       ├─> Top 域名
       └─> Top 客户端
   │
   ▼
9. 返回响应给客户端
```

### 配置更新流程

```
1. Web UI / API 请求
   │
   ▼
2. HTTP Handler 接收
   │
   ├─> 认证检查
   ├─> 参数验证
   └─> 调用配置修改器
   │
   ▼
3. ConfigModifier.Apply()
   │
   ├─> 更新内存配置
   ├─> 写入 YAML 文件
   └─> 触发重载
   │
   ▼
4. 模块重新配置
   │
   ├─> DNS 服务器重启
   │   ├─> 停止旧实例
   │   ├─> 应用新配置
   │   └─> 启动新实例
   │
   ├─> 过滤引擎重载
   │   ├─> 重新加载规则
   │   └─> 重建索引
   │
   └─> 其他模块更新
   │
   ▼
5. 返回成功响应
```

### 过滤规则更新流程

```
1. 触发更新
   │
   ├─> 手动触发 (API)
   ├─> 定时自动更新
   └─> 首次启动
   │
   ▼
2. 下载规则列表
   │
   ├─> 并发下载多个列表
   ├─> HTTP/HTTPS 请求
   ├─> 超时和重试
   └─> 验证内容
   │
   ▼
3. 保存到磁盘
   │
   ├─> 写入 data/filters/ 目录
   ├─> 文件名: <filter_id>.txt
   └─> 原子写入 (renameio)
   │
   ▼
4. 解析规则
   │
   ├─> 逐行解析
   ├─> 语法验证
   ├─> 规则分类
   │   ├─> 网络规则
   │   ├─> Host 规则
   │   └─> 修饰符规则
   └─> 构建索引
   │
   ▼
5. 重建过滤引擎
   │
   ├─> 创建新的 RuleStorage
   ├─> 创建新的 DNSEngine
   └─> 原子替换旧引擎
   │
   ▼
6. 清理和优化
   │
   ├─> 释放旧引擎内存
   ├─> 调用 GC
   └─> 记录更新日志
```

---

## 关键设计

### 1. 并发安全设计

**读写锁使用**:

```go
// DNS 服务器
type Server struct {
    serverLock sync.RWMutex  // 保护服务器状态
    engineLock sync.RWMutex  // 保护过滤引擎
}

// 配置管理
type DNSFilter struct {
    confMu *sync.RWMutex     // 保护配置
    filtersMu *sync.RWMutex  // 保护过滤器列表
}
```

**原子操作**:

```go
// 启用状态
enabled uint32  // 使用 atomic.LoadUint32/StoreUint32

// 保护更新标志
protectionUpdateInProgress atomic.Bool
```

### 2. 缓存策略

**多层缓存**:

1. **DNS 响应缓存** (dnsproxy):
   - LRU 缓存
   - 尊重 TTL
   - 可配置大小

2. **过滤结果缓存**:
   - 安全浏览缓存
   - 家长控制缓存
   - 安全搜索缓存

3. **客户端 ID 缓存**:
   - LRU 缓存
   - 临时存储
   - 1024 条目

### 3. 错误处理

**分层错误处理**:

```go
// 自定义错误类型
type PrivateRDNSError struct {
    err error
}

// 错误包装
return fmt.Errorf("preparing upstream: %w", err)

// 错误注解
return errors.Annotate(err, "filtering: %w")
```

**优雅降级**:
- 上游失败 → 使用备用上游
- 过滤失败 → 允许通过
- 日志失败 → 继续处理

### 4. 性能优化

**内存池**:

```go
// 缓冲区池
bufPool *syncutil.Pool[[]byte]

// 复用缓冲区
buf := d.bufPool.Get()
defer d.bufPool.Put(buf)
```

**批量操作**:
- 查询日志批量写入
- 统计数据批量更新
- 规则批量加载

**索引优化**:
- 域名哈希索引
- IP 地址前缀树
- 客户端快速查找

### 5. 配置持久化

**YAML 格式**:

```yaml
bind_host: 0.0.0.0
bind_port: 3000
users:
  - name: admin
    password: $2a$10$...
dns:
  bind_hosts:
    - 0.0.0.0
  port: 53
  upstream_dns:
    - https://dns.google/dns-query
  bootstrap_dns:
    - 9.9.9.10
filtering:
  protection_enabled: true
  filtering_enabled: true
  filters:
    - enabled: true
      url: https://...
      name: AdGuard DNS filter
      id: 1
```

**原子写入**:

```go
// 使用临时文件
tmpFile := configFile + ".tmp"
// 写入临时文件
// 原子重命名
os.Rename(tmpFile, configFile)
```

### 6. 模块化设计

**接口抽象**:

```go
// DHCP 接口
type DHCP interface {
    HostByIP(ip netip.Addr) (host string)
    IPByHost(host string) (ip netip.Addr)
    Enabled() (ok bool)
}

// 统计接口
type Interface interface {
    Update(e Entry)
    GetTopClientsIP(limit uint) []netip.Addr
    Clear()
}

// 查询日志接口
type QueryLog interface {
    Add(ctx context.Context, params *AddParams) error
    Clear() error
}
```

**依赖注入**:

```go
// 创建 DNS 服务器时注入依赖
func NewServer(p DNSCreateParams) (*Server, error) {
    return &Server{
        dnsFilter:   p.DNSFilter,
        stats:       p.Stats,
        queryLog:    p.QueryLog,
        dhcpServer:  p.DHCPServer,
        // ...
    }, nil
}
```

### 7. 热重载机制

**无中断重启**:

```go
func (s *Server) Reconfigure(conf *ServerConfig) error {
    // 1. 停止旧服务
    s.stopLocked()
    
    // 2. 等待连接关闭
    time.Sleep(100 * time.Millisecond)
    
    // 3. 应用新配置
    s.Prepare(conf)
    
    // 4. 启动新服务
    return s.startLocked()
}
```

**配置版本控制**:
- 配置文件版本号
- 自动迁移机制
- 向后兼容



### 8. 日志系统

**结构化日志**:

```go
// 使用 slog (Go 1.21+)
logger := slog.Default()
logger.InfoContext(ctx, "starting server", 
    "version", version.Full(),
    "port", port)

// 带前缀的子日志器
dnsLogger := baseLogger.With(slogutil.KeyPrefix, "dnsforward")
```

**日志级别**:
- DEBUG - 详细调试信息
- INFO - 一般信息
- WARN - 警告信息
- ERROR - 错误信息

**日志轮转**:
- 使用 lumberjack
- 按大小轮转
- 保留历史文件
- 压缩旧日志

---

## 部署架构

### 1. 单机部署

**标准部署**:

```
┌─────────────────────────────────────┐
│         物理/虚拟服务器              │
│                                      │
│  ┌────────────────────────────────┐ │
│  │     AdGuard Home 进程          │ │
│  │                                │ │
│  │  - DNS 服务 (53)               │ │
│  │  - Web UI (3000/80/443)        │ │
│  │  - DHCP 服务 (67/68)           │ │
│  └────────────────────────────────┘ │
│                                      │
│  ┌────────────────────────────────┐ │
│  │     文件系统                    │ │
│  │                                │ │
│  │  /opt/AdGuardHome/             │ │
│  │  ├── AdGuardHome (二进制)      │ │
│  │  ├── AdGuardHome.yaml (配置)   │ │
│  │  └── data/ (数据目录)          │ │
│  │      ├── filters/ (规则)       │ │
│  │      ├── sessions.db (会话)    │ │
│  │      ├── stats.db (统计)       │ │
│  │      └── querylog.json (日志)  │ │
│  └────────────────────────────────┘ │
└─────────────────────────────────────┘
```

**系统要求**:
- CPU: 1 核心 (推荐 2+)
- 内存: 512MB (推荐 1GB+)
- 磁盘: 100MB + 日志空间
- 操作系统: Linux/macOS/Windows/BSD

### 2. Docker 部署

**Docker Compose 配置**:

```yaml
version: '3'
services:
  adguardhome:
    image: adguard/adguardhome:latest
    container_name: adguardhome
    restart: unless-stopped
    ports:
      - "53:53/tcp"
      - "53:53/udp"
      - "67:67/udp"    # DHCP
      - "68:68/udp"    # DHCP
      - "80:80/tcp"    # Web UI
      - "443:443/tcp"  # HTTPS
      - "443:443/udp"  # QUIC
      - "3000:3000/tcp" # 初始设置
      - "853:853/tcp"  # DoT
      - "784:784/udp"  # DoQ
    volumes:
      - ./workdir:/opt/adguardhome/work
      - ./confdir:/opt/adguardhome/conf
    cap_add:
      - NET_ADMIN      # 用于 DHCP
```

**优势**:
- 隔离环境
- 易于升级
- 跨平台一致性
- 资源限制

### 3. Kubernetes 部署

**部署清单示例**:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: adguardhome
spec:
  replicas: 1
  selector:
    matchLabels:
      app: adguardhome
  template:
    metadata:
      labels:
        app: adguardhome
    spec:
      containers:
      - name: adguardhome
        image: adguard/adguardhome:latest
        ports:
        - containerPort: 53
          protocol: UDP
        - containerPort: 53
          protocol: TCP
        - containerPort: 3000
        volumeMounts:
        - name: config
          mountPath: /opt/adguardhome/conf
        - name: data
          mountPath: /opt/adguardhome/work
      volumes:
      - name: config
        persistentVolumeClaim:
          claimName: adguardhome-config
      - name: data
        persistentVolumeClaim:
          claimName: adguardhome-data
---
apiVersion: v1
kind: Service
metadata:
  name: adguardhome-dns
spec:
  type: LoadBalancer
  ports:
  - port: 53
    protocol: UDP
    name: dns-udp
  - port: 53
    protocol: TCP
    name: dns-tcp
  selector:
    app: adguardhome
```

### 4. 高可用部署

**主备模式**:

```
                  ┌─────────────┐
                  │   客户端     │
                  └──────┬──────┘
                         │
                         ▼
              ┌──────────────────────┐
              │   负载均衡器/VIP      │
              │   (Keepalived)       │
              └──────────┬───────────┘
                         │
          ┌──────────────┴──────────────┐
          │                             │
          ▼                             ▼
┌─────────────────┐           ┌─────────────────┐
│ AdGuard Home 1  │           │ AdGuard Home 2  │
│   (主节点)       │◄─────────►│   (备节点)       │
│                 │  配置同步  │                 │
└─────────────────┘           └─────────────────┘
```

**配置同步方案**:
- 使用 [adguardhome-sync](https://github.com/bakito/adguardhome-sync)
- 定期同步配置
- 双向同步支持
- 冲突解决

### 5. 网络拓扑

**家庭网络部署**:

```
互联网
  │
  ▼
┌─────────────┐
│   路由器     │
│  (网关)      │
└──────┬──────┘
       │
       ├─────────────────────────────┐
       │                             │
       ▼                             ▼
┌─────────────┐              ┌─────────────┐
│ AdGuard Home│              │  其他设备    │
│  (DNS服务器) │              │             │
│             │              │  - 电脑      │
│ 192.168.1.2 │              │  - 手机      │
└─────────────┘              │  - 智能电视  │
                             │  - IoT设备   │
                             └─────────────┘
```

**配置步骤**:
1. 安装 AdGuard Home
2. 配置路由器 DHCP，设置 DNS 为 AdGuard Home IP
3. 或手动配置每个设备的 DNS

**企业网络部署**:

```
                    互联网
                      │
                      ▼
              ┌───────────────┐
              │  边界防火墙    │
              └───────┬───────┘
                      │
              ┌───────┴───────┐
              │  核心交换机    │
              └───────┬───────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
        ▼             ▼             ▼
┌──────────────┐ ┌──────────┐ ┌──────────┐
│AdGuard Home 1│ │  内部DNS  │ │  客户端   │
│  (主DNS)     │ │  (AD等)   │ │  网段     │
└──────────────┘ └──────────┘ └──────────┘
        │
        ▼
┌──────────────┐
│AdGuard Home 2│
│  (备DNS)     │
└──────────────┘
```

---

## 安全机制

### 1. 认证和授权

**密码安全**:
- bcrypt 哈希 (cost=10)
- 不存储明文密码
- Session 管理

**速率限制**:
```go
type authRateLimiter struct {
    blockDuration time.Duration  // 阻止时长
    maxAttempts   uint            // 最大尝试次数
    failedAttempts map[string]*failedAuth
}
```

**Session 管理**:
- 基于 Cookie
- 可配置过期时间
- 存储在 bbolt 数据库
- 自动清理过期 Session

### 2. TLS/HTTPS 支持

**证书配置**:
```yaml
tls:
  enabled: true
  server_name: dns.example.com
  force_https: true
  port_https: 443
  port_dns_over_tls: 853
  port_dns_over_quic: 784
  certificate_chain: |
    -----BEGIN CERTIFICATE-----
    ...
  private_key: |
    -----BEGIN PRIVATE KEY-----
    ...
```

**支持的 TLS 协议**:
- TLS 1.2
- TLS 1.3
- 可配置密码套件

### 3. 访问控制

**IP 访问控制**:
```yaml
dns:
  allowed_clients:
    - 192.168.1.0/24
    - 10.0.0.0/8
  disallowed_clients:
    - 192.168.1.100
  blocked_hosts:
    - version.bind
    - id.server
```

**客户端识别**:
- IP 地址
- ClientID (DoH/DoT)
- MAC 地址 (DHCP)
- 子网匹配

### 4. 隐私保护

**查询日志匿名化**:
```yaml
querylog:
  anonymize_client_ip: true  # 掩码 /24 (IPv4) 或 /112 (IPv6)
```

**数据保留策略**:
- 可配置日志保留时间 (1/7/30/90 天)
- 自动轮转和清理
- 统计数据聚合

**本地处理**:
- 所有数据本地存储
- 不发送遥测数据
- 可选的外部服务 (安全浏览/家长控制)



---

## 性能特性

### 1. 性能指标

**典型性能**:
- 查询处理: < 10ms (本地缓存)
- 查询处理: 20-50ms (上游查询)
- 并发连接: 1000+ 同时连接
- 吞吐量: 10000+ 查询/秒 (取决于硬件)

**资源使用**:
- 内存: 50-200MB (取决于规则数量和缓存)
- CPU: 低负载 < 5%, 高负载 < 50%
- 磁盘 I/O: 主要是日志写入

### 2. 缓存机制

**DNS 缓存**:
```yaml
dns:
  cache_size: 4194304        # 4MB
  cache_ttl_min: 0           # 最小 TTL (秒)
  cache_ttl_max: 0           # 最大 TTL (秒)
  cache_optimistic: false    # 乐观缓存
```

**缓存策略**:
- LRU 淘汰算法
- 尊重 DNS TTL
- 可配置缓存大小
- 预取机制 (可选)

### 3. 并发处理

**Goroutine 池**:
```go
type Server struct {
    conf ServerConfig
    // MaxGoroutines 限制并发 goroutine 数量
    MaxGoroutines int
}
```

**连接复用**:
- HTTP/2 连接复用 (DoH)
- TLS 会话恢复
- Keep-Alive 连接

### 4. 优化建议

**规则列表优化**:
- 使用较少的大型列表而非多个小列表
- 定期清理无效规则
- 避免重复规则
- 使用域名哈希索引

**上游 DNS 优化**:
```yaml
dns:
  upstream_mode: parallel    # 或 fastest_addr
  upstream_dns:
    - https://dns.google/dns-query
    - https://cloudflare-dns.com/dns-query
  fastest_timeout: 1s        # 最快地址超时
```

**系统优化**:
```yaml
os:
  group: adguard
  user: adguard
  rlimit_nofile: 8192        # 文件描述符限制
```

---

## 监控和运维

### 1. 健康检查

**API 端点**:
```bash
# 获取状态
curl http://localhost:3000/control/status

# 响应示例
{
  "dns_addresses": ["127.0.0.1:53"],
  "dns_port": 53,
  "http_port": 3000,
  "protection_enabled": true,
  "running": true,
  "version": "v0.107.0"
}
```

**监控指标**:
- DNS 服务状态
- 查询速率
- 阻止率
- 上游响应时间
- 内存使用
- 磁盘使用

### 2. 日志管理

**日志位置**:
```
/opt/AdGuardHome/
├── AdGuardHome.log          # 应用日志
└── data/
    ├── querylog.json        # 查询日志
    ├── querylog.json.1      # 轮转日志
    └── stats.db             # 统计数据
```

**日志级别配置**:
```yaml
log:
  file: ""                   # 空表示 stdout
  max_backups: 0
  max_size: 100              # MB
  max_age: 3                 # 天
  compress: false
  local_time: false
  verbose: false
```

### 3. 备份和恢复

**需要备份的文件**:
```bash
# 配置文件
AdGuardHome.yaml

# 数据目录
data/
├── filters/                 # 过滤规则
├── sessions.db              # 用户会话
├── stats.db                 # 统计数据
└── querylog.json*           # 查询日志
```

**备份脚本示例**:
```bash
#!/bin/bash
BACKUP_DIR="/backup/adguardhome"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 备份配置和数据
tar -czf "$BACKUP_DIR/agh_backup_$DATE.tar.gz" \
    /opt/AdGuardHome/AdGuardHome.yaml \
    /opt/AdGuardHome/data/

# 保留最近 7 天的备份
find "$BACKUP_DIR" -name "agh_backup_*.tar.gz" -mtime +7 -delete
```

### 4. 更新管理

**自动更新**:
```yaml
# 禁用自动更新
# 启动参数: --no-check-update
```

**手动更新**:
```bash
# 通过 API
curl -X POST http://localhost:3000/control/update

# 或手动下载
wget https://github.com/AdguardTeam/AdGuardHome/releases/latest/download/AdGuardHome_linux_amd64.tar.gz
tar -xzf AdGuardHome_linux_amd64.tar.gz
./AdGuardHome -s stop
cp AdGuardHome /opt/AdGuardHome/
./AdGuardHome -s start
```

### 5. 故障排查

**常见问题**:

1. **端口冲突**:
```bash
# 检查端口占用
sudo netstat -tulpn | grep :53
sudo lsof -i :53

# 解决方案
# - 停止冲突服务 (如 systemd-resolved)
# - 更改 AdGuard Home 端口
```

2. **权限问题**:
```bash
# 绑定特权端口 (< 1024) 需要 root 或 CAP_NET_BIND_SERVICE
sudo setcap 'cap_net_bind_service=+ep' /opt/AdGuardHome/AdGuardHome
```

3. **DNS 解析失败**:
```bash
# 检查上游 DNS
dig @127.0.0.1 example.com

# 查看日志
tail -f /opt/AdGuardHome/AdGuardHome.log

# 测试上游连接
curl -v https://dns.google/dns-query
```

4. **过滤规则问题**:
```bash
# 检查规则语法
# 查看 Web UI 的过滤日志
# 临时禁用规则列表进行排查
```

---

## 扩展和集成

### 1. API 集成

**REST API**:
```bash
# 认证
curl -X POST http://localhost:3000/control/login \
  -H "Content-Type: application/json" \
  -d '{"name":"admin","password":"password"}'

# 获取统计
curl -b cookies.txt http://localhost:3000/control/stats

# 添加过滤规则
curl -X POST http://localhost:3000/control/filtering/add_url \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"name":"My List","url":"https://example.com/list.txt"}'
```

**Python 客户端**:
```python
from adguardhome import AdGuardHome

agh = AdGuardHome("http://localhost:3000", username="admin", password="password")

# 获取状态
status = agh.status()
print(f"Version: {status['version']}")

# 获取统计
stats = agh.stats.info()
print(f"Total queries: {stats['num_dns_queries']}")

# 添加客户端
agh.clients.add({
    "name": "My Device",
    "ids": ["192.168.1.100"],
    "use_global_settings": False,
    "filtering_enabled": True
})
```

### 2. 第三方工具

**AdGuardian-Term**:
- 终端实时监控
- 流量统计
- 查询日志查看

**AdGuard Home Sync**:
- 多实例配置同步
- 双向同步
- 冲突解决

**Home Assistant 集成**:
```yaml
# configuration.yaml
adguard:
  host: 192.168.1.2
  port: 3000
  username: admin
  password: !secret adguard_password
  ssl: false
  verify_ssl: false
```

### 3. 路由器集成

**OpenWrt**:
- LUCI 应用: luci-app-adguardhome
- 一键安装和配置
- 与 dnsmasq 集成

**GL.iNet 路由器**:
- 原生支持
- Web UI 集成
- 一键启用

**Asuswrt-Merlin**:
- 安装脚本支持
- 自动配置
- 固件集成

### 4. DNS-over-HTTPS 客户端配置

**浏览器配置**:

Firefox:
```
about:config
network.trr.mode = 2
network.trr.uri = https://your-server/dns-query
```

Chrome:
```
设置 -> 隐私和安全 -> 安全 -> 使用安全 DNS
自定义: https://your-server/dns-query
```

**操作系统配置**:

iOS/macOS:
- 安装 .mobileconfig 配置文件
- 访问: https://your-server/apple/doh.mobileconfig

Android:
- 设置 -> 网络和互联网 -> 私人 DNS
- 输入: your-server

Windows:
- 使用第三方工具 (如 Simple DNSCrypt)
- 或配置 DoH 客户端

---

## 最佳实践

### 1. 安全配置

**强化建议**:

```yaml
# 1. 使用强密码
users:
  - name: admin
    password: $2a$10$...  # 使用 htpasswd 生成

# 2. 启用 HTTPS
tls:
  enabled: true
  force_https: true
  certificate_chain: /path/to/cert.pem
  private_key: /path/to/key.pem

# 3. 限制访问
dns:
  allowed_clients:
    - 192.168.0.0/16
    - 10.0.0.0/8

# 4. 启用速率限制
dns:
  ratelimit: 20  # 每秒查询数

# 5. 禁用不需要的功能
dhcp:
  enabled: false  # 如果不使用 DHCP
```

### 2. 性能调优

**推荐配置**:

```yaml
dns:
  # 缓存优化
  cache_size: 8388608        # 8MB
  cache_ttl_min: 60          # 最小缓存 1 分钟
  cache_ttl_max: 86400       # 最大缓存 1 天
  
  # 并发优化
  upstream_mode: parallel    # 并行查询
  fastest_timeout: 1s
  
  # 上游优化
  upstream_dns:
    - https://dns.google/dns-query
    - https://cloudflare-dns.com/dns-query
  bootstrap_dns:
    - 8.8.8.8
    - 1.1.1.1

# 日志优化
querylog:
  enabled: true
  interval: 7                # 保留 7 天
  anonymize_client_ip: true

statistics:
  interval: 7                # 保留 7 天

# 系统优化
os:
  rlimit_nofile: 8192
```

### 3. 规则列表选择

**推荐列表**:

```yaml
filters:
  # 基础广告拦截
  - enabled: true
    url: https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt
    name: AdGuard DNS filter
    
  # 隐私保护
  - enabled: true
    url: https://raw.githubusercontent.com/AdguardTeam/FiltersRegistry/master/filters/filter_3_Spyware/filter.txt
    name: AdGuard Tracking Protection filter
    
  # 恶意软件
  - enabled: true
    url: https://raw.githubusercontent.com/AdguardTeam/FiltersRegistry/master/filters/filter_15_DnsFilter/filter.txt
    name: AdGuard DNS Malware filter
    
  # 社交媒体 (可选)
  - enabled: false
    url: https://raw.githubusercontent.com/AdguardTeam/FiltersRegistry/master/filters/filter_4_Social/filter.txt
    name: AdGuard Social Media filter
```

**规则数量建议**:
- 小型家庭: 50,000 - 100,000 条规则
- 中型网络: 100,000 - 300,000 条规则
- 大型网络: 300,000+ 条规则

### 4. 网络配置

**DNS 配置**:

```yaml
dns:
  # 监听地址
  bind_hosts:
    - 0.0.0.0        # 监听所有接口
  
  # 端口配置
  port: 53
  
  # 上游 DNS
  upstream_dns:
    - https://dns.google/dns-query
    - tls://dns.google
    - 8.8.8.8
  
  # 私有反向 DNS
  use_private_ptr_resolvers: true
  local_ptr_upstreams:
    - 192.168.1.1    # 路由器 DNS
  
  # 可信代理 (如果在反向代理后)
  trusted_proxies:
    - 127.0.0.0/8
    - ::1/128
```

### 5. 维护计划

**日常维护**:
- 每天: 检查服务状态
- 每周: 查看统计和日志
- 每月: 更新过滤规则
- 每季度: 系统更新和备份验证

**定期任务**:
```bash
# Cron 任务示例
# 每天凌晨 2 点备份
0 2 * * * /usr/local/bin/backup-adguardhome.sh

# 每周日凌晨 3 点更新规则
0 3 * * 0 curl -X POST http://localhost:3000/control/filtering/refresh

# 每月 1 号清理旧日志
0 4 1 * * find /opt/AdGuardHome/data -name "querylog.json.*" -mtime +30 -delete
```



---

## 架构优势

### 1. 技术优势

**单一二进制部署**:
- 无需复杂依赖
- 跨平台编译
- 易于分发和更新
- 嵌入式 Web UI

**高性能**:
- Go 语言原生并发
- 高效的内存管理
- 优化的 DNS 处理
- 智能缓存机制

**模块化设计**:
- 清晰的模块边界
- 接口抽象
- 易于测试和维护
- 可扩展性强

**协议支持全面**:
- 传统 DNS (UDP/TCP)
- 加密 DNS (DoH/DoT/DoQ)
- DNSCrypt
- DNS64

### 2. 功能优势

**全面的过滤能力**:
- 多种规则格式
- 实时更新
- 白名单支持
- 自定义规则

**安全特性**:
- 安全浏览
- 家长控制
- 恶意软件拦截
- 钓鱼网站防护

**易用性**:
- 直观的 Web UI
- 安装向导
- 实时统计
- 详细日志

**隐私保护**:
- 本地处理
- 数据匿名化
- 无遥测
- 开源透明

### 3. 运维优势

**轻量级**:
- 低资源占用
- 适合嵌入式设备
- 可运行在树莓派等

**可靠性**:
- 稳定运行
- 自动恢复
- 配置持久化
- 热重载支持

**可观测性**:
- 详细日志
- 实时统计
- 查询历史
- API 支持

---

## 架构局限

### 1. 技术限制

**DNS 层面限制**:
- 无法拦截 HTTPS 内容中的广告
- 无法处理加密的 SNI
- 对某些服务 (如 YouTube) 效果有限
- 无法拦截应用内广告 (某些情况)

**单点故障**:
- 单实例部署存在单点风险
- 需要额外配置实现高可用
- 配置同步需要第三方工具

**性能瓶颈**:
- 大量规则会影响性能
- 内存使用随规则数量增长
- 单机处理能力有限

### 2. 功能限制

**DHCP 功能**:
- 不如专业 DHCP 服务器功能丰富
- 某些高级特性缺失
- Windows 上不支持 DHCP

**统计功能**:
- 历史数据有限
- 无法导出详细报告
- 缺少高级分析功能

**集群支持**:
- 无原生集群支持
- 配置同步依赖第三方
- 负载均衡需要外部方案

### 3. 使用场景限制

**不适合场景**:
- 超大规模企业网络 (需要专业 DNS 方案)
- 需要复杂策略路由
- 需要深度包检测 (DPI)
- 需要内容过滤 (HTTP/HTTPS 内容)

**适合场景**:
- 家庭网络
- 小型办公室
- 个人使用
- 中小型企业
- 教育机构

---

## 未来发展方向

### 1. 功能增强

**计划中的功能**:
- 内容过滤代理 (类似浏览器扩展)
- 更强大的统计分析
- 机器学习辅助过滤
- 更好的移动端支持

**社区需求**:
- 集群原生支持
- 更多的集成选项
- 增强的 API
- 插件系统

### 2. 性能优化

**优化方向**:
- 更高效的规则匹配算法
- 更好的内存管理
- 并发性能提升
- 缓存策略优化

### 3. 生态建设

**生态系统**:
- 更多第三方集成
- 官方客户端库
- 插件市场
- 社区规则库

---

## 总结

### 核心特点

AdGuard Home 是一个**功能强大、易于部署、注重隐私**的网络级广告和跟踪器拦截 DNS 服务器。

**主要优势**:

1. **架构设计**:
   - 模块化、清晰的分层架构
   - Go 语言实现，性能优异
   - 单一二进制，部署简单

2. **功能完整**:
   - DNS 过滤、安全浏览、家长控制
   - 多协议支持 (DoH/DoT/DoQ)
   - DHCP 服务器集成
   - 详细的日志和统计

3. **易用性**:
   - 直观的 Web 管理界面
   - 简单的安装和配置
   - 丰富的文档和社区支持

4. **隐私保护**:
   - 本地数据处理
   - 开源透明
   - 可配置的匿名化

**适用场景**:

- ✅ 家庭网络保护
- ✅ 小型办公室
- ✅ 个人隐私保护
- ✅ 儿童上网管理
- ✅ IoT 设备保护
- ✅ 开发测试环境

**技术亮点**:

1. **DNS 转发层**: 基于 dnsproxy，支持多种协议和上游配置
2. **过滤引擎**: 使用 urlfilter，高效的规则匹配
3. **并发处理**: Go 原生并发，高性能处理
4. **模块化**: 清晰的模块划分，易于维护和扩展
5. **热重载**: 支持配置热更新，无需重启

### 架构评价

**优点**:
- ⭐ 架构清晰，模块职责明确
- ⭐ 代码质量高，注释完善
- ⭐ 性能优异，资源占用低
- ⭐ 功能全面，满足大多数需求
- ⭐ 社区活跃，持续更新

**改进空间**:
- 🔸 集群支持可以更完善
- 🔸 统计分析功能可以更强大
- 🔸 某些平台的 DHCP 支持有限
- 🔸 大规模部署的优化空间

### 技术选型建议

**推荐使用 AdGuard Home 的情况**:

1. **家庭用户**:
   - 保护家庭网络所有设备
   - 阻止广告和跟踪器
   - 家长控制功能

2. **小型企业**:
   - 员工数 < 100
   - 简单的 DNS 过滤需求
   - 预算有限

3. **技术爱好者**:
   - 自建服务
   - 学习 DNS 和网络安全
   - 定制化需求

**不推荐使用的情况**:

1. **大型企业**:
   - 需要企业级 DNS 方案
   - 复杂的策略需求
   - 需要专业支持

2. **特殊需求**:
   - 需要深度包检测
   - 需要 HTTP/HTTPS 内容过滤
   - 需要复杂的流量管理

### 学习价值

对于开发者，AdGuard Home 是一个**优秀的学习项目**:

1. **Go 语言实践**:
   - 并发编程
   - 网络编程
   - 系统编程

2. **架构设计**:
   - 模块化设计
   - 接口抽象
   - 依赖注入

3. **DNS 协议**:
   - DNS 协议实现
   - 加密 DNS
   - DNS 安全

4. **Web 开发**:
   - RESTful API
   - 前后端分离
   - 认证授权

5. **运维实践**:
   - 配置管理
   - 日志处理
   - 监控告警



---

## 参考资源

### 官方资源

**项目主页**:
- GitHub: https://github.com/AdguardTeam/AdGuardHome
- 官网: https://adguard.com/adguard-home/overview.html
- 文档: https://github.com/AdguardTeam/AdGuardHome/wiki

**下载和安装**:
- 发布页面: https://github.com/AdguardTeam/AdGuardHome/releases
- Docker Hub: https://hub.docker.com/r/adguard/adguardhome
- Snap Store: https://snapcraft.io/adguard-home

**API 文档**:
- OpenAPI 规范: https://github.com/AdguardTeam/AdGuardHome/tree/master/openapi
- 技术文档: 本项目的 `AGHTechDoc.md`

### 社区资源

**论坛和讨论**:
- Reddit: https://reddit.com/r/Adguard
- GitHub Discussions: https://github.com/AdguardTeam/AdGuardHome/discussions
- Telegram: https://t.me/adguard_en

**第三方工具**:
- AdGuard Home Sync: https://github.com/bakito/adguardhome-sync
- AdGuardian-Term: https://github.com/Lissy93/AdGuardian-Term
- Python 客户端: https://github.com/frenck/python-adguardhome
- Home Assistant 集成: https://www.home-assistant.io/integrations/adguard/

**规则列表**:
- AdGuard 过滤器: https://github.com/AdguardTeam/AdGuardSDNSFilter
- 过滤器注册表: https://github.com/AdguardTeam/FiltersRegistry
- 社区规则: https://filterlists.com/

### 技术文档

**Go 语言相关**:
- Go 官方文档: https://golang.org/doc/
- Go 并发模式: https://go.dev/blog/pipelines
- Go 网络编程: https://pkg.go.dev/net

**DNS 协议**:
- RFC 1035 (DNS): https://tools.ietf.org/html/rfc1035
- RFC 8484 (DoH): https://tools.ietf.org/html/rfc8484
- RFC 7858 (DoT): https://tools.ietf.org/html/rfc7858
- RFC 9250 (DoQ): https://tools.ietf.org/html/rfc9250

**依赖库文档**:
- dnsproxy: https://github.com/AdguardTeam/dnsproxy
- urlfilter: https://github.com/AdguardTeam/urlfilter
- golibs: https://github.com/AdguardTeam/golibs
- miekg/dns: https://github.com/miekg/dns

### 相关项目

**类似项目**:
- Pi-hole: https://pi-hole.net/
- Blocky: https://github.com/0xERR0R/blocky
- CoreDNS: https://coredns.io/
- Unbound: https://nlnetlabs.nl/projects/unbound/

**互补工具**:
- WireGuard: https://www.wireguard.com/ (VPN)
- Tailscale: https://tailscale.com/ (零配置 VPN)
- Nginx: https://nginx.org/ (反向代理)
- Prometheus: https://prometheus.io/ (监控)

### 学习资源

**DNS 安全**:
- DNS Security Extensions (DNSSEC)
- DNS over HTTPS (DoH)
- DNS over TLS (DoT)
- DNS Privacy

**网络安全**:
- 广告拦截技术
- 跟踪器防护
- 恶意软件防护
- 家长控制

**系统架构**:
- 微服务架构
- 高可用设计
- 性能优化
- 监控和运维

### 贡献指南

**如何贡献**:
- 代码贡献: https://github.com/AdguardTeam/AdGuardHome/blob/master/CONTRIBUTING.md
- 代码规范: https://github.com/AdguardTeam/CodeGuidelines
- 问题报告: https://github.com/AdguardTeam/AdGuardHome/issues
- 功能请求: https://github.com/AdguardTeam/AdGuardHome/discussions

**翻译贡献**:
- CrowdIn 项目: https://crowdin.com/project/adguard-applications/en#/adguard-home
- 翻译指南: https://kb.adguard.com/en/general/adguard-translations

---

## 附录

### A. 常用命令

**服务管理**:
```bash
# 启动服务
./AdGuardHome -s start

# 停止服务
./AdGuardHome -s stop

# 重启服务
./AdGuardHome -s restart

# 查看状态
./AdGuardHome -s status

# 安装为系统服务
./AdGuardHome -s install

# 卸载系统服务
./AdGuardHome -s uninstall
```

**配置管理**:
```bash
# 检查配置
./AdGuardHome --check-config

# 指定配置文件
./AdGuardHome -c /path/to/config.yaml

# 指定工作目录
./AdGuardHome -w /path/to/workdir

# 禁用更新检查
./AdGuardHome --no-check-update

# 使用本地前端文件 (开发)
./AdGuardHome --local-frontend
```

**测试命令**:
```bash
# 测试 DNS 查询
dig @127.0.0.1 example.com
nslookup example.com 127.0.0.1

# 测试 DoH
curl -H 'accept: application/dns-json' \
  'https://your-server/dns-query?name=example.com&type=A'

# 测试 DoT
kdig -d @127.0.0.1 +tls example.com

# 性能测试
dnsperf -s 127.0.0.1 -d queryfile.txt
```

### B. 配置示例

**最小配置**:
```yaml
bind_host: 0.0.0.0
bind_port: 3000
users:
  - name: admin
    password: $2a$10$...
dns:
  bind_hosts:
    - 0.0.0.0
  port: 53
  upstream_dns:
    - https://dns.google/dns-query
```

**完整配置示例**:
```yaml
# HTTP 服务配置
bind_host: 0.0.0.0
bind_port: 3000
auth_attempts: 5
block_auth_min: 15
http_proxy: ""
language: zh-cn
theme: auto

# 用户配置
users:
  - name: admin
    password: $2a$10$...

# DNS 配置
dns:
  bind_hosts:
    - 0.0.0.0
  port: 53
  
  # 统计和日志
  statistics_interval: 7
  querylog_enabled: true
  querylog_file_enabled: true
  querylog_interval: 7
  querylog_size_memory: 1000
  anonymize_client_ip: false
  
  # 保护功能
  protection_enabled: true
  blocking_mode: default
  blocking_ipv4: ""
  blocking_ipv6: ""
  blocked_response_ttl: 10
  
  # 家长控制和安全浏览
  parental_block_host: family-block.dns.adguard.com
  safebrowsing_block_host: standard-block.dns.adguard.com
  
  # 速率限制
  ratelimit: 20
  ratelimit_whitelist: []
  refuse_any: true
  
  # 上游 DNS
  upstream_dns:
    - https://dns.google/dns-query
    - https://cloudflare-dns.com/dns-query
  upstream_dns_file: ""
  bootstrap_dns:
    - 8.8.8.8
    - 1.1.1.1
  
  # 高级选项
  all_servers: false
  fastest_addr: false
  fastest_timeout: 1s
  allowed_clients: []
  disallowed_clients: []
  blocked_hosts:
    - version.bind
    - id.server
    - hostname.bind
  
  # 缓存
  cache_size: 4194304
  cache_ttl_min: 0
  cache_ttl_max: 0
  cache_optimistic: false
  
  # EDNS
  edns_client_subnet:
    custom_ip: ""
    enabled: false
    use_custom: false
  
  # 其他
  max_goroutines: 300
  handle_ddr: true
  ipset: []
  ipset_file: ""
  
  # 本地 PTR
  use_private_ptr_resolvers: true
  local_ptr_upstreams: []
  
  # 私有网络
  private_networks: []
  use_dns64: false
  dns64_prefixes: []
  
  # 服务端口
  serve_http3: false
  use_http3_upstreams: false

# TLS 配置
tls:
  enabled: false
  server_name: ""
  force_https: false
  port_https: 443
  port_dns_over_tls: 853
  port_dns_over_quic: 784
  port_dnscrypt: 0
  dnscrypt_config_file: ""
  allow_unencrypted_doh: false
  certificate_chain: ""
  private_key: ""
  certificate_path: ""
  private_key_path: ""
  strict_sni_check: false

# DHCP 配置
dhcp:
  enabled: false
  interface_name: ""
  local_domain_name: lan
  dhcpv4:
    gateway_ip: ""
    subnet_mask: ""
    range_start: ""
    range_end: ""
    lease_duration: 86400
    icmp_timeout_msec: 1000
    options: []
  dhcpv6:
    range_start: ""
    lease_duration: 86400
    ra_slaac_only: false
    ra_allow_slaac: false

# 客户端配置
clients:
  runtime_sources:
    whois: true
    arp: true
    rdns: true
    dhcp: true
    hosts: true
  persistent: []

# 日志配置
log:
  file: ""
  max_backups: 0
  max_size: 100
  max_age: 3
  compress: false
  local_time: false
  verbose: false

# 操作系统配置
os:
  group: ""
  user: ""
  rlimit_nofile: 0

# 过滤配置
filtering:
  protection_enabled: true
  filtering_enabled: true
  
  # 规则列表
  filters_update_interval: 24
  
  # 家长控制
  parental_enabled: false
  parental_sensitivity: 0
  
  # 安全浏览
  safebrowsing_enabled: false
  
  # 安全搜索
  safe_search:
    enabled: false
    bing: true
    duckduckgo: true
    google: true
    pixabay: true
    yandex: true
    youtube: true
  
  # 阻止的服务
  blocked_services:
    schedule:
      time_zone: Local
    ids: []

# 白名单过滤器
whitelist_filters: []

# 用户规则
user_rules: []

# DNS 重写
dns_rewrites: []

# 过滤器
filters:
  - enabled: true
    url: https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt
    name: AdGuard DNS filter
    id: 1
```

### C. 故障排查清单

**DNS 解析问题**:
- [ ] 检查 AdGuard Home 服务状态
- [ ] 验证端口 53 是否被占用
- [ ] 测试上游 DNS 连接
- [ ] 检查防火墙规则
- [ ] 查看 DNS 查询日志
- [ ] 验证客户端 DNS 配置

**Web UI 访问问题**:
- [ ] 检查 HTTP 服务端口
- [ ] 验证防火墙规则
- [ ] 检查 TLS 证书 (如果启用 HTTPS)
- [ ] 清除浏览器缓存
- [ ] 检查认证凭据

**过滤不生效**:
- [ ] 确认保护功能已启用
- [ ] 检查过滤规则是否加载
- [ ] 验证规则语法
- [ ] 查看过滤日志
- [ ] 测试特定域名

**性能问题**:
- [ ] 检查 CPU 和内存使用
- [ ] 查看并发连接数
- [ ] 优化规则列表
- [ ] 调整缓存大小
- [ ] 检查磁盘 I/O

**DHCP 问题**:
- [ ] 确认 DHCP 已启用
- [ ] 检查网络接口配置
- [ ] 验证 IP 地址范围
- [ ] 查看租约列表
- [ ] 检查权限 (需要 root)

---

## 版本历史

**文档版本**: v1.0  
**最后更新**: 2024-12-04  
**AdGuard Home 版本**: v0.107.x  
**作者**: AI Assistant  

**更新日志**:
- 2024-12-04: 初始版本，完整架构分析

---

## 许可证

本文档基于 AdGuard Home 项目分析编写。

AdGuard Home 项目采用 **GNU General Public License v3.0** 许可证。

详见: https://github.com/AdguardTeam/AdGuardHome/blob/master/LICENSE.txt

---

**文档结束**

如有问题或建议，欢迎通过以下方式反馈:
- GitHub Issues: https://github.com/AdguardTeam/AdGuardHome/issues
- GitHub Discussions: https://github.com/AdguardTeam/AdGuardHome/discussions

