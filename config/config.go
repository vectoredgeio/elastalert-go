package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	EsHost         string `yaml:"es_host"`
	EsPort         int    `yaml:"es_port"`
	RunEvery       string `yaml:"run_every"`
	BufferTime     string `yaml:"buffer_time"`
	WritebackIndex string `yaml:"writeback_index"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
