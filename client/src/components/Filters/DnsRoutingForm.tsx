import React from 'react';
import { useForm, Controller } from 'react-hook-form';
import { useTranslation, Trans } from 'react-i18next';
import { useSelector } from 'react-redux';
import { RootState } from '../../initialState';
import { validateRequiredValue } from '../../helpers/validators';
import { Input } from '../ui/Controls/Input';

type FormValues = {
    name: string;
    url: string;
    upstreamGroup: string;
    updateInterval: number;
    priority: number;
};

type Props = {
    closeModal: () => void;
    onSubmit: (values: FormValues) => void;
    processingAddFilter: boolean;
    processingConfigFilter: boolean;
    initialValues?: Partial<FormValues>;
};

export const DnsRoutingForm = ({
    closeModal,
    onSubmit,
    processingAddFilter,
    processingConfigFilter,
    initialValues,
}: Props) => {
    const { t } = useTranslation();
    const upstreamGroups = useSelector((state: RootState) => state.upstreamGroups.groups || []);

    const {
        control,
        handleSubmit,
        formState: { errors },
    } = useForm<FormValues>({
        defaultValues: {
            name: initialValues?.name || '',
            url: initialValues?.url || '',
            upstreamGroup: initialValues?.upstreamGroup || '',
            updateInterval: initialValues?.updateInterval ?? 0,
            priority: initialValues?.priority ?? 0,
        },
    });

    const processing = processingAddFilter || processingConfigFilter;

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <div className="modal-body modal-body--filters">
                {/* 名称 */}
                <div className="form__group">
                    <Controller
                        name="name"
                        control={control}
                        rules={{ validate: validateRequiredValue }}
                        render={({ field, fieldState }) => (
                            <Input
                                {...field}
                                type="text"
                                placeholder={t('enter_name_hint')}
                                error={fieldState.error?.message}
                                disabled={processing}
                                trimOnBlur
                            />
                        )}
                    />
                </div>

                {/* 规则URL/文件路径 */}
                <div className="form__group">
                    <Controller
                        name="url"
                        control={control}
                        rules={{ 
                            validate: (value) => {
                                // Validate URL format
                                if (!value || !value.trim()) {
                                    return t('form_error_required');
                                }
                                const trimmed = value.trim();
                                if (!trimmed.startsWith('http://') && !trimmed.startsWith('https://')) {
                                    return t('form_error_url_protocol');
                                }
                                if (trimmed.length < 12) {
                                    return t('form_error_url_too_short');
                                }
                                try {
                                    new URL(trimmed);
                                } catch {
                                    return t('form_error_url_format');
                                }
                                return undefined;
                            }
                        }}
                        render={({ field, fieldState }) => (
                            <Input
                                {...field}
                                type="text"
                                placeholder={t('enter_url_or_path_hint')}
                                error={fieldState.error?.message}
                                disabled={processing}
                                trimOnBlur
                            />
                        )}
                    />
                </div>

                <div className="form__description">
                    {t('enter_valid_routing_rule_url')}
                </div>

                {/* 目标上游组 */}
                <div className="form__group">
                    <div className="form__label">
                        <Trans>routing_rule_group</Trans>
                    </div>
                    <Controller
                        name="upstreamGroup"
                        control={control}
                        rules={{ validate: validateRequiredValue }}
                        render={({ field, fieldState }) => (
                            <>
                                <select
                                    {...field}
                                    className="form-control"
                                    disabled={processing}>
                                    <option value="">{t('custom_rule_select_group')}</option>
                                    {upstreamGroups
                                        .filter((group: any) => group.enabled)
                                        .map((group: any) => (
                                            <option key={group.id} value={group.id}>
                                                {group.name}
                                            </option>
                                        ))}
                                </select>
                                {fieldState.error && (
                                    <div className="form__error">{fieldState.error.message}</div>
                                )}
                            </>
                        )}
                    />
                </div>

                <div className="form__description">
                    {t('routing_rule_group_hint')}
                </div>

                {/* 定时更新间隔（分钟） */}
                <div className="form__group">
                    <div className="form__label">
                        <Trans>routing_rule_update_interval</Trans>
                    </div>
                    <Controller
                        name="updateInterval"
                        control={control}
                        rules={{
                            validate: (value) => {
                                const num = typeof value === 'string' ? parseInt(value) : value;
                                if (isNaN(num) || num < 0) {
                                    return t('form_error_negative');
                                }
                                if (num > 0 && num < 60) {
                                    return t('form_error_update_interval_too_short');
                                }
                                if (num > 525600) { // 1 year in minutes
                                    return t('form_error_update_interval_too_long');
                                }
                                return undefined;
                            }
                        }}
                        render={({ field, fieldState }) => (
                            <>
                                <input
                                    {...field}
                                    type="number"
                                    min="0"
                                    className="form-control"
                                    disabled={processing}
                                    value={field.value}
                                    onChange={(e) => field.onChange(parseInt(e.target.value, 10) || 0)}
                                />
                                {fieldState.error && (
                                    <div className="form__error">{fieldState.error.message}</div>
                                )}
                            </>
                        )}
                    />
                </div>

                <div className="form__description">
                    {t('routing_rule_update_interval_hint')}
                </div>

                {/* 优先级 */}
                <div className="form__group">
                    <div className="form__label">
                        <Trans>routing_rule_priority</Trans>
                    </div>
                    <Controller
                        name="priority"
                        control={control}
                        rules={{
                            validate: (value) => {
                                const num = typeof value === 'string' ? parseInt(value) : value;
                                if (isNaN(num) || num < 0) {
                                    return t('form_error_negative');
                                }
                                if (num > 1000) {
                                    return t('form_error_priority_too_high');
                                }
                                return undefined;
                            }
                        }}
                        render={({ field, fieldState }) => (
                            <>
                                <input
                                    {...field}
                                    type="number"
                                    min="0"
                                    max="1000"
                                    className="form-control"
                                    disabled={processing}
                                    value={field.value}
                                    onChange={(e) => field.onChange(parseInt(e.target.value, 10) || 0)}
                                />
                                {fieldState.error && (
                                    <div className="form__error">{fieldState.error.message}</div>
                                )}
                            </>
                        )}
                    />
                </div>

                <div className="form__description">
                    {t('routing_rule_priority_hint')}
                </div>
            </div>

            <div className="modal-footer">
                <button
                    type="button"
                    className="btn btn-secondary"
                    onClick={closeModal}
                    disabled={processing}>
                    <Trans>cancel_btn</Trans>
                </button>
                <button
                    type="submit"
                    className="btn btn-success"
                    disabled={processing}>
                    {processing ? (
                        <Trans>loading_table_status</Trans>
                    ) : (
                        <Trans>save_btn</Trans>
                    )}
                </button>
            </div>
        </form>
    );
};
