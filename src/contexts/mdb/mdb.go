package mdb // {{{
// }}}
import ( // {{{
	"bufio"
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

	db    []string
	table []string

	*ctx.Context
}

func (mdb *MDB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{
		"database": &ctx.Cache{Name: "数据库", Value: m.Confx("database", arg, 0), Help: "数据库"},
		"username": &ctx.Cache{Name: "账户", Value: m.Confx("username", arg, 1), Help: "账户"},
		"password": &ctx.Cache{Name: "密码", Value: m.Confx("password", arg, 2), Help: "密码"},
		"address":  &ctx.Cache{Name: "服务地址", Value: m.Confx("address", arg, 3), Help: "服务地址"},
		"protocol": &ctx.Cache{Name: "服务协议(tcp)", Value: m.Confx("protocol", arg, 4), Help: "服务协议"},
		"driver":   &ctx.Cache{Name: "数据库驱动(mysql)", Value: m.Confx("driver", arg, 5), Help: "数据库驱动"},
	}
	c.Configs = map[string]*ctx.Config{
		"table":  &ctx.Config{Name: "关系表", Value: "", Help: "关系表"},
		"field":  &ctx.Config{Name: "字段名", Value: "", Help: "字段名"},
		"where":  &ctx.Config{Name: "条件", Value: "", Help: "条件"},
		"group":  &ctx.Config{Name: "聚合", Value: "", Help: "聚合"},
		"order":  &ctx.Config{Name: "排序", Value: "", Help: "排序"},
		"limit":  &ctx.Config{Name: "分页", Value: "", Help: "分页"},
		"offset": &ctx.Config{Name: "偏移", Value: "", Help: "偏移"},
		"parse":  &ctx.Config{Name: "解析", Value: "", Help: "解析"},
	}

	s := new(MDB)
	s.Context = c
	return s
}

func (mdb *MDB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
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
	m.Log("info", "mdb open %s", m.Cap("database"))
	return false
}

// }}}
func (mdb *MDB) Close(m *ctx.Message, arg ...string) bool { // {{{
	return false
	switch mdb.Context {
	case m.Target():
		if mdb.DB != nil {
			m.Log("info", "mdb close %s", m.Cap("database"))
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
		"database": &ctx.Config{Name: "默认数据库", Value: "demo", Help: "默认数据库"},
		"username": &ctx.Config{Name: "默认用户名", Value: "demo", Help: "默认用户名"},
		"password": &ctx.Config{Name: "默认密码", Value: "demo", Help: "默认密码"},
		"protocol": &ctx.Config{Name: "默认协议", Value: "tcp", Help: "默认协议"},
		"address":  &ctx.Config{Name: "默认地址", Value: "", Help: "默认地址"},
		"driver":   &ctx.Config{Name: "数据库驱动(mysql)", Value: "mysql", Help: "数据库驱动"},

		"dbhelp": &ctx.Config{Name: "默认帮助", Value: "数据存储", Help: "默认帮助"},
		"dbname": &ctx.Config{Name: "默认模块名", Value: "db", Help: "默认模块名", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}
			return fmt.Sprintf("%s%d", x.Value, m.Capi("nsource", 1))
			// }}}
		}},

		"csv_col_sep": &ctx.Config{Name: "字段分隔符", Value: "\t", Help: "字段分隔符"},
		"csv_row_sep": &ctx.Config{Name: "记录分隔符", Value: "\n", Help: "记录分隔符"},
	},
	Commands: map[string]*ctx.Command{
		"open": &ctx.Command{
			Name: "open [database [username [password [address [protocol [driver]]]]]] [dbname name] [dbhelp help]",
			Help: "open打开数据库, database: 数据库名, username: 用户名, password: 密码, address: 服务地址, protocol: 服务协议, driver: 数据库类型, dbname: 模块名称, dbhelp: 帮助信息",
			Form: map[string]int{"dbname": 1, "dbhelp": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Start(m.Confx("dbname"), m.Confx("dbhelp"), arg...)
			}},
		"exec": &ctx.Command{Name: "exec sql [arg]", Help: "操作数据库, sql: SQL语句, arg: 操作参数",
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

					m.Log("info", "last(%s) nrow(%s)", m.Append("last", id), m.Append("nrow", n))
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
					m.Log("info", "rows(%d) cols(%d)", len(m.Meta[m.Meta["append"][0]]), len(m.Meta["append"]))
				} else {
					m.Log("info", "rows(0) cols(0)")
				}
			}
			// }}}
		}},
		"db": &ctx.Command{Name: "db [which]", Help: "查看或选择数据库信息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) { // {{{
				if len(arg) == 0 {
					msg := m.Spawn().Cmd("query", "show databases")
					mdb.db = []string{}
					for i, v := range msg.Meta[msg.Meta["append"][0]] {
						mdb.db = append(mdb.db, v)
						m.Echo("%d: %s\n", i, v)
					}
					return
				}

				db := arg[0]
				if i, e := strconv.Atoi(arg[0]); e == nil && i < len(mdb.db) {
					db = mdb.db[i]
				}
				m.Assert(m.Spawn().Cmd("exec", fmt.Sprintf("use %s", db)))
				m.Cap("database", db)
			}
			// }}}
		}},
		"table": &ctx.Command{Name: "table [which [field]]", Help: "查看关系表信息，which: 表名, field: 字段名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) { // {{{
				if len(arg) == 0 {
					msg := m.Spawn().Cmd("query", "show tables")
					mdb.table = []string{}
					for i, v := range msg.Meta[msg.Meta["append"][0]] {
						mdb.table = append(mdb.table, v)
						m.Echo("%d: %s\n", i, v)
					}
					return
				}

				table := arg[0]
				if i, e := strconv.Atoi(arg[0]); e == nil && i < len(mdb.table) {
					table = mdb.table[i]
				}

				msg := m.Spawn().Cmd("query", fmt.Sprintf("desc %s", table))
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
			Name: "show table fields... [where conditions] [group fields] [order fields] [limit fields] [offset fields] [save filename] [other rest...]",
			Help: "查询数据库, table: 表名, fields: 字段, where: 查询条件, group: 聚合字段, order: 排序字段",
			Form: map[string]int{"where": 1, "group": 1, "order": 1, "limit": 1, "offset": 1, "extras": 1, "extra_field": 1, "extra_format": 1, "trans_field": 1, "trans_map": 2, "save": 1, "export": 2, "other": -1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) { // {{{
					table := m.Confx("table", arg, 0)
					if i, e := strconv.Atoi(table); e == nil && i < len(mdb.table) {
						table = mdb.table[i]
					}

					fields := []string{"*"}
					if len(arg) > 1 {
						fields = arg[1:]
					} else if m.Confs("field") {
						fields = []string{m.Conf("field")}
					}
					field := strings.Join(fields, ",")

					where := m.Confx("where", m.Option("where"), "where %s")
					group := m.Confx("group", m.Option("group"), "group by %s")
					order := m.Confx("order", m.Option("order"), "order by %s")
					limit := m.Confx("limit", m.Option("limit"), "limit %s")
					offset := m.Confx("offset", m.Option("offset"), "offset %s")

					other := m.Meta["other"]
					for i, v := range other {
						if len(v) > 1 {
							switch v[0] {
							case '$':
								other[i] = m.Cap(v[1:])
							case '@':
								other[i] = m.Conf(v[1:])
							}
						}
					}

					msg := m.Spawn().Cmd("query", fmt.Sprintf("select %s from %s %s %s %s %s %s", field, table, where, group, order, limit, offset), other)
					m.Copy(msg, "append")
					if m.Optioni("query", msg.Code()); !m.Options("save") {
						m.Color(31, table).Echo(" %s %s %s %s %s %v\n", where, group, order, limit, offset, m.Meta["other"])
					}

					m.Table(func(maps map[string]string, lists []string, line int) bool {
						for i, v := range lists {
							if m.Options("save") {
								key := m.Meta["append"][i]
								value := maps[key]
								if key == m.Option("trans_field") {
									for i := 0; i < len(m.Meta["trans_map"])-1; i += 2 {
										if value == m.Meta["trans_map"][i] {
											value = m.Meta["trans_map"][i+1]
											break
										}
									}
								}

								m.Echo(value)
							} else if line == -1 {
								m.Color(32, v)
							} else {
								m.Echo(v)
							}
							if i < len(lists)-1 {
								m.Echo(m.Conf("csv_col_sep"))
							}
						}
						m.Echo(m.Conf("csv_row_sep"))
						return true
					})

					if m.Options("export") {
						f, e := os.Create(m.Option("export"))
						m.Assert(e)
						defer f.Close()

						m.Table(func(maps map[string]string, lists []string, line int) bool {
							if line > -1 {
								args := []interface{}{}
								for i := 0; i < len(lists); i++ {
									args = append(args, maps[m.Meta["append"][i]])
								}
								f.WriteString(fmt.Sprintf(m.Meta["export"][1], args...))
							}
							return true
						})
					}

					if m.Options("save") {
						f, e := os.Create(m.Option("save"))
						m.Assert(e)
						defer f.Close()

						for _, v := range m.Meta["result"] {
							f.WriteString(v)
						}
					}
				} // }}}
			}},
		"get": &ctx.Command{Name: "get field offset table where [parse func field]", Help: "执行查询语句",
			Form: map[string]int{"parse": 2},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				index := 0 // {{{
				if len(arg) > 1 {
					if i, e := strconv.Atoi(arg[1]); e == nil {
						index = i
					}
				}
				field := m.Confx("field", arg, 0)
				offset := m.Confx("offset", arg, 1, "offset %s")
				table := m.Confx("table", arg, 2, "from %s")
				where := m.Confx("where", arg, 3, "where %s")
				limit := "limit 1"

				msg := m.Spawn().Cmd("query", fmt.Sprintf("select %s %s %s %s %s", field, table, where, limit, offset))
				value := msg.Matrix(index, field)

				switch m.Confx("parse", m.Option("parse")) {
				case "json":
					extra := ""
					if len(m.Meta["parse"]) > 1 {
						extra = m.Meta["parse"][1]
					}
					data := map[string]interface{}{}
					if json.Unmarshal([]byte(value), &data); extra == "" {
						for k, v := range data {
							m.Echo("%s: %v\n", k, v)
						}
					} else if v, ok := data[extra]; ok {
						m.Echo("%v", v)
					}
				default:
					m.Echo("%v", value)
				}
				// }}}
			}},
		"csv": &ctx.Command{Name: "csv scv_file sql_file sql", Help: "执行查询语句",
			Form: map[string]int{"csv_col_sep": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				f, e := os.Open(arg[0])
				m.Assert(e)

				s, e := os.Create(arg[1])
				m.Assert(e)

				head := false
				bio := bufio.NewScanner(f)
				for bio.Scan() {
					line := bio.Text()
					if !head {
						head = true
						continue
					}

					args := []interface{}{}
					for _, v := range strings.Split(line, m.Confx("csv_col_sep")) {
						args = append(args, v)
					}

					s.WriteString(fmt.Sprintf(arg[2], args...))
					s.WriteString(";\n")
				}
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
