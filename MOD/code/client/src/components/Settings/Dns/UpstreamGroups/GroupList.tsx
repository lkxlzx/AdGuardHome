import React from 'react';
import { useTranslation } from 'react-i18next';
// @ts-expect-error FIXME: update react-table
import ReactTable from 'react-table';
import { useDispatch, useSelector } from 'react-redux';

import {
    updateUpstreamGroup,
    deleteUpstreamGroup,
    openEditModal,
    testUpstreamGroup,
} from '../../../../actions/upstreamGroups';
import { UpstreamGroup } from '../../../../types/upstreamGroups';
import { RootState } from '../../../../initialState';
import { TABLES_MIN_ROWS } from '../../../../helpers/constants';
import { LocalStorageHelper, LOCAL_STORAGE_KEYS } from '../../../../helpers/localStorageHelper';

interface GroupListProps {
    groups: UpstreamGroup[];
    processing: boolean;
    onAddGroup: () => void;
}

const GroupList: React.FC<GroupListProps> = ({ groups, processing, onAddGroup }) => {
    const { t } = useTranslation();
    const dispatch = useDispatch();

    const { processingUpdate, processingDelete, processingTest } = useSelector(
        (state: RootState) => state.upstreamGroups!,
    );

    const cellWrap = ({ value }: any) => (
        <div className="logs__row o-hidden">
            <span className="logs__text" title={value}>
                {value}
            </span>
        </div>
    );

    const renderCheckbox = ({ original }: any) => {
        const handleToggle = () => {
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
                    disabled={processingUpdate}
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
        const handleTest = () => {
            dispatch(testUpstreamGroup(original.id));
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
        const handleEdit = () => {
            dispatch(openEditModal(original));
        };

        const handleCopy = () => {
            const newGroup = {
                ...original,
                name: `${original.name} (Copy)`,
                is_default: false,
            };
            delete newGroup.id;
            dispatch(openEditModal(newGroup as UpstreamGroup));
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
                    className="btn btn-icon btn-outline-primary btn-sm mr-2"
                    onClick={handleCopy}
                    title={t('copy')}
                >
                    <svg className="icons icon12">
                        <use xlinkHref="#copy" />
                    </svg>
                </button>
                <button
                    type="button"
                    className="btn btn-icon btn-outline-secondary btn-sm"
                    onClick={handleDelete}
                    disabled={processingDelete}
                    title={t('delete')}
                >
                    <svg className="icons">
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
};

export default GroupList;
