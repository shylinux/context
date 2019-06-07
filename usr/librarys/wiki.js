var page = Page({
    conf: {
        border: 4,
        scroll_x: 50,
        scroll_y: 50,
    },
    onlayout: function(event, sizes) {
        var height = document.body.clientHeight-page.conf.border
        var width = document.body.clientWidth-page.conf.border
        page.conf.height = height
        page.conf.width = width

        sizes = sizes || {}
        sizes.header == undefined && (sizes.header = page.header.clientHeight)
        sizes.footer == undefined && (sizes.footer = page.footer.clientHeight)
        page.header.Size(width, sizes.header)
        page.footer.Size(width, sizes.footer)

        sizes.tree == undefined && (sizes.tree = page.tree.clientHeight)
        // page.tree.Size(width, sizes.tree)

        sizes.text = height - sizes.tree - sizes.header - sizes.footer - page.conf.border * 3
        page.text.Size(width, sizes.text)
    },
    oncontrol: function(event) {
        if (event.ctrlKey) {
            switch (event.key) {
                case "t":
                    page.tree.Tree()
                    break
                case "m":
                    page.text.Menu()
                    break
                case "n":
                    page.tree.Tree("none")
                    page.text.Menu("none")
                    break
            }
        }
    },

    initTree: function(page, pane, form, output) {
        var ui = kit.AppendChild(output, [
            {"view": ["back"], "name": "back"}, {"view": ["gap"]},
            {"view": ["tree"], "name": "tree"}, {"view": ["gap"]},
            {"view": ["list"], "name": "list"}, {"view": ["gap"]},
        ])

        pane.Conf("tree.display", "", function(value, old) {
            pane.style.display = value
            page.onlayout()
        })
        pane.Tree = function(value) {
            pane.Conf("tree.display", value || (pane.Conf("tree.display")? "": "none"))
        }

        ctx.Runs(page, form, function(msg) {
            ui.back.innerHTML = "", kit.AppendChild(ui.back, [
                {"button": ["知识", function(event) {
                    ctx.Search({"wiki_level": "", "wiki_class": "", "wiki_favor": ""})
                }]},
            ].concat(ctx.Search("wiki_class").split("/").map(function(value, index, array) {
                return value && {"button": [value, function(event) {
                    location.hash = "", ctx.Search({"wiki_class": array.slice(0, index+1).join("/")+"/", "wiki_favor":""})
                }]}
            })))

            ui.tree.innerHTML = "", kit.AppendChild(ui.tree, ctx.Table(msg, function(value, index) {
                return value.filename.endsWith("/") && {"text": [value.filename, "div"], click: function(event, target) {
                    location.hash = "", ctx.Search({"wiki_class": ctx.Search("wiki_class")+value.filename, "wiki_favor": ""})
                }}
            }))
            ui.list.innerHTML = "", kit.AppendChild(ui.list, ctx.Table(msg, function(value, index) {
                return !value.filename.endsWith("/") && {"text": [value.time.substr(5, 5)+" "+value.filename, "div"], click: function(event, target) {
                    location.hash = "", ctx.Search("wiki_favor", value.filename)
                }}
            }))
        })
        return
    },
    initText: function(page, pane, form, output) {
        var ui = kit.AppendChild(output, [
            {"view": ["menu", "div", "", "menu"]},
            {"view": ["text", "div", "", "text"]},
        ])
        ui.text.onscroll = function(event) {
            page.footer.State("text", kit.Position(ui.text))
        }
        ui.menu.onscroll = function(event) {
            page.footer.State("menu", kit.Position(ui.menu))
        }

        pane.Conf("menu.display", "", function(value, old) {
            ui.menu.style.display = value
        })
        pane.Conf("menu.float", !kit.isMobile, function(value, old) {
            page.onlayout()
        })
        pane.Conf("menu.scroll", true, function(value, old) {
            page.onlayout()
        })
        pane.Menu = function(value) {
            pane.Conf("menu.display", value || (pane.Conf("menu.display")? "": "none"))
        }
        pane.Size = function(width, height) {
            if (pane.Conf("menu.float")) {
                ui.menu.className = "menu left"
            } else {
                ui.menu.className = "menu"
            }
            if (pane.Conf("menu.float") && pane.Conf("menu.scroll")) {
                ui.menu.style.height = (height-8)+"px"
                ui.text.style.height = ((ui.menu.clientHeight||height)-8-20)+"px"
            } else {
                ui.menu.style.height = " "
                ui.text.style.height = " "
            }

            if (location.hash) {
                location.href = location.hash
            }
            ui.text.onscroll()
            ui.menu.onscroll()
        }

        ctx.Runs(page, form, function(msg) {
            ui.menu.innerHTML = "", ui.text.innerHTML = msg.result? msg.result.join(""): ""
            kit.AppendChild(ui.menu, [{"tree": kit.OrderText(pane, ui.text)}])
            page.footer.State("count", msg.visit_count)
            page.footer.State("visit", msg.visit_total)
            page.onlayout()
            return
        })
        return
    },
    init: function(page) {
        page.initField(page, function(init, pane, form) {
            var output = pane.querySelector("div.output")

            if (typeof init == "function") {
                var conf = init(page, pane, form, output)
                if (conf) {
                    kit.AppendChild(output, conf)
                }
            }
        })

        page.footer.Order({"text": "", "menu": "", "count": "0", "visit": "0"}, ["visit", "count", "menu", "text"])
        page.header.Order({"tree": "tree", "menu": "menu"}, ["tree", "menu"], function(event, item, value) {
            switch (item) {
                case "menu":
                    page.text.Menu()
                    break
                case "tree":
                    page.tree.Tree()
                    break
                default:
                    page.confirm("logout?") && page.login.Exit()
            }
        })
    },
})
