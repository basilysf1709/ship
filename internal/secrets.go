package shipinternal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const secretsFile = "secrets.env"
const remoteSecretsPath = "/root/.ship/secrets.env"

func secretsPath() string {
	return filepath.Join(shipDir, secretsFile)
}

func LoadSecrets() (map[string]string, error) {
	data, err := os.ReadFile(secretsPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", secretsPath(), err)
	}

	secrets := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("parse %s: invalid line %q", secretsPath(), line)
		}
		secrets[parts[0]] = parts[1]
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", secretsPath(), err)
	}
	return secrets, nil
}

func SaveSecrets(secrets map[string]string) error {
	if err := os.MkdirAll(shipDir, 0o755); err != nil {
		return fmt.Errorf("create .ship directory: %w", err)
	}

	keys := make([]string, 0, len(secrets))
	for key := range secrets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var lines []string
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s=%s", key, secrets[key]))
	}
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	if err := os.WriteFile(secretsPath(), []byte(content), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", secretsPath(), err)
	}
	return nil
}

func DeleteSecrets() error {
	if err := os.Remove(secretsPath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", secretsPath(), err)
	}
	return nil
}

func HasLocalSecrets() bool {
	_, err := os.Stat(secretsPath())
	return err == nil
}
