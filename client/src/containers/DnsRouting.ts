import { connect } from 'react-redux';
import {
    getDnsRoutingFilters,
    addDnsRoutingFilter,
    editDnsRoutingFilter,
    removeDnsRoutingFilter,
    toggleDnsRoutingFilter,
    refreshDnsRoutingFilters,
    toggleDnsRoutingModal,
} from '../actions/dnsRouting';
import { getDnsConfig, setDnsConfig } from '../actions/dnsConfig';
import { getUpstreamGroups } from '../actions/upstreamGroups';
import { addSuccessToast, addErrorToast } from '../actions/toasts';

import DnsRouting from '../components/Filters/DnsRouting';

const mapStateToProps = (state: any) => {
    const { dnsRouting, dnsConfig, upstreamGroups } = state;
    const props = { 
        dnsRouting,
        upstreamGroups: upstreamGroups.groups || [],
        dnsConfig,
    };
    return props;
};

const mapDispatchToProps = {
    getDnsRoutingFilters,
    addDnsRoutingFilter,
    editDnsRoutingFilter,
    removeDnsRoutingFilter,
    toggleDnsRoutingFilter,
    refreshDnsRoutingFilters,
    toggleDnsRoutingModal,
    getDnsConfig,
    setDnsConfig,
    getUpstreamGroups,
    addSuccessToast,
    addErrorToast,
};

export default connect(mapStateToProps, mapDispatchToProps)(DnsRouting);
