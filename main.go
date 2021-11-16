package main

import (
	"github.com/Makav3li94/yuz/cli"
	"os"
)


func main() {
	defer os.Exit(0)
	cmd := cli.CommandLine{}
	cmd.Run()
}