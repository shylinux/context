package main

import (
	_ "context"
	"context/cli"
	_ "context/ssh"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		cli.Index.Conf("log", os.Args[1])
	}

	if len(os.Args) > 2 {
		cli.Index.Conf("init.sh", os.Args[2])
	}

	cli.Index.Start()
}
