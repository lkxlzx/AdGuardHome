import React, { Component, Fragment } from 'react';
import { withTranslation } from 'react-i18next';
import { connect } from 'react-redux';
import { shallowEqual } from 'react-redux';

import Card from '../../../ui/Card';
import Form from './Form';
import Table from './UpstreamGroupsTable';
import Modal from './UpstreamGroupsModal';
import { setDnsConfig, testUpstreamGroup } from '../../../../actions/dnsConfig';
import { RootState, UpstreamGroup } from '../../../../initialState';
import { MODAL_TYPE } from '../../../../helpers/constants';

interface UpstreamDnsSettingsProps {
    t: (...args: unknown[]) => string;
    upstream_dns: string;
    fallback_dns: string;
    bootstrap_dns: string;
    upstream_mode: string;
    resolve_clients: boolean;
    local_ptr_upstreams: string;
    use_private_ptr_resolvers: boolean;
    upstream_timeout: number;
    upstream_groups: UpstreamGroup[];
    processingSetConfig: boolean;
    setDnsConfig: (config: any) => void;
    testUpstreamGroup: (groupId: string, upstreams: string[]) => void;
    dnsConfig: any;
    upstreamGroupTests: any;
}

interface UpstreamDnsSettingsState {
    isModalOpen: boolean;
    modalType: string;
    currentGroup?: UpstreamGroup;
}

class UpstreamDnsSettings extends Component<UpstreamDnsSettingsProps, UpstreamDnsSettingsState> {
    state: UpstreamDnsSettingsState = {
        isModalOpen: false,
        modalType: MODAL_TYPE.ADD,
        currentGroup: undefined,
    };

    handleFormSubmit = (values: any) => {
        this.props.setDnsConfig(values);
    };

    toggleModal = (config?: { type?: string; currentGroup?: UpstreamGroup }) => {
        this.setState((prevState) => ({
            isModalOpen: !prevState.isModalOpen,
            modalType: config?.type || MODAL_TYPE.ADD,
            currentGroup: config?.currentGroup,
        }));
    };

    handleAdd = (group: UpstreamGroup) => {
        const { upstream_groups } = this.props;
        
        // 如果新分组设置为默认组，将其他分组的 is_default 设为 false
        let newGroups: UpstreamGroup[];
        if (group.is_default) {
            newGroups = [
                ...upstream_groups.map((g) => ({ ...g, is_default: false })),
                group,
            ];
        } else {
            newGroups = [...upstream_groups, group];
        }

        this.props.setDnsConfig({
            upstream_groups: newGroups,
        });
    };

    handleUpdate = (target: UpstreamGroup, update: UpstreamGroup) => {
        const { upstream_groups } = this.props;
        
        // 如果更新后的分组设置为默认组，将其他分组的 is_default 设为 false
        let newGroups: UpstreamGroup[];
        if (update.is_default && !target.is_default) {
            newGroups = upstream_groups.map((g) => ({
                ...g,
                is_default: g.id === target.id,
                ...(g.id === target.id ? { name: update.name, upstreams: update.upstreams, enabled: update.enabled } : {}),
            }));
        } else {
            newGroups = upstream_groups.map((g) => (g.id === target.id ? update : g));
            
            // 如果取消了默认组，确保至少有一个默认组
            if (target.is_default && !update.is_default) {
                const hasDefault = newGroups.some((g) => g.is_default);
                if (!hasDefault && newGroups.length > 0) {
                    newGroups[0] = { ...newGroups[0], is_default: true };
                }
            }
        }

        this.props.setDnsConfig({
            upstream_groups: newGroups,
        });
    };

    handleDelete = (group: UpstreamGroup) => {
        const { t, upstream_groups } = this.props;

        if (group.is_default) {
            // eslint-disable-next-line no-alert
            alert(t('upstream_group_cannot_delete_default'));
            return;
        }

        // eslint-disable-next-line no-alert
        if (window.confirm(t('upstream_group_confirm_delete', { name: group.name }))) {
            const remainingGroups = upstream_groups.filter((g) => g.id !== group.id);

            const hasDefault = remainingGroups.some((g) => g.is_default);
            if (!hasDefault && remainingGroups.length > 0) {
                remainingGroups[0].is_default = true;
            }

            this.props.setDnsConfig({
                upstream_groups: remainingGroups,
            });
        }
    };

    handleSetDefault = (group: UpstreamGroup) => {
        const { upstream_groups } = this.props;
        const newGroups = upstream_groups.map((g) => ({
            ...g,
            is_default: g.id === group.id,
        }));

        // 直接保存，确保数据被发送到后端
        this.props.setDnsConfig({
            upstream_groups: newGroups,
        });
    };

    handleTest = (group: UpstreamGroup) => {
        // Convert upstreams from string to array if needed
        const upstreamsArray = Array.isArray(group.upstreams)
            ? group.upstreams
            : group.upstreams.split('\n').filter((s: string) => s.trim());
        
        this.props.testUpstreamGroup(group.id, upstreamsArray);
    };

    handleToggleEnabled = (group: UpstreamGroup) => {
        const updatedGroup: UpstreamGroup = {
            ...group,
            enabled: !group.enabled,
        };
        this.handleUpdate(group, updatedGroup);
    };

    handleSubmit = (values: any) => {
        const { modalType, currentGroup } = this.state;
        const { upstream_groups } = this.props;

        if (modalType === MODAL_TYPE.EDIT && currentGroup) {
            const updatedGroup: UpstreamGroup = {
                ...currentGroup,
                name: values.name,
                upstreams: values.upstreams,
                enabled: values.enabled !== undefined ? values.enabled : true,
                is_default: values.is_default || false,
            };
            this.handleUpdate(currentGroup, updatedGroup);
        } else {
            const isFirstGroup = upstream_groups.length === 0;
            const shouldBeDefault = values.is_default || isFirstGroup;

            const newGroup: UpstreamGroup = {
                id: `group_${Date.now()}`,
                name: values.name,
                upstreams: values.upstreams,
                enabled: values.enabled !== undefined ? values.enabled : true,
                is_default: shouldBeDefault,
            };

            this.handleAdd(newGroup);
        }

        this.toggleModal();
    };

    render() {
        const {
            t,
            upstream_dns,
            fallback_dns,
            bootstrap_dns,
            upstream_mode,
            resolve_clients,
            local_ptr_upstreams,
            use_private_ptr_resolvers,
            upstream_timeout,
            upstream_groups,
            processingSetConfig,
        } = this.props;
        const { isModalOpen, modalType, currentGroup } = this.state;

        // 准备分组管理内容
        const upstreamGroupsContent = (
            <>
                <Table
                    list={upstream_groups || []}
                    processing={processingSetConfig}
                    processingAdd={processingSetConfig}
                    processingDelete={processingSetConfig}
                    processingUpdate={processingSetConfig}
                    handleDelete={this.handleDelete}
                    toggleModal={this.toggleModal}
                    toggleDefault={this.handleSetDefault}
                    toggleEnabled={this.handleToggleEnabled}
                    handleTest={this.handleTest}
                    upstreamGroupTests={this.props.upstreamGroupTests || {}}
                />

                <div className="mt-3">
                    <button
                        data-testid="add-upstream-group"
                        type="button"
                        className="btn btn-success btn-standard"
                        onClick={() => this.toggleModal({ type: MODAL_TYPE.ADD })}
                        disabled={processingSetConfig}>
                        {t('upstream_group_add')}
                    </button>
                </div>
            </>
        );

        return (
            <Card title={t('upstream_dns')} bodyType="card-body box-body--settings">
                <div className="form">
                    <Form
                        initialValues={{
                            upstream_dns,
                            fallback_dns,
                            bootstrap_dns,
                            upstream_mode,
                            resolve_clients,
                            local_ptr_upstreams,
                            use_private_ptr_resolvers,
                            upstream_timeout,
                        }}
                        onSubmit={this.handleFormSubmit}
                        upstreamGroupsContent={upstreamGroupsContent}
                    />

                    <Modal
                        isOpen={isModalOpen}
                        modalType={modalType}
                        toggleModal={this.toggleModal}
                        handleSubmit={this.handleSubmit}
                        processing={processingSetConfig}
                        currentGroup={currentGroup}
                    />
                </div>
            </Card>
        );
    }
}

const mapStateToProps = (state: RootState) => ({
    upstream_dns: state.dnsConfig.upstream_dns,
    fallback_dns: state.dnsConfig.fallback_dns,
    bootstrap_dns: state.dnsConfig.bootstrap_dns,
    upstream_mode: state.dnsConfig.upstream_mode,
    resolve_clients: state.dnsConfig.resolve_clients,
    local_ptr_upstreams: state.dnsConfig.local_ptr_upstreams,
    use_private_ptr_resolvers: state.dnsConfig.use_private_ptr_resolvers,
    upstream_timeout: state.dnsConfig.upstream_timeout,
    upstream_groups: state.dnsConfig.upstream_groups || [],
    processingSetConfig: state.dnsConfig.processingSetConfig,
    dnsConfig: state.dnsConfig,
    upstreamGroupTests: state.dnsConfig.upstreamGroupTests || {},
});

const mapDispatchToProps = {
    setDnsConfig,
    testUpstreamGroup,
};

export default connect(mapStateToProps, mapDispatchToProps)(withTranslation()(UpstreamDnsSettings));
