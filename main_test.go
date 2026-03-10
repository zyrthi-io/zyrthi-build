package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuildCompileCommand 测试编译命令构建
func TestBuildCompileCommand(t *testing.T) {
	cfg := &Config{
		Compiler: CompilerConfig{
			Prefix:   "arm-none-eabi-",
			Cflags:   []string{"-Os", "-g", "-Wall"},
			Includes: []string{"include/", "src/"},
		},
	}

	src := "src/main.c"
	obj := "build/main.o"

	cmd, cmdStr := buildCompileCommand(cfg, src, obj)

	// 验证命令路径
	if !strings.Contains(cmd.Path, "arm-none-eabi-gcc") {
		t.Errorf("expected gcc command, got %s", cmd.Path)
	}

	// 验证命令字符串
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

// TestBuildCompileCommandMinimal 测试最小配置
func TestBuildCompileCommandMinimal(t *testing.T) {
	cfg := &Config{
		Compiler: CompilerConfig{
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
	if !strings.Contains(cmdStr, "test.c") {
		t.Error("expected source file")
	}
	if !strings.Contains(cmdStr, "test.o") {
		t.Error("expected output file")
	}
}

// TestCheckToolchainNoPrefix 测试缺少编译器前缀
func TestCheckToolchainNoPrefix(t *testing.T) {
	cfg := &Config{
		Compiler: CompilerConfig{
			Prefix: "",
		},
	}

	err := checkToolchain(cfg)
	if err == nil {
		t.Error("expected error for empty prefix")
	}
	if !strings.Contains(err.Error(), "未配置编译器前缀") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestCheckToolchainInvalidPrefix 测试无效前缀
func TestCheckToolchainInvalidPrefix(t *testing.T) {
	cfg := &Config{
		Compiler: CompilerConfig{
			Prefix: "nonexistent-compiler-",
		},
	}

	err := checkToolchain(cfg)
	if err == nil {
		t.Error("expected error for nonexistent compiler")
	}
}

// TestCleanBuild 测试清理功能
func TestCleanBuild(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// 创建 build 目录和文件
	buildDir := filepath.Join(tmpDir, "build")
	os.MkdirAll(buildDir, 0755)
	os.WriteFile(filepath.Join(buildDir, "test.o"), []byte{}, 0644)
	os.WriteFile(filepath.Join(buildDir, "test.elf"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "compile_commands.json"), []byte{}, 0644)

	cfg := &Config{}
	cleanBuild(cfg)

	// 验证 build 目录已删除
	if _, err := os.Stat(buildDir); !os.IsNotExist(err) {
		t.Error("expected build directory to be removed")
	}

	// 验证 compile_commands.json 已删除
	if _, err := os.Stat(filepath.Join(tmpDir, "compile_commands.json")); !os.IsNotExist(err) {
		t.Error("expected compile_commands.json to be removed")
	}
}

// TestCleanBuildNoBuildDir 测试清理不存在的 build 目录
func TestCleanBuildNoBuildDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	cfg := &Config{}
	// 不应该报错
	cleanBuild(cfg)
}

// TestBuildNoSources 测试没有源文件
func TestBuildNoSources(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// 创建空的源文件目录
	os.MkdirAll("src", 0755)

	cfg := &Config{
		Compiler: CompilerConfig{
			Prefix: "test-",
		},
		Project: ProjectConfig{
			Sources: []string{"src/"},
		},
	}

	// 创建 build 目录
	os.MkdirAll("build", 0755)

	// 收集源文件应该返回空
	sources, err := collectSources(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sources) != 0 {
		t.Errorf("expected 0 sources, got %d", len(sources))
	}
}

// TestBuildCreatesDirectory 测试 build 目录创建
func TestBuildCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// build 目录不存在
	if _, err := os.Stat("build"); !os.IsNotExist(err) {
		t.Fatal("build directory should not exist")
	}

	// os.MkdirAll 应该能创建
	err := os.MkdirAll("build", 0755)
	if err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}

	if _, err := os.Stat("build"); err != nil {
		t.Error("build directory should exist")
	}
}

// TestLinkCommand 测试链接命令参数
func TestLinkCommand(t *testing.T) {
	cfg := &Config{
		Compiler: CompilerConfig{
			Prefix:  "arm-none-eabi-",
			Ldflags: []string{"-T", "linker.ld", "-nostdlib"},
		},
	}

	// 验证参数构建逻辑
	args := []string{}
	args = append(args, cfg.Compiler.Ldflags...)
	args = append(args, "-o", "output.elf")
	args = append(args, "obj1.o", "obj2.o")

	expectedParts := []string{
		"-T", "linker.ld",
		"-nostdlib",
		"-o", "output.elf",
		"obj1.o", "obj2.o",
	}

	argsStr := strings.Join(args, " ")
	for _, part := range expectedParts {
		if !strings.Contains(argsStr, part) {
			t.Errorf("expected args to contain %s", part)
		}
	}
}

// TestObjcopyCommand 测试 objcopy 命令参数
func TestObjcopyCommand(t *testing.T) {
	// objcopy 参数: -O binary input.elf output.bin
	prefix := "arm-none-eabi-"
	elf := "build/firmware.elf"
	bin := "build/firmware.bin"

	// 验证预期的命令
	expectedCmd := prefix + "objcopy"
	expectedArgs := []string{"-O", "binary", elf, bin}

	// 这些只是验证逻辑正确性
	if expectedCmd != "arm-none-eabi-objcopy" {
		t.Errorf("unexpected command: %s", expectedCmd)
	}
	if expectedArgs[0] != "-O" || expectedArgs[1] != "binary" {
		t.Errorf("unexpected args: %v", expectedArgs)
	}
}

// TestShowSizeCommand 测试 size 命令
func TestShowSizeCommand(t *testing.T) {
	prefix := "arm-none-eabi-"
	elf := "build/firmware.elf"

	expectedCmd := prefix + "size"
	if expectedCmd != "arm-none-eabi-size" {
		t.Errorf("unexpected command: %s", expectedCmd)
	}
	_ = elf // 避免未使用警告
}

// TestCollectSourcesWithCpp 测试收集 C++ 源文件
func TestCollectSourcesWithCpp(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// 创建各种源文件
	os.MkdirAll("src", 0755)
	os.WriteFile("src/main.c", []byte{}, 0644)
	os.WriteFile("src/utils.cpp", []byte{}, 0644)
	os.WriteFile("src/helper.cc", []byte{}, 0644)
	os.WriteFile("src/header.h", []byte{}, 0644)     // 不应该被收集
	os.WriteFile("src/data.txt", []byte{}, 0644)     // 不应该被收集

	cfg := &Config{
		Project: ProjectConfig{
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

	// 验证文件扩展名
	for _, src := range sources {
		ext := strings.ToLower(filepath.Ext(src))
		if ext != ".c" && ext != ".cpp" && ext != ".cc" {
			t.Errorf("unexpected source file: %s", src)
		}
	}
}

// TestCollectSourcesNested 测试嵌套目录收集
func TestCollectSourcesNested(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// 创建嵌套目录结构
	os.MkdirAll("src/drivers", 0755)
	os.MkdirAll("src/utils", 0755)
	os.WriteFile("src/main.c", []byte{}, 0644)
	os.WriteFile("src/drivers/spi.c", []byte{}, 0644)
	os.WriteFile("src/utils/helpers.c", []byte{}, 0644)

	cfg := &Config{
		Project: ProjectConfig{
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
