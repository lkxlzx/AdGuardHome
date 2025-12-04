// DNS 上游分组相关的 TypeScript 类型定义

export interface UpstreamGroup {
    id: string;                    // 分组唯一标识符
    name: string;                  // 分组名称
    enabled: boolean;              // 是否启用
    is_default: boolean;           // 是否为默认分组
    upstream_dns: string[];        // 上游 DNS 服务器列表
    fallback_dns?: string[];       // 后备 DNS 服务器列表（可选）
    bootstrap_dns?: string[];      // Bootstrap DNS 服务器列表（可选）
    created_at: string;            // 创建时间
    updated_at: string;            // 更新时间
}

export interface UpstreamTestResult {
    upstream: string;              // 上游服务器地址
    success: boolean;              // 是否测试成功
    error?: string;                // 错误信息（如果失败）
    rtt?: number;                  // 响应时间（毫秒）
}

export interface TestResult {
    group_id: string;
    results: UpstreamTestResult[];
}

export interface UpstreamGroupsState {
    groups: UpstreamGroup[];       // 分组列表
    processing: boolean;           // 是否正在加载
    processingAdd: boolean;        // 是否正在添加
    processingUpdate: boolean;     // 是否正在更新
    processingDelete: boolean;     // 是否正在删除
    processingTest: boolean;       // 是否正在测试
    isModalOpen: boolean;          // 对话框是否打开
    modalType: 'add' | 'edit';     // 对话框类型
    currentGroup?: UpstreamGroup;  // 当前编辑的分组
    testResult?: TestResult;       // 测试结果
}
