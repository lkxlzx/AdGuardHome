import { connect } from 'react-redux';
import {
    getDnsRoutingStatus,
    addDnsRoutingRule,
    removeDnsRoutingRule,
    updateDnsRoutingRule,
    toggleDnsRoutingRule,
    refreshDnsRouting,
    toggleDnsRoutingModal,
} from '../actions/dnsRouting';
import { getDnsConfig, setDnsConfig } from '../actions/dnsConfig';
import { addSuccessToast, addErrorToast } from '../actions/toasts';

import DnsRouting from '../components/Filters/DnsRouting';

const mapStateToProps = (state: any) => {
    const { dnsRouting, dnsConfig } = state;
    const props = {
        dnsRouting,
        upstreamGroups: dnsConfig.upstream_groups || [],
        dnsConfig,
    };
    return props;
};

const mapDispatchToProps = {
    getDnsRoutingStatus,
    addDnsRoutingRule,
    removeDnsRoutingRule,
    updateDnsRoutingRule,
    toggleDnsRoutingRule,
    refreshDnsRouting,
    toggleDnsRoutingModal,
    getDnsConfig,
    setDnsConfig,
    addSuccessToast,
    addErrorToast,
};

export default connect(mapStateToProps, mapDispatchToProps)(DnsRouting);
