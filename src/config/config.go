package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type WatchMan struct {
	Path           []string
	StartColumn    int
	Suffix         []string
	TransferMethod string
	WatchAll       bool
	FileList       []string
}

type UdpInfo struct {
	Host string
	Port int
}

type MqttInfo struct {
	Host     string
	Port     int
	Qos      int
	Username string
	Password string
	Pubtopic string
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
	Mqttinfo MqttInfo
	Records  []Record
	Loginfo  Logger
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
	return err == nil
}
