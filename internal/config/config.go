package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	configFileName = "config"
	configFileExt  = "yaml"
	configFilePath = "."
)

func ReadConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName(configFileName)
	v.AddConfigPath(configFilePath)
	v.SetConfigType(configFileExt)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file, %s", err)
	}

	v.AutomaticEnv()

	return v, nil
}
