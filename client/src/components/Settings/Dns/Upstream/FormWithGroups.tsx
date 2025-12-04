import React, { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Trans, useTranslation } from 'react-i18next';
import { useDispatch, useSelector } from 'react-redux';
import clsx from 'clsx';

import { DNS_REQUEST_OPTIONS, UINT32_RANGE, UPSTREAM_CONFIGURATION_WIKI_LINK } from '../../../../helpers/constants';
import { removeEmptyLines } from '../../../../helpers/helpers';
import { RootState } from '../../../../initialState';
import { Checkbox } from '../../../ui/Controls/Checkbox';
import { Textarea } from '../../../ui/Controls/Textarea';
import { Radio } from '../../../ui/Controls/Radio';
import { Input } from '../../../ui/Controls/Input';
import { validateRequiredValue } from '../../../../helpers/validators';
import { toNumber } from '../../../../helpers/form';
import { getUpstreamGroups, openAddModal } from '../../../../actions/upstreamGroups';
import GroupList from '../UpstreamGroups/GroupList';
import GroupModal from '../UpstreamGroups/GroupModal';
import i18next from 'i18next';

type FormData = {
    upstream_mode: string;
    local_ptr_upstreams: string;
    use_private_ptr_resolvers: boolean;
    resolve_clients: boolean;
    upstream_timeout: number;
};

type FormProps = {
    initialValues?: Partial<FormData>;
    onSubmit: (data: FormData) => void;
};

const upstreamModeOptions = [
    {
        label: i18next.t('load_balancing'),
        desc: <Trans components={{ br: <br />, b: <b /> }}>load_balancing_desc</Trans>,
        value: DNS_REQUEST_OPTIONS.LOAD_BALANCING,
    },
    {
        label: i18next.t('parallel_requests'),
        desc: <Trans components={{ br: <br />, b: <b /> }}>upstream_parallel</Trans>,
        value: DNS_REQUEST_OPTIONS.PARALLEL,
    },
    {
        label: i18next.t('fastest_addr'),
        desc: <Trans components={{ br: <br />, b: <b /> }}>fastest_addr_desc</Trans>,
        value: DNS_REQUEST_OPTIONS.FASTEST_ADDR,
    },
];

const FormWithGroups = ({ initialValues, onSubmit }: FormProps) => {
    const { t } = useTranslation();
    const dispatch = useDispatch();

    const {
        control,
        handleSubmit,
        formState: { isSubmitting, isDirty },
    } = useForm<FormData>({
        mode: 'onBlur',
        defaultValues: {
            upstream_mode: initialValues?.upstream_mode || DNS_REQUEST_OPTIONS.LOAD_BALANCING,
            local_ptr_upstreams: initialValues?.local_ptr_upstreams || '',
            use_private_ptr_resolvers: initialValues?.use_private_ptr_resolvers || false,
            resolve_clients: initialValues?.resolve_clients || false,
            upstream_timeout: initialValues?.upstream_timeout || 0,
        },
    });

    const processingSetConfig = useSelector((state: RootState) => state.dnsConfig.processingSetConfig);
    const defaultLocalPtrUpstreams = useSelector((state: RootState) => state.dnsConfig.default_local_ptr_upstreams);

    const { groups, processing, isModalOpen } = useSelector(
        (state: RootState) => state.upstreamGroups!,
    );

    useEffect(() => {
        dispatch(getUpstreamGroups());
    }, [dispatch]);

    const handleAddGroup = () => {
        dispatch(openAddModal());
    };

    return (
        <>
            <form onSubmit={handleSubmit(onSubmit)} className="form--upstream">
                <div className="row">
                    <label className="col form__label" htmlFor="upstream_groups">
                        <Trans
                            components={{
                                a: <a href={UPSTREAM_CONFIGURATION_WIKI_LINK} target="_blank" rel="noopener noreferrer" />,
                            }}>
                            upstream_dns_help
                        </Trans>{' '}
                        <Trans
                            components={[
                                <a
                                    href="https://link.adtidy.org/forward.html?action=dns_kb_providers&from=ui&app=home"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    key="0">
                                    DNS providers
                                </a>,
                            ]}>
                            dns_providers
                        </Trans>
                    </label>

                    <div className="col-12 mb-4">
                        <GroupList
                            groups={groups}
                            processing={processing}
                            onAddGroup={handleAddGroup}
                        />
                    </div>

                    <div className="col-12">
                        <hr />
                    </div>

                    <div className="col-12 mb-4">
                        <Controller
                            name="upstream_mode"
                            control={control}
                            render={({ field }) => (
                                <Radio
                                    {...field}
                                    options={upstreamModeOptions}
                                    disabled={processingSetConfig}
                                />
                            )}
                        />
                    </div>

                    <div className="col-12">
                        <label className="form__label form__label--with-desc" htmlFor="local_ptr">
                            {t('local_ptr_title')}
                        </label>

                        <div className="form__desc form__desc--top">{t('local_ptr_desc')}</div>

                        <div className="form__desc form__desc--top">
                            {defaultLocalPtrUpstreams?.length > 0
                                ? t('local_ptr_default_resolver', {
                                      ip: defaultLocalPtrUpstreams.map((s: any) => `"${s}"`).join(', '),
                                  })
                                : t('local_ptr_no_default_resolver')}
                        </div>

                        <Controller
                            name="local_ptr_upstreams"
                            control={control}
                            render={({ field }) => (
                                <Textarea
                                    {...field}
                                    id="local_ptr_upstreams"
                                    data-testid="local_ptr_upstreams"
                                    placeholder={t('local_ptr_placeholder')}
                                    disabled={processingSetConfig}
                                    trimOnBlur
                                />
                            )}
                        />

                        <div className="mt-4">
                            <Controller
                                name="use_private_ptr_resolvers"
                                control={control}
                                render={({ field }) => (
                                    <Checkbox
                                        {...field}
                                        data-testid="dns_use_private_ptr_resolvers"
                                        title={t('use_private_ptr_resolvers_title')}
                                        subtitle={t('use_private_ptr_resolvers_desc')}
                                        disabled={processingSetConfig}
                                    />
                                )}
                            />
                        </div>
                    </div>

                    <div className="col-12">
                        <hr />
                    </div>

                    <div className="col-12 mb-4">
                        <Controller
                            name="resolve_clients"
                            control={control}
                            render={({ field }) => (
                                <Checkbox
                                    {...field}
                                    data-testid="dns_resolve_clients"
                                    title={t('resolve_clients_title')}
                                    subtitle={t('resolve_clients_desc')}
                                    disabled={processingSetConfig}
                                />
                            )}
                        />
                    </div>

                    <div className="col-12">
                        <hr />
                    </div>

                    <div className="col-12 col-md-7">
                        <div className="form__group">
                            <label htmlFor="upstream_timeout" className="form__label form__label--with-desc">
                                <Trans>upstream_timeout</Trans>
                            </label>

                            <div className="form__desc form__desc--top">
                                <Trans>upstream_timeout_desc</Trans>
                            </div>

                            <Controller
                                name="upstream_timeout"
                                control={control}
                                rules={{ validate: validateRequiredValue }}
                                render={({ field }) => (
                                    <Input
                                        {...field}
                                        type="number"
                                        id="upstream_timeout"
                                        data-testid="upstream_timeout"
                                        placeholder={t('form_enter_upstream_timeout')}
                                        disabled={processingSetConfig}
                                        min={1}
                                        max={UINT32_RANGE.MAX}
                                        onChange={(e) => {
                                            const { value } = e.target;
                                            field.onChange(toNumber(value));
                                        }}
                                    />
                                )}
                            />
                        </div>
                    </div>
                </div>

                <div className="card-actions">
                    <div className="btn-list">
                        <button
                            type="submit"
                            data-testid="dns_upstream_save"
                            className="btn btn-success btn-standard"
                            disabled={isSubmitting || !isDirty || processingSetConfig}>
                            {t('apply_btn')}
                        </button>
                    </div>
                </div>
            </form>
            {isModalOpen && <GroupModal />}
        </>
    );
};

export default FormWithGroups;
