package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Server Server `yaml:"server"`
	Proxy  Proxy  `yaml:"proxy"`
}

type Server struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}
type Proxy struct {
	Address string `yaml:"addr"`
}

var envRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}

	// 先做环境变量替换，再反序列化
	data = []byte(envRegex.ReplaceAllStringFunc(string(data), func(match string) string {
		// match 形如 ${SERVER_PORT:3200}
		inner := match[2 : len(match)-1] // "SERVER_PORT:3200"
		parts := splitEnvVar(inner)      // ["SERVER_PORT", "3200"]
		if val, ok := os.LookupEnv(parts[0]); ok {
			return val
		}
		// 没有设置环境变量且有默认值，返回默认值
		if len(parts) > 1 {
			return parts[1]
		}
		return ""
	}))

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config file failed: %w", err)
	}

	return &cfg, nil
}

// splitEnvVar 解析 "KEY:default" 或 "KEY"
func splitEnvVar(s string) []string {
	// 找第一个冒号，但要注意 URL 里也有冒号（如 http://...）
	// 所以只取第一个冒号之前的作为 key
	idx := -1
	for i, c := range s {
		if c == ':' {
			idx = i
			break
		}
	}
	if idx == -1 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+1:]}
}
