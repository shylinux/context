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
	"strconv"
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

func (aaa *AAA) Session(meta string) string {
	bs := md5.Sum([]byte(fmt.Sprintln("%d%d%s", time.Now().Unix(), rand.Int(), meta)))
	return hex.EncodeToString(bs[:])
}
func (aaa *AAA) Password(pwd string) string {
	bs := md5.Sum([]byte(fmt.Sprintln("password:%s", pwd)))
	return hex.EncodeToString(bs[:])
}
func (aaa *AAA) Input(stream string) []byte {
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
		"sessid":      &ctx.Cache{Name: "sessid", Value: aaa.Session(arg[1]), Help: "会话令牌"},
		"expire_time": &ctx.Cache{Name: "expire_time", Value: fmt.Sprintf("%d", int64(m.Confi("expire"))+now), Help: "会话超时"},
		"login_time": &ctx.Cache{Name: "login_time", Value: fmt.Sprintf("%d", now), Help: "登录时间", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
			if len(arg) > 0 {
				return arg[0]
			}

			n, e := strconv.Atoi(x.Value)
			m.Assert(e)
			return time.Unix(int64(n), 0).Format("15:03:04")
		}},
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
		stream = aaa.Password(stream)
	case "pub":
		public, e := x509.ParsePKIXPublicKey(aaa.Decode(stream))
		m.Assert(e)

		aaa.public = public.(*rsa.PublicKey)
		stream = aaa.Password(stream)
	case "key":
		private, e := x509.ParsePKCS1PrivateKey(aaa.Decode(stream))
		m.Assert(e)

		aaa.private = private
		aaa.public = &aaa.private.PublicKey
		stream = aaa.Password(stream)
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
		"expire": &ctx.Config{Name: "expire(s)", Value: "7200", Help: "会话超时"},
		"cert":   &ctx.Config{Name: "cert", Value: "etc/pem/cert.pem", Help: "证书文件"},
		"pub":    &ctx.Config{Name: "pub", Value: "etc/pem/pub.pem", Help: "公钥文件"},
		"key":    &ctx.Config{Name: "key", Value: "etc/pem/key.pem", Help: "私钥文件"},
	},
	Commands: map[string]*ctx.Command{
		"login": &ctx.Command{
			Name: "login [sessid]|[username password]",
			Form: map[string]int{"ip": 1, "openid": 1, "cert": 1, "pub": 1, "key": 1, "load": 1, "save": 1},
			Help: []string{"会话管理", "sessid: 令牌", "username: 账号", "password: 密码",
				"ip: 主机地址", "openid: 微信登录", "cert: 证书", "pub: 公钥", "key: 私钥", "load: 加载会话", "save: 保存会话"},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if aaa, ok := c.Server.(*AAA); m.Assert(ok) {
					method := ""
					for _, v := range []string{"ip", "openid", "cert", "pub", "key"} {
						if m.Has(v) {
							method = v
						}
					}
					if method != "" {
						m.Travel(func(m *ctx.Message, n int) bool {
							if n > 0 && m.Cap("method") == method {
								switch method {
								case "ip", "openid":
									if m.Cap("stream") == m.Option(method) {
										m.Cap("expire_time", fmt.Sprintf("%d", time.Now().Unix()+int64(m.Confi("expire"))))
										m.Echo(m.Cap("sessid"))
										return false
									}
								case "cert", "pub", "key":
									if m.Cap("stream") == aaa.Password(m.Option(method)) {
										m.Cap("expire_time", fmt.Sprintf("%d", time.Now().Unix()+int64(m.Confi("expire"))))
										m.Echo(m.Cap("sessid"))
										return false
									}
								}
							}
							return true
						}, c)

						if m.Results(0) {
							return
						}

						m.Start(fmt.Sprintf("user%d", m.Capi("nuser", 1)), "用户登录", method, m.Option(method))
						m.Echo(m.Cap("sessid"))
						return
					}

					switch len(arg) {
					case 2:
						m.Travel(func(m *ctx.Message, n int) bool {
							if n > 0 && m.Cap("method") == "password" && m.Cap("stream") == arg[0] {
								m.Assert(m.Cap("password") == aaa.Password(arg[1]))
								m.Cap("expire_time", fmt.Sprintf("%d", time.Now().Unix()+int64(m.Confi("expire"))))
								m.Echo(m.Cap("sessid"))
								return false
							}
							return true
						}, c)

						if m.Results(0) {
							m.Append("sessid", m.Result(0))
							return
						}
						if arg[0] == "" {
							return
						}

						m.Start(fmt.Sprintf("user%d", m.Capi("nuser", 1)), "密码登录", "password", arg[0])
						m.Cap("password", "password", aaa.Password(arg[1]), "密码登录")
						m.Append("sessid", m.Cap("sessid"))
						m.Echo(m.Cap("sessid"))
						return
					case 1:
						m.Sess("login", nil)
						m.Travel(func(m *ctx.Message, n int) bool {
							if n > 0 && m.Cap("sessid") == arg[0] {
								if int64(m.Capi("expire_time")) > time.Now().Unix() {
									m.Sess("login", m.Target().Message())
									m.Echo(m.Cap("stream"))
								} else {
									m.Target().Close(m)
								}
								return false
							}
							return true
						}, c)
					case 0:
						m.Travel(func(m *ctx.Message, n int) bool {
							if n > 0 {
								m.Add("append", "method", m.Cap("method"))
								m.Add("append", "stream", m.Cap("stream"))
							}
							return true
						}, c)
						m.Table()
					}
				}
			}},
		"right": &ctx.Command{Name: "right [user [check|owner|share group [order] [add|del]]]", Form: map[string]int{"from": 1}, Help: "权限管理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Travel(func(m *ctx.Message, n int) bool {
				if n == 0 {
					return true
				}
				if len(arg) == 0 {
					m.Add("append", "user", m.Cap("stream"))
					m.Add("append", "right", m.Confv("right"))
					return true
				}
				if m.Cap("stream") == arg[0] {
					if len(arg) == 1 { //查看所有权
						for k, v := range m.Confv("right").(map[string]interface{}) {
							m.Add("append", "group", k)
							m.Add("append", "right", v)
						}
						return true
					}
					if arg[1] == "check" { //权限检查
						if from := m.Confv("right", []interface{}{"right", "role"}); from != nil && from.(string) == "root" {
							m.Echo("root")
						}
						if len(arg) == 2 {
							return false
						}
						if from := m.Confv("right", []interface{}{arg[2], "right", "role"}); from != nil && from.(string) == "owner" {
							m.Echo("owner")
						}
						if len(arg) == 3 {
							return false
						}
						if from := m.Confv("right", []interface{}{arg[2], arg[3], "right", "role"}); from != nil && from.(string) == "share" {
							m.Echo("share")
						}
						return false
					}
					if len(arg) == 2 { //分配人事权
						if m.Option("from") != "root" {
							return false
						}
						switch arg[1] {
						case "add":
							m.Confv("right", []interface{}{"right", "role"}, "root")
							m.Confv("right", []interface{}{"right", "from"}, m.Option("from"))
						case "del":
							m.Confv("right", []interface{}{"right", "role"}, "")
						}
						return true
					}
					if len(arg) == 3 { //查看使用权
						for k, v := range m.Confv("right", arg[2]).(map[string]interface{}) {
							m.Add("append", "order", k)
							m.Add("append", "right", v)
						}
						return true
					}
					switch arg[1] {
					case "owner": //分配所有权
						if m.Cmd("right", m.Option("from"), "check").Result(0) == "" {
							return false
						}
						switch arg[3] {
						case "add":
							m.Confv("right", []interface{}{arg[2], "right", "role"}, "owner")
							m.Confv("right", []interface{}{arg[2], "right", "from"}, m.Option("from"))
						case "del":
							m.Confv("right", []interface{}{arg[2], "right", "role"}, "")
						}
					case "share": //分配使用权
						if m.Cmd("right", m.Option("from"), "check", arg[2]).Result(0) == "" {
							return false
						}
						switch arg[4] {
						case "add":
							m.Confv("right", []interface{}{arg[2], arg[3], "right", "role"}, "share")
							m.Confv("right", []interface{}{arg[2], arg[3], "right", "from"}, m.Option("from"))
						case "del":
							m.Confv("right", []interface{}{arg[2], arg[3], "right", "role"}, "")
						}
					}
					return false
				}
				return true
			}, c)
			m.Table()
		}},
		"cert": &ctx.Command{Name: "cert [filename]", Help: "导出证书", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.certificate != nil {
				certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: aaa.certificate.Raw}))
				if m.Echo(certificate); len(arg) > 0 {
					m.Assert(ioutil.WriteFile(arg[0], []byte(certificate), 0666))
				}
			}
		}},
		"pub": &ctx.Command{Name: "pub [filename]", Help: "导出公钥", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.public != nil {
				pub, e := x509.MarshalPKIXPublicKey(aaa.public)
				m.Assert(e)
				public := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pub}))
				if m.Echo(public); len(arg) > 0 {
					m.Assert(ioutil.WriteFile(arg[0], []byte(public), 0666))
				}
			}
		}},
		"key": &ctx.Command{Name: "key [filename]", Help: "导出私钥", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.private != nil {
				private := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(aaa.private)}))
				if m.Echo(private); len(arg) > 0 {
					m.Assert(ioutil.WriteFile(arg[0], []byte(private), 0666))
				}
			}
		}},
		"sign": &ctx.Command{Name: "sign content [signfile]", Help: "数字签名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.private != nil {
				h := md5.Sum(aaa.Input(arg[0]))
				b, e := rsa.SignPKCS1v15(crand.Reader, aaa.private, crypto.MD5, h[:])
				m.Assert(e)

				res := base64.StdEncoding.EncodeToString(b)
				if m.Echo(res); len(arg) > 1 {
					m.Assert(ioutil.WriteFile(arg[1], []byte(res), 0666))
				}
			}
		}},
		"verify": &ctx.Command{Name: "verify content signature", Help: "数字验签", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.public != nil {
				buf := make([]byte, 1024)
				n, e := base64.StdEncoding.Decode(buf, aaa.Input(arg[1]))
				m.Assert(e)
				buf = buf[:n]

				h := md5.Sum(aaa.Input(arg[0]))
				m.Echo("%t", rsa.VerifyPKCS1v15(aaa.public, crypto.MD5, h[:], buf) == nil)
			}
		}},
		"seal": &ctx.Command{Name: "seal content [sealfile]", Help: "数字加密", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.public != nil {
				b, e := rsa.EncryptPKCS1v15(crand.Reader, aaa.public, aaa.Input(arg[0]))
				m.Assert(e)

				res := base64.StdEncoding.EncodeToString(b)
				if m.Echo(res); len(arg) > 1 {
					m.Assert(ioutil.WriteFile(arg[1], []byte(res), 0666))
				}
			}
		}},
		"deal": &ctx.Command{Name: "deal content", Help: "数字解密", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.private != nil {
				buf := make([]byte, 1024)
				n, e := base64.StdEncoding.Decode(buf, aaa.Input(arg[0]))
				m.Assert(e)
				buf = buf[:n]

				b, e := rsa.DecryptPKCS1v15(crand.Reader, aaa.private, buf)
				m.Assert(e)
				m.Echo(string(b))
			}
		}},
		"newcipher": &ctx.Command{Name: "newcipher salt", Help: "加密算法", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) {
				salt := md5.Sum(aaa.Input(arg[0]))
				block, e := aes.NewCipher(salt[:])
				m.Assert(e)
				aaa.encrypt = cipher.NewCBCEncrypter(block, salt[:])
				aaa.decrypt = cipher.NewCBCDecrypter(block, salt[:])
			}
		}},
		"encrypt": &ctx.Command{Name: "encrypt content [enfile]", Help: "加密数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.encrypt != nil {
				content := aaa.Input(arg[0])

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
		}},
		"decrypt": &ctx.Command{Name: "decrypt content [defile]", Help: "解密数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) && aaa.decrypt != nil {
				content := aaa.Input(arg[0])

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
		}},
		"md5": &ctx.Command{Name: "md5 content", Help: "数字摘要", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if aaa, ok := m.Target().Server.(*AAA); m.Assert(ok) {
				h := md5.Sum(aaa.Input(arg[0]))
				m.Echo(hex.EncodeToString(h[:]))
			}
		}},
		"rsa": &ctx.Command{Name: "rsa gen|sign|verify|encrypt|decrypt|cert",
			Help: []string{"gen: 生成密钥, sgin: 私钥签名, verify: 公钥验签, encrypt: 公钥加密, decrypt: 私钥解密",
				"密钥: rsa gen [keyfile [pubfile [certfile]]]",
				"加密: rsa encrypt pub content [enfile]",
				"解密: rsa decrypt key content [defile]",
				"签名: rsa sign key content [signfile]",
				"验签: rsa verify pub content",
			},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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
						m.Log("fuck", "what %#v", template)
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

						h := md5.Sum(aaa.Input(arg[2]))
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
						n, e := base64.StdEncoding.Decode(buf, aaa.Input(arg[2]))
						m.Assert(e)
						buf = buf[:n]

						h := md5.Sum(aaa.Input(arg[3]))
						m.Echo("%t", rsa.VerifyPKCS1v15(public.(*rsa.PublicKey), crypto.MD5, h[:], buf) == nil)
					case "encrypt":
						public, e := x509.ParsePKIXPublicKey(aaa.Decode(arg[1]))
						m.Assert(e)

						b, e := rsa.EncryptPKCS1v15(crand.Reader, public.(*rsa.PublicKey), aaa.Input(arg[2]))
						m.Assert(e)

						res := base64.StdEncoding.EncodeToString(b)
						if m.Echo(res); len(arg) > 3 {
							ioutil.WriteFile(arg[3], []byte(res), 0666)
						}
					case "decrypt":
						private, e := x509.ParsePKCS1PrivateKey(aaa.Decode(arg[1]))
						m.Assert(e)

						buf := make([]byte, 1024)
						n, e := base64.StdEncoding.Decode(buf, aaa.Input(arg[2]))
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
							e := recover()
							m.Log("fuck", "what %v", e)
						}()

						root, e := x509.ParseCertificate(aaa.Decode(arg[1]))
						m.Assert(e)

						cert, e := x509.ParseCertificate(aaa.Decode(arg[2]))
						m.Assert(e)

						// ee := cert.CheckSignatureFrom(root)
						// m.Echo("%v", ee)
						//
						m.Log("fuck", "----")
						pool := &x509.CertPool{}
						m.Log("fuck", "----")
						m.Echo("%c", pool)
						m.Log("fuck", "----")
						pool.AddCert(root)
						m.Log("fuck", "----")
						c, e := cert.Verify(x509.VerifyOptions{Roots: pool})
						m.Log("fuck", "----")
						m.Echo("%c", c)
					}
				}
			}},
	},
}

func init() {
	aaa := &AAA{}
	aaa.Context = Index
	ctx.Index.Register(Index, aaa)
}
