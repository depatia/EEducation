package config

import "github.com/spf13/viper"

type Config struct {
	Port         int    `mapstructure:"PORT"`
	DBUrl        string `mapstructure:"DB_URL"`
	JWTSecretKey string `mapstructure:"JWT_SECRET_KEY"`
}

func LoadConfig() (cfg *Config, err error) {
	viper.AddConfigPath("../config/envs")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)

	return
}
