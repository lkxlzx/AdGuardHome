import { test, expect } from '@playwright/test';

// 测试配置
const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
const API_URL = process.env.API_URL || 'http://localhost:3000';

test.describe('DNS Upstream Groups', () => {
    test.beforeEach(async ({ page }) => {
        // 访问 DNS 设置页面
        await page.goto(`${BASE_URL}/#/dns`);
        await page.waitForLoadState('networkidle');
    });

    test.describe('创建分组流程', () => {
        test('应该能够打开创建分组对话框', async ({ page }) => {
            // 点击"添加DNS上游分组"按钮
            await page.click('button:has-text("添加DNS上游分组")');
            
            // 验证对话框打开
            await expect(page.locator('.ReactModal__Content')).toBeVisible();
            await expect(page.locator('text=添加分组')).toBeVisible();
        });

        test('应该能够创建新分组', async ({ page }) => {
            // 模拟 API 响应
            await page.route(`${API_URL}/control/dns/upstream_groups`, async (route) => {
                if (route.request().method() === 'POST') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify({
                            id: 'new-group-id',
                            name: 'Test Group',
                            enabled: true,
                            is_default: false,
                            upstream_dns: ['8.8.8.8', '1.1.1.1'],
                            created_at: new Date().toISOString(),
                            updated_at: new Date().toISOString(),
                        }),
                    });
                }
            });

            // 打开对话框
            await page.click('button:has-text("添加DNS上游分组")');
            
            // 填写表单
            await page.fill('input[name="name"]', 'Test Group');
            await page.fill('textarea[name="upstream_dns"]', '8.8.8.8\n1.1.1.1');
            
            // 保存
            await page.click('button:has-text("保存")');
            
            // 验证对话框关闭
            await expect(page.locator('.ReactModal__Content')).not.toBeVisible();
            
            // 验证成功提示
            await expect(page.locator('.toast--success')).toBeVisible();
        });

        test('应该验证必填字段', async ({ page }) => {
            // 打开对话框
            await page.click('button:has-text("添加DNS上游分组")');
            
            // 不填写任何内容，直接点击保存
            await page.click('button:has-text("保存")');
            
            // 验证错误提示
            await expect(page.locator('text=分组名称不能为空')).toBeVisible();
            await expect(page.locator('text=至少需要一个上游DNS服务器')).toBeVisible();
        });

        test('应该验证分组名称长度', async ({ page }) => {
            // 打开对话框
            await page.click('button:has-text("添加DNS上游分组")');
            
            // 输入超长名称
            const longName = 'a'.repeat(51);
            await page.fill('input[name="name"]', longName);
            await page.fill('textarea[name="upstream_dns"]', '8.8.8.8');
            
            // 点击保存
            await page.click('button:has-text("保存")');
            
            // 验证错误提示
            await expect(page.locator('text=分组名称不能超过50个字符')).toBeVisible();
        });
    });

    test.describe('编辑分组流程', () => {
        test('应该能够打开编辑对话框并预填充数据', async ({ page }) => {
            // 模拟已有分组
            await page.route(`${API_URL}/control/dns/upstream_groups`, async (route) => {
                if (route.request().method() === 'GET') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify([
                            {
                                id: 'group-1',
                                name: 'Existing Group',
                                enabled: true,
                                is_default: false,
                                upstream_dns: ['8.8.8.8'],
                                created_at: '2024-01-01T00:00:00Z',
                                updated_at: '2024-01-01T00:00:00Z',
                            },
                        ]),
                    });
                }
            });

            // 刷新页面加载分组
            await page.reload();
            await page.waitForLoadState('networkidle');
            
            // 点击编辑按钮
            await page.click('button[title="编辑"]');
            
            // 验证对话框打开并预填充
            await expect(page.locator('.ReactModal__Content')).toBeVisible();
            await expect(page.locator('input[name="name"]')).toHaveValue('Existing Group');
            await expect(page.locator('textarea[name="upstream_dns"]')).toHaveValue('8.8.8.8');
        });

        test('应该能够更新分组', async ({ page }) => {
            // 模拟 API
            await page.route(`${API_URL}/control/dns/upstream_groups/**`, async (route) => {
                if (route.request().method() === 'PUT') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify({
                            id: 'group-1',
                            name: 'Updated Group',
                            enabled: true,
                            is_default: false,
                            upstream_dns: ['1.1.1.1'],
                            created_at: '2024-01-01T00:00:00Z',
                            updated_at: new Date().toISOString(),
                        }),
                    });
                }
            });

            // 点击编辑按钮
            await page.click('button[title="编辑"]');
            
            // 修改数据
            await page.fill('input[name="name"]', 'Updated Group');
            await page.fill('textarea[name="upstream_dns"]', '1.1.1.1');
            
            // 保存
            await page.click('button:has-text("保存")');
            
            // 验证成功
            await expect(page.locator('.toast--success')).toBeVisible();
        });
    });

    test.describe('删除分组流程', () => {
        test('应该显示确认对话框', async ({ page }) => {
            // 点击删除按钮
            await page.click('button[title="删除"]');
            
            // 验证确认对话框
            await expect(page.locator('text=确认删除')).toBeVisible();
            await expect(page.locator('text=确定要删除分组')).toBeVisible();
        });

        test('应该能够删除非默认分组', async ({ page }) => {
            // 模拟 API
            await page.route(`${API_URL}/control/dns/upstream_groups/**`, async (route) => {
                if (route.request().method() === 'DELETE') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify({ success: true }),
                    });
                }
            });

            // 点击删除按钮
            await page.click('button[title="删除"]');
            
            // 确认删除
            await page.click('button:has-text("删除")');
            
            // 验证成功
            await expect(page.locator('.toast--success')).toBeVisible();
        });

        test('应该阻止删除默认分组', async ({ page }) => {
            // 模拟默认分组
            await page.route(`${API_URL}/control/dns/upstream_groups`, async (route) => {
                if (route.request().method() === 'GET') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify([
                            {
                                id: 'default-group',
                                name: 'Default Group',
                                enabled: true,
                                is_default: true,
                                upstream_dns: ['8.8.8.8'],
                                created_at: '2024-01-01T00:00:00Z',
                                updated_at: '2024-01-01T00:00:00Z',
                            },
                        ]),
                    });
                }
            });

            await page.reload();
            await page.waitForLoadState('networkidle');

            // 模拟删除失败
            await page.route(`${API_URL}/control/dns/upstream_groups/**`, async (route) => {
                if (route.request().method() === 'DELETE') {
                    await route.fulfill({
                        status: 409,
                        contentType: 'application/json',
                        body: JSON.stringify({
                            error: 'conflict',
                            message: '不能删除默认分组',
                        }),
                    });
                }
            });

            // 尝试删除
            await page.click('button[title="删除"]');
            await page.click('button:has-text("删除")');
            
            // 验证错误提示
            await expect(page.locator('.toast--error')).toBeVisible();
            await expect(page.locator('text=不能删除默认分组')).toBeVisible();
        });

        test('应该能够取消删除', async ({ page }) => {
            // 点击删除按钮
            await page.click('button[title="删除"]');
            
            // 取消删除
            await page.click('button:has-text("取消")');
            
            // 验证对话框关闭
            await expect(page.locator('text=确认删除')).not.toBeVisible();
        });
    });

    test.describe('设置默认分组流程', () => {
        test('应该能够设置默认分组', async ({ page }) => {
            // 模拟 API
            await page.route(`${API_URL}/control/dns/upstream_groups/**/default`, async (route) => {
                if (route.request().method() === 'POST') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify({ success: true }),
                    });
                }
            });

            // 勾选"设为默认分组"
            await page.check('input[type="checkbox"][name="is_default"]');
            
            // 验证成功
            await expect(page.locator('.toast--success')).toBeVisible();
        });

        test('应该显示默认标识', async ({ page }) => {
            // 模拟默认分组
            await page.route(`${API_URL}/control/dns/upstream_groups`, async (route) => {
                if (route.request().method() === 'GET') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify([
                            {
                                id: 'default-group',
                                name: 'Default Group',
                                enabled: true,
                                is_default: true,
                                upstream_dns: ['8.8.8.8'],
                                created_at: '2024-01-01T00:00:00Z',
                                updated_at: '2024-01-01T00:00:00Z',
                            },
                        ]),
                    });
                }
            });

            await page.reload();
            await page.waitForLoadState('networkidle');
            
            // 验证默认标识显示
            await expect(page.locator('.badge--success:has-text("默认")')).toBeVisible();
        });
    });

    test.describe('测试分组连通性流程', () => {
        test('应该能够测试分组', async ({ page }) => {
            // 模拟测试 API
            await page.route(`${API_URL}/control/dns/upstream_groups/**/test`, async (route) => {
                if (route.request().method() === 'POST') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify({
                            group_id: 'group-1',
                            results: [
                                { upstream: '8.8.8.8', success: true, rtt: 20 },
                                { upstream: '1.1.1.1', success: true, rtt: 15 },
                            ],
                        }),
                    });
                }
            });

            // 打开编辑对话框
            await page.click('button[title="编辑"]');
            
            // 点击检测按钮
            await page.click('button:has-text("检测")');
            
            // 验证加载状态
            await expect(page.locator('button:has-text("检测中")')).toBeVisible();
            
            // 等待测试完成
            await page.waitForSelector('text=测试完成');
            
            // 验证测试结果显示
            await expect(page.locator('text=8.8.8.8')).toBeVisible();
            await expect(page.locator('text=20ms')).toBeVisible();
            await expect(page.locator('text=1.1.1.1')).toBeVisible();
            await expect(page.locator('text=15ms')).toBeVisible();
        });

        test('应该显示测试失败的服务器', async ({ page }) => {
            // 模拟测试失败
            await page.route(`${API_URL}/control/dns/upstream_groups/**/test`, async (route) => {
                if (route.request().method() === 'POST') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify({
                            group_id: 'group-1',
                            results: [
                                { upstream: '8.8.8.8', success: true, rtt: 20 },
                                { upstream: '1.1.1.1', success: false, error: 'Connection timeout' },
                            ],
                        }),
                    });
                }
            });

            // 打开编辑对话框
            await page.click('button[title="编辑"]');
            
            // 点击检测按钮
            await page.click('button:has-text("检测")');
            
            // 等待测试完成
            await page.waitForSelector('text=测试完成');
            
            // 验证失败信息显示
            await expect(page.locator('text=Connection timeout')).toBeVisible();
        });
    });

    test.describe('启用/禁用分组', () => {
        test('应该能够切换分组启用状态', async ({ page }) => {
            // 模拟 API
            await page.route(`${API_URL}/control/dns/upstream_groups/**`, async (route) => {
                if (route.request().method() === 'PUT') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify({
                            id: 'group-1',
                            name: 'Test Group',
                            enabled: false,
                            is_default: false,
                            upstream_dns: ['8.8.8.8'],
                            created_at: '2024-01-01T00:00:00Z',
                            updated_at: new Date().toISOString(),
                        }),
                    });
                }
            });

            // 取消勾选启用复选框
            await page.uncheck('input[type="checkbox"][name="enabled"]');
            
            // 验证成功
            await expect(page.locator('.toast--success')).toBeVisible();
        });
    });

    test.describe('复制分组', () => {
        test('应该能够复制分组', async ({ page }) => {
            // 模拟 API
            await page.route(`${API_URL}/control/dns/upstream_groups`, async (route) => {
                if (route.request().method() === 'POST') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify({
                            id: 'new-copy-id',
                            name: 'Test Group Copy',
                            enabled: true,
                            is_default: false,
                            upstream_dns: ['8.8.8.8'],
                            created_at: new Date().toISOString(),
                            updated_at: new Date().toISOString(),
                        }),
                    });
                }
            });

            // 点击复制按钮
            await page.click('button[title="复制"]');
            
            // 验证成功
            await expect(page.locator('.toast--success')).toBeVisible();
        });
    });

    test.describe('空状态', () => {
        test('应该显示空状态提示', async ({ page }) => {
            // 模拟空列表
            await page.route(`${API_URL}/control/dns/upstream_groups`, async (route) => {
                if (route.request().method() === 'GET') {
                    await route.fulfill({
                        status: 200,
                        contentType: 'application/json',
                        body: JSON.stringify([]),
                    });
                }
            });

            await page.reload();
            await page.waitForLoadState('networkidle');
            
            // 验证空状态提示
            await expect(page.locator('text=暂无DNS上游分组')).toBeVisible();
            await expect(page.locator('text=点击下方按钮创建第一个分组')).toBeVisible();
        });
    });
});
