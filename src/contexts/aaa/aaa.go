package aaa // {{{
// }}}
import ( // {{{
	"contexts"

	"io"
	"io/ioutil"
	"os"

	"crypto"
	"crypto/md5"

	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"encoding/hex"
	"math/rand"

	"fmt"
	"strconv"
	"time"
)

// }}}

type AAA struct {
	share    map[string]*ctx.Context
	sessions map[string]*ctx.Context
	*ctx.Context
}

func (aaa *AAA) Session(meta string) string { // {{{
	bs := md5.Sum([]byte(fmt.Sprintln("%d%d%s", time.Now().Unix(), rand.Int(), meta)))
	sessid := hex.EncodeToString(bs[:])
	return sessid
}

// }}}

func (aaa *AAA) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	c.Index = map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void", Help: "void",
			Caches:   map[string]*ctx.Cache{"group": &ctx.Cache{}},
			Configs:  map[string]*ctx.Config{"rootname": &ctx.Config{}},
			Commands: map[string]*ctx.Command{"login": &ctx.Command{}},
		},
	}

	s := new(AAA)
	s.Context = c
	return s
}

// }}}
func (aaa *AAA) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	aaa.Context.Master(nil)
	aaa.Caches["group"] = &ctx.Cache{Name: "用户组", Value: "", Help: "用户组"}
	aaa.Caches["username"] = &ctx.Cache{Name: "用户名", Value: "", Help: "用户名"}
	aaa.Caches["password"] = &ctx.Cache{Name: "用户密码", Value: "", Help: "用户密码，加密存储", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
		if len(arg) > 0 {
			bs := md5.Sum([]byte(fmt.Sprintln("用户密码:%s", arg[0])))
			m.Assert(x.Value == "" || x.Value == hex.EncodeToString(bs[:]), "密码错误")
			m.Cap("expire", fmt.Sprintf("%d", time.Now().Unix()+int64(Pulse.Confi("expire"))))
			return hex.EncodeToString(bs[:])
		}
		return x.Value
	}}

	aaa.Caches["sessid"] = &ctx.Cache{Name: "会话令牌", Value: "", Help: "用户的会话标识"}
	aaa.Caches["expire"] = &ctx.Cache{Name: "会话超时", Value: "", Help: "用户的会话标识"}
	aaa.Caches["time"] = &ctx.Cache{Name: "登录时间", Value: fmt.Sprintf("%d", time.Now().Unix()), Help: "用户登录时间", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
		if len(arg) > 0 {
			return arg[0]
		}

		n, e := strconv.Atoi(x.Value)
		m.Assert(e)
		return time.Unix(int64(n), 0).Format("15:03:04")
	}}

	if m.Target() == Index {
		Pulse = m
	}
	return aaa
}

// }}}
func (aaa *AAA) Start(m *ctx.Message, arg ...string) bool { // {{{
	if len(arg) > 1 && m.Cap("sessid") == "" {
		m.Cap("group", arg[0])
		m.Cap("username", arg[1])
		m.Cap("stream", m.Cap("username"))
		m.Cap("sessid", aaa.Session(arg[1]))
		Pulse.Capi("nuser", 1)
		aaa.Owner = aaa.Context
	}

	m.Log("info", m.Source(), "%s login %s %s", Pulse.Cap("nuser"), m.Cap("group"), m.Cap("username"))
	return false
}

// }}}
func (aaa *AAA) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch aaa.Context {
	case m.Target():
		root := Pulse.Target().Server.(*AAA)
		delete(root.sessions, m.Cap("sessid"))
		m.Log("info", nil, "%d logout %s", Pulse.Capi("nuser", -1)+1, m.Cap("username"))
	case m.Source():
	}

	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "aaa", Help: "认证中心",
	Caches: map[string]*ctx.Cache{
		"nuser": &ctx.Cache{Name: "用户数量", Value: "0", Help: "用户数量"},
	},
	Configs: map[string]*ctx.Config{
		"rootname": &ctx.Config{Name: "根用户名", Value: "root", Help: "根用户名"},
		"expire":   &ctx.Config{Name: "会话超时(s)", Value: "120", Help: "会话超时"},
	},
	Commands: map[string]*ctx.Command{
		"login": &ctx.Command{Name: "login [sessid]|[[group] username password]]", Help: "用户登录", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Target(c) // {{{
			aaa := c.Server.(*AAA)

			switch len(arg) {
			case 0:
				m.Travel(c, func(m *ctx.Message) bool {
					m.Echo("%s(%s): %s\n", m.Target().Name, m.Cap("group"), m.Cap("time"))
					if int64(m.Capi("expire")) < time.Now().Unix() {
						m.Target().Close(m)
					}
					return true
				})
			case 1:
				s, ok := aaa.sessions[arg[0]]
				m.Assert(ok, "会话失败")
				m.Target(s)
				m.Assert(int64(m.Capi("expire")) > time.Now().Unix(), "会话失败")

				m.Log("info", m.Source(), "logon %s %s", m.Cap("username"), m.Cap("group"))
				m.Echo(m.Cap("username"))
			case 2, 3:
				group, username, password := arg[0], arg[0], arg[1]
				if len(arg) == 3 {
					username, password = arg[1], arg[2]
				}

				msg := m
				if username == Pulse.Conf("rootname") {
					msg = Pulse.Spawn(Pulse.Target())
					ctx.Index.Sessions["aaa"] = msg
					msg.Set("detail", group, username).Target().Start(msg)
				} else if msg = Pulse.Find(username, false); msg == nil {
					m.Start(username, "认证用户", group, username)
					msg = m
				} else {
					m.Target(msg.Target())
				}

				msg.Target().Sessions["aaa"] = msg

				msg.Cap("password", password)

				m.Login(msg)
				aaa.sessions[m.Cap("sessid")] = msg.Target()
				m.Echo(msg.Cap("sessid"))
			} // }}}
		}},
		"share": &ctx.Command{Name: "share user", Help: "用户登录", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) == 0 { // {{{
				aaa := m.Target().Server.(*AAA)
				for k, v := range aaa.share {
					m.Echo("%s: %s", k, v.Name)
				}
				return
			}

			group := m.Sess("aaa").Cap("group")
			m.Travel(c, func(msg *ctx.Message) bool {
				aaa := msg.Target().Server.(*AAA)
				if aaa.share == nil {
					aaa.share = make(map[string]*ctx.Context)
				}
				aaa.share[group] = m.Target()
				return true
			})
			// }}}
		}},
		"md5": &ctx.Command{Name: "md5 [file filename][content]", Help: "散列",
			Formats: map[string]int{"file": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if m.Options("file") { // {{{
					f, e := os.Open(m.Option("file"))
					m.Assert(e)

					h := md5.New()
					io.Copy(h, f)

					m.Echo(hex.EncodeToString(h.Sum([]byte{})[:]))
				} else if len(arg) > 0 {
					h := md5.Sum([]byte(arg[0]))
					m.Echo(hex.EncodeToString(h[:]))
				}
				// }}}
			}},
		"rsa": &ctx.Command{Name: "rsa gen|encrypt|decrypt|sign|verify [keyfile filename][key str][mmfile filename][mm str][signfile filename][signs str][file filename] content",
			Help: ` gen生成密钥, encrypt公钥加密, decrypt私钥解密, sgin私钥签名, verify公钥验签,
					keyfile密钥文件, key密钥字符串，mm加密文件, mm加密字符串, signfile签名文件，signs签名字符串,
					file数据文件，content数据内容.
					密钥: rsa gen keyfile key.pem
					加密: rsa encrypt keyfile pubkey.pem mmfile mm.txt hello
					解密: rsa decrypt keyfile key.pem mmfile mm.txt
					签名: rsa sign keyfile key.pem signfile sign.txt hello
					验签: rsa verify keyfile pubkey.pem signfile sign.txt hello`,
			Formats: map[string]int{"keyfile": 1, "key": 1, "mmfile": 1, "mm": 1, "signfile": 1, "signs": 1, "file": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if arg[0] == "gen" { // {{{
					keys, e := rsa.GenerateKey(crand.Reader, 1024)
					m.Assert(e)

					private := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keys)}))
					m.Append("private", private)
					m.Echo(private)

					pub, e := x509.MarshalPKIXPublicKey(&keys.PublicKey)
					m.Assert(e)

					public := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pub}))
					m.Append("public", public)
					m.Echo(public)

					if m.Options("keyfile") {
						ioutil.WriteFile(m.Option("keyfile"), []byte(private), 0666)
						ioutil.WriteFile("pub"+m.Option("keyfile"), []byte(public), 0666)
					}
					return
				}

				keys := []byte(m.Option("key"))
				if m.Options("keyfile") {
					b, e := ioutil.ReadFile(m.Option("keyfile"))
					m.Assert(e)
					keys = b
				}

				block, e := pem.Decode(keys)
				m.Assert(e)

				if arg[0] == "decrypt" {
					private, e := x509.ParsePKCS1PrivateKey(block.Bytes)
					m.Assert(e)

					mm := []byte(m.Option("mm"))
					if m.Options("mmfile") {
						b, e := ioutil.ReadFile(m.Option("mmfile"))
						m.Assert(e)
						mm = b
					}

					buf := make([]byte, 1024)
					n, e := base64.StdEncoding.Decode(buf, mm)
					m.Assert(e)
					buf = buf[:n]

					b, e := rsa.DecryptPKCS1v15(crand.Reader, private, buf)
					m.Assert(e)

					m.Echo(string(b))
					if m.Options("file") {
						ioutil.WriteFile(m.Option("file"), b, 0666)
					}
					return
				}

				var content []byte
				if m.Options("file") {
					b, e := ioutil.ReadFile(m.Option("file"))
					m.Assert(e)
					content = b
				} else if len(arg) > 1 {
					content = []byte(arg[1])
				}

				switch arg[0] {
				case "encrypt":
					public, e := x509.ParsePKIXPublicKey(block.Bytes)
					m.Assert(e)

					b, e := rsa.EncryptPKCS1v15(crand.Reader, public.(*rsa.PublicKey), content)
					m.Assert(e)

					res := base64.StdEncoding.EncodeToString(b)
					m.Echo(res)
					if m.Options("mmfile") {
						ioutil.WriteFile(m.Option("mmfile"), []byte(res), 0666)
					}

				case "sign":
					private, e := x509.ParsePKCS1PrivateKey(block.Bytes)
					m.Assert(e)

					h := md5.Sum(content)
					b, e := rsa.SignPKCS1v15(crand.Reader, private, crypto.MD5, h[:])
					m.Assert(e)

					res := base64.StdEncoding.EncodeToString(b)
					m.Echo(res)

					if m.Options("signfile") {
						ioutil.WriteFile(m.Option("signfile"), []byte(res), 0666)
					}

				case "verify":
					public, e := x509.ParsePKIXPublicKey(block.Bytes)
					m.Assert(e)

					sign := []byte(m.Option("sign"))
					if m.Options("signfile") {
						b, e := ioutil.ReadFile(m.Option("signfile"))
						m.Assert(e)
						sign = b
					}

					buf := make([]byte, 1024)
					n, e := base64.StdEncoding.Decode(buf, sign)
					m.Assert(e)
					buf = buf[:n]

					h := md5.Sum(content)
					m.Echo("%t", rsa.VerifyPKCS1v15(public.(*rsa.PublicKey), crypto.MD5, h[:], buf) == nil)
				}
				// }}}
			}},
		"deal": &ctx.Command{Name: "deal init|sell|buy|done [keyfile name][key str]", Help: "散列",
			Formats: map[string]int{"file": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if m.Options("file") { // {{{
					f, e := os.Open(m.Option("file"))
					m.Assert(e)

					h := md5.New()
					io.Copy(h, f)

					m.Echo(hex.EncodeToString(h.Sum([]byte{})[:]))
				} else if len(arg) > 0 {
					h := md5.Sum([]byte(arg[0]))
					m.Echo(hex.EncodeToString(h[:]))
				}
				// }}}
			}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void", Help: "void",
			Caches:  map[string]*ctx.Cache{"group": &ctx.Cache{}},
			Configs: map[string]*ctx.Config{"rootname": &ctx.Config{}},
			Commands: map[string]*ctx.Command{
				"login": &ctx.Command{},
				"check": &ctx.Command{},
			},
		},
	},
}

func init() {
	aaa := &AAA{}
	aaa.Context = Index
	ctx.Index.Register(Index, aaa)

	aaa.sessions = make(map[string]*ctx.Context)
}
