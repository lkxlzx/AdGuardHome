/**
 * Custom hook for managing DNS routing custom domain rules
 * This is completely independent from filtering custom rules
 */
import { useState, useCallback, useEffect } from 'react';
import { CustomRule } from '../components/Filters/CustomRuleModal';

interface DnsConfig {
    custom_domain_rules?: CustomRule[];
    [key: string]: any;
}

interface UseDnsRoutingCustomRulesProps {
    dnsConfig?: DnsConfig;
    setDnsConfig: (config: DnsConfig) => Promise<void>;
    getDnsConfig: (...args: unknown[]) => unknown;
    addSuccessToast: (message: string) => unknown;
    addErrorToast: (error: any) => unknown;
    t: (key: string) => string;
}

export const useDnsRoutingCustomRules = ({
    dnsConfig,
    setDnsConfig,
    getDnsConfig,
    addSuccessToast,
    addErrorToast,
    t,
}: UseDnsRoutingCustomRulesProps) => {
    const [customRules, setCustomRules] = useState<CustomRule[]>([]);
    const [isCustomRuleModalOpen, setIsCustomRuleModalOpen] = useState(false);
    const [editingRule, setEditingRule] = useState<CustomRule | null>(null);

    // Load custom rules from DNS config
    useEffect(() => {
        if (dnsConfig?.custom_domain_rules) {
            setCustomRules(dnsConfig.custom_domain_rules);
        }
    }, [dnsConfig]);

    const openCustomRuleModal = useCallback(() => {
        setEditingRule(null);
        setIsCustomRuleModalOpen(true);
    }, []);

    const closeCustomRuleModal = useCallback(() => {
        setIsCustomRuleModalOpen(false);
        setEditingRule(null);
    }, []);

    const handleCustomRuleSubmit = useCallback(async (rule: CustomRule) => {
        try {
            let updatedRules: CustomRule[];
            
            if (editingRule) {
                // Update existing rule
                updatedRules = customRules.map(r => 
                    r === editingRule ? rule : r
                );
            } else {
                // Add new rule
                updatedRules = [...customRules, rule];
            }
            
            const newConfig: DnsConfig = {
                ...dnsConfig,
                custom_domain_rules: updatedRules,
            };
            
            await setDnsConfig(newConfig);
            await getDnsConfig();
            
            setEditingRule(null);
            setIsCustomRuleModalOpen(false);
            addSuccessToast(t('custom_rule_saved'));
        } catch (error) {
            addErrorToast({ error });
        }
    }, [customRules, editingRule, dnsConfig, setDnsConfig, getDnsConfig, addSuccessToast, addErrorToast, t]);

    const handleEditCustomRule = useCallback((rule: CustomRule) => {
        setEditingRule(rule);
        setIsCustomRuleModalOpen(true);
    }, []);

    const handleToggleCustomRule = useCallback(async (rule: CustomRule) => {
        try {
            const updatedRules = customRules.map(r => 
                r === rule ? { ...r, enabled: !r.enabled } : r
            );
            
            const newConfig: DnsConfig = {
                ...dnsConfig,
                custom_domain_rules: updatedRules,
            };
            
            await setDnsConfig(newConfig);
            await getDnsConfig();
            
            setCustomRules(updatedRules);
            addSuccessToast(t('custom_rule_saved'));
        } catch (error) {
            addErrorToast({ error });
        }
    }, [customRules, dnsConfig, setDnsConfig, getDnsConfig, addSuccessToast, addErrorToast, t]);

    const handleDeleteCustomRule = useCallback(async (rule: CustomRule) => {
        if (window.confirm(t('list_confirm_delete'))) {
            try {
                const updatedRules = customRules.filter(r => r !== rule);
                
                const newConfig: DnsConfig = {
                    ...dnsConfig,
                    custom_domain_rules: updatedRules,
                };
                
                await setDnsConfig(newConfig);
                await getDnsConfig();
                
                setCustomRules(updatedRules);
                addSuccessToast(t('custom_rule_deleted'));
            } catch (error) {
                addErrorToast({ error });
            }
        }
    }, [customRules, dnsConfig, setDnsConfig, getDnsConfig, addSuccessToast, addErrorToast, t]);

    return {
        customRules,
        isCustomRuleModalOpen,
        editingRule,
        openCustomRuleModal,
        closeCustomRuleModal,
        handleCustomRuleSubmit,
        handleEditCustomRule,
        handleToggleCustomRule,
        handleDeleteCustomRule,
    };
};
