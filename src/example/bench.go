package main

import (
	"context"
	_ "context/cli"
	_ "context/mdb"
	_ "context/tcp"
	"os"
)

func main() {
	ctx.Start(os.Args[1:]...)
}
