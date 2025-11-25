import React, { Component } from 'react';

// @ts-expect-error FIXME: update react-table
import ReactTable from 'react-table';
import { withTranslation } from 'react-i18next';

import { MODAL_TYPE, TABLES_MIN_ROWS } from '../../../../helpers/constants';
import { LocalStorageHelper, LOCAL_STORAGE_KEYS } from '../../../../helpers/localStorageHelper';
import { UpstreamGroup } from '../../../../initialState';

interface TableProps {
    t: (...args: unknown[]) => string;
    list: UpstreamGroup[];
    processing: boolean;
    processingAdd: boolean;
    processingDelete: boolean;
    processingUpdate: boolean;
    handleDelete: (group: UpstreamGroup) => void;
    toggleModal: (config: { type: string; currentGroup?: UpstreamGroup }) => void;
    toggleDefault: (group: UpstreamGroup) => void;
    toggleEnabled: (group: UpstreamGroup) => void;
}

class Table extends Component<TableProps> {
    cellWrap = ({ value }: any) => (
        <div className="logs__row o-hidden">
            <span className="logs__text" title={value}>
                {value}
            </span>
        </div>
    );

    renderCheckbox = ({ original }: any) => {
        const { processing, toggleEnabled } = this.props;

        return (
            <label className="checkbox">
                <input
                    data-testid="group-enabled"
                    type="checkbox"
                    className="checkbox__input"
                    onChange={() => toggleEnabled(original)}
                    checked={original.enabled || false}
                    disabled={processing}
                />

                <span className="checkbox__label" />
            </label>
        );
    };

    columns = [
        {
            Header: this.props.t('enabled_table_header'),
            accessor: 'enabled',
            Cell: this.renderCheckbox,
            width: 90,
            className: 'text-center',
            resizable: false,
            sortable: false,
        },
        {
            Header: this.props.t('upstream_group_name'),
            accessor: 'name',
            Cell: ({ value, original }: any) => (
                <div className="logs__row o-hidden">
                    <span className="logs__text" title={value}>
                        {value}
                        {original.is_default && (
                            <span className="badge badge-success ml-2">{this.props.t('default_group')}</span>
                        )}
                    </span>
                </div>
            ),
        },
        {
            Header: this.props.t('upstream_group_servers_short'),
            accessor: 'upstreams',
            Cell: ({ value }: any) => {
                // Handle both string and array formats
                const servers = Array.isArray(value) 
                    ? value 
                    : (typeof value === 'string' ? value.split('\n').filter((s: string) => s.trim()) : []);
                const displayText = servers.join(', ');
                return (
                    <div className="logs__row o-hidden">
                        <span className="logs__text text-muted" title={displayText}>
                            {displayText}
                        </span>
                    </div>
                );
            },
        },
        {
            Header: this.props.t('actions_table_header'),
            accessor: 'actions',
            maxWidth: 150,
            sortable: false,
            resizable: false,
            Cell: (row: any) => {
                const { original } = row;

                return (
                    <div className="logs__row logs__row--center">
                        <button
                            data-testid="set-default-group"
                            type="button"
                            className="btn btn-icon btn-outline-success btn-sm mr-2"
                            onClick={() => this.props.toggleDefault(original)}
                            disabled={original.is_default || this.props.processing}
                            title={this.props.t('set_as_default')}>
                            <svg className="icons icon12">
                                <use xlinkHref="#check" />
                            </svg>
                        </button>

                        <button
                            data-testid="edit-group"
                            type="button"
                            className="btn btn-icon btn-outline-primary btn-sm mr-2"
                            onClick={() => {
                                this.props.toggleModal({
                                    type: MODAL_TYPE.EDIT,
                                    currentGroup: original,
                                });
                            }}
                            disabled={this.props.processingUpdate}
                            title={this.props.t('edit_table_action')}>
                            <svg className="icons icon12">
                                <use xlinkHref="#edit" />
                            </svg>
                        </button>

                        <button
                            data-testid="delete-group"
                            type="button"
                            className="btn btn-icon btn-outline-secondary btn-sm"
                            onClick={() => this.props.handleDelete(original)}
                            disabled={original.is_default}
                            title={this.props.t('delete_table_action')}>
                            <svg className="icons">
                                <use xlinkHref="#delete" />
                            </svg>
                        </button>
                    </div>
                );
            },
        },
    ];

    render() {
        const { t, list, processing, processingAdd, processingDelete } = this.props;

        return (
            <ReactTable
                data={list || []}
                columns={this.columns}
                loading={processing || processingAdd || processingDelete}
                className="-striped -highlight card-table-overflow"
                showPagination
                defaultPageSize={LocalStorageHelper.getItem(LOCAL_STORAGE_KEYS.UPSTREAM_GROUPS_PAGE_SIZE) || 10}
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
                noDataText={t('upstream_groups_empty_table')}
            />
        );
    }
}

export default withTranslation()(Table);
