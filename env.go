package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AddEnvToFile adds environment variables to .env file and ensures .gitignore is properly configured
// envPath and gitignorePath are optional, defaulting to ".env" and ".gitignore" respectively
func AddEnvToFile(key string, value string, paths ...string) error {
	envPath := ".env"
	gitignorePath := ".gitignore"

	if len(paths) > 0 {
		envPath = paths[0]
	}
	if len(paths) > 1 {
		gitignorePath = paths[1]
	}

	envVars, err := readEnvFile(envPath)
	if err != nil {
		return err
	}

	// Add/Update new environment variable
	envVars[key] = value

	if err := writeEnvFile(envPath, envVars); err != nil {
		return err
	}

	return handleGitignore(gitignorePath, envPath)
}

// readEnvFile reads and parses an existing .env file
func readEnvFile(path string) (map[string]string, error) {
	envVars := make(map[string]string)

	if data, err := os.ReadFile(path); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	}
	return envVars, nil
}

// writeEnvFile writes environment variables to a file
func writeEnvFile(path string, envVars map[string]string) error {
	var envContent strings.Builder
	for k, v := range envVars {
		envContent.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	return os.WriteFile(path, []byte(envContent.String()), 0644)
}

// handleGitignore ensures .env is in .gitignore
func handleGitignore(gitignorePath string, envPath string) error {
	var gitignoreContent []string
	envLineExists := false
	envFileName := filepath.Base(envPath)

	// Read existing .gitignore if it exists
	if data, err := os.ReadFile(gitignorePath); err == nil {
		gitignoreContent = strings.Split(string(data), "\n")
		for _, line := range gitignoreContent {
			if strings.TrimSpace(line) == envFileName {
				envLineExists = true
				break
			}
		}
	}

	// Add .env to .gitignore if not present
	if !envLineExists {
		// Remove empty lines at the end
		for len(gitignoreContent) > 0 && gitignoreContent[len(gitignoreContent)-1] == "" {
			gitignoreContent = gitignoreContent[:len(gitignoreContent)-1]
		}
		gitignoreContent = append(gitignoreContent, envFileName)

		if err := os.WriteFile(gitignorePath, []byte(strings.Join(gitignoreContent, "\n")+"\n"), 0644); err != nil {
			return fmt.Errorf("error writing .gitignore file: %v", err)
		}
	}
	return nil
}
