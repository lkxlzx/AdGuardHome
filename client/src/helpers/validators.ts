import i18next from 'i18next';

import {
    MAX_PORT,
    R_CIDR,
    R_CIDR_IPV6,
    R_HOST,
    R_IPV4,
    R_IPV6,
    R_MAC,
    R_URL_REQUIRES_PROTOCOL,
    STANDARD_WEB_PORT,
    UNSAFE_PORTS,
    R_CLIENT_ID,
    R_DOMAIN,
    MAX_PASSWORD_LENGTH,
    MIN_PASSWORD_LENGTH,
    R_IPV4_SUBNET,
    R_IPV6_SUBNET,
} from './constants';

import { ip4ToInt, isValidAbsolutePath } from './form';

import { isIpInCidr, parseSubnetMask } from './helpers';

// Validation functions
// If the value is valid, the validation function should return undefined.
/**
 * @param value {string|number}
 * @returns {undefined|string}
 */
export const validateRequiredValue = (value: any) => {
    const formattedValue = typeof value === 'string' ? value.trim() : value;
    if (formattedValue || formattedValue === 0 || (formattedValue && formattedValue.length !== 0)) {
        return undefined;
    }
    return i18next.t('form_error_required');
};

/**
 * @returns {undefined|string}
 * @param _
 * @param allValues
 */
export const validateIpv4RangeEnd = (_: any, allValues: any) => {
    if (!allValues || !allValues.v4 || !allValues.v4.range_end || !allValues.v4.range_start) {
        return undefined;
    }

    const { range_end, range_start } = allValues.v4;

    if (ip4ToInt(range_end) <= ip4ToInt(range_start)) {
        return i18next.t('greater_range_start_error');
    }

    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateIpv4 = (value: any) => {
    if (value && !R_IPV4.test(value)) {
        return i18next.t('form_error_ip4_format');
    }
    return undefined;
};

/**
 * @returns {undefined|string}
 * @param _
 * @param allValues
 */
export const validateNotInRange = (value: any, allValues: any) => {
    if (!allValues.v4) {
        return undefined;
    }

    const { range_start, range_end } = allValues.v4;

    if (range_start && validateIpv4(range_start)) {
        return undefined;
    }

    if (range_end && validateIpv4(range_end)) {
        return undefined;
    }

    const isAboveMin = range_start && ip4ToInt(value) >= ip4ToInt(range_start);
    const isBelowMax = range_end && ip4ToInt(value) <= ip4ToInt(range_end);

    if (isAboveMin && isBelowMax) {
        return i18next.t('out_of_range_error', {
            start: range_start,
            end: range_end,
        });
    }

    return undefined;
};

/**
 * @returns {undefined|string}
 * @param _
 * @param allValues
 */
export const validateGatewaySubnetMask = (_: any, allValues: any) => {
    if (!allValues || !allValues.v4 || !allValues.v4.subnet_mask || !allValues.v4.gateway_ip) {
        return i18next.t('gateway_or_subnet_invalid');
    }

    const { subnet_mask, gateway_ip } = allValues.v4;

    if (validateIpv4(gateway_ip)) {
        return i18next.t('gateway_or_subnet_invalid');
    }

    return parseSubnetMask(subnet_mask) ? undefined : i18next.t('gateway_or_subnet_invalid');
};

/**
 * @returns {undefined|string}
 * @param value
 * @param allValues
 */
export const validateIpForGatewaySubnetMask = (value: any, allValues: any) => {
    if (!allValues || !allValues.v4 || !value || !allValues.gateway_ip || !allValues.subnet_mask) {
        return undefined;
    }

    const { gateway_ip, subnet_mask } = allValues.v4;

    if ((gateway_ip && validateIpv4(gateway_ip)) || (subnet_mask && validateIpv4(subnet_mask))) {
        return undefined;
    }

    const subnetPrefix = parseSubnetMask(subnet_mask);

    if (!isIpInCidr(value, `${gateway_ip}/${subnetPrefix}`)) {
        return i18next.t('subnet_error');
    }

    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateClientId = (value: string) => {
    if (!value) {
        return undefined;
    }
    const formattedValue = value.trim();
    if (
        formattedValue &&
        !(
            R_IPV4.test(formattedValue) ||
            R_IPV6.test(formattedValue) ||
            R_MAC.test(formattedValue) ||
            R_CIDR.test(formattedValue) ||
            R_CIDR_IPV6.test(formattedValue) ||
            R_CLIENT_ID.test(formattedValue)
        )
    ) {
        return i18next.t('form_error_client_id_format');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateConfigClientId = (value: any) => {
    if (!value) {
        return undefined;
    }
    const formattedValue = value.trim();
    if (formattedValue && !R_CLIENT_ID.test(formattedValue)) {
        return i18next.t('form_error_client_id_format');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateServerName = (value: any) => {
    if (!value) {
        return undefined;
    }
    const formattedValue = value ? value.trim() : value;
    if (formattedValue && !R_DOMAIN.test(formattedValue)) {
        return i18next.t('form_error_server_name');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateIpv6 = (value: any) => {
    if (value && !R_IPV6.test(value)) {
        return i18next.t('form_error_ip6_format');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateIp = (value: any) => {
    if (value && !R_IPV4.test(value) && !R_IPV6.test(value)) {
        return i18next.t('form_error_ip_format');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateMac = (value: any) => {
    if (value && !R_MAC.test(value)) {
        return i18next.t('form_error_mac_format');
    }
    return undefined;
};

/**
 * @param value {number}
 * @returns {undefined|string}
 */
export const validatePort = (value: any) => {
    if ((value || value === 0) && (value < STANDARD_WEB_PORT || value > MAX_PORT)) {
        return i18next.t('form_error_port_range');
    }
    return undefined;
};

/**
 * @param value {number}
 * @returns {undefined|string}
 */
export const validateInstallPort = (value: any) => {
    if (value < 1 || value > MAX_PORT) {
        return i18next.t('form_error_port');
    }
    return undefined;
};

/**
 * @param value {number}
 * @returns {undefined|string}
 */
export const validatePortTLS = (value: any) => {
    if (value === 0) {
        return undefined;
    }
    if (value && (value < STANDARD_WEB_PORT || value > MAX_PORT)) {
        return i18next.t('form_error_port_range');
    }
    return undefined;
};

/**
 * @param value {number}
 * @returns {undefined|string}
 */
export const validatePortQuic = validatePortTLS;

/**
 * @param value {number}
 * @returns {undefined|string}
 */
export const validateIsSafePort = (value: any) => {
    if (UNSAFE_PORTS.includes(value)) {
        return i18next.t('form_error_port_unsafe');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateDomain = (value: any) => {
    if (value && !R_HOST.test(value)) {
        return i18next.t('form_error_domain_format');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateAnswer = (value: any) => {
    if (value && !R_IPV4.test(value) && !R_IPV6.test(value) && !R_HOST.test(value)) {
        return i18next.t('form_error_answer_format');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {undefined|string}
 */
export const validatePath = (value: any) => {
    if (value && !isValidAbsolutePath(value) && !R_URL_REQUIRES_PROTOCOL.test(value)) {
        return i18next.t('form_error_url_or_path_format');
    }
    return undefined;
};

/**
 * @param cidr {string}
 * @returns {Function}
 */
export const validateIpv4InCidr = (valueIp: any, allValues: any) => {
    if (!isIpInCidr(valueIp, allValues.cidr)) {
        return i18next.t('form_error_subnet', { ip: valueIp, cidr: allValues.cidr });
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {number}
 */
const utf8StringLength = (value: any) => {
    const encoder = new TextEncoder();
    const view = encoder.encode(value);

    return view.length;
};

/**
 * @param value {string}
 * @returns {Function}
 */
export const validatePasswordLength = (value: any) => {
    if (value) {
        const length = utf8StringLength(value);
        if (length < MIN_PASSWORD_LENGTH || length > MAX_PASSWORD_LENGTH) {
            // TODO: Make the i18n clearer with regards to bytes vs. characters.
            return i18next.t('form_error_password_length', {
                min: MIN_PASSWORD_LENGTH,
                max: MAX_PASSWORD_LENGTH,
            });
        }
    }

    return undefined;
};

/**
 * @param value {string}
 * @returns {Function}
 */
export const validateIpGateway = (value: any, allValues: any) => {
    if (value === allValues.gatewayIp) {
        return i18next.t('form_error_gateway_ip');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {Function}
 */
export const validateIPv4Subnet = (value: any) => {
    if (!R_IPV4_SUBNET.test(value)) {
        return i18next.t('rate_limit_subnet_len_ipv4_error');
    }
    return undefined;
};

/**
 * @param value {string}
 * @returns {Function}
 */
export const validateIPv6Subnet = (value: any) => {
    if (!R_IPV6_SUBNET.test(value)) {
        return i18next.t('rate_limit_subnet_len_ipv6_error');
    }
    return undefined;
};

/**
 * @returns {undefined|string}
 * @param value
 * @param allValues
 */
export const validatePlainDns = (value: any, allValues: any) => {
    const { enabled } = allValues;

    if (!enabled && !value) {
        return i18next.t('encryption_plain_dns_error');
    }

    return undefined;
};

/**
 * Validates DNS upstream servers list
 * Supports: IP, IP:port, [ipv6], [ipv6]:port, https://, tls://, quic://, sdns://
 * @param value {string}
 * @returns {undefined|string}
 */
export const validateUpstreamServers = (value: any) => {
    if (!value) {
        return undefined;
    }

    const lines = value.split('\n').map((line: string) => line.trim()).filter((line: string) => line.length > 0);
    
    if (lines.length === 0) {
        return undefined;
    }

    // Regex patterns for different upstream formats
    const patterns = {
        // Plain IP: 8.8.8.8
        plainIp: /^(\d{1,3}\.){3}\d{1,3}$/,
        // IP with port: 8.8.8.8:53
        ipWithPort: /^(\d{1,3}\.){3}\d{1,3}:\d{1,5}$/,
        // IPv6: [2001:4860:4860::8888]
        ipv6: /^\[([0-9a-fA-F:]+)\]$/,
        // IPv6 with port: [2001:4860:4860::8888]:53
        ipv6WithPort: /^\[([0-9a-fA-F:]+)\]:\d{1,5}$/,
        // Protocol-based: https://, tls://, quic://, sdns://
        protocol: /^(https?|tls|quic|sdns):\/\/.+/,
    };

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        
        // Skip comments
        if (line.startsWith('#')) {
            continue;
        }

        let isValid = false;

        // Check protocol-based formats first
        if (patterns.protocol.test(line)) {
            isValid = true;
        }
        // Check IPv6 formats
        else if (patterns.ipv6.test(line) || patterns.ipv6WithPort.test(line)) {
            isValid = true;
        }
        // Check IPv4 formats
        else if (patterns.plainIp.test(line) || patterns.ipWithPort.test(line)) {
            // Validate IP address octets
            const ipMatch = line.match(/^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})/);
            if (ipMatch) {
                const octets = [ipMatch[1], ipMatch[2], ipMatch[3], ipMatch[4]];
                const validOctets = octets.every((octet) => {
                    const num = parseInt(octet, 10);
                    return num >= 0 && num <= 255;
                });
                
                if (validOctets) {
                    // Validate port if present
                    const portMatch = line.match(/:(\d{1,5})$/);
                    if (portMatch) {
                        const port = parseInt(portMatch[1], 10);
                        isValid = port >= 1 && port <= 65535;
                    } else {
                        isValid = true;
                    }
                }
            }
        }

        if (!isValid) {
            return i18next.t('form_error_upstream_format', { line: i + 1, value: line });
        }
    }

    return undefined;
};
