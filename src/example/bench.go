package main

import (
	"context"
	_ "context/cli"

	_ "context/mdb"
	_ "context/tcp"

	_ "context/aaa"
	_ "context/ssh"

	_ "context/web"

	_ "context/lex"

	"os"
)

func main() {
	ctx.Start(os.Args[1:]...)
}
