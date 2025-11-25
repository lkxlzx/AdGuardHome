/**
 * Custom hook for managing DNS routing filters
 */
import { useCallback } from 'react';
import { MODAL_TYPE } from '../helpers/constants';

interface UseDnsRoutingFiltersProps {
    filtering: {
        modalType: string;
        modalFilterUrl: string;
    };
    addFilter: (...args: unknown[]) => unknown;
    editFilter: (...args: unknown[]) => unknown;
    removeFilter: (...args: unknown[]) => unknown;
    toggleFilterStatus: (...args: unknown[]) => unknown;
    refreshFilters: (...args: unknown[]) => unknown;
    toggleFilteringModal: (...args: unknown[]) => unknown;
    t: (key: string) => string;
}

export const useDnsRoutingFilters = ({
    filtering,
    addFilter,
    editFilter,
    removeFilter,
    toggleFilterStatus,
    refreshFilters,
    toggleFilteringModal,
    t,
}: UseDnsRoutingFiltersProps) => {
    const handleSubmit = useCallback((values: any) => {
        const { name, url, upstreamGroup, updateInterval, priority } = values;

        const filterData = {
            name,
            url,
            enabled: true,
            upstreamGroup: upstreamGroup || '',
            updateInterval: updateInterval !== undefined ? updateInterval : 0,
            priority: priority !== undefined ? priority : 0,
        };

        if (filtering.modalType === MODAL_TYPE.EDIT_FILTERS) {
            editFilter(filtering.modalFilterUrl, filterData, false, true);
        } else {
            addFilter(url, name, false, true, upstreamGroup, updateInterval, priority);
        }
    }, [filtering.modalType, filtering.modalFilterUrl, addFilter, editFilter]);

    const handleDelete = useCallback((url: any) => {
        if (window.confirm(t('list_confirm_delete'))) {
            removeFilter(url, false, true);
        }
    }, [removeFilter, t]);

    const toggleFilter = useCallback((url: any, data: any) => {
        toggleFilterStatus(url, data, false, true);
    }, [toggleFilterStatus]);

    const handleRefresh = useCallback(() => {
        refreshFilters({ whitelist: false, dns_routing: true });
    }, [refreshFilters]);

    const openAddFiltersModal = useCallback(() => {
        toggleFilteringModal({ type: MODAL_TYPE.ADD_FILTERS });
    }, [toggleFilteringModal]);

    return {
        handleSubmit,
        handleDelete,
        toggleFilter,
        handleRefresh,
        openAddFiltersModal,
    };
};
