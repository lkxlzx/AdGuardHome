# DNS 上游分组功能测试指南

## 快速开始

本指南帮助你快速运行 DNS 上游分组功能的所有测试。

---

## 前置要求

### 前端测试
- Node.js 16+
- npm 或 yarn

### 后端测试
- Go 1.21+

### E2E 测试
- Playwright (会自动安装)
- 运行中的开发服务器

---

## 运行测试

### 1. 前端单元测试

```bash
# 进入前端目录
cd client

# 安装依赖（首次运行）
npm install

# 运行所有单元测试
npm test

# 监听模式（开发时推荐）
npm run test:watch

# 生成覆盖率报告
npm test -- --coverage
```

**预期输出**:
```
✓ client/src/reducers/upstreamGroups.test.ts (27)
✓ client/src/selectors/upstreamGroups.test.ts (15)

Test Files  2 passed (2)
     Tests  42 passed (42)
```

---

### 2. E2E 测试

```bash
# 进入前端目录
cd client

# 安装 Playwright（首次运行）
npx playwright install

# 运行所有 E2E 测试
npm run test:e2e

# 交互式运行（推荐）
npm run test:e2e:interactive

# 调试模式
npm run test:e2e:debug

# 只运行特定浏览器
npx playwright test --project=chromium
```

**注意**: E2E 测试会自动启动开发服务器，无需手动启动。

**预期输出**:
```
Running 17 tests using 1 worker

✓ [chromium] › upstreamGroups.spec.ts:创建分组流程 (4)
✓ [chromium] › upstreamGroups.spec.ts:编辑分组流程 (2)
✓ [chromium] › upstreamGroups.spec.ts:删除分组流程 (4)
✓ [chromium] › upstreamGroups.spec.ts:设置默认分组流程 (2)
✓ [chromium] › upstreamGroups.spec.ts:测试分组连通性流程 (2)
✓ [chromium] › upstreamGroups.spec.ts:启用/禁用分组 (1)
✓ [chromium] › upstreamGroups.spec.ts:复制分组 (1)
✓ [chromium] › upstreamGroups.spec.ts:空状态 (1)

17 passed (30s)
```

---

### 3. 后端测试

```bash
# 在项目根目录

# 运行所有后端测试
go test ./internal/home/...

# 运行特定测试文件
go test ./internal/home/ -run TestHandleGetUpstreamGroups

# 详细输出
go test ./internal/home/ -v

# 生成覆盖率报告
go test ./internal/home/ -coverprofile=coverage.out
go tool cover -html=coverage.out

# 运行所有测试并显示覆盖率
go test ./internal/home/ -cover
```

**预期输出**:
```
=== RUN   TestValidateUpstreamGroupRequest
=== RUN   TestHandleGetUpstreamGroups
=== RUN   TestHandleAddUpstreamGroup
=== RUN   TestHandleUpdateUpstreamGroup
=== RUN   TestHandleDeleteUpstreamGroup
=== RUN   TestHandleSetDefaultGroup
=== RUN   TestDefaultGroupUniqueness
--- PASS: TestValidateUpstreamGroupRequest (0.00s)
--- PASS: TestHandleGetUpstreamGroups (0.01s)
--- PASS: TestHandleAddUpstreamGroup (0.02s)
--- PASS: TestHandleUpdateUpstreamGroup (0.01s)
--- PASS: TestHandleDeleteUpstreamGroup (0.01s)
--- PASS: TestHandleSetDefaultGroup (0.01s)
--- PASS: TestDefaultGroupUniqueness (0.01s)
PASS
coverage: 85.2% of statements
ok      github.com/AdguardTeam/AdGuardHome/internal/home        0.123s
```

---

## 运行所有测试

### 一键运行脚本

创建一个脚本来运行所有测试：

```bash
#!/bin/bash
# run-all-tests.sh

echo "=== 运行前端单元测试 ==="
cd client
npm test
if [ $? -ne 0 ]; then
    echo "❌ 前端单元测试失败"
    exit 1
fi

echo ""
echo "=== 运行后端测试 ==="
cd ..
go test ./internal/home/...
if [ $? -ne 0 ]; then
    echo "❌ 后端测试失败"
    exit 1
fi

echo ""
echo "=== 运行 E2E 测试 ==="
cd client
npm run test:e2e
if [ $? -ne 0 ]; then
    echo "❌ E2E 测试失败"
    exit 1
fi

echo ""
echo "✅ 所有测试通过！"
```

使用方法：
```bash
chmod +x run-all-tests.sh
./run-all-tests.sh
```

---

## 测试覆盖率

### 查看前端覆盖率

```bash
cd client
npm test -- --coverage

# 生成 HTML 报告
npm test -- --coverage --reporter=html

# 打开报告
open coverage/index.html  # macOS
xdg-open coverage/index.html  # Linux
start coverage/index.html  # Windows
```

### 查看后端覆盖率

```bash
go test ./internal/home/ -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 常见问题

### Q: 前端测试失败，提示找不到模块

**A**: 确保已安装依赖：
```bash
cd client
npm install
```

### Q: E2E 测试失败，提示浏览器未安装

**A**: 安装 Playwright 浏览器：
```bash
cd client
npx playwright install
```

### Q: E2E 测试超时

**A**: 增加超时时间或检查开发服务器是否正常启动：
```bash
# 手动启动开发服务器
cd client
npm run watch:hot

# 在另一个终端运行测试
npm run test:e2e
```

### Q: 后端测试失败，提示配置错误

**A**: 确保测试环境正确初始化，检查 `testutil.CleanupAndRequireSuccess` 调用。

### Q: 如何只运行特定的测试？

**A**: 
```bash
# 前端 - 使用 test.only
# 在测试文件中：it.only('test name', ...)

# 后端 - 使用 -run 标志
go test ./internal/home/ -run TestHandleGetUpstreamGroups

# E2E - 使用 --grep
npx playwright test --grep "创建分组"
```

---

## 调试测试

### 调试前端测试

```bash
# 使用 VS Code 调试器
# 在 .vscode/launch.json 中添加配置：
{
  "type": "node",
  "request": "launch",
  "name": "Vitest",
  "runtimeExecutable": "npm",
  "runtimeArgs": ["run", "test:watch"],
  "console": "integratedTerminal"
}
```

### 调试 E2E 测试

```bash
# 使用 Playwright Inspector
npm run test:e2e:debug

# 或者使用 UI 模式
npm run test:e2e:interactive
```

### 调试后端测试

```bash
# 使用 delve
go install github.com/go-delve/delve/cmd/dlv@latest
dlv test ./internal/home/ -- -test.run TestHandleGetUpstreamGroups
```

---

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Tests

on: [push, pull_request]

jobs:
  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install dependencies
        run: cd client && npm ci
      - name: Run unit tests
        run: cd client && npm test
      - name: Run E2E tests
        run: cd client && npm run test:e2e

  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        run: go test ./internal/home/... -cover
```

---

## 测试最佳实践

### 编写测试时

1. ✅ 使用描述性的测试名称
2. ✅ 每个测试只验证一个行为
3. ✅ 使用 AAA 模式 (Arrange, Act, Assert)
4. ✅ 测试边界条件和错误情况
5. ✅ 保持测试独立，不依赖执行顺序

### 运行测试时

1. ✅ 提交前运行所有测试
2. ✅ 定期检查测试覆盖率
3. ✅ 修复失败的测试，不要忽略
4. ✅ 保持测试快速运行

---

## 获取帮助

如果遇到问题：

1. 查看 [TESTING_SUMMARY.md](./TESTING_SUMMARY.md) 了解测试详情
2. 查看测试文件中的注释
3. 运行 `npm test -- --help` 或 `go test -h` 查看更多选项

---

**最后更新**: 2024-12-04  
**维护者**: DNS 上游分组功能团队
