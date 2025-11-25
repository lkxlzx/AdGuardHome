import React, { Component } from 'react';

import ReactModal from 'react-modal';
import { withTranslation } from 'react-i18next';

import { MODAL_TYPE } from '../../../../helpers/constants';
import { UpstreamGroupsForm } from './UpstreamGroupsForm';
import '../../../ui/Modal.css';

ReactModal.setAppElement('#root');

interface ModalProps {
    toggleModal: () => void;
    isOpen: boolean;
    handleSubmit: (values: any) => void;
    modalType: string;
    currentGroup?: any;
    t: (...args: unknown[]) => string;
    processing: boolean;
}

class UpstreamGroupsModal extends Component<ModalProps> {
    closeModal = () => {
        this.props.toggleModal();
    };

    render() {
        const { isOpen, handleSubmit, modalType, currentGroup, t, processing } = this.props;

        const isEdit = modalType === MODAL_TYPE.EDIT;
        const title = isEdit ? t('upstream_group_edit') : t('upstream_group_add');

        return (
            <ReactModal
                className="Modal__Bootstrap modal-dialog modal-dialog-centered"
                closeTimeoutMS={0}
                isOpen={isOpen}
                onRequestClose={this.closeModal}>
                <div className="modal-content">
                    <div className="modal-header">
                        <h4 className="modal-title">{title}</h4>

                        <button type="button" className="close" onClick={this.closeModal}>
                            <span className="sr-only">Close</span>
                        </button>
                    </div>

                    <UpstreamGroupsForm
                        initialValues={currentGroup}
                        onSubmit={handleSubmit}
                        processing={processing}
                        closeModal={this.closeModal}
                        isEdit={isEdit}
                    />
                </div>
            </ReactModal>
        );
    }
}

export default withTranslation()(UpstreamGroupsModal);
