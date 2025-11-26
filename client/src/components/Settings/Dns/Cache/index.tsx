import React from 'react';
import { useTranslation } from 'react-i18next';
import { shallowEqual, useDispatch, useSelector } from 'react-redux';

import Card from '../../../ui/Card';

import Form from './Form';
import { setDnsConfig } from '../../../../actions/dnsConfig';

import { replaceEmptyStringsWithZeroes, replaceZeroWithEmptyString } from '../../../../helpers/helpers';
import { RootState } from '../../../../initialState';

const CacheConfig = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();
    const {
        cache_enabled,
        cache_size,
        cache_ttl_max,
        cache_ttl_min,
        cache_optimistic,
        prefetch_enabled,
        prefetch_threshold,
        prefetch_time_window,
        prefetch_max_entries,
        prefetch_cleanup_interval,
        prefetch_soft_limit,
        prefetch_hard_limit,
    } = useSelector((state: RootState) => state.dnsConfig, shallowEqual);

    const handleFormSubmit = (values: any) => {
        const completedFields = replaceEmptyStringsWithZeroes(values);
        dispatch(setDnsConfig(completedFields));
    };

    return (
        <Card
            title={t('dns_cache_config')}
            subtitle={t('dns_cache_config_desc')}
            bodyType="card-body box-body--settings"
            id="dns-config">
            <div className="form">
                <Form
                    initialValues={{
                        cache_enabled,
                        cache_size: replaceZeroWithEmptyString(cache_size),
                        cache_ttl_max: replaceZeroWithEmptyString(cache_ttl_max),
                        cache_ttl_min: replaceZeroWithEmptyString(cache_ttl_min),
                        cache_optimistic,
                        prefetch_enabled: prefetch_enabled || false,
                        prefetch_threshold: replaceZeroWithEmptyString(prefetch_threshold),
                        prefetch_time_window: replaceZeroWithEmptyString(prefetch_time_window),
                        prefetch_max_entries: replaceZeroWithEmptyString(prefetch_max_entries),
                        prefetch_cleanup_interval: replaceZeroWithEmptyString(prefetch_cleanup_interval),
                        prefetch_soft_limit: replaceZeroWithEmptyString(prefetch_soft_limit),
                        prefetch_hard_limit: replaceZeroWithEmptyString(prefetch_hard_limit),
                    }}
                    onSubmit={handleFormSubmit}
                />
            </div>
        </Card>
    );
};

export default CacheConfig;
