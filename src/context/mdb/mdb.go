package mdb // {{{
// }}}
import ( // {{{
	"context"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"fmt"
)

// }}}

type MDB struct {
	*sql.DB
	*ctx.Context
}

func (mdb *MDB) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"source": &ctx.Cache{Name: "数据库参数", Value: "", Help: "数据库参数"},
		"driver": &ctx.Cache{Name: "数据库驱动", Value: "", Help: "数据库驱动"},
	}
	c.Configs = map[string]*ctx.Config{}

	if len(arg) > 0 {
		m.Cap("source", arg[0])
	}
	if len(arg) > 1 {
		m.Cap("driver", arg[1])
	} else {
		m.Cap("driver", m.Conf("driver"))
	}

	s := new(MDB)
	s.Context = c
	return s
}

// }}}
func (mdb *MDB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
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
		m.Cap("driver", m.Conf("driver"))
	}

	if m.Cap("source") == "" || m.Cap("driver") == "" {
		return false
	}

	db, e := sql.Open(m.Cap("driver"), m.Cap("source"))
	m.Assert(e)
	mdb.DB = db

	m.Log("info", "%s: %d open %s %s", mdb.Name, m.Capi("nsource", 1), m.Cap("driver"), m.Cap("source"))
	return false
}

// }}}
func (mdb *MDB) Close(m *ctx.Message, arg ...string) bool { // {{{
	if mdb.DB != nil && m.Target == mdb.Context {
		m.Log("info", "%s: %d close %s %s", mdb.Name, m.Capi("nsource", -1)+1, m.Cap("driver"), m.Cap("source"))
		mdb.DB.Close()
		mdb.DB = nil
		return true
	}

	return false
}

// }}}

var Index = &ctx.Context{Name: "mdb", Help: "内存数据库",
	Caches: map[string]*ctx.Cache{
		"nsource": &ctx.Cache{Name: "数据源数量", Value: "0", Help: "已打开数据库的数量"},
	},
	Configs: map[string]*ctx.Config{
		"driver": &ctx.Config{Name: "数据库驱动(mysql)", Value: "mysql", Help: "数据库驱动"},
	},
	Commands: map[string]*ctx.Command{
		"open": &ctx.Command{Name: "open name help [source [driver]]", Help: "打开数据库", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			m.Target = m.Master // {{{
			m.Start(arg[0], arg[1], arg[2:]...)
			return ""
			// }}}
		}},
		"exec": &ctx.Command{Name: "exec sql [arg]", Help: "执行操作语句",
			Appends: map[string]string{"LastInsertId": "最后插入元组的标识", "RowsAffected": "修改元组的数量"},
			Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
				mdb, ok := m.Target.Server.(*MDB) // {{{
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

				m.Add("append", "LastInsertId", fmt.Sprintf("%d", id))
				m.Add("append", "RowsAffected", fmt.Sprintf("%d", n))
				m.Log("info", "%s: last(%d) rows(%d)", m.Target.Name, id, n)
				return ""
				// }}}
			}},
		"query": &ctx.Command{Name: "query sql [arg]", Help: "执行查询语句", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			mdb, ok := m.Target.Server.(*MDB) // {{{
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

			m.Log("info", "%s: cols(%d) rows(%d)", m.Target.Name, len(m.Meta["append"]), len(m.Meta[m.Meta["append"][0]]))
			return ""
			// }}}
		}},
		"close": &ctx.Command{Name: "close name", Help: "关闭数据库", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			msg := m.Find(arg[0], m.Master) // {{{
			msg.Target.Close(msg)
			return ""
			// }}}
		}},
	},
}

func init() {
	mdb := &MDB{}
	mdb.Context = Index
	ctx.Index.Register(Index, mdb)
}
