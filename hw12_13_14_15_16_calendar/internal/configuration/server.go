package configuration

type ServerConf struct {
	Type string `mapstructure:"type"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}
