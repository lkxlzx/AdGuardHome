import React, { useState } from 'react';
import ReactModal from 'react-modal';
import { Trans, useTranslation } from 'react-i18next';
import { useSelector } from 'react-redux';
import { RootState } from '../../initialState';

interface CustomRuleModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (rule: CustomRule) => void;
    editingRule?: CustomRule | null;
}

export interface CustomRule {
    domain: string;
    matchType: 'DOMAIN' | 'DOMAIN-SUFFIX' | 'DOMAIN-KEYWORD';
    upstreamGroup: string;
    enabled?: boolean;
}

const CustomRuleModal: React.FC<CustomRuleModalProps> = ({ isOpen, onClose, onSubmit, editingRule }) => {
    const { t } = useTranslation();
    const [domain, setDomain] = useState('');
    const [matchType, setMatchType] = useState<'DOMAIN' | 'DOMAIN-SUFFIX' | 'DOMAIN-KEYWORD'>('DOMAIN-SUFFIX');
    const [upstreamGroup, setUpstreamGroup] = useState('');
    const [error, setError] = useState('');

    // Load editing rule data
    React.useEffect(() => {
        if (editingRule) {
            setDomain(editingRule.domain);
            setMatchType(editingRule.matchType);
            setUpstreamGroup(editingRule.upstreamGroup);
        }
    }, [editingRule]);

    // Get upstream groups from Redux state
    const upstreamGroups = useSelector((state: RootState) => state.dnsConfig.upstream_groups || []);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        
        // Validate domain
        if (!domain.trim()) {
            setError(t('form_error_required'));
            return;
        }

        // Validate upstream group
        if (!upstreamGroup) {
            setError(t('form_error_required'));
            return;
        }

        // Submit the rule
        onSubmit({
            domain: domain.trim(),
            matchType,
            upstreamGroup,
            enabled: editingRule?.enabled !== undefined ? editingRule.enabled : true,
        });

        // Reset form
        setDomain('');
        setMatchType('DOMAIN-SUFFIX');
        setUpstreamGroup('');
        setError('');
        onClose();
    };

    const handleClose = () => {
        setDomain('');
        setMatchType('DOMAIN-SUFFIX');
        setUpstreamGroup('');
        setError('');
        onClose();
    };

    return (
        <ReactModal
            className="Modal__Bootstrap modal-dialog modal-dialog-centered"
            closeTimeoutMS={0}
            isOpen={isOpen}
            onRequestClose={handleClose}>
            <div className="modal-content">
                <div className="modal-header">
                    <h4 className="modal-title">
                        {editingRule ? t('edit_custom_rule') : t('custom_rule_modal_title')}
                    </h4>
                    <button type="button" className="close" onClick={handleClose}>
                        <span className="sr-only">Close</span>
                    </button>
                </div>

                <form onSubmit={handleSubmit}>
                        <div className="modal-body">
                            {error && (
                                <div className="alert alert-danger" role="alert">
                                    {error}
                                </div>
                            )}

                            <div className="form-group">
                                <label htmlFor="domain">
                                    <Trans>custom_rule_domain</Trans>
                                </label>
                                <input
                                    type="text"
                                    className="form-control"
                                    id="domain"
                                    value={domain}
                                    onChange={(e) => setDomain(e.target.value)}
                                    placeholder={t('custom_rule_domain_placeholder')}
                                />
                                <small className="form-text text-muted">
                                    {matchType === 'DOMAIN' && '精确匹配: example.com'}
                                    {matchType === 'DOMAIN-SUFFIX' && '后缀匹配: *.example.com'}
                                    {matchType === 'DOMAIN-KEYWORD' && '关键字匹配: 包含该关键字的域名'}
                                </small>
                            </div>

                            <div className="form-group">
                                <label htmlFor="matchType">
                                    <Trans>custom_rule_match_type</Trans>
                                </label>
                                <select
                                    className="form-control"
                                    id="matchType"
                                    value={matchType}
                                    onChange={(e) => setMatchType(e.target.value as any)}>
                                    <option value="DOMAIN">
                                        {t('custom_rule_match_exact')}
                                    </option>
                                    <option value="DOMAIN-SUFFIX">
                                        {t('custom_rule_match_suffix')}
                                    </option>
                                    <option value="DOMAIN-KEYWORD">
                                        {t('custom_rule_match_keyword')}
                                    </option>
                                </select>
                            </div>

                            <div className="form-group">
                                <label htmlFor="upstreamGroup">
                                    <Trans>custom_rule_upstream_group</Trans>
                                </label>
                                <select
                                    className="form-control"
                                    id="upstreamGroup"
                                    value={upstreamGroup}
                                    onChange={(e) => setUpstreamGroup(e.target.value)}>
                                    <option value="">
                                        {t('custom_rule_select_group')}
                                    </option>
                                    {upstreamGroups
                                        .filter((group: any) => group.enabled)
                                        .map((group: any) => (
                                            <option key={group.id} value={group.id}>
                                                {group.name}
                                            </option>
                                        ))}
                                </select>
                            </div>
                        </div>

                        <div className="modal-footer">
                            <button
                                type="button"
                                className="btn btn-secondary"
                                onClick={handleClose}>
                                <Trans>cancel_btn</Trans>
                            </button>
                            <button type="submit" className="btn btn-success">
                                {editingRule ? t('save_btn') : t('add_btn')}
                            </button>
                        </div>
                    </form>
                </div>
            </ReactModal>
    );
};

export default CustomRuleModal;
