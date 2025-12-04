import { createAction } from 'redux-actions';
import i18next from 'i18next';
import apiClient from '../api/Api';
import { addErrorToast, addSuccessToast } from './toasts';

// Modal actions
export const toggleDnsRoutingModal = createAction('DNS_ROUTING_MODAL_TOGGLE');

// Get DNS routing filters
export const getDnsRoutingFiltersRequest = createAction('GET_DNS_ROUTING_FILTERS_REQUEST');
export const getDnsRoutingFiltersFailure = createAction('GET_DNS_ROUTING_FILTERS_FAILURE');
export const getDnsRoutingFiltersSuccess = createAction('GET_DNS_ROUTING_FILTERS_SUCCESS');

export const getDnsRoutingFilters = () => async (dispatch: any) => {
    dispatch(getDnsRoutingFiltersRequest());
    try {
        const status = await apiClient.getFilteringStatus();
        dispatch(getDnsRoutingFiltersSuccess(status.dns_routing_filters || []));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(getDnsRoutingFiltersFailure());
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
        await apiClient.addFilter({
            url,
            name,
            whitelist: false,
            dns_routing: true,
            upstream_group: upstreamGroup,
            update_interval: updateInterval,
            priority,
        });
        dispatch(addDnsRoutingFilterSuccess());
        dispatch(toggleDnsRoutingModal());
        dispatch(addSuccessToast('dns_routing_rule_added'));
        dispatch(getDnsRoutingFilters());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(addDnsRoutingFilterFailure());
    }
};

// Edit DNS routing filter
export const editDnsRoutingFilterRequest = createAction('EDIT_DNS_ROUTING_FILTER_REQUEST');
export const editDnsRoutingFilterFailure = createAction('EDIT_DNS_ROUTING_FILTER_FAILURE');
export const editDnsRoutingFilterSuccess = createAction('EDIT_DNS_ROUTING_FILTER_SUCCESS');

export const editDnsRoutingFilter = (
    url: string,
    data: {
        name: string;
        upstreamGroup: string;
        updateInterval: number;
        priority: number;
    }
) => async (dispatch: any) => {
    dispatch(editDnsRoutingFilterRequest());
    try {
        await apiClient.setFilterUrl({
            url,
            data: {
                name: data.name,
                url,
                whitelist: false,
                dns_routing: true,
                upstream_group: data.upstreamGroup,
                update_interval: data.updateInterval,
                priority: data.priority,
            },
        });
        dispatch(editDnsRoutingFilterSuccess());
        dispatch(toggleDnsRoutingModal());
        dispatch(addSuccessToast('dns_routing_rule_updated'));
        dispatch(getDnsRoutingFilters());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(editDnsRoutingFilterFailure());
    }
};

// Remove DNS routing filter
export const removeDnsRoutingFilterRequest = createAction('REMOVE_DNS_ROUTING_FILTER_REQUEST');
export const removeDnsRoutingFilterFailure = createAction('REMOVE_DNS_ROUTING_FILTER_FAILURE');
export const removeDnsRoutingFilterSuccess = createAction('REMOVE_DNS_ROUTING_FILTER_SUCCESS');

export const removeDnsRoutingFilter = (url: string) => async (dispatch: any) => {
    dispatch(removeDnsRoutingFilterRequest());
    try {
        await apiClient.removeFilter({ url, whitelist: false, dns_routing: true });
        dispatch(removeDnsRoutingFilterSuccess());
        dispatch(addSuccessToast('dns_routing_rule_removed'));
        dispatch(getDnsRoutingFilters());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(removeDnsRoutingFilterFailure());
    }
};

// Toggle DNS routing filter status
export const toggleDnsRoutingFilterRequest = createAction('TOGGLE_DNS_ROUTING_FILTER_REQUEST');
export const toggleDnsRoutingFilterFailure = createAction('TOGGLE_DNS_ROUTING_FILTER_FAILURE');
export const toggleDnsRoutingFilterSuccess = createAction('TOGGLE_DNS_ROUTING_FILTER_SUCCESS');

export const toggleDnsRoutingFilter = (url: string, data: any) => async (dispatch: any) => {
    dispatch(toggleDnsRoutingFilterRequest());
    try {
        await apiClient.setFilterUrl({
            url,
            data: {
                ...data,
                whitelist: false,
                dns_routing: true,
            },
        });
        dispatch(toggleDnsRoutingFilterSuccess());
        dispatch(getDnsRoutingFilters());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(toggleDnsRoutingFilterFailure());
    }
};

// Refresh DNS routing filters
export const refreshDnsRoutingFiltersRequest = createAction('REFRESH_DNS_ROUTING_FILTERS_REQUEST');
export const refreshDnsRoutingFiltersFailure = createAction('REFRESH_DNS_ROUTING_FILTERS_FAILURE');
export const refreshDnsRoutingFiltersSuccess = createAction('REFRESH_DNS_ROUTING_FILTERS_SUCCESS');

export const refreshDnsRoutingFilters = (url?: string) => async (dispatch: any) => {
    dispatch(refreshDnsRoutingFiltersRequest());
    try {
        const config: any = { whitelist: false, dns_routing: true };
        if (url) {
            config.url = url;
        }
        await apiClient.refreshFilters(config);
        dispatch(refreshDnsRoutingFiltersSuccess());
        dispatch(addSuccessToast('dns_routing_refreshed'));
        dispatch(getDnsRoutingFilters());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(refreshDnsRoutingFiltersFailure());
    }
};
