import React, { Component } from 'react';
// @ts-expect-error FIXME: update react-table
import ReactTable from 'react-table';
import { withTranslation, Trans } from 'react-i18next';
import CellWrap from '../ui/CellWrap';

interface CustomRulesTableProps {
    customRules: any[];
    loading: boolean;
    onEdit: (rule: any) => void;
    onDelete: (rule: any) => void;
    onToggle?: (rule: any) => void;
    upstreamGroups?: any[];
    t: (...args: unknown[]) => string;
}

class CustomRulesTable extends Component<CustomRulesTableProps> {
    columns = [
        {
            Header: <Trans>enabled_table_header</Trans>,
            accessor: 'enabled',
            className: 'text-center',
            width: 90,
            Cell: ({ value, original }: any) => {
                const { onToggle } = this.props as any;
                return (
                    <label className="checkbox checkbox--form">
                        <input
                            type="checkbox"
                            className="checkbox__input"
                            checked={value}
                            onChange={() => onToggle && onToggle(original)}
                        />
                        <span className="checkbox__label" />
                    </label>
                );
            },
        },
        {
            Header: <Trans>custom_rule_domain</Trans>,
            accessor: 'domain',
            minWidth: 180,
            Cell: CellWrap,
        },
        {
            Header: <Trans>custom_rule_match_type</Trans>,
            accessor: 'matchType',
            className: 'text-center',
            minWidth: 120,
            Cell: ({ value }: any) => {
                const { t } = this.props;
                const typeMap: Record<string, string> = {
                    'DOMAIN': t('custom_rule_match_exact'),
                    'DOMAIN-SUFFIX': t('custom_rule_match_suffix'),
                    'DOMAIN-KEYWORD': t('custom_rule_match_keyword'),
                };
                return <div className="logs__row">{typeMap[value] || value}</div>;
            },
        },
        {
            Header: <Trans>dns_group_table_header</Trans>,
            accessor: 'upstreamGroup',
            className: 'text-center',
            minWidth: 150,
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
            Header: <Trans>actions_table_header</Trans>,
            accessor: 'actions',
            className: 'text-center',
            width: 100,
            sortable: false,
            resizable: false,
            Cell: (row: any) => {
                const { original } = row;
                const { t, onEdit, onDelete } = this.props;

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
        const { loading, customRules, t } = this.props;

        return (
            <ReactTable
                data={customRules}
                columns={this.columns}
                showPagination={customRules.length > 10}
                defaultPageSize={10}
                loading={loading}
                minRows={6}
                ofText="/"
                previousText={t('previous_btn')}
                nextText={t('next_btn')}
                pageText={t('page_table_footer_text')}
                rowsText={t('rows_table_footer_text')}
                loadingText={t('loading_table_status')}
                noDataText={t('no_custom_rules_added')}
            />
        );
    }
}

export default withTranslation()(CustomRulesTable);
