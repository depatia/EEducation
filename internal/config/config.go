package config

import "github.com/spf13/viper"

type Config struct {
	Port    int           `mapstructure:"port"`
	DBPath  string        `mapstructure:"db_path"`
	Clients ClientsConfig `mapstructure:"clients"`
}

type Client struct {
	Addr         string `mapstructure:"addr"`
	Timeout      int    `mapstructure:"timeout"`
	RetriesCount int    `mapstructure:"retries_count"`
}

type ClientsConfig struct {
	Lessons Client `mapstructure:"lessons"`
}

func LoadConfig() (cfg *Config, err error) {
	viper.AddConfigPath("../config")
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()

	if err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)

	return
}
