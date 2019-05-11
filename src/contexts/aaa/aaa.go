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
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
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
			"unique":  map[string]interface{}{"session": true, "bench": true, "relay": true},
			"public":  map[string]interface{}{"userrole": true, "username": true, "cert": true},
			"single":  map[string]interface{}{"password": true, "token": true, "uuid": true, "ppid": true},
			"secrete": map[string]interface{}{"password": true, "token": true, "uuid": true, "ppid": true},
		}, Help: "散列"},

		"secrete_key": &ctx.Config{Name: "secrete_key", Value: map[string]interface{}{"password": 1, "uuid": 1}, Help: "私钥文件"},
		"expire":      &ctx.Config{Name: "expire(s)", Value: "72000", Help: "会话超时"},
		"cert":        &ctx.Config{Name: "cert", Value: "etc/pem/cert.pem", Help: "证书文件"},
		"pub":         &ctx.Config{Name: "pub", Value: "etc/pem/pub.pem", Help: "公钥文件"},
		"key":         &ctx.Config{Name: "key", Value: "etc/pem/key.pem", Help: "私钥文件"},
	},
	Commands: map[string]*ctx.Command{
		"init": &ctx.Command{Name: "init", Help: "数字摘要", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Conf("runtime", "node.cert", m.Cmdx("nfs.load", os.Getenv("node_cert")))
			m.Conf("runtime", "node.key", m.Cmdx("nfs.load", os.Getenv("node_key")))
			m.Conf("runtime", "user.cert", m.Cmdx("nfs.load", os.Getenv("user_cert")))
			m.Conf("runtime", "user.key", m.Cmdx("nfs.load", os.Getenv("user_key")))
			return
		}},
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
		"auth": &ctx.Command{Name: "auth [id] [delete data|ship|node] [[ship] type [meta]] [[data] key [val]] [[node] key [val]]",
			Help: "权限区块链", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
										m.Add("append", "up_ship", val["ship"])
									}
								}
							}
						}
						if !up {
							m.Add("append", "up_key", "")
							m.Add("append", "up_type", "")
							m.Add("append", "up_ship", "")
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
				if len(arg) == 0 { // 查看节点
					m.Echo(t)
					return
				}

				p, route, block, chain := s, "ship", []map[string]string{}, []map[string]string{}
				for i := 0; i < len(arg); i += 2 {
					if node := m.Confm("auth", arg[i]); node != nil {
						if i++; p != "" { // 朋友链接
							expire := kit.Int(m.Time(m.Conf("auth_expire"), "stamp"))
							m.Confv("auth", []string{arg[i-1], "ship", p}, map[string]interface{}{
								"create_time": m.Time(), "expire_time": expire,
								"type": t, "meta": a, "ship": "5",
							})
							m.Confv("auth", []string{p, "ship", arg[i-1]}, map[string]interface{}{
								"create_time": m.Time(), "expire_time": expire,
								"type": node["type"], "meta": node["meta"], "ship": "4",
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
						return
					}

					switch route {
					case "ship": // 链接操作
						if i >= len(arg)-1 {
							if p == "" { // 节点列表
								m.Confm("auth", func(k string, node map[string]interface{}) {
									if i > len(arg)-1 || node["type"].(string) == arg[i] || strings.HasSuffix(k, arg[i]) || strings.HasPrefix(k, arg[i]) {
										m.Add("append", "create_time", node["create_time"])
										m.Add("append", "key", k)
										m.Add("append", "type", node["type"])
										m.Add("append", "meta", node["meta"])
									}
								})
							} else { // 链接列表
								if i == len(arg)-1 {
									m.Confm("auth", []string{arg[i]}, func(node map[string]interface{}) {
										m.Confv("auth", []string{p, "ship", arg[i]}, node)
									})
								}

								m.Confm("auth", []string{p, "ship"}, func(k string, ship map[string]interface{}) {
									if node := m.Confm("auth", k); i > len(arg)-1 || ship["type"].(string) == arg[i] || strings.HasSuffix(k, arg[i]) || strings.HasPrefix(k, arg[i]) {
										m.Add("append", "create_time", node["create_time"])
										m.Add("append", "key", k)
										m.Add("append", "ship", ship["ship"])
										m.Add("append", "type", node["type"])
										m.Add("append", "meta", node["meta"])
									}
								})
							}
							m.Sort("create_time", "time_r").Set("result").Table()
							return
						}

						// 删除链接
						if arg[i] == "delete" {
							m.Confm("auth", []string{p, "ship"}, func(ship map[string]interface{}) {
								for _, k := range arg[i+1:] {
									m.Confm("auth", []string{k, "ship"}, func(peer map[string]interface{}) {
										m.Log("info", "delete peer %s %s %s", k, s, kit.Formats(peer[s]))
										delete(peer, s)
									})
									m.Log("info", "delete ship %s %s %s", s, k, kit.Formats(ship[k]))
									delete(ship, k)
								}
							})
							return
						}

						// 检查链接
						if arg[i] == "check" {
							has := "false"
							m.Confm("auth", []string{p, "ship"}, func(k string, ship map[string]interface{}) {
								if i == len(arg)-2 && (ship["meta"] != arg[i+1] && k != arg[i+1]) {
									return
								}
								if i == len(arg)-3 && (ship["type"] != arg[i+1] || ship["meta"] != arg[i+2]) {
									return
								}

								if ship["expire_time"] == nil || int64(kit.Int(ship["expire_time"])) > time.Now().Unix() {
									has = k
								}
							})
							m.Set("result").Echo(has)
							return
						}

						meta := []string{arg[i]}

						// 加密节点
						if m.Confs("auth_type", []string{"secrete", arg[i]}) {
							meta = append(meta, Password(arg[i+1]))
						} else {
							meta = append(meta, arg[i+1])
						}
						// 私有节点
						if !m.Confs("auth_type", []string{"public", arg[i]}) {
							if m.Confs("auth_type", []string{"unique", arg[i]}) {
								meta = append(meta, "uniq")
							} else {
								meta = append(meta, p)
							}
						}

						// h := m.Cmdx("aaa.hash", meta)
						h, _ := kit.Hash(meta)
						if !m.Confs("auth", h) {
							m.Set("result")
							if m.Confs("auth_type", []string{"single", arg[i]}) && m.Confs("auth", p) && m.Cmds("aaa.auth", p, arg[i]) {
								m.Log("fuck", "password %s", h)
								return // 单点认证失败
							}

							// 创建节点
							block = append(block, map[string]string{"hash": h, "type": arg[i], "meta": meta[1]})
							m.Echo(h)
						}

						if s != "" { // 祖孙链接
							chain = append(chain, map[string]string{"node": s, "ship": "3", "hash": h, "type": arg[i], "meta": meta[1]})
							chain = append(chain, map[string]string{"node": h, "ship": "2", "hash": s, "type": arg[i], "meta": meta[1]})
						}
						if p != "" { // 父子链接
							chain = append(chain, map[string]string{"node": p, "ship": "1", "hash": h, "type": arg[i], "meta": meta[1]})
							chain = append(chain, map[string]string{"node": h, "ship": "0", "hash": p, "type": t, "meta": a})
						}

						p, t, a = h, arg[i], meta[1]
						m.Echo(h)
					case "node": // 节点操作
						if i > len(arg)-1 { // 查看节点
							m.Cmdy("aaa.config", "auth", p)
						} else if arg[i] == "delete" { // 删除节点
							m.Confm("auth", []string{p, "ship"}, func(ship map[string]interface{}) {
								for k, _ := range ship {
									m.Confm("auth", []string{k, "ship"}, func(peer map[string]interface{}) {
										m.Log("info", "delete peer %s %s %s", k, s, kit.Formats(peer[s]))
										delete(peer, s)
									})
									m.Log("info", "delete ship %s %s %s", s, k, kit.Formats(ship[k]))
									delete(ship, k)
								}
								m.Log("info", "delete node %s %s", s, kit.Formats(m.Confm("auth", s)))
								delete(m.Confm("auth"), s)
							})
						} else if i < len(arg)-1 { // 修改属性
							m.Confv("auth", []string{p, arg[i]}, arg[i+1])
						} else { // 搜索属性
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
						}
						return
					case "data": // 数据操作
						if i > len(arg)-1 { // 查看数据
							m.Cmdy("ctx.config", "auth", strings.Join([]string{p, "data"}, "."))
						} else if arg[i] == "delete" { // 删除数据
							m.Confm("auth", []string{s, "data"}, func(data map[string]interface{}) {
								for _, k := range arg[i+1:] {
									m.Log("info", "delete data %s %s %s", s, k, kit.Formats(data[k]))
									delete(data, k)
								}
							})
						} else if i < len(arg)-1 { // 修改数据
							if arg[i] == "option" {
								m.Confv("auth", []string{p, "data", arg[i+1]}, m.Optionv(arg[i+1]))
							} else {
								m.Confv("auth", []string{p, "data", arg[i]}, arg[i+1])
							}
							m.Echo(arg[i+1])
						} else { // 搜索数据
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
						}
						return
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

		"role": &ctx.Command{Name: "role [name [componet [name [command [name]]]]|[user [name [password|uuid code]]]]",
			Help: "用户角色, componet: 组件管理, user: 用户管理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) == 0 { // 角色列表
					m.Cmdy("aaa.auth", "ship", "userrole")
					return
				}

				role, arg := arg[0], arg[1:]
				switch arg[0] {
				case "componet", "command":
					componets, commands := []string{}, []string{}
					for i := 0; i < len(arg); i++ { // 解析参数
						if arg[i] == "command" {
							for i = i + 1; i < len(arg); i++ {
								if arg[i] == "componet" {
									break
								}
								commands = append(commands, arg[i])
							}
							continue
						}
						if arg[i] == "componet" {
							continue
						}
						componets = append(componets, arg[i])
					}
					m.Log("info", "componet: %v, command: %v", componets, commands)

					if len(componets) == 0 { // 查看组件
						m.Cmdy("aaa.auth", "ship", "userrole", role, "componet")
						return
					}
					for i := 0; i < len(componets); i++ {
						if len(commands) == 0 { // 查看命令
							m.Cmdy("aaa.auth", "ship", "userrole", role, "componet", componets[i], "command")
							continue
						}
						for j := 0; j < len(commands); j++ { // 添加命令
							m.Cmd("aaa.auth", "ship", "userrole", role, "componet", componets[i], "command", commands[j])
						}
					}

				case "user":
					if len(arg) == 1 { // 查看用户
						m.Cmdy("aaa.auth", "ship", "userrole", role, "username")
						break
					}
					for i := 1; i < len(arg); i++ { // 添加用户
						if m.Cmd("aaa.auth", "ship", "username", arg[i], "userrole", role); i < len(arg)-2 {
							switch arg[i+1] {
							case "password", "uuid", "cert":
								m.Cmd("aaa.auth", "ship", "username", arg[i], arg[i+1], arg[i+2])
								i += 2
							}
						}
					}
				}
				return
			}},
		"user": &ctx.Command{Name: "user cookie [role]|[login [password|uuid [code]]]|[service [name [value]]]|[session [select|create]]",
			Help: "用户认证, cookie: cookie管理, session: 会话管理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) == 0 { // 查看用户
					m.Cmdy("aaa.auth", "ship", "username")
					return
				}

				switch arg[0] {
				case "role": // 用户角色
					m.Cmdy("aaa.auth", "ship", "username", m.Option("username"), "userrole")

				case "login": // 用户登录
					m.Cmdy("aaa.auth", "username", m.Option("username"), arg[1], arg[2])

				case "cookie":
					if len(arg) > 3 { // 设置cookie
						m.Cmdy("aaa.auth", "username", m.Option("username"), "data", strings.Join(arg[:3], "."), arg[3])
						arg = arg[:3]
					}

					// 查看cookie
					m.Cmdy("aaa.auth", "username", m.Option("username"), "data", strings.Join(arg, "."))
				case "session":
					if len(arg) == 1 { // 查看会话
						m.Cmdy("aaa.auth", "ship", "username", m.Option("username"), "session")
						return
					}

					switch arg[1] {
					case "select": // 选择会话
						defer func() { m.Log("info", "sessid: %s", m.Result(0)) }()

						if m.Options("sessid") && m.Cmds("aaa.auth", "ship", "username", m.Option("username"), "check", m.Option("sessid")) {
							m.Echo(m.Option("sessid"))
							return
						}
						if m.Cmdy("aaa.auth", "ship", "username", m.Option("username"), "session"); m.Appends("key") {
							m.Set("result").Echo(m.Append("key"))
							return
						}
						fallthrough
					case "create": // 创建会话
						m.Cmdy("aaa.auth", "ship", "username", m.Option("username"), "session", kit.Select("web", arg, 2))
						m.Cmd("aaa.auth", m.Result(0), "data", "current.ctx", "mdb")
					}
				}
				return
			}},
		"sess": &ctx.Command{Name: "sess [sessid] [current [pod|ctx|dir|env [value]]]|[bench [select|create]]",
			Help: "会话管理, current: 指针管理, bench: 空间管理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) == 0 { // 会话列表
					m.Cmdy("aaa.auth", "ship", "session")
					return
				}

				sid := m.Option("sessid")
				if m.Conf("auth", []string{arg[0], "type"}) == "session" {
					sid, arg = arg[0], arg[1:]
				}
				if len(arg) == 0 {
					m.Echo(sid)
					return
				}

				switch arg[0] {
				case "user": // 查看用户
					m.Cmdy("aaa.auth", sid, "ship", "username")

				case "current":
					if len(arg) > 2 { // 设置指针
						m.Cmd("aaa.auth", sid, "data", strings.Join(arg[:2], "."), arg[2])
						arg = arg[:2]
					}

					// 查看指针
					m.Cmdy("aaa.auth", sid, "data", strings.Join(arg, "."))

				case "bench":
					if len(arg) == 1 { // 查看空间
						m.Cmdy("aaa.auth", m.Option("sessid"), "ship", "bench")
						return
					}

					switch arg[1] {
					case "select": // 选择空间
						defer func() { m.Log("info", "bench: %s", m.Result(0)) }()

						if m.Options("bench") && m.Cmds("aaa.auth", m.Option("sessid"), "ship", "check", m.Option("bench")) {
							m.Echo(m.Option("bench"))
							return
						}
						if m.Cmdy("aaa.auth", m.Option("sessid"), "ship", "bench"); m.Appends("key") {
							m.Set("result").Echo(m.Append("key"))
							return
						}
						fallthrough
					case "create": // 创建空间
						m.Cmdy("aaa.auth", m.Option("sessid"), "ship", "bench", kit.Select("web", arg, 2))
						m.Cmd("aaa.auth", m.Result(0), "data", "name", "web")
					}
				}
				return
			}},
		"work": &ctx.Command{Name: "work [benchid] [sesion]|[delete]|[rename name]|[share public|protect|private][data arg...]|[right [componet [command [argument]]]]",
			Help: []string{"工作空间",
				"session: 查看会话",
				"delete: 删除空间",
				"rename [name]: 命名空间",
				"share [public|protect|private]: 共享空间",
				"data arg...: 读写数据",
				"right [componet [command [arguments]]]: 权限检查",
			}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) == 0 { // 空间列表
					m.Cmdy("aaa.auth", "ship", "bench")
					return
				}

				bid := m.Option("bench")
				if m.Conf("auth", []string{arg[0], "type"}) == "bench" {
					bid, arg = arg[0], arg[1:]
				}
				if len(arg) == 0 {
					m.Echo(bid)
					return
				}

				switch arg[0] {
				case "session": // 查看会话
					m.Cmdy("aaa.auth", bid, "ship", "session")

				case "delete": // 删除空间
					m.Cmdy("aaa.auth", bid, "delete", "node")

				case "rename": // 命名空间
					if len(arg) > 1 {
						m.Cmd("aaa.auth", bid, "data", "name", arg[1])
					}
					m.Cmdy("aaa.auth", bid, "data", "name")

				case "share": // 共享空间
					if len(arg) > 1 {
						m.Cmdy("aaa.auth", bid, "data", "share", arg[1])
					}
					m.Cmdy("aaa.auth", bid, "data", "share")

				case "data": // 读写数据
					m.Cmdy("aaa.auth", bid, arg)

				case "right":
					if len(arg) == 1 { // 查看权限
						m.Cmd("aaa.auth", m.Option("bench"), "ship", "componet").CopyTo(m, "append")
						m.Cmd("aaa.auth", m.Option("bench"), "ship", "command").CopyTo(m, "append")
						m.Table()
						return
					}

					// 检查权限
					m.Cmd("aaa.auth", "ship", "username", m.Option("username"), "userrole").Table(func(node map[string]string) {
						if m.Options("userrole") && node["meta"] != m.Option("userrole") {
							return // 失败
						} else if node["meta"] == "root" { // 超级用户

						} else if len(arg) > 2 { // 接口权限
							if m.Cmds("aaa.auth", m.Option("bench"), "ship", "check", arg[2]) {

							} else if cid := m.Cmdx("aaa.auth", "ship", "userrole", node["meta"], "componet", arg[1], "check", arg[2]); kit.Right(cid) {
								m.Cmd("aaa.auth", m.Option("bench"), cid)
							} else {
								return // 失败
							}
						} else if len(arg) > 1 { // 组件权限
							if m.Cmds("aaa.auth", m.Option("bench"), "ship", "check", arg[1]) {

							} else if cid := m.Cmdx("aaa.auth", "ship", "userrole", node["meta"], "check", arg[1]); kit.Right(cid) {
								m.Cmd("aaa.auth", m.Option("bench"), cid)
							} else {
								return // 失败
							}
						}

						m.Log("info", "role: %s %v", node["meta"], arg[1:])
						m.Echo(node["meta"])
					})

					m.Log("right", "bench: %s sessid: %s user: %s com: %v result: %v",
						m.Option("bench"), m.Option("sessid"), m.Option("username"), arg[2:], m.Result(0))
				}
				return
			}},

		"relay": &ctx.Command{Name: "relay check hash | share role", Help: "授权", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 { // 会话列表
				m.Cmdy("aaa.auth", "relay")
				return
			}

			switch arg[0] {
			case "check":
				if relay := m.Confm("auth", []string{arg[1], "data"}); relay != nil {
					if kit.Select("", arg, 2) == "userrole" && kit.Int(relay["count"]) > 0 {
						relay["count"] = kit.Int(relay["count"]) - 1
						m.Echo("%s", relay["userrole"])
					}
					for k, v := range relay {
						m.Append(k, v)
					}
				}
			case "share":
				m.Echo(m.Cmd("aaa.auth", "relay", "right").Result(0))
				m.Conf("auth", []string{m.Result(0), "data"}, map[string]interface{}{
					"userrole": kit.Select("tech", arg, 1),
					"username": kit.Select("", arg, 2),
					"count":    kit.Select("1", arg, 3),
				})
			}
			return
		}},
		"login": &ctx.Command{Name: "login nodesess", Help: "登录", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmd("aaa.auth", "username", m.Option("username"), "session", "nodes").Table(func(sess map[string]string) {
					m.Cmd("aaa.auth", sess["key"], "nodes").Table(func(node map[string]string) {
						m.Add("append", "key", node["key"])
						m.Add("append", "meta", node["meta"])
					})
				})
				m.Table()
				return
			}
			m.Cmd("aaa.auth", arg[0], "username", m.Option("username"))
			return
		}},
		"share": &ctx.Command{Name: "share nodesess", Help: "共享", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmd("aaa.auth", "username", m.Option("username"), "session", "nodes").Table(func(sess map[string]string) {
					m.Cmd("aaa.auth", sess["key"], "nodes").Table(func(node map[string]string) {
						m.Add("key", node["key"])
						m.Add("meta", node["meta"])
					})
				})
				m.Table()
				return
			}
			m.Cmd("aaa.auth", arg[0], "username", m.Option("username"))
			return
		}},

		"rsa": &ctx.Command{Name: "rsa gen|sign|verify|encrypt|decrypt|cert",
			Form: map[string]int{"common": -1},
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

						common := map[string]interface{}{}
						for i := 0; i < len(m.Meta["common"]); i += 2 {
							kit.Chain(common, m.Meta["common"][i], m.Meta["common"][i+1])
						}

						// 生成证书
						template := x509.Certificate{
							SerialNumber: big.NewInt(1),
							IsCA:         true,
							BasicConstraintsValid: true,
							KeyUsage:              x509.KeyUsageCertSign,
							Subject:               pkix.Name{CommonName: kit.Format(common)},
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
						if e != nil {
							cert, e := x509.ParseCertificate(aaa.Decode(arg[1]))
							m.Assert(e)
							public = cert.PublicKey
						}

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
					case "info":
						cert, e := x509.ParseCertificate(aaa.Decode(arg[1]))
						m.Assert(e)

						var common interface{}
						json.Unmarshal([]byte(cert.Subject.CommonName), &common)
						m.Put("option", "common", common).Cmdy("ctx.trans", "common", "format", "object")
					case "grant":
						private, e := x509.ParsePKCS1PrivateKey(aaa.Decode(arg[1]))
						m.Assert(e)

						parent, e := x509.ParseCertificate(aaa.Decode(arg[2]))
						m.Assert(e)

						for _, v := range arg[3:] {
							template, e := x509.ParseCertificate(aaa.Decode(v))
							m.Assert(e)

							cert, e := x509.CreateCertificate(crand.Reader, template, parent, template.PublicKey, private)
							m.Assert(e)

							certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert}))
							m.Echo(certificate)
						}
					case "check":
						parent, e := x509.ParseCertificate(aaa.Decode(arg[1]))
						m.Assert(e)

						for _, v := range arg[2:] {
							template, e := x509.ParseCertificate(aaa.Decode(v))
							m.Assert(e)

							if e = template.CheckSignatureFrom(parent); e != nil {
								m.Echo("error: ").Echo("%v", e)
							}
						}
						m.Echo("true")
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
