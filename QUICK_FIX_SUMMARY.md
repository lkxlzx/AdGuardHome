# 快速修复总结

## 修复的问题

1. ✅ **前端转圈问题**：添加了 30 秒的 API 超时配置
2. ✅ **格式统一**：将 `upstreams` 改为数组格式，与 `bootstrap_dns` 保持一致
3. ✅ **自动转换**：前端自动处理字符串和数组的转换

## 修改的文件

### 后端
1. `internal/dnsforward/config.go` - `Upstreams` 改为 `[]string`
2. `internal/dnsforward/upstream_groups.go` - 更新处理逻辑
3. `internal/dnsforward/dnsforward.go` - 更新默认组逻辑

### 前端
4. `client/src/actions/dnsConfig.ts` - 添加数据转换
5. `client/src/reducers/dnsConfig.ts` - 添加数据转换
6. `client/src/api/Api.ts` - 添加超时配置

### 配置
7. `AdGuardHome.yaml` - 更新为数组格式

## 新的格式

现在 `upstreams` 和 `bootstrap_dns` 使用相同的数组格式：

```yaml
bootstrap_dns:
  - 9.9.9.10
  - 149.112.112.10

upstream_groups:
  - id: group_xxx
    name: 国内DNS
    upstreams:
      - 114.114.114.114
      - 223.5.5.5
```

## 如何使用

在前端表单中，每行输入一个上游服务器地址：

```
114.114.114.114
223.5.5.5
```

保存后会自动转换为数组格式并保存到配置文件。
