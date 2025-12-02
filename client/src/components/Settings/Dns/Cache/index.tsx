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
        prefetch_threshold_window,
        prefetch_retention_time,
        prefetch_dynamic_retention_max_multiplier,
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
                        prefetch_threshold: replaceZeroWithEmptyString(prefetch_threshold) || 2,
                        prefetch_threshold_window: replaceZeroWithEmptyString(prefetch_threshold_window) || 600,
                        prefetch_retention_time: replaceZeroWithEmptyString(prefetch_retention_time) || 0,
                        prefetch_dynamic_retention_max_multiplier: replaceZeroWithEmptyString(prefetch_dynamic_retention_max_multiplier) || 10,
                    }}
                    onSubmit={handleFormSubmit}
                />
            </div>
        </Card>
    );
};

export default CacheConfig;
