package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type WatchMan struct {
	Path           string
	StartColumn    int
	TransferMethod string
	WatchType      string
	FileList       []string
}

type UdpInfo struct {
	Host string
	Port int
}

type Record struct {
	File   string
	Column int
}

type Logger struct {
	LogPath     string
	LogLevel    string
	MaxSize     int
	MaxAge      int
	ServiceName string
}

type Config struct {
	Watchman WatchMan
	Udpinfo  UdpInfo
	Records  []Record
	Loginfo  Logger
}

func init() {

}

func GetConfig() Config {

	config := Config{}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found - ", err.Error())
		} else {
			fmt.Println(err.Error())
		}
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("unable to decode into struct, %v", err)
	}

	return config
}

func UpdateConfig() bool {
	err := viper.WriteConfig()
	if err != nil {
		return false
	}
	return true
}
