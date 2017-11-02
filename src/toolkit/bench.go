package main

import (
	"context"
	_ "context/cli"
	_ "context/ssh"
	_ "context/tcp"
	_ "context/web"
	"os"
)

func main() {
	ctx.Index.Init(os.Args[1:]...)
}
