package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadOrCreateUserConfig loads user configuration from the given path or creates a default one
func LoadOrCreateUserConfig(configPath string) (*UserConfig, error) {
	// Try to read existing config
	data, err := os.ReadFile(configPath)
	if err == nil {
		var config UserConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("parsing user config: %w", err)
		}
		return &config, nil
	}

	// If file doesn't exist, create default config
	if os.IsNotExist(err) {
		config := UserConfig{
			APIConfigPath: "apis.yaml", // Default in current directory
		}

		data, err := yaml.Marshal(&config)
		if err != nil {
			return nil, fmt.Errorf("serializing default config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return nil, fmt.Errorf("writing default config: %w", err)
		}

		return &config, nil
	}

	return nil, fmt.Errorf("reading user config: %w", err)
}

// LoadConfig loads API configuration from the given file path
func LoadConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &config, nil
}

// InitUserConfigDir ensures the user configuration directory exists
func InitUserConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home directory: %w", err)
	}

	apiDir := filepath.Join(homeDir, ".api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		return "", fmt.Errorf("creating directory: %w", err)
	}

	return apiDir, nil
}

// CollectModuleInfo collects all parameters and request configurations from the module chain
func CollectModuleInfo(modules map[string]Module, path []string, apiName string) ([]ParamDef, []RequestConfig, *APISpec, error) {
	if len(path) == 0 {
		return nil, nil, nil, fmt.Errorf("invalid module path")
	}

	currentName := path[0]
	currentModule, ok := modules[currentName]
	if !ok {
		return nil, nil, nil, fmt.Errorf("module '%s' not found", currentName)
	}

	var params []ParamDef
	params = append(params, currentModule.Params...)

	var reqs []RequestConfig
	if currentModule.Request != nil {
		reqs = append(reqs, *currentModule.Request)
	}

	// If there are submodules to traverse
	if len(path) > 1 {
		if currentModule.Modules == nil {
			return nil, nil, nil, fmt.Errorf("submodule '%s' not found in module '%s'", path[1], currentName)
		}
		childParams, childReqs, apiSpec, err := CollectModuleInfo(currentModule.Modules, path[1:], apiName)
		if err != nil {
			return nil, nil, nil, err
		}
		params = append(params, childParams...)
		reqs = append(reqs, childReqs...)
		return params, reqs, apiSpec, nil
	}

	// Find target API in current module
	if apiSpec, ok := currentModule.APIs[apiName]; ok {
		return params, reqs, &apiSpec, nil
	}

	return nil, nil, nil, fmt.Errorf("API '%s' not found in module '%s'", apiName, currentName)
}

// MergeRequestConfigs merges all request configurations in the module chain
func MergeRequestConfigs(moduleReqs []RequestConfig, apiReq *RequestSpec) *RequestSpec {
	mergedReq := *apiReq
	mergedHeaders := make(map[string]string)

	// Apply module-level headers in order
	for _, req := range moduleReqs {
		for k, v := range req.Headers {
			mergedHeaders[k] = v
		}
	}

	// Apply API-specific headers (highest priority)
	if apiReq.Headers != nil {
		for k, v := range apiReq.Headers {
			mergedHeaders[k] = v
		}
	}

	mergedReq.Headers = mergedHeaders
	return &mergedReq
}
