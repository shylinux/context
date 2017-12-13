package main

import (
	"context"
	_ "context/aaa"
	_ "context/cli"

	_ "context/mdb"
	_ "context/tcp"
	_ "context/web"

	_ "context/ssh"

	_ "context/lex"

	"os"
)

func main() {
	ctx.Start(os.Args[1:]...)
}
