package mdb

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"

	"contexts/ctx"
	"toolkit"

	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type MDB struct {
	conn redis.Conn
	*sql.DB
	*ctx.Context
}

func (mdb *MDB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{
		"database": &ctx.Cache{Name: "database", Value: m.Confx("database", arg, 0), Help: "数据库"},
		"username": &ctx.Cache{Name: "username", Value: m.Confx("username", arg, 1), Help: "用户名"},
		"password": &ctx.Cache{Name: "password", Value: m.Confx("password", arg, 2), Help: "密码"},
		"address":  &ctx.Cache{Name: "address", Value: m.Confx("address", arg, 3), Help: "地址"},
		"protocol": &ctx.Cache{Name: "protocol(tcp)", Value: m.Confx("protocol", arg, 4), Help: "协议"},
		"driver":   &ctx.Cache{Name: "driver(mysql)", Value: m.Confx("driver", arg, 5), Help: "驱动"},
		"redis":    &ctx.Cache{Name: "redis", Value: "", Help: "数据缓存"},
	}
	c.Configs = map[string]*ctx.Config{
		"dbs":    &ctx.Config{Name: "dbs", Value: []string{}, Help: "数据库"},
		"tabs":   &ctx.Config{Name: "tabs", Value: []string{}, Help: "关系表"},
		"limit":  &ctx.Config{Name: "limit", Value: "10", Help: "分页"},
		"offset": &ctx.Config{Name: "offset", Value: "0", Help: "偏移"},
	}

	return &MDB{Context: c}
}
func (mdb *MDB) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return mdb
}
func (mdb *MDB) Start(m *ctx.Message, arg ...string) bool {
	if db, e := sql.Open(m.Cap("driver"), fmt.Sprintf("%s:%s@%s(%s)/%s", m.Cap("username"), m.Cap("password"),
		m.Cap("protocol"), m.Cap("address"), m.Cap("database"))); m.Assert(e) {
		m.Log("info", "mdb open %s", m.Cap("stream", m.Cap("database")))
		mdb.DB = db
	}
	return false
}
func (mdb *MDB) Close(m *ctx.Message, arg ...string) bool {
	return false
}

var Index = &ctx.Context{Name: "mdb", Help: "数据中心",
	Caches: map[string]*ctx.Cache{
		"nsource": &ctx.Cache{Name: "nsource", Value: "0", Help: "已打开数据库的数量"},
	},
	Configs: map[string]*ctx.Config{
		"database": &ctx.Config{Name: "database", Value: "demo", Help: "默认数据库"},
		"username": &ctx.Config{Name: "username", Value: "demo", Help: "默认账户"},
		"password": &ctx.Config{Name: "password", Value: "demo", Help: "默认密码"},
		"address":  &ctx.Config{Name: "address", Value: ":6379", Help: "默认地址"},
		"protocol": &ctx.Config{Name: "protocol(tcp)", Value: "tcp", Help: "默认协议"},
		"driver":   &ctx.Config{Name: "driver(mysql)", Value: "mysql", Help: "默认驱动"},

		"ktv": &ctx.Config{Name: "ktv", Value: map[string]interface{}{
			"conf": map[string]interface{}{"expire": "24h"}, "data": map[string]interface{}{},
		}, Help: "缓存数据"},

		"temp":        &ctx.Config{Name: "temp", Value: map[string]interface{}{}, Help: "缓存数据"},
		"temp_view":   &ctx.Config{Name: "temp_view", Value: map[string]interface{}{}, Help: "缓存数据"},
		"temp_expire": &ctx.Config{Name: "temp_expire(s)", Value: "3000", Help: "缓存数据"},

		"note": &ctx.Config{Name: "note", Value: map[string]interface{}{
			"faa01a8fc2fc92dae3fbc02ac1b4ec75": map[string]interface{}{
				"create_time": "1990-07-30 07:08:09", "access_time": "2017-11-01 02:03:04",
				"type": "index", "name": "shy", "data": "", "ship": map[string]interface{}{
					"prev": map[string]interface{}{"type": "index", "data": ""},
				},
			},
			"81c5709d091eb04bd31ee751c3f81023": map[string]interface{}{
				"create_time": "1990-07-30 07:08:09", "access_time": "2017-11-01 02:03:04",
				"meta": []interface{}{"text", "text", "place", "place", "label", "label", "friend", "friend", "username", "username"},
				"view": map[string]interface{}{
					"list": map[string]interface{}{"name": "left", "create_date": "right"},
					"edit": map[string]interface{}{"model": "hidden", "username": "hidden"},
				},
				"bind": map[string]interface{}{},
				"type": "model", "name": "shy", "data": "", "ship": map[string]interface{}{
					"prev": map[string]interface{}{"type": "model", "data": ""},
				},
			},
		}, Help: "数据结构"},
		"note_view": &ctx.Config{Name: "note_view", Value: map[string]interface{}{
			"default": []interface{}{"key", "create_time", "type", "name", "model", "value"},
			"base":    []interface{}{"key", "create_time", "type", "name", "model", "value"},
			"full":    []interface{}{"key", "create_time", "access_time", "type", "name", "model", "value", "view", "data", "ship"},
		}, Help: "数据视图"},
	},
	Commands: map[string]*ctx.Command{
		"open": &ctx.Command{Name: "open [database [username [password [address [protocol [driver]]]]]]",
			Help: "打开数据库, database: 数据库名, username: 用户名, password: 密码, address: 服务地址, protocol: 服务协议, driver: 数据库类型",
			Form: map[string]int{"dbname": 1, "dbhelp": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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

				if ret, e := mdb.Exec(arg[0], which...); m.Assert(e) {
					id, _ := ret.LastInsertId()
					n, _ := ret.RowsAffected()
					m.Log("info", "last(%s) nrow(%s)", m.Append("last", id), m.Append("nrow", n))
					m.Echo("%d", n)
				}
			}
			return
		}},
		"query": &ctx.Command{Name: "query sql [arg]", Help: "查询数据库, sql: SQL语句, arg: 查询参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) && mdb.DB != nil {
				which := make([]interface{}, 0, len(arg))
				for _, v := range arg[1:] {
					which = append(which, v)
				}

				if rows, e := mdb.Query(arg[0], which...); m.Assert(e) {
					defer rows.Close()
					if cols, e := rows.Columns(); m.Assert(e) {
						num := len(cols)

						for rows.Next() {
							vals := make([]interface{}, num)
							ptrs := make([]interface{}, num)
							for i := range vals {
								ptrs[i] = &vals[i]
							}
							rows.Scan(ptrs...)

							for i, k := range cols {
								m.Push(k, vals[i])
							}
						}

						if len(m.Meta["append"]) > 0 {
							m.Log("info", "rows(%d) cols(%d)", len(m.Meta[m.Meta["append"][0]]), len(m.Meta["append"]))
						} else {
							m.Log("info", "rows(0) cols(0)")
						}
						m.Table()
					}
				}

			}
			return
		}},
		"redis": &ctx.Command{Name: "redis [open address]|[cmd...]", Help: "缓存数据库", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if mdb, ok := m.Target().Server.(*MDB); m.Assert(ok) {
				switch arg[0] {
				case "open":
					if mdb.conn, e = redis.Dial("tcp", m.Cap("redis", arg[1]), redis.DialKeepAlive(time.Second*10)); m.Assert(e) {
						m.Log("info", "redis: %v", arg[1])
					}

				default:
					if mdb.conn == nil || mdb.conn.Err() != nil {
						if m.Caps("redis") {
							m.Cmd("mdb.redis", "open", m.Cap("redis"))
						} else {
							m.Echo("not open")
							break
						}
					}

					args := []interface{}{}
					for _, v := range arg[1:] {
						args = append(args, v)
					}

					if res, err := mdb.conn.Do(arg[0], args...); m.Assert(err) {
						switch val := res.(type) {
						case redis.Error:
							m.Echo("%v", val)

						case []interface{}:
							for i, v := range val {
								m.Add("append", "index", i)
								m.Add("append", "value", v)
							}
							m.Table()

						default:
							var data interface{}
							if str := kit.Format(res); json.Unmarshal([]byte(str), &data) == nil {
								m.Echo(kit.Formats(data))
							} else {
								m.Echo(str)
							}
						}
					}
				}
			}
			return
		}},

		"db": &ctx.Command{Name: "db [which]", Help: "查看或选择数据库", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmdy(".query", "show databases").Table()
				return
			}

			m.Assert(m.Cmdy("exec", fmt.Sprintf("use %s", arg[0])))
			m.Echo(m.Cap("database", arg[0]))
			return
		}},
		"tab": &ctx.Command{Name: "tab [which [field]]", Help: "查看关系表，which: 表名, field: 字段名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
			Form: map[string]int{"where": 1, "eq": 2, "ne": 2, "in": 2, "like": 2, "begin": 2, "group": 1, "order": 1, "desc": 0, "limit": 1, "offset": 1, "other": -1},
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
				for i := 0; i < len(m.Meta["ne"]); i += 2 {
					where = append(where, fmt.Sprintf("%s!='%s'", m.Meta["ne"][i], m.Meta["ne"][i+1]))
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
		"update": &ctx.Command{Name: "update [table [condition [field [value]]...]]",
			Help: "修改数据, table: 关系表, condition: 条件语句, field: 字段名, value: 字段值",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				switch len(arg) {
				case 0:
					m.Cmdy(".show")
				case 1:
					m.Cmdy(".show", arg[0])
				case 2:
					m.Cmdy(".show", arg[0], "where", arg[1])
				case 3:
					m.Cmdy(".show", arg[0], arg[2], "where", arg[1])
				default:
					fields := []string{}
					values := []string{}
					for i := 2; i < len(arg)-1; i += 2 {
						fields = append(fields, arg[i])
						values = append(values, fmt.Sprintf("%s='%s'", arg[i], arg[i+1]))
					}

					stmt := []string{"update", arg[0]}
					stmt = append(stmt, "set", strings.Join(values, ","))
					stmt = append(stmt, "where", arg[1])
					m.Cmd(".exec", strings.Join(stmt, " "))

					m.Cmdy(".show", arg[0], fields, "where", arg[1])
				}
				return
			}},

		"ktv": &ctx.Command{Name: "ktv", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				now := kit.Int(m.Time("stamp"))
				m.Confm("ktv", "data", func(key string, value map[string]interface{}) {
					m.Push("key", key)
					m.Push("expire", kit.Int(value["expire"])-now)
					m.Push("value", value["value"])
				})
				m.Table()
				return
			}
			if len(arg) == 1 {
				if m.Confi("ktv", []string{"data", arg[0], "expire"}) < kit.Int(m.Time("stamp")) {
					m.Conf("ktv", []string{"data", arg[0]}, "")
				}
				m.Echo(m.Conf("ktv", []string{"data", arg[0], "value"}))
				return
			}
			m.Confv("ktv", []string{"data", arg[0]}, map[string]interface{}{
				"expire": m.Time(kit.Select(m.Conf("ktv", "conf.expire"), arg, 2), "stamp"),
				"value":  arg[1],
			})
			m.Echo(arg[1])
			return
		}},
		"temp": &ctx.Command{Name: "temp [type [meta [data]]] [tid [node|ship|data] [chain... [select ...]]]",
			Form: map[string]int{"select": -1, "limit": 1}, Help: "缓存数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) > 0 && arg[0] == "check" {
					h := ""
					for i := 1; i < len(arg)-1; i += 2 {
						switch arg[i] {
						case "url", "trans":
							h = ""
						}
						if h = m.Cmdx("aaa.hash", arg[i], arg[i+1], h); !m.Confs("temp", h) {
							return
						}
						expire := kit.Time(m.Conf("temp", []string{h, "create_time"})) + kit.Int(m.Confx("temp_expire")) - kit.Time(m.Time())
						m.Log("info", "expire: %ds", expire)
						if expire < 0 {
							return
						}
					}
					m.Echo(h)
					return
				}

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
							} else if arg[0] == "hash" { //  添加缓存
								m.Echo(hash)
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

							if view := m.Confm("temp", hh); view != nil && false { // 加载缓存
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

				h, arg := arg[0], arg[1:]
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
		"note": &ctx.Command{Name: "note [model [name [type name]...]]|[index [name data...]]|[value name data...]|[name model data...]",
			Form: map[string]int{"eq": 2, "begin": 2, "offset": 1, "limit": 1}, Help: "记事", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				offset := kit.Int(kit.Select(m.Conf("table", "offset"), m.Option("table.offset")))
				limit := kit.Int(kit.Select(m.Conf("table", "limit"), m.Option("table.limit")))

				// 节点列表
				if len(arg) == 0 {
					m.CopyFuck(m.Cmd("mdb.config", "note", "format", "table", "fields", "create_time", "access_time", "type", "name", "data", "ship"), "append").Set("result").Table()
					return
				}

				// 节点详情
				if note := m.Confm("note", arg[0]); note != nil {
					m.CopyFuck(m.Cmd("mdb.config", "note", arg[0]), "append").Set("result").Table()
					return
				}

				// 节点列表
				hm, _ := kit.Hash("type", arg[0], "name", "shy")
				if len(arg) == 2 && arg[0] == "value" {
					hm, _ = kit.Hash("type", "index", "name", arg[1])
					hm = m.Conf("note", []string{hm, "ship", "value", "data"})
					arg = arg[1:]
				} else if len(arg) == 2 && arg[0] == "note" {
					hm, _ = kit.Hash("type", "model", "name", arg[1])
					hm = m.Conf("note", []string{hm, "ship", "note", "data"})
					arg = arg[1:]
				}
				if len(arg) == 1 && arg[0] != "show" {
					for i := offset; hm != "" && i < limit+offset; hm, i = m.Conf("note", []string{hm, "ship", "prev", "data"}), i+1 {
						model := m.Confm("note", hm)
						m.Add("append", "key", hm)
						m.Add("append", "create_time", model["create_time"])
						m.Add("append", "access_time", model["access_time"])
						m.Add("append", "type", model["type"])
						m.Add("append", "name", model["name"])
						m.Add("append", "view", kit.Format(model["view"]))
						m.Add("append", "data", kit.Format(model["data"]))
						m.Add("append", "ship", kit.Format(model["ship"]))
					}
					m.Table()
					return
				}

				switch arg[0] {
				case "show":
					if len(arg) == 1 { // 查看索引
						m.Cmdy("mdb.note", "index")
						break
					}
					if len(arg) == 2 {
						if arg[1] == "model" { // 查看模型
							m.Cmdy("mdb.note", "model")
						} else { // 查看数值
							m.Cmdy("mdb.note", "value", arg[1])
						}
						break
					}

					fields := kit.View(arg[3:], m.Confm("note_view"))

					hv, _ := kit.Hash("type", "value", "name", arg[1], "data", kit.Select(arg[2], m.Option(arg[2])))
					hn := m.Conf("note", []string{hv, "ship", "note", "data"})
					if arg[1] == "model" {
						hm, _ := kit.Hash("type", "model", "name", arg[2])
						hv, hn = "prev", m.Conf("note", []string{hm, "ship", "note", "data"})
					}

					fuck := 0
					for i := 0; hn != "" && i < limit+offset; hn, i = m.Conf("note", []string{hn, "ship", hv, "data"}), i+1 {
						m.Log("fuck", "what %d %d %d %s", offset, limit, i, hn)
						// m.Log("fuck", "what hn: %v %v", hn, kit.Formats(m.Confv("note", hn)))
						// 翻页
						if fuck++; fuck > 1000 {
							break
						}

						if i < offset {
							continue
						}

						// 关系表
						note := m.Confm("note", hn)
						hvs := kit.Trans(note["data"])
						hm := kit.Format(kit.Chain(note, "ship.model.data"))

						// 值转换
						value := []interface{}{}
						values := map[string]interface{}{}
						m.Confm("note", []string{hm, "data"}, func(i int, model map[string]interface{}) {
							v := m.Conf("note", []string{hvs[i], "data"})
							value = append(value, map[string]interface{}{"type": model["type"], "name": model["name"], "value": v})
							values[kit.Format(model["name"])] = v
						})

						// 行筛选
						miss := false
						if !miss && m.Has("eq") {
							for j := 0; j < len(m.Meta["eq"]); j += 2 {
								if kit.Select(kit.Format(note[m.Meta["eq"][j]]), kit.Format(values[m.Meta["eq"][j]])) != m.Meta["eq"][j+1] {
									miss = true
									break
								}
							}
						}
						if !miss && m.Has("begin") {
							for j := 0; j < len(m.Meta["begin"]); j += 2 {
								if !strings.HasPrefix(kit.Select(kit.Format(note[m.Meta["begin"][j]]),
									kit.Format(values[m.Meta["begin"][j]])), m.Meta["begin"][j+1]) {
									miss = true
									break
								}
							}
						}
						if miss {
							i--
							continue
						}

						// 列筛选
						for j := 0; j < len(fields); j++ {
							switch fields[j] {
							case "key":
								m.Add("append", "key", hn)
							case "create_time", "access_time", "type", "name":
								m.Add("append", fields[j], note[fields[j]])
							case "model":
								m.Add("append", "model", m.Conf("note", []string{hm, "name"}))
							case "view":
								m.Add("append", "view", kit.Format(m.Conf("note", []string{hm, "view"})))
							case "value":
								m.Add("append", "value", kit.Format(value))
							case "data", "ship":
								m.Add("append", fields[j], kit.Format(note[fields[j]]))
							default:
								m.Add("append", fields[j], kit.Format(values[fields[j]]))
							}
						}
					}
					m.Table()

				case "model":
					// 模板详情
					hm, _ := kit.Hash("type", arg[0], "name", arg[1])
					if len(arg) == 2 {
						m.CopyFuck(m.Cmd("mdb.config", "note", hm), "append").Set("result").Table()
						break
					}

					// 模板视图
					if arg[2] == "view" {
						for i := 4; i < len(arg)-1; i += 2 {
							m.Conf("note", []string{hm, "view", arg[3], arg[i]}, arg[i+1])
						}
						break
					}

					// 操作模板
					data := []interface{}{}
					if model := m.Confm("note", hm); model == nil { // 添加模板
						view := map[string]interface{}{}
						m.Confm("note", "81c5709d091eb04bd31ee751c3f81023.view", func(key string, fields map[string]interface{}) {
							vs := map[string]interface{}{}
							for k, v := range fields {
								vs[k] = v
							}
							view[key] = vs
						})

						prev := m.Conf("note", []string{"81c5709d091eb04bd31ee751c3f81023", "ship", "prev", "data"})
						m.Confv("note", hm, map[string]interface{}{
							"type": "model", "name": arg[1], "data": data, "view": view,
							"create_time": m.Time(), "access_time": m.Time(),
							"ship": map[string]interface{}{
								"prev": map[string]interface{}{"type": "model", "data": prev},
								"note": map[string]interface{}{"type": "note", "data": ""},
							},
						})
						m.Conf("note", []string{"81c5709d091eb04bd31ee751c3f81023", "ship", "prev", "data"}, hm)
					} else { // 修改模板
						data = m.Confv("note", []string{hm, "data"}).([]interface{})
						m.Confv("note", []string{hm, "access_time"}, m.Time())
					}

					// 操作元素
					if len(data) == 0 {
						arg = append(arg, kit.Trans(m.Confv("note", "81c5709d091eb04bd31ee751c3f81023.meta"))...)
						for i := 2; i < len(arg)-1; i += 2 {
							data = append(data, map[string]interface{}{"name": arg[i], "type": arg[i+1]})

							hi, _ := kit.Hash("type", "index", "name", arg[i+1])
							if index := m.Confm("note", hi); index == nil {
								m.Cmd("mdb.note", "index", arg[i+1])
							}
						}
					}

					m.Confv("note", []string{hm, "data"}, data)
					m.Echo(hm)

				case "index":
					// 操作索引
					data := arg[2:]
					hi, _ := kit.Hash("type", arg[0], "name", arg[1])
					if index := m.Confm("note", hi); index == nil { // 添加索引
						prev := m.Conf("note", []string{"faa01a8fc2fc92dae3fbc02ac1b4ec75", "ship", "prev", "data"})
						m.Confv("note", hi, map[string]interface{}{
							"create_time": m.Time(), "access_time": m.Time(),
							"type": "index", "name": arg[1], "data": data, "ship": map[string]interface{}{
								"prev":  map[string]interface{}{"type": "index", "data": prev},
								"value": map[string]interface{}{"type": "value", "data": ""},
							},
						})
						m.Confv("note", []string{"faa01a8fc2fc92dae3fbc02ac1b4ec75", "ship", "prev", "data"}, hi)
					} else { // 修改索引
						m.Confv("note", []string{hi, "access_time"}, m.Time())
						data, _ = m.Confv("note", []string{hi, "data"}).([]string)
					}

					// 操作元素
					m.Confv("note", []string{hi, "data"}, data)
					m.Echo(hi)

				case "value":
					hi := m.Cmdx("mdb.note", "index", arg[1])
					hv, _ := kit.Hash("type", arg[0], "name", arg[1], "data", arg[2])
					if value := m.Confm("note", hv); value == nil {
						prev := m.Conf("note", []string{hi, "ship", "value", "data"})
						m.Confv("note", hv, map[string]interface{}{
							"create_time": m.Time(), "access_time": m.Time(),
							"type": arg[0], "name": arg[1], "data": arg[2], "ship": map[string]interface{}{
								"prev":  map[string]interface{}{"type": "value", "data": prev},
								"index": map[string]interface{}{"type": "index", "data": hi},
								"note":  map[string]interface{}{"type": "note", "data": ""},
							},
						})
						m.Conf("note", []string{hi, "ship", "value", "data"}, hv)
					} else {
						m.Confv("note", []string{hv, "access_time"}, m.Time())
					}
					m.Echo(hv)

				default:
					if len(arg) == 2 {
					}

					hm, _ := kit.Hash("type", "model", "name", arg[1])
					hn, _ := kit.Hash("type", "note", "name", arg[0], "uniq")
					hp := m.Conf("note", []string{hm, "ship", "note", "data"})
					if hp == hn {
						hp = ""
					}
					ship := map[string]interface{}{
						"prev":  map[string]interface{}{"type": "note", "data": hp},
						"model": map[string]interface{}{"type": "model", "data": hm},
					}

					data := []interface{}{}
					m.Confm("note", []string{hm, "data"}, func(i int, index map[string]interface{}) {
						hv := m.Cmdx("mdb.note", "value", index["type"], kit.Select(m.Option(kit.Format(index["name"])), arg, i+2))
						data = append(data, hv)

						ship[hv] = map[string]interface{}{"type": "note", "data": m.Conf("note", []string{hv, "ship", "note", "data"})}
						m.Conf("note", []string{hv, "ship", "note", "data"}, hn)
					})

					m.Confv("note", hn, map[string]interface{}{
						"create_time": m.Time(), "access_time": m.Time(),
						"type": "note", "name": arg[0], "data": data, "ship": ship,
					})
					m.Conf("note", []string{hm, "ship", "note", "data"}, hn)
					m.Echo(hn)
				}
				return

				sync := len(arg) > 0 && arg[0] == "sync"
				if sync {
					m.Cmdy("ssh.sh", "sub", "context", "mdb", "note", arg)
					m.Set("result").Table()
					arg = arg[1:]
				}

				h, _ := kit.Hash("uniq")
				data := map[string]interface{}{
					"create_time": m.Time(),
					"type":        arg[0],
					"title":       arg[1],
					"content":     arg[2],
				}
				for i := 3; i < len(arg)-1; i += 2 {
					kit.Chain(data, data[arg[i]], arg[i+1])
				}
				m.Conf("note", h, data)
				m.Echo(h)
				return
			}},
	},
}

func init() {
	ctx.Index.Register(Index, &MDB{Context: Index})
}
