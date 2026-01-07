package configuration

type StorageConf struct {
	Type     string `mapstructure:"type"`
	Database DbConf `mapstructure:"database"`
}
