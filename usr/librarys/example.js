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
                set: function(value, force) {
                    if (value == data && !force) {
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
                    result = [{view: ["item", "div", key.length>1? line[key[0]]+"("+line[key[1]]+")": (key.length>0? line[key[0]]: "null")], click: cb}]
                    break

                case "code":
                    result = [{type: "code", list: [{text: [key.length>1? line[key[0]]+"("+line[key[1]]+")": (key.length>0? line[key[0]]: "null")], click: cb}]}]
                    break

                case "table":
                    result = [{view: [""], list: [
                        {view: ["", "table"], list: JSON.parse(line.text || "[]").map(function(item, index) {
                            return {type: "tr", list: item.map(function(value) {
                                return {text: [value, index == 0? "th": "td"]}
                            })}
                        })},
                    ]}]
                    break

                case "field":
                    line = JSON.parse(line.text)

                case "plugin":
                    var id = "plugin"+page.ID()
                    result = [{view: [line.view, "fieldset"], data: {id: id}, list: [
                        {text: [line.name, "legend"]},
                        {name: "option", view: ["option", "form"], data: {Run: cb}, list: [{type: "input", style: {"display": "none"}}]},
                        {name: "output", view: ["output", "div"]},
                        {script: "Plugin("+id+","+line.inputs+","+line.init+")"},
                    ]}]
                    break
            }

            ui = kit.AppendChild(parent, result)
            ui.last.Meta = line
            return ui
        },
        alert: function(text) {
            alert(text)
        },
        confirm: function(text) {
            return confirm(text)
        },
        reload: function() {
            location.reload()
        },
        oninput: function(event, local) {
            var target = event.target
            kit.History.add("key", (event.ctrlKey? "Control+": "")+(event.shiftKey? "Shift+": "")+event.key)

            if (event.ctrlKey) {
                if (typeof local == "function" && local(event)) {
                    event.stopPropagation()
                    return true
                }
                switch (event.key) {
                    case "a":
                    case "e":
                    case "f":
                    case "b":
                    case "h":
                    case "d":
                        break
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
                    default:
                        return false

                }
                event.stopPropagation()
                return true
            }
            switch (event.key) {
                case "Escape":
                    target.blur()
                    event.stopPropagation()
                    return true
                default:
                    if (kit.HitText(target, "jk")) {
                        kit.DelText(target, target.selectionStart-2, 2)
                        target.blur()
                        return true
                    }
            }
            return false
        },
        ontoast: function(text, title, duration) {
            var args = typeof text == "object"? text: {text: text, title: title, duration: duration}
            var toast = kit.ModifyView("fieldset.toast", {
                display: "block", dialog: [args.width||200, args.height||40], padding: 10,
            })

            var list = [{text: [args.text||""]}]
            args.inputs && args.inputs.forEach(function(input) {
                if (typeof input == "string") {
                    list.push({inner: input, type: "label", style: {"margin-right": "5px"}})
                    list.push({input: [input, page.oninput]})
                } else {
                    list.push({inner: input[0], type: "label", style: {"margin-right": "5px"}})
                    var option = []
                    for (var i = 1; i < input.length; i++) {
                        option.push({type: "option", inner: input[i]})
                    }
                    list.push({name: input[0], type: "select", list: option})
                }
                list.push({type: "br"})
            })
            args.button && args.button.forEach(function(input) {
                list.push({type: "button", inner: input, click: function(event) {
                    var values = {}
                    toast.querySelectorAll("input").forEach(function(input) {
                        values[input.name] = input.value
                    })
                    toast.querySelectorAll("select").forEach(function(input) {
                        values[input.name] = input.value
                    })
                    typeof args.cb == "function" && args.cb(input, values)
                    toast.style.display = "none"
                }})
            })

            kit.ModifyNode(toast.querySelector("legend"), args.title||"tips")
            var ui = kit.AppendChild(kit.ModifyNode(toast.querySelector("div.output"), ""), list)
            args.duration !=- 1 && setTimeout(function(){toast.style.display = "none"}, args.duration||3000)
            page.toast = toast
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

        initHeader: function(page, pane, form, output) {
            var state = {}, list = [], cb = function(event, item, value) {
            }
            pane.State = function(name, value) {
                state[name] = value, pane.Show()
            }
            pane.Order = function(value, cbs) {
                list = value, cb = cbs || cb, pane.Show()
            }
            pane.Show = function() {
                output.innerHTML = "", kit.AppendChild(output, [
                    {"view": ["title", "div", "shycontext"]},
                    {"view": ["state"], list: list.map(function(item) {return {text: [state[item], "div"], click: function(event) {
                        cb(event, item, state[item])
                    }}})},
                ])
            }
            return
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
        initFooter: function(page, pane, form, output) {
            var state = {}, list = []
            pane.State = function(name, value) {
                state[name] = value, pane.Show()
            }
            pane.Order = function(value) {
                list = value, pane.Show()
            }

            pane.Show = function() {
                output.innerHTML = "", kit.AppendChild(output, [
                    {"view": ["title", "div", "<a href='mailto:shylinux@163.com'>shylinux@163.com</>"]},
                    {"view": ["state"], list: list.map(function(item) {return {text: [item+":"+state[item], "div"]}})},
                ])
            }
            return
        },
        initField: function(page, cb) {
            document.querySelectorAll("body>fieldset").forEach(function(pane) {
                var form = pane.querySelector("form.option")
                page[form.dataset.componet_name] = pane

                // pane init
                pane.which = page.Sync(form.dataset.componet_name)
                pane.ShowDialog = function(width, height) {
                    if (pane.style.display != "block") {
                        page.dialog && page.dialog != pane && page.dialog.style.display == "block" && page.dialog.Show()
                        pane.style.display = "block", page.dialog = pane
                        kit.ModifyView(pane, {window: [width||80, height||200]})
                        return true
                    }
                    pane.style.display = "none"
                    delete(page.dialog)
                    return false
                }
                pane.Size = function(width, height) {
                    pane.style.display = (width<=0 || height<=0)? "none": "block"
                    pane.style.width = width+"px"
                    pane.style.height = height+"px"
                }

                // form init
                pane.Run = form.Run = function(cmds, cb) {
                    ctx.Run(page, form.dataset, cmds, cb)
                }
                pane.Runs = form.Runs = function(cmds, cb) {
                    ctx.Run(page, form.dataset, cmds, function(msg) {
                        ctx.Table(msg, function(line, index) {
                            cb(line, index, msg)
                        })
                    })
                }
                form.onsubmit = function(event) {
                    event.preventDefault()
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
function Plugin(field, inputs, plugin) {
    var option = field.querySelector("form.option")
    var output = field.querySelector("div.output")

    function run(event) {
        var args = []
        option.querySelectorAll("input").forEach(function(item, index){
            item.type == "text" && args.push(item.value)
        })
        option.Run(event, args.slice(1), function(msg) {
            (option.ondaemon || function(msg) {
                output.innerHTML = "",
                msg.append? kit.AppendTable(kit.AppendChild(output, "table"), ctx.Table(msg), msg.append)
                     :kit.AppendChild(output, [{type: "code", list: [{text: [msg.result.join(""), "pre"]}]}])
            })(msg)
        })
    }
    field.onclick = function(event) {
        page.plugin = field
        page.footer.State("action", field.id)
    }

    var ui = kit.AppendChild(option, inputs.map(function(item, index, inputs) {
        item.type == "button"? item.onclick = function(event) {
            plugin[item.click]? plugin[item.click](event, item, option, field): run(event)

        }: (item.onkeyup = function(event) {
            page.oninput(event, function(event) {
                switch (event.key) {
                    case "i":
                        var next = field.nextSibling;
                        next && next.Select()
                        break
                    case "o":
                        var prev = field.previousSibling;
                        prev && prev.Select()
                        break
                    case "c":
                        output.innerHTML = ""
                        break
                    case "r":
                        output.innerHTML = ""
                    case "j":
                        run(event)
                        break
                    case "l":
                        page.action.scrollTo(0, option.parentNode.offsetTop)
                        break
                    case "m":
                        page.View(field.parentNode, "plugin", field.Meta, [], option.Run)
                        event.stopPropagation()
                        break
                    default:
                        return false
                }
                return true
            })
            event.key == "Enter" && (index == inputs.length-1? run(event): event.target.parentNode.nextSibling.childNodes[1].focus())
        }, field.Select = function() {
            ui.last.childNodes[1].focus()
        })
        return {type: "div", list: [{type: "label", inner: item.label||""}, {type: "input", name: item.name, data: item}]}
    }))
    ui.last.childNodes[1].focus()

    plugin = plugin || {}, plugin.__proto__ = {
        show: function() {},
        init: function() {},
    }
    plugin.init(page, page.action, field, option, output)
    page[field.id] = plugin
}
