// Redux selectors for upstream groups
// Using memoization to prevent unnecessary re-renders

import { RootState } from '../initialState';
import { UpstreamGroup } from '../types/upstreamGroups';

// Base selector
export const selectUpstreamGroupsState = (state: RootState) => state.upstreamGroups;

// Memoized selectors
export const selectGroups = (state: RootState): UpstreamGroup[] => {
    return state.upstreamGroups?.groups || [];
};

export const selectProcessing = (state: RootState): boolean => {
    return state.upstreamGroups?.processing || false;
};

export const selectProcessingUpdate = (state: RootState): boolean => {
    return state.upstreamGroups?.processingUpdate || false;
};

export const selectProcessingDelete = (state: RootState): boolean => {
    return state.upstreamGroups?.processingDelete || false;
};

export const selectProcessingTest = (state: RootState): boolean => {
    return state.upstreamGroups?.processingTest || false;
};

export const selectIsModalOpen = (state: RootState): boolean => {
    return state.upstreamGroups?.isModalOpen || false;
};

export const selectModalType = (state: RootState): 'add' | 'edit' => {
    return state.upstreamGroups?.modalType || 'add';
};

export const selectCurrentGroup = (state: RootState): UpstreamGroup | undefined => {
    return state.upstreamGroups?.currentGroup;
};

export const selectTestResult = (state: RootState) => {
    return state.upstreamGroups?.testResult;
};

// Derived selectors
export const selectEnabledGroups = (state: RootState): UpstreamGroup[] => {
    const groups = selectGroups(state);
    return groups.filter(group => group.enabled);
};

export const selectDefaultGroup = (state: RootState): UpstreamGroup | undefined => {
    const groups = selectGroups(state);
    return groups.find(group => group.is_default);
};

export const selectGroupById = (state: RootState, id: string): UpstreamGroup | undefined => {
    const groups = selectGroups(state);
    return groups.find(group => group.id === id);
};

export const selectGroupsCount = (state: RootState): number => {
    return selectGroups(state).length;
};

export const selectEnabledGroupsCount = (state: RootState): number => {
    return selectEnabledGroups(state).length;
};
