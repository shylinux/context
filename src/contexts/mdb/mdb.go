package mdb

import (
	"contexts/ctx"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

type MDB struct {
	*sql.DB
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
		"dbs":    &ctx.Config{Name: "dbs", Value: []string{}, Help: "关系表"},
		"tables": &ctx.Config{Name: "dbs", Value: []string{}, Help: "关系表"},
		"table":  &ctx.Config{Name: "关系表", Value: "0", Help: "关系表"},
		"field":  &ctx.Config{Name: "字段名", Value: "", Help: "字段名"},
		"where":  &ctx.Config{Name: "条件", Value: "", Help: "条件"},
		"group":  &ctx.Config{Name: "聚合", Value: "", Help: "聚合"},
		"order":  &ctx.Config{Name: "排序", Value: "", Help: "排序"},
		"limit":  &ctx.Config{Name: "分页", Value: "10", Help: "分页"},
		"offset": &ctx.Config{Name: "偏移", Value: "0", Help: "偏移"},
		"parse":  &ctx.Config{Name: "解析", Value: "", Help: "解析"},
	}

	s := new(MDB)
	s.Context = c
	return s
}
func (mdb *MDB) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return mdb
}
func (mdb *MDB) Start(m *ctx.Message, arg ...string) bool {
	db, e := sql.Open(m.Cap("driver"), fmt.Sprintf("%s:%s@%s(%s)/%s",
		m.Cap("username"), m.Cap("password"), m.Cap("protocol"), m.Cap("address"), m.Cap("database")))
	m.Assert(e)
	mdb.DB = db
	m.Log("info", "mdb open %s", m.Cap("database"))
	return false
}
func (mdb *MDB) Close(m *ctx.Message, arg ...string) bool {
	switch mdb.Context {
	case m.Target():
	case m.Source():
	}
	return false
}

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

		"csv_col_sep": &ctx.Config{Name: "字段分隔符", Value: "\t", Help: "字段分隔符"},
		"csv_row_sep": &ctx.Config{Name: "记录分隔符", Value: "\n", Help: "记录分隔符"},
	},
	Commands: map[string]*ctx.Command{
		"open": &ctx.Command{Name: "open [database [username [password [address [protocol [driver]]]]]]",
			Help: "open打开数据库, database: 数据库名, username: 用户名, password: 密码, address: 服务地址, protocol: 服务协议, driver: 数据库类型",
			Form: map[string]int{"dbname": 1, "dbhelp": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				dbname := fmt.Sprintf("db%d", m.Capi("nsource", 1))
				dbhelp := "数据源"
				if m.Has("dbname") {
					dbname = m.Option("dbname")
				}
				if m.Has("dbhelp") {
					dbname = m.Option("dbhelp")
				}
				m.Start(dbname, dbhelp, arg...)
			}},
		"exec": &ctx.Command{Name: "exec sql [arg]", Help: "操作数据库, sql: SQL语句, arg: 操作参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) && mdb.DB != nil {
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
				m.Echo("%d", n)
			}
		}},
		"query": &ctx.Command{Name: "query sql [arg]", Help: "查询数据库, sql: SQL语句, arg: 查询参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) && mdb.DB != nil {
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
						case nil:
							m.Add("append", k, "")
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
		}},
		"db": &ctx.Command{Name: "db [which]", Help: "查看或选择数据库信息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) && mdb.DB != nil {
				if len(arg) == 0 {
					msg := m.Spawn().Cmd("query", "show databases")
					dbs := []string{}
					for i, v := range msg.Meta[msg.Meta["append"][0]] {
						dbs = append(dbs, v)
						m.Echo("%d: %s\n", i, v)
					}
					m.Target().Configs["dbs"].Value = dbs
					return
				}

				db := m.Confv("dbs", arg[0]).(string)
				m.Assert(m.Spawn().Cmd("exec", fmt.Sprintf("use %s", db)))
				m.Echo(m.Cap("database", db))
			}
		}},
		"tab": &ctx.Command{Name: "tab[which [field]]", Help: "查看关系表信息，which: 表名, field: 字段名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*MDB); m.Assert(ok) {
				if len(arg) == 0 {
					msg := m.Spawn().Cmd("query", "show tables")
					tables := []string{}
					for i, v := range msg.Meta[msg.Meta["append"][0]] {
						tables = append(tables, v)
						m.Echo("%d: %s\n", i, v)
					}
					m.Target().Configs["tables"].Value = tables
					return
				}

				table := arg[0]
				switch v := m.Confv("tables", arg[0]).(type) {
				case string:
					table = v
				}

				msg := m.Spawn().Cmd("query", fmt.Sprintf("desc %s", table))
				m.Copy(msg, "append")
				m.Table()
			}
		}},
		"show": &ctx.Command{Name: "show table fields...",
			Help: "查询数据库, table: 表名, fields: 字段, where: 查询条件, group: 聚合字段, order: 排序字段",
			Form: map[string]int{"where": 1, "group": 1, "desc": 0, "order": 1, "limit": 1, "offset": 1, "other": -1,
				"extra_field": 2, "extra_fields": 1, "extra_format": 1, "trans_field": 1, "trans_map": 2},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if _, ok := m.Target().Server.(*MDB); m.Assert(ok) {
					if len(arg) == 0 {
						msg := m.Spawn().Cmd("query", "show tables")
						for _, v := range msg.Meta[msg.Meta["append"][0]] {
							m.Add("append", "table", v)
						}
						m.Table()
						return
					}

					table := m.Confx("table", arg, 0)
					if v := m.Confv("tables", table); v != nil {
						table = v.(string)
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
					if m.Has("desc") {
						order = order + " desc"
					}
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

					m.Target().Configs["template_value"] = &ctx.Config{}
					if m.Has("extra_field") && len(m.Meta[m.Option("extra_field")]) > 0 {
						format := "%v"
						if m.Has("extra_format") {
							format = m.Option("extra_format")
						}
						for i := 0; i < len(m.Meta[m.Option("extra_field")]); i++ {
							json.Unmarshal([]byte(m.Meta[m.Option("extra_field")][i]), &m.Target().Configs["template_value"].Value)
							if m.Meta["extra_field"][1] != "" {
								m.Target().Configs["template_value"].Value = m.Confv("template_value", m.Meta["extra_field"][1])
							}
							fields = strings.Split(m.Option("extra_fields"), " ")
							switch v := m.Confv("template_value").(type) {
							case map[string]interface{}:
								for _, k := range fields {
									if k == "" {
										continue
									}
									if val, ok := v[k]; ok {
										m.Add("append", k, fmt.Sprintf(format, val))
									} else {
										m.Add("append", k, fmt.Sprintf(format, ""))
									}
								}
								m.Meta[m.Option("extra_field")][i] = ""
							case float64:
								m.Meta[m.Option("extra_field")][i] = fmt.Sprintf(format, int(v))
							case nil:
								m.Meta[m.Option("extra_field")][i] = ""
							default:
								m.Meta[m.Option("extra_field")][i] = fmt.Sprintf(format, v)
							}
						}
					}

					if m.Has("trans_field") {
						trans := map[string]string{}
						for i := 0; i < len(m.Meta["trans_map"]); i += 2 {
							trans[m.Meta["trans_map"][i]] = m.Meta["trans_map"][i+1]
						}
						for i := 0; i < len(m.Meta[m.Option("trans_field")]); i++ {
							if t, ok := trans[m.Meta[m.Option("trans_field")][i]]; ok {
								m.Meta[m.Option("trans_field")][i] = t
							}
						}
					}

					m.Color(31, table).Echo(" %s %s %s %s %s %v\n", where, group, order, limit, offset, m.Meta["other"])

					m.Table(func(maps map[string]string, lists []string, line int) bool {
						for i, v := range lists {
							if line == -1 {
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
				}
			}},
		"set": &ctx.Command{Name: "set table [field value] where condition", Help: "查看或选择数据库信息", Form: map[string]int{"where": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) && mdb.DB != nil {
				sql := []string{"update", arg[0], "set"}
				fields := []string{}
				for i := 1; i < len(arg)-1; i += 2 {
					sql = append(sql, fmt.Sprintf("%s='%s'", arg[i], arg[i+1]))
					fields = append(fields, arg[i])
					if i < len(arg)-2 {
						sql = append(sql, ",")
					}
				}
				sql = append(sql, fmt.Sprintf("where %s", m.Option("where")))
				m.Spawn().Cmd("exec", strings.Join(sql, " "))
				msg := m.Spawn().Cmd("show", arg[0], fields, "where", m.Option("where"))
				m.Copy(msg, "result").Copy(msg, "append")
			}
		}},
	},
}

func init() {
	mdb := &MDB{}
	mdb.Context = Index
	ctx.Index.Register(Index, mdb)
}
