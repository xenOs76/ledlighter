package main

import (
	"errors"
	"fmt"
	"net"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Leds []Led `mapstructure:"leds"`
}

func GetLedsMap(cfg *Config) (map[int]Led, error) {
	ledsMap := make(map[int]Led)

	if len(cfg.Leds) < 1 {
		err := fmt.Errorf("config check: Yaml leds map does not contain any led data: %v", cfg.Leds)
		return nil, errors.New(err.Error())
	}

	var cLedsI interface{} = cfg.Leds

	if _, ok := cLedsI.([]Led); !ok {
		err := fmt.Errorf("config check: wrong type for cfg.Leds. Expected []Led, got %v", cfg.Leds)
		return nil, errors.New(err.Error())
	}

	for _, l := range cfg.Leds {
		validLedKind := false
		validateLedKind := func() {
			for _, v := range LedKinds {
				if l.Kind == v {
					validLedKind = true
				}
			}
		}
		validateLedKind()

		if !validLedKind {
			err := fmt.Errorf("config check: unable to validate Led kind %v for Led id %v", l.Kind, l.Id)
			return nil, errors.New(err.Error())
		}

		if ip := net.ParseIP(l.Address); ip == nil {
			err := fmt.Errorf("config check: unable to validate IP %v for Led id %v", l.Address, l.Id)
			return nil, errors.New(err.Error())
		}

		// TODO: earlier panic on defining a Led with wrongly typed id
		// var lIdInterface interface{} = l.Id
		// if _, ok := lIdInterface.(int); !ok {
		// 	err := fmt.Errorf("config check: unable to validate Id %v for Led", l.Id)
		// 	return nil, errors.New(err.Error())
		// }

		ledsMap[l.Id] = l
	}
	return ledsMap, nil
}

// TODO: add config's hot reload for Docker
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("testdata")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into config struct: %s", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		// log.Print("config change detected")
	})

	return &config, nil
}
