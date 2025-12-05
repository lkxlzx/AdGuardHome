import { handleActions } from 'redux-actions';
import { MODAL_TYPE } from '../helpers/constants';

const initialState = {
    isModalOpen: false,
    modalType: MODAL_TYPE.ADD_FILTERS,
    modalFilterUrl: '',
    modalFilter: null,
    processingFilters: false,
    processingAddFilter: false,
    processingEditFilter: false,
    processingRemoveFilter: false,
    processingToggleFilter: false,
    processingRefreshFilters: false,
    filters: [],
    // Request tracking for concurrent request management
    lastRequestId: 0,
    lastSuccessfulRequestId: 0,
};

const dnsRouting = handleActions(
    {
        // Modal
        DNS_ROUTING_MODAL_TOGGLE: (state, { payload }: any) => {
            if (payload && typeof payload === 'object') {
                const { type, url, filter } = payload;
                return {
                    ...state,
                    isModalOpen: !state.isModalOpen,
                    modalType: type || MODAL_TYPE.ADD_FILTERS,
                    modalFilterUrl: url || '',
                    modalFilter: filter || null,
                };
            }
            return {
                ...state,
                isModalOpen: !state.isModalOpen,
                modalType: MODAL_TYPE.ADD_FILTERS,
                modalFilterUrl: '',
                modalFilter: null,
            };
        },

        // Get filters
        GET_DNS_ROUTING_FILTERS_REQUEST: (state) => {
            const requestId = state.lastRequestId + 1;
            return {
                ...state,
                processingFilters: true,
                lastRequestId: requestId,
            };
        },
        GET_DNS_ROUTING_FILTERS_FAILURE: (state, { payload }: any) => {
            // Only update if this is the latest request
            const requestId = payload?.requestId || 0;
            if (requestId < state.lastSuccessfulRequestId) {
                return state; // Ignore outdated request
            }
            return {
                ...state,
                processingFilters: false,
            };
        },
        GET_DNS_ROUTING_FILTERS_SUCCESS: (state, { payload }: any) => {
            // Only update if this is the latest request
            const requestId = payload?.requestId || state.lastRequestId;
            if (requestId < state.lastSuccessfulRequestId) {
                return state; // Ignore outdated response
            }
            return {
                ...state,
                processingFilters: false,
                filters: payload?.data || payload || [],
                lastSuccessfulRequestId: requestId,
            };
        },

        // Add filter
        ADD_DNS_ROUTING_FILTER_REQUEST: (state) => ({
            ...state,
            processingAddFilter: true,
        }),
        ADD_DNS_ROUTING_FILTER_FAILURE: (state) => ({
            ...state,
            processingAddFilter: false,
        }),
        ADD_DNS_ROUTING_FILTER_SUCCESS: (state) => ({
            ...state,
            processingAddFilter: false,
        }),

        // Edit filter
        EDIT_DNS_ROUTING_FILTER_REQUEST: (state) => ({
            ...state,
            processingEditFilter: true,
        }),
        EDIT_DNS_ROUTING_FILTER_FAILURE: (state) => ({
            ...state,
            processingEditFilter: false,
        }),
        EDIT_DNS_ROUTING_FILTER_SUCCESS: (state) => ({
            ...state,
            processingEditFilter: false,
        }),

        // Remove filter
        REMOVE_DNS_ROUTING_FILTER_REQUEST: (state) => ({
            ...state,
            processingRemoveFilter: true,
        }),
        REMOVE_DNS_ROUTING_FILTER_FAILURE: (state) => ({
            ...state,
            processingRemoveFilter: false,
        }),
        REMOVE_DNS_ROUTING_FILTER_SUCCESS: (state) => ({
            ...state,
            processingRemoveFilter: false,
        }),

        // Toggle filter
        TOGGLE_DNS_ROUTING_FILTER_REQUEST: (state) => ({
            ...state,
            processingToggleFilter: true,
        }),
        TOGGLE_DNS_ROUTING_FILTER_FAILURE: (state) => ({
            ...state,
            processingToggleFilter: false,
        }),
        TOGGLE_DNS_ROUTING_FILTER_SUCCESS: (state) => ({
            ...state,
            processingToggleFilter: false,
        }),

        // Refresh filters
        REFRESH_DNS_ROUTING_FILTERS_REQUEST: (state) => ({
            ...state,
            processingRefreshFilters: true,
        }),
        REFRESH_DNS_ROUTING_FILTERS_FAILURE: (state) => ({
            ...state,
            processingRefreshFilters: false,
        }),
        REFRESH_DNS_ROUTING_FILTERS_SUCCESS: (state) => ({
            ...state,
            processingRefreshFilters: false,
        }),
    },
    initialState
);

export default dnsRouting;
