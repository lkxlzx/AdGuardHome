import React, { Component, Fragment } from 'react';
import { Trans, withTranslation } from 'react-i18next';

import Table from './UpstreamGroupsTable';
import Modal from './UpstreamGroupsModal';
import Card from '../../../ui/Card';

import { MODAL_TYPE } from '../../../../helpers/constants';
import { UpstreamGroup } from '../../../../initialState';

interface UpstreamGroupsProps {
    t: (...args: unknown[]) => string;
    groups: UpstreamGroup[];
    processing: boolean;
    processingAdd: boolean;
    processingDelete: boolean;
    processingUpdate: boolean;
    onAdd: (group: UpstreamGroup) => void;
    onUpdate: (target: UpstreamGroup, update: UpstreamGroup) => void;
    onDelete: (group: UpstreamGroup) => void;
    onSetDefault: (group: UpstreamGroup) => void;
    onTest: (group: UpstreamGroup) => void;
    upstreamGroupTests: any;
}

interface UpstreamGroupsState {
    isModalOpen: boolean;
    modalType: string;
    currentGroup?: UpstreamGroup;
}

class UpstreamGroups extends Component<UpstreamGroupsProps, UpstreamGroupsState> {
    state: UpstreamGroupsState = {
        isModalOpen: false,
        modalType: MODAL_TYPE.ADD,
        currentGroup: undefined,
    };

    toggleModal = (config?: { type?: string; currentGroup?: UpstreamGroup }) => {
        this.setState((prevState) => ({
            isModalOpen: !prevState.isModalOpen,
            modalType: config?.type || MODAL_TYPE.ADD,
            currentGroup: config?.currentGroup,
        }));
    };

    handleDelete = (group: UpstreamGroup) => {
        const { t, onDelete } = this.props;

        if (group.is_default) {
            // eslint-disable-next-line no-alert
            alert(t('upstream_group_cannot_delete_default'));
            return;
        }

        // eslint-disable-next-line no-alert
        if (window.confirm(t('upstream_group_confirm_delete', { name: group.name }))) {
            onDelete(group);
        }
    };

    handleSubmit = (values: any) => {
        const { modalType, currentGroup } = this.state;
        const { onAdd, onUpdate, groups } = this.props;

        if (modalType === MODAL_TYPE.EDIT && currentGroup) {
            const updatedGroup: UpstreamGroup = {
                ...currentGroup,
                name: values.name,
                upstreams: values.upstreams,
                enabled: values.enabled !== undefined ? values.enabled : true,
                is_default: values.is_default || false,
            };
            onUpdate(currentGroup, updatedGroup);
        } else {
            const isFirstGroup = groups.length === 0;
            const shouldBeDefault = values.is_default || isFirstGroup;
            
            const newGroup: UpstreamGroup = {
                id: `group_${Date.now()}`,
                name: values.name,
                upstreams: values.upstreams,
                enabled: values.enabled !== undefined ? values.enabled : true,
                is_default: shouldBeDefault,
            };
            
            onAdd(newGroup);
        }

        this.toggleModal();
    };

    handleSetDefault = (group: UpstreamGroup) => {
        this.props.onSetDefault(group);
    };

    handleToggleEnabled = (group: UpstreamGroup) => {
        const { onUpdate } = this.props;
        const updatedGroup: UpstreamGroup = {
            ...group,
            enabled: !group.enabled,
        };
        onUpdate(group, updatedGroup);
    };

    render() {
        const { t, groups, processing, processingAdd, processingDelete, processingUpdate, onTest, upstreamGroupTests } = this.props;
        const { isModalOpen, modalType, currentGroup } = this.state;

        return (
            <Fragment>
                <Card title={t('upstream_dns_groups')} bodyType="card-body box-body--settings">
                    <Fragment>
                        <div className="mb-3">
                            <p className="form__desc">{t('upstream_groups_desc')}</p>
                        </div>

                        <Table
                            list={groups}
                            processing={processing}
                            processingAdd={processingAdd}
                            processingDelete={processingDelete}
                            processingUpdate={processingUpdate}
                            handleDelete={this.handleDelete}
                            toggleModal={this.toggleModal}
                            toggleDefault={this.handleSetDefault}
                            toggleEnabled={this.handleToggleEnabled}
                            handleTest={onTest}
                            upstreamGroupTests={upstreamGroupTests}
                        />

                        <div className="card-actions">
                            <button
                                data-testid="add-upstream-group"
                                type="button"
                                className="btn btn-success btn-standard"
                                onClick={() => this.toggleModal({ type: MODAL_TYPE.ADD })}
                                disabled={processingAdd}>
                                <Trans>upstream_group_add</Trans>
                            </button>
                        </div>

                        <Modal
                            isOpen={isModalOpen}
                            modalType={modalType}
                            toggleModal={this.toggleModal}
                            handleSubmit={this.handleSubmit}
                            processing={processingAdd || processingUpdate}
                            currentGroup={currentGroup}
                        />
                    </Fragment>
                </Card>
            </Fragment>
        );
    }
}

export default withTranslation()(UpstreamGroups);
