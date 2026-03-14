package shipinternal

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const secretsFile = "secrets.env"
const remoteSecretsPath = "/root/.ship/secrets.env"

var secretKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func secretsPath() string {
	return filepath.Join(shipDir, secretsFile)
}

func LoadSecrets() (map[string]string, error) {
	if err := validateSecretsFilePermissions(); err != nil {
		return nil, err
	}

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
		if err := ValidateSecretKey(parts[0]); err != nil {
			return nil, err
		}
		if err := ValidateSecretValue(parts[1]); err != nil {
			return nil, err
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
		if err := ValidateSecretKey(key); err != nil {
			return err
		}
		if err := ValidateSecretValue(secrets[key]); err != nil {
			return err
		}
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
	tempPath := secretsPath() + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", tempPath, err)
	}
	if err := os.Rename(tempPath, secretsPath()); err != nil {
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

func ValidateSecretKey(key string) error {
	if !secretKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid secret key %q: expected [A-Za-z_][A-Za-z0-9_]*", key)
	}
	return nil
}

func ValidateSecretValue(value string) error {
	if strings.Contains(value, "\x00") {
		return fmt.Errorf("invalid secret value: NUL bytes are not supported")
	}
	if strings.Contains(value, "\n") {
		return fmt.Errorf("invalid secret value: multiline values are not supported in %s", secretsPath())
	}
	return nil
}

func SecretsChecksum() (string, error) {
	secrets, err := LoadSecrets()
	if err != nil {
		return "", err
	}
	if len(secrets) == 0 {
		return "", nil
	}

	keys := make([]string, 0, len(secrets))
	for key := range secrets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	hash := sha256.New()
	for _, key := range keys {
		_, _ = hash.Write([]byte(key))
		_, _ = hash.Write([]byte("="))
		_, _ = hash.Write([]byte(secrets[key]))
		_, _ = hash.Write([]byte("\n"))
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func validateSecretsFilePermissions() error {
	info, err := os.Stat(secretsPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat %s: %w", secretsPath(), err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%s must not be group/world accessible; run chmod 600 %s", secretsPath(), secretsPath())
	}
	return nil
}
