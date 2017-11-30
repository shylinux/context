package ssh

import (
	"context"
	_ "context/cli"
)

type SSH struct {
	*ctx.Context
}

func (ssh *SSH) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return ssh
}

func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool {
	return true
}

func (ssh *SSH) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}
	c.Commands = map[string]*ctx.Command{}

	s := new(SSH)
	s.Context = c
	return s
}

func (ssh *SSH) Close(m *ctx.Message, arg ...string) bool {
	return true
}

var Index = &ctx.Context{Name: "ssh", Help: "加密终端",
	Caches:   map[string]*ctx.Cache{},
	Configs:  map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
