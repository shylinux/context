var page = Page({check: false, conf: {border: 4},
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

    Action: {
        title: function(event, item, value, page) {
            ctx.Search("layout", ctx.Search("layout")? "": "max")
        },
        user: function(event, item, value, page) {
            page.carte.Pane.Show(event, shy({
                "修改昵称": function(event) {
                    var name = kit.prompt("new name")
                    name && page.login.Pane.Run(event, ["rename", name], function(msg) {
                        page.header.Pane.State("user", name)
                    })
                },
                "退出登录": function(event) {
                    kit.confirm("logout?") && page.login.Pane.Exit()
                },
            }, ["修改昵称", "退出登录"], function(event, value, meta) {
                meta[value](event)
            }))
        },
        menu: function() {
            page.text.Pane.Menu()
        },
        tree: function() {
            page.tree.Pane.Tree()
        },

    },
    Button: shy({title: "github.com/shylinux/context", tree: "tree", menu: "menu"}, ["tree", "menu"], function(key, value) {var meta = arguments.callee.meta
        return kit.isNone(key)? meta: kit.isNone(value)? meta[key]: (meta[key] = value, page.header.Pane.Show())
    }),
    Status: shy({title: '<a href="mailto:shylinux@163.com">shylinux@163.com</a>', text: "0", menu: "0"}, ["text", "menu"], function(key, value) {var meta = arguments.callee.meta
        return kit.isNone(key)? meta: kit.isNone(value)? meta[key]: (meta[key] = value, page.footer.Pane.Show())
    }),

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

                ctx.Run(event, form.dataset, [], function(msg) {
                    ui.back.innerHTML = "", kit.AppendChild(ui.back, [
                        {"button": ["知识", function(event) {
                            ctx.Search({"level": "", "class": "", "favor": ""})
                        }]},
                    ].concat(ctx.Search("class").split("/").map(function(value, index, array) {
                        return value && {"button": [value, function(event) {
                            location.hash = "", ctx.Search({"class": array.slice(0, index+1).join("/")+"/", "favor":""})
                        }]}
                    })))

                    ui.tree.innerHTML = "", kit.AppendChild(ui.tree, msg.Table(function(value, index) {
                        return value.file.endsWith("/") && {"text": [value.file, "div"], click: function(event, target) {
                            location.hash = "", ctx.Search({"class": ctx.Search("class")+value.file, "favor": ""})
                        }}
                    }))
                    ui.list.innerHTML = "", kit.AppendChild(ui.list, msg.Table(function(value, index) {
                        return !value.file.endsWith("/") && {"text": [value.time.substr(5, 5)+" "+value.file, "div"], click: function(event, target) {
                            location.hash = "", ctx.Search("favor", value.file)
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
            page.Status("text", kit.Position(ui.text))
        }
        ui.menu.onscroll = function(event) {
            page.Status("menu", kit.Position(ui.menu))
        }
        return {
            Menu: function(value) {
                page.Conf("menu.display", value || (page.Conf("menu.display")? "": "none"))
            },
            Size: function(width, height) {
                if (kit.device.isMobile) {
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
                page.Conf("menu.float", !kit.device.isMobile, function(value, old) {
                    page.onlayout()
                })
                page.Conf("menu.scroll", true, function(value, old) {
                    page.onlayout()
                })

                ctx.Search("layout") == "max" && (page.Conf("tree.display", "none"), page.Conf("menu.display", "none"))

                ctx.Run(event, form.dataset, [], function(msg) {
                    ui.menu.innerHTML = "", ui.text.innerHTML = msg.result? msg.result.join(""): ""
                    kit.AppendChild(ui.menu, [{"tree": kit.OrderText(field, ui.text)}])
                    page.Status("count", msg.visit_count)
                    page.Status("visit", msg.visit_total)
                    page.onlayout()
                    return
                })
            },
        }
    },
    init: function(page) {
        page.tree.Pane.Show()
        page.text.Pane.Show()
        page.Button("tree", "tree")
    },
})
