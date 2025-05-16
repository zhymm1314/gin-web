package config

type ApiUrlConfig struct {
	Urls map[string]any
}

type ApiEnv struct {
	Dev  string `mapstructure:"dev" json:"dev" yaml:"dev"`
	Pre  string `mapstructure:"pre" json:"pre" yaml:"pre"`
	Prod string `mapstructure:"prod" json:"prod" yaml:"prod"`
}
