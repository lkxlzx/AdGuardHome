import { handleActions } from 'redux-actions';
import * as actions from '../actions/dnsRouting';

const dnsRouting = handleActions(
    {
        [actions.getDnsRoutingStatusRequest.toString()]: (state) => ({
            ...state,
            processing: true,
        }),
        [actions.getDnsRoutingStatusSuccess.toString()]: (state, { payload }: any) => ({
            ...state,
            processing: false,
            rules: payload.rules || [],
        }),
        [actions.getDnsRoutingStatusFailure.toString()]: (state) => ({
            ...state,
            processing: false,
        }),

        [actions.addDnsRoutingRuleRequest.toString()]: (state) => ({
            ...state,
            processingAdd: true,
        }),
        [actions.addDnsRoutingRuleSuccess.toString()]: (state) => ({
            ...state,
            processingAdd: false,
            isModalOpen: false,
        }),
        [actions.addDnsRoutingRuleFailure.toString()]: (state) => ({
            ...state,
            processingAdd: false,
        }),

        [actions.removeDnsRoutingRuleRequest.toString()]: (state) => ({
            ...state,
            processingRemove: true,
        }),
        [actions.removeDnsRoutingRuleSuccess.toString()]: (state) => ({
            ...state,
            processingRemove: false,
        }),
        [actions.removeDnsRoutingRuleFailure.toString()]: (state) => ({
            ...state,
            processingRemove: false,
        }),

        [actions.updateDnsRoutingRuleRequest.toString()]: (state) => ({
            ...state,
            processingUpdate: true,
        }),
        [actions.updateDnsRoutingRuleSuccess.toString()]: (state) => ({
            ...state,
            processingUpdate: false,
            isModalOpen: false,
        }),
        [actions.updateDnsRoutingRuleFailure.toString()]: (state) => ({
            ...state,
            processingUpdate: false,
        }),

        [actions.toggleDnsRoutingRuleRequest.toString()]: (state) => ({
            ...state,
            processingToggle: true,
        }),
        [actions.toggleDnsRoutingRuleSuccess.toString()]: (state) => ({
            ...state,
            processingToggle: false,
        }),
        [actions.toggleDnsRoutingRuleFailure.toString()]: (state) => ({
            ...state,
            processingToggle: false,
        }),

        [actions.refreshDnsRoutingRequest.toString()]: (state) => ({
            ...state,
            processingRefresh: true,
        }),
        [actions.refreshDnsRoutingSuccess.toString()]: (state) => ({
            ...state,
            processingRefresh: false,
        }),
        [actions.refreshDnsRoutingFailure.toString()]: (state) => ({
            ...state,
            processingRefresh: false,
        }),

        [actions.toggleDnsRoutingModal.toString()]: (state, { payload }: any) => {
            if (payload) {
                return {
                    ...state,
                    isModalOpen: !state.isModalOpen,
                    modalType: payload.type || null,
                    modalUrl: payload.url || null,
                };
            }

            return {
                ...state,
                isModalOpen: !state.isModalOpen,
                modalType: null,
                modalUrl: null,
            };
        },
    },
    {
        processing: false,
        processingAdd: false,
        processingRemove: false,
        processingUpdate: false,
        processingToggle: false,
        processingRefresh: false,
        rules: [],
        isModalOpen: false,
        modalType: null,
        modalUrl: null,
    },
);

export default dnsRouting;
