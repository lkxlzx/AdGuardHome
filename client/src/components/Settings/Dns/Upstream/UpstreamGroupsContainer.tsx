import React, { Component } from 'react';
import { connect } from 'react-redux';
import { shallowEqual } from 'react-redux';

import UpstreamGroups from './index';
import { setDnsConfig } from '../../../../actions/dnsConfig';
import { RootState, UpstreamGroup } from '../../../../initialState';

interface ContainerProps {
    upstream_groups: UpstreamGroup[];
    processingSetConfig: boolean;
    setDnsConfig: (config: any) => void;
    dnsConfig: any;
}

class UpstreamGroupsContainer extends Component<ContainerProps> {
    handleAdd = (group: UpstreamGroup) => {
        const { upstream_groups } = this.props;
        
        let newGroups: UpstreamGroup[];
        
        // 如果新分组设置为默认组，需要将其他分组的 is_default 设置为 false
        if (group.is_default) {
            newGroups = [
                ...upstream_groups.map((g) => ({ ...g, is_default: false })),
                group,
            ];
        } else {
            newGroups = [...upstream_groups, group];
        }
        
        // 只发送 upstream_groups 字段
        this.props.setDnsConfig({
            upstream_groups: newGroups,
        });
    };

    handleUpdate = (target: UpstreamGroup, update: UpstreamGroup) => {
        const { upstream_groups, dnsConfig } = this.props;
        
        console.log('handleUpdate - target:', target);
        console.log('handleUpdate - update:', update);
        console.log('handleUpdate - existing groups:', upstream_groups);
        
        // 如果更新后的分组设置为默认组，使用与 handleSetDefault 相同的逻辑
        let newGroups: UpstreamGroup[];
        if (update.is_default && !target.is_default) {
            // 设置为默认组：将所有分组的 is_default 设为 false，只有目标分组为 true
            newGroups = upstream_groups.map((g) => ({
                ...g,
                is_default: g.id === target.id,
                // 更新目标分组的其他字段
                ...(g.id === target.id ? { name: update.name, upstreams: update.upstreams, enabled: update.enabled } : {}),
            }));
        } else {
            // 普通更新：只更新目标分组
            newGroups = upstream_groups.map((g) => (g.id === target.id ? update : g));
            
            // 如果取消了默认组，确保至少有一个默认组
            if (target.is_default && !update.is_default) {
                const hasDefault = newGroups.some((g) => g.is_default);
                if (!hasDefault && newGroups.length > 0) {
                    newGroups[0] = { ...newGroups[0], is_default: true };
                }
            }
        }
        
        console.log('handleUpdate - newGroups:', newGroups);
        
        // 只发送 upstream_groups 字段，避免其他字段干扰
        const configToSave = {
            upstream_groups: newGroups,
        };
        console.log('handleUpdate - configToSave:', configToSave);
        console.log('handleUpdate - configToSave.upstream_groups:', configToSave.upstream_groups);
        
        this.props.setDnsConfig(configToSave);
    };

    handleDelete = (group: UpstreamGroup) => {
        const { upstream_groups, dnsConfig } = this.props;
        const remainingGroups = upstream_groups.filter((g) => g.id !== group.id);

        // 如果删除后没有默认组，设置第一个为默认组
        const hasDefault = remainingGroups.some((g) => g.is_default);
        if (!hasDefault && remainingGroups.length > 0) {
            remainingGroups[0].is_default = true;
        }

        // 只发送 upstream_groups 字段
        this.props.setDnsConfig({
            upstream_groups: remainingGroups,
        });
    };

    handleSetDefault = (group: UpstreamGroup) => {
        const { upstream_groups } = this.props;
        const newGroups = upstream_groups.map((g) => ({
            ...g,
            is_default: g.id === group.id,
        }));

        // 只发送 upstream_groups 字段
        this.props.setDnsConfig({
            upstream_groups: newGroups,
        });
    };

    render() {
        const { upstream_groups, processingSetConfig } = this.props;

        return (
            <UpstreamGroups
                groups={upstream_groups || []}
                processing={processingSetConfig}
                processingAdd={processingSetConfig}
                processingDelete={processingSetConfig}
                processingUpdate={processingSetConfig}
                onAdd={this.handleAdd}
                onUpdate={this.handleUpdate}
                onDelete={this.handleDelete}
                onSetDefault={this.handleSetDefault}
            />
        );
    }
}

const mapStateToProps = (state: RootState) => ({
    upstream_groups: state.dnsConfig.upstream_groups || [],
    processingSetConfig: state.dnsConfig.processingSetConfig,
    dnsConfig: state.dnsConfig,
});

const mapDispatchToProps = {
    setDnsConfig,
};

export default connect(mapStateToProps, mapDispatchToProps)(UpstreamGroupsContainer);
