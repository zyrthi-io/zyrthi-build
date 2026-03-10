package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config zyrthi.yaml 配置结构
type Config struct {
	Platform string         `yaml:"platform"`
	Chip     string         `yaml:"chip"`
	Compiler CompilerConfig `yaml:"compiler"`
	Flash    FlashConfig    `yaml:"flash"`
	Project  ProjectConfig  `yaml:"project"`
}

// CompilerConfig 编译器配置
type CompilerConfig struct {
	Prefix   string   `yaml:"prefix"`
	Cflags   []string `yaml:"cflags"`
	Ldflags  []string `yaml:"ldflags"`
	Includes []string `yaml:"includes"`
}

// FlashConfig 烧录配置
type FlashConfig struct {
	Plugin      string `yaml:"plugin"`
	EntryAddr   string `yaml:"entry_addr"`
	FlashSize   string `yaml:"flash_size"`
	DefaultBaud int    `yaml:"default_baud"`
}

// ProjectConfig 项目配置
type ProjectConfig struct {
	Name    string   `yaml:"name"`
	Sources []string `yaml:"sources"`
}

// CompileCommand compile_commands.json 条目
type CompileCommand struct {
	Directory string `json:"directory"`
	Command   string `json:"command"`
	File      string `json:"file"`
}

// Load 从 zyrthi.yaml 加载配置
func Load(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	if cfg.Project.Name == "" {
		cfg.Project.Name = "firmware"
	}
	if len(cfg.Project.Sources) == 0 {
		cfg.Project.Sources = []string{"src/"}
	}

	return &cfg, nil
}

// WriteCompileCommands 生成 compile_commands.json
func WriteCompileCommands(commands []CompileCommand) error {
	cwd, _ := os.Getwd()
	for i := range commands {
		commands[i].Directory = cwd
	}

	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("compile_commands.json", data, 0644)
}