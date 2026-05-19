package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	OpenAI OpenAIConfig `yaml:"openai"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type OpenAIConfig struct {
	APIKey           string  `yaml:"api_key"`
	BaseURL          string  `yaml:"base_url"`
	Model            string  `yaml:"model"`
	MaxTokens        int     `yaml:"max_tokens"`
	Temperature      float32 `yaml:"temperature"`
	TopP             float32 `yaml:"top_p"`
	PresencePenalty  float32 `yaml:"presence_penalty"`
	FrequencyPenalty float32 `yaml:"frequency_penalty"`
}

var (
	cfg  *Config
	once sync.Once
)

// Load 加载配置文件（线程安全，单例模式）
func Load(path string) (*Config, error) {
	var err error
	once.Do(func() {
		cfg = &Config{}
		var data []byte
		data, err = os.ReadFile(path)
		if err != nil {
			return
		}
		err = yaml.Unmarshal(data, cfg)
	})
	return cfg, err
}

// Get 获取已加载的配置
func Get() *Config {
	return cfg
}
