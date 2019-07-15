package main

import (
	"contexts/cli"
	"contexts/ctx"
	_ "contexts/nfs"
	"toolkit"

	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
)

func Merge(left []int, right []int) []int {
	result := make([]int, 0, len(left)+len(right))
	for len(left) > 0 || len(right) > 0 {
		if len(right) == 0 || (len(left) > 0 && (left[0] < right[0])) {
			result, left = append(result, left[0]), left[1:]
		} else {
			result, right = append(result, right[0]), right[1:]
		}
	}
	return result
}
func MergeSort(m *ctx.Message, level int, data []int) []int {
	if len(data) < 2 {
		m.Add("append", kit.Format(level), fmt.Sprintf("[][]"))
		return data
	}
	middle := len(data) / 2
	m.Add("append", kit.Format(level), fmt.Sprintf("%v%v", data[:middle], data[middle:]))
	return Merge(MergeSort(m, level+1, data[:middle]), MergeSort(m, level+1, data[middle:]))
}
func QuickSort(m *ctx.Message, level int, data []int, left int, right int) {
	if left >= right {
		return
	}

	p, l, r := left, left+1, right
	for l < r {
		for ; p < r && data[p] < data[r]; r-- {
		}
		if p < r {
			p, data[p], data[r] = r, data[r], data[p]
		}
		for ; l < p && data[l] < data[p]; l++ {
		}
		if l < p {
			p, data[l], data[p] = l, data[p], data[l]
		}
	}

	m.Add("append", kit.Format(level), fmt.Sprintf("%v%v", data[left:p+1], data[p+1:right+1]))
	QuickSort(m, level+1, data, left, p)
	QuickSort(m, level+1, data, p+1, right)
}

var Index = &ctx.Context{Name: "sort", Help: "sort code",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"data": &ctx.Config{Name: "data", Value: map[string]interface{}{
			"seed": []int{47, 59, 81, 40, 56, 0, 94, 11, 18, 25},
		}},
		"_index": &ctx.Config{Name: "index", Value: []interface{}{
			map[string]interface{}{"componet_name": "select", "componet_help": "选择排序",
				"componet_tmpl": "componet", "componet_view": "componet", "componet_init": "",
				"componet_type": "public", "componet_ctx": "sort", "componet_cmd": "select",
				"componet_args": []interface{}{}, "inputs": []interface{}{
					map[string]interface{}{"type": "text", "name": "pod", "imports": "plugin_pod"},
					map[string]interface{}{"type": "button", "value": "执行"},
				},
			},
		}},
	},
	Commands: map[string]*ctx.Command{
		"_init": &ctx.Command{Name: "_init", Help: "_init", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 0 {
				keys := []string{}
				for key, _ := range c.Commands {
					if !strings.HasPrefix(key, "_") && !strings.HasPrefix(key, "/") {
						keys = append(keys, key)
					}
				}

				list := []interface{}{}
				for _, key := range keys {
					cmd := c.Commands[key]

					list = append(list, map[string]interface{}{"componet_name": cmd.Name, "componet_help": cmd.Help,
						"componet_tmpl": "componet", "componet_view": "", "componet_init": "",
						"componet_ctx": "cli." + arg[0], "componet_cmd": key,
						"componet_args": []interface{}{"@text", "@total"}, "inputs": []interface{}{
							map[string]interface{}{"type": "input"},
							map[string]interface{}{"type": "button", "value": "show"},
						},
					})
				}

				m.Confv("ssh.componet", arg[0], list)
			}
			return
		}},
		"data": &ctx.Command{Name: "data", Help: "data", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			data := []int{}
			for i := 0; i < kit.Int(kit.Select("10", arg, 0)); i++ {
				data = append(data, rand.Intn(kit.Int(kit.Select("100", arg, 1))))
			}
			m.Confv("data", "seed", data)
			m.Echo("data: %v", data)
			return
		}},
		"select": &ctx.Command{Name: "select", Help: "select", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			data := append([]int{}, m.Confv("data", "seed").([]int)...)

			m.Echo("data: %v\n", data)
			for i := 0; i < len(data)-1; i++ {
				for j := i + 1; j < len(data); j++ {
					if data[j] < data[i] {
						data[i], data[j] = data[j], data[i]
					}
				}
				m.Echo("data: %d %v    %v\n", i, data[:i+1], data[i+1:])
			}
			m.Echo("data: %v\n", data)
			return
		}},
		"insert": &ctx.Command{Name: "insert", Help: "insert", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			data := append([]int{}, m.Confv("data", "seed").([]int)...)

			m.Echo("data: %v\n", data)
			for i, j := 1, 0; i < len(data); i++ {
				tmp := data[i]
				for j = i - 1; j >= 0; j-- {
					if data[j] < tmp {
						break
					}
					data[j+1] = data[j]
				}
				data[j+1] = tmp
				m.Echo("data: %d %v    %v\n", i, data[:i+1], data[i+1:])
			}
			m.Echo("data: %v\n", data)
			return
		}},
		"bubble": &ctx.Command{Name: "bubble", Help: "bubble", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			data := append([]int{}, m.Confv("data", "seed").([]int)...)
			m.Echo("data: %v\n", data)
			for i := 1; i < len(data); i++ {
				finish := true
				for j := 0; j < len(data)-i; j++ {
					if data[j] > data[j+1] {
						finish, data[j], data[j+1] = false, data[j+1], data[j]
					}
				}
				if finish {
					break
				}
				m.Echo("data: %d %v    %v\n", i, data[:len(data)-i], data[len(data)-i:])
			}
			m.Echo("data: %v\n", data)
			return
		}},
		"quick": &ctx.Command{Name: "quick", Help: "quick", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			data := append([]int{}, m.Confv("data", "seed").([]int)...)
			m.Echo("data: %v\n", data)
			QuickSort(m, 0, data, 0, len(data)-1)
			for i := 0; i < len(data); i++ {
				meta, ok := m.Meta[kit.Format(i)]
				if !ok {
					break
				}
				m.Echo("data: %v %v\n", i, meta)
			}
			m.Echo("data: %v\n", data)
			return
		}},
		"merge": &ctx.Command{Name: "merge", Help: "merge", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			data := append([]int{}, m.Confv("data", "seed").([]int)...)
			m.Echo("data: %v\n", data)
			data = MergeSort(m, 0, data)
			for i := 0; int(math.Exp2(float64(i))) < len(data); i++ {
				meta, ok := m.Meta[kit.Format(i)]
				if !ok {
					break
				}
				m.Echo("data: %v %v\n", i, meta)
			}
			m.Echo("data: %v\n", data)
			return
		}},
	},
}

func main() {
	fmt.Print(cli.Index.Plugin(Index, os.Args[1:]))
}
