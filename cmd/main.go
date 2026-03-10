package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zyrthi-io/zyrthi-build/internal/config"
)

var (
	flagConfig  = flag.String("config", "zyrthi.yaml", "配置文件路径")
	flagClean   = flag.Bool("clean", false, "清理编译产物")
	flagVerbose = flag.Bool("v", false, "详细输出")
)

func main() {
	flag.Parse()

	// 读取配置
	cfg, err := config.Load(*flagConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法读取配置文件: %v\n", err)
		os.Exit(1)
	}

	// 清理模式
	if *flagClean {
		cleanBuild(cfg)
		return
	}

	// 检查工具链
	if err := checkToolchain(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 编译
	if err := build(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "编译失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("编译成功")
}

// checkToolchain 检查工具链是否存在
func checkToolchain(cfg *config.Config) error {
	prefix := cfg.Compiler.Prefix
	if prefix == "" {
		return fmt.Errorf("未配置编译器前缀")
	}

	tools := []string{"gcc", "g++", "objcopy", "size"}
	for _, tool := range tools {
		cmd := prefix + tool
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("找不到编译器: %s", cmd)
		}
	}

	return nil
}

// build 执行编译
func build(cfg *config.Config) error {
	buildDir := "build"

	// 创建 build 目录
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return err
	}

	// 收集源文件
	sources, err := collectSources(cfg)
	if err != nil {
		return err
	}

	if len(sources) == 0 {
		return fmt.Errorf("未找到源文件")
	}

	fmt.Printf("编译 %d 个源文件...\n", len(sources))

	// 编译每个源文件
	objects := make([]string, 0, len(sources))
	compileCommands := make([]config.CompileCommand, 0, len(sources))

	for _, src := range sources {
		obj := filepath.Join(buildDir, strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))+".o")
		objects = append(objects, obj)

		// 编译命令
		cmd, cmdStr := buildCompileCommand(cfg, src, obj)
		compileCommands = append(compileCommands, config.CompileCommand{
			Directory: ".",
			Command:   cmdStr,
			File:      src,
		})

		if *flagVerbose {
			fmt.Println(cmdStr)
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("编译 %s 失败: %v", src, err)
		}
	}

	// 生成 compile_commands.json
	if err := config.WriteCompileCommands(compileCommands); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 无法生成 compile_commands.json: %v\n", err)
	}

	// 链接
	elfFile := filepath.Join(buildDir, cfg.Project.Name+".elf")
	if err := link(cfg, objects, elfFile); err != nil {
		return err
	}

	// 生成 bin
	binFile := filepath.Join(buildDir, cfg.Project.Name+".bin")
	if err := objcopy(cfg, elfFile, binFile); err != nil {
		return err
	}

	// 显示大小
	showSize(cfg, elfFile)

	return nil
}

// collectSources 收集源文件
func collectSources(cfg *config.Config) ([]string, error) {
	sources := make([]string, 0)

	for _, dir := range cfg.Project.Sources {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".c" || ext == ".cpp" || ext == ".cc" {
				sources = append(sources, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return sources, nil
}

// buildCompileCommand 构建编译命令
func buildCompileCommand(cfg *config.Config, src, obj string) (*exec.Cmd, string) {
	args := []string{}

	// 编译选项
	args = append(args, "-c")
	args = append(args, cfg.Compiler.Cflags...)

	// 包含路径
	for _, inc := range cfg.Compiler.Includes {
		args = append(args, "-I"+inc)
	}

	// 输出
	args = append(args, "-o", obj)

	// 源文件
	args = append(args, src)

	cmd := exec.Command(cfg.Compiler.Prefix+"gcc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 构建命令字符串
	cmdStr := cfg.Compiler.Prefix + "gcc " + strings.Join(args, " ")

	return cmd, cmdStr
}

// link 链接
func link(cfg *config.Config, objects []string, output string) error {
	args := []string{}

	// 链接选项
	args = append(args, cfg.Compiler.Ldflags...)

	// 输出
	args = append(args, "-o", output)

	// 目标文件
	args = append(args, objects...)

	if *flagVerbose {
		fmt.Println(cfg.Compiler.Prefix+"g++", strings.Join(args, " "))
	}

	cmd := exec.Command(cfg.Compiler.Prefix+"g++", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// objcopy 生成 bin 文件
func objcopy(cfg *config.Config, elf, bin string) error {
	cmd := exec.Command(cfg.Compiler.Prefix+"objcopy", "-O", "binary", elf, bin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// showSize 显示大小
func showSize(cfg *config.Config, elf string) {
	cmd := exec.Command(cfg.Compiler.Prefix+"size", elf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// cleanBuild 清理编译产物
func cleanBuild(cfg *config.Config) {
	buildDir := "build"
	if _, err := os.Stat(buildDir); err == nil {
		os.RemoveAll(buildDir)
		fmt.Println("已清理 build 目录")
	}

	// 清理 compile_commands.json
	os.Remove("compile_commands.json")
}