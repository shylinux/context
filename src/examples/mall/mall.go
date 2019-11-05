package mall

import (
	"contexts/ctx"
	"contexts/web"
	"toolkit"
)

var Index = &ctx.Context{Name: "mall", Help: "交易中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"salary": {Name: "salary table month total base tax", Help: "工资", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) < 3 || arg[2] == "" {
				m.Cmdy("ssh.data", "show", arg[0], "fields", "id", "month",
					"total", "base", "zhu", "old", "bad", "mis", "tax", "rest")
				return
			}

			total := kit.Int(arg[2])
			base := kit.Int(kit.Select(arg[2], arg, 3))
			tax := kit.Int(kit.Select("0", arg, 4))

			zhu := base * 120 / 1000
			if len(arg) > 5 {
				zhu = kit.Int(arg[5])
			}
			old := base * 80 / 1000
			if len(arg) > 6 {
				old = kit.Int(arg[6])
			}
			bad := base * 20 / 1000
			if len(arg) > 7 {
				bad = kit.Int(arg[7])
			}
			mis := base * 5 / 1000
			if len(arg) > 8 {
				mis = kit.Int(arg[8])
			}

			rest := total - zhu - old - bad - mis - tax
			m.Cmdy("ssh.data", "insert", arg[0], "month", arg[1], "total", total, "base", base, "tax",
				tax, "zhu", zhu, "old", old, "bad", bad, "mis", mis, "rest", rest)
			return
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.WEB{Context: Index})
}
