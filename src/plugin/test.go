package main
import (
	"bufio"
	"bytes"
	"contexts/ctx"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
	kit "toolkit"
)

var Index = &ctx.Context{Name: "test", Help: "test",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"nread": {Name: "nread", Value: 1},
		"nwork": {Name: "nwork", Value: 10},
		"nskip": {Name: "nwork", Value: 100},
		"nwrite":{Name: "nwrite", Value: 1},
		"outdir": {Name: "outdir", Value: "tmp"},
		"prefix0": {Name: "prefix0", Value: "uri["},
		"prefix1": {Name: "prefix1", Value: "request_param["},
		"timeout": {Name: "timeout", Value: "10s"},
	},
	Commands: map[string]*ctx.Command{
		"diff": {Name: "diff", Help:"diff", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			f, e := os.Open(arg[0])
			m.Assert(e)
			defer f.Close()
			http.DefaultClient.Timeout = kit.Duration(m.Conf("timeout"))

			input := make(chan []string, kit.Int(m.Confx("nread")))
			output := make(chan []string, kit.Int(m.Confx("nwrite")))

			s, e := f.Stat()
			m.Assert(e)
			total, count, nline := int(s.Size()), 0, 0

			mu := sync.Mutex{}
			wg := sync.WaitGroup{}
			all, skip, diff, same, error := 0, 0, 0, 0, 0

			begin := time.Now()
			api := map[string]int{}
			go func() {
				for bio := bufio.NewScanner(f); bio.Scan(); {
					text := bio.Text()
					count += len(text)+1
					// fmt.Printf("line: %d %d%% \n", nline, int(count*100.0/total))

					word := strings.Split(text, " ")
					if len(word) != 2 {
						continue
					}
					uri := word[0][len(m.Conf("prefix0")) : len(word[0])-1]
					arg := word[1][len(m.Conf("prefix1")) : len(word[1])-1]
					if len(arg) > m.Confi("nskip") {
						skip++
						continue
					}

					nline++
					input <- []string{uri, arg, fmt.Sprintf("%d", nline)}
					api[uri]++
				}
				close(input)
				wg.Wait()
				close(output)
			}()

			for i := 0; i < m.Confi("nwork"); i++ {
				go func(msg *ctx.Message) {
					wg.Add(1)
					defer func() {
						wg.Done()
					}()

					for {
						word, ok := <-input
						if !ok {
							break
						}
						uri, args, key := word[0], word[1], word[2]

						begin := time.Now()
						res1, e1 := http.Post(arg[1]+uri, "application/json", bytes.NewReader([]byte(args)))
						t1 := time.Since(begin)

						begin = time.Now()
						res2, e2 := http.Post(arg[2]+uri, "application/json", bytes.NewReader([]byte(args)))
						t2 := time.Since(begin)

						size1, size2 := 0, 0
						result := "error"
						mu.Lock()
						all++
						mu.Unlock()
						var t3, t4 time.Duration
						if e1 != nil || e2 != nil {
							mu.Lock()
							error++
							mu.Unlock()
							fmt.Printf("%d%% %d cost: %v/%v error: %v %v\n", int(count*100.0/total), error, t1, t2, e1, e2)

						} else if res1.StatusCode != http.StatusOK || res2.StatusCode != http.StatusOK {
							mu.Lock()
							error++
							mu.Unlock()
							fmt.Printf("%d%% %d %s %s cost: %v/%v error: %v %v\n", int(count*100.0/total), error, "error", uri, t1, t2, res1.StatusCode, res2.StatusCode)
						} else {

							begin = time.Now()
							b1, _ := ioutil.ReadAll(res1.Body)
							b2, _ := ioutil.ReadAll(res2.Body)
							t3 = time.Since(begin)

							begin = time.Now()
							num := 0
							if bytes.Compare(b1, b2) == 0 {
								mu.Lock()
								same++
								mu.Unlock()
								num = same
								result = "same"

							} else {
								mu.Lock()
								diff++
								mu.Unlock()
								num = diff
								result = "diff"

								result1 := []string{
									fmt.Sprintf("index:%v uri:", key),
									fmt.Sprintf(uri),
									fmt.Sprintf(" arguments:"),
									fmt.Sprintf(args),
								}

								result1 = append(result1, "\n")
								result1 = append(result1, "\n")
								result1 = append(result1, "result0:")
								result1 = append(result1, string(b1))
								result1 = append(result1, "\n")
								result1 = append(result1, "\n")
								result1 = append(result1, "result1:")
								result1 = append(result1, string(b2))
								result1 = append(result1, "\n")
								result1 = append(result1, "\n")
								output <-  result1
							}
							size1 = len(b1)
							size2 = len(b2)
							t4 = time.Since(begin)
							fmt.Printf("%d%% %d %s %s size: %v/%v cost: %v/%v diff: %v/%v\n", int(count*100.0/total), num, result, uri, len(b1), len(b2), t1, t2, t3, t4)
						}

						mu.Lock()
						m.Add("append", "uri", uri)
						m.Add("append", "time1", t1/1000000)
						m.Add("append", "time2", t2/1000000)
						m.Add("append", "time3", t3/1000000)
						m.Add("append", "time4", t4/1000000)
						m.Add("append", "size1", size1)
						m.Add("append", "size2", size2)
						m.Add("append", "action", result)
						mu.Unlock()
					}
				}(m.Spawn())
			}

			files := map[string]*os.File{}
			for {
				data, ok := <-output
				if !ok {
					break
				}

				uri := data[1]
				if files[uri] == nil {
					f, _ := os.Create(path.Join(m.Conf("outdir"), strings.Replace(uri, "/", "_", -1)+".txt"))
					defer f.Close()
					files[uri] = f
				}
				for _, v:= range data {
					fmt.Fprintf(files[uri], v)
				}
			}

			fmt.Printf("\n\nnuri: %v nreq: %v skip: %v same: %v diff: %v error: %v cost: %v\n\n", len(api), nline, skip, same, diff, error, time.Since(begin))

			m.Sort("time1", "int").Table()
			return
		}},
		"output": {Name: "output", Help:"output", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			f, e := os.Open(arg[0])
			m.Assert(e)
			defer f.Close()
			for bio := bufio.NewScanner(f); bio.Scan(); {
				text := bio.Text()

				word := strings.Split(text, " ")
				if len(word) != 2 {
					continue
				}
				uri := word[0][len(m.Conf("prefix0")) : len(word[0])-1]
				arg := word[1][len(m.Conf("prefix1")) : len(word[1])-1]
				if len(arg) > m.Confi("nskip") {
					continue
				}
				m.Echo("./hey -n 10000 -c 100 -m POST -T \"application/json\" -d '%s' http://127.0.0.1:6363%s\n", arg, uri)
			}
			return
		}},
	},
}
func main() {
	fmt.Print(Index.Plugin(os.Args[1:]))
}
