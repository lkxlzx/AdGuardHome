import { describe, it, expect } from 'vitest';
import upstreamGroupsReducer from './upstreamGroups';
import * as actions from '../actions/upstreamGroups';
import { UpstreamGroupsState, UpstreamGroup } from '../types/upstreamGroups';

describe('upstreamGroups reducer', () => {
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

    const mockGroup: UpstreamGroup = {
        id: 'test-id-1',
        name: 'Test Group',
        enabled: true,
        is_default: false,
        upstream_dns: ['8.8.8.8', '1.1.1.1'],
        bootstrap_dns: ['9.9.9.9'],
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
    };

    describe('initial state', () => {
        it('should return the initial state', () => {
            const state = upstreamGroupsReducer(undefined, { type: '@@INIT' } as any);
            expect(state).toEqual(initialState);
        });
    });

    describe('getUpstreamGroups actions', () => {
        it('should handle getUpstreamGroupsRequest', () => {
            const action = actions.getUpstreamGroupsRequest();
            const state = upstreamGroupsReducer(initialState, action as any);
            
            expect(state.processing).toBe(true);
        });

        it('should handle getUpstreamGroupsSuccess', () => {
            const groups = [mockGroup];
            const action = actions.getUpstreamGroupsSuccess(groups);
            const state = upstreamGroupsReducer(
                { ...initialState, processing: true },
                action as any
            );
            
            expect(state.processing).toBe(false);
            expect(state.groups).toEqual(groups);
        });

        it('should handle getUpstreamGroupsFailure', () => {
            const action = actions.getUpstreamGroupsFailure();
            const state = upstreamGroupsReducer(
                { ...initialState, processing: true },
                action as any
            );
            
            expect(state.processing).toBe(false);
        });
    });

    describe('addUpstreamGroup actions', () => {
        it('should handle addUpstreamGroupRequest', () => {
            const action = actions.addUpstreamGroupRequest();
            const state = upstreamGroupsReducer(initialState, action as any);
            
            expect(state.processingAdd).toBe(true);
        });

        it('should handle addUpstreamGroupSuccess', () => {
            const action = actions.addUpstreamGroupSuccess();
            const state = upstreamGroupsReducer(
                { ...initialState, processingAdd: true, isModalOpen: true },
                action as any
            );
            
            expect(state.processingAdd).toBe(false);
            expect(state.isModalOpen).toBe(false);
        });

        it('should handle addUpstreamGroupFailure', () => {
            const action = actions.addUpstreamGroupFailure();
            const state = upstreamGroupsReducer(
                { ...initialState, processingAdd: true },
                action as any
            );
            
            expect(state.processingAdd).toBe(false);
        });
    });

    describe('updateUpstreamGroup actions', () => {
        it('should handle updateUpstreamGroupRequest', () => {
            const action = actions.updateUpstreamGroupRequest();
            const state = upstreamGroupsReducer(initialState, action as any);
            
            expect(state.processingUpdate).toBe(true);
        });

        it('should handle updateUpstreamGroupSuccess', () => {
            const action = actions.updateUpstreamGroupSuccess();
            const state = upstreamGroupsReducer(
                { ...initialState, processingUpdate: true, isModalOpen: true },
                action as any
            );
            
            expect(state.processingUpdate).toBe(false);
            expect(state.isModalOpen).toBe(false);
        });

        it('should handle updateUpstreamGroupFailure', () => {
            const action = actions.updateUpstreamGroupFailure();
            const state = upstreamGroupsReducer(
                { ...initialState, processingUpdate: true },
                action as any
            );
            
            expect(state.processingUpdate).toBe(false);
        });
    });

    describe('deleteUpstreamGroup actions', () => {
        it('should handle deleteUpstreamGroupRequest', () => {
            const action = actions.deleteUpstreamGroupRequest();
            const state = upstreamGroupsReducer(initialState, action as any);
            
            expect(state.processingDelete).toBe(true);
        });

        it('should handle deleteUpstreamGroupSuccess', () => {
            const action = actions.deleteUpstreamGroupSuccess();
            const state = upstreamGroupsReducer(
                { ...initialState, processingDelete: true },
                action as any
            );
            
            expect(state.processingDelete).toBe(false);
        });

        it('should handle deleteUpstreamGroupFailure', () => {
            const action = actions.deleteUpstreamGroupFailure();
            const state = upstreamGroupsReducer(
                { ...initialState, processingDelete: true },
                action as any
            );
            
            expect(state.processingDelete).toBe(false);
        });
    });

    describe('setDefaultGroup actions', () => {
        it('should handle setDefaultGroupRequest', () => {
            const action = actions.setDefaultGroupRequest();
            const state = upstreamGroupsReducer(initialState, action as any);
            
            expect(state.processingUpdate).toBe(true);
        });

        it('should handle setDefaultGroupSuccess', () => {
            const action = actions.setDefaultGroupSuccess();
            const state = upstreamGroupsReducer(
                { ...initialState, processingUpdate: true },
                action as any
            );
            
            expect(state.processingUpdate).toBe(false);
        });

        it('should handle setDefaultGroupFailure', () => {
            const action = actions.setDefaultGroupFailure();
            const state = upstreamGroupsReducer(
                { ...initialState, processingUpdate: true },
                action as any
            );
            
            expect(state.processingUpdate).toBe(false);
        });
    });

    describe('testUpstreamGroup actions', () => {
        it('should handle testUpstreamGroupRequest', () => {
            const action = actions.testUpstreamGroupRequest();
            const state = upstreamGroupsReducer(initialState, action as any);
            
            expect(state.processingTest).toBe(true);
            expect(state.testResult).toBeUndefined();
        });

        it('should handle testUpstreamGroupSuccess', () => {
            const testResult = {
                group_id: 'test-id-1',
                results: [
                    { upstream: '8.8.8.8', success: true, rtt: 20 },
                    { upstream: '1.1.1.1', success: true, rtt: 15 },
                ],
            };
            const action = actions.testUpstreamGroupSuccess(testResult);
            const state = upstreamGroupsReducer(
                { ...initialState, processingTest: true },
                action as any
            );
            
            expect(state.processingTest).toBe(false);
            expect(state.testResult).toEqual(testResult);
        });

        it('should handle testUpstreamGroupFailure', () => {
            const action = actions.testUpstreamGroupFailure();
            const state = upstreamGroupsReducer(
                { ...initialState, processingTest: true },
                action as any
            );
            
            expect(state.processingTest).toBe(false);
        });
    });

    describe('modal actions', () => {
        it('should handle openModal for add', () => {
            const action = actions.openModal({ modalType: 'add' });
            const state = upstreamGroupsReducer(initialState, action as any);
            
            expect(state.isModalOpen).toBe(true);
            expect(state.modalType).toBe('add');
            expect(state.currentGroup).toBeUndefined();
            expect(state.testResult).toBeUndefined();
        });

        it('should handle openModal for edit', () => {
            const action = actions.openModal({
                modalType: 'edit',
                currentGroup: mockGroup,
            });
            const state = upstreamGroupsReducer(initialState, action as any);
            
            expect(state.isModalOpen).toBe(true);
            expect(state.modalType).toBe('edit');
            expect(state.currentGroup).toEqual(mockGroup);
        });

        it('should handle closeModal', () => {
            const action = actions.closeModal();
            const state = upstreamGroupsReducer(
                {
                    ...initialState,
                    isModalOpen: true,
                    currentGroup: mockGroup,
                    testResult: { group_id: 'test', results: [] },
                },
                action as any
            );
            
            expect(state.isModalOpen).toBe(false);
            expect(state.currentGroup).toBeUndefined();
            expect(state.testResult).toBeUndefined();
        });
    });

    describe('state immutability', () => {
        it('should not mutate the original state', () => {
            const originalState = { ...initialState };
            const action = actions.getUpstreamGroupsRequest();
            
            upstreamGroupsReducer(initialState, action as any);
            
            expect(initialState).toEqual(originalState);
        });
    });
});
