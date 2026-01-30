package helpers

import (
	"bufio"
	"encoding/json"
	"log-analyzer/models"
	"os"
	"path/filepath"
)

func getConfigPath(filename string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(cwd, "config", filename)
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(ex)
	return filepath.Join(dir, "..", "config", filename), nil
}

func ReadRules() ([]models.Rules, error) {
	configPath, err := getConfigPath("rules.json")
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Rules []models.Rules `json:"rules"`
	}

	err = json.Unmarshal(file, &wrapper)
	if err != nil {
		return nil, err
	}

	return wrapper.Rules, nil
}

func ReadPaths() (models.PathConfig, error) {
	configPath, err := getConfigPath("paths.json")
	if err != nil {
		return models.PathConfig{}, err
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return models.PathConfig{}, err
	}

	var config models.PathConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return models.PathConfig{}, err
	}

	configDir := filepath.Dir(configPath)
	for key, logPath := range config.Logs {
		absPath := filepath.Join(configDir, logPath)
		config.Logs[key] = absPath
	}

	return config, nil
}

func ReadLogFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func GetLogNames() ([]string, error) {
	config, err := ReadPaths()
	if err != nil {
		return nil, err
	}

	var names []string
	for name := range config.Logs {
		names = append(names, name)
	}
	return names, nil
}
