package main

import (
	"context"
	_ "context/aaa"
	_ "context/cli"
	_ "context/tcp"
	_ "context/web"
	_ "context/web/mp"
)

func main() {
	ctx.Start()
}
