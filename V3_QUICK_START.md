# V3开发快速开始指南

## 🎯 当前状态

- **当前分支**: v3
- **基于版本**: v2 (2f1e0490)
- **V2备份**: 已保护，不可修改
- **开始日期**: 2024-11-25

## 📋 立即可以开始的任务

### 1️⃣ 问题#4: 提取公共域名匹配函数（2小时）

**文件**:
- `internal/dnsforward/upstream_groups.go`
- `internal/dnsforward/process.go`

**任务**:
```go
// 创建新文件: internal/dnsforward/domain_utils.go
package dnsforward

// MatchDomainPattern 统一的域名匹配函数
func MatchDomainPattern(domain, pattern string) bool {
    // 实现统一的匹配逻辑
}
```

**步骤**:
1. 创建 `internal/dnsforward/domain_utils.go`
2. 提取公共匹配逻辑
3. 更新两个文件使用新函数
4. 添加单元测试
5. 删除重复代码

---

### 2️⃣ 问题#7: 移除前端类型断言（3小时）

**文件**: `client/src/components/Filters/DnsRouting.tsx`

**任务**:
```typescript
// 定义正确的Props接口
interface DnsRoutingProps {
    // ... 现有props
    dnsConfig: DnsConfig;
    setDnsConfig: (config: DnsConfig) => Promise<void>;
}

// 移除所有 (this.props as any)
```

**步骤**:
1. 定义完整的Props接口
2. 定义DnsConfig类型
3. 替换所有 `as any`
4. 验证TypeScript编译无错误

---

### 3️⃣ 问题#9: 删除未使用的函数（1小时）

**文件**: `internal/dnsforward/upstream_groups.go`

**要删除的函数**:
- `GetEnabledUpstreamGroups()`
- `GetUpstreamGroupByName()`
- `getUpstreamGroupByID()` (与`GetUpstreamGroupByID()`重复)

**步骤**:
1. 确认函数未被调用
2. 删除函数定义
3. 运行测试确保无影响

---

### 4️⃣ 问题#11: 添加输入验证（2小时）

**文件**: `internal/dnsforward/http_dns_routing.go`

**任务**:
```go
func validateDnsRoutingRequest(req *dnsRoutingAddReq) error {
    // 验证URL格式
    if req.URL != "" {
        if _, err := url.Parse(req.URL); err != nil {
            return fmt.Errorf("invalid URL: %w", err)
        }
    }
    
    // 验证名称长度
    if len(req.Name) > 100 {
        return errors.Error("name too long")
    }
    
    // 验证GroupID存在
    // ...
    
    return nil
}
```

**步骤**:
1. 创建验证函数
2. 在所有HTTP处理函数中调用
3. 添加错误消息翻译
4. 添加单元测试

---

### 5️⃣ 问题#12: 添加加载状态（2小时）

**文件**: `client/src/components/Filters/DnsRouting.tsx`

**任务**:
```typescript
// 添加loading状态
const [isSubmitting, setIsSubmitting] = useState(false);

// 在提交时设置状态
const handleSubmit = async () => {
    setIsSubmitting(true);
    try {
        await submitAction();
    } finally {
        setIsSubmitting(false);
    }
};

// 禁用按钮
<button disabled={isSubmitting}>
    {isSubmitting ? 'Loading...' : 'Submit'}
</button>
```

---

### 6️⃣ 功能1: Priority范围验证（2小时）

**文件**: `internal/filtering/http.go`

**任务**:
```go
func validatePriority(priority int) error {
    if priority < 0 || priority > 100 {
        return fmt.Errorf("priority must be between 0 and 100, got %d", priority)
    }
    return nil
}
```

**步骤**:
1. 添加验证函数
2. 在`handleFilteringAddURL`中调用
3. 在`handleFilteringSetURL`中调用
4. 添加错误消息
5. 添加单元测试

---

## 🛠️ 开发环境设置

### 前端开发
```bash
cd client
npm install
npm run dev  # 开发模式
npm run build-prod  # 生产构建
```

### 后端开发
```bash
go build -o AdGuardHome_v3.exe
.\AdGuardHome_v3.exe
```

### 全平台构建
```bash
.\build-all-platforms.bat
```

---

## 📝 提交规范

### 提交消息格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

### 类型（type）
- `feat`: 新功能
- `fix`: Bug修复
- `refactor`: 重构
- `perf`: 性能优化
- `docs`: 文档更新
- `test`: 测试相关
- `chore`: 构建/工具相关

### 示例
```bash
git commit -m "fix(dns-routing): remove duplicate domain matching logic

Extracted common domain matching function to domain_utils.go
to eliminate code duplication between upstream_groups.go and process.go.

Fixes #4 from CODE_REVIEW_REPORT.md"
```

---

## 🧪 测试

### 运行测试
```bash
# 所有测试
go test ./...

# 特定包
go test ./internal/dnsforward/...

# 带覆盖率
go test -cover ./...

# 基准测试
go test -bench=. ./internal/dnsforward/
```

### 前端测试
```bash
cd client
npm test
```

---

## 📊 进度跟踪

### 本周目标（第1周）
- [ ] 问题#4: 提取公共函数
- [ ] 问题#7: 移除类型断言
- [ ] 问题#9: 删除未使用函数
- [ ] 问题#11: 添加输入验证
- [ ] 问题#12: 添加加载状态
- [ ] 功能1: Priority范围验证

### 完成标准
- ✅ 代码通过所有测试
- ✅ 代码通过lint检查
- ✅ 添加了必要的文档
- ✅ 提交消息符合规范

---

## 🔗 相关文档

- `V3_DEVELOPMENT_PLAN.md` - 完整开发计划
- `CODE_REVIEW_REPORT.md` - 代码审查报告
- `FEATURE_PRIORITY_COMPLETE.md` - Priority功能文档
- `RELEASE_NOTES_V2.md` - V2发布说明

---

## 💡 提示

### 开发技巧
1. 小步提交，频繁推送
2. 每个任务创建独立的commit
3. 先写测试，再写代码（TDD）
4. 使用`git stash`保存临时工作

### 调试技巧
1. 使用`--verbose`标志查看详细日志
2. 检查`AdGuardHome.yaml`配置
3. 查看浏览器控制台错误
4. 使用Go的`-race`标志检测竞态条件

### 性能分析
```bash
# CPU性能分析
go test -cpuprofile=cpu.prof -bench=.

# 内存性能分析
go test -memprofile=mem.prof -bench=.

# 查看分析结果
go tool pprof cpu.prof
```

---

## 🚀 开始开发

选择一个任务，开始编码吧！建议从简单的任务开始：

1. **最简单**: 问题#9（删除未使用函数）- 1小时
2. **简单**: 问题#11（添加输入验证）- 2小时
3. **中等**: 问题#4（提取公共函数）- 2小时
4. **中等**: 问题#12（添加加载状态）- 2小时
5. **中等**: 功能1（Priority验证）- 2小时
6. **复杂**: 问题#7（移除类型断言）- 3小时

祝开发顺利！🎉
