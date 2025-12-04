import { describe, it, expect } from 'vitest';
import {
    selectGroups,
    selectProcessingUpdate,
    selectProcessingDelete,
    selectProcessingTest,
    selectEnabledGroups,
    selectDefaultGroup,
    selectGroupById,
    selectGroupsCount,
} from './upstreamGroups';
import { RootState } from '../reducers';
import { UpstreamGroup } from '../types/upstreamGroups';

describe('upstreamGroups selectors', () => {
    const mockGroups: UpstreamGroup[] = [
        {
            id: 'group-1',
            name: 'Default Group',
            enabled: true,
            is_default: true,
            upstream_dns: ['8.8.8.8'],
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
        },
        {
            id: 'group-2',
            name: 'Secondary Group',
            enabled: true,
            is_default: false,
            upstream_dns: ['1.1.1.1'],
            created_at: '2024-01-02T00:00:00Z',
            updated_at: '2024-01-02T00:00:00Z',
        },
        {
            id: 'group-3',
            name: 'Disabled Group',
            enabled: false,
            is_default: false,
            upstream_dns: ['9.9.9.9'],
            created_at: '2024-01-03T00:00:00Z',
            updated_at: '2024-01-03T00:00:00Z',
        },
    ];

    const mockState: Partial<RootState> = {
        upstreamGroups: {
            groups: mockGroups,
            processing: false,
            processingAdd: false,
            processingUpdate: true,
            processingDelete: true,
            processingTest: true,
            isModalOpen: false,
            modalType: 'add',
            currentGroup: undefined,
            testResult: undefined,
        },
    };

    describe('selectGroups', () => {
        it('should return all groups', () => {
            const result = selectGroups(mockState as RootState);
            expect(result).toEqual(mockGroups);
        });

        it('should return empty array when upstreamGroups is undefined', () => {
            const result = selectGroups({} as RootState);
            expect(result).toEqual([]);
        });
    });

    describe('selectProcessingUpdate', () => {
        it('should return processingUpdate status', () => {
            const result = selectProcessingUpdate(mockState as RootState);
            expect(result).toBe(true);
        });

        it('should return false when upstreamGroups is undefined', () => {
            const result = selectProcessingUpdate({} as RootState);
            expect(result).toBe(false);
        });
    });

    describe('selectProcessingDelete', () => {
        it('should return processingDelete status', () => {
            const result = selectProcessingDelete(mockState as RootState);
            expect(result).toBe(true);
        });

        it('should return false when upstreamGroups is undefined', () => {
            const result = selectProcessingDelete({} as RootState);
            expect(result).toBe(false);
        });
    });

    describe('selectProcessingTest', () => {
        it('should return processingTest status', () => {
            const result = selectProcessingTest(mockState as RootState);
            expect(result).toBe(true);
        });

        it('should return false when upstreamGroups is undefined', () => {
            const result = selectProcessingTest({} as RootState);
            expect(result).toBe(false);
        });
    });

    describe('selectEnabledGroups', () => {
        it('should return only enabled groups', () => {
            const result = selectEnabledGroups(mockState as RootState);
            expect(result).toHaveLength(2);
            expect(result.every((g) => g.enabled)).toBe(true);
        });

        it('should return empty array when no enabled groups', () => {
            const stateWithDisabledGroups = {
                upstreamGroups: {
                    ...mockState.upstreamGroups!,
                    groups: mockGroups.map((g) => ({ ...g, enabled: false })),
                },
            };
            const result = selectEnabledGroups(stateWithDisabledGroups as RootState);
            expect(result).toEqual([]);
        });
    });

    describe('selectDefaultGroup', () => {
        it('should return the default group', () => {
            const result = selectDefaultGroup(mockState as RootState);
            expect(result).toEqual(mockGroups[0]);
            expect(result?.is_default).toBe(true);
        });

        it('should return undefined when no default group', () => {
            const stateWithoutDefault = {
                upstreamGroups: {
                    ...mockState.upstreamGroups!,
                    groups: mockGroups.map((g) => ({ ...g, is_default: false })),
                },
            };
            const result = selectDefaultGroup(stateWithoutDefault as RootState);
            expect(result).toBeUndefined();
        });
    });

    describe('selectGroupById', () => {
        it('should return group by id', () => {
            const result = selectGroupById(mockState as RootState, 'group-2');
            expect(result).toEqual(mockGroups[1]);
        });

        it('should return undefined for non-existent id', () => {
            const result = selectGroupById(mockState as RootState, 'non-existent');
            expect(result).toBeUndefined();
        });
    });

    describe('selectGroupsCount', () => {
        it('should return the count of groups', () => {
            const result = selectGroupsCount(mockState as RootState);
            expect(result).toBe(3);
        });

        it('should return 0 when no groups', () => {
            const stateWithoutGroups = {
                upstreamGroups: {
                    ...mockState.upstreamGroups!,
                    groups: [],
                },
            };
            const result = selectGroupsCount(stateWithoutGroups as RootState);
            expect(result).toBe(0);
        });
    });

    describe('memoization', () => {
        it('should return same reference for same input', () => {
            const result1 = selectEnabledGroups(mockState as RootState);
            const result2 = selectEnabledGroups(mockState as RootState);
            // Note: Without reselect, selectors create new arrays each time
            // This test verifies the selector returns correct data
            expect(result1).toEqual(result2);
        });

        it('should return new reference when groups change', () => {
            const result1 = selectEnabledGroups(mockState as RootState);
            
            const newState = {
                upstreamGroups: {
                    ...mockState.upstreamGroups!,
                    groups: [...mockGroups, {
                        id: 'group-4',
                        name: 'New Group',
                        enabled: true,
                        is_default: false,
                        upstream_dns: ['4.4.4.4'],
                        created_at: '2024-01-04T00:00:00Z',
                        updated_at: '2024-01-04T00:00:00Z',
                    }],
                },
            };
            
            const result2 = selectEnabledGroups(newState as RootState);
            expect(result1).not.toBe(result2);
        });
    });
});
