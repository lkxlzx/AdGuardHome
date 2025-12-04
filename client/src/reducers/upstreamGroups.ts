import { handleActions } from 'redux-actions';

import * as actions from '../actions/upstreamGroups';
import { UpstreamGroupsState } from '../types/upstreamGroups';

const initialState: UpstreamGroupsState = {
    groups: [],
    processing: false,
    processingAdd: false,
    processingUpdate: false,
    processingDelete: false,
    processingTest: false,
    isModalOpen: false,
    modalType: 'add',
    currentGroup: undefined,
    testResult: undefined,
};

const upstreamGroups = handleActions(
    {
        // 获取分组列表
        [actions.getUpstreamGroupsRequest.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processing: true,
        }),
        [actions.getUpstreamGroupsSuccess.toString()]: (state: UpstreamGroupsState, { payload }: any) => ({
            ...state,
            groups: payload || [],
            processing: false,
        }),
        [actions.getUpstreamGroupsFailure.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processing: false,
        }),

        // 添加分组
        [actions.addUpstreamGroupRequest.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingAdd: true,
        }),
        [actions.addUpstreamGroupSuccess.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingAdd: false,
            isModalOpen: false,
        }),
        [actions.addUpstreamGroupFailure.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingAdd: false,
        }),

        // 更新分组
        [actions.updateUpstreamGroupRequest.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingUpdate: true,
        }),
        [actions.updateUpstreamGroupSuccess.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingUpdate: false,
            isModalOpen: false,
        }),
        [actions.updateUpstreamGroupFailure.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingUpdate: false,
        }),

        // 删除分组
        [actions.deleteUpstreamGroupRequest.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingDelete: true,
        }),
        [actions.deleteUpstreamGroupSuccess.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingDelete: false,
        }),
        [actions.deleteUpstreamGroupFailure.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingDelete: false,
        }),

        // 设置默认分组
        [actions.setDefaultGroupRequest.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingUpdate: true,
        }),
        [actions.setDefaultGroupSuccess.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingUpdate: false,
        }),
        [actions.setDefaultGroupFailure.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingUpdate: false,
        }),

        // 测试分组
        [actions.testUpstreamGroupRequest.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingTest: true,
            testResult: undefined,
        }),
        [actions.testUpstreamGroupSuccess.toString()]: (state: UpstreamGroupsState, { payload }: any) => ({
            ...state,
            processingTest: false,
            testResult: payload,
        }),
        [actions.testUpstreamGroupFailure.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            processingTest: false,
        }),

        // 对话框
        [actions.openModal.toString()]: (state: UpstreamGroupsState, { payload }: any) => ({
            ...state,
            isModalOpen: true,
            modalType: payload.modalType,
            currentGroup: payload.currentGroup,
            testResult: undefined,
        }),
        [actions.closeModal.toString()]: (state: UpstreamGroupsState) => ({
            ...state,
            isModalOpen: false,
            currentGroup: undefined,
            testResult: undefined,
        }),
    },
    initialState,
);

export default upstreamGroups;
