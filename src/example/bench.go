package main

import (
	"context"
	_ "context/aaa"
	_ "context/cli"
	_ "context/tcp"
	_ "context/web"
)

func main() {
	ctx.Start()
}
