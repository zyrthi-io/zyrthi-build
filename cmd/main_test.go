package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zyrthi-io/zyrthi-build/internal/config"
)

func TestBuildCompileCommand(t *testing.T) {
	cfg := &config.Config{
		Compiler: config.CompilerConfig{
			Prefix:   "arm-none-eabi-",
			Cflags:   []string{"-Os", "-g", "-Wall"},
			Includes: []string{"include/", "src/"},
		},
	}

	src := "src/main.c"
	obj := "build/main.o"

	cmd, cmdStr := buildCompileCommand(cfg, src, obj)

	if !strings.Contains(cmd.Path, "arm-none-eabi-gcc") {
		t.Errorf("expected gcc command, got %s", cmd.Path)
	}

	expectedParts := []string{
		"arm-none-eabi-gcc",
		"-c",
		"-Os",
		"-g",
		"-Wall",
		"-Iinclude/",
		"-Isrc/",
		"-o", "build/main.o",
		"src/main.c",
	}

	for _, part := range expectedParts {
		if !strings.Contains(cmdStr, part) {
			t.Errorf("expected command to contain %s, got %s", part, cmdStr)
		}
	}
}

func TestBuildCompileCommandMinimal(t *testing.T) {
	cfg := &config.Config{
		Compiler: config.CompilerConfig{
			Prefix:   "gcc",
			Cflags:   []string{},
			Includes: []string{},
		},
	}

	cmd, cmdStr := buildCompileCommand(cfg, "test.c", "test.o")

	if cmd == nil {
		t.Error("expected cmd to be non-nil")
	}
	if !strings.Contains(cmdStr, "-c") {
		t.Error("expected -c flag")
	}
}

func TestCheckToolchainNoPrefix(t *testing.T) {
	cfg := &config.Config{
		Compiler: config.CompilerConfig{
			Prefix: "",
		},
	}

	err := checkToolchain(cfg)
	if err == nil {
		t.Error("expected error for empty prefix")
	}
}

func TestCheckToolchainInvalidPrefix(t *testing.T) {
	cfg := &config.Config{
		Compiler: config.CompilerConfig{
			Prefix: "nonexistent-compiler-",
		},
	}

	err := checkToolchain(cfg)
	if err == nil {
		t.Error("expected error for nonexistent compiler")
	}
}

func TestCleanBuild(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	buildDir := filepath.Join(tmpDir, "build")
	os.MkdirAll(buildDir, 0755)
	os.WriteFile(filepath.Join(buildDir, "test.o"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "compile_commands.json"), []byte{}, 0644)

	cfg := &config.Config{}
	cleanBuild(cfg)

	if _, err := os.Stat(buildDir); !os.IsNotExist(err) {
		t.Error("expected build directory to be removed")
	}
}

func TestCollectSources(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.c"), []byte{}, 0644)
	os.WriteFile(filepath.Join(srcDir, "helper.cpp"), []byte{}, 0644)
	os.WriteFile(filepath.Join(srcDir, "readme.txt"), []byte{}, 0644)

	cfg := &config.Config{
		Project: config.ProjectConfig{
			Sources: []string{srcDir},
		},
	}

	sources, err := collectSources(cfg)
	if err != nil {
		t.Fatalf("collectSources error: %v", err)
	}

	if len(sources) != 2 {
		t.Errorf("expected 2 source files, got %d", len(sources))
	}
}

func TestCollectSourcesNested(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	os.MkdirAll("src/drivers", 0755)
	os.MkdirAll("src/utils", 0755)
	os.WriteFile("src/main.c", []byte{}, 0644)
	os.WriteFile("src/drivers/spi.c", []byte{}, 0644)
	os.WriteFile("src/utils/helpers.c", []byte{}, 0644)

	cfg := &config.Config{
		Project: config.ProjectConfig{
			Sources: []string{"src/"},
		},
	}

	sources, err := collectSources(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sources) != 3 {
		t.Errorf("expected 3 source files, got %d", len(sources))
	}
}

func TestCollectSourcesWithCpp(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	os.MkdirAll("src", 0755)
	os.WriteFile("src/main.c", []byte{}, 0644)
	os.WriteFile("src/utils.cpp", []byte{}, 0644)
	os.WriteFile("src/helper.cc", []byte{}, 0644)
	os.WriteFile("src/header.h", []byte{}, 0644)

	cfg := &config.Config{
		Project: config.ProjectConfig{
			Sources: []string{"src/"},
		},
	}

	sources, err := collectSources(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sources) != 3 {
		t.Errorf("expected 3 source files, got %d: %v", len(sources), sources)
	}
}
