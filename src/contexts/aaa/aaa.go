package aaa

import (
	"contexts/ctx"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"strings"
	"time"
	"toolkit"
)

type AAA struct {
	certificate *x509.Certificate
	public      *rsa.PublicKey
	private     *rsa.PrivateKey
	encrypt     cipher.BlockMode
	decrypt     cipher.BlockMode

	*ctx.Context
}

func Auto(m *ctx.Message, arg ...string) {
	msg := m.Spawn().Add("option", "auto_cmd", "").Cmd("auth", arg)
	msg.Table(func(maps map[string]string, list []string, line int) bool {
		if line >= 0 {
			m.Add("append", "value", maps["key"])
			m.Add("append", "name", fmt.Sprintf("%s: %s", maps["type"], maps["meta"]))
			m.Add("append", "help", fmt.Sprintf("%s", maps["create_time"]))
		}
		return true
	})
}

func Password(pwd string) string {
	bs := md5.Sum([]byte(fmt.Sprintln("password:%s", pwd)))
	return hex.EncodeToString(bs[:])
}
func Input(stream string) []byte {
	if b, e := ioutil.ReadFile(stream); e == nil {
		return b
	}
	return []byte(stream)
}
func (aaa *AAA) Decode(stream string) []byte {
	buf, e := ioutil.ReadFile(stream)
	if e != nil {
		buf = []byte(stream)
	}
	block, _ := pem.Decode(buf)
	return block.Bytes
}

func (aaa *AAA) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	now := time.Now().Unix()
	c.Caches = map[string]*ctx.Cache{
		"method":      &ctx.Cache{Name: "method", Value: arg[0], Help: "登录方式"},
		"sessid":      &ctx.Cache{Name: "sessid", Value: "", Help: "会话令牌"},
		"login_time":  &ctx.Cache{Name: "login_time", Value: fmt.Sprintf("%d", now), Help: "登录时间"},
		"expire_time": &ctx.Cache{Name: "expire_time", Value: fmt.Sprintf("%d", int64(m.Confi("expire"))+now), Help: "会话超时"},
	}
	c.Configs = map[string]*ctx.Config{
		"right": &ctx.Config{Name: "right", Value: map[string]interface{}{}, Help: "用户权限"},
	}

	s := new(AAA)
	s.Context = c
	return s
}
func (aaa *AAA) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return aaa
}
func (aaa *AAA) Start(m *ctx.Message, arg ...string) bool {
	stream := arg[1]
	switch arg[0] {
	case "cert":
		cert, e := x509.ParseCertificate(aaa.Decode(stream))
		m.Assert(e)

		aaa.certificate = cert
		aaa.public = cert.PublicKey.(*rsa.PublicKey)
		stream = Password(stream)
	case "pub":
		public, e := x509.ParsePKIXPublicKey(aaa.Decode(stream))
		m.Assert(e)

		aaa.public = public.(*rsa.PublicKey)
		stream = Password(stream)
	case "key":
		private, e := x509.ParsePKCS1PrivateKey(aaa.Decode(stream))
		m.Assert(e)

		aaa.private = private
		aaa.public = &aaa.private.PublicKey
		stream = Password(stream)
	}
	m.Log("info", "%d login %s", m.Capi("nuser"), m.Cap("stream", stream))
	return false
}
func (aaa *AAA) Close(m *ctx.Message, arg ...string) bool {
	return false
}

var Index = &ctx.Context{Name: "aaa", Help: "认证中心",
	Caches: map[string]*ctx.Cache{
		"nuser": &ctx.Cache{Name: "nuser", Value: "0", Help: "用户数量"},
	},
	Configs: map[string]*ctx.Config{
		"hash":        &ctx.Config{Name: "hash", Value: map[string]interface{}{}, Help: "散列"},
		"auth":        &ctx.Config{Name: "auth", Value: map[string]interface{}{}, Help: "散列"},
		"auth_expire": &ctx.Config{Name: "auth_expire", Value: "10m", Help: "权限超时"},
		"auth_type": &ctx.Config{Name: "auth_type", Value: map[string]interface{}{
			"session":  map[string]interface{}{"unique": true},
			"bench":    map[string]interface{}{"unique": true},
			"username": map[string]interface{}{"public": true},
			"userrole": map[string]interface{}{"public": true},
			"password": map[string]interface{}{"secrete": true, "single": true},
			"uuid":     map[string]interface{}{"secrete": true, "single": true},
			"ppid":     map[string]interface{}{"secrete": true, "single": true},
		}, Help: "散列"},

		"secrete_key": &ctx.Config{Name: "secrete_key", Value: map[string]interface{}{"password": 1, "uuid": 1}, Help: "私钥文件"},
		"expire":      &ctx.Config{Name: "expire(s)", Value: "72000", Help: "会话超时"},
		"cert":        &ctx.Config{Name: "cert", Value: "etc/pem/cert.pem", Help: "证书文件"},
		"pub":         &ctx.Config{Name: "pub", Value: "etc/pem/pub.pem", Help: "公钥文件"},
		"key":         &ctx.Config{Name: "key", Value: "etc/pem/key.pem", Help: "私钥文件"},
	},
	Commands: map[string]*ctx.Command{
		"hash": &ctx.Command{Name: "hash [meta...]", Help: "数字摘要", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmdy("ctx.config", "hash")
				return
			}

			hs, meta := kit.Hash(arg)
			m.Log("info", "%s: %v", hs, meta)
			m.Confv("hash", hs, meta)
			m.Echo(hs)
			return
		}},
		"auth": &ctx.Command{Name: "auth [id] [[ship] type [meta]] [[data] key [val]] [[node] key [val]]", Help: "权限区块链", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 { // 节点列表
				m.Confm("auth", func(key string, node map[string]interface{}) {
					up := false
					if ship, ok := node["ship"].(map[string]interface{}); ok {
						for k, v := range ship {
							val := v.(map[string]interface{})
							switch val["ship"].(string) {
							case "0":
								if !up {
									up = true
									m.Add("append", "up_key", k)
									m.Add("append", "up_type", val["type"])
								}
							}
						}
					}
					if !up {
						m.Add("append", "up_key", "")
						m.Add("append", "up_type", "")
					}
					m.Add("append", "key", key)
					m.Add("append", "type", node["type"])
					m.Add("append", "meta", node["meta"])
				})
				m.Table()
				return
			}

			s, t, a := "", "", ""
			if v := m.Confm("auth", arg[0]); v != nil {
				s, t, a, arg = arg[0], kit.Format(v["type"]), kit.Format(v["meta"]), arg[1:]
			}

			if len(arg) > 0 && arg[0] == "delete" {
				switch arg[1] {
				case "data":
					if data := m.Confm("auth", []string{s, "data"}); data != nil {
						for _, k := range arg[2:] {
							m.Log("info", "delete data %s %s %v", s, k, data[k])
							delete(data, k)
						}
					}
				case "ship":
					if ship := m.Confm("auth", []string{s, "ship"}); ship != nil {
						for _, k := range arg[2:] {
							if val, ok := ship[k].(map[string]interface{}); ok {
								m.Add("append", "key", k)
								m.Add("append", "ship", val["ship"])
								m.Add("append", "type", val["type"])
								m.Add("append", "meta", val["meta"])
							}

							m.Log("info", "delete ship %s %s %v", s, k, ship[k])
							delete(ship, k)
							if peer := m.Confm("auth", []string{k, "ship"}); peer != nil {
								m.Log("info", "delete ship %s %s %v", k, s, peer[s])
								delete(peer, s)
							}
						}
						m.Table()
					}
				case "node":
					if ship := m.Confm("auth", []string{s, "ship"}); ship != nil {
						for k, _ := range ship {
							if val, ok := ship[k].(map[string]interface{}); ok {
								m.Add("append", "key", k)
								m.Add("append", "ship", val["ship"])
								m.Add("append", "type", val["type"])
								m.Add("append", "meta", val["meta"])
							}

							m.Log("info", "delete ship %s %s %v", s, k, ship[k])
							delete(ship, k)
							if peer := m.Confm("auth", []string{k, "ship"}); peer != nil {
								m.Log("info", "delete ship %s %s %v", k, s, peer[s])
								delete(peer, s)
							}
						}
						m.Log("info", "delete node %s %v", s, m.Confm("auth", s))
						delete(m.Confm("auth"), s)
						m.Table()
					}
				}
				return
			}

			if len(arg) == 0 { // 查看节点
				m.Echo(t)
				return
			}

			p, route, block, chain := s, "ship", []map[string]string{}, []map[string]string{}
			for i := 0; i < len(arg); i += 2 {
				if p == "" {
					m.Confm("auth", func(k string, node map[string]interface{}) {
						if strings.HasSuffix(k, arg[i]) || strings.HasPrefix(k, arg[i]) {
							arg[i] = k
						}
					})
				} else {
					m.Confm("auth", []string{p, "ship"}, func(k string, ship map[string]interface{}) {
						if strings.HasSuffix(k, arg[i]) || strings.HasPrefix(k, arg[i]) {
							arg[i] = k
						}
					})
				}

				if node := m.Confm("auth", arg[i]); node != nil {
					if i++; p != "" { // 添加链接
						d, e := time.ParseDuration(m.Conf("auth_expire"))
						m.Assert(e)
						expire := time.Now().Add(d).Unix()
						m.Confv("auth", []string{p, "ship", arg[i-1]}, map[string]interface{}{
							"create_time": m.Time(), "type": node["type"], "meta": node["meta"], "ship": "4", "expire_time": expire,
						})

						m.Confv("auth", []string{arg[i-1], "ship", p}, map[string]interface{}{
							"create_time": m.Time(), "type": t, "meta": a, "ship": "5", "expire_time": expire,
						})

					}
					p, t, a = arg[i-1], node["type"].(string), node["meta"].(string)
				}

				if i < len(arg) {
					switch arg[i] { // 切换类型
					case "data", "node", "ship":
						route, i = arg[i], i+1
					}
				}

				if p == "" && route != "ship" {
					break
				}

				switch route {
				case "ship": // 链接操作
					if i > len(arg)-1 {
						m.Confm("auth", []string{p, "ship"}, func(k string, ship map[string]interface{}) {
							if node := m.Confm("auth", k); node != nil {
								m.Add("append", "key", k)
								m.Add("append", "ship", ship["ship"])
								m.Add("append", "type", node["type"])
								m.Add("append", "meta", node["meta"])
								m.Add("append", "create_time", node["create_time"])
							}
						})
						m.Set("result").Table()
						break
					} else if i == len(arg)-1 { // 读取链接
						if p == "" {
							m.Confm("auth", func(k string, node map[string]interface{}) {
								if node["type"].(string) == arg[i] || strings.HasSuffix(k, arg[i]) || strings.HasPrefix(k, arg[i]) {
									m.Add("append", "key", k)
									m.Add("append", "type", node["type"])
									m.Add("append", "meta", node["meta"])
									m.Add("append", "create_time", node["create_time"])
								}
							})
						} else {
							if node := m.Confm("auth", []string{arg[i]}); node != nil {
								m.Confv("auth", []string{p, "ship", arg[i]}, node)
							}

							m.Confm("auth", []string{p, "ship"}, func(k string, ship map[string]interface{}) {
								if node := m.Confm("auth", k); ship["type"].(string) == arg[i] || strings.HasSuffix(k, arg[i]) || strings.HasPrefix(k, arg[i]) {
									m.Add("append", "key", k)
									m.Add("append", "ship", ship["ship"])
									m.Add("append", "type", node["type"])
									m.Add("append", "meta", node["meta"])
									m.Add("append", "create_time", node["create_time"])
								}
							})
						}
						m.Set("result").Table()
						return
					} else if i == len(arg)-2 {
						if p != "" && arg[i] == "session" {
							m.Confm("auth", []string{p, "ship"}, func(k string, ship map[string]interface{}) {
								if node := m.Confm("auth", k); ship["type"].(string) == arg[i] && (ship["meta"] == arg[i+1] || strings.HasSuffix(k, arg[i+1]) || strings.HasPrefix(k, arg[i+1])) {
									m.Add("append", "key", k)
									m.Add("append", "ship", ship["ship"])
									m.Add("append", "type", node["type"])
									m.Add("append", "meta", node["meta"])
									m.Add("append", "create_time", node["create_time"])
								}
							})
							m.Set("result").Table()
							return
						}
					}

					if arg[i] == "check" {
						has := "false"
						m.Confm("auth", []string{p, "ship"}, func(k string, ship map[string]interface{}) {
							if ship["meta"] == arg[i+1] {
								if ship["expire_time"] == nil || int64(kit.Int(ship["expire_time"])) > time.Now().Unix() {
									has = k
								}
							}
						})
						m.Set("result").Echo(has)
						return
					}

					meta := []string{arg[i]}
					if m.Confs("auth_type", []string{arg[i], "secrete"}) {
						meta = append(meta, Password(arg[i+1])) // 加密节点
					} else {
						meta = append(meta, arg[i+1])
					}
					if t != "session" && !m.Confs("auth_type", []string{arg[i], "public"}) {
						meta = append(meta, p) // 私有节点
					}
					if m.Confs("auth_type", []string{arg[i], "unique"}) {
						meta = append(meta, "time", "rand") // 惟一节点
					}

					h := m.Cmdx("aaa.hash", meta)
					if !m.Confs("auth", h) {
						if m.Confs("auth_type", []string{arg[i], "single"}) && m.Confs("auth", p) && m.Cmds("aaa.auth", p, arg[i]) {
							m.Set("result")
							return // 单点认证失败
						}

						// 创建节点
						block = append(block, map[string]string{"hash": h, "type": arg[i], "meta": meta[1]})
					}

					if s != "" { // 创建根链接
						chain = append(chain, map[string]string{"node": s, "ship": "3", "hash": h, "type": arg[i], "meta": meta[1]})
						chain = append(chain, map[string]string{"node": h, "ship": "2", "hash": s, "type": arg[i], "meta": meta[1]})
					}
					if p != "" { // 创建父链接
						chain = append(chain, map[string]string{"node": p, "ship": "1", "hash": h, "type": arg[i], "meta": meta[1]})
						chain = append(chain, map[string]string{"node": h, "ship": "0", "hash": p, "type": t, "meta": a})
					}

					p, t, a = h, arg[i], meta[1]
					m.Echo(h)
				case "node": // 节点操作
					if i > len(arg)-1 { // 查看节点
						m.Cmdy("aaa.config", "auth", p)
						return
					} else if i == len(arg)-1 { // 查询节点
						ps := []string{p}
						for j := 0; j < len(ps); j++ {
							if value := m.Confv("auth", []string{ps[j], arg[i]}); value != nil {
								m.Put("option", "data", value).Cmdy("ctx.trans", "data")
								break
							}

							m.Confm("auth", []string{ps[j], "ship"}, func(key string, ship map[string]interface{}) {
								if ship["ship"] != "0" {
									ps = append(ps, key)
								}
							})
						}
						return
					} else { // 修改节点
						m.Confv("auth", []string{p, arg[i]}, arg[i+1])
					}
				case "data": // 数据操作
					if i > len(arg)-1 { // 查看数据
						m.Cmdy("ctx.config", "auth", strings.Join([]string{p, "data"}, "."))
						return
					} else if i == len(arg)-1 { // 相询数据
						ps := []string{p}
						for j := 0; j < len(ps); j++ {
							if value := m.Confv("auth", []string{ps[j], "data", arg[i]}); value != nil {
								m.Set("append").Set("result").Put("option", "data", value).Cmdy("ctx.trans", "data")
								break
							}

							m.Confm("auth", []string{ps[j], "ship"}, func(key string, ship map[string]interface{}) {
								if ship["ship"] != "0" {
									ps = append(ps, key)
								}
							})
						}
						return
					} else { // 修改数据
						if arg[i] == "option" {
							m.Confv("auth", []string{p, "data", arg[i+1]}, m.Optionv(arg[i+1]))
						} else {
							m.Confv("auth", []string{p, "data", arg[i]}, arg[i+1])
						}
					}
				}
			}

			m.Log("debug", "block: %v chain: %v", len(block), len(chain))
			for _, b := range block { // 添加节点
				m.Confv("auth", b["hash"], map[string]interface{}{"create_time": m.Time(), "type": b["type"], "meta": b["meta"]})
			}
			for _, c := range chain { // 添加链接
				m.Confv("auth", []interface{}{c["node"], "ship", c["hash"]}, map[string]interface{}{"ship": c["ship"], "type": c["type"], "meta": c["meta"]})
			}
			return
		}},
		"role": &ctx.Command{Name: "role [name [[componet] name [[command] name]]]", Help: "用户角色",
			Auto: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) bool {
				switch len(arg) {
				case 0: // 查看角色
					Auto(m, "ship", "userrole")
				case 1: // 查看组件
					Auto(m, "ship", "userrole", arg[0], "componet")
				case 2: // 查看功能
					Auto(m, "ship", "userrole", arg[0], "componet", arg[1], "command")
				}
				return true
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				switch len(arg) {
				case 0: // 查看角色
					m.Cmdy("aaa.auth", "ship", "userrole")
				case 1: // 查看组件
					m.Cmdy("aaa.auth", "ship", "userrole", arg[0], "componet")
				case 2: // 查看功能
					m.Cmdy("aaa.auth", "ship", "userrole", arg[0], "componet", arg[1], "command")
				case 3: // 查看接口
					m.Cmdy("aaa.auth", "ship", "userrole", arg[0], "componet", arg[1], "command", arg[2], "componet")
				default: // 添加接口
					if arg[1] == "componet" && arg[3] == "command" {
						for _, v := range arg[4:] {
							m.Cmdy("aaa.auth", "ship", "userrole", arg[0], "componet", arg[2], "command", v)
						}
					}
				}
				return
			}},
		"user": &ctx.Command{Name: "user [role username password] [username]", Help: "用户认证",
			Auto: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) bool {
				switch len(arg) {
				case 0: // 查看用户
					Auto(m, "ship", "username")
				case 1: // 查看会话
					Auto(m, "ship", "username", arg[0], "session")
				}
				return true
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				switch len(arg) {
				case 0: // 查看用户
					m.Cmdy("aaa.auth", "ship", "username")
				case 1: // 查看角色
					m.Cmdy("aaa.auth", "ship", "username", arg[0], "userrole")
				case 2: // 查看会话
					m.Cmdy("aaa.auth", "ship", "username", arg[0], "session", arg[1])
				case 3: // 用户认证
					if m.Cmds("aaa.auth", "ship", "username", arg[0]) && (arg[1] == "password" || arg[1] == "uuid") {
						m.Cmdy("aaa.auth", "username", arg[0], arg[1], arg[2])
						break
					}
					fallthrough
				default: // 添加用户
					for i := 1; i < len(arg); i += 2 {
						if m.Cmdy("aaa.auth", "ship", "username", arg[i], "userrole", arg[0]); i < len(arg)-1 {
							m.Cmd("aaa.auth", "ship", "username", arg[i], "password", arg[i+1])
						}
					}
				}
				return
			}},
		"sess": &ctx.Command{Name: "sess [sessid [meta]|[username password password]]|[type ip ip]", Help: "会话管理",
			Auto: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) bool {
				switch len(arg) {
				case 0: // 查看会话
					Auto(m, "ship", "session")
				case 1: // 查看会话
					Auto(m, "ship", "session", arg[0])
				}
				return true
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				switch len(arg) {
				case 0: // 查看会话
					m.Cmdy("aaa.auth", "ship", "session")
				case 1: // 查看会话
					m.Cmdy("aaa.auth", arg[0])
				case 2: // 查询会话
					m.Cmdy("aaa.auth", arg[0], "ship", arg[1])
				case 3: // 创建会话
					m.Cmdy("aaa.auth", "ship", "session", arg[0], arg[1], arg[2])
				case 4: // 用户登录
					m.Cmdy("aaa.auth", arg[0], "ship", "username", arg[1], arg[2], arg[3])
				}
				return
			}},
		"work": &ctx.Command{Name: "work [sessid create|select]|[benchid] [right [username [componet name [command name [argument name]]]]]", Help: "工作任务",
			Auto: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (goon bool) {
				switch len(arg) {
				case 0:
					Auto(m, "ship", "bench")
					Auto(m, "ship", "session")
				default:
					switch m.Conf("auth", []string{arg[0], "type"}) {
					case "session":
						if len(arg) == 1 {
							m.Auto("create", "create", "创建空间")
							m.Auto("select", "select", "查找空间")
						} else {

						}
					case "bench":
						if len(arg) == 1 {
							m.Auto("delete", "delete", "删除空间")
							m.Auto("rename", "rename", "命名空间")
							m.Auto("right", "right [username [componet [command]]]", "权限检查")
						} else {
						}
					default:
						m.Auto("invalid id")
					}
				}
				return true
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) == 0 { // 查看空间
					m.Cmdy("aaa.auth", "ship", "bench")
					return
				}

				bid := ""
				switch m.Conf("auth", []string{arg[0], "type"}) {
				case "session": // 创建空间
					bid, arg = m.Cmdx("aaa.auth", arg[0], "ship", "bench", arg[1]), arg[2:]
					m.Cmd("aaa.auth", bid, "data", "name", "web")
					defer func() { m.Set("result").Echo(bid) }()
				case "bench": // 查询空间
					bid, arg = arg[0], arg[1:]
				default:
					return
				}
				if len(arg) == 0 {
					m.Echo(bid)
					return
				}

				switch arg[0] {
				case "delete": // 删除空间
					m.Cmd("aaa.auth", bid, "delete", "node")
				case "rename": // 命名空间
					m.Cmd("aaa.auth", bid, "data", "name", arg[1])
				case "right": // 权限检查
					m.Cmd("aaa.auth", "ship", "username", arg[1], "userrole").Table(func(node map[string]string) {
						if node["meta"] == "root" { // 超级用户
							m.Log("info", "root %s", arg[1])
							m.Echo("true")
						} else if len(arg) >= 4 { // 功能权限
							if m.Cmds("aaa.auth", bid, "ship", "check", arg[3]) {
								m.Echo("true")
							} else if cid := m.Cmdx("aaa.auth", "ship", "userrole", node["meta"], "componet", arg[2], "check", arg[3]); kit.Right(cid) {
								m.Cmd("aaa.auth", bid, cid)
								m.Echo("true")
							}
						} else if len(arg) >= 3 { // 组件权限
							if m.Cmds("aaa.auth", bid, "ship", "check", arg[2]) {
								m.Echo("true")
							} else if cid := m.Cmdx("aaa.auth", "ship", "userrole", node["meta"], "check", arg[2]); kit.Right(cid) {
								m.Cmd("aaa.auth", bid, cid)
								m.Echo("true")
							}
						}
					})

					m.Log("right", "bench: %s sessid: %s user: %s com: %v result: %v",
						m.Option("bench"), m.Option("sessid"), m.Option("username"), arg[2:], m.Result(0))
				default: // 读写数据
					m.Cmdx("aaa.auth", bid, arg)
				}
				return
			}},

		"rsa": &ctx.Command{Name: "rsa gen|sign|verify|encrypt|decrypt|cert",
			Help: []string{"gen: 生成密钥, sgin: 私钥签名, verify: 公钥验签, encrypt: 公钥加密, decrypt: 私钥解密",
				"密钥: rsa gen [keyfile [pubfile [certfile]]]",
				"加密: rsa encrypt pub content [enfile]",
				"解密: rsa decrypt key content [defile]",
				"签名: rsa sign key content [signfile]",
				"验签: rsa verify pub content",
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) {
					switch arg[0] {
					case "gen":
						// 生成私钥
						keys, e := rsa.GenerateKey(crand.Reader, 1024)
						m.Assert(e)

						private := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keys)}))
						m.Echo(m.Append("private", private))

						// 生成公钥
						pub, e := x509.MarshalPKIXPublicKey(&keys.PublicKey)
						m.Assert(e)

						public := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pub}))
						m.Echo(m.Append("public", public))

						// 生成证书
						template := x509.Certificate{
							SerialNumber: big.NewInt(1),
							IsCA:         true,
							KeyUsage:     x509.KeyUsageCertSign,
						}
						cert, e := x509.CreateCertificate(crand.Reader, &template, &template, &keys.PublicKey, keys)
						m.Assert(e)

						certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert}))
						m.Echo(m.Append("certificate", certificate))

						// 输出文件
						if len(arg) > 1 {
							ioutil.WriteFile(arg[1], []byte(private), 0666)
						}
						if len(arg) > 2 {
							ioutil.WriteFile(arg[2], []byte(public), 0666)
						}
						if len(arg) > 3 {
							ioutil.WriteFile(arg[3], []byte(certificate), 0666)
						}
					case "sign":
						private, e := x509.ParsePKCS1PrivateKey(aaa.Decode(arg[1]))
						m.Assert(e)

						h := md5.Sum(Input(arg[2]))
						b, e := rsa.SignPKCS1v15(crand.Reader, private, crypto.MD5, h[:])
						m.Assert(e)

						res := base64.StdEncoding.EncodeToString(b)
						if m.Echo(res); len(arg) > 3 {
							ioutil.WriteFile(arg[3], []byte(res), 0666)
						}
					case "verify":
						public, e := x509.ParsePKIXPublicKey(aaa.Decode(arg[1]))
						m.Assert(e)

						buf := make([]byte, 1024)
						n, e := base64.StdEncoding.Decode(buf, Input(arg[2]))
						m.Assert(e)
						buf = buf[:n]

						h := md5.Sum(Input(arg[3]))
						m.Echo("%t", rsa.VerifyPKCS1v15(public.(*rsa.PublicKey), crypto.MD5, h[:], buf) == nil)
					case "encrypt":
						public, e := x509.ParsePKIXPublicKey(aaa.Decode(arg[1]))
						m.Assert(e)

						b, e := rsa.EncryptPKCS1v15(crand.Reader, public.(*rsa.PublicKey), Input(arg[2]))
						m.Assert(e)

						res := base64.StdEncoding.EncodeToString(b)
						if m.Echo(res); len(arg) > 3 {
							ioutil.WriteFile(arg[3], []byte(res), 0666)
						}
					case "decrypt":
						private, e := x509.ParsePKCS1PrivateKey(aaa.Decode(arg[1]))
						m.Assert(e)

						buf := make([]byte, 1024)
						n, e := base64.StdEncoding.Decode(buf, Input(arg[2]))
						m.Assert(e)
						buf = buf[:n]

						b, e := rsa.DecryptPKCS1v15(crand.Reader, private, buf)
						m.Assert(e)

						if m.Echo(string(b)); len(arg) > 3 {
							ioutil.WriteFile(arg[3], b, 0666)
						}
					case "cert":
						private, e := x509.ParsePKCS1PrivateKey(aaa.Decode(arg[1]))
						m.Assert(e)

						cert, e := x509.ParseCertificate(aaa.Decode(arg[2]))
						m.Assert(e)

						public, e := x509.ParsePKIXPublicKey(aaa.Decode(arg[3]))
						m.Assert(e)

						template := &x509.Certificate{
							SerialNumber: big.NewInt(rand.Int63()),
							NotBefore:    time.Now(),
							NotAfter:     time.Now().AddDate(1, 0, 0),
						}
						buf, e := x509.CreateCertificate(crand.Reader, template, cert, public, private)

						certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: buf}))
						if m.Echo(certificate); len(arg) > 4 {
							ioutil.WriteFile(arg[4], []byte(certificate), 0666)
						}
					case "check":
						defer func() {
							recover()
						}()

						root, e := x509.ParseCertificate(aaa.Decode(arg[1]))
						m.Assert(e)

						cert, e := x509.ParseCertificate(aaa.Decode(arg[2]))
						m.Assert(e)

						// ee := cert.CheckSignatureFrom(root)
						// m.Echo("%v", ee)
						//
						pool := &x509.CertPool{}
						m.Echo("%c", pool)
						pool.AddCert(root)
						c, e := cert.Verify(x509.VerifyOptions{Roots: pool})
						m.Echo("%c", c)
					}
				}
				return
			}},
		"cert": &ctx.Command{Name: "cert [filename]", Help: "导出证书", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.certificate != nil {
				certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: aaa.certificate.Raw}))
				if m.Echo(certificate); len(arg) > 0 {
					m.Assert(ioutil.WriteFile(arg[0], []byte(certificate), 0666))
				}
			}
			return
		}},
		"pub": &ctx.Command{Name: "pub [filename]", Help: "导出公钥", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.public != nil {
				pub, e := x509.MarshalPKIXPublicKey(aaa.public)
				m.Assert(e)
				public := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pub}))
				if m.Echo(public); len(arg) > 0 {
					m.Assert(ioutil.WriteFile(arg[0], []byte(public), 0666))
				}
			}
			return
		}},
		"keys": &ctx.Command{Name: "keys [filename]", Help: "导出私钥", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.private != nil {
				private := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(aaa.private)}))
				if m.Echo(private); len(arg) > 0 {
					m.Assert(ioutil.WriteFile(arg[0], []byte(private), 0666))
				}
			}
			return
		}},
		"sign": &ctx.Command{Name: "sign content [signfile]", Help: "数字签名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.private != nil {
				h := md5.Sum(Input(arg[0]))
				b, e := rsa.SignPKCS1v15(crand.Reader, aaa.private, crypto.MD5, h[:])
				m.Assert(e)

				res := base64.StdEncoding.EncodeToString(b)
				if m.Echo(res); len(arg) > 1 {
					m.Assert(ioutil.WriteFile(arg[1], []byte(res), 0666))
				}
			}
			return
		}},
		"verify": &ctx.Command{Name: "verify content signature", Help: "数字验签", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.public != nil {
				buf := make([]byte, 1024)
				n, e := base64.StdEncoding.Decode(buf, Input(arg[1]))
				m.Assert(e)
				buf = buf[:n]

				h := md5.Sum(Input(arg[0]))
				m.Echo("%t", rsa.VerifyPKCS1v15(aaa.public, crypto.MD5, h[:], buf) == nil)
			}
			return
		}},
		"seal": &ctx.Command{Name: "seal content [sealfile]", Help: "数字加密", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.public != nil {
				b, e := rsa.EncryptPKCS1v15(crand.Reader, aaa.public, Input(arg[0]))
				m.Assert(e)

				res := base64.StdEncoding.EncodeToString(b)
				if m.Echo(res); len(arg) > 1 {
					m.Assert(ioutil.WriteFile(arg[1], []byte(res), 0666))
				}
			}
			return
		}},
		"deal": &ctx.Command{Name: "deal content", Help: "数字解密", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.private != nil {
				buf := make([]byte, 1024)
				n, e := base64.StdEncoding.Decode(buf, Input(arg[0]))
				m.Assert(e)
				buf = buf[:n]

				b, e := rsa.DecryptPKCS1v15(crand.Reader, aaa.private, buf)
				m.Assert(e)
				m.Echo(string(b))
			}
			return
		}},

		"newcipher": &ctx.Command{Name: "newcipher salt", Help: "加密算法", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) {
				salt := md5.Sum(Input(arg[0]))
				block, e := aes.NewCipher(salt[:])
				m.Assert(e)
				aaa.encrypt = cipher.NewCBCEncrypter(block, salt[:])
				aaa.decrypt = cipher.NewCBCDecrypter(block, salt[:])
			}
			return
		}},
		"encrypt": &ctx.Command{Name: "encrypt content [enfile]", Help: "加密数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.encrypt != nil {
				content := Input(arg[0])

				bsize := aaa.encrypt.BlockSize()
				size := (len(content) / bsize) * bsize
				if len(content)%bsize != 0 {
					size += bsize
				}

				buf := make([]byte, size)
				for pos := 0; pos < len(content); pos += bsize {
					end := pos + bsize
					if end > len(content) {
						end = len(content)
					}

					b := make([]byte, bsize)
					copy(b, content[pos:end])

					aaa.encrypt.CryptBlocks(buf[pos:pos+bsize], b)
				}

				res := base64.StdEncoding.EncodeToString(buf)
				if m.Echo(res); len(arg) > 1 {
					m.Assert(ioutil.WriteFile(arg[1], []byte(res), 0666))
				}
			}
			return
		}},
		"decrypt": &ctx.Command{Name: "decrypt content [defile]", Help: "解密数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.decrypt != nil {
				content := Input(arg[0])

				buf := make([]byte, 1024)
				n, e := base64.StdEncoding.Decode(buf, content)
				m.Assert(e)
				buf = buf[:n]

				res := make([]byte, n)
				aaa.decrypt.CryptBlocks(res, buf)

				if m.Echo(string(res)); len(arg) > 1 {
					m.Assert(ioutil.WriteFile(arg[1], res, 0666))
				}
			}
			return
		}},
	},
}

func init() {
	aaa := &AAA{}
	aaa.Context = Index
	ctx.Index.Register(Index, aaa)
}
