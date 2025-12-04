import React from 'react';
import { useTranslation } from 'react-i18next';
// @ts-expect-error FIXME: update react-table
import ReactTable from 'react-table';
import { useDispatch, useSelector } from 'react-redux';

import {
    updateUpstreamGroup,
    deleteUpstreamGroup,
    setDefaultGroup,
    openEditModal,
    testUpstreamGroup,
    testUpstreamGroupRequest,
    testUpstreamGroupSuccess,
    testUpstreamGroupFailure,
} from '../../../../actions/upstreamGroups';
import { addSuccessToast, addErrorToast } from '../../../../actions/toasts';
import apiClient from '../../../../api/Api';
import { UpstreamGroup } from '../../../../types/upstreamGroups';
import { RootState } from '../../../../initialState';
import { TABLES_MIN_ROWS } from '../../../../helpers/constants';
import { LocalStorageHelper, LOCAL_STORAGE_KEYS } from '../../../../helpers/localStorageHelper';
import {
    selectProcessingUpdate,
    selectProcessingDelete,
    selectProcessingTest,
} from '../../../../selectors/upstreamGroups';

interface GroupListProps {
    groups: UpstreamGroup[];
    processing: boolean;
    onAddGroup: () => void;
}

const GroupList: React.FC<GroupListProps> = React.memo(({ groups, processing, onAddGroup }) => {
    const { t } = useTranslation();
    const dispatch = useDispatch();

    // Use memoized selectors for better performance
    const processingUpdate = useSelector(selectProcessingUpdate);
    const processingDelete = useSelector(selectProcessingDelete);
    const processingTest = useSelector(selectProcessingTest);

    const cellWrap = React.useCallback(({ value }: any) => (
        <div className="logs__row o-hidden">
            <span className="logs__text" title={value}>
                {value}
            </span>
        </div>
    ), []);

    const renderCheckbox = ({ original }: any) => {
        const handleToggle = () => {
            // Prevent disabling default group
            if (original.is_default && original.enabled) {
                dispatch(addErrorToast({ error: t('cannot_disable_default_group') }));
                return;
            }
            
            dispatch(
                updateUpstreamGroup(original.id, {
                    ...original,
                    enabled: !original.enabled,
                }),
            );
        };

        return (
            <label className="checkbox">
                <input
                    type="checkbox"
                    className="checkbox__input"
                    onChange={handleToggle}
                    checked={original.enabled}
                    disabled={processingUpdate || original.is_default}
                    title={original.is_default ? t('cannot_disable_default_group') : ''}
                />
                <span className="checkbox__label" />
            </label>
        );
    };

    const renderGroupName = ({ original }: any) => (
        <div className="logs__row">
            <span className="logs__text">
                {original.name}
                {original.is_default && (
                    <span className="badge badge-success ml-2">{t('default')}</span>
                )}
            </span>
        </div>
    );

    const renderTest = ({ original }: any) => {
        const handleTest = async () => {
            dispatch(testUpstreamGroupRequest());
            try {
                const data = await apiClient.testUpstreamGroup(original.id);
                
                // Build detailed result message
                const resultLines = data.results.map((r: any) => {
                    if (r.success) {
                        return `- ${r.upstream}: ${t('success')} (${r.rtt}ms)`;
                    } else {
                        // Translate error message
                        const errorKey = `error_${r.error}`;
                        const translatedError = t(errorKey, r.error);
                        return `- ${r.upstream}: ${t('failed')} - ${translatedError}`;
                    }
                });
                
                const message = resultLines.join('\n');
                const successCount = data.results.filter((r: any) => r.success).length;
                const totalCount = data.results.length;
                
                if (successCount === totalCount) {
                    // All success
                    dispatch(addSuccessToast(message));
                } else if (successCount > 0) {
                    // Partial success - show as info
                    dispatch(addSuccessToast(message));
                } else {
                    // All failed
                    dispatch(addErrorToast({ error: message }));
                }
                
                dispatch(testUpstreamGroupSuccess(data));
            } catch (error) {
                dispatch(addErrorToast({ error }));
                dispatch(testUpstreamGroupFailure());
            }
        };

        return (
            <button
                type="button"
                className="btn btn-primary btn-sm"
                onClick={handleTest}
                disabled={processingTest}
            >
                {t('test')}
            </button>
        );
    };

    const renderUpstreams = ({ original }: any) => {
        const upstreams = original.upstream_dns.join(', ');
        return cellWrap({ value: upstreams });
    };

    const renderActions = ({ original }: any) => {
        const handleSetDefault = () => {
            if (original.is_default) {
                return;
            }
            dispatch(setDefaultGroup(original.id));
        };

        const handleEdit = () => {
            dispatch(openEditModal(original));
        };

        const handleDelete = () => {
            if (original.is_default) {
                // eslint-disable-next-line no-alert
                alert(t('cannot_delete_default_group'));
                return;
            }
            // eslint-disable-next-line no-alert
            if (window.confirm(t('confirm_delete_group', { name: original.name }))) {
                dispatch(deleteUpstreamGroup(original.id));
            }
        };

        return (
            <div className="logs__row logs__row--center">
                <button
                    type="button"
                    className="btn btn-icon btn-outline-success btn-sm mr-2"
                    onClick={handleSetDefault}
                    disabled={original.is_default || processingUpdate}
                    title={original.is_default ? t('default_group') : t('set_as_default')}
                >
                    <svg className="icons icon12">
                        <use xlinkHref="#check" />
                    </svg>
                </button>
                <button
                    type="button"
                    className="btn btn-icon btn-outline-primary btn-sm mr-2"
                    onClick={handleEdit}
                    disabled={processingUpdate}
                    title={t('edit')}
                >
                    <svg className="icons icon12">
                        <use xlinkHref="#edit" />
                    </svg>
                </button>
                <button
                    type="button"
                    className="btn btn-icon btn-outline-secondary btn-sm"
                    onClick={handleDelete}
                    disabled={processingDelete}
                    title={t('delete')}
                >
                    <svg className="icons icon12">
                        <use xlinkHref="#delete" />
                    </svg>
                </button>
            </div>
        );
    };

    const columns = [
        {
            Header: t('enabled_table_header'),
            accessor: 'enabled',
            Cell: renderCheckbox,
            width: 90,
            className: 'text-center',
            resizable: false,
        },
        {
            Header: t('group_name'),
            accessor: 'name',
            Cell: renderGroupName,
            minWidth: 150,
        },
        {
            Header: t('test'),
            accessor: 'test',
            Cell: renderTest,
            width: 100,
            className: 'text-center',
            sortable: false,
            resizable: false,
        },
        {
            Header: t('upstream_servers'),
            accessor: 'upstream_dns',
            Cell: renderUpstreams,
            minWidth: 200,
        },
        {
            Header: t('actions_table_header'),
            accessor: 'actions',
            Cell: renderActions,
            maxWidth: 150,
            className: 'text-center',
            sortable: false,
            resizable: false,
        },
    ];

    return (
        <>
            <ReactTable
                data={groups || []}
                columns={columns}
                loading={processing}
                className="-striped -highlight card-table-overflow"
                showPagination
                defaultPageSize={
                    LocalStorageHelper.getItem(LOCAL_STORAGE_KEYS.UPSTREAM_GROUPS_PAGE_SIZE) || 10
                }
                onPageSizeChange={(size: any) =>
                    LocalStorageHelper.setItem(LOCAL_STORAGE_KEYS.UPSTREAM_GROUPS_PAGE_SIZE, size)
                }
                minRows={TABLES_MIN_ROWS}
                ofText="/"
                previousText={t('previous_btn')}
                nextText={t('next_btn')}
                pageText={t('page_table_footer_text')}
                rowsText={t('rows_table_footer_text')}
                loadingText={t('loading_table_status')}
                noDataText={t('no_upstream_groups')}
            />
            <div className="card-actions">
                <button
                    type="button"
                    className="btn btn-success btn-standard"
                    onClick={onAddGroup}
                >
                    {t('add_upstream_group')}
                </button>
            </div>
        </>
    );
});

GroupList.displayName = 'GroupList';

export default GroupList;
