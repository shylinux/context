package mdb

import (
	"contexts/ctx"
	"toolkit"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"encoding/json"
	"fmt"
	"strings"
)

type MDB struct {
	*sql.DB
	*ctx.Context
}

func (mdb *MDB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{
		"database": &ctx.Cache{Name: "database", Value: m.Confx("database", arg, 0), Help: "数据库"},
		"username": &ctx.Cache{Name: "username", Value: m.Confx("username", arg, 1), Help: "账户"},
		"password": &ctx.Cache{Name: "password", Value: m.Confx("password", arg, 2), Help: "密码"},
		"address":  &ctx.Cache{Name: "address", Value: m.Confx("address", arg, 3), Help: "地址"},
		"protocol": &ctx.Cache{Name: "protocol(tcp)", Value: m.Confx("protocol", arg, 4), Help: "协议"},
		"driver":   &ctx.Cache{Name: "driver(mysql)", Value: m.Confx("driver", arg, 5), Help: "驱动"},
	}
	c.Configs = map[string]*ctx.Config{
		"dbs":    &ctx.Config{Name: "dbs", Value: []string{}, Help: "数据库"},
		"tabs":   &ctx.Config{Name: "tabs", Value: []string{}, Help: "关系表"},
		"limit":  &ctx.Config{Name: "limit", Value: "10", Help: "分页"},
		"offset": &ctx.Config{Name: "offset", Value: "0", Help: "偏移"},
	}

	s := new(MDB)
	s.Context = c
	return s
}
func (mdb *MDB) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return mdb
}
func (mdb *MDB) Start(m *ctx.Message, arg ...string) bool {
	db, e := sql.Open(m.Cap("driver"), fmt.Sprintf("%s:%s@%s(%s)/%s", m.Cap("username"), m.Cap("password"), m.Cap("protocol"), m.Cap("address"), m.Cap("database")))
	m.Assert(e)
	mdb.DB = db
	m.Log("info", "mdb open %s", m.Cap("stream", m.Cap("database")))
	return false
}
func (mdb *MDB) Close(m *ctx.Message, arg ...string) bool {
	if mdb.DB != nil {
		return false
	}
	return true
}

var Index = &ctx.Context{Name: "mdb", Help: "数据中心",
	Caches: map[string]*ctx.Cache{
		"nsource": &ctx.Cache{Name: "nsource", Value: "0", Help: "已打开数据库的数量"},
	},
	Configs: map[string]*ctx.Config{
		"database": &ctx.Config{Name: "database", Value: "demo", Help: "默认数据库"},
		"username": &ctx.Config{Name: "username", Value: "demo", Help: "默认账户"},
		"password": &ctx.Config{Name: "password", Value: "demo", Help: "默认密码"},
		"protocol": &ctx.Config{Name: "protocol(tcp)", Value: "tcp", Help: "默认协议"},
		"address":  &ctx.Config{Name: "address", Value: "", Help: "默认地址"},
		"driver":   &ctx.Config{Name: "driver(mysql)", Value: "mysql", Help: "默认驱动"},

		"temp":      &ctx.Config{Name: "temp", Value: map[string]interface{}{}, Help: "缓存数据"},
		"temp_view": &ctx.Config{Name: "temp_view", Value: map[string]interface{}{}, Help: "缓存数据"},
	},
	Commands: map[string]*ctx.Command{
		"temp": &ctx.Command{Name: "temp [type [meta [data]]] [tid [node|ship|data] [chain... [select ...]]]", Form: map[string]int{"select": -1}, Help: "缓存数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 2 { // 添加数据
				if temp := m.Confm("temp", arg[0]); temp == nil {
					h := m.Cmdx("aaa.hash", arg[0], arg[1])
					m.Confv("temp", h, map[string]interface{}{
						"create_time": m.Time(), "expire_time": m.Time("60s"),
						"type": arg[0], "meta": arg[1], "data": m.Optionv(arg[2]),
					})
					arg[2], arg = h, arg[2:]
				}
			}

			if len(arg) > 1 {
				if temp, msg := m.Confm("temp", arg[0]), m; temp != nil {
					hash, arg := arg[0], arg[1:]
					switch arg[0] {
					case "node": // 查看节点
						m.Put("option", "temp", temp).Cmdy("ctx.trans", "temp")
						return
					case "ship": //查看链接
						for k, v := range temp {
							val := v.(map[string]interface{})
							m.Add("append", "key", k)
							m.Add("append", "create_time", val["create_time"])
							m.Add("append", "type", val["type"])
							m.Add("append", "meta", val["meta"])
						}
						m.Sort("create_time", "time_r").Table()
						return
					case "data": // 查看数据
						arg = arg[1:]
					}

					trans := strings.Join(append([]string{hash, "data"}, arg...), ".")
					h := m.Cmdx("aaa.hash", "trans", trans)

					if len(arg) == 0 || temp["type"].(string) == "trans" {
						h = hash
					} else { // 转换数据
						if view := m.Confm("temp", h); view != nil && false { // 加载缓存
							msg = m.Spawn()
							switch data := view["data"].(type) {
							case map[string][]string:
								msg.Meta = data
							case map[string]interface{}:
								for k, v := range data {
									switch val := v.(type) {
									case []interface{}:
										msg.Add("append", k, val)
									}
								}
								m.Confv("temp", []string{h, "data"}, msg.Meta)
							}
							temp = view
						} else if arg[0] == "" { //  添加缓存
							b, _ := json.MarshalIndent(temp["data"], "", "  ")
							m.Echo(string(b))
						} else {
							msg = m.Put("option", "temp", temp["data"]).Cmd("ctx.trans", "temp", arg)

							m.Confv("temp", h, map[string]interface{}{
								"create_time": m.Time(), "expire_time": m.Time("60s"),
								"type": "trans", "meta": trans, "data": msg.Meta,
								"ship": map[string]interface{}{
									hash: map[string]interface{}{"create_time": m.Time(), "ship": "0", "type": temp["type"], "meta": temp["meta"]},
								},
							})
							m.Confv("temp", []string{hash, "ship", h}, map[string]interface{}{
								"create_time": m.Time(), "ship": "1", "type": "trans", "meta": trans,
							})
							temp = m.Confm("temp", h)
						}
					}

					if m.Options("select") { // 过滤数据
						chain := strings.Join(m.Optionv("select").([]string), " ")
						hh := m.Cmdx("aaa.hash", "select", chain, h)

						if view := m.Confm("temp", hh); view != nil { // 加载缓存
							msg = msg.Spawn()
							switch data := view["data"].(type) {
							case map[string][]string:
								msg.Meta = data
							case map[string]interface{}:
								for k, v := range data {
									switch val := v.(type) {
									case []interface{}:
										msg.Add("append", k, val)
									}
								}
								m.Confv("temp", []string{h, "data"}, msg.Meta)
							}
						} else { // 添加缓存
							msg = msg.Spawn().Copy(msg, "append").Cmd("select", m.Optionv("select"))

							m.Confv("temp", hh, map[string]interface{}{
								"create_time": m.Time(), "expire_time": m.Time("60s"),
								"type": "select", "meta": chain, "data": msg.Meta,
								"ship": map[string]interface{}{
									h: map[string]interface{}{"create_time": m.Time(), "ship": "0", "type": temp["type"], "meta": temp["meta"]},
								},
							})

							m.Confv("temp", []string{h, "ship", hh}, map[string]interface{}{
								"create_time": m.Time(), "ship": "1", "type": "select", "meta": chain,
							})
						}
					}

					msg.CopyTo(m)
					return
				}
			}

			arg, h := kit.Slice(arg)
			if h != "" {
				if temp := m.Confm("temp", h); temp != nil {
					m.Echo(h)
					return
				}
			}

			// 缓存列表
			m.Confm("temp", func(key string, temp map[string]interface{}) {
				if len(arg) == 0 || strings.HasPrefix(key, arg[0]) || strings.HasSuffix(key, arg[0]) || (temp["type"].(string) == arg[0] && (len(arg) == 1 || strings.Contains(temp["meta"].(string), arg[1]))) {
					m.Add("append", "key", key)
					m.Add("append", "create_time", temp["create_time"])
					m.Add("append", "type", temp["type"])
					m.Add("append", "meta", temp["meta"])
				}
			})
			m.Sort("create_time", "time_r").Table().Cmd("select", m.Optionv("select"))
			return
		}},

		"open": &ctx.Command{Name: "open [database [username [password [address [protocol [driver]]]]]]",
			Help: "打开数据库, database: 数据库名, username: 用户名, password: 密码, address: 服务地址, protocol: 服务协议, driver: 数据库类型",
			Form: map[string]int{"dbname": 1, "dbhelp": 1},
			Auto: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) bool {
				switch len(arg) {
				case 0:
					m.Auto("", fmt.Sprintf("database(%s)", m.Conf("database")), "数据库")
				case 1:
					m.Auto("", fmt.Sprintf("username(%s)", m.Conf("username")), "账号")
				case 2:
					m.Auto("", fmt.Sprintf("password(%s)", m.Conf("password")), "密码")
				case 3:
					m.Auto("", fmt.Sprintf("address(%s)", m.Conf("address")), "地址")
				case 4:
					m.Auto("", fmt.Sprintf("protocol(%s)", m.Conf("protocol")), "协议")
				case 5:
					m.Auto("", fmt.Sprintf("driver(%s)", m.Conf("driver")), "驱动")
				default:
					m.Auto("", "[dbname name]", "模块名")
					m.Auto("", "[dbhelp help]", "帮助文档")
				}
				return true
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				m.Start(kit.Select(fmt.Sprintf("db%d", m.Capi("nsource", 1)), m.Option("dbname")),
					kit.Select("数据源", m.Option("dbhelp")), arg...)
				return
			}},
		"exec": &ctx.Command{Name: "exec sql [arg]", Help: "操作数据库, sql: SQL语句, arg: 操作参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
			return
		}},
		"query": &ctx.Command{Name: "query sql [arg]", Help: "查询数据库, sql: SQL语句, arg: 查询参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
				m.Table()
			}
			return
		}},

		"db": &ctx.Command{Name: "db [which]", Help: "查看或选择数据库",
			Auto: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) bool {
				if len(arg) == 0 {
					m.Put("option", "auto_cmd", "").Spawn().Cmd("query", "show databases").Table(func(line map[string]string) {
						for _, v := range line {
							m.Auto(v, "", "")
						}
					})
				}
				return true
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) == 0 {
					m.Cmdy(".query", "show databases").Table()
					return
				}

				m.Assert(m.Cmdy("exec", fmt.Sprintf("use %s", arg[0])))
				m.Echo(m.Cap("database", arg[0]))
				return
			}},
		"tab": &ctx.Command{Name: "tab [which [field]]", Help: "查看关系表，which: 表名, field: 字段名",
			Auto: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) bool {
				switch len(arg) {
				case 0:
					m.Put("option", "auto_cmd", "").Spawn().Cmd("query", "show tables").Table(func(line map[string]string) {
						for _, v := range line {
							m.Auto(v, "", "")
						}
					})
				case 1:
					m.Put("option", "auto_cmd", "").Spawn().Cmd("query", fmt.Sprintf("desc %s", arg[0])).Table(func(line map[string]string) {
						m.Auto(line["Field"], line["Type"], line["Default"])
					})
				}
				return true
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				switch len(arg) {
				case 0:
					m.Cmdy(".query", "show tables").Table()
				case 1:
					m.Cmdy(".query", fmt.Sprintf("desc %s", arg[0])).Table()
				case 2:
					m.Cmdy(".query", fmt.Sprintf("desc %s", arg[0])).Cmd("select", "Field", arg[1]).CopyTo(m)
				}
				return
			}},
		"show": &ctx.Command{Name: "show table fields...", Help: "查询数据, table: 表名, fields: 字段, where: 查询条件, group: 聚合字段, order: 排序字段",
			Form: map[string]int{"where": 1, "eq": 2, "like": 2, "in": 2, "begin": 2, "group": 1, "order": 1, "desc": 0, "limit": 1, "offset": 1, "other": -1},
			Auto: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) bool {
				if len(arg) == 0 {
					m.Put("option", "auto_cmd", "").Spawn().Cmd("query", "show tables").Table(func(line map[string]string) {
						for _, v := range line {
							m.Auto(v, "", "")
						}
					})
					return true
				}

				m.Auto("where", "stmt", "条件语句")
				m.Auto("eq", "field value", "条件语句")
				m.Auto("in", "field value", "条件语句")
				m.Auto("like", "field value", "条件语句")
				m.Auto("begin", "field value", "条件语句")

				m.Auto("group", "field", "聚合")
				m.Auto("order", "field", "排序")
				m.Auto("desc", "", "降序")
				m.Auto("limit", "10", "分页")
				m.Auto("offset", "0", "偏移")
				m.Auto("other", "stmt", "其它")
				return true
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) == 0 {
					m.Cmdy(".query", "show tables")
					return
				}

				stmt := []string{
					fmt.Sprintf("select %s", kit.Select("*", strings.Join(arg[1:], ","))),
					fmt.Sprintf("from %s", arg[0]),
				}

				where := []string{}
				if m.Has("where") {
					where = append(where, m.Option("where"))
				}
				for i := 0; i < len(m.Meta["eq"]); i += 2 {
					where = append(where, fmt.Sprintf("%s='%s'", m.Meta["eq"][i], m.Meta["eq"][i+1]))
				}
				for i := 0; i < len(m.Meta["in"]); i += 2 {
					where = append(where, fmt.Sprintf("%s in (%s)", m.Meta["in"][i], m.Meta["in"][i+1]))
				}
				for i := 0; i < len(m.Meta["like"]); i += 2 {
					where = append(where, fmt.Sprintf("%s like '%s'", m.Meta["like"][i], m.Meta["like"][i+1]))
				}
				for i := 0; i < len(m.Meta["begin"]); i += 2 {
					where = append(where, fmt.Sprintf("%s like '%s%%'", m.Meta["begin"][i], m.Meta["begin"][i+1]))
				}
				if len(where) > 0 {
					stmt = append(stmt, "where", strings.Join(where, " and "))
				}

				if m.Has("group") {
					stmt = append(stmt, fmt.Sprintf("group by %s", m.Option("group")))
				}
				if m.Has("order") {
					stmt = append(stmt, fmt.Sprintf("order by %s", m.Option("order")))
				}
				if m.Has("desc") {
					stmt = append(stmt, "desc")
				}

				stmt = append(stmt, fmt.Sprintf("limit %s", m.Confx("limit")))
				stmt = append(stmt, fmt.Sprintf("offset %s", m.Confx("offset")))

				for _, v := range m.Meta["other"] {
					stmt = append(stmt, m.Parse(v))
				}

				m.Cmdy(".query", strings.Join(stmt, " "))
				return
			}},
		"update": &ctx.Command{Name: "update table condition [set field value]", Help: "修改数据, table: 关系表, condition: 条件语句, set: 修改数据",
			Auto: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) bool {
				if len(arg) == 0 {
					m.Put("option", "auto_cmd", "").Spawn().Cmd("query", "show tables").Table(func(line map[string]string) {
						for _, v := range line {
							m.Auto(v, "", "")
						}
					})
					return true
				}
				if len(arg) == 2 {
					// m.Put("option", "auto_cmd", "").Spawn().Cmd("show", arg[0], "where", arg[1]).Table(func(line map[string]string) {
					// 	for _, v := range line {
					// 		m.Auto(v, "", "")
					// 	}
					// })
					return true
				}
				return true
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) == 2 {
					m.Cmd(".show", arg[0], "where", arg[1]).CopyTo(m, "append")
				}

				fields := []string{}
				values := []string{}
				for i := 3; i < len(arg)-1; i += 2 {
					fields = append(fields, arg[i])
					values = append(values, fmt.Sprintf("%s='%s'", arg[i], arg[i+1]))
				}

				stmt := []string{fmt.Sprintf("update %s", arg[0])}
				stmt = append(stmt, "set", strings.Join(values, ","))
				stmt = append(stmt, fmt.Sprintf("where %s", arg[1]))

				m.Cmd(".exec", strings.Join(stmt, " "))
				m.Cmdy("show", arg[0], fields, "where", arg[1])
				return
			}},
	},
}

func init() {
	mdb := &MDB{}
	mdb.Context = Index
	ctx.Index.Register(Index, mdb)
}
