import React from 'react';
import { useTranslation } from 'react-i18next';
import { shallowEqual, useDispatch, useSelector } from 'react-redux';

import FormWithGroups from './FormWithGroups';
import Card from '../../../ui/Card';
import { setDnsConfig } from '../../../../actions/dnsConfig';
import { RootState } from '../../../../initialState';

const Upstream = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();
    const {
        upstream_mode,
        resolve_clients,
        local_ptr_upstreams,
        use_private_ptr_resolvers,
        upstream_timeout,
    } = useSelector((state: RootState) => state.dnsConfig, shallowEqual);

    const handleSubmit = (values: any) => {
        const {
            upstream_mode,
            resolve_clients,
            local_ptr_upstreams,
            use_private_ptr_resolvers,
            upstream_timeout,
        } = values;

        const dnsConfig = {
            upstream_mode,
            resolve_clients,
            local_ptr_upstreams,
            use_private_ptr_resolvers,
            upstream_timeout,
        };

        dispatch(setDnsConfig(dnsConfig));
    };

    return (
        <Card title={t('upstream_dns')} bodyType="card-body box-body--settings">
            <div className="row">
                <div className="col">
                    <FormWithGroups
                        initialValues={{
                            upstream_mode,
                            resolve_clients,
                            local_ptr_upstreams,
                            use_private_ptr_resolvers,
                            upstream_timeout,
                        }}
                        onSubmit={handleSubmit}
                    />
                </div>
            </div>
        </Card>
    );
};

export default Upstream;
