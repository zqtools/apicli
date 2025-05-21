package config

// Config represents the main API configuration
type Config struct {
	Modules map[string]Module `yaml:"modules"`
}

// UserConfig represents user-specific configuration
type UserConfig struct {
	APIConfigPath string `yaml:"api_config_path"`
}

// Module represents an API module configuration
type Module struct {
	Description string            `yaml:"description"`
	Params      []ParamDef       `yaml:"params,omitempty"`
	Request     *RequestConfig   `yaml:"request,omitempty"`
	Modules     map[string]Module `yaml:"modules,omitempty"`
	APIs        map[string]APISpec `yaml:"apis,omitempty"`
}

// APISpec represents an API specification
type APISpec struct {
	Params  []ParamDef  `yaml:"params"`
	Request RequestSpec `yaml:"request"`
}

// ParamDef represents a parameter definition
type ParamDef struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Description string `yaml:"description"`
}

// RequestConfig represents module-level request configuration
type RequestConfig struct {
	Headers map[string]string `yaml:"headers,omitempty"`
}

// RequestSpec represents API-specific request configuration
type RequestSpec struct {
	Method   string            `yaml:"method"`
	URL      string            `yaml:"url"`
	Body     string            `yaml:"body,omitempty"`
	BodyFile string           `yaml:"body_file,omitempty"`
	Form     map[string]string `yaml:"form,omitempty"`
	Params   []QueryParam     `yaml:"params,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty"`
}

// QueryParam represents a URL query parameter
type QueryParam struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
