import React, { useEffect } from 'react';
import { withTranslation } from 'react-i18next';

import PageTitle from '../ui/PageTitle';
import Card from '../ui/Card';
import Modal from './Modal';
import Actions from './Actions';
import Table from './Table';
import CustomRuleModal, { CustomRule } from './CustomRuleModal';
import CustomRulesTable from './CustomRulesTable';

import { getCurrentFilter } from '../../helpers/helpers';
import { useDnsRoutingFilters } from '../../hooks/useDnsRoutingFilters';
import { useCustomRules } from '../../hooks/useCustomRules';

interface DnsConfig {
    custom_domain_rules?: CustomRule[];
    [key: string]: any;
}

interface DnsRoutingProps {
    getFilteringStatus: (...args: unknown[]) => unknown;
    getDnsConfig: (...args: unknown[]) => unknown;
    setDnsConfig: (config: DnsConfig) => Promise<void>;
    dnsConfig?: DnsConfig;
    addSuccessToast: (message: string) => unknown;
    addErrorToast: (error: any) => unknown;
    filtering: {
        modalType: string;
        modalFilterUrl: string;
        isModalOpen: boolean;
        isFilterAdded: boolean;
        processingRefreshFilters: boolean;
        processingRemoveFilter: boolean;
        processingAddFilter: boolean;
        processingConfigFilter: boolean;
        processingFilters: boolean;
        whitelistFilters: any[];
        dnsRoutingFilters?: any[];
    };
    upstreamGroups: any[];
    removeFilter: (...args: unknown[]) => unknown;
    toggleFilterStatus: (...args: unknown[]) => unknown;
    addFilter: (...args: unknown[]) => unknown;
    toggleFilteringModal: (...args: unknown[]) => unknown;
    handleRulesChange: (...args: unknown[]) => unknown;
    refreshFilters: (...args: unknown[]) => unknown;
    editFilter: (...args: unknown[]) => unknown;
    t: (...args: unknown[]) => string;
}

const DnsRouting: React.FC<DnsRoutingProps> = (props) => {
    const {
        getFilteringStatus,
        getDnsConfig,
        setDnsConfig,
        dnsConfig,
        addSuccessToast,
        addErrorToast,
        filtering,
        upstreamGroups,
        removeFilter,
        toggleFilterStatus,
        addFilter,
        toggleFilteringModal,
        refreshFilters,
        editFilter,
        t,
    } = props;

    // Initialize data on mount
    useEffect(() => {
        getFilteringStatus();
        getDnsConfig();
    }, [getFilteringStatus, getDnsConfig]);

    // Use custom hooks for business logic
    const {
        handleSubmit,
        handleDelete,
        toggleFilter,
        handleRefresh,
        openAddFiltersModal,
    } = useDnsRoutingFilters({
        filtering,
        addFilter,
        editFilter,
        removeFilter,
        toggleFilterStatus,
        refreshFilters,
        toggleFilteringModal,
        t,
    });

    const {
        customRules,
        isCustomRuleModalOpen,
        editingRule,
        openCustomRuleModal,
        closeCustomRuleModal,
        handleCustomRuleSubmit,
        handleEditCustomRule,
        handleToggleCustomRule,
        handleDeleteCustomRule,
    } = useCustomRules({
        dnsConfig,
        setDnsConfig,
        getDnsConfig,
        addSuccessToast,
        addErrorToast,
        t,
    });

    // Compute derived state
    const {
        dnsRoutingFilters,
        isModalOpen,
        isFilterAdded,
        processingRefreshFilters,
        processingRemoveFilter,
        processingAddFilter,
        processingConfigFilter,
        processingFilters,
        modalType,
        modalFilterUrl,
    } = filtering;

    const currentFilterData = getCurrentFilter(modalFilterUrl, dnsRoutingFilters || []);
    const loading =
        processingConfigFilter ||
        processingFilters ||
        processingAddFilter ||
        processingRemoveFilter ||
        processingRefreshFilters;

    const whitelist = false;

    return (
        <>
            <PageTitle title={t('dns_routing')} subtitle={t('dns_routing_desc')} />
            <div className="content">
                <div className="row">
                    <div className="col-md-12">
                        <Card subtitle={t('dns_routing_hint')}>
                            <Table
                                filters={dnsRoutingFilters || []}
                                loading={loading}
                                processingConfigFilter={processingConfigFilter}
                                toggleFilteringModal={toggleFilteringModal}
                                handleDelete={handleDelete}
                                toggleFilter={toggleFilter}
                                whitelist={whitelist}
                                upstreamGroups={upstreamGroups}
                                showUpstreamGroup={true}
                            />

                            <Actions
                                handleAdd={openAddFiltersModal}
                                handleRefresh={handleRefresh}
                                processingRefreshFilters={processingRefreshFilters}
                                whitelist={whitelist}
                                isRoutingRule={true}
                            />
                        </Card>

                        <Card 
                            title={t('custom_rules_card_title')} 
                            subtitle={t('custom_rules_card_subtitle')}>
                            <CustomRulesTable
                                customRules={customRules}
                                loading={false}
                                onEdit={handleEditCustomRule}
                                onDelete={handleDeleteCustomRule}
                                onToggle={handleToggleCustomRule}
                                upstreamGroups={upstreamGroups}
                            />
                            
                            <div className="card-actions">
                                <button 
                                    className="btn btn-success btn-standard mb-2" 
                                    type="button"
                                    onClick={openCustomRuleModal}>
                                    {t('add_custom_rule')}
                                </button>
                            </div>
                        </Card>
                    </div>
                </div>
            </div>

            <Modal
                filters={dnsRoutingFilters || []}
                isOpen={isModalOpen}
                toggleFilteringModal={toggleFilteringModal}
                addFilter={addFilter}
                isFilterAdded={isFilterAdded}
                processingAddFilter={processingAddFilter}
                processingConfigFilter={processingConfigFilter}
                handleSubmit={handleSubmit}
                modalType={modalType}
                currentFilterData={currentFilterData}
                whitelist={whitelist}
                isRoutingRule={true}
            />

            <CustomRuleModal
                isOpen={isCustomRuleModalOpen}
                onClose={closeCustomRuleModal}
                onSubmit={handleCustomRuleSubmit}
                editingRule={editingRule}
            />
        </>
    );
};

export default withTranslation()(DnsRouting);
