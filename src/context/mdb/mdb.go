package mdb // {{{
// }}}
import ( // {{{
	"context"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"errors"
	"fmt"
	"log"
)

// }}}

type MDB struct {
	db *sql.DB
	*ctx.Context
}

func (mdb *MDB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	mdb.Configs["source"] = &ctx.Config{Name: "source", Value: "", Help: "数据库参数"}
	mdb.Configs["driver"] = &ctx.Config{Name: "driver", Value: "", Help: "数据库驱动"}

	return mdb
}

// }}}
func (mdb *MDB) Start(m *ctx.Message, arg ...string) bool { // {{{
	mdb.Capi("nsource", 1)
	defer mdb.Capi("nsource", -1)

	if len(arg) > 0 {
		mdb.Conf("source", arg[0])

		if len(arg) > 1 {
			mdb.Conf("driver", arg[1])
		}
	}

	if mdb.Conf("source") == "" || mdb.Conf("driver") == "" {
		return true
	}

	db, e := sql.Open(mdb.Conf("driver"), mdb.Conf("source"))
	mdb.Assert(e)
	mdb.db = db
	defer mdb.db.Close()

	log.Println(mdb.Name, "open:", mdb.Conf("driver"), mdb.Conf("source"))
	defer log.Println(mdb.Name, "close:", mdb.Conf("driver"), mdb.Conf("source"))

	for _, p := range m.Meta["prepare"] {
		_, e := db.Exec(p)
		mdb.Assert(e)
	}

	return true
}

// }}}
func (mdb *MDB) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(MDB)
	s.Context = c
	return s
}

// }}}
func (mdb *MDB) Exit(m *ctx.Message, arg ...string) bool { // {{{
	return true
}

// }}}

var Index = &ctx.Context{Name: "mdb", Help: "内存数据库",
	Caches: map[string]*ctx.Cache{
		"nsource": &ctx.Cache{Name: "数据源数量", Value: "0", Help: "数据库连接的数量"},
	},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"open": &ctx.Command{Name: "open [source [driver]]", Help: "打开数据库",
			Options: map[string]string{
				"prepare": "打开数据库时自动执行的语句",
			},
			Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
				m.Start("db"+c.Cap("nsource"), arg...) // {{{
				return ""
				// }}}
			}},
		"exec": &ctx.Command{Name: "exec sql [arg]", Help: "执行SQL语句",
			Appends: map[string]string{
				"LastInsertId": "最后插入元组的标识",
				"RowsAffected": "修改元组的数量",
			},
			Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
				mdb, ok := m.Target.Server.(*MDB) // {{{
				if !ok {
					m.Assert(errors.New("目标模块类型错误"))
				}
				if len(arg) == 0 {
					m.Assert(errors.New("缺少参数"))
				}

				which := make([]interface{}, 0, len(arg))
				for _, v := range arg[1:] {
					which = append(which, v)
				}

				ret, e := mdb.db.Exec(arg[0], which...)
				m.Assert(e)

				id, e := ret.LastInsertId()
				m.Assert(e)
				n, e := ret.RowsAffected()
				m.Assert(e)

				m.Add("append", "LastInsertId", fmt.Sprintf("%d", id))
				m.Add("append", "RowsAffected", fmt.Sprintf("%d", n))
				return ""
				// }}}
			}},
		"query": &ctx.Command{Name: "query sql [arg]", Help: "执行SQL语句", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			mdb, ok := m.Target.Server.(*MDB) // {{{
			if !ok {
				m.Assert(errors.New("目标模块类型错误"))
			}
			if len(arg) == 0 {
				m.Assert(errors.New("缺少参数"))
			}

			which := make([]interface{}, 0, len(arg))
			for _, v := range arg[1:] {
				which = append(which, v)
			}

			rows, e := mdb.db.Query(arg[0], which...)
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
					}
				}
			}

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
