package configuration

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Storage StorageConf `mapstructure:"storage"`
	Server  ServerConf  `mapstructure:"server"`
}

func NewConfigFrom(filePath string) Config {
	pathInfo := getConfigPathInfoFor(filePath)

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
