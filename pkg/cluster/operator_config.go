package cluster

import (
	"fmt"

	"github.com/spf13/viper"
)

// OperatorConfig defines the operator manager configuration loaded from environment
// variables and flags via Viper.
type OperatorConfig struct {
	MetricsAddr         string `mapstructure:"metrics-bind-address"`
	HealthProbeAddr     string `mapstructure:"health-probe-bind-address"`
	LeaderElection      bool   `mapstructure:"leader-elect"`
	MonitoringNamespace string `mapstructure:"dsc-monitoring-namespace"`
	LogMode             string `mapstructure:"log-mode"`
	PprofAddr           string `mapstructure:"pprof-bind-address"`

	// Zap logging configuration
	ZapDevel        bool   `mapstructure:"zap-devel"`
	ZapEncoder      string `mapstructure:"zap-encoder"`
	ZapLogLevel     string `mapstructure:"zap-log-level"`
	ZapStacktrace   string `mapstructure:"zap-stacktrace-level"`
	ZapTimeEncoding string `mapstructure:"zap-time-encoding"`
}

// LoadConfig loads operator configuration from Viper.
// Viper must be configured with appropriate env prefix and flags before calling this function.
func LoadConfig() (*OperatorConfig, error) {
	var operatorConfig OperatorConfig
	if err := viper.Unmarshal(&operatorConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal operator manager config: %w", err)
	}
	return &operatorConfig, nil
}
