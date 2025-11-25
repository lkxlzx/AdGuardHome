import { connect } from 'react-redux';
import {
    setRules,
    getFilteringStatus,
    addFilter,
    removeFilter,
    toggleFilterStatus,
    toggleFilteringModal,
    refreshFilters,
    handleRulesChange,
    editFilter,
} from '../actions/filtering';
import { getDnsConfig, setDnsConfig } from '../actions/dnsConfig';
import { addSuccessToast, addErrorToast } from '../actions/toasts';

import DnsRouting from '../components/Filters/DnsRouting';

const mapStateToProps = (state: any) => {
    const { filtering, dnsConfig } = state;
    const props = { 
        filtering,
        upstreamGroups: dnsConfig.upstream_groups || [],
        dnsConfig,
    };
    return props;
};

const mapDispatchToProps = {
    setRules,
    getFilteringStatus,
    addFilter,
    removeFilter,
    toggleFilterStatus,
    toggleFilteringModal,
    refreshFilters,
    handleRulesChange,
    editFilter,
    getDnsConfig,
    setDnsConfig,
    addSuccessToast,
    addErrorToast,
};

export default connect(mapStateToProps, mapDispatchToProps)(DnsRouting);
