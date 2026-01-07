package configuration

type ServerConf struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}
