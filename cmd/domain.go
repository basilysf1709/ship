package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

func newDomainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain",
		Short: "Configure domains and TLS via Caddy",
	}
	cmd.AddCommand(newDomainSetupCommand())
	return cmd
}

func newDomainSetupCommand() *cobra.Command {
	var domains []string
	var appPort int

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure Caddy for domains and TLS",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectConfig, err := loadProjectConfig()
			if err != nil {
				return err
			}

			proxy := shipinternal.ProxyConfig{}
			if projectConfig.Proxy != nil {
				proxy = *projectConfig.Proxy
			}
			if len(domains) > 0 {
				proxy.Domains = domains
			}
			if appPort > 0 {
				proxy.AppPort = appPort
			}
			if !proxy.HasDomains() {
				return fmt.Errorf("no domains configured; use --domain or add proxy.domains to ship.json")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
			defer cancel()
			state, client, err := currentServerClient(ctx, 30*time.Second)
			if err != nil {
				return err
			}
			defer client.Close()

			if err := configureProxy(ctx, client, proxy); err != nil {
				return err
			}
			if err := shipinternal.SaveProxyRuntimeConfig(proxy); err != nil {
				return err
			}
			return writeCommandOutput(cmd, fmt.Sprintf("STATUS=DOMAIN_CONFIGURED\nSERVER_IP=%s\n", state.IP), map[string]any{
				"status":    "DOMAIN_CONFIGURED",
				"server_ip": state.IP,
				"domains":   proxy.Domains,
				"app_port":  proxy.EffectiveAppPort(),
			})
		},
	}
	cmd.Flags().StringSliceVar(&domains, "domain", nil, "Domain names to configure")
	cmd.Flags().IntVar(&appPort, "app-port", 0, "Local upstream port for reverse proxy")
	return cmd
}
