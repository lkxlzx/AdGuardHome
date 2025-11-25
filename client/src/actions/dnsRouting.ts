import { createAction } from 'redux-actions';
import apiClient from '../api/Api';
import { addErrorToast, addSuccessToast } from './toasts';

// Action types
export const getDnsRoutingStatusRequest = createAction('GET_DNS_ROUTING_STATUS_REQUEST');
export const getDnsRoutingStatusSuccess = createAction('GET_DNS_ROUTING_STATUS_SUCCESS');
export const getDnsRoutingStatusFailure = createAction('GET_DNS_ROUTING_STATUS_FAILURE');

export const addDnsRoutingRuleRequest = createAction('ADD_DNS_ROUTING_RULE_REQUEST');
export const addDnsRoutingRuleSuccess = createAction('ADD_DNS_ROUTING_RULE_SUCCESS');
export const addDnsRoutingRuleFailure = createAction('ADD_DNS_ROUTING_RULE_FAILURE');

export const removeDnsRoutingRuleRequest = createAction('REMOVE_DNS_ROUTING_RULE_REQUEST');
export const removeDnsRoutingRuleSuccess = createAction('REMOVE_DNS_ROUTING_RULE_SUCCESS');
export const removeDnsRoutingRuleFailure = createAction('REMOVE_DNS_ROUTING_RULE_FAILURE');

export const updateDnsRoutingRuleRequest = createAction('UPDATE_DNS_ROUTING_RULE_REQUEST');
export const updateDnsRoutingRuleSuccess = createAction('UPDATE_DNS_ROUTING_RULE_SUCCESS');
export const updateDnsRoutingRuleFailure = createAction('UPDATE_DNS_ROUTING_RULE_FAILURE');

export const toggleDnsRoutingRuleRequest = createAction('TOGGLE_DNS_ROUTING_RULE_REQUEST');
export const toggleDnsRoutingRuleSuccess = createAction('TOGGLE_DNS_ROUTING_RULE_SUCCESS');
export const toggleDnsRoutingRuleFailure = createAction('TOGGLE_DNS_ROUTING_RULE_FAILURE');

export const refreshDnsRoutingRequest = createAction('REFRESH_DNS_ROUTING_REQUEST');
export const refreshDnsRoutingSuccess = createAction('REFRESH_DNS_ROUTING_SUCCESS');
export const refreshDnsRoutingFailure = createAction('REFRESH_DNS_ROUTING_FAILURE');

export const toggleDnsRoutingModal = createAction('TOGGLE_DNS_ROUTING_MODAL');

// Action creators
export const getDnsRoutingStatus = () => async (dispatch: any) => {
    dispatch(getDnsRoutingStatusRequest());
    try {
        const data = await apiClient.getDnsRoutingStatus();
        dispatch(getDnsRoutingStatusSuccess(data));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(getDnsRoutingStatusFailure());
    }
};

export const addDnsRoutingRule =
    (url: string, name: string, groupId: string) => async (dispatch: any) => {
        dispatch(addDnsRoutingRuleRequest());
        try {
            await apiClient.addDnsRoutingUrl(url, name, groupId, true);
            dispatch(addDnsRoutingRuleSuccess());
            dispatch(addSuccessToast('dns_routing_rule_added'));
            dispatch(getDnsRoutingStatus());
        } catch (error) {
            dispatch(addErrorToast({ error }));
            dispatch(addDnsRoutingRuleFailure());
        }
    };

export const removeDnsRoutingRule = (url: string) => async (dispatch: any) => {
    dispatch(removeDnsRoutingRuleRequest());
    try {
        await apiClient.removeDnsRoutingUrl(url);
        dispatch(removeDnsRoutingRuleSuccess());
        dispatch(addSuccessToast('dns_routing_rule_removed'));
        dispatch(getDnsRoutingStatus());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(removeDnsRoutingRuleFailure());
    }
};

export const updateDnsRoutingRule =
    (url: string, data: any) => async (dispatch: any) => {
        dispatch(updateDnsRoutingRuleRequest());
        try {
            await apiClient.setDnsRoutingUrl(url, data);
            dispatch(updateDnsRoutingRuleSuccess());
            dispatch(addSuccessToast('dns_routing_rule_updated'));
            dispatch(getDnsRoutingStatus());
        } catch (error) {
            dispatch(addErrorToast({ error }));
            dispatch(updateDnsRoutingRuleFailure());
        }
    };

export const toggleDnsRoutingRule =
    (url: string, data: any) => async (dispatch: any) => {
        dispatch(toggleDnsRoutingRuleRequest());
        try {
            await apiClient.setDnsRoutingUrl(url, data);
            dispatch(toggleDnsRoutingRuleSuccess());
            dispatch(getDnsRoutingStatus());
        } catch (error) {
            dispatch(addErrorToast({ error }));
            dispatch(toggleDnsRoutingRuleFailure());
        }
    };

export const refreshDnsRouting = () => async (dispatch: any) => {
    dispatch(refreshDnsRoutingRequest());
    try {
        await apiClient.refreshDnsRouting();
        dispatch(refreshDnsRoutingSuccess());
        dispatch(addSuccessToast('dns_routing_refreshed'));
        dispatch(getDnsRoutingStatus());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(refreshDnsRoutingFailure());
    }
};
