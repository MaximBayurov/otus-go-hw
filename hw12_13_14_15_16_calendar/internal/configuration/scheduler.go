package configuration

import "time"

type SchedulerConf struct {
	Interval         int64 `mapstructure:"interval"`
	CleanupInterval  int64 `mapstructure:"cleanup_interval"`
	BatchSize        int   `mapstructure:"batch_size"`
	CleanupThreshold int64 `mapstructure:"cleanup_threshold"`
}

func (sc *SchedulerConf) GetInterval() time.Duration {
	return time.Duration(sc.Interval) * time.Minute
}

func (sc *SchedulerConf) GetCleanupInterval() time.Duration {
	return time.Duration(sc.CleanupInterval) * time.Hour
}

func (sc *SchedulerConf) GetCleanupThreshold() time.Duration {
	return time.Duration(sc.CleanupThreshold) * time.Hour
}
