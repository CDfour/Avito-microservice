package config

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

type config struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
}

func LoadConfig() (*config, error) {

	logrus.Info("Starting loading config")

	config := &config{}

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		logrus.Fatalln("ReadFile: ", err)
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		logrus.Fatalln("Unmarshal: ", err)
	}

	if config.Username == "" {
		return nil, ErrNoUsername
	}
	if config.Password == "" {
		return nil, ErrNoPassword
	}
	if config.Host == "" {
		return nil, ErrNoHost
	}
	if config.Port == "" {
		return nil, ErrNoPort
	}
	if config.Database == "" {
		return nil, ErrNoDatabase
	}

	logrus.Info("Ending loading config")

	return config, nil
}
