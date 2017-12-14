package ssh

import (
	"context"
)

type SSH struct {
	*ctx.Context
}

func (ssh *SSH) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}
	c.Commands = map[string]*ctx.Command{}

	s := new(SSH)
	s.Context = c
	return s
}

func (ssh *SSH) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return ssh
}

func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool {
	return false
}

func (ssh *SSH) Close(m *ctx.Message, arg ...string) bool {
	return false
}

var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches:   map[string]*ctx.Cache{},
	Configs:  map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
