# 上游组测试功能实现总结

## 已完成

### ✅ 后端API（100%）

1. **创建测试接口文件**
   - `internal/dnsforward/http_upstream_test.go`
   - 实现 `handleTestUpstreamGroup` 处理测试请求
   - 实现 `testSingleUpstream` 测试单个上游
   - 实现 `createTestMessage` 创建测试DNS查询

2. **注册路由**
   - 在 `internal/dnsforward/http.go` 中添加路由
   - `POST /control/test_upstream_group`

### API接口说明

**请求：**
```bash
POST /control/test_upstream_group
Content-Type: application/json

{
  "group_id": "group_1763970331409",
  "upstreams": ["114.114.114.114", "223.5.5.5"]
}
```

**响应：**
```json
{
  "status": "ok",
  "results": [
    {
      "upstream": "114.114.114.114",
      "status": "ok",
      "response_time": "45ms",
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

## 待实现（前端部分）

### 📝 需要手动完成的步骤

#### 1. 添加API客户端方法

**文件：** `client/src/api/Api.ts`

```typescript
// 在Api类中添加方法
async testUpstreamGroup(groupId: string, upstreams: string[]) {
    const path = '/control/test_upstream_group';
    const data = {
        group_id: groupId,
        upstreams: upstreams,
    };
    return this.makeRequest(path, 'POST', data);
}
```

#### 2. 添加Redux Actions

**文件：** `client/src/actions/dnsConfig.ts`

在文件末尾添加：

```typescript
// Test upstream group actions
export const testUpstreamGroupRequest = createAction('TEST_UPSTREAM_GROUP_REQUEST');
export const testUpstreamGroupFailure = createAction('TEST_UPSTREAM_GROUP_FAILURE');
export const testUpstreamGroupSuccess = createAction('TEST_UPSTREAM_GROUP_SUCCESS');

export const testUpstreamGroup = (groupId: string, upstreams: string[]) => async (dispatch: any) => {
    dispatch(testUpstreamGroupRequest({ groupId }));
    try {
        const data = await apiClient.testUpstreamGroup(groupId, upstreams);
        dispatch(testUpstreamGroupSuccess({ groupId, result: data }));
        return data;
    } catch (error) {
        dispatch(testUpstreamGroupFailure({ groupId, error }));
        dispatch(addErrorToast({ error }));
        throw error;
    }
};
```

#### 3. 修改Reducer

**文件：** `client/src/reducers/dnsConfig.ts`

在state接口中添加：

```typescript
interface DnsConfigState {
    // ... 现有字段
    upstreamGroupTests: {
        [groupId: string]: {
            testing: boolean;
            result?: any;
            error?: string;
        };
    };
}
```

在reducer中添加处理：

```typescript
case 'TEST_UPSTREAM_GROUP_REQUEST':
    return {
        ...state,
        upstreamGroupTests: {
            ...state.upstreamGroupTests,
            [action.payload.groupId]: {
                testing: true,
            },
        },
    };

case 'TEST_UPSTREAM_GROUP_SUCCESS':
    return {
        ...state,
        upstreamGroupTests: {
            ...state.upstreamGroupTests,
            [action.payload.groupId]: {
                testing: false,
                result: action.payload.result,
            },
        },
    };

case 'TEST_UPSTREAM_GROUP_FAILURE':
    return {
        ...state,
        upstreamGroupTests: {
            ...state.upstreamGroupTests,
            [action.payload.groupId]: {
                testing: false,
                error: action.payload.error,
            },
        },
    };
```

#### 4. 修改表格组件

**文件：** `client/src/components/Settings/Dns/Upstream/UpstreamGroupsTable.tsx`

在columns数组中，在"分组名称"列之后添加：

```typescript
{
    Header: this.props.t('test_upstream_group'),
    accessor: 'test',
    maxWidth: 120,
    sortable: false,
    resizable: false,
    Cell: (row: any) => {
        const { original } = row;
        const testState = this.props.upstreamGroupTests?.[original.id] || {};
        
        if (testState.testing) {
            return (
                <div className="logs__row logs__row--center">
                    <button type="button" className="btn btn-sm btn-outline-secondary" disabled>
                        <span className="spinner-border spinner-border-sm mr-2" />
                        {this.props.t('testing')}
                    </button>
                </div>
            );
        }
        
        if (testState.result) {
            const { summary } = testState.result;
            const allSuccess = summary.failed === 0;
            
            return (
                <div className="logs__row logs__row--center">
                    <button
                        type="button"
                        className={`btn btn-sm ${allSuccess ? 'btn-outline-success' : 'btn-outline-warning'}`}
                        onClick={() => this.props.handleTest(original)}
                        title={`${summary.success}/${summary.total} OK, ${summary.avg_response_time}`}>
                        <svg className="icons icon12">
                            <use xlinkHref={allSuccess ? '#check' : '#cross'} />
                        </svg>
                        <span className="ml-1">{summary.avg_response_time}</span>
                    </button>
                </div>
            );
        }
        
        return (
            <div className="logs__row logs__row--center">
                <button
                    type="button"
                    className="btn btn-sm btn-outline-primary"
                    onClick={() => this.props.handleTest(original)}
                    disabled={this.props.processing}
                    title={this.props.t('test_upstream_group')}>
                    <svg className="icons icon12">
                        <use xlinkHref="#play" />
                    </svg>
                    <span className="ml-1">{this.props.t('test')}</span>
                </button>
            </div>
        );
    },
},
```

#### 5. 修改Container组件

**文件：** `client/src/components/Settings/Dns/Upstream/UpstreamGroupsContainer.tsx`

添加props映射：

```typescript
const mapStateToProps = (state: RootState) => ({
    // ... 现有映射
    upstreamGroupTests: state.dnsConfig.upstreamGroupTests,
});

const mapDispatchToProps = {
    // ... 现有映射
    handleTest: (group: UpstreamGroup) => testUpstreamGroup(group.id, group.upstreams),
};
```

#### 6. 添加国际化文本

**文件：** `client/src/__locales/zh-cn.json`

```json
{
    "test_upstream_group": "检测",
    "test": "测试",
    "testing": "测试中...",
    "test_success": "测试成功",
    "test_failed": "测试失败"
}
```

**文件：** `client/src/__locales/en.json`

```json
{
    "test_upstream_group": "Test",
    "test": "Test",
    "testing": "Testing...",
    "test_success": "Test Successful",
    "test_failed": "Test Failed"
}
```

## 测试步骤

### 1. 编译后端
```bash
go build -o AdGuardHome.exe
```

### 2. 编译前端
```bash
cd client
npm run build
```

### 3. 启动服务
```bash
.\AdGuardHome.exe
```

### 4. 测试功能
1. 打开浏览器访问 AdGuardHome
2. 进入"设置" → "DNS设置" → "上游DNS服务器"
3. 在上游组表格中，点击任意组的"测试"按钮
4. 观察测试过程和结果显示

## 预期效果

1. **点击测试按钮**：按钮变为"测试中..."并显示加载动画
2. **测试完成**：显示测试结果（✓或✗）和平均响应时间
3. **再次点击**：重新测试，更新结果

## 注意事项

1. 测试使用 `google.com` 作为测试域名
2. 每个上游超时时间为5秒
3. 测试结果保存在Redux中，刷新页面后清除
4. 如果所有上游都测试失败，显示警告图标

---

*后端实现已完成，前端需要按照上述步骤手动完成*
*预计前端实现时间：30-40分钟*
