package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type PingConfig struct {
	Size     int           `yaml:"size"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

const (
	online  = "online"
	offline = "offline"
)

type Host struct {
	Id         int    `json:"id"`
	Name       string `yaml:"name" json:"name"`
	Addr       string `yaml:"addr" json:"addr"`
	Status     string `json:"status"`
	PingConfig `yaml:",inline" json:"-"`
}

type AppConfig struct {
	PageTitle  string `yaml:"pageTitle"`
	ListenHost string `yaml:"listenHost"`
	ListenPort int    `yaml:"listenPort"`
	PingConfig `yaml:",inline"`
	Hosts      []Host `yaml:"hosts,flow"`
}

func (a *AppConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%d", a.ListenHost, a.ListenPort)
}

func (a *AppConfig) MakeFullPath(path string, protocol string) string {
	return fmt.Sprintf("%s://%s%s", protocol, a.ListenAddr(), path)
}

func getRawConfig(filename string) ([]byte, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func getConfig(rawConfig []byte) (AppConfig, error) {
	config := AppConfig{
		PageTitle:  "pingovalka",
		ListenHost: "localhost",
		ListenPort: 8000,
		PingConfig: PingConfig{Size: 64, Interval: time.Second, Timeout: time.Second},
	}
	err := yaml.Unmarshal(rawConfig, &config)
	if err != nil {
		return config, err
	}

	for i, h := range config.Hosts {
		config.Hosts[i].Id = i
		config.Hosts[i].Status = offline
		if h.Size == 0 {
			config.Hosts[i].Size = config.Size
		}
		if h.Interval == time.Second*0 {
			config.Hosts[i].Interval = config.Interval
		}
		if h.Timeout == time.Second*0 {
			config.Hosts[i].Timeout = config.Timeout
		}
	}

	return config, nil

}

func getConfigFileName() (string, error) {
	if len(os.Args) < 2 {
		err := fmt.Errorf("Usage:\n%s CONFIG_FILE", os.Args[0])
		return "", err
	}
	return os.Args[1], nil
}
