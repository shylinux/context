package main

import (
	"contexts/cli"
	"contexts/ctx"
	"toolkit"

	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func Input(m *ctx.Message, file string, input chan []string) {
	f, e := os.Open(file)
	m.Assert(e)
	defer f.Close()

	nline := 0
	for bio := bufio.NewScanner(f); nline < kit.Int(m.Confx("limit")) && bio.Scan(); {
		word := strings.Split(bio.Text(), " ")
		if len(word) != 2 {
			continue
		}

		uri := word[0][len(m.Conf("prefix0")) : len(word[0])-1]
		arg := word[1][len(m.Conf("prefix1")) : len(word[1])-1]
		if len(arg) > m.Confi("nskip") {
			continue
		}

		input <- []string{fmt.Sprintf("%d", nline), uri, arg}
		if nline++; nline%kit.Int(m.Confx("nsleep")) == 0 {
			fmt.Printf("nline:%d sleep 1s...\n", nline)
			time.Sleep(time.Second)
		}
	}
	close(input)
}

func Output(m *ctx.Message, output <-chan []string) {
	os.MkdirAll(m.Conf("outdir"), 0777)

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
		for _, v := range data {
			fmt.Fprintf(files[uri], v)
		}
	}
}

func Process(m *ctx.Message, file string, cb func(*ctx.Message, *http.Client, []string, chan []string)) (int32, time.Duration) {
	nline, begin := int32(0), time.Now()
	input := make(chan []string, kit.Int(m.Confx("nread")))
	output := make(chan []string, kit.Int(m.Confx("nwrite")))

	go Input(m, file, input)
	go Output(m, output)

	wg := sync.WaitGroup{}
	for i := 0; i < m.Confi("nwork"); i++ {
		go func(msg *ctx.Message) {
			wg.Add(1)
			defer wg.Done()

			for {
				word, ok := <-input
				if !ok {
					break
				}
				atomic.AddInt32(&nline, 1)
				cb(msg, &http.Client{Timeout: kit.Duration(m.Conf("timeout"))}, word, output)
			}
		}(m.Spawn())
	}
	runtime.Gosched()
	wg.Wait()
	close(output)
	return nline, time.Since(begin)
}

func Compare(b1 []byte, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	if bytes.Compare(b1, b2) == 0 {
		return true
	}

	var d1, d2 interface{}
	json.Unmarshal(b1, &d1)
	json.Unmarshal(b2, &d2)
	s1, _ := json.Marshal(d1)
	s2, _ := json.Marshal(d2)
	if bytes.Compare(s1, s2) == 0 {
		return true
	}
	return false
}

var Index = &ctx.Context{Name: "test", Help: "测试工具",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"nread":  {Name: "nread", Help: "读取Channel长度", Value: 1},
		"nwork":  {Name: "nwork", Help: "协程数量", Value: 20},
		"limit":  {Name: "limit", Help: "请求数量", Value: 100},
		"nskip":  {Name: "nskip", Help: "请求限长", Value: 100},
		"nsleep": {Name: "nsleep", Help: "阻塞时长", Value: "10000"},
		"nwrite": {Name: "nwrite", Help: "输出Channel长度", Value: 1},
		"outdir": {Name: "outdir", Help: "输出目录", Value: "tmp"},

		"timeout": {Name: "timeout", Help: "请求超时", Value: "10s"},
		"prefix0": {Name: "prefix0", Help: "请求前缀", Value: "uri["},
		"prefix1": {Name: "prefix1", Help: "参数前缀", Value: "request_param["},

		"_index": &ctx.Config{Name: "index", Value: []interface{}{
			map[string]interface{}{"componet_name": "status", "componet_help": "状态",
				"componet_tmpl": "componet", "componet_view": "Company", "componet_init": "",
				"componet_type": "private", "componet_ctx": "test", "componet_cmd": "diff",
				"componet_args": []interface{}{}, "inputs": []interface{}{
					map[string]interface{}{"type": "text", "name": "pod", "imports": "plugin_pod"},
					map[string]interface{}{"type": "select", "name": "sub", "values": []interface{}{"status", ""}},
					map[string]interface{}{"type": "button", "value": "执行"},
				},
			},
		}},
	},
	Commands: map[string]*ctx.Command{
		"diff": {Name: "diff file server1 server2", Form: map[string]int{"nsleep": 1}, Help: "接口对比工具", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) < 3 {
				m.Echo("usage: %v", "diff file server1 server2")
				return
			}

			var diff, same, error int32 = 0, 0, 0
			api := map[string]int{}
			mu := sync.Mutex{}

			nline, cost := Process(m, arg[0], func(msg *ctx.Message, client *http.Client, word []string, output chan []string) {
				key, uri, args := word[0], word[1], word[2]
				mu.Lock()
				api[uri]++
				mu.Unlock()

				begin := time.Now()
				res1, e1 := client.Post(arg[1]+uri, "application/json", bytes.NewReader([]byte(args)))
				t1 := time.Since(begin)

				begin = time.Now()
				res2, e2 := client.Post(arg[2]+uri, "application/json", bytes.NewReader([]byte(args)))
				t2 := time.Since(begin)

				size1, size2 := 0, 0
				result := "error"
				var t3, t4 time.Duration
				if e1 != nil || e2 != nil {
					atomic.AddInt32(&error, 1)
					fmt.Printf("%v %d cost: %v/%v error: %v %v\n", key, error, t1, t2, e1, e2)

				} else if res1.StatusCode != http.StatusOK || res2.StatusCode != http.StatusOK {
					atomic.AddInt32(&error, 1)
					fmt.Printf("%v %d %s %s cost: %v/%v error: %v %v\n", key, error, "error", uri, t1, t2, res1.StatusCode, res2.StatusCode)

				} else {
					begin = time.Now()
					b1, _ := ioutil.ReadAll(res1.Body)
					b2, _ := ioutil.ReadAll(res2.Body)
					t3 = time.Since(begin)

					begin = time.Now()
					var num int32
					if Compare(b1, b2) {
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
						output <- result1
					}
					size1 = len(b1)
					size2 = len(b2)
					t4 = time.Since(begin)
					fmt.Printf("%v %d %s %s size: %v/%v cost: %v/%v diff: %v/%v\n", key, num, result, uri, len(b1), len(b2), t1, t2, t3, t4)
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
			})

			fmt.Printf("\n\nnuri: %v nreq: %v same: %v diff: %v error: %v cost: %v\n\n", len(api), nline, same, diff, error, cost)

			m.Sort("time1", "int").Table()
			return
		}},
		"cost": {Name: "cost file server nroute", Help: "接口耗时测试", Form: map[string]int{"nwork": 1, "limit": 1, "nsleep": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			var success int32 = 0
			var times time.Duration
			mu := sync.Mutex{}
			limit := kit.Int(m.Confx("limit"))
			nline := 0
			_, cost := Process(m, arg[0], func(msg *ctx.Message, client *http.Client, word []string, output chan []string) {
				key, uri, args := word[0], word[1], word[2]
				for _, host := range arg[1:] {
					fmt.Printf("%v/%v post: %v\t%v\n", key, limit, host+uri, args)

					begin := time.Now()
					res, err := client.Post(host+uri, "application/json", bytes.NewReader([]byte(args)))
					if res.StatusCode == http.StatusOK {
						io.Copy(ioutil.Discard, res.Body)
						atomic.AddInt32(&success, 1)
					} else {
						fmt.Printf("%v/%v error: %v\n", key, limit, err)
					}
					t := time.Since(begin)
					times += t

					mu.Lock()
					nline++
					m.Add("append", "host", host)
					m.Add("append", "uri", uri)
					m.Add("append", "cost", t/1000000)
					mu.Unlock()
				}
			})

			m.Sort("cost", "int_r")
			m.Echo("\n\nnclient: %v nreq: %v success: %v time: %v average: %vms",
				m.Confx("nwork"), nline, success, cost, int(times)/int(nline)/1000000)
			return
		}},
		"cmd": {Name: "cmd file", Help: "生成测试命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			str := kit.Select("./hey -n 10000 -c 100 -m POST -T \"application/json\" -d '%s' http://127.0.0.1:6363%s\n", arg, 1)
			Process(m, arg[0], func(msg *ctx.Message, client *http.Client, word []string, output chan []string) {
				m.Echo(str, word[2], word[1])
			})
			return
		}},
	},
}

func main() {
	fmt.Print(cli.Index.Plugin(Index, os.Args[1:]))
}
