import { handleActions } from 'redux-actions';
import { MODAL_TYPE } from '../helpers/constants';

const initialState = {
    isModalOpen: false,
    modalType: MODAL_TYPE.ADD_FILTERS,
    modalFilterUrl: '',
    processingFilters: false,
    processingAddFilter: false,
    processingEditFilter: false,
    processingRemoveFilter: false,
    processingToggleFilter: false,
    processingRefreshFilters: false,
    filters: [],
};

const dnsRouting = handleActions(
    {
        // Modal
        DNS_ROUTING_MODAL_TOGGLE: (state, { payload }: any) => {
            if (payload && typeof payload === 'object') {
                const { type, url } = payload;
                return {
                    ...state,
                    isModalOpen: !state.isModalOpen,
                    modalType: type || MODAL_TYPE.ADD_FILTERS,
                    modalFilterUrl: url || '',
                };
            }
            return {
                ...state,
                isModalOpen: !state.isModalOpen,
                modalType: MODAL_TYPE.ADD_FILTERS,
                modalFilterUrl: '',
            };
        },

        // Get filters
        GET_DNS_ROUTING_FILTERS_REQUEST: (state) => ({
            ...state,
            processingFilters: true,
        }),
        GET_DNS_ROUTING_FILTERS_FAILURE: (state) => ({
            ...state,
            processingFilters: false,
        }),
        GET_DNS_ROUTING_FILTERS_SUCCESS: (state, { payload }: any) => ({
            ...state,
            processingFilters: false,
            filters: payload || [],
        }),

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
