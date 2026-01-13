package configuration

import (
	"errors"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Logger    LoggerConf    `mapstructure:"logger"`
	Storage   StorageConf   `mapstructure:"storage"`
	Server    ServerConf    `mapstructure:"server,omitempty"`
	Broker    BrokerConf    `mapstructure:"broker,omitempty"`
	Scheduler SchedulerConf `mapstructure:"scheduler,omitempty"`
	Sender    SenderConf    `mapstructure:"sender,omitempty"`
}

func NewConfigFrom(filePath string) Config {
	pathInfo := getConfigPathInfoFor(filePath)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName(pathInfo.Name)
	viper.AddConfigPath(pathInfo.Path)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Println("config file not found, using defaults")
		} else {
			log.Fatal("error reading config:", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("unable to decode config into struct:", err)
	}

	return config
}
