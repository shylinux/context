var page = Page({
    conf: {
        border: 4,
        scroll_x: 50,
        scroll_y: 50,
    },
    onlayout: function(event, sizes) {
        return
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
        page.tree.Size(width, sizes.tree)

        sizes.text = height - sizes.tree - sizes.header - sizes.footer
        page.text.Size(width, sizes.text)
    },
    initList: function(page, pane, form, output) {
        ctx.Runs(page, form, function(msg) {
            output.innerHTML = ""
            kit.AppendChild(output, [{"tree": ctx.Table(msg, function(value, index) {
                return {"leaf": [value.file, function(event, target) {
                    ctx.Search("wiki_favor", value.file)
                }]}
            })}])
        })
        return
    },

    initTree: function(page, pane, form, output) {
        // if (!ctx.isMobile) {
        //     pane.style.float = "left"
        // }

        ctx.Runs(page, form, function(msg) {
            output.innerHTML = ""
            var back = [{"button": ["知识", function(event) {
                ctx.Search({"wiki_level": "", "wiki_class": "", "wiki_favor": ""})
            }]}]
            ctx.Search("wiki_class").split("/").forEach(function(value, index, array) {
                if (value) {
                    var favor = []
                    for (var i = 0; i <= index; i++) {
                        favor.push(array[i])
                    }
                    favor.push("")
                    back.push({"button": [value, function(event) {
                        ctx.Search({"wiki_class": favor.join("/"), "wiki_favor":""})
                    }]})
                }
            })

            var ui = kit.AppendChild(output, [
                {"view": ["back"], "list": back},
                {"view": ["tree"], "list": [{"tree": ctx.Table(msg, function(value, index) {
                    if (value.filename.endsWith("/")) {
                        return {"leaf": [value.filename, function(event, target) {
                            ctx.Search({"wiki_class": ctx.Search("wiki_class")+value.filename, "wiki_favor": ""})
                        }]}
                    }
                })}]},
                {"view": ["list"], "list": [{"tree": ctx.Table(msg, function(value, index) {
                    if (!value.filename.endsWith("/")) {
                        return {"leaf": [value.time.substr(5, 5)+" "+value.filename, function(event, target) {
                            ctx.Search("wiki_favor", value.filename)
                        }]}
                    }
                })}]},
            ])
        })
        return
    },
    initText: function(page, pane, form, output) {
        ctx.Runs(page, form, function(msg) {
            if (!msg.result) {
                return
            }
            output.innerHTML = ""
            var ui = kit.AppendChild(output, [
                {"view": ["menu", "div", "", "menu"]},
                {"view": ["text", "div", msg.result.join(""), "text"]},
            ])

            ui.text.querySelectorAll("table").forEach(function(value, index, array) {
                kit.OrderTable(value)
            })
            ui.text.querySelectorAll("table.wiki_list").forEach(function(value, index, array) {
                kit.OrderTable(value, "path", function(event) {
                    var text = event.target.innerText
                    ctx.Search({"wiki_class": text})
                })
            })

            ui.text.querySelectorAll("a").forEach(function(value, index, array) {
                kit.OrderLink(value, pane)
            })


            var i = 0, j = 0, k = 0
            var h0 = [], h2 = [], h3 = []
            ui.text.querySelectorAll("h2,h3,h4").forEach(function(value, index, array) {
                var id = ""
                var level = 0
                var text = value.innerText
                var ratio = parseInt(value.offsetTop/pane.scrollHeight*100)

                if (value.tagName == "H2") {
                    j=k=0
                    h2 = []
                    id = ++i+"."
                    text = id+" "+text
                    h0.push({"fork": [text+" ("+ratio+"%)", h2, function(event) {
                        console.log(text)
                        location.hash = id
                    }]})
                } else if (value.tagName == "H3") {
                    k=0
                    h3 = []
                    id = i+"."+(++j)
                    text = id+" "+text
                    h2.push({"fork": [text+" ("+ratio+"%)", h3, function(event) {
                        console.log(text)
                        location.hash = id
                    }]})
                } else if (value.tagName == "H4") {
                    id = i+"."+j+"."+(++k)
                    text = id+" "+text
                    h3.push({"leaf": [text+" ("+ratio+"%)", function(event) {
                        console.log(text)
                        location.hash = id
                    }]})
                }
                value.innerText = text
                value.id = id
            })
            kit.AppendChild(ui.menu, [{"tree": h0}])


            ui.text.style.width = document.body.offsetWidth-30+"px"
            if (i > 0 && !ctx.isMobile) {
                ui.menu.style.position = "absolute"
                var width = ui.menu.offsetWidth
                var height = ui.menu.offsetHeight>400?ui.menu.offsetHeight:600

                pane.style.marginLeft = width+10+"px"
                ui.menu.style.marginLeft = -width-20+"px"
                ui.text.style.height = height+"px"
                ui.text.style.width = pane.offsetWidth-30+"px"
            }
            if (location.hash) {
                location.href = location.hash
            }
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
    },
})
