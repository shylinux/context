var page = Page({
    conf: {border: 4},
    onlayout: function(event, sizes) {
        var height = document.body.clientHeight-page.conf.border
        var width = document.body.clientWidth-page.conf.border
        page.conf.height = height
        page.conf.width = width

        sizes = sizes || {}
        sizes.header == undefined && (sizes.header = page.header.clientHeight)
        sizes.footer == undefined && (sizes.footer = page.footer.clientHeight)
        page.header.Pane.Size(width, sizes.header)
        page.footer.Pane.Size(width, sizes.footer)

        sizes.tree == undefined && (sizes.tree = page.tree.clientHeight)
        // page.tree.Size(width, sizes.tree)

        sizes.text = height - sizes.tree - sizes.header - sizes.footer - page.conf.border * 3
        page.text.Pane.Size(width, sizes.text)
    },
    oncontrol: function(event) {
        if (event.ctrlKey) {
            switch (event.key) {
                case "t":
                    page.tree.Pane.Tree()
                    break
                case "m":
                    page.text.Pane.Menu()
                    break
                case "n":
                    page.tree.Pane.Tree("none")
                    page.text.Pane.Menu("none")
                    break
            }
        }
    },

    initTree: function(page, field, form, output) {
        var ui = kit.AppendChild(output, [
            {"view": ["back"], "name": "back"}, {"view": ["gap"]},
            {"view": ["tree"], "name": "tree"}, {"view": ["gap"]},
            {"view": ["list"], "name": "list"}, {"view": ["gap"]},
        ])

        return {
            Tree: function(value) {
                page.Conf("tree.display", value || (page.Conf("tree.display")? "": "none"))
            },
            Show: function() {
                page.Conf("tree.display", "", function(value, old) {
                    field.style.display = value
                    page.onlayout()
                })

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
                        return value.file.endsWith("/") && {"text": [value.file, "div"], click: function(event, target) {
                            location.hash = "", ctx.Search({"wiki_class": ctx.Search("wiki_class")+value.file, "wiki_favor": ""})
                        }}
                    }))
                    ui.list.innerHTML = "", kit.AppendChild(ui.list, ctx.Table(msg, function(value, index) {
                        return !value.file.endsWith("/") && {"text": [value.time.substr(5, 5)+" "+value.file, "div"], click: function(event, target) {
                            location.hash = "", ctx.Search("wiki_favor", value.file)
                        }}
                    }))
                })
            },
        }
    },
    initText: function(page, field, form, output) {
        var ui = kit.AppendChild(output, [
            {"view": ["menu", "div", "", "menu"]},
            {"view": ["text", "div", "", "text"]},
        ])
        ui.text.onscroll = function(event) {
            page.footer.Pane.State("text", kit.Position(ui.text))
        }
        ui.menu.onscroll = function(event) {
            page.footer.Pane.State("menu", kit.Position(ui.menu))
        }
        return {
            Menu: function(value) {
                page.Conf("menu.display", value || (page.Conf("menu.display")? "": "none"))
            },
            Size: function(width, height) {
                if (kit.isMobile) {
                    return
                }
                if (page.Conf("menu.float")) {
                    ui.menu.className = "menu left"
                } else {
                    ui.menu.className = "menu"
                }
                if (page.Conf("menu.float") && page.Conf("menu.scroll")) {
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
            },
            Show: function() {
                page.Conf("menu.display", "", function(value, old) {
                    ui.menu.style.display = value
                })
                page.Conf("menu.float", !kit.isMobile, function(value, old) {
                    page.onlayout()
                })
                page.Conf("menu.scroll", true, function(value, old) {
                    page.onlayout()
                })

                ctx.Search("layout") == "max" && (page.Conf("tree.display", "none"), page.Conf("menu.display", "none"))

                ctx.Runs(page, form, function(msg) {
                    ui.menu.innerHTML = "", ui.text.innerHTML = msg.result? msg.result.join(""): ""
                    kit.AppendChild(ui.menu, [{"tree": kit.OrderText(field, ui.text)}])
                    page.footer.Pane.State("count", msg.visit_count)
                    page.footer.Pane.State("visit", msg.visit_total)
                    page.onlayout()
                    return
                })
            },
        }
    },
    init: function(page) {
        page.footer.Pane.Order({"text": "", "menu": "", "count": "0", "visit": "0"}, ["visit", "count", "menu", "text"])
        page.header.Pane.Order({"tree": "tree", "menu": "menu"}, ["tree", "menu"], function(event, item, value) {
            switch (item) {
                case "menu":
                    page.text.Pane.Menu()
                    break
                case "tree":
                    page.tree.Pane.Tree()
                    break
                case "title":
                    ctx.Search("layout", ctx.Search("layout")? "": "max")
                    break

                default:
                    page.confirm("logout?") && page.login.Pane.Exit()
            }
        })
        page.header.style.height = "32px"
        page.footer.style.height = "32px"
        page.tree.Pane.Show()
        page.text.Pane.Show()
    },
})
