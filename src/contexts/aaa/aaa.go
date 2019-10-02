package aaa

import (
	"gopkg.in/gomail.v2"
	"strconv"

	"contexts/ctx"
	"toolkit"

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
)

type AAA struct {
	certificate *x509.Certificate
	public      *rsa.PublicKey
	private     *rsa.PrivateKey
	encrypt     cipher.BlockMode
	decrypt     cipher.BlockMode

	*ctx.Context
}

func Input(stream string) []byte {
	if b, e := ioutil.ReadFile(stream); e == nil {
		return b
	}
	return []byte(stream)
}
func Decode(stream string) []byte {
	block, _ := pem.Decode(Input(stream))
	return block.Bytes
}
func Password(pwd string) string {
	bs := md5.Sum([]byte(fmt.Sprintln("password:%s", pwd)))
	return hex.EncodeToString(bs[:])
}

func (aaa *AAA) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	return &AAA{Context: c}
}
func (aaa *AAA) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return aaa
}
func (aaa *AAA) Start(m *ctx.Message, arg ...string) bool {
	stream := arg[1]
	switch arg[0] {
	case "cert":
		cert, e := x509.ParseCertificate(Decode(stream))
		m.Assert(e)

		aaa.certificate = cert
		aaa.public = cert.PublicKey.(*rsa.PublicKey)
		stream = Password(stream)
	case "pub":
		public, e := x509.ParsePKIXPublicKey(Decode(stream))
		m.Assert(e)

		aaa.public = public.(*rsa.PublicKey)
		stream = Password(stream)
	case "key":
		private, e := x509.ParsePKCS1PrivateKey(Decode(stream))
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
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"hash": &ctx.Config{Name: "hash", Value: map[string]interface{}{}, Help: "散列"},
		"auth": &ctx.Config{Name: "auth", Value: map[string]interface{}{}, Help: "散列"},
		"auth_type": &ctx.Config{Name: "auth_type", Value: map[string]interface{}{
			"unique":  map[string]interface{}{"session": true, "relay": true},
			"public":  map[string]interface{}{"userrole": true, "username": true, "cert": true, "access": true},
			"single":  map[string]interface{}{"password": true, "token": true, "uuid": true, "ppid": true},
			"secrete": map[string]interface{}{"password": true, "token": true, "uuid": true, "ppid": true},
		}, Help: "散列"},

		"short": &ctx.Config{Name: "short", Value: map[string]interface{}{}, Help: "散列"},
		"email": &ctx.Config{Name: "email", Value: map[string]interface{}{
			"self": "shylinux@163.com", "smtp": "smtp.163.com", "port": "25",
		}, Help: "邮件服务"},
	},
	Commands: map[string]*ctx.Command{
		"_init": &ctx.Command{Name: "_init", Help: "初始化", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Conf("runtime", "node.cert", m.Cmdx("nfs.load", os.Getenv("node_cert")))
			m.Conf("runtime", "node.key", m.Cmdx("nfs.load", os.Getenv("node_key")))
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
		"auth": &ctx.Command{Name: "auth [id|(key val)]... [node|ship|data] [check|delete] key [val]",
			Help: "权限区块链", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				// 节点列表
				if len(arg) == 0 {
					m.Confm("auth", func(key string, node map[string]interface{}) {
						up := false
						if ship, ok := node["ship"].(map[string]interface{}); ok {
							for k, v := range ship {
								val := v.(map[string]interface{})
								switch val["ship"].(string) {
								case "0":
									if !up {
										up = true
										m.Push("up_key", k[:8])
										m.Push("up_type", val["type"])
									}
								}
							}
						}
						if !up {
							m.Push("up_key", "")
							m.Push("up_type", "")
						}

						m.Push("key", key)
						m.Push("type", node["type"])
						m.Push("meta", node["meta"])
					})
					m.Sort("type").Table()
					return
				}

				p, t, a := "", "", ""
				s, route, block, chain := "", "ship", []map[string]string{}, []map[string]string{}
				for i := 0; i < len(arg); i += 2 {
					// 查找节点
					if node := m.Confm("auth", arg[i]); node != nil {
						// 切换节点
						p, t, a = arg[i], node["type"].(string), node["meta"].(string)

						// 一级节点
						if i++; s == "" {
							s = p
						}
					}

					// 切换类型
					if i < len(arg) {
						switch arg[i] {
						case "data", "node", "ship":
							route, i = arg[i], i+1
						}
					}

					if p == "" && route != "ship" {
						return
					}

					switch route {
					// 链接操作
					case "ship":
						// 节点列表
						if i >= len(arg)-1 {
							if p == "" {
								// 所有节点
								m.Confm("auth", func(k string, node map[string]interface{}) {
									if t, _ := node["type"].(string); i > len(arg)-1 || t == arg[i] || strings.HasSuffix(k, arg[i]) || strings.HasPrefix(k, arg[i]) {
										m.Push("create_time", node["create_time"])
										m.Push("key", k)
										m.Push("type", node["type"])
										m.Push("meta", node["meta"])
									}
								})
							} else {
								// 添加关系
								if i == len(arg)-1 {
									m.Confm("auth", []string{arg[i]}, func(node map[string]interface{}) {
										m.Confv("auth", []string{p, "ship", arg[i]}, node)
									})
								}

								// 关联节点
								m.Confm("auth", []string{p, "ship"}, func(k string, ship map[string]interface{}) {
									if node := m.Confm("auth", k); i > len(arg)-1 || ship["type"].(string) == arg[i] || strings.HasSuffix(k, arg[i]) || strings.HasPrefix(k, arg[i]) {
										m.Push("create_time", node["create_time"])
										m.Push("key", k)
										m.Push("ship", ship["ship"])
										m.Push("type", node["type"])
										m.Push("meta", node["meta"])
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

						// 节点哈希
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
						h := kit.Hashs(meta)

						// 新的节点
						if !m.Confs("auth", h) {
							// 单点认证
							if m.Set("result"); m.Confs("auth_type", []string{"single", arg[i]}) && m.Confs("auth", p) && m.Cmds("aaa.auth", p, arg[i]) {
								m.Log("fuck", "password %s", h)
								// 认证失败
								return
							}

							// 创建节点
							block = append(block, map[string]string{"hash": h, "type": arg[i], "meta": meta[1]})
						}

						// 祖孙链接
						if s != "" {
							chain = append(chain, map[string]string{"node": s, "ship": "3", "hash": h, "type": arg[i], "meta": meta[1]})
							chain = append(chain, map[string]string{"node": h, "ship": "2", "hash": s, "type": arg[i], "meta": meta[1]})
						}
						// 父子链接
						if p != "" {
							chain = append(chain, map[string]string{"node": p, "ship": "1", "hash": h, "type": arg[i], "meta": meta[1]})
							chain = append(chain, map[string]string{"node": h, "ship": "0", "hash": p, "type": t, "meta": a})
						}
						// 切换节点
						p, t, a = h, arg[i], meta[1]
						m.Echo(h)

					// 节点操作
					case "node":
						if i > len(arg)-1 {
							// 查看节点
							m.Set("result").Cmdy("aaa.config", "auth", p)

						} else if arg[i] == "delete" {
							// 删除节点
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

						} else if i < len(arg)-1 {
							// 修改属性
							m.Confv("auth", []string{p, arg[i]}, arg[i+1])
							break

						} else {
							// 搜索属性
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

					// 数据操作
					case "data":
						if i > len(arg)-1 {
							// 查看数据
							m.Set("result").Echo(m.Conf("auth", strings.Join([]string{p, "data"}, ".")))

						} else if i == len(arg)-1 {
							// 查看数据
							m.Set("result").Echo(m.Conf("auth", strings.Join([]string{p, "data", arg[i]}, ".")))

						} else if arg[i] == "delete" {
							// 删除数据
							m.Confm("auth", []string{s, "data"}, func(data map[string]interface{}) {
								for _, k := range arg[i+1:] {
									m.Log("info", "delete data %s %s %s", s, k, kit.Formats(data[k]))
									delete(data, k)
								}
							})

						} else if i < len(arg)-1 {
							// 修改数据
							if m.Set("result"); arg[i] == "option" {
								m.Confv("auth", []string{p, "data", arg[i+1]}, m.Optionv(arg[i+1]))
							} else {
								m.Confv("auth", []string{p, "data", arg[i]}, arg[i+1])
							}
							m.Echo(arg[i+1])
							break

						} else {
							// 搜索数据
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

				// 添加节点
				for _, b := range block {
					m.Confv("auth", b["hash"], map[string]interface{}{"create_time": m.Time(), "type": b["type"], "meta": b["meta"]})
				}
				// 添加链接
				for _, c := range chain {
					m.Confv("auth", []interface{}{c["node"], "ship", c["hash"]}, map[string]interface{}{"ship": c["ship"], "type": c["type"], "meta": c["meta"]})
				}
				m.Log("debug", "block: %v chain: %v", len(block), len(chain))
				return
			}},

		"role": &ctx.Command{Name: "role [name [componet [name [command [name]]]]|[check [componet [command]]]|[user [name [password|cert|uuid code]]]]",
			Help: []string{"用户角色, name",
				"权限管理, componet [name [command name]]",
				"权限检查, check [componet [command]]",
				"用户管理, user [name [password|cert|uuid code]]",
			}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				// 角色列表
				if len(arg) == 0 {
					m.Cmdy("aaa.auth", "ship", "userrole")
					return
				}

				role, arg := kit.Select("void", arg[0]), arg[1:]
				switch arg[0] {
				// 权限管理
				case "componet", "command":
					// 解析参数
					componets, commands := []string{}, []string{}
					for i := 0; i < len(arg); i++ {
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

					// 组件列表
					if len(componets) == 0 {
						m.Cmdy("aaa.auth", "ship", "userrole", role, "componet")
						return
					}

					for i := 0; i < len(componets); i++ {
						// 命令列表
						if len(commands) == 0 {
							m.Cmdy("aaa.auth", "ship", "userrole", role, "componet", componets[i], "command")
							continue
						}
						// 添加命令
						for j := 0; j < len(commands); j++ {
							m.Cmd("aaa.auth", "ship", "userrole", role, "componet", componets[i], "command", commands[j])
						}
					}

				// 检查权限
				case "check":
					switch len(arg) {
					case 1: // 超级权限
						m.Echo("%t", role == "root")
					case 2: // 组件权限
						m.Echo("%t", role == "root" || m.Cmds("aaa.auth", "userrole", role, "check", "componet", arg[1]))
					case 3: // 命令权限
						m.Echo("%t", role == "root" || m.Cmds("aaa.auth", "userrole", role, "componet", arg[1], "check", "command", arg[2]))
					default: // 参数权限
						m.Echo("%t", role == "root" || m.Cmds("aaa.auth", "userrole", role, "componet", arg[1], "check", "command", arg[2]))
					}
					m.Log("right", "%v %v: %v %v %v", m.Result(0), m.Option("sessid"), m.Option("username"), role, arg[1:])

				// 用户管理
				case "user":
					// 用户列表
					if len(arg) == 1 {
						m.Cmdy("aaa.auth", "ship", "userrole", role, "username")
						break
					}

					// 添加用户
					for i := 1; i < len(arg); i++ {
						if m.Cmd("aaa.auth", "ship", "username", arg[i], "userrole", role); i < len(arg)-2 {
							switch arg[i+1] {
							case "password", "cert", "uuid":
								// 添加认证
								m.Cmd("aaa.auth", "ship", "username", arg[i], arg[i+1], arg[i+2])
								i += 2
							}
						}
					}
				}
				return
			}},
		"user": &ctx.Command{Name: "user role|login|cookie|session|key", Help: []string{"用户管理",
			"role: 查看角色",
			"login [password|cert|uuid [code]]: 用户登录",
			"cookie [spide [key [value]]]: 读写缓存",
			"session [select|create [meta]]: 会话管理",
			"key [value]: 用户数据",
		}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 用户列表
			if len(arg) == 0 {
				m.Cmdy("aaa.auth", "ship", "username")
				return
			}

			switch arg[0] {
			// 角色列表
			case "role":
				m.Cmdy("aaa.auth", "ship", "username", m.Option("username"), "userrole")

			// 用户登录
			case "login":
				m.Cmdy("aaa.auth", "username", m.Option("username"), arg[1], arg[2])

			case "cookie":
				// 设置缓存
				if len(arg) > 3 {
					m.Cmdy("aaa.auth", "username", m.Option("username"), "data", strings.Join(arg[:3], "."), arg[3])
					arg = arg[:3]
				}

				// 查看缓存
				m.Cmdy("aaa.auth", "username", m.Option("username"), "data", strings.Join(arg, "."))

			case "session":
				// 查看会话
				if len(arg) == 1 {
					m.Cmdy("aaa.auth", "ship", "username", m.Option("username"), "session")
					return
				}

				switch arg[1] {
				case "select":
					defer func() { m.Log("info", "sessid: %s", m.Result(0)) }()

					// 检查会话
					if m.Options("sessid") && m.Cmds("aaa.auth", "ship", "username", m.Option("username"), "check", m.Option("sessid")) {
						m.Echo(m.Option("sessid"))
						return
					}
					// 选择会话
					if m.Cmdy("aaa.auth", "ship", "username", m.Option("username"), "session"); m.Appends("key") {
						m.Set("result").Echo(m.Append("key"))
						return
					}
					fallthrough
				case "create":
					// 创建会话
					m.Cmdy("aaa.auth", "ship", "username", m.Option("username"), "session", kit.Select("web", arg, 2))
					m.Cmd("aaa.auth", m.Result(0), "data", "current.ctx", "mdb")
				}
			default:
				// 读写数据
				m.Option("format", "object")
				m.Cmdy("aaa.auth", "username", m.Option("username"), "data", arg)
			}
			return
		}},
		"sess": &ctx.Command{Name: "sess [sid] [user|access|key [val]]", Help: "会话管理, user: 用户列表, access: 访问管理, key [val]: 读写数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 会话列表
			if len(arg) == 0 {
				m.Cmdy("aaa.auth", "ship", "session")
				return
			}

			// 查找会话
			sid := m.Option("sessid")
			if m.Conf("auth", []string{arg[0], "type"}) == "session" {
				if sid, arg = arg[0], arg[1:]; len(arg) == 0 {
					m.Echo(sid)
					return
				}
			}

			switch arg[0] {
			// 查看用户
			case "user":
				m.Cmdy("aaa.auth", sid, "ship", "username")

			case "access":
				// 查看访问
				if len(arg) == 1 {
					m.Cmdy("aaa.auth", sid, "access")
					break
				}
				// 添加访问
				if sid != "" {
					m.Cmdy("aaa.auth", sid, "access", arg[1])
				}
				// 查看会话
				if len(arg) == 2 {
					m.Cmdy("aaa.auth", "access", arg[1], "ship", "session")
					break
				}
				// 读写数据
				m.Cmdy("aaa.auth", "access", arg[1], "data", arg[2:])

			default:
				// 读写数据
				m.Option("format", "object")
				m.Cmdy("aaa.auth", sid, "data", arg)
			}
			return
		}},
		"short": &ctx.Command{Name: "short", Help: "短码", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// cc2309e0cb95ab3cabced1b3e7141105
			if len(arg) == 0 {
				return
			}
			length := 6
			short := arg[0][:length]

			if len(arg[0]) == 32 {
				m.Confm("aaa.short", short, func(index int, value string) {
					if value == arg[0] {
						m.Echo("%s%02x", short, index)
					}
				})
				if m.Result() != "" {
					return
				}

				m.Confv("aaa.short", []string{short, "-2"}, arg[0])
				if v, ok := m.Confv("aaa.short", short).([]interface{}); ok {
					m.Echo("%s%02x", short, len(v)-1)
				}

			} else if len(arg[0]) > 0 {
				if i, e := strconv.ParseInt(arg[0][length:], 16, 64); e == nil {
					m.Echo(m.Conf("aaa.short", []interface{}{short, int(i)}))
				} else {
					m.Echo(arg[0])
				}
			}
			return
		}},
		"relay": &ctx.Command{Name: "relay [rid] [check userrole]|[count num]|[share [type [role [name [count]]]]]", Help: "授权", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 授权列表
			if len(arg) == 0 {
				m.Cmdy("aaa.auth", "relay")
				return
			}

			rid := m.Option("relay")
			if m.Confm("auth", arg[0]) != nil {
				if rid, arg = arg[0], arg[1:]; len(arg) == 0 {
					m.Echo(rid)
					return
				}
			}

			switch arg[0] {
			// 检查授权
			case "check":
				if relay := m.Confm("auth", []string{rid, "data"}); relay != nil {
					if kit.Select("", arg, 1) == "userrole" && kit.Int(relay["count"]) > 0 {
						m.Echo("%s", relay["userrole"])
					}
					for k, v := range relay {
						m.Append(k, v)
					}
					if kit.Int(relay["count"]) > 0 {
						relay["count"] = kit.Int(relay["count"]) - 1
					}
				}

			// 分享权限
			case "share":
				if len(arg) == 1 {
					m.Cmdy("aaa.auth", "username", m.Option("username"), "relay")
					break
				}
				relay := m.Cmdx("aaa.auth", "username", m.Option("username"), "relay", arg[1])
				m.Cmd("aaa.auth", relay, "data", "from", m.Option("username"), "count", "1", arg[2:])
				m.Echo(relay)

			// 授权计数
			case "count":
				m.Cmdy("aaa.auth", rid, "data", "count", kit.Select("1", arg, 1))

			case "clear":
				m.Cmdy("aaa.auth", "relay")

			default:
				m.Cmdy("aaa.auth", rid, "data", arg)
			}
			return
		}},
		"email": &ctx.Command{Name: "email name title content", Help: "发送邮件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			msg := gomail.NewMessage()
			msg.SetHeader("From", m.Conf("email", "self"))
			msg.SetHeader("To", arg[0])
			msg.SetHeader("Subject", arg[1])
			msg.SetBody("text/html", strings.Join(arg[2:], ""))
			m.Assert(gomail.NewDialer(m.Conf("email", "smtp"), kit.Int(m.Conf("email", "port")), m.Conf("email", "self"), m.Conf("email", "code")).DialAndSend(msg))
			m.Echo("success")
			return
		}},
		"location": &ctx.Command{Name: "location [latitude [longitude [location]]]", Help: "地理位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			h := m.Cmdx("aaa.auth", "username", m.Option("username"))

			// 位置列表
			if len(arg) < 2 {
				m.Confm("auth", []string{h, "data", "location"}, func(index int, value map[string]interface{}) {
					m.Push("create_time", value["create_time"])
					m.Push("index", index)
					m.Push("location", value["location"])
					m.Push("latitude", value["latitude"])
					m.Push("longitude", value["longitude"])
				})
				m.Table()
				return
			}

			switch arg[0] {
			// 添加位置
			default:
				m.Conf("auth", []string{h, "data", "location", "-2"}, map[string]interface{}{
					"create_time": m.Time(),
					"location":    kit.Select("", arg, 2),
					"latitude":    arg[0], "longitude": arg[1],
				})
				m.Echo("%d", len(m.Confv("auth", []string{h, "data", "location"}).([]interface{}))-1)
			}
			return
		}},
		"clip": &ctx.Command{Name: "clip", Help: "粘贴板", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 文本列表
			h := m.Cmdx("aaa.auth", "username", m.Option("username"))
			if len(arg) == 0 {
				m.Cmdy("aaa.config", "auth", h+".data.clip")
				return
			}

			switch arg[0] {
			// 清空列表
			case "clear":
				m.Conf("auth", []string{h, "data", "clip"}, []interface{}{})

			// 添加文本
			default:
				m.Conf("auth", []string{h, "data", "clip", kit.Select("-2", arg, 1)}, arg[0])
				m.Echo("%d", len(m.Confv("auth", []string{h, "data", "clip"}).([]interface{}))-1)
			}
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
			}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
					private, e := x509.ParsePKCS1PrivateKey(Decode(arg[1]))
					m.Assert(e)

					h := md5.Sum(Input(arg[2]))
					b, e := rsa.SignPKCS1v15(crand.Reader, private, crypto.MD5, h[:])
					m.Assert(e)

					res := base64.StdEncoding.EncodeToString(b)
					if m.Echo(res); len(arg) > 3 {
						ioutil.WriteFile(arg[3], []byte(res), 0666)
					}
				case "verify":
					public, e := x509.ParsePKIXPublicKey(Decode(arg[1]))
					if e != nil {
						cert, e := x509.ParseCertificate(Decode(arg[1]))
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
					public, e := x509.ParsePKIXPublicKey(Decode(arg[1]))
					m.Assert(e)

					b, e := rsa.EncryptPKCS1v15(crand.Reader, public.(*rsa.PublicKey), Input(arg[2]))
					m.Assert(e)

					res := base64.StdEncoding.EncodeToString(b)
					if m.Echo(res); len(arg) > 3 {
						ioutil.WriteFile(arg[3], []byte(res), 0666)
					}
				case "decrypt":
					private, e := x509.ParsePKCS1PrivateKey(Decode(arg[1]))
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
					private, e := x509.ParsePKCS1PrivateKey(Decode(arg[1]))
					m.Assert(e)

					cert, e := x509.ParseCertificate(Decode(arg[2]))
					m.Assert(e)

					public, e := x509.ParsePKIXPublicKey(Decode(arg[3]))
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
					cert, e := x509.ParseCertificate(Decode(arg[1]))
					m.Assert(e)

					var common interface{}
					json.Unmarshal([]byte(cert.Subject.CommonName), &common)
					m.Put("option", "common", common).Cmdy("ctx.trans", "common", "format", "object")
				case "grant":
					private, e := x509.ParsePKCS1PrivateKey(Decode(arg[1]))
					m.Assert(e)

					parent, e := x509.ParseCertificate(Decode(arg[2]))
					m.Assert(e)

					for _, v := range arg[3:] {
						template, e := x509.ParseCertificate(Decode(v))
						m.Assert(e)

						cert, e := x509.CreateCertificate(crand.Reader, template, parent, template.PublicKey, private)
						m.Assert(e)

						certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert}))
						m.Echo(certificate)
					}
				case "check":
					parent, e := x509.ParseCertificate(Decode(arg[1]))
					m.Assert(e)

					for _, v := range arg[2:] {
						template, e := x509.ParseCertificate(Decode(v))
						m.Assert(e)

						if e = template.CheckSignatureFrom(parent); e != nil {
							m.Echo("error: ").Echo("%v", e)
						}
					}
					m.Echo("true")
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
	ctx.Index.Register(Index, &AAA{Context: Index})
}
