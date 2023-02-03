package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/comfyprog/pingovalka/frontend"
	"gopkg.in/yaml.v3"
)

type PingConfig struct {
	Size     int           `yaml:"size"`
	Interval time.Duration `yaml:"interval"`
}

type Host struct {
	Name       string `yaml:"name"`
	Addr       string `yaml:"addr"`
	PingConfig `yaml:",inline"`
}

type AppConfig struct {
	ListenHost string `yaml:"listenHost"`
	ListenPort int    `yaml:"listenPort"`
	PingConfig `yaml:",inline"`
	Hosts      []Host `yaml:"hosts,flow"`
}

func (a *AppConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%d", a.ListenHost, a.ListenPort)
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
		ListenHost: "localhost",
		ListenPort: 8000,
		PingConfig: PingConfig{Size: 64, Interval: time.Second},
	}
	err := yaml.Unmarshal(rawConfig, &config)
	if err != nil {
		return config, err
	}

	for i, h := range config.Hosts {
		if h.Size == 0 {
			config.Hosts[i].Size = config.Size
		}
		if h.Interval == time.Second*0 {
			config.Hosts[i].Interval = config.Interval
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

func main() {
	configFilename, err := getConfigFileName()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rawConfig, err := getRawConfig(configFilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	config, err := getConfig(rawConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(config)

	frontendFs, err := fs.Sub(frontend.FrontendFs, "dist")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	http.Handle("/", http.FileServer(http.FS(frontendFs)))

	log.Fatal(http.ListenAndServe(config.ListenAddr(), nil))
}