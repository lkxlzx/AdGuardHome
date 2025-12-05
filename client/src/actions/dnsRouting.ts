import { createAction } from 'redux-actions';
import i18next from 'i18next';
import apiClient from '../api/Api';
import { addErrorToast, addSuccessToast } from './toasts';

// Helper function to get appropriate error message based on error type
const getErrorMessage = (error: any, defaultKey = 'error_generic') => {
    if (error.response?.status === 403) {
        return i18next.t('error_permission_denied');
    } else if (error.response?.status >= 500) {
        return i18next.t('error_server');
    } else if (!error.response) {
        return i18next.t('error_network');
    }
    return i18next.t(defaultKey);
};

// Modal actions
export const toggleDnsRoutingModal = createAction('DNS_ROUTING_MODAL_TOGGLE');

// Get DNS routing filters
export const getDnsRoutingFiltersRequest = createAction('GET_DNS_ROUTING_FILTERS_REQUEST');
export const getDnsRoutingFiltersFailure = createAction('GET_DNS_ROUTING_FILTERS_FAILURE');
export const getDnsRoutingFiltersSuccess = createAction('GET_DNS_ROUTING_FILTERS_SUCCESS');

export const getDnsRoutingFilters = () => async (dispatch: any, getState: any) => {
    const requestAction = getDnsRoutingFiltersRequest();
    dispatch(requestAction);
    
    // Get the request ID from the updated state
    const requestId = getState().dnsRouting.lastRequestId;
    
    try {
        const rules = await apiClient.getDnsRoutingRules();
        dispatch(getDnsRoutingFiltersSuccess({ data: rules || [], requestId }));
    } catch (error: any) {
        const errorMessage = getErrorMessage(error, 'error_loading_dns_routing_rules');
        dispatch(addErrorToast({ error, message: errorMessage }));
        dispatch(getDnsRoutingFiltersFailure({ requestId }));
    }
};

// Add DNS routing filter
export const addDnsRoutingFilterRequest = createAction('ADD_DNS_ROUTING_FILTER_REQUEST');
export const addDnsRoutingFilterFailure = createAction('ADD_DNS_ROUTING_FILTER_FAILURE');
export const addDnsRoutingFilterSuccess = createAction('ADD_DNS_ROUTING_FILTER_SUCCESS');

export const addDnsRoutingFilter = (
    url: string,
    name: string,
    upstreamGroup: string,
    updateInterval: number = 0,
    priority: number = 0
) => async (dispatch: any) => {
    dispatch(addDnsRoutingFilterRequest());
    try {
        await apiClient.addDnsRoutingRule({
            url,
            name,
            upstream_group: upstreamGroup,
            update_interval: updateInterval,
            priority,
        });
        dispatch(addDnsRoutingFilterSuccess());
        dispatch(toggleDnsRoutingModal());
        dispatch(addSuccessToast('dns_routing_rule_added'));
        dispatch(getDnsRoutingFilters());
    } catch (error: any) {
        const errorMessage = getErrorMessage(error, 'error_adding_dns_routing_rule');
        dispatch(addErrorToast({ error, message: errorMessage }));
        dispatch(addDnsRoutingFilterFailure());
    }
};

// Edit DNS routing filter
export const editDnsRoutingFilterRequest = createAction('EDIT_DNS_ROUTING_FILTER_REQUEST');
export const editDnsRoutingFilterFailure = createAction('EDIT_DNS_ROUTING_FILTER_FAILURE');
export const editDnsRoutingFilterSuccess = createAction('EDIT_DNS_ROUTING_FILTER_SUCCESS');

export const editDnsRoutingFilter = (
    id: number,
    data: {
        name: string;
        url: string;
        upstreamGroup: string;
        updateInterval: number;
        priority: number;
        enabled?: boolean;
    }
) => async (dispatch: any) => {
    dispatch(editDnsRoutingFilterRequest());
    try {
        await apiClient.updateDnsRoutingRule({
            id,
            name: data.name,
            url: data.url,
            upstream_group: data.upstreamGroup,
            update_interval: data.updateInterval,
            priority: data.priority,
            enabled: data.enabled !== undefined ? data.enabled : true,
        });
        dispatch(editDnsRoutingFilterSuccess());
        dispatch(toggleDnsRoutingModal());
        dispatch(addSuccessToast('dns_routing_rule_updated'));
        dispatch(getDnsRoutingFilters());
    } catch (error: any) {
        const errorMessage = getErrorMessage(error, 'error_updating_dns_routing_rule');
        dispatch(addErrorToast({ error, message: errorMessage }));
        dispatch(editDnsRoutingFilterFailure());
    }
};

// Remove DNS routing filter
export const removeDnsRoutingFilterRequest = createAction('REMOVE_DNS_ROUTING_FILTER_REQUEST');
export const removeDnsRoutingFilterFailure = createAction('REMOVE_DNS_ROUTING_FILTER_FAILURE');
export const removeDnsRoutingFilterSuccess = createAction('REMOVE_DNS_ROUTING_FILTER_SUCCESS');

export const removeDnsRoutingFilter = (id: number) => async (dispatch: any) => {
    dispatch(removeDnsRoutingFilterRequest());
    try {
        await apiClient.deleteDnsRoutingRule({ id });
        dispatch(removeDnsRoutingFilterSuccess());
        dispatch(addSuccessToast('dns_routing_rule_removed'));
        dispatch(getDnsRoutingFilters());
    } catch (error: any) {
        const errorMessage = getErrorMessage(error, 'error_deleting_dns_routing_rule');
        dispatch(addErrorToast({ error, message: errorMessage }));
        dispatch(removeDnsRoutingFilterFailure());
    }
};

// Toggle DNS routing filter status
export const toggleDnsRoutingFilterRequest = createAction('TOGGLE_DNS_ROUTING_FILTER_REQUEST');
export const toggleDnsRoutingFilterFailure = createAction('TOGGLE_DNS_ROUTING_FILTER_FAILURE');
export const toggleDnsRoutingFilterSuccess = createAction('TOGGLE_DNS_ROUTING_FILTER_SUCCESS');

export const toggleDnsRoutingFilter = (filter: any) => async (dispatch: any) => {
    dispatch(toggleDnsRoutingFilterRequest());
    try {
        await apiClient.updateDnsRoutingRule({
            id: filter.id,
            name: filter.name,
            url: filter.url,
            upstream_group: filter.upstream_group,
            update_interval: filter.update_interval || 0,
            priority: filter.priority || 0,
            enabled: !filter.enabled,
        });
        dispatch(toggleDnsRoutingFilterSuccess());
        dispatch(getDnsRoutingFilters());
    } catch (error: any) {
        const errorMessage = getErrorMessage(error, 'error_toggling_dns_routing_rule');
        dispatch(addErrorToast({ error, message: errorMessage }));
        dispatch(toggleDnsRoutingFilterFailure());
    }
};

// Refresh DNS routing filters
export const refreshDnsRoutingFiltersRequest = createAction('REFRESH_DNS_ROUTING_FILTERS_REQUEST');
export const refreshDnsRoutingFiltersFailure = createAction('REFRESH_DNS_ROUTING_FILTERS_FAILURE');
export const refreshDnsRoutingFiltersSuccess = createAction('REFRESH_DNS_ROUTING_FILTERS_SUCCESS');

export const refreshDnsRoutingFilters = (id?: number) => async (dispatch: any) => {
    dispatch(refreshDnsRoutingFiltersRequest());
    try {
        await apiClient.refreshDnsRoutingRule({ id: id || 0 });
        dispatch(refreshDnsRoutingFiltersSuccess());
        dispatch(addSuccessToast('dns_routing_refreshed'));
        dispatch(getDnsRoutingFilters());
    } catch (error: any) {
        const errorMessage = getErrorMessage(error, 'error_refreshing_dns_routing_rule');
        dispatch(addErrorToast({ error, message: errorMessage }));
        dispatch(refreshDnsRoutingFiltersFailure());
    }
};
