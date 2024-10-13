package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	JwtPrivateKey string `mapstructure:"JWT_PRIVATE_KEY"`
	JwtPublicKey  string `mapstructure:"JWT_PUBLIC_KEY"`
	DbUsername    string `mapstructure:"DB_USERNAME"`
	DbPassword    string `mapstructure:"DB_PASSWORD"`
	DbHost        string `mapstructure:"DB_HOST"`
	DbPort        string `mapstructure:"DB_PORT"`
	DbDatabase    string `mapstructure:"DB_DATABASE"`
}

var AppConfig Config

func LoadEnv(configPath string) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		return
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		fmt.Println(err)
		return
	}
}
