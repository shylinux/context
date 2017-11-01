package ssh // {{{
// }}}
import ( // {{{
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"strings"
)

// }}}

type SSH struct {
	w   io.WriteCloser
	bio *bufio.Reader

	meta map[string][]string
	data []byte
	head map[string][]string
	body []string
	len  int

	*ctx.Context
}

func (ssh *SSH) Scan() bool { // {{{
	if ssh.meta != nil {
		return false
	}

	meta := make(map[string][]string)

	for {
		l, e := ssh.bio.ReadString('\n')
		ssh.Check(e)

		log.Printf("%s >> %s", ssh.Name, l)
		if len(l) == 1 {
			break
		}

		ls := strings.SplitN(l, ":", 2)
		ls[0] = strings.TrimSpace(ls[0])
		ls[1] = strings.TrimSpace(ls[1])

		if m, ok := meta[ls[0]]; ok {
			meta[ls[0]] = append(m, ls[1])
		} else {
			meta[ls[0]] = []string{ls[1]}
		}
	}

	ssh.meta = meta
	return true
}

// }}}
func (ssh *SSH) Has(field string) bool { // {{{
	_, ok := ssh.meta[field]
	return ok
}

// }}}
func (ssh *SSH) Get(field string) string { // {{{
	if m, ok := ssh.meta[field]; ok {
		return m[0]
	}
	return ""
}

// }}}

func (ssh *SSH) Echo(name string, value interface{}) { // {{{
	fmt.Fprintln(ssh.w, name+":", value)
	log.Println(ssh.Name, "<<", name+":", value)
}

// }}}
func (ssh *SSH) Print(str string, arg ...interface{}) { // {{{
	if ssh.body == nil {
		ssh.body = make([]string, 0, 3)
	}

	s := fmt.Sprintf(str, arg...)
	ssh.body = append(ssh.body, s)
	ssh.len += len(s)
}

// }}}
func (ssh *SSH) End() { // {{{
	ssh.Echo("len", ssh.len)
	fmt.Fprintln(ssh.w)
	log.Println(ssh.Name, "<<")

	if ssh.len > 0 {
		for i, v := range ssh.body {
			n, e := fmt.Fprintf(ssh.w, v)
			log.Println(ssh.Name, "<<", i, n, v)
			ssh.Check(e)
		}
	}

	// ssh.Check(ssh.w.Flush())
	log.Println("\n")

	ssh.meta = nil
	ssh.body = nil
	ssh.len = 0
}

// }}}

func (ssh *SSH) Begin() bool { // {{{
	// ssh.Conf("log", ssh.Conf("log"))
	for k, v := range ssh.Configs {
		ssh.Conf(k, v.Value)
	}

	return true
}

// }}}
func (ssh *SSH) Start() bool { // {{{
	ssh.Begin()

	if ssh.Cap("status") == "start" {
		return false
	}

	defer ssh.Cap("status", "stop")

	if ssh.Conf("client") == "yes" {
		conn, e := tls.Dial("tcp", ssh.Conf("address"), &tls.Config{InsecureSkipVerify: true})
		ssh.Check(e)
		ssh.w = conn
		ssh.bio = bufio.NewReader(conn)
		log.Println(ssh.Name, "->", conn.RemoteAddr(), "connect")
		ssh.Cap("status", "start")

		ssh.Echo("master", ssh.Conf("master"))
		ssh.End()

		ssh.Scan()
		if ssh.Get("master") == "yes" {
			ssh.Conf("master", "no")
		} else {
			ssh.Conf("master", "yes")
		}

		log.Println("fuck")
		ssh.End()

		if ssh.Conf("master") == "yes" {
		} else {
			for ssh.Scan() {
				m := &ctx.Message{Meta: ssh.meta}
				m.Context = ssh.Context
				ssh.Root.Find([]string{"cli"}).Post(m)
				ssh.End()
			}
		}

	} else {
		pair, e := tls.LoadX509KeyPair(ssh.Conf("cert"), ssh.Conf("key"))
		ssh.Check(e)

		ls, e := tls.Listen("tcp", ssh.Conf("address"), &tls.Config{Certificates: []tls.Certificate{pair}})
		ssh.Check(e)
		ssh.Cap("status", "start")
		for {
			conn, e := ls.Accept()
			ssh.Check(e)
			log.Println(ssh.Name, "<-", conn.RemoteAddr(), "connect")

			go func() {
				s := ssh.Context.Spawn(conn.RemoteAddr().String())
				psh := s.Server.(*SSH)
				psh.w = conn
				psh.bio = bufio.NewReader(conn)

				psh.Scan()
				if psh.Get("master") == "yes" {
					psh.Conf("master", "master", "no", "主控或被控")
				} else {
					psh.Conf("master", "master", "yes", "主控或被控")
				}

				psh.Echo("master", psh.Conf("master"))
				psh.End()

				psh.Scan()

				if psh.Conf("master") == "yes" {
				} else {
					psh.meta = nil
					for psh.Scan() {
						m := &ctx.Message{Meta: psh.meta, Wait: make(chan bool)}
						psh.Root.Find([]string{"cli"}).Post(m)
						for _, v := range m.Meta["result"] {
							psh.Echo("result", v)
						}

						psh.End()
					}
				}
			}()
		}
	}
	return true
}

// }}}
func (ssh *SSH) Fork(c *ctx.Context, key string) ctx.Server { // {{{
	s := new(SSH)
	s.Context = c
	return s
}

// }}}
func (ssh *SSH) Spawn(c *ctx.Context, key string) ctx.Server { // {{{
	s := new(SSH)
	s.Context = c
	return s
}

// }}}

var Index = &ctx.Context{Name: "ssh", Help: "远程控制",
	Caches: map[string]*ctx.Cache{
		"status": &ctx.Cache{"status", "stop", "服务器状态", nil},
	},
	Configs: map[string]*ctx.Config{
		"master":  &ctx.Config{"master", "yes", "主控或被控", nil},
		"client":  &ctx.Config{"client", "yes", "连接或监听", nil},
		"address": &ctx.Config{"address", ":9090", "连接或监听的地址", nil},
		"cert":    &ctx.Config{"cert", "etc/cert.pem", "证书文件", nil},
		"key":     &ctx.Config{"key", "etc/key.pem", "私钥文件", nil},
	},
	Commands: map[string]*ctx.Command{
		"remote": &ctx.Command{"remote", "远程命令", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			ssh := c.Server.(*SSH) // {{{
			if c.Conf("master") == "yes" {
				for _, v := range arg[1:] {
					ssh.Echo("detail", v)
				}
				ssh.End()
			}
			ssh.Scan()
			for _, v := range ssh.meta["result"] {
				msg.Echo(v)
			}
			msg.Echo("\n")
			return ""
			// }}}
		}},
	},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
