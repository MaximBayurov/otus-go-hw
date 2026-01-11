package configuration

import "time"

type DbConf struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	DbName          string        `mapstructure:"dbname"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"pass"`
	MaxIdleConn     int           `mapstructure:"conn.max_idle"`
	MaxOpenConn     int           `mapstructure:"conn.max_open"`
	MaxLifetimeConn time.Duration `mapstructure:"conn.lifetime"`
}
