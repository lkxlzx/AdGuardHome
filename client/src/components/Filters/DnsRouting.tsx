import React, { Component } from 'react';
import { withTranslation } from 'react-i18next';

import PageTitle from '../ui/PageTitle';
import Card from '../ui/Card';
import Modal from './Modal';
import Actions from './Actions';
import Table from './Table';
import CustomRuleModal, { CustomRule } from './CustomRuleModal';
import CustomRulesTable from './CustomRulesTable';

import { MODAL_TYPE } from '../../helpers/constants';

import { getCurrentFilter } from '../../helpers/helpers';

interface DnsRoutingProps {
    getFilteringStatus: (...args: unknown[]) => unknown;
    getDnsConfig: (...args: unknown[]) => unknown;
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

interface DnsRoutingState {
    isCustomRuleModalOpen: boolean;
    customRules: CustomRule[];
    editingRule: CustomRule | null;
}

class DnsRouting extends Component<DnsRoutingProps, DnsRoutingState> {
    constructor(props: DnsRoutingProps) {
        super(props);
        this.state = {
            isCustomRuleModalOpen: false,
            customRules: [],
            editingRule: null,
        };
    }

    componentDidMount() {
        this.props.getFilteringStatus();
        this.props.getDnsConfig();
    }

    componentDidUpdate(prevProps: DnsRoutingProps) {
        // Load custom rules from DNS config when it's updated
        const { dnsConfig } = this.props as any;
        const prevDnsConfig = (prevProps as any).dnsConfig;
        
        if (dnsConfig && dnsConfig.custom_domain_rules && 
            dnsConfig.custom_domain_rules !== prevDnsConfig?.custom_domain_rules) {
            this.setState({ customRules: dnsConfig.custom_domain_rules || [] });
        }
    }

    handleSubmit = (values: any) => {
        const { name, url, upstreamGroup, updateInterval } = values;

        const { filtering } = this.props;

        // Prepare data with upstream group and update interval
        const filterData = {
            name,
            url,
            enabled: true,
            upstreamGroup: upstreamGroup || '',
            updateInterval: updateInterval !== undefined ? updateInterval : 0,
        };

        if (filtering.modalType === MODAL_TYPE.EDIT_FILTERS) {
            // Edit existing DNS routing rule
            this.props.editFilter(filtering.modalFilterUrl, filterData, false, true);
        } else {
            // Add new DNS routing rule
            this.props.addFilter(url, name, false, true, upstreamGroup, updateInterval);
        }
    };

    handleDelete = (url: any) => {
        if (window.confirm(this.props.t('list_confirm_delete'))) {
            // Delete DNS routing rule
            this.props.removeFilter(url, false, true);
        }
    };

    toggleFilter = (url: any, data: any) => {
        // Toggle DNS routing rule
        this.props.toggleFilterStatus(url, data, false, true);
    };

    handleRefresh = () => {
        // Refresh DNS routing rules
        this.props.refreshFilters({ whitelist: false, dns_routing: true });
    };

    openAddFiltersModal = () => {
        this.props.toggleFilteringModal({ type: MODAL_TYPE.ADD_FILTERS });
    };

    openCustomRuleModal = () => {
        this.setState({ isCustomRuleModalOpen: true });
    };

    closeCustomRuleModal = () => {
        this.setState({ 
            isCustomRuleModalOpen: false,
            editingRule: null 
        });
    };

    handleCustomRuleSubmit = async (rule: CustomRule) => {
        const { customRules, editingRule } = this.state;
        const { t, addSuccessToast, addErrorToast, getDnsConfig } = this.props;
        
        try {
            let updatedRules: CustomRule[];
            
            if (editingRule) {
                // Edit existing rule
                updatedRules = customRules.map(r => 
                    r === editingRule ? rule : r
                );
            } else {
                // Add new rule
                updatedRules = [...customRules, rule];
            }
            
            // Save to backend via DNS config
            const dnsConfig = (this.props as any).dnsConfig;
            const newConfig = {
                ...dnsConfig,
                custom_domain_rules: updatedRules,
            };
            
            // Call API to save - wait for success before updating state
            await (this.props as any).setDnsConfig(newConfig);
            
            // Reload DNS config from server
            await getDnsConfig();
            
            // Only update local state after successful API call
            this.setState({ 
                customRules: updatedRules, 
                editingRule: null,
                isCustomRuleModalOpen: false 
            });
            
            // Show success message
            addSuccessToast(t('custom_rule_saved'));
        } catch (error) {
            // On error, state remains unchanged
            addErrorToast({ error });
        }
    };

    handleEditCustomRule = (rule: CustomRule) => {
        this.setState({ 
            editingRule: rule,
            isCustomRuleModalOpen: true 
        });
    };

    handleToggleCustomRule = async (rule: CustomRule) => {
        const { t, addSuccessToast, addErrorToast, getDnsConfig } = this.props;
        
        try {
            const { customRules } = this.state;
            const updatedRules = customRules.map(r => 
                r === rule ? { ...r, enabled: !r.enabled } : r
            );
            
            // Save to backend via DNS config
            const dnsConfig = (this.props as any).dnsConfig;
            const newConfig = {
                ...dnsConfig,
                custom_domain_rules: updatedRules,
            };
            
            // Call API to save - wait for success before updating state
            await (this.props as any).setDnsConfig(newConfig);
            
            // Reload DNS config from server
            await getDnsConfig();
            
            // Only update local state after successful API call
            this.setState({ customRules: updatedRules });
            
            // Show success message
            addSuccessToast(t('custom_rule_saved'));
        } catch (error) {
            // On error, state remains unchanged
            addErrorToast({ error });
        }
    };

    handleDeleteCustomRule = async (rule: CustomRule) => {
        const { t, addSuccessToast, addErrorToast, getDnsConfig } = this.props;
        
        if (window.confirm(t('list_confirm_delete'))) {
            try {
                const { customRules } = this.state;
                const updatedRules = customRules.filter(r => r !== rule);
                
                // Save to backend via DNS config
                const dnsConfig = (this.props as any).dnsConfig;
                const newConfig = {
                    ...dnsConfig,
                    custom_domain_rules: updatedRules,
                };
                
                // Call API to save - wait for success before updating state
                await (this.props as any).setDnsConfig(newConfig);
                
                // Reload DNS config from server
                await getDnsConfig();
                
                // Only update local state after successful API call
                this.setState({ customRules: updatedRules });
                
                // Show success message
                addSuccessToast(t('custom_rule_deleted'));
            } catch (error) {
                // On error, state remains unchanged
                addErrorToast({ error });
            }
        }
    };

    render() {
        const {
            t,
            toggleFilteringModal,
            addFilter,
            upstreamGroups,
            filtering: {
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
            },
        } = this.props;
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
                                    handleDelete={this.handleDelete}
                                    toggleFilter={this.toggleFilter}
                                    whitelist={whitelist}
                                    upstreamGroups={upstreamGroups}
                                    showUpstreamGroup={true}
                                />

                                <Actions
                                    handleAdd={this.openAddFiltersModal}
                                    handleRefresh={this.handleRefresh}
                                    processingRefreshFilters={processingRefreshFilters}
                                    whitelist={whitelist}
                                    isRoutingRule={true}
                                />
                            </Card>

                            <Card 
                                title={t('custom_rules_card_title')} 
                                subtitle={t('custom_rules_card_subtitle')}>
                                <CustomRulesTable
                                    customRules={this.state.customRules}
                                    loading={false}
                                    onEdit={this.handleEditCustomRule}
                                    onDelete={this.handleDeleteCustomRule}
                                    onToggle={this.handleToggleCustomRule}
                                    upstreamGroups={upstreamGroups}
                                />
                                
                                <div className="card-actions">
                                    <button 
                                        className="btn btn-success btn-standard mb-2" 
                                        type="button"
                                        onClick={this.openCustomRuleModal}>
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
                    handleSubmit={this.handleSubmit}
                    modalType={modalType}
                    currentFilterData={currentFilterData}
                    whitelist={whitelist}
                    isRoutingRule={true}
                />

                <CustomRuleModal
                    isOpen={this.state.isCustomRuleModalOpen}
                    onClose={this.closeCustomRuleModal}
                    onSubmit={this.handleCustomRuleSubmit}
                    editingRule={this.state.editingRule}
                />
            </>
        );
    }
}

export default withTranslation()(DnsRouting);
