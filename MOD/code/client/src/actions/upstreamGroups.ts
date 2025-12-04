import { createAction } from 'redux-actions';
import i18next from 'i18next';

import apiClient from '../api/Api';
import { addErrorToast, addSuccessToast } from './toasts';
import { UpstreamGroup } from '../types/upstreamGroups';

// 获取分组列表
export const getUpstreamGroupsRequest = createAction('GET_UPSTREAM_GROUPS_REQUEST');
export const getUpstreamGroupsSuccess = createAction('GET_UPSTREAM_GROUPS_SUCCESS');
export const getUpstreamGroupsFailure = createAction('GET_UPSTREAM_GROUPS_FAILURE');

export const getUpstreamGroups = () => async (dispatch: any) => {
    dispatch(getUpstreamGroupsRequest());
    try {
        const data = await apiClient.getUpstreamGroups();
        dispatch(getUpstreamGroupsSuccess(data));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(getUpstreamGroupsFailure());
    }
};

// 添加分组
export const addUpstreamGroupRequest = createAction('ADD_UPSTREAM_GROUP_REQUEST');
export const addUpstreamGroupSuccess = createAction('ADD_UPSTREAM_GROUP_SUCCESS');
export const addUpstreamGroupFailure = createAction('ADD_UPSTREAM_GROUP_FAILURE');

export const addUpstreamGroup = (group: Partial<UpstreamGroup>) => async (dispatch: any) => {
    dispatch(addUpstreamGroupRequest());
    try {
        const data = await apiClient.addUpstreamGroup(group);
        dispatch(addUpstreamGroupSuccess(data));
        dispatch(addSuccessToast(i18next.t('upstream_group_created')));
        dispatch(getUpstreamGroups());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(addUpstreamGroupFailure());
    }
};

// 更新分组
export const updateUpstreamGroupRequest = createAction('UPDATE_UPSTREAM_GROUP_REQUEST');
export const updateUpstreamGroupSuccess = createAction('UPDATE_UPSTREAM_GROUP_SUCCESS');
export const updateUpstreamGroupFailure = createAction('UPDATE_UPSTREAM_GROUP_FAILURE');

export const updateUpstreamGroup = (id: string, group: Partial<UpstreamGroup>) => async (dispatch: any) => {
    dispatch(updateUpstreamGroupRequest());
    try {
        const data = await apiClient.updateUpstreamGroup(id, group);
        dispatch(updateUpstreamGroupSuccess(data));
        dispatch(addSuccessToast(i18next.t('upstream_group_updated')));
        dispatch(getUpstreamGroups());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(updateUpstreamGroupFailure());
    }
};

// 删除分组
export const deleteUpstreamGroupRequest = createAction('DELETE_UPSTREAM_GROUP_REQUEST');
export const deleteUpstreamGroupSuccess = createAction('DELETE_UPSTREAM_GROUP_SUCCESS');
export const deleteUpstreamGroupFailure = createAction('DELETE_UPSTREAM_GROUP_FAILURE');

export const deleteUpstreamGroup = (id: string) => async (dispatch: any) => {
    dispatch(deleteUpstreamGroupRequest());
    try {
        await apiClient.deleteUpstreamGroup(id);
        dispatch(deleteUpstreamGroupSuccess(id));
        dispatch(addSuccessToast(i18next.t('upstream_group_deleted')));
        dispatch(getUpstreamGroups());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(deleteUpstreamGroupFailure());
    }
};

// 设置默认分组
export const setDefaultGroupRequest = createAction('SET_DEFAULT_GROUP_REQUEST');
export const setDefaultGroupSuccess = createAction('SET_DEFAULT_GROUP_SUCCESS');
export const setDefaultGroupFailure = createAction('SET_DEFAULT_GROUP_FAILURE');

export const setDefaultGroup = (id: string) => async (dispatch: any) => {
    dispatch(setDefaultGroupRequest());
    try {
        await apiClient.setDefaultUpstreamGroup(id);
        dispatch(setDefaultGroupSuccess(id));
        dispatch(addSuccessToast(i18next.t('default_group_set')));
        dispatch(getUpstreamGroups());
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(setDefaultGroupFailure());
    }
};

// 测试分组连通性
export const testUpstreamGroupRequest = createAction('TEST_UPSTREAM_GROUP_REQUEST');
export const testUpstreamGroupSuccess = createAction('TEST_UPSTREAM_GROUP_SUCCESS');
export const testUpstreamGroupFailure = createAction('TEST_UPSTREAM_GROUP_FAILURE');

export const testUpstreamGroup = (id: string) => async (dispatch: any) => {
    dispatch(testUpstreamGroupRequest());
    try {
        const data = await apiClient.testUpstreamGroup(id);
        dispatch(testUpstreamGroupSuccess(data));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(testUpstreamGroupFailure());
    }
};

// 打开/关闭对话框
export const openModal = createAction('OPEN_UPSTREAM_GROUP_MODAL');
export const closeModal = createAction('CLOSE_UPSTREAM_GROUP_MODAL');

export const openAddModal = () => (dispatch: any) => {
    dispatch(openModal({ modalType: 'add', currentGroup: undefined }));
};

export const openEditModal = (group: UpstreamGroup) => (dispatch: any) => {
    dispatch(openModal({ modalType: 'edit', currentGroup: group }));
};

export const closeGroupModal = () => (dispatch: any) => {
    dispatch(closeModal());
};
