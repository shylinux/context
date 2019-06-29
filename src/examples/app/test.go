package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("usage: %s file server", os.Args[0])
		return
	}

	prefix0 := len("uri[")
	prefix1 := len("request_param[")

	if os.Mkdir("tmp", 0777) != nil {
	}

	begin := time.Now()
	f, e := os.Open(os.Args[1])
	if e != nil {
		fmt.Printf("%s\n", e)
		os.Exit(1)
	}
	defer f.Close()
	bio := bufio.NewScanner(f)
	output := map[string]*os.File{}
	nreq := 0

	for bio.Scan() {
		word := strings.Split(bio.Text(), " ")
		if len(word) != 2 {
			continue
		}
		uri := word[0][prefix0:len(word[0])-1]
		arg := word[1][prefix1:len(word[1])-1]
		if output[uri] == nil {
			name := path.Join("tmp", strings.Replace(uri, "/", "_", -1)+".txt")
			f, e = os.Create(name)
			output[uri] = f
		}
		nreq++
		br := bytes.NewReader([]byte(arg))
		fmt.Printf("%d post: %v\n", nreq, os.Args[2]+uri)
		res, e := http.Post(os.Args[2]+uri, "application/json", br)
		fmt.Fprintf(output[uri], uri)
		fmt.Fprintf(output[uri], " arguments:")
		fmt.Fprintf(output[uri], arg)
		fmt.Fprintf(output[uri], " result:")
		if e != nil {
			fmt.Fprintf(output[uri], "%v", e)
		} else if res.StatusCode != http.StatusOK {
			fmt.Fprintf(output[uri], res.Status)
		} else {
			io.Copy(output[uri], res.Body)
		}
		fmt.Fprintf(output[uri], "\n")
	}
	fmt.Printf("nuri: %v nreq: %v cost: %v", len(output), nreq, time.Since(begin))
}
