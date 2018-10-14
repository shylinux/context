package wiki

import (
	"bufio"
	"contexts/ctx"
	"contexts/web"
	"encoding/xml"
	"github.com/gomarkdown/markdown"
	"io/ioutil"
	"net/http"
	"path"

	"bytes"

	"fmt"
	"os"
	"strings"
)

type WIKI struct {
	web.WEB
}

var Index = &ctx.Context{Name: "wiki", Help: "文档中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"which":    &ctx.Config{Name: "which", Value: "redis.note", Help: "路由数量"},
		"wiki_dir": &ctx.Config{Name: "wiki_dir", Value: "usr/wiki", Help: "路由数量"},
		"wiki_list_show": &ctx.Config{Name: "wiki_list_show", Value: map[string]interface{}{
			"md": true,
		}, Help: "路由数量"},
	},
	Commands: map[string]*ctx.Command{
		"/blog": &ctx.Command{Name: "/blog", Help: "博客", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if m.Has("title") && m.Has("content") {
			}

			m.Echo("blog service")

		}},
		"/wiki_tags": &ctx.Command{Name: "/wiki_tags ", Help: "博客", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) > 0 {
				m.Option("dir", arg[0])
			}

			yac := m.Find("yac.parse4", true)

			msg := m.Sess("nfs").Cmd("dir", path.Join(m.Conf("wiki_dir"), "src", m.Option("dir")), "dir_name", "path")
			for _, v := range msg.Meta["filename"] {
				name := strings.TrimSpace(v)
				es := strings.Split(name, ".")
				switch es[len(es)-1] {
				case "pyc", "o", "gz", "tar":
					continue
				case "c":
				case "py":
				case "h":
				default:
					continue
				}

				f, e := os.Open(name)
				m.Assert(e)
				defer f.Close()

				bio := bufio.NewScanner(f)
				for line := 1; bio.Scan(); line++ {
					yac.Options("silent", true)
					l := yac.Cmd("parse", "code", "void", bio.Text())

					key := ""
					switch l.Result(1) {
					case "struct":
						switch l.Result(2) {
						case "struct", "}":
							key = l.Result(3)
						case "typedef":
							if l.Result(3) == "struct" {
								key = l.Result(5)
							}
						}
					case "function":
						switch l.Result(3) {
						case "*":
							key = l.Result(4)
						default:
							key = l.Result(3)
						}
					case "variable":
						switch l.Result(2) {
						case "struct":
							key = l.Result(4)
						}
					case "define":
						key = l.Result(3)
					}
					if key != "" {
						m.Confv("define", strings.Join([]string{key, "position", "-2"}, "."), map[string]interface{}{
							"file": strings.TrimPrefix(name, m.Confx("wiki_dir")+"/src"),
							"line": line,
							"type": l.Result(1),
						})
					}

					yac.Meta = nil
				}
			}
		}},
		"/wiki_body": &ctx.Command{Name: "/wiki_body", Help: "维基", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			which := path.Join(m.Conf("wiki_dir"), m.Confx("which"))
			st, _ := os.Stat(which)
			if ls, e := ioutil.ReadFile(which); e == nil {
				pre := false
				es := strings.Split(m.Confx("which"), ".")
				if len(es) > 0 {
					switch es[len(es)-1] {
					case "md":
						m.Option("modify_count", 1)
						m.Option("modify_time", st.ModTime().Format("2006/01/02 15:03:04"))

						switch v := m.Confv("record", []interface{}{m.Option("path"), "local", "modify_count"}).(type) {
						case int:
							if m.Confv("record", []interface{}{m.Option("path"), "local", "modify_time"}).(string) != m.Option("modify_time") {
								m.Confv("record", []interface{}{m.Option("path"), "local", "modify_time"}, m.Option("modify_time"))
								m.Confv("record", []interface{}{m.Option("path"), "local", "modify_count"}, v+1)
							}
							m.Option("modify_count", v+1)
						case float64:
							if m.Confv("record", []interface{}{m.Option("path"), "local", "modify_time"}).(string) != m.Option("modify_time") {
								m.Confv("record", []interface{}{m.Option("path"), "local", "modify_time"}, m.Option("modify_time"))
								m.Confv("record", []interface{}{m.Option("path"), "local", "modify_count"}, v+1)
							}
							m.Option("modify_count", v+1)
						case nil:
							m.Confv("record", []interface{}{m.Option("path"), "local"}, map[string]interface{}{
								"modify_count": m.Optioni("modify_count"),
								"modify_time":  m.Option("modify_time"),
							})
						default:
							m.Log("fuck", "5")
						}

						ls = markdown.ToHTML(ls, nil, nil)
					default:
						pre = true
					}
				}

				if pre {
					m.Option("nline", bytes.Count(ls, []byte("\n")))
					m.Option("nbyte", len(ls))
					m.Add("append", "code", string(ls))
					m.Add("append", "body", "")
				} else {
					m.Add("append", "body", string(ls))
					m.Add("append", "code", "")
				}
				return
			}

			if m.Options("query") {
				if v, ok := m.Confv("define", m.Option("query")).(map[string]interface{}); ok {
					for _, val := range v["position"].([]interface{}) {
						value := val.(map[string]interface{})
						m.Add("append", "name", fmt.Sprintf("src/%v#hash_%v", value["file"], value["line"]))
					}
					return
				}
				msg := m.Sess("nfs").Cmd("dir", path.Join(m.Conf("wiki_dir"), m.Option("dir")), "dir_name", "path")
				for _, v := range msg.Meta["filename"] {
					name := strings.TrimPrefix(strings.TrimSpace(v), m.Conf("wiki_dir"))
					es := strings.Split(name, ".")
					switch es[len(es)-1] {
					case "pyc", "o", "gz", "tar":
						continue
					}
					if strings.Contains(name, m.Option("query")) {
						m.Add("append", "name", name)
					}
				}
				return
			}

			msg := m.Spawn().Cmd("/wiki_list")
			m.Copy(msg, "append").Copy(msg, "option")
		}},
		"/wiki_list": &ctx.Command{Name: "/wiki_list", Help: "维基", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			ls, e := ioutil.ReadDir(path.Join(m.Conf("wiki_dir"), m.Option("which")))
			m.Option("dir", m.Option("which"))
			if e != nil {
				dir, _ := path.Split(m.Option("which"))
				m.Option("dir", dir)
				ls, e = ioutil.ReadDir(path.Join(m.Conf("wiki_dir"), dir))
			}

			parent, _ := path.Split(strings.TrimSuffix(m.Option("dir"), "/"))
			m.Option("parent", parent)
			for _, l := range ls {
				if l.Name()[0] == '.' {
					continue
				}
				if !l.IsDir() {
					es := strings.Split(l.Name(), ".")
					if len(es) > 0 {
						if show, ok := m.Confv("wiki_list_show", es[len(es)-1]).(bool); !ok || !show {
							continue
						}
					}
				}

				m.Add("append", "name", l.Name())
				m.Add("append", "time", l.ModTime().Format("2006-01-02 15:04:05"))
				if l.IsDir() {
					m.Add("append", "pend", "/")
				} else {
					m.Add("append", "pend", "")
				}
				m.Option("time_layout", "2006-01-02 15:04:05")
				m.Sort("time", "time_r")
			}
		}},
		"/wiki/": &ctx.Command{Name: "/wiki", Help: "维基", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Option("which", strings.TrimPrefix(key, "/wiki/"))
			if f, e := os.Stat(path.Join(m.Conf("wiki_dir"), m.Option("which"))); e == nil && !f.IsDir() && (strings.HasSuffix(m.Option("which"), ".json") || strings.HasSuffix(m.Option("which"), ".js") || strings.HasSuffix(m.Option("which"), ".css")) {
				m.Append("directory", path.Join(m.Conf("wiki_dir"), m.Option("which")))
				return
			}

			m.Append("template", "wiki")
		}},
		"/wx/": &ctx.Command{Name: "/wx/", Help: "微信", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if !m.Sess("aaa").Cmd("wx").Results(0) {
				return
			}
			if m.Has("echostr") {
				m.Echo(m.Option("echostr"))
				return
			}
			r := m.Optionv("request").(*http.Request)

			switch r.Header.Get("Content-Type") {
			case "text/xml":
				type Article struct {
					XMLName     xml.Name `xml:"item"`
					PicUrl      string
					Title       string
					Description string
					Url         string
				}
				type WXMsg struct {
					XMLName      xml.Name `xml:"xml"`
					ToUserName   string
					FromUserName string
					CreateTime   int32
					MsgId        int64
					MsgType      string

					Event    string
					EventKey string

					Content string

					Format      string
					Recognition string

					PicUrl  string
					MediaId string

					Location_X float64
					Location_Y float64
					Scale      int64
					Label      string

					ArticleCount int
					Articles     struct {
						XMLName  xml.Name `xml:"Articles"`
						Articles []*Article
					}
				}

				var data WXMsg

				b, e := ioutil.ReadAll(r.Body)
				e = xml.Unmarshal(b, &data)

				// de := xml.NewDecoder(r.Body)
				// e := de.Decode(&data)
				m.Assert(e)

				var echo WXMsg
				echo.FromUserName = data.ToUserName
				echo.ToUserName = data.FromUserName
				echo.CreateTime = data.CreateTime

				fs, e := ioutil.ReadDir("usr/wiki")
				m.Assert(e)
				msg := m.Spawn()
				for _, f := range fs {
					if !strings.HasSuffix(f.Name(), ".md") {
						continue
					}
					msg.Add("append", "name", f.Name())
					msg.Add("append", "title", strings.TrimSuffix(f.Name(), ".md")+"源码解析")
					msg.Add("append", "time", f.ModTime().Format("01/02 15:03"))
				}
				msg.Option("time_layout", "01/02 15:03")
				msg.Sort("time", "time_r")

				articles := []*Article{}
				articles = append(articles, &Article{PicUrl: "http://mmbiz.qpic.cn/mmbiz_jpg/sCJZHmp0V0doWEFBe6gS2HjgB0abiaK7H5WjkXGTvAI0CkCFrVJDEBBbJX8Kz0VegZ54ZoCo4We0sKJUOTuf1Tw/0",
					Title: "wiki首页", Description: "技术文章", Url: "https://shylinux.com/wiki/"})
				for i, v := range msg.Meta["title"] {
					if i > 6 {
						continue
					}

					articles = append(articles, &Article{PicUrl: "http://mmbiz.qpic.cn/mmbiz_jpg/sCJZHmp0V0doWEFBe6gS2HjgB0abiaK7H5WjkXGTvAI0CkCFrVJDEBBbJX8Kz0VegZ54ZoCo4We0sKJUOTuf1Tw/0",
						Title: msg.Meta["time"][i] + " " + v, Description: "技术文章", Url: "https://shylinux.com/wiki/" + msg.Meta["name"][i]})
				}

				switch data.MsgType {
				case "event":
					echo.MsgType = "news"
					echo.Articles.Articles = articles
					echo.ArticleCount = len(echo.Articles.Articles)
				case "text":
					echo.MsgType = "news"
					echo.Articles.Articles = articles
					echo.ArticleCount = len(echo.Articles.Articles)
				case "voice":
					echo.MsgType = "text"
					echo.Content = "你好"
				case "image":
					echo.MsgType = "text"
					echo.Content = "你好"
				case "location":
					echo.MsgType = "text"
					echo.Content = "你好"
				}

				b, e = xml.Marshal(echo)
				m.Echo(string(b))
			}
		}},
	},
}

func init() {
	wiki := &WIKI{}
	wiki.Context = Index
	web.Index.Register(Index, wiki)
}
