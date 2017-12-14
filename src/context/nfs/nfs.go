package nfs

import (
	"context"
)

type NFS struct {
	*ctx.Context
}

func (nfs *NFS) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(NFS)
	s.Context = c
	return s
}

func (nfs *NFS) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return nfs
}

func (nfs *NFS) Start(m *ctx.Message, arg ...string) bool {
	return false
}

func (nfs *NFS) Exit(m *ctx.Message, arg ...string) bool {
	return false
}

var Index = &ctx.Context{Name: "nfs", Help: "存储中心",
	Caches:   map[string]*ctx.Cache{},
	Configs:  map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{},
}

func init() {
	nfs := &NFS{}
	nfs.Context = Index
	ctx.Index.Register(Index, nfs)
}
