package wiki

import (
	"os/exec"
	"time"

	"bufio"
	"bytes"
	"contexts/ctx"
	"contexts/web"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gomarkdown/markdown"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
		"old_get": &ctx.Command{
			Name: "get [method GET|POST] [file name filename] url arg...",
			Help: "访问服务, method: 请求方法, file: 发送文件, url: 请求地址, arg: 请求参数",
			Form: map[string]int{"method": 1, "content_type": 1, "headers": 2, "file": 2, "body_type": 1, "body": 1, "fields": 1, "value": 1, "json_route": 1, "json_key": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if web, ok := m.Target().Server.(*web.WEB); m.Assert(ok) {
					if web.Client == nil {
						web.Client = &http.Client{}
					}

					if m.Has("value") {
						args := strings.Split(m.Option("value"), " ")
						values := []interface{}{}
						for _, v := range args {
							if len(v) > 1 && v[0] == '$' {
								values = append(values, m.Cap(v[1:]))
							} else {
								values = append(values, v)
							}
						}
						arg[0] = fmt.Sprintf(arg[0], values...)
					}

					method := m.Confx("method")
					uri := web.Merge(m, arg[0], arg[1:]...)
					m.Log("info", "%s %s", method, uri)
					m.Echo("%s: %s\n", method, uri)

					var body io.Reader
					index := strings.Index(uri, "?")
					content_type := ""

					switch method {
					case "POST":
						if m.Options("file") {
							file, e := os.Open(m.Meta["file"][1])
							m.Assert(e)
							defer file.Close()

							if m.Option("body_type") == "json" {
								content_type = "application/json"
								body = file
								break
							}
							buf := &bytes.Buffer{}
							writer := multipart.NewWriter(buf)

							part, e := writer.CreateFormFile(m.Option("file"), filepath.Base(m.Meta["file"][1]))
							m.Assert(e)
							io.Copy(part, file)

							for i := 0; i < len(arg)-1; i += 2 {
								value := arg[i+1]
								if len(arg[i+1]) > 1 {
									switch arg[i+1][0] {
									case '$':
										value = m.Cap(arg[i+1][1:])
									case '@':
										value = m.Conf(arg[i+1][1:])
									}
								}
								writer.WriteField(arg[i], value)
							}

							content_type = writer.FormDataContentType()
							body = buf
							writer.Close()
						} else if m.Option("body_type") == "json" {
							if m.Options("body") {
								data := []interface{}{}
								for _, v := range arg[1:] {
									if len(v) > 1 && v[0] == '$' {
										v = m.Cap(v[1:])
									}
									data = append(data, v)
								}
								body = strings.NewReader(fmt.Sprintf(m.Option("body"), data...))
							} else {
								data := map[string]interface{}{}
								for i := 1; i < len(arg)-1; i += 2 {
									switch arg[i+1] {
									case "false":
										data[arg[i]] = false
									case "true":
										data[arg[i]] = true
									default:
										if len(arg[i+1]) > 1 && arg[i+1][0] == '$' {
											data[arg[i]] = m.Cap(arg[i+1][1:])
										} else {
											data[arg[i]] = arg[i+1]
										}
									}
								}

								b, e := json.Marshal(data)
								m.Assert(e)
								body = bytes.NewReader(b)
							}

							content_type = "application/json"
							if index > -1 {
								uri = uri[:index]
							}

						} else if index > 0 {
							content_type = "application/x-www-form-urlencoded"
							body = strings.NewReader(uri[index+1:])
							uri = uri[:index]
						}
					}

					req, e := http.NewRequest(method, uri, body)
					m.Assert(e)
					for i := 0; i < len(m.Meta["headers"]); i += 2 {
						req.Header.Set(m.Meta["headers"][i], m.Meta["headers"][i+1])
					}

					if len(content_type) > 0 {
						req.Header.Set("Content-Type", content_type)
						m.Log("info", "content-type: %s", content_type)
					}

					for _, v := range m.Confv("cookie").(map[string]interface{}) {
						req.AddCookie(v.(*http.Cookie))
					}

					res, e := web.Client.Do(req)
					m.Assert(e)

					for _, v := range res.Cookies() {
						m.Confv("cookie", v.Name, v)
						m.Log("info", "set-cookie %s: %v", v.Name, v.Value)
					}

					if m.Confs("logheaders") {
						for k, v := range res.Header {
							m.Log("info", "%s: %v", k, v)
						}
					}

					if m.Confs("output") {
						if _, e := os.Stat(m.Conf("output")); e == nil {
							name := path.Join(m.Conf("output"), fmt.Sprintf("%d", time.Now().Unix()))
							f, e := os.Create(name)
							m.Assert(e)
							io.Copy(f, res.Body)
							if m.Confs("editor") {
								cmd := exec.Command(m.Conf("editor"), name)
								cmd.Stdin = os.Stdin
								cmd.Stdout = os.Stdout
								cmd.Stderr = os.Stderr
								cmd.Run()
							} else {
								m.Echo("write to %s\n", name)
							}
							return
						}
					}

					buf, e := ioutil.ReadAll(res.Body)
					m.Assert(e)

					ct := res.Header.Get("Content-Type")
					if len(ct) >= 16 && ct[:16] == "application/json" {
						var result interface{}
						json.Unmarshal(buf, &result)
						m.Option("response_json", result)
						if m.Has("json_route") {
							routes := strings.Split(m.Option("json_route"), ".")
							for _, k := range routes {
								if len(k) > 0 && k[0] == '$' {
									k = m.Cap(k[1:])
								}
								switch r := result.(type) {
								case map[string]interface{}:
									result = r[k]
								}
							}
						}

						fields := map[string]bool{}
						for _, k := range strings.Split(m.Option("fields"), " ") {
							if k == "" {
								continue
							}
							if fields[k] = true; len(fields) == 1 {
								m.Meta["append"] = append(m.Meta["append"], "index")
							}
							m.Meta["append"] = append(m.Meta["append"], k)
						}

						if len(fields) > 0 {

							switch ret := result.(type) {
							case map[string]interface{}:
								m.Append("index", "0")
								for k, v := range ret {
									switch value := v.(type) {
									case string:
										m.Append(k, strings.Replace(value, "\n", " ", -1))
									case float64:
										m.Append(k, fmt.Sprintf("%d", int(value)))
									default:
										if _, ok := fields[k]; ok {
											m.Append(k, fmt.Sprintf("%v", value))
										}
									}
								}
							case []interface{}:
								for i, r := range ret {
									m.Add("append", "index", i)
									if rr, ok := r.(map[string]interface{}); ok {
										for k, v := range rr {
											switch value := v.(type) {
											case string:
												if _, ok := fields[k]; len(fields) == 0 || ok {
													m.Add("append", k, strings.Replace(value, "\n", " ", -1))
												}
											case float64:
												if _, ok := fields[k]; len(fields) == 0 || ok {
													m.Add("append", k, fmt.Sprintf("%d", int64(value)))
												}
											case bool:
												if _, ok := fields[k]; len(fields) == 0 || ok {
													m.Add("append", k, fmt.Sprintf("%v", value))
												}
											case map[string]interface{}:
												for kk, vv := range value {
													key := k + "." + kk
													if _, ok := fields[key]; len(fields) == 0 || ok {
														m.Add("append", key, strings.Replace(fmt.Sprintf("%v", vv), "\n", " ", -1))
													}
												}
											default:
												if _, ok := fields[k]; ok {
													m.Add("append", k, fmt.Sprintf("%v", value))
												}
											}
										}
									}
								}

								if m.Has("json_key") {
									m.Sort(m.Option("json_key"))
								}
								m.Meta["index"] = nil
								for i, _ := range ret {
									m.Add("append", "index", i)
								}
							}
						}
					}

					if m.Table(); len(m.Meta["append"]) == 0 {
						m.Echo("%s", string(buf))
					}
				}
			}},
	},
}

func init() {
	wiki := &WIKI{}
	wiki.Context = Index
	web.Index.Register(Index, wiki)
}
