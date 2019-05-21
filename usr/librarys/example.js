function Page(page) {
    page.__proto__ = {
        Sync: function(m) {
            var meta = m
            var data = ""
            var list = []
            return {
                change: function(cb) {
                    list.push(cb)
                    return list.length-1
                },
                eq: function(value) {
                    return data == value
                },
                neq: function(value) {
                    return data != value
                },
                get: function() {
                    return data
                },
                set: function(value) {
                    if (value == data) {
                        return value
                    }
                    old_value = data, data = value
                    meta && kit.Log(meta, value, old_value)
                    for (var i = 0; i < list.length; i++) {
                        list[i](value, old_value)
                    }
                    return value
                },
            }
        },
        View: function(type, line, key, cb) {
            switch (type) {
                case "icon":
                    return [{view: ["item", "div"], list: [{type: "img", data: {src: line[key[0]]}}, {}]}]

                case "text":
                    switch (key.length) {
                        case 0:
                            return [{view: ["item", "div", "null"], click: cb}]
                        case 1:
                            return [{view: ["item", "div", line[key[0]]], click: cb}]
                        default:
                            return [{view: ["item", "div", line[key[0]]+"("+line[key[1]]+")"], click: cb}]
                    }
                    break

                case "table":
                    var data = JSON.parse(line.text)
                    var list = []
                    var line = []
                    for (var k in data[0]) {
                        line.push({view: ["", "th", k]})
                    }
                    list.push({view: ["", "tr"], list: line})
                    for (var i = 0; i < data.length; i++) {
                        var line = []
                        for (var k in data[i]) {
                            line.push({view: ["", "td", data[i][k]]})
                        }
                        list.push({view: ["", "tr"], list: line})
                    }
                    var result = [{view: [""], list: [{view: ["", "table"], list: list}]}]
                    return result

                case "field":
                    var data = JSON.parse(line.text)
                    var input = [{type: "input", style: {"display": "none"}}]
                    for (var i = 0; i < data.input.length; i++) {
                        input.push(data.input[i])
                    }

                    var result = [{view: ["", "fieldset"], list: [
                        {name: "form", view: ["", "form"], dataset: {
                            componet_group: data.componet_group,
                            componet_name: data.componet_name,
                            cmds: data.cmds,
                        }, list: input},
                        {name: "table", view: ["", "table"]},
                        {view: ["", "code"], list: [{name: "code", view: ["", "pre"]}]},
                    ]}]
                    return result

            }
        },
        reload: function() {
            location.reload()
        },
        onscroll: function(event, target, action) {
            switch (action) {
                case "scroll":
                    if (event.target == document.body) {
                        kit.ScrollPage(event, page.conf)
                    }
                    break
            }
        },
        showToast: function(text) {

        },

        initHeader: function(page, field, option, output) {
            return [{"text": ["shycontext", "div", "title"]}]
        },
        initBanner: function(page, field, option, output) {
            field.querySelectorAll("li").forEach(function(item) {
                item.onclick = function(event) {
                    ctx.Search("componet_group", item.innerText)
                    if (item.innerText == "login") {
                        ctx.Cookie("sessid", "")
                    }
                }
            })
            return [{"text": ["shylinux", "div", "title"]}]
        },
        initFooter: function(page, field, option) {
            return [{"view": ["title", "div", "<a href='mailto:shylinux@163.com'>shylinux@163.com</>"]}]
        },
        initField: function(page, cb) {
            document.querySelectorAll("body>fieldset").forEach(function(pane) {
                var form = pane.querySelector("form.option")
                page[form.dataset.componet_name] = pane

                // pane init
                pane.which = page.Sync(form.dataset.componet_name)
                pane.ShowWindow = function(width, height) {
                    kit.ModifyView(pane, {window: [width||80, height||40]})
                }
                pane.ShowDialog = function(width, height) {
                    if (pane.style.display != "block") {
                        pane.style.display = "block"
                        kit.ModifyView(pane, {dialog: [width||800, height||400]})
                        return true
                    }
                    pane.style.display = "none"
                    return false
                }
                pane.Size = function(width, height) {
                    pane.style.display = width==0? "none": "block"
                    pane.style.width = width+"px"
                    pane.style.height = height+"px"
                }

                // form init
                form.Run = function(cmds, cb) {
                    ctx.Run(page, form.dataset, cmds, cb)
                }
                form.Runs = function(cmds, cb) {
                    ctx.Run(page, form.dataset, cmds, function(msg) {
                        ctx.Table(msg, function(line, index) {
                            cb(line, index, msg)
                        })
                    })
                }

                var conf = cb(page[pane.dataset.init], pane, form)
            })

            document.querySelectorAll("body>fieldset").forEach(function(pane) {
                for (var k in pane.Listen) {
                    page[k].which.change(pane.Listen[k])
                }
            })
        },
    }
    window.onload = function() {
        page.init(page)

        window.onresize = page.size
        document.body.onkeydown = function(event) {
            page.onscroll && page.onscroll(event, document.body, "scroll")
        }
        document.body.onkeyup = function(event) {
            page.oncontrol && page.oncontrol(event, document.body, "control")
        }
    }
    return page
}
