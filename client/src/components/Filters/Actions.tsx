import React from 'react';
import { withTranslation, Trans } from 'react-i18next';

interface ActionsProps {
    handleAdd: (...args: unknown[]) => unknown;
    handleRefresh: (...args: unknown[]) => unknown;
    handleAddCustomRule?: (...args: unknown[]) => unknown;
    processingRefreshFilters: boolean;
    whitelist?: boolean;
    isRoutingRule?: boolean;
}

const Actions = ({ handleAdd, handleRefresh, handleAddCustomRule, processingRefreshFilters, whitelist, isRoutingRule }: ActionsProps) => (
    <div className="card-actions">
        <button className="btn btn-success btn-standard mr-2 btn-large mb-2" type="submit" onClick={handleAdd}>
            {isRoutingRule ? <Trans>add_routing_rule</Trans> : (whitelist ? <Trans>add_allowlist</Trans> : <Trans>add_blocklist</Trans>)}
        </button>

        <button
            className="btn btn-primary btn-standard mr-2 mb-2"
            type="submit"
            onClick={handleRefresh}
            disabled={processingRefreshFilters}>
            <Trans>check_updates_btn</Trans>
        </button>

        {isRoutingRule && handleAddCustomRule && (
            <button
                className="btn btn-outline-success btn-standard mb-2"
                type="button"
                onClick={handleAddCustomRule}>
                <Trans>add_custom_rule</Trans>
            </button>
        )}
    </div>
);

export default withTranslation()(Actions);
