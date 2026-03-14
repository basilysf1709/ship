package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	var template string
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a starter ship.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				if _, err := os.Stat("ship.json"); err == nil {
					return fmt.Errorf("ship.json already exists; use --force to overwrite")
				}
			}
			if template == "" {
				template = detectInitTemplate()
			}
			content, err := renderInitTemplate(template)
			if err != nil {
				return err
			}
			if err := os.WriteFile("ship.json", []byte(content), 0o644); err != nil {
				return fmt.Errorf("write ship.json: %w", err)
			}
			return writeCommandOutput(cmd, fmt.Sprintf("STATUS=INIT_COMPLETE\nTEMPLATE=%s\n", template), map[string]any{
				"status":   "INIT_COMPLETE",
				"template": template,
			})
		},
	}
	cmd.Flags().StringVar(&template, "template", "", "Template: docker, node, go, static")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing ship.json")
	return cmd
}

func detectInitTemplate() string {
	if _, err := os.Stat("Dockerfile"); err == nil {
		return "docker"
	}
	if _, err := os.Stat("package.json"); err == nil {
		return "node"
	}
	if _, err := os.Stat("go.mod"); err == nil {
		return "go"
	}
	return "static"
}

func renderInitTemplate(template string) (string, error) {
	switch template {
	case "docker":
		return "{\n  \"status\": {\n    \"healthcheck_path\": \"/\"\n  }\n}\n", nil
	case "node":
		return "{\n  \"deploy\": {\n    \"local_commands\": [\n      \"npm ci\",\n      \"npm run build\",\n      \"tar -czf release.tar.gz dist package.json package-lock.json\"\n    ],\n    \"uploads\": [\n      {\n        \"source\": \"release.tar.gz\",\n        \"destination\": \"/opt/app/release.tar.gz\",\n        \"mode\": \"0644\"\n      }\n    ],\n    \"remote_commands\": [\n      \"mkdir -p /opt/app\",\n      \"cd /opt/app && tar -xzf release.tar.gz\",\n      \"cd /opt/app && npm ci --omit=dev\",\n      \"pkill -f 'npm start' || true\",\n      \"cd /opt/app && nohup npm start >/var/log/app.log 2>&1 &\"\n    ],\n    \"cleanup_local\": [\n      \"release.tar.gz\"\n    ]\n  },\n  \"status\": {\n    \"healthcheck_path\": \"/\"\n  }\n}\n", nil
	case "go":
		return "{\n  \"deploy\": {\n    \"local_commands\": [\n      \"go build -o app .\"\n    ],\n    \"uploads\": [\n      {\n        \"source\": \"app\",\n        \"destination\": \"/opt/app/app\",\n        \"mode\": \"0755\"\n      }\n    ],\n    \"remote_commands\": [\n      \"mkdir -p /opt/app\",\n      \"pkill -f '/opt/app/app' || true\",\n      \"nohup /opt/app/app >/var/log/app.log 2>&1 &\"\n    ],\n    \"cleanup_local\": [\n      \"app\"\n    ]\n  },\n  \"status\": {\n    \"healthcheck_path\": \"/\"\n  }\n}\n", nil
	case "static":
		return "{\n  \"deploy\": {\n    \"local_commands\": [\n      \"tar -czf site.tar.gz dist\"\n    ],\n    \"uploads\": [\n      {\n        \"source\": \"site.tar.gz\",\n        \"destination\": \"/var/www/site.tar.gz\",\n        \"mode\": \"0644\"\n      }\n    ],\n    \"remote_commands\": [\n      \"mkdir -p /var/www\",\n      \"cd /var/www && tar -xzf site.tar.gz\"\n    ],\n    \"cleanup_local\": [\n      \"site.tar.gz\"\n    ]\n  }\n}\n", nil
	default:
		return "", fmt.Errorf("unsupported template %q", template)
	}
}
