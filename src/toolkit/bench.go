package main

import (
	"context"
	_ "context/cli"
	_ "context/ssh"
	_ "context/tcp"
	_ "context/web"
)

func main() {
	ctx.Start()
}
