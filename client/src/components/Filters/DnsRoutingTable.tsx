import React, { Component } from 'react';
// @ts-expect-error FIXME: update react-table
import ReactTable from 'react-table';
import { withTranslation, Trans } from 'react-i18next';
import CellWrap from '../ui/CellWrap';
import { formatDetailedDateTime } from '../../helpers/helpers';

interface DnsRoutingTableProps {
    filters: any[];
    loading: boolean;
    onEdit: (filter: any) => void;
    onDelete: (filter: any) => void;
    onToggle: (filter: any) => void;
    onRefresh: (filter: any) => void;
    upstreamGroups?: any[];
    t: (...args: unknown[]) => string;
}

class DnsRoutingTable extends Component<DnsRoutingTableProps> {
    getDateCell = (row: any) => {
        const { value } = row;
        return (
            <div className="logs__row" title={value}>
                {formatDetailedDateTime(value)}
            </div>
        );
    };

    renderCheckbox = ({ original }: any) => {
        const { onToggle } = this.props;
        const { enabled } = original;

        return (
            <label className="checkbox checkbox--form">
                <input
                    type="checkbox"
                    className="checkbox__input"
                    checked={enabled}
                    onChange={() => onToggle(original)}
                />
                <span className="checkbox__label" />
            </label>
        );
    };

    columns = [
        {
            Header: <Trans>enabled_table_header</Trans>,
            accessor: 'enabled',
            className: 'text-center',
            width: 90,
            Cell: this.renderCheckbox,
        },
        {
            Header: <Trans>routing_rule_name</Trans>,
            accessor: 'name',
            minWidth: 150,
            Cell: CellWrap,
        },
        {
            Header: <Trans>routing_rule_url</Trans>,
            accessor: 'url',
            minWidth: 200,
            Cell: (row: any) => {
                const { value } = row;
                return (
                    <div className="logs__row" title={value}>
                        <code style={{ fontSize: '13px', background: 'transparent', padding: 0, border: 'none' }}>{value}</code>
                    </div>
                );
            },
        },
        {
            Header: <Trans>dns_group_table_header</Trans>,
            accessor: 'upstream_group',
            className: 'text-center',
            minWidth: 120,
            Cell: ({ value }: any) => {
                const { upstreamGroups } = this.props;
                if (!upstreamGroups || !value) {
                    return <div className="logs__row">{value || '-'}</div>;
                }
                
                const group = upstreamGroups.find((g: any) => g.id === value);
                const displayName = group ? group.name : value;
                
                return <div className="logs__row">{displayName}</div>;
            },
        },
        {
            Header: <Trans>routing_rule_count</Trans>,
            accessor: 'rules_count',
            className: 'text-center',
            width: 100,
            Cell: ({ value }: any) => (
                <div className="logs__row">{value ? value.toLocaleString() : 0}</div>
            ),
        },
        {
            Header: <Trans>last_updated</Trans>,
            accessor: 'last_updated',
            className: 'text-center',
            minWidth: 150,
            Cell: this.getDateCell,
        },
        {
            Header: <Trans>actions_table_header</Trans>,
            accessor: 'actions',
            className: 'text-center',
            width: 150,
            sortable: false,
            resizable: false,
            Cell: (row: any) => {
                const { original } = row;
                const { t, onEdit, onRefresh, onDelete } = this.props;

                return (
                    <div className="logs__row logs__row--center">
                        <button
                            type="button"
                            className="btn btn-icon btn-outline-primary btn-sm mr-2"
                            title={t('edit_table_action')}
                            onClick={() => onEdit(original)}>
                            <svg className="icons icon12">
                                <use xlinkHref="#edit" />
                            </svg>
                        </button>

                        <button
                            type="button"
                            className="btn btn-icon btn-outline-info btn-sm mr-2"
                            title={t('refresh_table_action')}
                            onClick={() => onRefresh(original)}>
                            <svg className="icons icon12">
                                <use xlinkHref="#refresh" />
                            </svg>
                        </button>

                        <button
                            type="button"
                            className="btn btn-icon btn-outline-secondary btn-sm"
                            onClick={() => onDelete(original)}
                            title={t('delete_table_action')}>
                            <svg className="icons icon12">
                                <use xlinkHref="#delete" />
                            </svg>
                        </button>
                    </div>
                );
            },
        },
    ];

    render() {
        const { loading, filters, t } = this.props;

        return (
            <ReactTable
                data={filters}
                columns={this.columns}
                showPagination={filters.length > 10}
                defaultPageSize={10}
                loading={loading}
                minRows={7}
                className="-striped -highlight card-table-overflow"
                ofText="/"
                previousText={t('previous_btn')}
                nextText={t('next_btn')}
                pageText={t('page_table_footer_text')}
                rowsText={t('rows_table_footer_text')}
                loadingText={t('loading_table_status')}
                noDataText={
                    <div className="text-center p-4">
                        <svg className="icons icon--24 mb-3 text-muted">
                            <use xlinkHref="#list" />
                        </svg>
                        <p className="text-muted mb-1">{t('no_dns_routing_rules')}</p>
                        <p className="text-muted small mb-0">{t('click_add_button_to_create_rule')}</p>
                    </div>
                }
            />
        );
    }
}

export default withTranslation()(DnsRoutingTable);
