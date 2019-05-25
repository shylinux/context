function Page(page) {
    page.__proto__ = {
        _id: 1, ID: function() {
            return this._id++
        },
        Sync: function(m) {
            var meta = m, data = "", list = []
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
        View: function(parent, type, line, key, cb) {
            var ui = {}
            var result = []
            switch (type) {
                case "icon":
                    result = [{view: ["item", "div"], list: [{type: "img", data: {src: line[key[0]]}}, {}]}]
                    break

                case "text":
                    switch (key.length) {
                        case 0:
                            result = [{view: ["item", "div", "null"], click: cb}]
                            break
                        case 1:
                            result = [{view: ["item", "div", line[key[0]]], click: cb}]
                            break
                        default:
                            result = [{view: ["item", "div", line[key[0]]+"("+line[key[1]]+")"], click: cb}]
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
                    break

                case "field":
                    line = JSON.parse(line.text)

                case "plugin":
                    var id = "plugin"+page.ID()
                    var input = [{type: "input", style: {"display": "none"}}]
                    JSON.parse(line.inputs || "[]").forEach(function(item, index, inputs) {
                        function run(event) {
                            var args = []
                            ui.option.querySelectorAll("input").forEach(function(item, index){
                                if (index==0) {
                                    return
                                }
                                if (item.type == "text") {
                                    args.push(item.value)
                                }
                            })
                            ui.option.Run(event, args, function(msg) {
                                ui.option.ondaemon(msg)
                            })
                        }

                        item.type == "button"? item.onclick = function(event) {
                            var plugin = page[id];
                            (plugin[item.click] || function() {
                                run(event)
                            })(item, index, inputs, event, ui.option)

                        }: item.onkeyup = function(event) {
                            event.key == "Enter" && (index == inputs.length-1? run(event): event.target.nextSibling.focus())
                            if (event.ctrlKey) {
                                switch (event.key) {
                                    case "k":
                                        kit.DelText(target, target.selectionStart)
                                        break
                                    case "u":
                                        kit.DelText(target, 0, target.selectionEnd)
                                        break
                                    case "w":
                                        var start = target.selectionStart-2
                                        var end = target.selectionEnd-1
                                        for (var i = start; i >= 0; i--) {
                                            if (target.value[end] == " " && target.value[i] != " ") {
                                                break
                                            }
                                            if (target.value[end] != " " && target.value[i] == " ") {
                                                break
                                            }
                                        }
                                        kit.DelText(target, i+1, end-i)
                                        break

                                    case "c":
                                        ui.output.innerHTML = ""
                                        break
                                    case "r":
                                        ui.output.innerHTML = ""
                                    case "j":
                                        run(event)
                                        break
                                    case "l":
                                        page.action.scrollTo(0, ui.option.parentNode.offsetTop)
                                        break
                                    case "m":
                                        event.stopPropagation()
                                        var uis = page.View(parent, type, line, key, cb)
                                        page.action.scrollTo(0, uis.option.parentNode.offsetTop)
                                        ui.option.querySelectorAll("input")[1].focus()
                                        break
                                }
                            }
                        }
                        input.push({type: "input", data: item})
                    })

                    var result = [{view: [line.view, "fieldset"], data: {id: id}, list: [
                        {script: "Plugin("+id+","+line.init+")"},
                        {text: [line.name, "legend"]},
                        {name: "option", view: ["option", "form"], data: {Run: cb}, list: input},
                        {name: "output", view: ["output", "div"]},
                    ]}]
                    break
            }

            ui = kit.AppendChild(parent, result)
            return ui
        },
        reload: function() {
            location.reload()
        },
        showToast: function(text) {

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
                    pane.style.display = (width==0 || height==0)? "none": "block"
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

        window.onresize = function(event) {
            page.onlayout && page.onlayout(event)
        }
        document.body.onkeydown = function(event) {
            page.onscroll && page.onscroll(event, document.body, "scroll")
        }
        document.body.onkeyup = function(event) {
            page.oncontrol && page.oncontrol(event, document.body, "control")
        }
    }
    return page
}
function Plugin(field, plugin) {
    var option = field.querySelector("form.option")
    var output = field.querySelector("div.output")

    plugin.__proto__ = {
        field: field,
    }
    plugin.init(page, page.action, field, option, output)
    page[field.id] = plugin
}
