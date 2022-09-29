package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	platform := flag.String("p", "", "操作系统平台(windows, linux, darwin...)")
	arch := flag.String("a", "", "处理器架构(amd64, 386, arm, arm64...)")
	help := flag.Bool("h", false, "显示此帮助")
	flag.Parse()

	if *help || len(*platform) == 0 || len(*arch) == 0 {
		fmt.Println("使用方法: go run build/main.go -p <platform> -a <arch>")
		flag.PrintDefaults()
		fmt.Println("注：所有支持的平台及处理器架构可以在这里查看: https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63")
		return
	}

	os.Setenv("GOOS", *platform)
	os.Setenv("GOARCH", *arch)

	var outName string
	if *platform == "windows" {
		outName = "cns.exe"
	} else {
		outName = "cns"
	}

	cmd := exec.Command("go", "build", "-o", outName)
	if err := cmd.Run(); err != nil {
		fmt.Println("编译失败，请检查arch与platform是否正确")
	} else {
		fmt.Printf("编译成功，生成可执行文件：%s\n", outName)
	}
}
