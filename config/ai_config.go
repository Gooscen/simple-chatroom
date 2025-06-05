package config

import (
	"os"
)

// AI配置结构
type AIConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

// 获取AI配置
func GetAIConfig() AIConfig {
	config := AIConfig{
		Provider: getEnvWithDefault("AI_PROVIDER", "openai"),
		APIKey:   getEnvWithDefault("AI_API_KEY", ""),
		BaseURL:  getEnvWithDefault("AI_BASE_URL", "https://api.openai.com/v1"),
		Model:    getEnvWithDefault("AI_MODEL", "gpt-3.5-turbo"),
	}

	// 兼容旧的环境变量名
	if config.APIKey == "" {
		config.APIKey = os.Getenv("OPENAI_API_KEY")
	}

	return config
}

// 获取环境变量，如果不存在则使用默认值
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// 检查AI服务是否已配置
func IsAIConfigured() bool {
	config := GetAIConfig()
	return config.APIKey != "" && config.APIKey != "your-api-key-here"
}
