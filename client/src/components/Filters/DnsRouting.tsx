import React, { useEffect } from 'react';
import { withTranslation } from 'react-i18next';

import PageTitle from '../ui/PageTitle';
import Card from '../ui/Card';
import { DnsRoutingModal } from './DnsRoutingModal';
import DnsRoutingActions from './DnsRoutingActions';
import DnsRoutingTable from './DnsRoutingTable';
import CustomRuleModal, { CustomRule } from './CustomRuleModal';
import CustomRulesTable from './CustomRulesTable';

import { useDnsRoutingCustomRules } from '../../hooks/useDnsRoutingCustomRules';

interface DnsConfig {
    custom_domain_rules?: CustomRule[];
    [key: string]: any;
}

interface DnsRoutingProps {
    getDnsRoutingFilters: () => unknown;
    addDnsRoutingFilter: (url: string, name: string, upstreamGroup: string, updateInterval: number, priority: number) => unknown;
    editDnsRoutingFilter: (url: string, data: any) => unknown;
    removeDnsRoutingFilter: (url: string) => unknown;
    toggleDnsRoutingFilter: (url: string, data: any) => unknown;
    refreshDnsRoutingFilters: (url?: string) => unknown;
    toggleDnsRoutingModal: (payload?: any) => unknown;
    getDnsConfig: (...args: unknown[]) => unknown;
    setDnsConfig: (config: DnsConfig) => Promise<void>;
    dnsConfig?: DnsConfig;
    addSuccessToast: (message: string) => unknown;
    addErrorToast: (error: any) => unknown;
    dnsRouting: {
        modalType: string;
        modalFilterUrl: string;
        isModalOpen: boolean;
        processingRefreshFilters: boolean;
        processingRemoveFilter: boolean;
        processingAddFilter: boolean;
        processingEditFilter: boolean;
        processingFilters: boolean;
        filters: any[];
    };
    upstreamGroups: any[];
    t: (...args: unknown[]) => string;
}

const DnsRouting: React.FC<DnsRoutingProps> = (props) => {
    const {
        getDnsRoutingFilters,
        addDnsRoutingFilter,
        editDnsRoutingFilter,
        removeDnsRoutingFilter,
        toggleDnsRoutingFilter,
        refreshDnsRoutingFilters,
        toggleDnsRoutingModal,
        getDnsConfig,
        setDnsConfig,
        dnsConfig,
        addSuccessToast,
        addErrorToast,
        dnsRouting,
        upstreamGroups,
        t,
    } = props;

    // Initialize data on mount
    useEffect(() => {
        getDnsRoutingFilters();
        getDnsConfig();
    }, [getDnsRoutingFilters, getDnsConfig]);

    // Handle form submit
    const handleSubmit = (values: any) => {
        const { name, url, upstreamGroup, updateInterval, priority } = values;
        
        if (dnsRouting.modalType === 'EDIT_FILTERS') {
            editDnsRoutingFilter(dnsRouting.modalFilterUrl, {
                name,
                upstreamGroup,
                updateInterval,
                priority,
            });
        } else {
            addDnsRoutingFilter(url, name, upstreamGroup, updateInterval, priority);
        }
    };

    const handleEdit = (filter: any) => {
        toggleDnsRoutingModal({ type: 'EDIT_FILTERS', url: filter.url });
    };

    const handleDelete = (filter: any) => {
        if (window.confirm(t('list_confirm_delete'))) {
            removeDnsRoutingFilter(filter.url);
        }
    };

    const handleToggle = (filter: any) => {
        const data = {
            ...filter,
            enabled: !filter.enabled,
        };
        toggleDnsRoutingFilter(filter.url, data);
    };

    const handleRefreshSingle = (filter: any) => {
        refreshDnsRoutingFilters(filter.url);
    };

    const handleRefreshAll = () => {
        refreshDnsRoutingFilters();
    };

    const handleAddRule = () => {
        toggleDnsRoutingModal({ type: 'ADD_FILTERS' });
    };

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
    } = useDnsRoutingCustomRules({
        dnsConfig,
        setDnsConfig,
        getDnsConfig,
        addSuccessToast,
        addErrorToast,
        t,
    });

    // Compute derived state
    const {
        filters,
        isModalOpen,
        processingRefreshFilters,
        processingRemoveFilter,
        processingAddFilter,
        processingEditFilter,
        processingFilters,
        modalType,
        modalFilterUrl,
    } = dnsRouting;

    // Get current filter data for editing
    const currentFilterData = filters.find((f: any) => f.url === modalFilterUrl) || null;
    
    const loading =
        processingEditFilter ||
        processingFilters ||
        processingAddFilter ||
        processingRemoveFilter ||
        processingRefreshFilters;

    return (
        <>
            <PageTitle title={t('dns_routing')} subtitle={t('dns_routing_desc')} />
            <div className="content">
                <div className="row">
                    <div className="col-md-12">
                        <Card subtitle={t('dns_routing_hint')}>
                            <DnsRoutingTable
                                filters={filters || []}
                                loading={loading}
                                onEdit={handleEdit}
                                onDelete={handleDelete}
                                onToggle={handleToggle}
                                onRefresh={handleRefreshSingle}
                                upstreamGroups={upstreamGroups}
                            />

                            <DnsRoutingActions
                                onAddRule={handleAddRule}
                                onCheckUpdates={handleRefreshAll}
                                processingRefresh={processingRefreshFilters}
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

            <DnsRoutingModal
                isOpen={isModalOpen}
                closeModal={() => toggleDnsRoutingModal()}
                onSubmit={handleSubmit}
                processingAddFilter={processingAddFilter}
                processingConfigFilter={processingEditFilter}
                modalType={modalType}
                currentFilterData={currentFilterData}
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
