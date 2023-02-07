package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

type PingConfig struct {
	Size     int           `yaml:"size"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

const (
	online   = "online"
	offline  = "offline"
	progname = "pingovalka"
)

type Host struct {
	Id         int    `json:"id"`
	Name       string `yaml:"name" json:"name"`
	Addr       string `yaml:"addr" json:"addr"`
	Status     string `json:"status"`
	PingConfig `yaml:",inline" json:"-"`
}

type BasicAuthCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type AppConfig struct {
	PageTitle  string `yaml:"pageTitle"`
	ListenHost string `yaml:"listenHost"`
	ListenPort int    `yaml:"listenPort"`
	PingConfig `yaml:",inline"`
	Hosts      []Host                 `yaml:"hosts,flow"`
	BasicAuth  []BasicAuthCredentials `yaml:"basicAuth,flow"`
}

func (a *AppConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%d", a.ListenHost, a.ListenPort)
}

func (a *AppConfig) MakeFullPath(path string, protocol string) string {
	return fmt.Sprintf("%s://%s%s", protocol, a.ListenAddr(), path)
}

func (a *AppConfig) HasBasicAuthConfigured() bool {
	return len(a.BasicAuth) > 0
}

func getFlags(args []string) (filename string, showVersion bool, output string, err error) {
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)
	var buf bytes.Buffer
	flags.SetOutput(&buf)

	fileUsage := "path to config file"
	flags.StringVar(&filename, "config", "config.yml", fileUsage)
	flags.StringVar(&filename, "c", "config.yml", fileUsage+" (shorthand)")

	versionUsage := "show program version"
	flags.BoolVar(&showVersion, "version", false, versionUsage)
	flags.BoolVar(&showVersion, "v", false, versionUsage+" (shorthand)")

	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage of %s:\n", progname)
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), "\nProgram uses the github.com/go-ping/ping library "+
			"that attempts to send an \"unprivileged\" ping via UDP.\n"+
			"On Linux, this must be enabled with the following sysctl command:\n"+
			"\tsudo sysctl -w net.ipv4.ping_group_range=\"0 2147483647\"\n")
	}

	err = flags.Parse(args)
	output = buf.String()

	return
}

func getVersionString(version string) string {
	return fmt.Sprintf("%s v%s built with %s", progname, version, runtime.Version())
}

func getRawConfig(filename string) ([]byte, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func parseConfig(rawConfig []byte) (AppConfig, error) {
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
