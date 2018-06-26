package mdb // {{{
// }}}
import ( // {{{
	"contexts"
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"

	"fmt"
	"os"
	"strconv"
	"strings"
)

// }}}

type MDB struct {
	*sql.DB

	table []string
	*ctx.Context
}

func (mdb *MDB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{
		"database": &ctx.Cache{Name: "数据库", Value: m.Confx("database", arg, 0), Help: "数据库驱动"},
		"username": &ctx.Cache{Name: "用户名", Value: m.Confx("username", arg, 1), Help: "数据库驱动"},
		"password": &ctx.Cache{Name: "密码", Value: m.Confx("password", arg, 2), Help: "数据库驱动"},
		"protocol": &ctx.Cache{Name: "协议", Value: m.Confx("protocol", arg, 4), Help: "数据库驱动"},
		"address":  &ctx.Cache{Name: "地址", Value: m.Confx("address", arg, 3), Help: "数据库驱动"},
		"driver":   &ctx.Cache{Name: "数据库驱动(mysql)", Value: m.Confx("driver", arg, 5), Help: "数据库驱动"},
	}
	c.Configs = map[string]*ctx.Config{
		"table": &ctx.Config{Name: "关系表", Value: "", Help: "关系表"},
		"field": &ctx.Config{Name: "字段名", Value: "", Help: "字段名"},
		"where": &ctx.Config{Name: "条件", Value: "", Help: "条件"},
		"group": &ctx.Config{Name: "聚合", Value: "", Help: "聚合"},
		"order": &ctx.Config{Name: "排序", Value: "", Help: "排序"},
		"parse": &ctx.Config{Name: "解析", Value: "", Help: "解析"},
	}

	s := new(MDB)
	s.Context = c
	return s
}

func (mdb *MDB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	mdb.Context.Master(nil)
	if mdb.Context == Index {
		Pulse = m
	}
	return mdb
}

// }}}
func (mdb *MDB) Start(m *ctx.Message, arg ...string) bool { // {{{
	db, e := sql.Open(m.Cap("driver"), fmt.Sprintf("%s:%s@%s(%s)/%s",
		m.Cap("username"), m.Cap("password"), m.Cap("protocol"), m.Cap("address"), m.Cap("database")))
	m.Assert(e)
	mdb.DB = db

	m.Log("info", nil, "%d open %s %s", m.Capi("nsource"), m.Cap("driver"), m.Cap("stream", m.Cap("database")))
	return false
}

// }}}
func (mdb *MDB) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch mdb.Context {
	case m.Target():
		if mdb.DB != nil {
			m.Log("info", nil, "close")
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
		"nsource": &ctx.Cache{Name: "数据源数量", Value: "0", Help: "已打开数据库的数量"},
	},
	Configs: map[string]*ctx.Config{
		"database": &ctx.Config{Name: "默认数据库", Value: "demo", Help: "数据库驱动"},
		"username": &ctx.Config{Name: "默认用户名", Value: "demo", Help: "数据库驱动"},
		"password": &ctx.Config{Name: "默认密码", Value: "demo", Help: "数据库驱动"},
		"protocol": &ctx.Config{Name: "默认协议", Value: "tcp", Help: "数据库驱动"},
		"address":  &ctx.Config{Name: "默认地址", Value: "", Help: "数据库驱动"},
		"driver":   &ctx.Config{Name: "数据库驱动(mysql)", Value: "mysql", Help: "数据库驱动"},

		"dbhelp": &ctx.Config{Name: "默认帮助", Value: "数据存储", Help: "默认帮助"},
		"dbname": &ctx.Config{Name: "默认模块名", Value: "db", Help: "默认模块名", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}
			return fmt.Sprintf("%s%d", x.Value, m.Capi("nsource", 1))
			// }}}
		}},

		"csv_sep": &ctx.Config{Name: "字段分隔符", Value: "\t", Help: "字段分隔符"},
	},
	Commands: map[string]*ctx.Command{
		"open": &ctx.Command{
			Name: "open [database [username [password [address [protocol [driver]]]]]] [dbname name] [dbhelp help]",
			Help: "open打开数据库, database: 数据库名, username: 用户名, password: 密码, address: 主机地址, protocol: 主机协议, driver: 数据库类型, dbname: 模块名称, dbhelp: 帮助信息",
			Form: map[string]int{"dbname": 1, "dbhelp": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Start(m.Confx("dbname"), m.Confx("dbhelp"), arg...)
			}},
		"exec": &ctx.Command{Name: "exec sql [arg]", Help: "操作数据库, sql: SQL语句, arg: 查询参数",
			Appends: map[string]string{"last": "最后插入元组的标识", "nrow": "修改元组的数量"},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) { // {{{
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

					m.Log("info", nil, "last(%s) nrow(%s)", m.Append("last", id), m.Append("nrow", n))
					m.Echo("%d", id).Echo("%d", n)
				}
				// }}}
			}},
		"query": &ctx.Command{Name: "query sql [arg]", Help: "查询数据库, sql: SQL语句, arg: 查询参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) { // {{{
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
							m.Add("append", k, fmt.Sprintf("%v", b))
						}
					}
				}

				if len(m.Meta["append"]) > 0 {
					m.Log("info", nil, "rows(%d) cols(%d)", len(m.Meta[m.Meta["append"][0]]), len(m.Meta["append"]))
				} else {
					m.Log("info", nil, "rows(0) cols(0)")
				}
			}
			// }}}
		}},
		"table": &ctx.Command{Name: "table [which [field]]", Help: "查看关系表信息，which: 表名, field: 字段名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) { // {{{
				msg := m.Spawn()
				if len(arg) == 0 {
					msg.Cmd("query", "show tables")
					mdb.table = []string{}
					for i, v := range msg.Meta[msg.Meta["append"][0]] {
						mdb.table = append(mdb.table, v)
						m.Echo("%d: %s\n", i, v)
					}
					return
				}

				table := arg[0]
				index, e := strconv.Atoi(arg[0])
				if e == nil && index < len(mdb.table) {
					table = mdb.table[index]
				}

				msg.Cmd("query", fmt.Sprintf("desc %s", table))
				if len(arg) == 1 {
					for _, v := range msg.Meta[msg.Meta["append"][0]] {
						m.Echo("%s\n", v)
					}
					return
				}

				for i, v := range msg.Meta[msg.Meta["append"][0]] {
					if v == arg[1] {
						for _, k := range msg.Meta["append"] {
							m.Echo("%s: %s\n", k, msg.Meta[k][i])
						}
					}
				}
			}
			// }}}
		}},
		"show": &ctx.Command{
			Name: "show table fields... [where conditions] [group fields] [order fields] [save filename]",
			Help: "查询数据库, table: 表名, fields: 字段, where: 查询条件, group: 聚合字段, order: 排序字段",
			Form: map[string]int{"where": 1, "group": 1, "order": 1, "extras": 1, "save": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if m.Options("extras") { // {{{
					arg = append(arg, "extra")
				}
				fields := strings.Join(arg[1:], ",")

				where := m.Optionx("where", "where %s")
				group := m.Optionx("group", "group by %s")
				order := m.Optionx("order", "order by %s")

				msg := m.Spawn().Cmd("query", fmt.Sprintf("select %s from %s %s %s %s", fields, arg[0], where, group, order))
				if !m.Options("save") {
					m.Echo("\033[31m%s\033[0m %s %s %s\n", arg[0], where, group, order)
				}
				msg.Table(func(maps map[string]string, lists []string, index int) bool {
					for i, v := range lists {
						if m.Options("save") {
							m.Echo(maps[msg.Meta["append"][i]])
						} else if index == -1 {
							m.Echo("\033[32m%s\033[0m", v)
						} else {
							m.Echo(v)
						}
						if i < len(lists)-1 {
							m.Echo(m.Conf("csv_sep"))
						}
					}
					m.Echo("\n")
					return true
				})

				if m.Options("save") {
					f, e := os.Create(m.Option("save"))
					m.Assert(e)
					defer f.Close()

					for _, v := range m.Meta["result"] {
						f.WriteString(v)
					}
				}
				// }}}
			}},
		"get": &ctx.Command{Name: "get [where str] [parse str] [table [field]]", Help: "执行查询语句",
			Form: map[string]int{"where": 1, "parse": 2},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				where := m.Confx("where") // {{{
				if where != "" {
					where = "where " + where
				}

				parse := m.Confx("parse", m.Option("parse"))
				extra := m.Confx("extra", m.Meta["parse"], 1)
				table := m.Confx("table", arg, 0)
				field := m.Confx("field", arg, 1)

				msg := m.Spawn().Cmd("query", fmt.Sprintf("select %s from %s %s", field, table, where))
				msg.Table(func(row map[string]string, lists []string, index int) bool {
					if index == -1 {
						return true
					}
					data := map[string]interface{}{}
					switch parse {
					case "json":
						if json.Unmarshal([]byte(row[field]), &data); extra == "" {
							for k, v := range data {
								m.Echo("%s: %v\n", k, v)
							}
						} else if v, ok := data[extra]; ok {
							m.Echo("%v", v)
						}
					default:
						m.Echo("%v", row[field])
					}
					return false
				}) // }}}
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
