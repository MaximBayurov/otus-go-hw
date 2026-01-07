package configuration

type LoggerConf struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}
