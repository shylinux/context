package wiki

import (
	"github.com/gomarkdown/markdown"

	"contexts/ctx"
	"contexts/web"
	"toolkit"

	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

var Index = &ctx.Context{Name: "wiki", Help: "文档中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"login": &ctx.Config{Name: "login", Value: map[string]interface{}{"check": "false"}, Help: "默认组件"},

		"componet_group": &ctx.Config{Name: "component_group", Value: "index", Help: "默认组件"},
		"componet": &ctx.Config{Name: "componet", Value: map[string]interface{}{
			"index": []interface{}{
				map[string]interface{}{"componet_name": "wiki", "componet_tmpl": "head", "metas": []interface{}{
					map[string]interface{}{"name": "viewport", "content": "width=device-width, initial-scale=0.7, user-scalable=no"},
				}, "favicon": "favicon.ico", "styles": []interface{}{"example.css", "wiki.css"}},
				map[string]interface{}{"componet_name": "header", "componet_tmpl": "fieldset",
					"componet_view": "Header", "componet_init": "initHeader",
					"title": "shylinux 天行健，君子以自强不息",
				},

				map[string]interface{}{"componet_name": "tree", "componet_tmpl": "fieldset",
					"componet_view": "Tree", "componet_init": "initTree",
					"componet_ctx": "web.wiki", "componet_cmd": "wiki_tree", "arguments": []interface{}{"@wiki_class"},
				},
				map[string]interface{}{"componet_name": "text", "componet_tmpl": "fieldset",
					"componet_view": "Text", "componet_init": "initText",
					"componet_ctx": "web.wiki", "componet_cmd": "wiki_text", "arguments": []interface{}{"@wiki_favor"},
				},

				map[string]interface{}{"componet_name": "footer", "componet_tmpl": "fieldset",
					"componet_view": "Footer", "componet_init": "initFooter",
					"title": "shycontext 地势坤，君子以厚德载物",
				},
				map[string]interface{}{"componet_name": "tail", "componet_tmpl": "tail",
					"scripts": []interface{}{"toolkit.js", "context.js", "example.js", "wiki.js"},
				},
			},
		}, Help: "组件列表"},

		"wiki_level": &ctx.Config{Name: "wiki_level", Value: "wiki/自然/编程", Help: "路由数量"},
		"wiki_favor": &ctx.Config{Name: "wiki_favor", Value: "index.md", Help: "路由数量"},
		"wiki_visit": &ctx.Config{Name: "wiki_visit", Value: map[string]interface{}{}, Help: "路由数量"},

		"wiki_dir":  &ctx.Config{Name: "wiki_dir", Value: "wiki", Help: "路由数量"},
		"wiki_list": &ctx.Config{Name: "wiki_list", Value: []interface{}{}, Help: "路由数量"},
		"wiki_list_show": &ctx.Config{Name: "wiki_list_show", Value: map[string]interface{}{
			"md": true,
		}, Help: "路由数量"},
	},
	Commands: map[string]*ctx.Command{
		"wiki_tree": &ctx.Command{Name: "wiki_tree", Help: "wiki_tree", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("nfs.dir", path.Join(m.Confx("wiki_level"), kit.Select(m.Option("wiki_class"), arg, 0)), "dir_sort", "time", "time_r")
			return
		}},
		"wiki_text": &ctx.Command{Name: "wiki_text", Help: "wiki_text", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			which := m.Cmdx("nfs.path", path.Join(m.Confx("wiki_level"), m.Option("wiki_class"), m.Confx("wiki_favor", arg, 0)))

			if ls, e := ioutil.ReadFile(which); e == nil {
				m.Confi("wiki_visit", []string{which, m.Option("remote_ip")},
					m.Confi("wiki_visit", []string{which, m.Option("remote_ip")})+1)
				m.Append("visit_count", m.Confi("wiki_visit", []string{which, m.Option("remote_ip")}))
				m.Append("visit_total", len(m.Confm("wiki_visit", []string{which})))

				buffer := bytes.NewBuffer([]byte{})
				temp, e := template.New("temp").Funcs(ctx.CGI).Parse(string(ls))
				if e != nil {
					m.Log("info", "parse %s %s", which, e)
				}
				temp.Execute(buffer, m)
				ls = buffer.Bytes()

				ls = markdown.ToHTML(ls, nil, nil)
				m.Echo(string(ls))
			} else {
				msg := m.Cmd("nfs.dir", path.Join(m.Confx("wiki_level"), m.Option("wiki_class")),
					"dir_deep", "dir_type", "dir", "time", "path")
				msg.Table(func(index int, value map[string]string) {
					msg.Meta["path"][index] = strings.TrimPrefix(value["path"], path.Join("usr", m.Confx("wiki_level")))
				})
				m.Echo(msg.ToHTML("wiki_list"))
			}
			return
		}},
		"wiki_list": &ctx.Command{Name: "wiki_list sort_field sort_order", Help: "wiki_list", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			sort_field, sort_order := "time", "time_r"
			if len(arg) > 0 {
				sort_field, arg = arg[0], arg[1:]
			}
			if len(arg) > 0 {
				sort_order, arg = arg[0], arg[1:]
			}

			dir := path.Join(m.Conf("wiki_dir"), m.Option("wiki_class"))
			md, e := ioutil.ReadDir(dir)
			m.Assert(e)

			for _, v := range md {
				if strings.HasSuffix(v.Name(), ".md") {
					f, e := os.Open(path.Join(dir, v.Name()))
					m.Assert(e)
					defer f.Close()

					title := ""
					nline, ncode := 0, 0
					h2, h3, h4 := 0, 0, 0
					for bio := bufio.NewScanner(f); bio.Scan(); {
						line := bio.Text()
						nline++

						if strings.HasPrefix(line, "## ") {
							h2++
							if title == "" {
								title = line[3:]
							}
						} else if strings.HasPrefix(line, "### ") {
							h3++
						} else if strings.HasPrefix(line, "#### ") {
							h4++
						} else if strings.HasPrefix(line, "```") {
							ncode++
						}
					}

					m.Add("append", "time", v.ModTime().Format("2006/01/02"))
					m.Add("append", "file", v.Name())
					m.Add("append", "size", v.Size())

					m.Add("append", "line", nline)
					m.Add("append", "code", ncode/2)
					m.Add("append", "h4", h4)
					m.Add("append", "h3", h3)
					m.Add("append", "h2", h2)
					m.Add("append", "title", title)
				}
			}
			m.Sort(sort_field, sort_order).Table()

			m.Target().Configs["wiki_list"].Value = []interface{}{}
			m.Table(func(maps map[string]string, list []string, line int) bool {
				if line > 0 {
					m.Confv("wiki_list", -2, maps)
				}
				return true
			})
			return
		}},

		"/wiki_tags": &ctx.Command{Name: "/wiki_tags ", Help: "博客", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
			return
		}},
		"/wiki_body": &ctx.Command{Name: "/wiki_body", Help: "维基", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
				return e
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
			return
		}},
		"/wiki_list": &ctx.Command{Name: "/wiki_list", Help: "维基", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
				m.Option("time_format", "2006-01-02 15:04:05")
				m.Sort("time", "time_r")
			}
			return
		}},
		"/wiki/": &ctx.Command{Name: "/wiki", Help: "维基", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Option("which", strings.TrimPrefix(key, "/wiki/"))
			if f, e := os.Stat(path.Join(m.Conf("wiki_dir"), m.Option("which"))); e == nil && !f.IsDir() && (strings.HasSuffix(m.Option("which"), ".json") || strings.HasSuffix(m.Option("which"), ".js") || strings.HasSuffix(m.Option("which"), ".css")) {
				m.Append("directory", path.Join(m.Conf("wiki_dir"), m.Option("which")))
				return e
			}

			m.Append("template", "wiki")
			return
		}},
		"/wx/": &ctx.Command{Name: "/wx/", Help: "微信", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
				msg.Option("time_format", "01/02 15:03")
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
			return
		}},
	},
}

func init() {
	wiki := &web.WEB{}
	wiki.Context = Index
	web.Index.Register(Index, wiki)
}
