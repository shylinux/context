package main
import (
	"bufio"
	"bytes"
	"sync/atomic"
	"contexts/ctx"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
	kit "toolkit"
)

var Index = &ctx.Context{Name: "test", Help: "接口测试工具",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"nread": {Name: "nread", Help: "读取Channel长度", Value: 1},
		"nwork": {Name: "nwork", Help: "协程数量", Value: 20},
		"limit":{Name: "limit", Help: "请求数量", Value: 100},
		"nskip": {Name: "nskip", Help: "请求限长", Value: 100},
		"nwrite":{Name: "nwrite", Help: "输出Channel长度", Value: 1},
		"outdir": {Name: "outdir", Help: "输出目录", Value: "tmp"},
		"prefix0": {Name: "prefix0", Help: "请求前缀", Value: "uri["},
		"prefix1": {Name: "prefix1", Help: "参数前缀", Value: "request_param["},
		"timeout": {Name: "timeout", Help: "请求超时", Value: "10s"},
	},
	Commands: map[string]*ctx.Command{
		"diff": {Name: "diff file server1 server2", Help:"接口对比工具", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			os.MkdirAll(m.Conf("outdir"), 0777)

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
			var all, skip, diff, same, error int32 = 0, 0, 0, 0, 0

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
						atomic.AddInt32(&all, 1)
						var t3, t4 time.Duration
						if e1 != nil || e2 != nil {
							atomic.AddInt32(&error, 1)
							fmt.Printf("%d%% %d cost: %v/%v error: %v %v\n", int(count*100.0/total), error, t1, t2, e1, e2)

						} else if res1.StatusCode != http.StatusOK || res2.StatusCode != http.StatusOK {
							atomic.AddInt32(&error, 1)
							fmt.Printf("%d%% %d %s %s cost: %v/%v error: %v %v\n", int(count*100.0/total), error, "error", uri, t1, t2, res1.StatusCode, res2.StatusCode)
						} else {

							begin = time.Now()
							b1, _ := ioutil.ReadAll(res1.Body)
							b2, _ := ioutil.ReadAll(res2.Body)
							t3 = time.Since(begin)

							begin = time.Now()
							var num int32
							if bytes.Compare(b1, b2) == 0 {
								atomic.AddInt32(&same, 1)
								num = same
								result = "same"

							} else {
								atomic.AddInt32(&diff, 1)
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
		"cost": {Name: "cost file server nroute", Help:"接口耗时测试",
			Form: map[string]int{"nwork": 1, "limit": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			os.MkdirAll(m.Conf("outdir"), 0777)

			f, e := os.Open(arg[0])
			m.Assert(e)
			defer f.Close()

			input := make(chan []string, kit.Int(m.Confx("nread")))
			output := make(chan []string, kit.Int(m.Confx("nwrite")))

			count, nline := 0, 0

			wg := sync.WaitGroup{}
			limit := kit.Int(m.Confx("limit"))

			var skip, success int32 = 0, 0
			begin := time.Now()
			api := map[string]int{}
			go func() {
				for bio := bufio.NewScanner(f); bio.Scan() && nline < limit; {
					text := bio.Text()
					count += len(text)+1

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

					// fmt.Printf("line: %d %d%% %v\n", nline, int(nline/limit), uri)
					nline++
					input <- []string{uri, arg, fmt.Sprintf("%d", nline)}
					api[uri]++
				}
				close(input)
				wg.Wait()
				close(output)
			}()

			var times time.Duration
			for i := 0; i < kit.Int(m.Confx("nwork")); i++ {
				go func(msg *ctx.Message) {
					wg.Add(1)
					defer func() {
						wg.Done()
					}()
					client := http.Client{Timeout:kit.Duration(m.Conf("timeout"))}

					for {
						word, ok := <-input
						if !ok {
							break
						}
						uri, args, key := word[0], word[1], word[2]

						fmt.Printf("%v/%v\tpost: %v\t%v\n", key, limit, arg[1]+uri, args)
						begin := time.Now()
						res, err := client.Post(arg[1]+uri, "application/json", bytes.NewReader([]byte(args)))
						if res.StatusCode == http.StatusOK {
							io.Copy(ioutil.Discard, res.Body)
							atomic.AddInt32(&success, 1)
						} else {
							fmt.Printf("%v/%v\terror: %v\n", key, limit, err)
						}
						times += time.Since(begin)
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
			m.Echo("\n\nnclient: %v skip: %v nreq: %v success: %v time: %v average: %vms",
				m.Confx("nwork"), skip, nline, success, time.Since(begin), int(times)/nline/1000000)
			return
		}},
		"cmd": {Name: "cmd file", Help:"生成测试命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
	kit.DisableLog = true
	fmt.Print(Index.Plugin(os.Args[1:]))
}
