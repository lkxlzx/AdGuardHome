@echo off
echo ========================================
echo DNS路由独立性测试脚本
echo ========================================
echo.

echo 测试配置：
echo - protection_enabled: false
echo - filtering_enabled: false
echo - DNS路由过滤器: enabled
echo.

echo 测试1: 查询中国域名 (应该路由到china上游组)
echo ----------------------------------------
nslookup baidu.com 127.0.0.1
echo.

echo 测试2: 查询自定义域名 (应该路由到custom上游组)
echo ----------------------------------------
nslookup example.com 127.0.0.1
echo.

echo 测试3: 查询普通域名 (应该使用默认上游)
echo ----------------------------------------
nslookup google.com 127.0.0.1
echo.

echo ========================================
echo 测试完成！
echo.
echo 请检查AdGuardHome日志，应该看到：
echo [debug] DNS routing matched host=baidu.com upstream_group=china
echo [debug] returning DNS routing rule upstream_group=china
echo.
echo 如果看到这些日志，说明DNS路由引擎独立工作成功！
echo ========================================
pause
