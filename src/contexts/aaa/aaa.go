package aaa // {{{
// }}}
import ( // {{{
	"contexts"
	"math/big"

	"bufio"
	"io"
	"io/ioutil"
	"os"

	"crypto"
	"crypto/md5"
	"strings"

	"crypto/aes"
	"crypto/cipher"
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
	public      *rsa.PublicKey
	private     *rsa.PrivateKey
	certificate *x509.Certificate
	encrypt     cipher.BlockMode
	decrypt     cipher.BlockMode

	lark     chan *ctx.Message
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
	c.Configs = map[string]*ctx.Config{
		"lark": &ctx.Config{Name: "lark", Value: map[string]interface{}{}, Help: "用户密码，加密存储"},
	}

	s := new(AAA)
	s.Context = c
	s.sessions = aaa.sessions
	if m.Has("cert") {
		s.certificate = m.Optionv("certificate").(*x509.Certificate)
		s.public = s.certificate.PublicKey.(*rsa.PublicKey)
	}
	if m.Has("pub") {
		s.public = m.Optionv("public").(*rsa.PublicKey)
	}
	if m.Has("key") {
		s.private = m.Optionv("private").(*rsa.PrivateKey)
		s.public = &s.private.PublicKey

	}

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
	if arg[0] == "lark" {
		aaa.lark = make(chan *ctx.Message)

		for {
			msg := <-aaa.lark
			from := msg.Option("username")
			m.Log("lark", "%v", msg.Meta["detail"])
			m.Travel(func(m *ctx.Message, n int) bool {
				m.Log("fuck", "why-%v=%v", m.Cap("username"), msg.Detail(1))
				if m.Cap("username") == msg.Detail(1) {
					m.Log("fuck", "why-%v=%v", m.Cap("username"), msg.Detail(1))
					m.Confv("lark", strings.Join([]string{from, "-2"}, "."),
						map[string]interface{}{"time": msg.Time(), "type": "recv", "text": msg.Detail(2)})
				}
				return true
			}, aaa.Context)
		}
		return true
	}
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
		"pub":      &ctx.Config{Name: "pub", Value: "etc/pub.pem", Help: "公钥文件"},
		"key":      &ctx.Config{Name: "key", Value: "etc/key.pem", Help: "私钥文件"},

		"aaa_name": &ctx.Config{Name: "aaa_name", Value: "user", Help: "默认模块名", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}
			return fmt.Sprintf("%s%d", x.Value, m.Capi("nuser", 1))
			// }}}
		}},
		"aaa_help": &ctx.Config{Name: "aaa_help", Value: "登录用户", Help: "默认模块帮助"},
	},
	Commands: map[string]*ctx.Command{
		"login": &ctx.Command{
			Name: "login [sessid]|[username password]|[cert certfile]|[pub pubfile]|[key keyfile]|[ip ipstr]|[load|save filename]",
			Help: "用户登录, sessid: 会话ID, username: 用户名, password: 密码, load: 加载用户信息, save: 保存用户信息, filename: 文件名",
			Form: map[string]int{"cert": 1, "pub": 1, "key": 1, "ip": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) { // {{{
					stream := ""
					if m.Has("ip") {
						stream = m.Option("ip")
					}
					if m.Has("pub") {
						stream = m.Option("pub")
						buf, e := ioutil.ReadFile(m.Option("pub"))
						if e != nil {
							buf = []byte(m.Option("pub"))
							e = nil
							stream = "RSA PUBLIC KEY"
						}
						block, _ := pem.Decode(buf)
						public, e := x509.ParsePKIXPublicKey(block.Bytes)
						m.Assert(e)
						m.Optionv("public", public)
					}
					if m.Options("cert") {
						stream = m.Option("cert")
						buf, e := ioutil.ReadFile(m.Option("cert"))
						if e != nil {
							buf = []byte(m.Option("cert"))
							e = nil
							stream = "CERTIFICATE"
						}
						block, _ := pem.Decode(buf)
						cert, e := x509.ParseCertificate(block.Bytes)
						m.Assert(e)
						m.Optionv("certificate", cert)
					}
					if m.Has("key") {
						stream = m.Option("key")
						buf, e := ioutil.ReadFile(m.Option("key"))
						if e != nil {
							buf = []byte(m.Option("key"))
							e = nil
							stream = "RSA PRIVATE KEY"
						}
						block, buf := pem.Decode(buf)
						private, e := x509.ParsePKCS1PrivateKey(block.Bytes)
						m.Assert(e)
						m.Optionv("private", private)
					}

					if stream != "" {
						m.Start(arg[0], m.Confx("aaa_help"), arg[0], "", aaa.Session(arg[0]))
						m.Cap("stream", stream)
						return
					}

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
								m.Appendv("aaa", msg)
								m.Sess("aaa", msg)
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
									msg := m.Spawn()
									msg.Start(word[0], "用户", word[0], word[1], word[2])
									msg.Spawn().Cmd("config", "load", fmt.Sprintf("etc/%s.json", word[0]), "lark")
								}
							}
						case "save":
							if f, e := os.Create(arg[1]); m.Assert(e) {
								m.Travel(func(m *ctx.Message, i int) bool {
									if i > 0 && m.Cap("username") != "root" {
										f.WriteString(fmt.Sprintf("%s:%s:%s\n", m.Cap("username"), m.Cap("password"), m.Cap("sessid")))
										m.Spawn().Cmd("config", "save", fmt.Sprintf("etc/%s.json", m.Cap("username")), "lark")
									}
									return true
								})
							}
						default:
							find := false
							m.Travel(func(m *ctx.Message, line int) bool {
								if line > 0 && m.Cap("username") == arg[0] {
									if m.Cap("password") == aaa.Password(arg[1]) {
										m.Sess("aaa", m.Target())
										m.Echo(m.Cap("sessid"))
										m.Appendv("aaa", m)
										m.Sess("aaa", m)
									} else {
										m.Sess("aaa", c)
									}
									find = true
									return false
								}
								return true
							}, c)

							if !find {
								m.Start(arg[0], m.Confx("aaa_help"), arg[0], aaa.Password(arg[1]), aaa.Session(arg[0]))
								m.Cap("stream", arg[0])
								m.Echo(m.Cap("sessid"))
								m.Appendv("aaa", m)
								m.Sess("aaa", m)
							}
						}
					}
				} // }}}
			}},
		"certificate": &ctx.Command{Name: "certificate filename", Help: "散列",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.certificate != nil { // {{{
					certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: aaa.certificate.Raw}))
					if m.Echo(certificate); len(arg) > 0 {
						m.Assert(ioutil.WriteFile(arg[0], []byte(certificate), 0666))
					}
				}
				// }}}
			}},
		"public": &ctx.Command{Name: "public filename", Help: "散列",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.public != nil { // {{{
					pub, e := x509.MarshalPKIXPublicKey(aaa.public)
					m.Assert(e)
					public := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pub}))
					if m.Echo(public); len(arg) > 0 {
						m.Assert(ioutil.WriteFile(arg[0], []byte(public), 0666))
					}
				}
				// }}}
			}},
		"private": &ctx.Command{Name: "private filename", Help: "散列",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.private != nil { // {{{
					private := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(aaa.private)}))
					if m.Echo(private); len(arg) > 0 {
						m.Assert(ioutil.WriteFile(arg[0], []byte(private), 0666))
					}
				}
				// }}}
			}},
		"sign": &ctx.Command{Name: "sign [file filename][content] [sign signfile]", Help: "散列",
			Form: map[string]int{"file": 1, "sign": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.private != nil { // {{{
					var content []byte
					if m.Has("file") {
						b, e := ioutil.ReadFile(m.Option("file"))
						m.Assert(e)
						content = b
					} else if len(arg) > 0 {
						content = []byte(arg[0])
					}

					h := md5.Sum(content)
					b, e := rsa.SignPKCS1v15(crand.Reader, aaa.private, crypto.MD5, h[:])
					m.Assert(e)

					res := base64.StdEncoding.EncodeToString(b)
					if m.Echo(res); m.Has("sign") {
						m.Assert(ioutil.WriteFile(m.Option("sign"), []byte(res), 0666))
					}
				}
				// }}}
			}},
		"verify": &ctx.Command{Name: "verify [file filename][content] [sign signfile][signature]", Help: "散列",
			Form: map[string]int{"file": 1, "sign": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.public != nil { // {{{
					var content []byte
					if m.Has("file") {
						b, e := ioutil.ReadFile(m.Option("file"))
						m.Assert(e)
						content = b
					} else if len(arg) > 0 {
						content, arg = []byte(arg[0]), arg[1:]
					}

					var sign []byte
					if m.Has("sign") {
						b, e := ioutil.ReadFile(m.Option("sign"))
						m.Assert(e)
						sign = b
					} else if len(arg) > 0 {
						sign, arg = []byte(arg[0]), arg[1:]
					}

					buf := make([]byte, 1024)
					n, e := base64.StdEncoding.Decode(buf, sign)
					m.Assert(e)
					buf = buf[:n]

					h := md5.Sum(content)
					m.Echo("%t", rsa.VerifyPKCS1v15(aaa.public, crypto.MD5, h[:], buf) == nil)
				}
				// }}}
			}},
		"seal": &ctx.Command{Name: "seal [file filename][content] [seal sealfile]", Help: "散列",
			Form: map[string]int{"file": 1, "seal": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.public != nil { // {{{
					var content []byte
					if m.Has("file") {
						b, e := ioutil.ReadFile(m.Option("file"))
						m.Assert(e)
						content = b
					} else if len(arg) > 0 {
						content = []byte(arg[0])
					}

					b, e := rsa.EncryptPKCS1v15(crand.Reader, aaa.public, content)
					m.Assert(e)

					res := base64.StdEncoding.EncodeToString(b)
					if m.Echo(res); m.Has("seal") {
						m.Assert(ioutil.WriteFile(m.Option("seal"), []byte(res), 0666))
					}
				}
				// }}}
			}},
		"deal": &ctx.Command{Name: "deal [file filename][content]", Help: "散列",
			Form: map[string]int{"file": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.private != nil { // {{{
					var content []byte
					if m.Has("file") {
						b, e := ioutil.ReadFile(m.Option("file"))
						m.Assert(e)
						content = b
					} else if len(arg) > 0 {
						content, arg = []byte(arg[0]), arg[1:]
					}

					buf := make([]byte, 1024)
					n, e := base64.StdEncoding.Decode(buf, content)
					m.Assert(e)
					buf = buf[:n]

					b, e := rsa.DecryptPKCS1v15(crand.Reader, aaa.private, buf)
					m.Assert(e)
					m.Echo(string(b))
				}

				// }}}
			}},
		"newcipher": &ctx.Command{Name: "newcipher", Help: "散列",
			Form: map[string]int{"file": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) { // {{{
					salt := md5.Sum([]byte(arg[0]))
					block, e := aes.NewCipher(salt[:])
					m.Assert(e)
					aaa.encrypt = cipher.NewCBCEncrypter(block, salt[:])
					aaa.decrypt = cipher.NewCBCDecrypter(block, salt[:])
				}
				// }}}
			}},
		"encrypt": &ctx.Command{Name: "encrypt [file filename][content] [enfile]", Help: "散列",
			Form: map[string]int{"file": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.encrypt != nil { // {{{
					var content []byte
					if m.Has("file") {
						b, e := ioutil.ReadFile(m.Option("file"))
						m.Assert(e)
						content = b
					} else if len(arg) > 0 {
						content, arg = []byte(arg[0]), arg[1:]
					}

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
					if m.Echo(res); len(arg) > 0 {
						m.Assert(ioutil.WriteFile(arg[0], []byte(res), 0666))
					}
				}
				// }}}
			}},
		"decrypt": &ctx.Command{Name: "decrypt [file filename][content] [defile]", Help: "散列",
			Form: map[string]int{"file": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.decrypt != nil { // {{{
					var content []byte
					if m.Has("file") {
						b, e := ioutil.ReadFile(m.Option("file"))
						m.Assert(e)
						content = b
					} else if len(arg) > 0 {
						content, arg = []byte(arg[0]), arg[1:]
					}

					buf := make([]byte, 1024)
					n, e := base64.StdEncoding.Decode(buf, content)
					m.Assert(e)
					buf = buf[:n]

					res := make([]byte, n)
					aaa.decrypt.CryptBlocks(res, buf)

					if m.Echo(string(res)); len(arg) > 0 {
						m.Assert(ioutil.WriteFile(arg[0], res, 0666))
					}
				}
				// }}}
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

					template := x509.Certificate{SerialNumber: big.NewInt(1)}
					cert, e := x509.CreateCertificate(crand.Reader, &template, &template, &keys.PublicKey, keys)
					m.Assert(e)

					certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert}))
					m.Append("certificate", certificate)
					m.Echo(certificate)

					if m.Options("keyfile") {
						ioutil.WriteFile(m.Option("keyfile"), []byte(private), 0666)
						ioutil.WriteFile("pub"+m.Option("keyfile"), []byte(public), 0666)
						ioutil.WriteFile("cert"+m.Option("keyfile"), []byte(certificate), 0666)
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
		"lark": &ctx.Command{Name: "lark who message", Help: "散列",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if m.Option("username") == "" {
					m.Option("username", m.Sess("aaa", false).Cap("username"))
				}
				if aaa, ok := c.Server.(*AAA); m.Assert(ok) && aaa.lark != nil { // {{{
					m.Travel(func(m *ctx.Message, n int) bool {
						if n == 0 || m.Cap("username") != m.Option("username") {
							return true
						}

						switch len(arg) {
						case 0:
							for k, v := range m.Confv("lark").(map[string]interface{}) {
								for _, x := range v.([]interface{}) {
									val := x.(map[string]interface{})
									m.Add("append", "friend", k)
									m.Add("append", "time", val["time"])
									m.Add("append", "type", val["type"])
									if val["type"].(string) == "send" {
										m.Add("append", "text", fmt.Sprintf("<< %v", val["text"]))
									} else {
										m.Add("append", "text", fmt.Sprintf(">> %v", val["text"]))
									}
								}
							}
						case 1:
							for _, v := range m.Confv("lark", arg[0]).([]interface{}) {
								val := v.(map[string]interface{})
								m.Add("append", "time", val["time"])
								m.Add("append", "text", val["text"])
							}
						case 2:
							m.Confv("lark", strings.Join([]string{arg[0], "-2"}, "."),
								map[string]interface{}{"time": m.Time(), "type": "send", "text": arg[1]})
							aaa.lark <- m
							m.Echo("%s send done", m.Time())
						}
						return false
					}, c)
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
