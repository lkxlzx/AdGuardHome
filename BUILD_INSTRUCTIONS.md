# 编译说明

## 最新编译信息

- **编译时间**：2025年11月24日 19:07
- **可执行文件**：`AdGuardHome.exe`
- **文件大小**：32.2 MB
- **包含修改**：
  - ✅ Upstream Groups 格式统一（数组格式）
  - ✅ API 超时配置（30秒）
  - ✅ 前端数据自动转换
  - ✅ 后端数据结构更新

## 编译步骤

### 1. 编译前端

```bash
cd client
npm run build-prod
```

编译结果会输出到 `build/static/` 目录。

### 2. 生成 Go 嵌入文件

```bash
go generate ./...
```

这会将前端资源嵌入到 Go 代码中。

### 3. 编译后端

```bash
go build -o AdGuardHome.exe -ldflags="-s -w"
```

参数说明：
- `-o AdGuardHome.exe`：指定输出文件名
- `-ldflags="-s -w"`：去除调试信息，减小文件大小

## 验证编译结果

```bash
.\AdGuardHome.exe --version
```

## 运行程序

```bash
.\AdGuardHome.exe
```

默认会在 `http://localhost:80` 启动管理界面。

## 配置文件

程序会自动读取 `AdGuardHome.yaml` 配置文件。新版本支持的 `upstream_groups` 格式：

```yaml
upstream_groups:
  - id: group_xxx
    name: 国内DNS
    upstreams:
      - 114.114.114.114
      - 223.5.5.5
    enabled: true
    is_default: true
```

## 注意事项

1. **配置兼容性**：新版本会自动处理旧格式的配置文件
2. **数据迁移**：如果你的配置文件中 `upstreams` 是字符串格式，需要手动改为数组格式
3. **前端缓存**：如果前端没有更新，请清除浏览器缓存（Ctrl+F5）

## 开发模式

如果需要开发调试，可以使用：

```bash
# 前端开发服务器（热重载）
cd client
npm run watch:hot

# 后端（另一个终端）
go run .
```

## 故障排除

### 前端编译失败

```bash
cd client
npm install
npm run build-prod
```

### 后端编译失败

```bash
go mod tidy
go build -o AdGuardHome.exe
```

### 运行时错误

检查 `AdGuardHome.yaml` 格式是否正确，特别是 `upstream_groups` 部分。
