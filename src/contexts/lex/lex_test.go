package lex

import (
	"context"
	"testing"
)

func TestLEX(t *testing.T) {

	m := ctx.Pulse.Spawn(Index)
	seed := map[string]map[string]string{
		// "shy?": map[string]string{
		// 	"s":     "",
		// 	"sh":    "sh",
		// 	"she":   "sh",
		// 	"shy":   "shy",
		// 	"shyyy": "shy",
		// },
		// "shy*": map[string]string{
		// 	"s":     "",
		// 	"sh":    "sh",
		// 	"she":   "sh",
		// 	"shy":   "shy",
		// 	"shyyy": "shyyy",
		// },
		// "shy+": map[string]string{
		// 	"s":     "",
		// 	"sh":    "",
		// 	"she":   "",
		// 	"shy":   "shy",
		// 	"shyyy": "shyyy",
		// },
		// "s?hy": map[string]string{
		// 	"s":   "",
		// 	"sh":  "",
		// 	"she": "",
		// 	"shy": "shy",
		// 	"hy":  "hy",
		// },
		// "s*hy": map[string]string{
		// 	"s":     "",
		// 	"sh":    "",
		// 	"she":   "",
		// 	"shy":   "shy",
		// 	"ssshy": "ssshy",
		// 	"hy":    "hy",
		// },
		// "s+hy": map[string]string{
		// 	"s":     "",
		// 	"sh":    "",
		// 	"she":   "",
		// 	"shy":   "shy",
		// 	"ssshy": "ssshy",
		// 	"hy":    "",
		// },
		// "sh[xyz]?": map[string]string{
		// 	"s":     "",
		// 	"sh":    "sh",
		// 	"she":   "sh",
		// 	"shy":   "shy",
		// 	"shyyy": "shy",
		// },
		// "sh[xyz]*": map[string]string{
		// 	"s":     "",
		// 	"sh":    "sh",
		// 	"she":   "sh",
		// 	"shy":   "shy",
		// 	"shyyy": "shyyy",
		// 	"shyxz": "shyxz",
		// },
		// "sh[xyz]+": map[string]string{
		// 	"s":      "",
		// 	"sh":     "",
		// 	"she":    "",
		// 	"shy":    "shy",
		// 	"shyyy":  "shyyy",
		// 	"shyxzy": "shyxzy",
		// },
		// "[xyz]?sh": map[string]string{
		// 	"s":      "",
		// 	"sh":     "sh",
		// 	"zsh":    "zsh",
		// 	"zxyshy": "",
		// },
		// "[xyz]*sh": map[string]string{
		// 	"s":      "",
		// 	"sh":     "sh",
		// 	"zsh":    "zsh",
		// 	"zxyshy": "zxysh",
		// },
		// "[xyz]+sh": map[string]string{
		// 	"s":      "",
		// 	"sh":     "",
		// 	"zsh":    "zsh",
		// 	"zxyshy": "zxysh",
		// },
		// "[0-9]+": map[string]string{
		// 	"hello": "",
		// 	"hi123": "",
		// 	"123":   "123",
		// 	"123hi": "123",
		// },
		// "0x[0-9a-fA-F]+": map[string]string{
		// 	"hello":     "",
		// 	"0xhi123":   "",
		// 	"0x123":     "0x123",
		// 	"0xab123ab": "0xab123ab",
		// 	"0x123ab":   "0x123ab",
		// },
		"[a-zA-Z][a-zA-Z0-9]*": map[string]string{
			"hello": "hello",
			"hi123": "hi123",
			"123":   "",
		},
		"\"[^\"]*\"": map[string]string{
			"hello":      "",
			"0xhi123":    "",
			"\"hi\"":     "\"hi\"",
			"\"\\\"hi\"": "\"\\\"hi\"",
		},
	}
	m.Conf("debug", "on")
	Index.Begin(m)
	for k, s := range seed {
		Index.Start(m)
		m.Cmd("train", k)
		for i, v := range s {
			if m.Cmd("parse", i) != v {
				t.Error("train&parse error:", k, i, v)
			}
		}
	}

	Index.Start(m)
	m.Cmd("train", "[ \n\t]+", "1")
	m.Cmd("train", "[a-zA-Z][a-zA-Z0-9]*", "2", "2")
	m.Cmd("train", "0x[0-9]+", "3", "2")
	m.Cmd("train", "[0-9]+", "3", "2")
	m.Cmd("train", "\"[^\"]*\"", "4", "2")
	m.Cmd("train", "'[^']*'", "4", "2")

	lex := Index.Server.(*LEX)
	for _, v := range lex.seed {
		t.Log(v.page, v.hash, v.word)
	}

	m.Cmd("split", "hello 0x2134 \"hi he\" meet\\ you")
	// m.Cmd("parse", "0x54 nice to meet")
	// m.Cmd("parse", "737 nice to meet")
	// m.Cmd("parse", "\"73 u\" nice to meet")
	// m.Cmd("parse", "'hh h' nice to meet")
}
