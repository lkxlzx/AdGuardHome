import React, { useEffect } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { useDispatch, useSelector } from 'react-redux';
import { useForm, Controller } from 'react-hook-form';
import ReactModal from 'react-modal';

import {
    addUpstreamGroup,
    updateUpstreamGroup,
    closeGroupModal,
    testUpstreamGroup,
} from '../../../../actions/upstreamGroups';
import { RootState } from '../../../../initialState';
import { validateRequiredValue } from '../../../../helpers/validators';

interface FormData {
    name: string;
    upstream_dns: string;
    fallback_dns?: string;
    bootstrap_dns?: string;
    enabled: boolean;
    is_default: boolean;
}

const GroupModal: React.FC = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();

    const { modalType, currentGroup, processingAdd, processingUpdate, processingTest, testResult } =
        useSelector((state: RootState) => state.upstreamGroups!);

    const isEdit = modalType === 'edit';
    const processing = processingAdd || processingUpdate;

    const {
        control,
        handleSubmit,
        formState: { errors },
        reset,
    } = useForm<FormData>({
        defaultValues: {
            name: currentGroup?.name || '',
            upstream_dns: currentGroup?.upstream_dns?.join('\n') || '',
            fallback_dns: (currentGroup as any)?.fallback_dns?.join('\n') || '',
            bootstrap_dns: currentGroup?.bootstrap_dns?.join('\n') || '',
            enabled: currentGroup?.enabled ?? true,
            is_default: currentGroup?.is_default ?? false,
        },
    });

    useEffect(() => {
        if (currentGroup) {
            reset({
                name: currentGroup.name,
                upstream_dns: currentGroup.upstream_dns.join('\n'),
                fallback_dns: (currentGroup as any)?.fallback_dns?.join('\n') || '',
                bootstrap_dns: currentGroup.bootstrap_dns?.join('\n') || '',
                enabled: currentGroup.enabled,
                is_default: currentGroup.is_default,
            });
        }
    }, [currentGroup, reset]);

    const onSubmit = (data: FormData) => {
        const upstreamDns = data.upstream_dns
            .split('\n')
            .map((s) => s.trim())
            .filter((s) => s.length > 0);

        const fallbackDns = data.fallback_dns
            ? data.fallback_dns
                  .split('\n')
                  .map((s) => s.trim())
                  .filter((s) => s.length > 0)
            : [];

        const bootstrapDns = data.bootstrap_dns
            ? data.bootstrap_dns
                  .split('\n')
                  .map((s) => s.trim())
                  .filter((s) => s.length > 0)
            : [];

        const groupData = {
            name: data.name,
            upstream_dns: upstreamDns,
            fallback_dns: fallbackDns.length > 0 ? fallbackDns : undefined,
            bootstrap_dns: bootstrapDns.length > 0 ? bootstrapDns : undefined,
            enabled: data.enabled,
            is_default: data.is_default,
        };

        if (isEdit && currentGroup) {
            dispatch(updateUpstreamGroup(currentGroup.id, groupData));
        } else {
            dispatch(addUpstreamGroup(groupData));
        }
    };

    const handleClose = () => {
        dispatch(closeGroupModal());
    };

    const handleTest = () => {
        if (currentGroup?.id) {
            dispatch(testUpstreamGroup(currentGroup.id));
        }
    };

    return (
        <ReactModal
            className="Modal__Bootstrap modal-dialog modal-dialog-centered upstream-group-modal"
            closeTimeoutMS={0}
            isOpen
            onRequestClose={handleClose}
        >
            <div className="modal-content">
                <div className="modal-header">
                    <h4 className="modal-title">
                        {isEdit ? <Trans>edit_group</Trans> : <Trans>add_upstream_group</Trans>}
                    </h4>
                    <button type="button" className="close" onClick={handleClose}>
                        <span className="sr-only">Close</span>
                    </button>
                </div>

                <form onSubmit={handleSubmit(onSubmit)}>
                    <div className="modal-body">
                        <div className="form__group">
                            <label className="form__label" htmlFor="name">
                                <Trans>group_name</Trans>
                            </label>
                            <Controller
                                name="name"
                                control={control}
                                rules={{
                                    validate: {
                                        required: validateRequiredValue,
                                    },
                                    maxLength: {
                                        value: 50,
                                        message: t('group_name_max_length'),
                                    },
                                }}
                                render={({ field, fieldState }) => (
                                    <>
                                        <input
                                            {...field}
                                            type="text"
                                            className="form-control"
                                            id="name"
                                            placeholder={t('group_name_placeholder')}
                                        />
                                        {fieldState.error && (
                                            <div className="form__message form__message--error">
                                                {fieldState.error.message}
                                            </div>
                                        )}
                                    </>
                                )}
                            />
                        </div>

                        <div className="form__group">
                            <label className="form__label" htmlFor="upstream_dns">
                                <Trans>upstream_servers</Trans>
                            </label>
                            <Controller
                                name="upstream_dns"
                                control={control}
                                rules={{
                                    validate: {
                                        required: validateRequiredValue,
                                        hasServers: (value) => {
                                            const servers = value
                                                .split('\n')
                                                .map((s: string) => s.trim())
                                                .filter((s: string) => s.length > 0);
                                            return (
                                                servers.length > 0 || t('upstream_servers_required')
                                            );
                                        },
                                    },
                                }}
                                render={({ field, fieldState }) => (
                                    <>
                                        <textarea
                                            {...field}
                                            className="form-control form-control--textarea"
                                            id="upstream_dns"
                                            rows={6}
                                            placeholder={t('upstream_dns_placeholder')}
                                        />
                                        {fieldState.error && (
                                            <div className="form__message form__message--error">
                                                <Trans>form_error_required</Trans>
                                            </div>
                                        )}
                                    </>
                                )}
                            />
                            <div className="form__desc form__desc--top">
                                <Trans>upstream_servers_help</Trans>
                            </div>
                        </div>

                        <div className="form__group">
                            <label className="form__label form__label--with-desc" htmlFor="fallback_dns">
                                <Trans>fallback_dns_title</Trans>
                            </label>
                            <div className="form__desc form__desc--top">
                                <Trans>fallback_dns_desc</Trans>
                            </div>
                            <Controller
                                name="fallback_dns"
                                control={control}
                                render={({ field }) => (
                                    <textarea
                                        {...field}
                                        className="form-control form-control--textarea"
                                        id="fallback_dns"
                                        rows={3}
                                        placeholder={t('fallback_dns_placeholder')}
                                    />
                                )}
                            />
                        </div>

                        <div className="form__group">
                            <label className="form__label form__label--with-desc" htmlFor="bootstrap_dns">
                                <Trans>bootstrap_dns</Trans>
                            </label>
                            <div className="form__desc form__desc--top">
                                <Trans>bootstrap_dns_desc</Trans>
                            </div>
                            <Controller
                                name="bootstrap_dns"
                                control={control}
                                render={({ field }) => (
                                    <textarea
                                        {...field}
                                        className="form-control form-control--textarea"
                                        id="bootstrap_dns"
                                        rows={3}
                                        placeholder={t('bootstrap_dns')}
                                    />
                                )}
                            />
                        </div>

                        <div className="form__group">
                            <label className="checkbox">
                                <Controller
                                    name="enabled"
                                    control={control}
                                    render={({ field: { value, onChange, ...field } }) => (
                                        <>
                                            <input
                                                {...field}
                                                type="checkbox"
                                                className="checkbox__input"
                                                id="enabled"
                                                checked={value}
                                                onChange={(e) => onChange(e.target.checked)}
                                            />
                                            <span className="checkbox__label">
                                                <Trans>enable_group</Trans>
                                            </span>
                                        </>
                                    )}
                                />
                            </label>
                        </div>

                        <div className="form__group">
                            <label className="checkbox">
                                <Controller
                                    name="is_default"
                                    control={control}
                                    render={({ field: { value, onChange, ...field } }) => (
                                        <>
                                            <input
                                                {...field}
                                                type="checkbox"
                                                className="checkbox__input"
                                                id="is_default"
                                                checked={value}
                                                onChange={(e) => onChange(e.target.checked)}
                                            />
                                            <span className="checkbox__label">
                                                <Trans>set_as_default</Trans>
                                            </span>
                                        </>
                                    )}
                                />
                            </label>
                        </div>

                        {testResult && (
                            <div className="alert alert-info test-results-alert">
                                <h6>
                                    <Trans>test_results</Trans>
                                </h6>
                                <ul>
                                    {testResult.results.map((result, index) => (
                                        <li key={index}>
                                            {result.upstream}:{' '}
                                            {result.success ? (
                                                <span className="text-success">
                                                    <Trans>success</Trans> ({result.rtt}ms)
                                                </span>
                                            ) : (
                                                <span className="text-danger">
                                                    <Trans>failed</Trans> - {result.error}
                                                </span>
                                            )}
                                        </li>
                                    ))}
                                </ul>
                            </div>
                        )}
                    </div>

                    <div className="modal-footer">
                        <div className="btn-list">
                            <button
                                type="button"
                                className="btn btn-secondary btn-standard"
                                onClick={handleClose}
                                disabled={processing}
                            >
                                <Trans>cancel_btn</Trans>
                            </button>
                            {isEdit && (
                                <button
                                    type="button"
                                    className="btn btn-primary btn-standard"
                                    onClick={handleTest}
                                    disabled={processingTest || !currentGroup?.id}
                                >
                                    {processingTest ? <Trans>testing</Trans> : <Trans>test</Trans>}
                                </button>
                            )}
                            <button
                                type="submit"
                                className="btn btn-success btn-standard"
                                disabled={processing}
                            >
                                {processing ? <Trans>saving</Trans> : <Trans>save_btn</Trans>}
                            </button>
                        </div>
                    </div>
                </form>
            </div>
        </ReactModal>
    );
};

export default GroupModal;
