package aaa // {{{
// }}}
import ( // {{{
	"contexts"

	"bufio"
	"io"
	"io/ioutil"
	"os"

	"crypto"
	"crypto/md5"
	"strings"

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
	sessions map[string]*ctx.Message
	*ctx.Context
}

func (aaa *AAA) Session(meta string) string { // {{{
	bs := md5.Sum([]byte(fmt.Sprintln("%d%d%s", time.Now().Unix(), rand.Int(), meta)))
	return hex.EncodeToString(bs[:])
}

// }}}
func (aaa *AAA) Password(pwd string) string { // {{{
	bs := md5.Sum([]byte(fmt.Sprintln("用户密码:%s", pwd)))
	return hex.EncodeToString(bs[:])
}

// }}}

func (aaa *AAA) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"time": &ctx.Cache{Name: "time", Value: fmt.Sprintf("%d", time.Now().Unix()), Help: "登录时间", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}

			n, e := strconv.Atoi(x.Value)
			m.Assert(e)
			return time.Unix(int64(n), 0).Format("15:03:04")
			// }}}
		}},
		"username": &ctx.Cache{Name: "username", Value: arg[0], Help: "用户名"},
		"password": &ctx.Cache{Name: "password", Value: arg[1], Help: "用户密码，加密存储"},
		"sessid":   &ctx.Cache{Name: "sessid", Value: arg[2], Help: "会话令牌"},
		"expire":   &ctx.Cache{Name: "expire", Value: fmt.Sprintf("%d", int64(m.Confi("expire"))+time.Now().Unix()), Help: "会话超时"},
	}
	c.Configs = map[string]*ctx.Config{}

	s := new(AAA)
	s.Context = c
	s.sessions = aaa.sessions
	return s
}

// }}}
func (aaa *AAA) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	return aaa
}

// }}}
func (aaa *AAA) Start(m *ctx.Message, arg ...string) bool { // {{{
	aaa.sessions[m.Cap("sessid")] = m
	m.Log("info", "%d login %s", m.Capi("nuser", 1), m.Cap("stream", arg[0]))
	return false
}

// }}}
func (aaa *AAA) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch aaa.Context {
	case m.Target():
		if int64(m.Capi("expire")) > time.Now().Unix() {
			return false
		}
		delete(aaa.sessions, m.Cap("sessid"))
		m.Log("info", "%d logout %s", m.Capi("nuser", -1), m.Cap("username"))
	case m.Source():
	}
	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "aaa", Help: "认证中心",
	Caches: map[string]*ctx.Cache{
		"nuser": &ctx.Cache{Name: "nuser", Value: "0", Help: "用户数量"},
	},
	Configs: map[string]*ctx.Config{
		"rootname": &ctx.Config{Name: "rootname", Value: "root", Help: "根用户名"},
		"expire":   &ctx.Config{Name: "expire(s)", Value: "7200", Help: "会话超时"},
		"cert":     &ctx.Config{Name: "cert", Value: "etc/cert.pem", Help: "证书文件"},
		"key":      &ctx.Config{Name: "key", Value: "etc/key.pem", Help: "私钥文件"},
	},
	Commands: map[string]*ctx.Command{
		"login": &ctx.Command{
			Name: "login [sessid]|[username password]|[load|save filename]",
			Help: "用户登录, sessid: 会话ID, username: 用户名, password: 密码, load: 加载用户信息, save: 保存用户信息, filename: 文件名",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) { // {{{
					switch len(arg) {
					case 0:
						m.Travel(func(m *ctx.Message, i int) bool {
							if i > 0 {
								m.Echo("%s: %s\n", m.Cap("username"), m.Cap("sessid"))
							}
							return true
						})
					case 1:
						if msg, ok := aaa.sessions[arg[0]]; ok {
							if int64(msg.Capi("expire")) > time.Now().Unix() {
								m.Echo(msg.Cap("username"))
								m.Copy(msg, "target")
							} else {
								delete(aaa.sessions, arg[0])
								msg.Target().Close(msg)
								m.Capi("nuser", -1)
							}
						}
					default:
						switch arg[0] {
						case "load":
							if f, e := os.Open(arg[1]); m.Assert(e) {
								for bio := bufio.NewScanner(f); bio.Scan(); {
									word := strings.SplitN(bio.Text(), ":", 3)
									m.Spawn().Start(word[0], "用户", word[0], word[1], word[2])
								}
							}
						case "save":
							if f, e := os.Create(arg[1]); m.Assert(e) {
								m.Travel(func(m *ctx.Message, i int) bool {
									if i > 0 {
										f.WriteString(fmt.Sprintf("%s:%s:%s\n", m.Cap("username"), m.Cap("password"), m.Cap("sessid")))
									}
									return true
								})
							}
						default:
							if msg := m.Find(arg[0], false); msg == nil {
								m.Start(arg[0], "用户", arg[0], aaa.Password(arg[1]), aaa.Session(arg[0]))
								m.Echo(m.Cap("sessid"))
							} else if msg.Cap("password") != aaa.Password(arg[1]) {
								return
							} else {
								m.Echo(msg.Cap("sessid"))
								m.Copy(msg, "target")
							}
						}
					}
				} // }}}
			}},
		"md5": &ctx.Command{Name: "md5 [file filename][content]", Help: "散列",
			Form: map[string]int{"file": 1},
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
			Form: map[string]int{"keyfile": 1, "key": 1, "mmfile": 1, "mm": 1, "signfile": 1, "signs": 1, "file": 1},
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
			Form: map[string]int{"file": 1},
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
}

func init() {
	aaa := &AAA{}
	aaa.Context = Index
	ctx.Index.Register(Index, aaa)

	aaa.sessions = make(map[string]*ctx.Message)
}
