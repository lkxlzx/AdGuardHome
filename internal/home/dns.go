package home

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/agh"
	"github.com/AdguardTeam/AdGuardHome/internal/aghalg"
	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/aghnet"
	"github.com/AdguardTeam/AdGuardHome/internal/client"
	"github.com/AdguardTeam/AdGuardHome/internal/dnsforward"
	"github.com/AdguardTeam/AdGuardHome/internal/dnsrouting"
	"github.com/AdguardTeam/AdGuardHome/internal/dnsroutingfiles"
	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/AdGuardHome/internal/querylog"
	"github.com/AdguardTeam/AdGuardHome/internal/stats"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/log"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/netutil/urlutil"
	"github.com/ameshkov/dnscrypt/v2"
	yaml "go.yaml.in/yaml/v4"
)

// Default listening ports.
const (
	defaultPortDNS   uint16 = 53
	defaultPortHTTP  uint16 = 80
	defaultPortHTTPS uint16 = 443
	defaultPortQUIC  uint16 = 853
	defaultPortTLS   uint16 = 853
)

// initDNS updates all the fields of the [globalContext] needed to initialize
// the DNS server and initializes it at last.  It also must not be called unless
// [config] and [globalContext] are initialized.  baseLogger, tlsMgr,
// confModifier, and httpReg must not be nil.
func initDNS(
	ctx context.Context,
	baseLogger *slog.Logger,
	tlsMgr *tlsManager,
	confModifier agh.ConfigModifier,
	httpReg aghhttp.Registrar,
	statsDir string,
	querylogDir string,
	workDir string,
) (err error) {
	// Ensure a default upstream group exists before initializing DNS server
	ensureDefaultGroup(ctx, baseLogger.With(slogutil.KeyPrefix, "upstream_groups"))

	anonymizer := config.anonymizer()

	statsConf := stats.Config{
		Logger:            baseLogger.With(slogutil.KeyPrefix, "stats"),
		Filename:          filepath.Join(statsDir, "stats.db"),
		Limit:             time.Duration(config.Stats.Interval),
		ConfigModifier:    confModifier,
		HTTPReg:           httpReg,
		Enabled:           config.Stats.Enabled,
		ShouldCountClient: globalContext.clients.shouldCountClient,
	}

	engine, err := aghnet.NewIgnoreEngine(config.Stats.Ignored)
	if err != nil {
		return fmt.Errorf("statistics: ignored list: %w", err)
	}

	statsConf.Ignored = engine
	globalContext.stats, err = stats.New(statsConf)
	if err != nil {
		return fmt.Errorf("init stats: %w", err)
	}

	conf := querylog.Config{
		Logger:            baseLogger.With(slogutil.KeyPrefix, "querylog"),
		Anonymizer:        anonymizer,
		ConfigModifier:    confModifier,
		HTTPReg:           httpReg,
		FindClient:        globalContext.clients.findMultiple,
		BaseDir:           querylogDir,
		AnonymizeClientIP: config.DNS.AnonymizeClientIP,
		RotationIvl:       time.Duration(config.QueryLog.Interval),
		MemSize:           config.QueryLog.MemSize,
		Enabled:           config.QueryLog.Enabled,
		FileEnabled:       config.QueryLog.FileEnabled,
	}

	engine, err = aghnet.NewIgnoreEngine(config.QueryLog.Ignored)
	if err != nil {
		return fmt.Errorf("querylog: ignored list: %w", err)
	}

	conf.Ignored = engine
	globalContext.queryLog, err = querylog.New(conf)
	if err != nil {
		return fmt.Errorf("init querylog: %w", err)
	}

	// Set up DNS routing rules update callback
	config.Filtering.OnDnsRoutingRulesUpdated = func(filterID int64, upstreamGroup string, priority int, rules []interface{}) {
		if globalContext.dnsServer != nil {
			globalContext.dnsServer.UpdateDnsRoutingRules(filterID, upstreamGroup, priority, rules)
		}
	}
	
	globalContext.filters, err = filtering.New(config.Filtering, nil)
	if err != nil {
		// Don't wrap the error, since it's informative enough as is.
		return err
	}

	// Initialize DNS routing file manager
	err = initDnsRoutingFileManager(ctx, baseLogger, workDir, confModifier)
	if err != nil {
		return fmt.Errorf("init dns routing file manager: %w", err)
	}

	return initDNSServer(
		ctx,
		globalContext.filters,
		globalContext.stats,
		globalContext.queryLog,
		globalContext.dhcpServer,
		anonymizer,
		httpReg,
		tlsMgr,
		baseLogger,
		confModifier,
	)
}

// initDNSServer initializes the [context.dnsServer].  To only use the internal
// proxy, none of the arguments are required, but tlsMgr and l still must not be
// nil, in other cases all the arguments also must not be nil.  It also must not
// be called unless [config] and [globalContext] are initialized.
//
// TODO(e.burkov): Use [dnsforward.DNSCreateParams] as a parameter.
func initDNSServer(
	ctx context.Context,
	filters *filtering.DNSFilter,
	sts stats.Interface,
	qlog querylog.QueryLog,
	dhcpSrv dnsforward.DHCP,
	anonymizer *aghnet.IPMut,
	httpReg aghhttp.Registrar,
	tlsMgr *tlsManager,
	l *slog.Logger,
	confModifier agh.ConfigModifier,
) (err error) {
	globalContext.dnsServer, err = dnsforward.NewServer(dnsforward.DNSCreateParams{
		Logger:      l,
		DNSFilter:   filters,
		Stats:       sts,
		QueryLog:    qlog,
		PrivateNets: parseSubnetSet(config.DNS.PrivateNets),
		Anonymizer:  anonymizer,
		DHCPServer:  dhcpSrv,
		EtcHosts:    globalContext.etcHosts,
		LocalDomain: config.DHCP.LocalDomainName,
	})
	defer func() {
		if err != nil {
			closeDNSServer(ctx)
		}
	}()
	if err != nil {
		return fmt.Errorf("dnsforward.NewServer: %w", err)
	}

	globalContext.clients.clientChecker = globalContext.dnsServer

	dnsConf, err := newServerConfig(
		&config.DNS,
		config.Clients.Sources,
		tlsMgr.config(),
		tlsMgr,
		httpReg,
		globalContext.clients.storage,
		confModifier,
	)
	if err != nil {
		return fmt.Errorf("newServerConfig: %w", err)
	}

	// Try to prepare the server with disabled private RDNS resolution if it
	// failed to prepare as is.  See TODO on [dnsforward.PrivateRDNSError].
	err = globalContext.dnsServer.Prepare(ctx, dnsConf)
	if privRDNSErr := (&dnsforward.PrivateRDNSError{}); errors.As(err, &privRDNSErr) {
		l.WarnContext(ctx, "private rdns resolution failed; disabling", slogutil.KeyError, err)

		dnsConf.UsePrivateRDNS = false
		err = globalContext.dnsServer.Prepare(ctx, dnsConf)
	}

	if err != nil {
		return fmt.Errorf("dnsServer.Prepare: %w", err)
	}

	return nil
}

// parseSubnetSet parses a slice of subnets.  If the slice is empty, it returns
// a subnet set that matches all locally served networks, see
// [netutil.IsLocallyServed].
func parseSubnetSet(nets []netutil.Prefix) (s netutil.SubnetSet) {
	switch len(nets) {
	case 0:
		// Use an optimized function-based matcher.
		return netutil.SubnetSetFunc(netutil.IsLocallyServed)
	case 1:
		return nets[0].Prefix
	default:
		return netutil.SliceSubnetSet(netutil.UnembedPrefixes(nets))
	}
}

func isRunning() bool {
	return globalContext.dnsServer != nil && globalContext.dnsServer.IsRunning()
}

func ipsToTCPAddrs(ips []netip.Addr, port uint16) (tcpAddrs []*net.TCPAddr) {
	if ips == nil {
		return nil
	}

	tcpAddrs = make([]*net.TCPAddr, 0, len(ips))
	for _, ip := range ips {
		tcpAddrs = append(tcpAddrs, net.TCPAddrFromAddrPort(netip.AddrPortFrom(ip, port)))
	}

	return tcpAddrs
}

func ipsToUDPAddrs(ips []netip.Addr, port uint16) (udpAddrs []*net.UDPAddr) {
	if ips == nil {
		return nil
	}

	udpAddrs = make([]*net.UDPAddr, 0, len(ips))
	for _, ip := range ips {
		udpAddrs = append(udpAddrs, net.UDPAddrFromAddrPort(netip.AddrPortFrom(ip, port)))
	}

	return udpAddrs
}

// newServerConfig converts values from the configuration file into the internal
// DNS server configuration.  All arguments must not be nil.
func newServerConfig(
	dnsConf *dnsConfig,
	clientSrcConf *clientSourcesConfig,
	tlsConf *tlsConfigSettings,
	tlsMgr *tlsManager,
	httpReg aghhttp.Registrar,
	clientsContainer dnsforward.ClientsContainer,
	confModifier agh.ConfigModifier,
) (newConf *dnsforward.ServerConfig, err error) {
	hosts := aghalg.CoalesceSlice(dnsConf.BindHosts, []netip.Addr{netutil.IPv4Localhost()})

	fwdConf := dnsConf.Config
	
	// Use default upstream group if configured
	for _, group := range config.DNS.UpstreamGroups {
		if group.IsDefault && group.Enabled {
			fwdConf.UpstreamDNS = group.UpstreamDNS
			if len(group.FallbackDNS) > 0 {
				fwdConf.FallbackDNS = group.FallbackDNS
			}
			if len(group.BootstrapDNS) > 0 {
				fwdConf.BootstrapDNS = group.BootstrapDNS
			}
			break
		}
	}
	
	// Set up custom domain rules callbacks
	fwdConf.CustomDomainRulesGetter = func() []dnsforward.CustomDomainRuleConfig {
		config.RLock()
		defer config.RUnlock()
		
		rules := make([]dnsforward.CustomDomainRuleConfig, len(config.DNS.CustomDomainRules))
		for i, rule := range config.DNS.CustomDomainRules {
			rules[i] = dnsforward.CustomDomainRuleConfig{
				Domain:        rule.Domain,
				MatchType:     rule.MatchType,
				UpstreamGroup: rule.UpstreamGroup,
				Enabled:       rule.Enabled,
			}
		}
		return rules
	}
	
	fwdConf.CustomDomainRulesSetter = func(rules []dnsforward.CustomDomainRuleConfig) {
		config.Lock()
		defer config.Unlock()
		
		config.DNS.CustomDomainRules = make([]CustomDomainRule, len(rules))
		for i, rule := range rules {
			config.DNS.CustomDomainRules[i] = CustomDomainRule{
				Domain:        rule.Domain,
				MatchType:     rule.MatchType,
				UpstreamGroup: rule.UpstreamGroup,
				Enabled:       rule.Enabled,
			}
		}
	}
	
	// Set up upstream group getter
	fwdConf.UpstreamGroupGetter = func(groupID string) *dnsforward.UpstreamGroupConfig {
		config.RLock()
		defer config.RUnlock()
		
		for _, group := range config.DNS.UpstreamGroups {
			if group.ID == groupID && group.Enabled {
				return &dnsforward.UpstreamGroupConfig{
					ID:           group.ID,
					Name:         group.Name,
					Enabled:      group.Enabled,
					UpstreamDNS:  group.UpstreamDNS,
					FallbackDNS:  group.FallbackDNS,
					BootstrapDNS: group.BootstrapDNS,
				}
			}
		}
		return nil
	}
	
	// Set up DNS routing rules getter
	fwdConf.DnsRoutingRulesGetter = func() []dnsforward.DnsRoutingRuleConfig {
		config.RLock()
		defer config.RUnlock()
		
		// Get DNS routing rules from Filters array (marked with DnsRouting=true)
		var rules []dnsforward.DnsRoutingRuleConfig
		for _, filter := range config.Filters {
			if filter.DnsRouting && filter.Enabled {
				rules = append(rules, dnsforward.DnsRoutingRuleConfig{
					ID:            int64(filter.ID),
					Enabled:       filter.Enabled,
					URL:           filter.URL,
					Name:          filter.Name,
					UpstreamGroup: filter.UpstreamGroup,
					Priority:      filter.Priority,
					RulesCount:    filter.RulesCount,
					LastUpdated:   filter.LastUpdated.Format(time.RFC3339),
				})
			}
		}
		
		// Also get from DnsRoutingRules array if it exists
		for _, rule := range config.DnsRoutingRules {
			if rule.Enabled {
				rules = append(rules, dnsforward.DnsRoutingRuleConfig{
					ID:            rule.ID,
					Enabled:       rule.Enabled,
					URL:           rule.URL,
					Name:          rule.Name,
					UpstreamGroup: rule.UpstreamGroup,
					Priority:      rule.Priority,
					RulesCount:    rule.RulesCount,
					LastUpdated:   rule.LastUpdated,
				})
			}
		}
		
		return rules
	}
	
	fwdConf.ClientsContainer = clientsContainer

	intTLSConf, err := newDNSTLSConfig(tlsConf, hosts)
	if err != nil {
		return nil, fmt.Errorf("constructing tls config: %w", err)
	}

	newConf = &dnsforward.ServerConfig{
		UDPListenAddrs:         ipsToUDPAddrs(hosts, dnsConf.Port),
		TCPListenAddrs:         ipsToTCPAddrs(hosts, dnsConf.Port),
		Config:                 fwdConf,
		TLSConf:                intTLSConf,
		TLSAllowUnencryptedDoH: tlsConf.AllowUnencryptedDoH,
		UpstreamTimeout:        time.Duration(dnsConf.UpstreamTimeout),
		TLSv12Roots:            tlsMgr.rootCerts,
		ConfModifier:           confModifier,
		HTTPReg:                httpReg,
		LocalPTRResolvers:      dnsConf.PrivateRDNSResolvers,
		UseDNS64:               dnsConf.UseDNS64,
		DNS64Prefixes:          dnsConf.DNS64Prefixes,
		UsePrivateRDNS:         dnsConf.UsePrivateRDNS,
		ServeHTTP3:             dnsConf.ServeHTTP3,
		UseHTTP3Upstreams:      dnsConf.UseHTTP3Upstreams,
		ServePlainDNS:          dnsConf.ServePlainDNS,
		PendingRequestsEnabled: dnsConf.PendingRequests.Enabled,
	}

	var initialAddresses []netip.Addr
	// Context.stats may be nil here if initDNSServer is called from
	// [cmdlineUpdate].
	if sts := globalContext.stats; sts != nil {
		const initialClientsNum = 100
		initialAddresses = globalContext.stats.TopClientsIP(initialClientsNum)
	}

	// Do not set DialContext, PrivateSubnets, and UsePrivateRDNS, because they
	// are set by [dnsforward.Server.Prepare].
	newConf.AddrProcConf = &client.DefaultAddrProcConfig{
		Exchanger:        globalContext.dnsServer,
		AddressUpdater:   &globalContext.clients,
		InitialAddresses: initialAddresses,
		CatchPanics:      true,
		UseRDNS:          clientSrcConf.RDNS,
		UseWHOIS:         clientSrcConf.WHOIS,
	}

	return newConf, nil
}

// newDNSTLSConfig converts values from the configuration file into the internal
// TLS settings for the DNS server.  conf must not be nil.
func newDNSTLSConfig(
	conf *tlsConfigSettings,
	addrs []netip.Addr,
) (dnsConf *dnsforward.TLSConfig, err error) {
	if !conf.Enabled {
		return &dnsforward.TLSConfig{}, nil
	}

	dnsCryptConf, err := newDNSCryptConfig(conf, addrs)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		return nil, err
	}

	dnsConf = &dnsforward.TLSConfig{
		DNSCryptConf:   dnsCryptConf,
		ServerName:     conf.ServerName,
		StrictSNICheck: conf.StrictSNICheck,
	}

	if conf.PortHTTPS != 0 {
		dnsConf.HTTPSListenAddrs = ipsToTCPAddrs(addrs, conf.PortHTTPS)
	}

	if conf.PortDNSOverTLS != 0 {
		dnsConf.TLSListenAddrs = ipsToTCPAddrs(addrs, conf.PortDNSOverTLS)
	}

	if conf.PortDNSOverQUIC != 0 {
		dnsConf.QUICListenAddrs = ipsToUDPAddrs(addrs, conf.PortDNSOverQUIC)
	}

	cert, err := tls.X509KeyPair(conf.CertificateChainData, conf.PrivateKeyData)
	if err != nil {
		err = fmt.Errorf("parsing tls key pair: %w", err)
		if conf.AllowUnencryptedDoH || dnsCryptConf != nil {
			// TODO(s.chzhen):  Use [slog.Logger].
			log.Info("warning: %s", err)

			return dnsConf, nil
		}

		// Don't wrap the error, because it's already annotated.
		return nil, err
	}

	dnsConf.Cert = &cert

	return dnsConf, nil
}

// newDNSCryptConfig converts values from the configuration file into the
// internal DNSCrypt settings for the DNS server.  conf must not be nil.
func newDNSCryptConfig(
	conf *tlsConfigSettings,
	addrs []netip.Addr,
) (dnsCryptConf *dnsforward.DNSCryptConfig, err error) {
	if conf.PortDNSCrypt == 0 {
		return nil, nil
	}

	if conf.DNSCryptConfigFile == "" {
		return nil, fmt.Errorf("dnscrypt_config_file: %w", errors.ErrEmptyValue)
	}

	f, err := os.Open(conf.DNSCryptConfigFile)
	if err != nil {
		return nil, fmt.Errorf("opening dnscrypt config: %w", err)
	}
	defer func() { err = errors.WithDeferred(err, f.Close()) }()

	rc := &dnscrypt.ResolverConfig{}
	err = yaml.NewDecoder(f).Decode(rc)
	if err != nil {
		return nil, fmt.Errorf("decoding dnscrypt config: %w", err)
	}

	cert, err := rc.CreateCert()
	if err != nil {
		return nil, fmt.Errorf("creating dnscrypt cert: %w", err)
	}

	return &dnsforward.DNSCryptConfig{
		ResolverCert:   cert,
		UDPListenAddrs: ipsToUDPAddrs(addrs, conf.PortDNSCrypt),
		TCPListenAddrs: ipsToTCPAddrs(addrs, conf.PortDNSCrypt),
		ProviderName:   rc.ProviderName,
	}, nil
}

// dnsEncryption contains different types of TLS encryption addresses.
type dnsEncryption struct {
	https string
	tls   string
	quic  string
}

// getDNSEncryption returns the TLS encryption addresses that AdGuard Home
// listens on.  tlsMgr must not be nil.
func getDNSEncryption(tlsMgr *tlsManager) (de dnsEncryption) {
	tlsConf := tlsMgr.config()

	if !tlsConf.Enabled || len(tlsConf.ServerName) == 0 {
		return dnsEncryption{}
	}

	hostname := tlsConf.ServerName
	if tlsConf.PortHTTPS != 0 {
		addr := hostname
		if p := tlsConf.PortHTTPS; p != defaultPortHTTPS {
			addr = netutil.JoinHostPort(addr, p)
		}

		de.https = (&url.URL{
			Scheme: urlutil.SchemeHTTPS,
			Host:   addr,
			Path:   "/dns-query",
		}).String()
	}

	if p := tlsConf.PortDNSOverTLS; p != 0 {
		de.tls = (&url.URL{
			Scheme: "tls",
			Host:   netutil.JoinHostPort(hostname, p),
		}).String()
	}

	if p := tlsConf.PortDNSOverQUIC; p != 0 {
		de.quic = (&url.URL{
			Scheme: "quic",
			Host:   netutil.JoinHostPort(hostname, p),
		}).String()
	}

	return de
}

func startDNSServer() error {
	config.RLock()
	defer config.RUnlock()

	if isRunning() {
		return fmt.Errorf("unable to start forwarding DNS server: Already running")
	}

	globalContext.filters.EnableFilters(false)

	// TODO(s.chzhen):  Pass context.
	ctx := context.TODO()
	err := globalContext.clients.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting clients container: %w", err)
	}

	err = globalContext.dnsServer.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting dns server: %w", err)
	}

	globalContext.filters.Start()
	globalContext.stats.Start()

	err = globalContext.queryLog.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting query log: %w", err)
	}

	return nil
}

func stopDNSServer(ctx context.Context) (err error) {
	if !isRunning() {
		return nil
	}

	err = globalContext.dnsServer.Stop(ctx)
	if err != nil {
		return fmt.Errorf("stopping forwarding dns server: %w", err)
	}

	err = globalContext.clients.close(ctx)
	if err != nil {
		return fmt.Errorf("closing clients container: %w", err)
	}

	closeDNSServer(ctx)

	return nil
}

func closeDNSServer(ctx context.Context) {
	// DNS forward module must be closed BEFORE stats or queryLog because it depends on them
	if globalContext.dnsServer != nil {
		globalContext.dnsServer.Close(ctx)
		globalContext.dnsServer = nil
	}

	if globalContext.filters != nil {
		globalContext.filters.Close()
	}

	if globalContext.dnsRoutingFileManager != nil {
		err := globalContext.dnsRoutingFileManager.Close()
		if err != nil {
			log.Error("closing dns routing file manager: %s", err)
		}
	}

	if globalContext.stats != nil {
		err := globalContext.stats.Close()
		if err != nil {
			log.Error("closing stats: %s", err)
		}
	}

	if globalContext.queryLog != nil {
		err := globalContext.queryLog.Shutdown(ctx)
		if err != nil {
			log.Error("closing query log: %s", err)
		}
	}

	log.Debug("all dns modules are closed")
}

// checkStatsAndQuerylogDirs checks and returns directory paths to store
// statistics and query log.
func checkStatsAndQuerylogDirs(
	conf *configuration,
	workDir string,
) (statsDir, querylogDir string, err error) {
	baseDir := filepath.Join(workDir, dataDir)

	statsDir = conf.Stats.DirPath
	if statsDir == "" {
		statsDir = baseDir
	} else {
		err = checkDir(statsDir)
		if err != nil {
			return "", "", fmt.Errorf("statistics: custom directory: %w", err)
		}
	}

	querylogDir = conf.QueryLog.DirPath
	if querylogDir == "" {
		querylogDir = baseDir
	} else {
		err = checkDir(querylogDir)
		if err != nil {
			return "", "", fmt.Errorf("querylog: custom directory: %w", err)
		}
	}

	return statsDir, querylogDir, nil
}

// checkDir checks if the path is a directory.  It's used to check for
// misconfiguration at startup.
func checkDir(path string) (err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(path); err != nil {
		// Don't wrap the error, since it's informative enough as is.
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("%q is not a directory", path)
	}

	return nil
}

// initDnsRoutingFileManager initializes the DNS routing file manager.
func initDnsRoutingFileManager(ctx context.Context, baseLogger *slog.Logger, workDir string, confModifier agh.ConfigModifier) (err error) {
	if workDir == "" {
		workDir = "."
	}

	fmConfig := dnsroutingfiles.Config{
		DataDir: workDir,
		Logger:  baseLogger.With(slogutil.KeyPrefix, "dns_routing_files"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		OnRouterUpdate: func(ctx context.Context) error {
			// Reload DNS router when rules change
			// This callback is called WITHOUT holding any File Manager locks
			if globalContext.dnsServer != nil {
				// Use a goroutine to avoid blocking and potential deadlocks
				go func() {
					defer func() {
						if r := recover(); r != nil {
							baseLogger.ErrorContext(ctx, "panic in router reload", "panic", r)
						}
					}()
					reloadDnsRoutingRules(ctx, baseLogger)
				}()
			}
			return nil
		},
		GetRuleConfig: func(ruleID int64) *dnsroutingfiles.DomainListRule {
			// Get rule config from the config file
			config.RLock()
			defer config.RUnlock()

			for _, filter := range config.Filtering.Filters {
				if int64(filter.ID) == ruleID && filter.DnsRouting {
					return &dnsroutingfiles.DomainListRule{
						ID:             int64(filter.ID),
						Name:           filter.Name,
						URL:            filter.URL,
						UpstreamGroup:  filter.UpstreamGroup,
						Priority:       filter.Priority,
						Enabled:        filter.Enabled,
						RulesCount:     filter.RulesCount,
						LastUpdated:    filter.LastUpdated,
						UpdateInterval: filter.UpdateInterval,
					}
				}
			}
			return nil
		},
		SaveRuleConfig: func(ctx context.Context, rule *dnsroutingfiles.DomainListRule) error {
			// Update config file with new values
			config.Lock()
			for i := range config.Filtering.Filters {
				if int64(config.Filtering.Filters[i].ID) == rule.ID && config.Filtering.Filters[i].DnsRouting {
					config.Filtering.Filters[i].RulesCount = rule.RulesCount
					config.Filtering.Filters[i].LastUpdated = rule.LastUpdated
					break
				}
			}
			config.Unlock()

			// Save config to disk using confModifier
			confModifier.Apply(ctx)
			return nil
		},
	}

	globalContext.dnsRoutingFileManager, err = dnsroutingfiles.NewManager(fmConfig)
	if err != nil {
		return fmt.Errorf("creating dns routing file manager: %w", err)
	}

	// Load all persisted rules
	if err = globalContext.dnsRoutingFileManager.LoadAll(ctx); err != nil {
		return fmt.Errorf("loading dns routing rules: %w", err)
	}

	// Initialize the atomic ID counter with the maximum existing ID
	config.RLock()
	var maxID int64
	for _, filter := range config.Filtering.Filters {
		if int64(filter.ID) > maxID {
			maxID = int64(filter.ID)
		}
	}
	config.RUnlock()
	nextDnsRoutingRuleID.Store(maxID)

	// Load rules into router (this will be called again after DNS server is initialized)
	// Note: This is safe to call even if dnsServer is nil, it will just return early
	reloadDnsRoutingRules(ctx, baseLogger)

	// Start auto-updates for all rules
	if fm, ok := globalContext.dnsRoutingFileManager.(interface{ StartAutoUpdates() }); ok {
		fm.StartAutoUpdates()
	}

	return nil
}

// reloadDnsRoutingRules reloads all DNS routing rules from File Manager into the router.
func reloadDnsRoutingRules(ctx context.Context, baseLogger *slog.Logger) {
	if globalContext.dnsServer == nil {
		return
	}

	if globalContext.dnsRoutingFileManager == nil {
		return
	}

	baseLogger.InfoContext(ctx, "reloading DNS routing rules from config file")

	// Get all domain list rules from config file (source of truth)
	config.RLock()
	filters := config.Filtering.Filters
	config.RUnlock()

	// Load each domain list rule file and update router
	for _, filter := range filters {
		if !filter.DnsRouting {
			continue
		}
		
		// Handle disabled rules by removing them from router
		if !filter.Enabled {
			// Remove the source from router if it exists
			if globalContext.dnsServer != nil {
				globalContext.dnsServer.RemoveDnsRoutingSource(int64(filter.ID))
			}
			baseLogger.DebugContext(ctx, "removed disabled rule from router",
				"id", filter.ID,
				"name", filter.Name)
			continue
		}

		// Get file path
		filePath := globalContext.dnsRoutingFileManager.GetRuleFilePath(int64(filter.ID))

		// Read rule file (already in AdGuard format)
		data, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				baseLogger.WarnContext(ctx, "rule file does not exist, skipping",
					"id", filter.ID,
					"path", filePath)
			} else {
				baseLogger.ErrorContext(ctx, "failed to read rule file",
					"id", filter.ID,
					"path", filePath,
					"error", err)
			}
			continue
		}

		// Parse AdGuard format rules with improved error handling
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		rulesInterface := make([]interface{}, 0, len(lines))
		invalidCount := 0
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse AdGuard format to determine match type
			var domain string
			var matchType string
			valid := true

			if strings.HasPrefix(line, "||") && strings.HasSuffix(line, "^") {
				// ||example.com^ -> DOMAIN-SUFFIX
				domain = strings.TrimSuffix(strings.TrimPrefix(line, "||"), "^")
				matchType = "DOMAIN-SUFFIX"
			} else if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
				// |example.com| -> DOMAIN
				domain = strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|")
				matchType = "DOMAIN"
			} else if strings.HasPrefix(line, "@@||") {
				// @@||example.com^ -> Allowlist rule (skip for DNS routing)
				continue
			} else if strings.Contains(line, "$") {
				// Rules with modifiers (skip for DNS routing)
				continue
			} else if strings.HasPrefix(line, "/") && strings.HasSuffix(line, "/") {
				// Regex rules (skip for DNS routing)
				continue
			} else {
				// Plain domain or keyword -> DOMAIN-KEYWORD
				domain = line
				matchType = "DOMAIN-KEYWORD"
			}

			// Validate domain is not empty
			if domain == "" {
				invalidCount++
				continue
			}

			if valid {
				// Create ParsedRule
				parsedRule := dnsrouting.ParsedRule{
					Domain:    domain,
					MatchType: matchType,
				}
				rulesInterface = append(rulesInterface, parsedRule)
			}
		}
		
		if invalidCount > 0 {
			baseLogger.DebugContext(ctx, "skipped invalid rules",
				"id", filter.ID,
				"invalid_count", invalidCount)
		}

		// Update router with these rules
		globalContext.dnsServer.UpdateDnsRoutingRules(
			int64(filter.ID),
			filter.UpstreamGroup,
			filter.Priority,
			rulesInterface,
		)

		baseLogger.InfoContext(ctx, "loaded domain list rule into router",
			"id", filter.ID,
			"name", filter.Name,
			"rules_count", len(rulesInterface))
	}

	// Reload custom rules
	globalContext.dnsServer.ReloadDnsRouter(ctx)

	// Count DNS routing rules
	dnsRoutingCount := 0
	for _, filter := range filters {
		if filter.DnsRouting {
			dnsRoutingCount++
		}
	}

	baseLogger.InfoContext(ctx, "DNS routing rules reload complete",
		"domain_list_rules", dnsRoutingCount)
}
