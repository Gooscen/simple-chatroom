package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// 完整的配置结构
type Config struct {
	AI       AIConfig       `yaml:"ai"`
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
}

// AI配置结构
type AIConfig struct {
	Provider  string `yaml:"provider"`
	APIKey    string `yaml:"api_key"`
	BaseURL   string `yaml:"base_url"`
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens"`
	Timeout   int    `yaml:"timeout"`
}

// 数据库配置结构
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// 服务器配置结构
type ServerConfig struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	Debug bool   `yaml:"debug"`
}

var globalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	// 尝试读取配置文件
	configFile := "config/config.yml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// 如果配置文件不存在，使用默认配置
		fmt.Println("配置文件不存在，使用默认配置。请复制 config.example.yml 为 config.yml 并配置您的API密钥。")
		globalConfig = getDefaultConfig()
		return globalConfig, nil
	}

	// 读取YAML文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 如果配置文件中某些值为空，使用环境变量或默认值
	fillConfigFromEnv(&config)

	globalConfig = &config
	fmt.Printf("配置加载成功，AI提供商: %s\n", config.AI.Provider)
	return globalConfig, nil
}

// 从环境变量填充配置
func fillConfigFromEnv(config *Config) {
	if config.AI.APIKey == "" || config.AI.APIKey == "your-api-key-here" {
		if envKey := os.Getenv("AI_API_KEY"); envKey != "" {
			config.AI.APIKey = envKey
		}
	}

	if config.AI.Provider == "" {
		config.AI.Provider = os.Getenv("AI_PROVIDER")
	}

	if config.AI.BaseURL == "" {
		config.AI.BaseURL = os.Getenv("AI_BASE_URL")
	}

	if config.AI.Model == "" {
		config.AI.Model = os.Getenv("AI_MODEL")
	}

	if config.AI.MaxTokens == 0 {
		config.AI.MaxTokens = 150
	}

	if config.AI.Timeout == 0 {
		config.AI.Timeout = 30
	}
}

// 获取默认配置
func getDefaultConfig() *Config {
	return &Config{
		AI: AIConfig{
			Provider:  "deepseek",
			APIKey:    "",
			BaseURL:   "https://api.deepseek.com",
			Model:     "deepseek-chat",
			MaxTokens: 150,
			Timeout:   30,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "chatroom",
		},
		Server: ServerConfig{
			Host:  "localhost",
			Port:  8081,
			Debug: true,
		},
	}
}

// GetAIConfig 获取AI配置（保持向后兼容）
func GetAIConfig() AIConfig {
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("加载配置失败，使用默认配置: %v\n", err)
		return getDefaultConfig().AI
	}
	return config.AI
}

// 获取环境变量，如果不存在则使用默认值
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// IsAIConfigured 检查AI服务是否已配置
func IsAIConfigured() bool {
	config := GetAIConfig()
	return config.APIKey != "" && config.APIKey != "your-openai-api-key-here" && config.APIKey != "your-api-key-here"
}

// GetTimeout 获取超时时间
func GetTimeout() time.Duration {
	config := GetAIConfig()
	return time.Duration(config.Timeout) * time.Second
}
