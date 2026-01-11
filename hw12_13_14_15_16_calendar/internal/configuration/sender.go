package configuration

import "time"

type SenderConf struct {
	Workers       int   `mapstructure:"workers"`
	RetryInterval int64 `mapstructure:"retry_interval"`
	MaxRetries    int   `mapstructure:"max_retries"`
}

func (sc *SenderConf) GetRetryInterval() time.Duration {
	return time.Duration(sc.RetryInterval) * time.Second
}
