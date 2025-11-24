# 需要添加到翻译文件的内容

## 中文翻译 (client/src/__locales/zh-cn.json)

在文件末尾添加以下内容（注意在最后一个现有条目后添加逗号）：

```json
    "upstream_groups_title": "上游 DNS 服务器分组",
    "upstream_groups_desc": "创建上游 DNS 服务器分组，可在 DNS 分流规则中使用这些分组。未命中分流规则的查询将使用默认组",
    "upstream_groups_empty": "暂无分组，点击下方按钮添加第一个分组（将自动设为默认组）",
    "upstream_group_name": "分组名称",
    "upstream_group_name_placeholder": "例如：国内 DNS、国外 DNS",
    "upstream_group_servers": "DNS 服务器列表",
    "upstream_group_servers_placeholder": "每行一个 DNS 服务器地址",
    "upstream_group_add": "添加分组",
    "upstream_group_confirm_delete": "确定要删除此分组吗？",
    "upstream_group_cannot_delete_default": "无法删除默认组，请先设置其他组为默认组",
    "default_group": "默认组",
    "default_group_hint": "分流规则未命中时使用此组",
    "set_as_default": "设为默认",
    "dns_routing": "DNS 分流",
    "dns_routing_desc": "为特定域名指定专用的上游 DNS 服务器组",
    "dns_routing_hint": "根据域名规则将 DNS 查询路由到不同的上游服务器组",
    "dns_routing_no_rules": "暂无分流规则",
    "dns_routing_rule_added": "分流规则已添加",
    "dns_routing_rule_deleted": "分流规则已删除",
    "dns_routing_rule_updated": "分流规则已更新",
    "dns_routing_confirm_delete": "确定要删除此分流规则吗？",
    "dns_routing_domain": "域名",
    "dns_routing_domain_placeholder": "例如：example.com 或 *.example.com",
    "dns_routing_upstream_group": "上游服务器组",
    "dns_routing_description": "描述",
    "dns_routing_description_placeholder": "可选的规则描述",
    "dns_routing_enabled": "启用",
    "dns_routing_add_rule": "添加规则",
    "dns_routing_edit_rule": "编辑规则",
    "save": "保存",
    "cancel": "取消",
    "edit": "编辑",
    "delete": "删除",
    "add": "添加",
    "enabled": "已启用",
    "disabled": "已禁用",
    "actions": "操作"
```

## 英文翻译 (client/src/__locales/en.json)

```json
    "upstream_groups_title": "Upstream DNS Server Groups",
    "upstream_groups_desc": "Create upstream DNS server groups that can be used in DNS routing rules. Queries that don't match routing rules will use the default group",
    "upstream_groups_empty": "No groups yet. Click the button below to add your first group (it will be set as default automatically)",
    "upstream_group_name": "Group Name",
    "upstream_group_name_placeholder": "e.g., Domestic DNS, Foreign DNS",
    "upstream_group_servers": "DNS Server List",
    "upstream_group_servers_placeholder": "One DNS server address per line",
    "upstream_group_add": "Add Group",
    "upstream_group_confirm_delete": "Are you sure you want to delete this group?",
    "upstream_group_cannot_delete_default": "Cannot delete the default group. Please set another group as default first",
    "default_group": "Default",
    "default_group_hint": "Used when routing rules don't match",
    "set_as_default": "Set as Default",
    "dns_routing": "DNS Routing",
    "dns_routing_desc": "Specify dedicated upstream DNS server groups for specific domains",
    "dns_routing_hint": "Route DNS queries to different upstream server groups based on domain rules",
    "dns_routing_no_rules": "No routing rules yet",
    "dns_routing_rule_added": "Routing rule added",
    "dns_routing_rule_deleted": "Routing rule deleted",
    "dns_routing_rule_updated": "Routing rule updated",
    "dns_routing_confirm_delete": "Are you sure you want to delete this routing rule?",
    "dns_routing_domain": "Domain",
    "dns_routing_domain_placeholder": "e.g., example.com or *.example.com",
    "dns_routing_upstream_group": "Upstream Server Group",
    "dns_routing_description": "Description",
    "dns_routing_description_placeholder": "Optional rule description",
    "dns_routing_enabled": "Enabled",
    "dns_routing_add_rule": "Add Rule",
    "dns_routing_edit_rule": "Edit Rule",
    "save": "Save",
    "cancel": "Cancel",
    "edit": "Edit",
    "delete": "Delete",
    "add": "Add",
    "enabled": "Enabled",
    "disabled": "Disabled",
    "actions": "Actions"
```

## 注意事项

1. 这些翻译需要手动添加到对应的 JSON 文件中
2. 确保 JSON 格式正确，注意逗号的使用
3. 如果某些键（如 save、cancel、edit 等）已存在，则无需重复添加
