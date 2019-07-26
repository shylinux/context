package web

import (
// "github.com/PuerkitoBio/goquery"
)

func parseHTML() {
	/*
		page, e := goquery.NewDocumentFromReader(res.Body)
		m.Assert(e)

		page.Find(kit.Select("html", m.Option("parse_chain"))).Each(func(n int, s *goquery.Selection) {
			if m.Options("parse_select") {
				for i := 0; i < len(m.Meta["parse_select"])-2; i += 3 {
					item := s.Find(m.Meta["parse_select"][i+1])
					if m.Meta["parse_select"][i+1] == "" {
						item = s
					}
					if v, ok := item.Attr(m.Meta["parse_select"][i+2]); ok {
						m.Add("append", m.Meta["parse_select"][i], v)
						m.Log("info", "item attr %v", v)
					} else {
						m.Add("append", m.Meta["parse_select"][i], strings.Replace(item.Text(), "\n", "", -1))
						m.Log("info", "item text %v", item.Text())
					}
				}
				return
			}

			s.Find("a").Each(func(n int, s *goquery.Selection) {
				if attr, ok := s.Attr("href"); ok {
					s.SetAttr("href", proxy(m, attr))
				}
			})
			s.Find("img").Each(func(n int, s *goquery.Selection) {
				if attr, ok := s.Attr("src"); ok {
					s.SetAttr("src", proxy(m, attr))
				}
				if attr, ok := s.Attr("r-lazyload"); ok {
					s.SetAttr("src", proxy(m, attr))
				}
			})
			s.Find("script").Each(func(n int, s *goquery.Selection) {
				if attr, ok := s.Attr("src"); ok {
					s.SetAttr("src", proxy(m, attr))
				}
			})

			if html, e := s.Html(); e == nil {
				m.Add("append", "html", html)
			}
		})
		m.Table()

	*/
}
