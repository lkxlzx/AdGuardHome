# 上游组测试功能实现方案

## 需求描述

在上游组管理表格中添加"检测"列，允许用户单独测试每个上游组的有效性。

## UI设计

### 表格列结构

```
| 已启用 | 分组名称 | 检测 | 上游服务器地址 | 操作 |
|--------|----------|------|----------------|------|
| ✓      | 国内     | [测试]| 114.114.114.114| [✓][✎][🗑] |
| ✓      | 海外     | [测试]| 8.8.8.8, 1.1.1.1| [✓][✎][🗑] |
```

### 测试按钮状态

1. **默认状态**：显示"测试"按钮
2. **测试中**：显示加载动画 + "测试中..."
3. **成功**：显示绿色✓图标 + 响应时间（如"45ms"）
4. **失败**：显示红色✗图标 + 错误信息

## 技术实现

### 1. 后端API

#### 新增API端点

**POST /control/test_upstream_group**

请求体：
```json
{
  "group_id": "group_1763970331409",
  "upstreams": ["114.114.114.114", "223.5.5.5"]
}
```

响应：
```json
{
  "status": "ok",
  "results": [
    {
      "upstream": "114.114.114.114",
      "status": "ok",
      "response_time": "45ms",
      "error": ""
    },
    {
      "upstream": "223.5.5.5",
      "status": "ok",
      "response_time": "52ms",
      "error": ""
    }
  ],
  "summary": {
    "total": 2,
    "success": 2,
    "failed": 0,
    "avg_response_time": "48.5ms"
  }
}
```

#### 实现文件

**internal/dnsforward/http_upstream_test.go** (新建)

```go
package dnsforward

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
    "github.com/AdguardTeam/dnsproxy/upstream"
)

// testUpstreamGroupRequest represents the request to test an upstream group
type testUpstreamGroupRequest struct {
    GroupID   string   `json:"group_id"`
    Upstreams []string `json:"upstreams"`
}

// upstreamTestResult represents the test result for a single upstream
type upstreamTestResult struct {
    Upstream     string `json:"upstream"`
    Status       string `json:"status"` // "ok" or "error"
    ResponseTime string `json:"response_time"`
    Error        string `json:"error"`
}

// testUpstreamGroupResponse represents the response of upstream group test
type testUpstreamGroupResponse struct {
    Status  string               `json:"status"`
    Results []upstreamTestResult `json:"results"`
    Summary struct {
        Total           int    `json:"total"`
        Success         int    `json:"success"`
        Failed          int    `json:"failed"`
        AvgResponseTime string `json:"avg_response_time"`
    } `json:"summary"`
}

// handleTestUpstreamGroup handles the POST /control/test_upstream_group request
func (s *Server) handleTestUpstreamGroup(w http.ResponseWriter, r *http.Request) {
    req := &testUpstreamGroupRequest{}
    
    if err := json.NewDecoder(r.Body).Decode(req); err != nil {
        aghhttp.Error(r, w, http.StatusBadRequest, "invalid request body: %s", err)
        return
    }
    
    if len(req.Upstreams) == 0 {
        aghhttp.Error(r, w, http.StatusBadRequest, "no upstreams provided")
        return
    }
    
    // Test each upstream
    results := make([]upstreamTestResult, 0, len(req.Upstreams))
    var totalTime time.Duration
    successCount := 0
    
    for _, upstreamAddr := range req.Upstreams {
        result := s.testSingleUpstream(upstreamAddr)
        results = append(results, result)
        
        if result.Status == "ok" {
            successCount++
            // Parse response time
            if duration, err := time.ParseDuration(result.ResponseTime); err == nil {
                totalTime += duration
            }
        }
    }
    
    // Calculate average response time
    avgResponseTime := "N/A"
    if successCount > 0 {
        avgDuration := totalTime / time.Duration(successCount)
        avgResponseTime = avgDuration.Round(time.Millisecond).String()
    }
    
    // Build response
    resp := &testUpstreamGroupResponse{
        Status:  "ok",
        Results: results,
    }
    resp.Summary.Total = len(req.Upstreams)
    resp.Summary.Success = successCount
    resp.Summary.Failed = len(req.Upstreams) - successCount
    resp.Summary.AvgResponseTime = avgResponseTime
    
    aghhttp.WriteJSONResponse(w, r, resp)
}

// testSingleUpstream tests a single upstream server
func (s *Server) testSingleUpstream(upstreamAddr string) upstreamTestResult {
    result := upstreamTestResult{
        Upstream: upstreamAddr,
        Status:   "error",
    }
    
    // Create upstream
    u, err := upstream.AddressToUpstream(upstreamAddr, &upstream.Options{
        Timeout: 5 * time.Second,
    })
    if err != nil {
        result.Error = fmt.Sprintf("failed to create upstream: %s", err)
        return result
    }
    
    // Test with a simple DNS query (google.com A record)
    req := createTestMessage()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    start := time.Now()
    _, err = u.Exchange(req)
    elapsed := time.Since(start)
    
    if err != nil {
        result.Error = fmt.Sprintf("query failed: %s", err)
        return result
    }
    
    result.Status = "ok"
    result.ResponseTime = elapsed.Round(time.Millisecond).String()
    return result
}

// createTestMessage creates a test DNS query message
func createTestMessage() *dns.Msg {
    req := &dns.Msg{
        MsgHdr: dns.MsgHdr{
            Id:               dns.Id(),
            RecursionDesired: true,
        },
        Question: []dns.Question{{
            Name:   "google.com.",
            Qtype:  dns.TypeA,
            Qclass: dns.ClassINET,
        }},
    }
    return req
}
```

#### 注册路由

**internal/dnsforward/http.go**

```go
// 在 registerHandlers() 函数中添加：
s.mux.HandleFunc("/control/test_upstream_group", s.handleTestUpstreamGroup)
```

### 2. 前端实现

#### 2.1 Redux Action

**client/src/actions/index.ts**

```typescript
// 添加新的action types
export const TEST_UPSTREAM_GROUP_REQUEST = 'TEST_UPSTREAM_GROUP_REQUEST';
export const TEST_UPSTREAM_GROUP_SUCCESS = 'TEST_UPSTREAM_GROUP_SUCCESS';
export const TEST_UPSTREAM_GROUP_FAILURE = 'TEST_UPSTREAM_GROUP_FAILURE';

// 添加action creator
export const testUpstreamGroup = (groupId: string, upstreams: string[]) => async (dispatch: any, getState: any) => {
    dispatch({ type: TEST_UPSTREAM_GROUP_REQUEST, payload: { groupId } });
    
    try {
        const response = await apiClient.post('/control/test_upstream_group', {
            group_id: groupId,
            upstreams: upstreams,
        });
        
        dispatch({
            type: TEST_UPSTREAM_GROUP_SUCCESS,
            payload: {
                groupId,
                result: response.data,
            },
        });
        
        return response.data;
    } catch (error) {
        dispatch({
            type: TEST_UPSTREAM_GROUP_FAILURE,
            payload: {
                groupId,
                error: error.message,
            },
        });
        throw error;
    }
};
```

#### 2.2 Redux Reducer

**client/src/reducers/upstreamGroups.ts**

```typescript
// 添加到state
interface UpstreamGroupsState {
    // ... 现有字段
    testResults: {
        [groupId: string]: {
            testing: boolean;
            result?: TestResult;
            error?: string;
        };
    };
}

// 添加reducer处理
case TEST_UPSTREAM_GROUP_REQUEST:
    return {
        ...state,
        testResults: {
            ...state.testResults,
            [action.payload.groupId]: {
                testing: true,
            },
        },
    };

case TEST_UPSTREAM_GROUP_SUCCESS:
    return {
        ...state,
        testResults: {
            ...state.testResults,
            [action.payload.groupId]: {
                testing: false,
                result: action.payload.result,
            },
        },
    };

case TEST_UPSTREAM_GROUP_FAILURE:
    return {
        ...state,
        testResults: {
            ...state.testResults,
            [action.payload.groupId]: {
                testing: false,
                error: action.payload.error,
            },
        },
    };
```

#### 2.3 表格组件修改

**client/src/components/Settings/Dns/Upstream/UpstreamGroupsTable.tsx**

```typescript
// 添加测试列
{
    Header: this.props.t('test_upstream_group'),
    accessor: 'test',
    maxWidth: 120,
    sortable: false,
    resizable: false,
    Cell: (row: any) => {
        const { original } = row;
        const testState = this.props.testResults[original.id] || {};
        
        return (
            <div className="logs__row logs__row--center">
                {testState.testing ? (
                    <button
                        type="button"
                        className="btn btn-sm btn-outline-secondary"
                        disabled>
                        <span className="spinner-border spinner-border-sm mr-2" />
                        {this.props.t('testing')}
                    </button>
                ) : testState.result ? (
                    <button
                        type="button"
                        className={`btn btn-sm ${
                            testState.result.summary.failed === 0
                                ? 'btn-outline-success'
                                : 'btn-outline-warning'
                        }`}
                        onClick={() => this.props.handleTest(original)}
                        title={`${testState.result.summary.success}/${testState.result.summary.total} OK, ${testState.result.summary.avg_response_time}`}>
                        {testState.result.summary.failed === 0 ? (
                            <svg className="icons icon12">
                                <use xlinkHref="#check" />
                            </svg>
                        ) : (
                            <svg className="icons icon12">
                                <use xlinkHref="#cross" />
                            </svg>
                        )}
                        <span className="ml-1">{testState.result.summary.avg_response_time}</span>
                    </button>
                ) : (
                    <button
                        type="button"
                        className="btn btn-sm btn-outline-primary"
                        onClick={() => this.props.handleTest(original)}
                        disabled={this.props.processing}
                        title={this.props.t('test_upstream_group')}>
                        <svg className="icons icon12">
                            <use xlinkHref="#test" />
                        </svg>
                        <span className="ml-1">{this.props.t('test')}</span>
                    </button>
                )}
            </div>
        );
    },
},
```

#### 2.4 Container组件

**client/src/components/Settings/Dns/Upstream/UpstreamGroupsContainer.tsx**

```typescript
// 添加mapStateToProps
const mapStateToProps = (state: RootState) => ({
    // ... 现有映射
    testResults: state.upstreamGroups.testResults,
});

// 添加mapDispatchToProps
const mapDispatchToProps = {
    // ... 现有映射
    handleTest: (group: UpstreamGroup) => testUpstreamGroup(group.id, group.upstreams),
};
```

### 3. 国际化

**client/src/__locales/zh-cn.json**

```json
{
    "test_upstream_group": "检测",
    "test": "测试",
    "testing": "测试中...",
    "test_success": "测试成功",
    "test_failed": "测试失败"
}
```

**client/src/__locales/en.json**

```json
{
    "test_upstream_group": "Test",
    "test": "Test",
    "testing": "Testing...",
    "test_success": "Test Successful",
    "test_failed": "Test Failed"
}
```

## 实现步骤

### 阶段1：后端API（30分钟）
1. 创建 `internal/dnsforward/http_upstream_test.go`
2. 实现 `handleTestUpstreamGroup` 和 `testSingleUpstream`
3. 在 `http.go` 中注册路由
4. 测试API接口

### 阶段2：前端Redux（20分钟）
1. 添加action types和action creators
2. 修改reducer添加testResults状态
3. 测试Redux流程

### 阶段3：前端UI（30分钟）
1. 修改 `UpstreamGroupsTable.tsx` 添加测试列
2. 修改 `UpstreamGroupsContainer.tsx` 连接Redux
3. 添加国际化文本
4. 测试UI交互

### 阶段4：测试和优化（20分钟）
1. 端到端测试
2. 错误处理优化
3. UI细节调整

## 预期效果

1. **用户体验**：点击测试按钮后，立即显示"测试中..."状态
2. **测试结果**：显示每个上游的测试结果和平均响应时间
3. **错误提示**：如果测试失败，显示具体错误信息
4. **性能**：测试超时5秒，避免长时间等待

## 注意事项

1. **并发测试**：多个上游串行测试，避免网络拥塞
2. **缓存结果**：测试结果保存在Redux中，刷新页面后清除
3. **权限控制**：复用现有的API认证机制
4. **测试域名**：使用 `google.com` 作为测试域名，确保全球可达

---

*预计总开发时间：1.5-2小时*
*难度：中等*
