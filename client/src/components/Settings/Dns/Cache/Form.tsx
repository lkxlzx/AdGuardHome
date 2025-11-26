import React from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { useDispatch, useSelector } from 'react-redux';

import i18next from 'i18next';
import { clearDnsCache } from '../../../../actions/dnsConfig';
import { CACHE_CONFIG_FIELDS, UINT32_RANGE } from '../../../../helpers/constants';
import { replaceZeroWithEmptyString } from '../../../../helpers/helpers';
import { RootState } from '../../../../initialState';
import { Checkbox } from '../../../ui/Controls/Checkbox';

const INPUTS_FIELDS = [
    {
        name: CACHE_CONFIG_FIELDS.cache_size,
        title: i18next.t('cache_size'),
        description: i18next.t('cache_size_desc'),
        placeholder: i18next.t('enter_cache_size'),
    },
    {
        name: CACHE_CONFIG_FIELDS.cache_ttl_min,
        title: i18next.t('cache_ttl_min_override'),
        description: i18next.t('cache_ttl_min_override_desc'),
        placeholder: i18next.t('enter_cache_ttl_min_override'),
    },
    {
        name: CACHE_CONFIG_FIELDS.cache_ttl_max,
        title: i18next.t('cache_ttl_max_override'),
        description: i18next.t('cache_ttl_max_override_desc'),
        placeholder: i18next.t('enter_cache_ttl_max_override'),
    },
];

const PREFETCH_INPUTS_FIELDS = [
    {
        name: CACHE_CONFIG_FIELDS.prefetch_threshold,
        title: i18next.t('prefetch_threshold'),
        description: i18next.t('prefetch_threshold_desc'),
        placeholder: i18next.t('prefetch_threshold_placeholder'),
        min: 1,
        max: 100,
    },
    {
        name: CACHE_CONFIG_FIELDS.prefetch_time_window,
        title: i18next.t('prefetch_time_window'),
        description: i18next.t('prefetch_time_window_desc'),
        placeholder: i18next.t('prefetch_time_window_placeholder'),
        min: 60,
        max: 86400,
    },
    {
        name: CACHE_CONFIG_FIELDS.prefetch_max_entries,
        title: i18next.t('prefetch_max_entries'),
        description: i18next.t('prefetch_max_entries_desc'),
        placeholder: i18next.t('prefetch_max_entries_placeholder'),
        min: 1000,
        max: 100000,
    },
    {
        name: CACHE_CONFIG_FIELDS.prefetch_cleanup_interval,
        title: i18next.t('prefetch_cleanup_interval'),
        description: i18next.t('prefetch_cleanup_interval_desc'),
        placeholder: i18next.t('prefetch_cleanup_interval_placeholder'),
        min: 900,
        max: 14400,
    },
    {
        name: CACHE_CONFIG_FIELDS.prefetch_soft_limit,
        title: i18next.t('prefetch_soft_limit'),
        description: i18next.t('prefetch_soft_limit_desc'),
        placeholder: i18next.t('prefetch_soft_limit_placeholder'),
        min: 10,
        max: 500,
    },
    {
        name: CACHE_CONFIG_FIELDS.prefetch_hard_limit,
        title: i18next.t('prefetch_hard_limit'),
        description: i18next.t('prefetch_hard_limit_desc'),
        placeholder: i18next.t('prefetch_hard_limit_placeholder'),
        min: 50,
        max: 1000,
    },
];

type FormData = {
    cache_enabled: boolean;
    cache_size: number;
    cache_ttl_min: number;
    cache_ttl_max: number;
    cache_optimistic: boolean;
    prefetch_enabled: boolean;
    prefetch_threshold: number;
    prefetch_time_window: number;
    prefetch_max_entries: number;
    prefetch_cleanup_interval: number;
    prefetch_soft_limit: number;
    prefetch_hard_limit: number;
};

type CacheFormProps = {
    initialValues?: Partial<FormData>;
    onSubmit: (data: FormData) => void;
};

const Form = ({ initialValues, onSubmit }: CacheFormProps) => {
    const { t } = useTranslation();
    const dispatch = useDispatch();

    const { processingSetConfig } = useSelector((state: RootState) => state.dnsConfig);

    const {
        register,
        handleSubmit,
        watch,
        control,
        formState: { isSubmitting },
    } = useForm<FormData>({
        mode: 'onBlur',
        defaultValues: {
            cache_enabled: initialValues?.cache_enabled || false,
            cache_size: initialValues?.cache_size || 0,
            cache_ttl_min: initialValues?.cache_ttl_min || 0,
            cache_ttl_max: initialValues?.cache_ttl_max || 0,
            cache_optimistic: initialValues?.cache_optimistic || false,
            prefetch_enabled: initialValues?.prefetch_enabled || false,
            prefetch_threshold: initialValues?.prefetch_threshold || 5,
            prefetch_time_window: initialValues?.prefetch_time_window || 3600,
            prefetch_max_entries: initialValues?.prefetch_max_entries || 10000,
            prefetch_cleanup_interval: initialValues?.prefetch_cleanup_interval || 3600,
            prefetch_soft_limit: initialValues?.prefetch_soft_limit || 50,
            prefetch_hard_limit: initialValues?.prefetch_hard_limit || 150,
        },
    });

    const cache_enabled = watch('cache_enabled');
    const cache_size = watch('cache_size');
    const cache_ttl_min = watch('cache_ttl_min');
    const cache_ttl_max = watch('cache_ttl_max');
    const prefetch_enabled = watch('prefetch_enabled');
    const prefetch_soft_limit = watch('prefetch_soft_limit');
    const prefetch_hard_limit = watch('prefetch_hard_limit');

    const minExceedsMax = cache_ttl_min > 0 && cache_ttl_max > 0 && cache_ttl_min > cache_ttl_max;
    const cacheSizeZeroWhenEnabled = cache_enabled && cache_size === 0;
    const prefetchSoftExceedsHard =
        prefetch_soft_limit > 0 && prefetch_hard_limit > 0 && prefetch_soft_limit > prefetch_hard_limit;

    const handleClearCache = () => {
        if (window.confirm(t('confirm_dns_cache_clear'))) {
            dispatch(clearDnsCache());
        }
    };

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <div className="row">
                <div className="col-12 col-md-7">
                    <div className="form__group form__group--settings">
                        <Controller
                            name="cache_enabled"
                            control={control}
                            render={({ field }) => (
                                <Checkbox
                                    {...field}
                                    data-testid="dns_cache_enabled"
                                    title={t('cache_enabled')}
                                    subtitle={t('cache_enabled_desc')}
                                    disabled={processingSetConfig}
                                />
                            )}
                        />
                    </div>
                </div>

                {INPUTS_FIELDS.map(({ name, title, description, placeholder }) => (
                    <div className="col-12" key={name}>
                        <div className="col-12 col-md-7 p-0">
                            <div className="form__group form__group--settings">
                                <label htmlFor={name} className="form__label form__label--with-desc">
                                    {title}
                                </label>

                                <div className="form__desc form__desc--top">{description}</div>

                                <input
                                    type="number"
                                    data-testid={`dns_${name}`}
                                    className="form-control"
                                    placeholder={placeholder}
                                    disabled={processingSetConfig}
                                    min={0}
                                    max={UINT32_RANGE.MAX}
                                    {...register(name as keyof FormData, {
                                        valueAsNumber: true,
                                        setValueAs: (value) => replaceZeroWithEmptyString(value),
                                    })}
                                />

                                {name === CACHE_CONFIG_FIELDS.cache_size && cacheSizeZeroWhenEnabled && (
                                    <span className="form__message form__message--error">
                                        {t('cache_size_validation')}
                                    </span>
                                )}
                            </div>
                        </div>
                    </div>
                ))}
                {minExceedsMax && <span className="text-danger pl-3 pb-3">{t('ttl_cache_validation')}</span>}
            </div>

            <div className="row">
                <div className="col-12 col-md-7">
                    <div className="form__group form__group--settings">
                        <Controller
                            name="cache_optimistic"
                            control={control}
                            render={({ field }) => (
                                <Checkbox
                                    {...field}
                                    data-testid="dns_cache_optimistic"
                                    title={t('cache_optimistic')}
                                    subtitle={t('cache_optimistic_desc')}
                                    disabled={processingSetConfig}
                                />
                            )}
                        />
                    </div>
                </div>
            </div>

            <hr className="my-4" />

            <div className="row">
                <div className="col-12">
                    <h5 className="mb-3">{t('prefetch_settings')}</h5>
                </div>
                <div className="col-12 col-md-7">
                    <div className="form__group form__group--settings">
                        <Controller
                            name="prefetch_enabled"
                            control={control}
                            render={({ field }) => (
                                <Checkbox
                                    {...field}
                                    data-testid="dns_prefetch_enabled"
                                    title={t('prefetch_enabled')}
                                    subtitle={t('prefetch_enabled_desc')}
                                    disabled={processingSetConfig}
                                />
                            )}
                        />
                    </div>
                </div>

                {prefetch_enabled &&
                    PREFETCH_INPUTS_FIELDS.map(({ name, title, description, placeholder, min, max }) => (
                        <div className="col-12" key={name}>
                            <div className="col-12 col-md-7 p-0">
                                <div className="form__group form__group--settings">
                                    <label htmlFor={name} className="form__label form__label--with-desc">
                                        {title}
                                    </label>

                                    <div className="form__desc form__desc--top">{description}</div>

                                    <input
                                        type="number"
                                        data-testid={`dns_${name}`}
                                        className="form-control"
                                        placeholder={placeholder}
                                        disabled={processingSetConfig}
                                        min={min}
                                        max={max}
                                        {...register(name as keyof FormData, {
                                            valueAsNumber: true,
                                            setValueAs: (value) => replaceZeroWithEmptyString(value),
                                        })}
                                    />
                                </div>
                            </div>
                        </div>
                    ))}
                {prefetchSoftExceedsHard && (
                    <span className="text-danger pl-3 pb-3">{t('prefetch_limit_validation')}</span>
                )}
            </div>

            <button
                type="submit"
                data-testid="dns_save"
                className="btn btn-success btn-standard btn-large"
                disabled={
                    isSubmitting ||
                    processingSetConfig ||
                    minExceedsMax ||
                    cacheSizeZeroWhenEnabled ||
                    prefetchSoftExceedsHard
                }>
                {t('save_btn')}
            </button>

            <button
                type="button"
                data-testid="dns_clear"
                className="btn btn-outline-secondary btn-standard form__button"
                onClick={handleClearCache}>
                {t('clear_cache')}
            </button>
        </form>
    );
};

export default Form;
