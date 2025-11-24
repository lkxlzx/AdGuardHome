import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Controller, useForm } from 'react-hook-form';
import { Input } from '../../../ui/Controls/Input';
import { Textarea } from '../../../ui/Controls/Textarea';
import { Checkbox } from '../../../ui/Controls/Checkbox';
import { UpstreamGroup } from '../../../../initialState';

interface UpstreamGroupsProps {
    groups: UpstreamGroup[];
    onChange: (groups: UpstreamGroup[]) => void;
    disabled?: boolean;
}

const UpstreamGroups: React.FC<UpstreamGroupsProps> = ({ groups, onChange, disabled }) => {
    const { t } = useTranslation();
    const [editingId, setEditingId] = useState<string | null>(null);
    const [isAdding, setIsAdding] = useState(false);

    const { control, handleSubmit, reset } = useForm<{ name: string; upstreams: string }>({
        defaultValues: { name: '', upstreams: '' },
    });

    const handleAdd = () => {
        setIsAdding(true);
        reset({ name: '', upstreams: '' });
    };

    const handleEdit = (group: UpstreamGroup) => {
        setEditingId(group.id);
        reset({ name: group.name, upstreams: group.upstreams });
    };

    const handleSave = (data: { name: string; upstreams: string }) => {
        if (!data.name.trim() || !data.upstreams.trim()) {
            return;
        }

        if (isAdding) {
            const isFirstGroup = groups.length === 0;
            const newGroup: UpstreamGroup = {
                id: `group_${Date.now()}`,
                name: data.name.trim(),
                upstreams: data.upstreams.trim(),
                isDefault: isFirstGroup,
            };
            onChange([...groups, newGroup]);
            setIsAdding(false);
        } else if (editingId) {
            const updatedGroups = groups.map((g) =>
                g.id === editingId ? { ...g, name: data.name.trim(), upstreams: data.upstreams.trim() } : g
            );
            onChange(updatedGroups);
            setEditingId(null);
        }

        reset({ name: '', upstreams: '' });
    };

    const handleCancel = () => {
        setIsAdding(false);
        setEditingId(null);
        reset({ name: '', upstreams: '' });
    };

    const handleDelete = (id: string) => {
        const group = groups.find((g) => g.id === id);
        if (group?.isDefault) {
            alert(t('upstream_group_cannot_delete_default'));
            return;
        }
        if (window.confirm(t('upstream_group_confirm_delete'))) {
            const remainingGroups = groups.filter((g) => g.id !== id);

            const hasDefault = remainingGroups.some((g) => g.isDefault);
            if (!hasDefault && remainingGroups.length > 0) {
                remainingGroups[0].isDefault = true;
            }

            onChange(remainingGroups);
        }
    };

    const handleSetDefault = (id: string) => {
        const updatedGroups = groups.map((g) => ({
            ...g,
            isDefault: g.id === id,
        }));
        onChange(updatedGroups);
    };

    return (
        <div className="upstream-groups-table">
            <div className="mb-3">
                <p className="form__desc">{t('upstream_groups_desc')}</p>
            </div>

            <div className="table-responsive">
                <table className="table table-hover">
                    <thead>
                        <tr>
                            <th style={{ width: '80px' }}>{t('enabled')}</th>
                            <th style={{ width: '200px' }}>{t('upstream_group_name')}</th>
                            <th>{t('upstream_group_servers_short')}</th>
                            <th style={{ width: '150px' }}>{t('actions')}</th>
                        </tr>
                    </thead>
                    <tbody>
                        {groups.length === 0 && !isAdding && (
                            <tr>
                                <td colSpan={4} className="text-center text-muted py-4">
                                    {t('upstream_groups_empty_table')}
                                </td>
                            </tr>
                        )}

                        {groups.map((group) =>
                            editingId === group.id ? (
                                <tr key={group.id} className="editing-row">
                                    <td colSpan={4}>
                                        <form onSubmit={handleSubmit(handleSave)} className="p-3">
                                            <div className="row">
                                                <div className="col-md-3 mb-3">
                                                    <label>{t('upstream_group_name')}</label>
                                                    <Controller
                                                        name="name"
                                                        control={control}
                                                        rules={{ required: true }}
                                                        render={({ field }) => (
                                                            <Input
                                                                {...field}
                                                                placeholder={t('upstream_group_name_placeholder')}
                                                                disabled={disabled}
                                                            />
                                                        )}
                                                    />
                                                </div>
                                                <div className="col-md-9 mb-3">
                                                    <label>{t('upstream_group_servers')}</label>
                                                    <Controller
                                                        name="upstreams"
                                                        control={control}
                                                        rules={{ required: true }}
                                                        render={({ field }) => (
                                                            <Textarea
                                                                {...field}
                                                                placeholder={t('upstream_group_servers_placeholder')}
                                                                disabled={disabled}
                                                                rows={2}
                                                            />
                                                        )}
                                                    />
                                                </div>
                                            </div>
                                            <div className="d-flex gap-2">
                                                <button type="submit" className="btn btn-success btn-sm" disabled={disabled}>
                                                    {t('save')}
                                                </button>
                                                <button
                                                    type="button"
                                                    className="btn btn-secondary btn-sm"
                                                    onClick={handleCancel}
                                                    disabled={disabled}>
                                                    {t('cancel')}
                                                </button>
                                            </div>
                                        </form>
                                    </td>
                                </tr>
                            ) : (
                                <tr key={group.id} className={group.isDefault ? 'table-warning' : ''}>
                                    <td>
                                        <div className="form-check">
                                            <input
                                                type="radio"
                                                className="form-check-input"
                                                checked={group.isDefault || false}
                                                onChange={() => handleSetDefault(group.id)}
                                                disabled={disabled}
                                            />
                                        </div>
                                    </td>
                                    <td>
                                        <strong>{group.name}</strong>
                                        {group.isDefault && (
                                            <span className="badge bg-primary ms-2">{t('default_group')}</span>
                                        )}
                                    </td>
                                    <td>
                                        <small className="text-muted">{group.upstreams.split('\n').join(', ')}</small>
                                    </td>
                                    <td>
                                        <div className="btn-group btn-group-sm">
                                            <button
                                                type="button"
                                                className="btn btn-outline-primary"
                                                onClick={() => handleEdit(group)}
                                                disabled={disabled || isAdding || editingId !== null}>
                                                {t('edit')}
                                            </button>
                                            <button
                                                type="button"
                                                className="btn btn-outline-danger"
                                                onClick={() => handleDelete(group.id)}
                                                disabled={disabled || isAdding || editingId !== null || group.isDefault}>
                                                {t('delete')}
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            )
                        )}

                        {isAdding && (
                            <tr className="editing-row">
                                <td colSpan={4}>
                                    <form onSubmit={handleSubmit(handleSave)} className="p-3 bg-light">
                                        <div className="row">
                                            <div className="col-md-3 mb-3">
                                                <label>{t('upstream_group_name')}</label>
                                                <Controller
                                                    name="name"
                                                    control={control}
                                                    rules={{ required: true }}
                                                    render={({ field }) => (
                                                        <Input
                                                            {...field}
                                                            placeholder={t('upstream_group_name_placeholder')}
                                                            disabled={disabled}
                                                        />
                                                    )}
                                                />
                                            </div>
                                            <div className="col-md-9 mb-3">
                                                <label>{t('upstream_group_servers')}</label>
                                                <Controller
                                                    name="upstreams"
                                                    control={control}
                                                    rules={{ required: true }}
                                                    render={({ field }) => (
                                                        <Textarea
                                                            {...field}
                                                            placeholder={t('upstream_group_servers_placeholder')}
                                                            disabled={disabled}
                                                            rows={2}
                                                        />
                                                    )}
                                                />
                                            </div>
                                        </div>
                                        <div className="d-flex gap-2">
                                            <button type="submit" className="btn btn-success btn-sm" disabled={disabled}>
                                                {t('add')}
                                            </button>
                                            <button
                                                type="button"
                                                className="btn btn-secondary btn-sm"
                                                onClick={handleCancel}
                                                disabled={disabled}>
                                                {t('cancel')}
                                            </button>
                                        </div>
                                    </form>
                                </td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>

            <div className="mt-3">
                <button
                    type="button"
                    className="btn btn-success btn-sm"
                    onClick={handleAdd}
                    disabled={disabled || isAdding || editingId !== null}>
                    {t('upstream_group_add')}
                </button>
                <button
                    type="button"
                    className="btn btn-primary btn-sm ms-2"
                    disabled={disabled}>
                    {t('upstream_group_import')}
                </button>
            </div>
        </div>
    );
};

export default UpstreamGroups;
