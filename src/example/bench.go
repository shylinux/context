package main

import (
	"contexts"
	_ "contexts/aaa"
	_ "contexts/cli"
	_ "contexts/ssh"

	_ "contexts/mdb"
	_ "contexts/nfs"
	_ "contexts/tcp"
	_ "contexts/web"

	_ "contexts/lex"
	_ "contexts/log"
	_ "contexts/yac"

	"os"
)

func main() {
	ctx.Start(os.Args[1:]...)
}
