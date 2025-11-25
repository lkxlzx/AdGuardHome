import React from 'react';
import { useForm, Controller } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { validateRequiredValue } from '../../../../helpers/validators';
import { Input } from '../../../ui/Controls/Input';
import { Textarea } from '../../../ui/Controls/Textarea';
import { Checkbox } from '../../../ui/Controls/Checkbox';

type FormValues = {
    name: string;
    upstreams: string;
    enabled: boolean;
    is_default: boolean;
};

const defaultValues: FormValues = {
    name: '',
    upstreams: '',
    enabled: true,
    is_default: false,
};

type Props = {
    closeModal: () => void;
    onSubmit: (values: FormValues) => void;
    processing: boolean;
    initialValues?: any;
    isEdit?: boolean;
};

export const UpstreamGroupsForm = ({ closeModal, onSubmit, processing, initialValues, isEdit }: Props) => {
    const { t } = useTranslation();

    const { handleSubmit, control } = useForm({
        defaultValues: {
            ...defaultValues,
            ...initialValues,
        },
        mode: 'onBlur',
    });

    const handleFormSubmit = (values: FormValues) => {
        onSubmit(values);
    };

    return (
        <form onSubmit={handleSubmit(handleFormSubmit)}>
            <div className="modal-body">
                <div className="form__group">
                    <Controller
                        name="name"
                        control={control}
                        rules={{
                            required: t('form_error_required'),
                            validate: validateRequiredValue,
                        }}
                        render={({ field, fieldState }) => (
                            <Input
                                {...field}
                                type="text"
                                data-testid="upstream_group_name"
                                label={t('upstream_group_name')}
                                placeholder={t('upstream_group_name_placeholder')}
                                error={fieldState.error?.message}
                                disabled={processing}
                            />
                        )}
                    />
                </div>

                <div className="form__group">
                    <Controller
                        name="upstreams"
                        control={control}
                        rules={{
                            required: t('form_error_required'),
                            validate: validateRequiredValue,
                        }}
                        render={({ field, fieldState }) => (
                            <Textarea
                                {...field}
                                data-testid="upstream_group_servers"
                                label={t('upstream_group_servers')}
                                placeholder={t('upstream_group_servers_placeholder')}
                                error={fieldState.error?.message}
                                disabled={processing}
                                rows={4}
                            />
                        )}
                    />
                </div>

                <div className="form__desc">
                    <p>{t('upstream_group_servers_desc')}</p>
                </div>

                <div className="form__group">
                    <Controller
                        name="enabled"
                        control={control}
                        render={({ field: { value, onChange, ...field } }) => (
                            <Checkbox
                                {...field}
                                value={value || false}
                                onChange={onChange}
                                title={t('upstream_group_enabled_label')}
                                disabled={processing}
                            />
                        )}
                    />
                </div>

                <div className="form__group">
                    <Controller
                        name="is_default"
                        control={control}
                        render={({ field: { value, onChange, ...field } }) => (
                            <Checkbox
                                {...field}
                                value={value || false}
                                onChange={onChange}
                                title={t('upstream_group_set_as_default')}
                                disabled={processing}
                            />
                        )}
                    />
                </div>
            </div>

            <div className="modal-footer">
                <button
                    type="button"
                    className="btn btn-secondary btn-standard"
                    disabled={processing}
                    onClick={closeModal}>
                    {t('cancel_btn')}
                </button>

                <button type="submit" className="btn btn-success btn-standard" disabled={processing}>
                    {t('save_btn')}
                </button>
            </div>
        </form>
    );
};
