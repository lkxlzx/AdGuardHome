import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { shallowEqual, useDispatch, useSelector } from 'react-redux';

import Form from './Form';
import UpstreamGroups from './UpstreamGroups';

import Card from '../../../ui/Card';
import { setDnsConfig } from '../../../../actions/dnsConfig';
import { RootState } from '../../../../initialState';

interface UpstreamGroup {
    id: string;
    name: string;
    upstreams: string;
}

const Upstream = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();
    const {
        upstream_dns,
        fallback_dns,
        bootstrap_dns,
        upstream_mode,
        resolve_clients,
        local_ptr_upstreams,
        use_private_ptr_resolvers,
        upstream_timeout,
    } = useSelector((state: RootState) => state.dnsConfig, shallowEqual);

    const upstream_dns_file = useSelector((state: RootState) => state.dnsConfig.upstream_dns_file);
    const upstream_groups = useSelector((state: RootState) => state.dnsConfig.upstream_groups);

    // 本地状态管理上游分组
    const [groups, setGroups] = useState<UpstreamGroup[]>([]);

    useEffect(() => {
        // 从 Redux 加载分组数据
        if (upstream_groups && Array.isArray(upstream_groups)) {
            setGroups(upstream_groups);
        }
    }, [upstream_groups]);

    const handleSubmit = (values: any) => {
        const {
            fallback_dns,
            bootstrap_dns,
            upstream_dns,
            upstream_mode,
            resolve_clients,
            local_ptr_upstreams,
            use_private_ptr_resolvers,
            upstream_timeout,
        } = values;

        const dnsConfig = {
            fallback_dns,
            bootstrap_dns,
            upstream_mode,
            resolve_clients,
            local_ptr_upstreams,
            use_private_ptr_resolvers,
            upstream_timeout,
            upstream_groups: groups, // 保存分组数据
            ...(upstream_dns_file ? null : { upstream_dns }),
        };

        dispatch(setDnsConfig(dnsConfig));
    };

    const handleGroupsChange = (newGroups: UpstreamGroup[]) => {
        setGroups(newGroups);
        
        // 立即保存分组更改
        const dnsConfig = {
            fallback_dns,
            bootstrap_dns,
            upstream_mode,
            resolve_clients,
            local_ptr_upstreams,
            use_private_ptr_resolvers,
            upstream_timeout,
            upstream_groups: newGroups,
            ...(upstream_dns_file ? null : { upstream_dns }),
        };

        dispatch(setDnsConfig(dnsConfig));
    };

    const upstreamDns = upstream_dns_file
        ? t('upstream_dns_configured_in_file', { path: upstream_dns_file })
        : upstream_dns;

    const processingSetConfig = useSelector((state: RootState) => state.dnsConfig.processingSetConfig);

    return (
        <Card title={t('upstream_dns')} bodyType="card-body box-body--settings">
            <div className="row">
                <div className="col">
                    <Form
                        initialValues={{
                            upstream_dns: upstreamDns,
                            fallback_dns,
                            bootstrap_dns,
                            upstream_mode,
                            resolve_clients,
                            local_ptr_upstreams,
                            use_private_ptr_resolvers,
                            upstream_timeout,
                        }}
                        onSubmit={handleSubmit}
                    />
                    
                    <UpstreamGroups
                        groups={groups}
                        onChange={handleGroupsChange}
                        disabled={processingSetConfig}
                    />
                </div>
            </div>
        </Card>
    );
};

export default Upstream;
