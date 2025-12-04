import React from 'react';
import ReactModal from 'react-modal';
import { useTranslation, Trans } from 'react-i18next';
import { DnsRoutingForm } from './DnsRoutingForm';
import { MODAL_TYPE } from '../../helpers/constants';

type FormValues = {
    name: string;
    url: string;
    upstreamGroup: string;
    updateInterval: number;
    priority: number;
};

type Props = {
    isOpen: boolean;
    closeModal: () => void;
    onSubmit: (values: FormValues) => void;
    processingAddFilter: boolean;
    processingConfigFilter: boolean;
    modalType: string;
    currentFilterData?: any;
};

export const DnsRoutingModal = ({
    isOpen,
    closeModal,
    onSubmit,
    processingAddFilter,
    processingConfigFilter,
    modalType,
    currentFilterData,
}: Props) => {
    const { t } = useTranslation();

    const isEdit = modalType === MODAL_TYPE.EDIT_FILTERS;
    const title = isEdit ? t('edit_routing_rule') : t('new_routing_rule');

    const initialValues = isEdit && currentFilterData ? {
        name: currentFilterData.name || '',
        url: currentFilterData.url || '',
        upstreamGroup: currentFilterData.upstream_group || '',
        updateInterval: currentFilterData.update_interval ?? 0,
        priority: currentFilterData.priority ?? 0,
    } : undefined;

    return (
        <ReactModal
            className="Modal__Bootstrap modal-dialog modal-dialog-centered"
            closeTimeoutMS={0}
            isOpen={isOpen}
            onRequestClose={closeModal}>
            <div className="modal-content">
                <div className="modal-header">
                    <h4 className="modal-title">{title}</h4>
                    <button
                        type="button"
                        className="close"
                        onClick={closeModal}>
                        <span className="sr-only">Close</span>
                    </button>
                </div>

                <DnsRoutingForm
                    closeModal={closeModal}
                    onSubmit={onSubmit}
                    processingAddFilter={processingAddFilter}
                    processingConfigFilter={processingConfigFilter}
                    initialValues={initialValues}
                />
            </div>
        </ReactModal>
    );
};
