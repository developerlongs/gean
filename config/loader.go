// Package config handles lean-quickstart configuration loading.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// GenesisConfig represents config.yaml from lean-quickstart.
type GenesisConfig struct {
	GenesisTime    uint64 `yaml:"GENESIS_TIME"`
	ValidatorCount uint64 `yaml:"VALIDATOR_COUNT"`
}

// Load loads all configuration from a genesis directory.
func Load(genesisDir, nodeID string) (*GenesisConfig, []uint64, []string, error) {
	cfg, err := loadConfig(genesisDir)
	if err != nil {
		return nil, nil, nil, err
	}

	var validatorIndices []uint64
	if nodeID != "" {
		validatorIndices, err = loadValidatorIndices(genesisDir, nodeID)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	bootnodes, _ := loadBootnodes(genesisDir) // Ignore error, bootnodes optional

	return cfg, validatorIndices, bootnodes, nil
}

func loadConfig(dir string) (*GenesisConfig, error) {
	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read config.yaml: %w", err)
	}
	var cfg GenesisConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config.yaml: %w", err)
	}
	return &cfg, nil
}

func loadValidatorIndices(dir, nodeID string) ([]uint64, error) {
	data, err := os.ReadFile(filepath.Join(dir, "validators.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read validators.yaml: %w", err)
	}
	var assignments map[string][]uint64
	if err := yaml.Unmarshal(data, &assignments); err != nil {
		return nil, fmt.Errorf("parse validators.yaml: %w", err)
	}
	indices, ok := assignments[nodeID]
	if !ok {
		return nil, fmt.Errorf("node %q not found in validators.yaml", nodeID)
	}
	return indices, nil
}

func loadBootnodes(dir string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "nodes.yaml"))
	if err != nil {
		return nil, err
	}
	var nodes []string
	if err := yaml.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}
