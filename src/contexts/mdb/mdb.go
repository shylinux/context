package mdb // {{{
// }}}
import ( // {{{
	"contexts"
	"strconv"
	"strings"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"fmt"
)

// }}}

type MDB struct {
	*sql.DB

	list     map[string][]string
	list_key []string

	table []string
	*ctx.Context
}

func (mdb *MDB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"source": &ctx.Cache{Name: "数据库参数", Value: "", Help: "数据库参数"},
		"driver": &ctx.Cache{Name: "数据库驱动", Value: "", Help: "数据库驱动"},
	}
	c.Configs = map[string]*ctx.Config{}

	s := new(MDB)
	s.Context = c
	return s
}

// }}}
func (mdb *MDB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	mdb.Context.Master(nil)
	if mdb.Context == Index {
		Pulse = m
	}
	return mdb
}

// }}}
func (mdb *MDB) Start(m *ctx.Message, arg ...string) bool { // {{{
	if len(arg) > 0 {
		m.Cap("source", arg[0])
	}
	if len(arg) > 1 {
		m.Cap("driver", arg[1])
	} else {
		m.Cap("driver", Pulse.Conf("driver"))
	}
	if m.Cap("source") == "" || m.Cap("driver") == "" {
		return false
	}

	db, e := sql.Open(m.Cap("driver"), m.Cap("source"))
	m.Assert(e)
	mdb.DB = db

	m.Log("info", nil, "%d open %s %s", Pulse.Capi("nsource"), m.Cap("driver"), m.Cap("stream", m.Cap("source")))
	return false
}

// }}}
func (mdb *MDB) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch mdb.Context {
	case m.Target():
		if mdb.DB != nil {
			m.Log("info", nil, "%d close %s %s", Pulse.Capi("nsource", -1)+1, m.Cap("driver"), m.Cap("source"))
			mdb.DB.Close()
			mdb.DB = nil
		}
	case m.Source():
	}
	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "mdb", Help: "数据中心",
	Caches: map[string]*ctx.Cache{
		"count":   &ctx.Cache{Name: "count", Value: "0", Help: "主机协议"},
		"nsource": &ctx.Cache{Name: "数据源数量", Value: "0", Help: "已打开数据库的数量"},
	},
	Configs: map[string]*ctx.Config{
		"driver": &ctx.Config{Name: "数据库驱动(mysql)", Value: "mysql", Help: "数据库驱动"},
	},
	Commands: map[string]*ctx.Command{
		"open": &ctx.Command{Name: "open source [driver]", Help: "打开数据库", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Assert(len(arg) > 0, "缺少参数") // {{{
			m.Start(fmt.Sprintf("db%d", Pulse.Capi("nsource", 1)), "数据存储", arg...)
			Pulse.Cap("stream", Pulse.Cap("nsource"))
			m.Echo(m.Target().Name)
			// }}}
		}},
		"exec": &ctx.Command{Name: "exec sql [arg]", Help: "操作数据库",
			Appends: map[string]string{"last": "最后插入元组的标识", "nrow": "修改元组的数量"},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				mdb, ok := m.Target().Server.(*MDB) // {{{
				m.Assert(ok, "目标模块类型错误")
				m.Assert(len(arg) > 0, "缺少参数")
				m.Assert(mdb.DB != nil, "数据库未打开")

				which := make([]interface{}, 0, len(arg))
				for _, v := range arg[1:] {
					which = append(which, v)
				}

				ret, e := mdb.Exec(arg[0], which...)
				m.Assert(e)
				id, e := ret.LastInsertId()
				m.Assert(e)
				n, e := ret.RowsAffected()
				m.Assert(e)

				m.Echo("%d", id).Echo("%d", n)
				m.Add("append", "last", fmt.Sprintf("%d", id))
				m.Add("append", "nrow", fmt.Sprintf("%d", n))
				m.Log("info", nil, "last(%d) nrow(%d)", id, n)
				// }}}
			}},
		"query": &ctx.Command{Name: "query sql [arg]", Help: "执行查询语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			mdb, ok := m.Target().Server.(*MDB) // {{{
			m.Assert(ok, "目标模块类型错误")
			m.Assert(len(arg) > 0, "缺少参数")
			m.Assert(mdb.DB != nil, "数据库未打开")

			which := make([]interface{}, 0, len(arg))
			for _, v := range arg[1:] {
				which = append(which, v)
			}

			rows, e := mdb.Query(arg[0], which...)
			m.Assert(e)
			defer rows.Close()

			cols, e := rows.Columns()
			m.Assert(e)
			num := len(cols)

			for rows.Next() {
				vals := make([]interface{}, num)
				ptrs := make([]interface{}, num)
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				rows.Scan(ptrs...)

				for i, k := range cols {
					switch b := vals[i].(type) {
					case []byte:
						m.Add("append", k, string(b))
					case int64:
						m.Add("append", k, fmt.Sprintf("%d", b))
					default:
						m.Add("append", k, "")
					}
				}
			}

			if len(m.Meta["append"]) > 0 {
				m.Log("info", nil, "rows(%d) cols(%d)", len(m.Meta[m.Meta["append"][0]]), len(m.Meta["append"]))
			} else {
				m.Log("info", nil, "rows(0) cols(0)")
			}
			// }}}
		}},
		"table": &ctx.Command{Name: "table", Help: "执行查询语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			mdb, ok := m.Target().Server.(*MDB) // {{{
			m.Assert(ok)
			msg := m.Spawn(m.Target())
			if len(arg) > 0 {
				table := arg[0]
				index, e := strconv.Atoi(arg[0])
				if e == nil && index < len(mdb.table) {
					table = mdb.table[index]
				}
				msg.Cmd("query", fmt.Sprintf("desc %s", table))

				if len(arg) > 1 {
					for i, v := range msg.Meta[msg.Meta["append"][0]] {
						if v == arg[1] {
							for _, k := range msg.Meta["append"] {
								m.Echo("%s: %s\n", k, msg.Meta[k][i])
							}
						}
					}
				} else {
					for _, v := range msg.Meta[msg.Meta["append"][0]] {
						m.Echo("%s\n", v)
					}
				}
			} else {
				msg.Cmd("query", "show tables")
				mdb.table = []string{}
				for i, v := range msg.Meta[msg.Meta["append"][0]] {
					mdb.table = append(mdb.table, v)
					m.Echo("%d: %s\n", i, v)
				}
			}
			// }}}
		}},
		"list": &ctx.Command{Name: "list add table field [where condition]", Help: "执行查询语句",
			Formats: map[string]int{"where": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				mdb, ok := m.Target().Server.(*MDB) // {{{
				m.Assert(ok)
				if len(arg) == 0 {
					for _, k := range mdb.list_key {
						m.Echo("%s: %v\n", k, mdb.list[k])
					}
					return
				}

				switch arg[0] {
				case "add":
					if mdb.list == nil {
						mdb.list = make(map[string][]string)
					}
					mdb.list[m.Cap("count")] = append(mdb.list[m.Cap("count")], m.Option("where"))
					mdb.list[m.Cap("count")] = append(mdb.list[m.Cap("count")], arg[1:]...)
					mdb.list_key = append(mdb.list_key, m.Cap("count"))
					m.Capi("count", 1)
				case "set":
					mdb.list[arg[1]] = []string{m.Option("where")}
					mdb.list[arg[1]] = append(mdb.list[arg[1]], arg[2:]...)
				default:
					if table, ok := mdb.list[arg[0]]; ok {
						msg := m.Spawn(m.Target())

						fields := strings.Join(table[2:], ",")
						condition := ""
						if len(arg) > 1 && len(arg[1]) > 0 {
							condition = fmt.Sprintf("where %s", arg[1])
						} else if len(table[0]) > 0 {
							condition = fmt.Sprintf("where %s", table[0])
						}

						other := ""
						if len(arg) > 2 {
							other = strings.Join(arg[2:], " ")
						}

						msg.Cmd("query", fmt.Sprintf("select %s from %s %s %s", fields, table[1], condition, other))
						m.Echo("%s %s\n", table[1], condition)
						for _, k := range msg.Meta["append"] {
							m.Echo("%s\t", k)
						}
						m.Echo("\n")
						for i := 0; i < len(msg.Meta[msg.Meta["append"][0]]); i++ {
							for _, k := range msg.Meta["append"] {
								m.Echo("%s\t", msg.Meta[k][i])
							}
							m.Echo("\n")
						}
					}
				}
				// }}}
			}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void",
			Commands: map[string]*ctx.Command{
				"open": &ctx.Command{},
			},
		},
	},
}

func init() {
	mdb := &MDB{}
	mdb.Context = Index
	ctx.Index.Register(Index, mdb)
}
