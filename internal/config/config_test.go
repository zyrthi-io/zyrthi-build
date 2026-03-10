package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: esp32c3
compiler:
  prefix: riscv32-esp-elf-
  cflags:
    - -Os
    - -g
  ldflags:
    - -nostdlib
  includes:
    - include/
flash:
  plugin: https://example.com/plugin.wasm
  entry_addr: "0x0"
project:
  name: test-project
  sources:
    - src/
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.Platform != "esp32" {
		t.Errorf("expected platform 'esp32', got %s", cfg.Platform)
	}
	if cfg.Chip != "esp32c3" {
		t.Errorf("expected chip 'esp32c3', got %s", cfg.Chip)
	}
	if cfg.Compiler.Prefix != "riscv32-esp-elf-" {
		t.Errorf("expected compiler prefix 'riscv32-esp-elf-', got %s", cfg.Compiler.Prefix)
	}
	if cfg.Project.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got %s", cfg.Project.Name)
	}
	if len(cfg.Compiler.Cflags) != 2 {
		t.Errorf("expected 2 cflags, got %d", len(cfg.Compiler.Cflags))
	}
}

func TestLoadDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: esp32c3
compiler:
  prefix: riscv32-esp-elf-
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.Project.Name != "firmware" {
		t.Errorf("expected default project name 'firmware', got %s", cfg.Project.Name)
	}
	if len(cfg.Project.Sources) != 1 || cfg.Project.Sources[0] != "src/" {
		t.Errorf("expected default sources ['src/'], got %v", cfg.Project.Sources)
	}
}

func TestLoadNotExist(t *testing.T) {
	_, err := Load("/nonexistent/zyrthi.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: [invalid yaml
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestWriteCompileCommands(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	commands := []CompileCommand{
		{Directory: ".", Command: "gcc -c main.c", File: "main.c"},
		{Directory: ".", Command: "gcc -c helper.c", File: "helper.c"},
	}

	err := WriteCompileCommands(commands)
	if err != nil {
		t.Fatalf("WriteCompileCommands error: %v", err)
	}

	if _, err := os.Stat("compile_commands.json"); err != nil {
		t.Error("compile_commands.json should exist")
	}

	data, err := os.ReadFile("compile_commands.json")
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("compile_commands.json should not be empty")
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Platform: "esp32",
		Chip:     "esp32c3",
		Compiler: CompilerConfig{
			Prefix:   "riscv32-esp-elf-",
			Cflags:   []string{"-Os", "-g"},
			Ldflags:  []string{"-nostdlib"},
			Includes: []string{"include/"},
		},
		Flash: FlashConfig{
			Plugin:      "https://example.com/plugin.wasm",
			EntryAddr:   "0x0",
			FlashSize:   "4MB",
			DefaultBaud: 115200,
		},
		Project: ProjectConfig{
			Name:    "test-project",
			Sources: []string{"src/"},
		},
	}

	if cfg.Platform != "esp32" {
		t.Errorf("expected platform 'esp32', got %s", cfg.Platform)
	}
}

func TestCompilerConfigStruct(t *testing.T) {
	cc := CompilerConfig{
		Prefix:   "arm-none-eabi-",
		Cflags:   []string{"-mcpu=cortex-m4", "-mthumb"},
		Ldflags:  []string{"-T", "linker.ld"},
		Includes: []string{"include/", "lib/include/"},
	}

	if cc.Prefix != "arm-none-eabi-" {
		t.Errorf("expected prefix 'arm-none-eabi-', got %s", cc.Prefix)
	}
}

func TestFlashConfigStruct(t *testing.T) {
	fc := FlashConfig{
		Plugin:      "https://example.com/plugin.wasm",
		EntryAddr:   "0x0",
		FlashSize:   "4MB",
		DefaultBaud: 115200,
	}

	if fc.Plugin != "https://example.com/plugin.wasm" {
		t.Errorf("expected plugin URL, got %s", fc.Plugin)
	}
}

func TestProjectConfigStruct(t *testing.T) {
	pc := ProjectConfig{
		Name:    "my-firmware",
		Sources: []string{"src/", "lib/src/"},
	}

	if pc.Name != "my-firmware" {
		t.Errorf("expected name 'my-firmware', got %s", pc.Name)
	}
}

func TestCompileCommandStruct(t *testing.T) {
	cc := CompileCommand{
		Directory: "/project",
		Command:   "gcc -c -o main.o main.c",
		File:      "main.c",
	}

	if cc.Directory != "/project" {
		t.Errorf("expected directory '/project', got %s", cc.Directory)
	}
}