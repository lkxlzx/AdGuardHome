import React from 'react';
import { Trans } from 'react-i18next';

interface DnsRoutingActionsProps {
    onAddRule: () => void;
    onCheckUpdates: () => void;
    processingRefresh: boolean;
}

const DnsRoutingActions: React.FC<DnsRoutingActionsProps> = ({
    onAddRule,
    onCheckUpdates,
    processingRefresh,
}) => (
    <div className="card-actions">
        <button
            className="btn btn-success btn-standard mr-2 btn-large mb-2"
            type="button"
            onClick={onAddRule}>
            <Trans>add_routing_rule</Trans>
        </button>

        <button
            className="btn btn-primary btn-standard mb-2"
            type="button"
            onClick={onCheckUpdates}
            disabled={processingRefresh}>
            <Trans>check_updates_btn</Trans>
        </button>
    </div>
);

export default DnsRoutingActions;
