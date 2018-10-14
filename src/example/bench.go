package main

import (
	"os"

	"contexts"
	_ "contexts/aaa"
	_ "contexts/cli"
	_ "contexts/web"

	_ "contexts/lex"
	_ "contexts/log"
	_ "contexts/yac"

	_ "contexts/mdb"
	_ "contexts/nfs"
	_ "contexts/ssh"
	_ "contexts/tcp"
)

func main() {
	ctx.Start(os.Args[1:]...)
}
