import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Controller, useForm } from 'react-hook-form';
import { Input } from '../../../ui/Controls/Input';
import { Textarea } from '../../../ui/Controls/Textarea';

interface UpstreamGroup {
    id: string;
    name: string;
    upstreams: string;
    isDefault?: boolean; // 是否为默认组
}

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
            // 如果是第一个分组，自动设为默认组
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
                g.id === editingId
                    ? { ...g, name: data.name.trim(), upstreams: data.upstreams.trim() }
                    : g
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
            
            // 如果删除后没有默认组，将第一个设为默认
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
        <div className="upstream-groups">
            <div className="upstream-groups__header">
                <h4 className="upstream-groups__title">{t('upstream_groups_title')}</h4>
                <p className="form__desc">{t('upstream_groups_desc')}</p>
            </div>

            {groups.length === 0 && !isAdding && (
                <div className="upstream-groups__empty">
                    <p className="text-muted">{t('upstream_groups_empty')}</p>
                </div>
            )}

            {groups.length > 0 && (
                <div className="upstream-groups__list">
                    {groups.map((group) => (
                        <div
                            key={group.id}
                            className={`upstream-group-item ${group.isDefault ? 'upstream-group-item--default' : ''}`}>
                            {editingId === group.id ? (
                                <form onSubmit={handleSubmit(handleSave)} className="upstream-group-form">
                                    <div className="form-group">
                                        <label htmlFor={`edit-name-${group.id}`}>
                                            {t('upstream_group_name')}
                                        </label>
                                        <Controller
                                            name="name"
                                            control={control}
                                            rules={{ required: true }}
                                            render={({ field }) => (
                                                <Input
                                                    {...field}
                                                    id={`edit-name-${group.id}`}
                                                    placeholder={t('upstream_group_name_placeholder')}
                                                    disabled={disabled}
                                                />
                                            )}
                                        />
                                    </div>
                                    <div className="form-group">
                                        <label htmlFor={`edit-upstreams-${group.id}`}>
                                            {t('upstream_group_servers')}
                                        </label>
                                        <Controller
                                            name="upstreams"
                                            control={control}
                                            rules={{ required: true }}
                                            render={({ field }) => (
                                                <Textarea
                                                    {...field}
                                                    id={`edit-upstreams-${group.id}`}
                                                    placeholder={t('upstream_group_servers_placeholder')}
                                                    disabled={disabled}
                                                    rows={3}
                                                />
                                            )}
                                        />
                                    </div>
                                    <div className="upstream-group-actions">
                                        <button
                                            type="submit"
                                            className="btn btn-sm btn-success"
                                            disabled={disabled}>
                                            {t('save')}
                                        </button>
                                        <button
                                            type="button"
                                            className="btn btn-sm btn-secondary"
                                            onClick={handleCancel}
                                            disabled={disabled}>
                                            {t('cancel')}
                                        </button>
                                    </div>
                                </form>
                            ) : (
                                <div className="upstream-group-display">
                                    <div className="upstream-group-header">
                                        <div className="upstream-group-title-wrapper">
                                            <h5 className="upstream-group-name">
                                                {group.name}
                                                {group.isDefault && (
                                                    <span className="badge badge-primary ml-2">
                                                        {t('default_group')}
                                                    </span>
                                                )}
                                            </h5>
                                            {group.isDefault && (
                                                <p className="upstream-group-default-hint">
                                                    {t('default_group_hint')}
                                                </p>
                                            )}
                                        </div>
                                        <div className="upstream-group-buttons">
                                            {!group.isDefault && (
                                                <button
                                                    type="button"
                                                    className="btn btn-sm btn-outline-secondary"
                                                    onClick={() => handleSetDefault(group.id)}
                                                    disabled={disabled || isAdding || editingId !== null}>
                                                    {t('set_as_default')}
                                                </button>
                                            )}
                                            <button
                                                type="button"
                                                className="btn btn-sm btn-outline-primary"
                                                onClick={() => handleEdit(group)}
                                                disabled={disabled || isAdding || editingId !== null}>
                                                {t('edit')}
                                            </button>
                                            <button
                                                type="button"
                                                className="btn btn-sm btn-outline-danger"
                                                onClick={() => handleDelete(group.id)}
                                                disabled={disabled || isAdding || editingId !== null || group.isDefault}>
                                                {t('delete')}
                                            </button>
                                        </div>
                                    </div>
                                    <div className="upstream-group-content">
                                        <pre className="upstream-group-servers">{group.upstreams}</pre>
                                    </div>
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            )}

            {isAdding && (
                <div className="upstream-group-item upstream-group-item--new">
                    <form onSubmit={handleSubmit(handleSave)} className="upstream-group-form">
                        <div className="form-group">
                            <label htmlFor="new-group-name">{t('upstream_group_name')}</label>
                            <Controller
                                name="name"
                                control={control}
                                rules={{ required: true }}
                                render={({ field }) => (
                                    <Input
                                        {...field}
                                        id="new-group-name"
                                        placeholder={t('upstream_group_name_placeholder')}
                                        disabled={disabled}
                                    />
                                )}
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="new-group-upstreams">{t('upstream_group_servers')}</label>
                            <Controller
                                name="upstreams"
                                control={control}
                                rules={{ required: true }}
                                render={({ field }) => (
                                    <Textarea
                                        {...field}
                                        id="new-group-upstreams"
                                        placeholder={t('upstream_group_servers_placeholder')}
                                        disabled={disabled}
                                        rows={3}
                                    />
                                )}
                            />
                        </div>
                        <div className="upstream-group-actions">
                            <button type="submit" className="btn btn-sm btn-success" disabled={disabled}>
                                {t('add')}
                            </button>
                            <button
                                type="button"
                                className="btn btn-sm btn-secondary"
                                onClick={handleCancel}
                                disabled={disabled}>
                                {t('cancel')}
                            </button>
                        </div>
                    </form>
                </div>
            )}

            {!isAdding && editingId === null && (
                <button
                    type="button"
                    className="btn btn-outline-success btn-sm mt-3"
                    onClick={handleAdd}
                    disabled={disabled}>
                    <i className="fa fa-plus mr-2" />
                    {t('upstream_group_add')}
                </button>
            )}

            <style jsx>{`
                .upstream-groups {
                    margin-top: 2rem;
                    padding-top: 2rem;
                    border-top: 1px solid #dee2e6;
                }

                .upstream-groups__header {
                    margin-bottom: 1.5rem;
                }

                .upstream-groups__title {
                    font-size: 1.1rem;
                    font-weight: 600;
                    margin-bottom: 0.5rem;
                }

                .upstream-groups__empty {
                    padding: 2rem;
                    text-align: center;
                    background-color: #f8f9fa;
                    border: 1px dashed #dee2e6;
                    border-radius: 4px;
                    margin-bottom: 1rem;
                }

                .upstream-groups__empty .text-muted {
                    margin: 0;
                    color: #6c757d;
                }

                .upstream-groups__list {
                    display: flex;
                    flex-direction: column;
                    gap: 1rem;
                    margin-bottom: 1rem;
                }

                .upstream-group-item {
                    border: 1px solid #dee2e6;
                    border-radius: 4px;
                    padding: 1rem;
                    background-color: #f8f9fa;
                }

                .upstream-group-item--default {
                    background-color: #fff3cd;
                    border-color: #ffc107;
                    border-width: 2px;
                }

                .upstream-group-item--new {
                    background-color: #e7f3ff;
                    border-color: #0d6efd;
                }

                .upstream-group-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 0.75rem;
                }

                .upstream-group-title-wrapper {
                    flex: 1;
                }

                .upstream-group-name {
                    font-size: 1rem;
                    font-weight: 600;
                    margin: 0;
                    color: #495057;
                    display: flex;
                    align-items: center;
                }

                .upstream-group-default-hint {
                    font-size: 0.875rem;
                    color: #856404;
                    margin: 0.25rem 0 0 0;
                }

                .badge {
                    display: inline-block;
                    padding: 0.25em 0.6em;
                    font-size: 0.75rem;
                    font-weight: 600;
                    line-height: 1;
                    text-align: center;
                    white-space: nowrap;
                    vertical-align: baseline;
                    border-radius: 0.25rem;
                }

                .badge-primary {
                    color: #fff;
                    background-color: #0d6efd;
                }

                .ml-2 {
                    margin-left: 0.5rem;
                }

                .upstream-group-buttons {
                    display: flex;
                    gap: 0.5rem;
                }

                .upstream-group-content {
                    margin-top: 0.5rem;
                }

                .upstream-group-servers {
                    background-color: #fff;
                    border: 1px solid #ced4da;
                    border-radius: 4px;
                    padding: 0.75rem;
                    margin: 0;
                    font-size: 0.875rem;
                    color: #495057;
                    white-space: pre-wrap;
                    word-break: break-all;
                }

                .upstream-group-form {
                    display: flex;
                    flex-direction: column;
                    gap: 1rem;
                }

                .upstream-group-actions {
                    display: flex;
                    gap: 0.5rem;
                }

                .form-group {
                    margin-bottom: 0;
                }

                .form-group label {
                    display: block;
                    margin-bottom: 0.5rem;
                    font-weight: 500;
                    font-size: 0.875rem;
                    color: #495057;
                }
            `}</style>
        </div>
    );
};

export default UpstreamGroups;
