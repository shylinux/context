package ctx

import (
	"fmt"
	"testing"
)

func TestCtx(t *testing.T) {
	context := Context{Name: "root", Help: "默认",
		Caches: map[string]*Cache{
			"nclient": &Cache{"nclient", "10", "连接数量", func(c *Context) string {
				return "10"
			}},
		},
		Configs: map[string]*Config{
			"limit": &Config{"limit", "10", "最大连接数", func(c *Context, arg string) {
			}},
		},
		Commands: map[string]*Command{
			"session": &Command{"session", "会话管理", func(c *Context, arg ...string) string {
				return "ok"
			}},
		},
	}
	context.Index = map[string]*Context{"root": &context}

	ctx := context.Index["root"]
	ctx.Add("context", "hi", "hi", "nice")
	if _, ok := context.Contexts["hi"]; !ok {
		t.Fatal("root.ctxs add error")
	}
	if c, ok := context.Index["hi"]; ok {
		ctx.Add("command", "hi", "session")
		if _, ok := c.Commands["session"]; ok {
			if c.Cmd("session") != "ok" {
				t.Fatal("hi.cmds.session: run error")
			}
		} else {
			t.Fatal("hi.cmds: add error")
		}

		ctx.Add("config", "hi", "limit")
		ctx.Add("cache", "hi", "nclient")

	} else {
		t.Fatal("root.index add error")
	}
	return

	if true {
		ctx := context.Index["hi"]
		if ctx.Cmd("session", "nice") == "ok" {
			t.Log("hi.cmds.session: run")
		} else {
			t.Fatal("hi.cmds.session: run error")
		}
		t.Log()

		ctx.Add("context", "he", "he", "nice")
		for k, v := range ctx.Contexts {
			t.Log("hi.ctxs", k, v.Name, v.Help)
		}
		if len(ctx.Contexts) != 1 {
			t.Fatal("hi.ctxs: add error")
		}
		for k, v := range ctx.Index {
			t.Log("hi.index:", k, v.Name, v.Help)
		}
		if len(ctx.Index) != 3 {
			t.Fatal("hi.index: add error")
		}
		t.Log()

		ctx.Add("command", "he", "session")
		for k, v := range ctx.Contexts["he"].Commands {
			t.Log("he.cmds:", k, v.Name, v.Help)
		}
		if len(ctx.Contexts["he"].Commands) != 1 {
			t.Fatal("he.cmds: add error")
		}

	}

	for k, v := range ctx.Index {
		t.Log("root.index:", k, v.Name, v.Help)
	}
	if len(ctx.Index) != 3 {
		t.Fatal("root.index add error")
	}
	t.Log()

	for k, v := range ctx.Index {
		t.Log(fmt.Sprintf("root.index.%s.cmds: %d", k, len(v.Commands)))
	}
	t.Log()
	ctx.Del("command", "hi", "session")
	for k, v := range ctx.Index {
		t.Log(fmt.Sprintf("root.index.%s.cmds: %d", k, len(v.Commands)))
	}
	t.Log()

	ctx.Del("context", "hi")
	for k, v := range ctx.Contexts {
		t.Log("root.ctxs:", k, v.Name, v.Help)
	}
	if len(ctx.Contexts) != 0 {
		t.Fatal("root.ctxs: del error")
	}
	for k, v := range ctx.Index {
		t.Log("root.index:", k, v.Name, v.Help)
	}
	if len(ctx.Index) != 1 {
		t.Fatal("root.index del error")
	}
}
