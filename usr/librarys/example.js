function Page(page) {
    var id = 1
    var conf = {}, conf_cb = {}
    var sync = {}
    page.__proto__ = {
        __proto__: kit,
        ID: function() {
            return id++
        },
        Conf: function(key, value, cb) {
            if (key == undefined) {
                return conf
            }
            if (cb != undefined) {
                conf_cb[key] = cb
            }
            if (value != undefined) {
                var old = conf[key]
                conf[key] = value
                conf_cb[key] && conf_cb[key](value, old)
            }
            return conf[key]
        },
        Sync: function(m) {
            var meta = m, data = "", list = []
            return sync[m] || (sync[m] = {
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
                    if (value == undefined) {
                        return
                    }
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
            })
        },
        View: function(parent, type, line, key, cb) {
            var text = line, list = [], ui = {}
            switch (type) {
                case "icon":
                    list.push({img: [line[key[0]], function(event) {
                        // event.target.scrollIntoView()
                    }]})
                    break

                case "text":
                    list.push({text: [key.length>1? line[key[0]]+"("+line[key[1]]+")":
                        (key.length>0? line[key[0]]: "null"), "span"], click: cb})
                    break

                case "code":
                    list.push({view: ["code", key.length>1? line[key[0]]+"("+line[key[1]]+")":
                        (key.length>0? line[key[0]]: "null")], click: cb})
                    break

                case "table":
                    list.push({type: "table", list: JSON.parse(line.text || "[]").map(function(item, index) {
                        return {type: "tr", list: item.map(function(value) {
                            return {text: [value, index == 0? "th": "td"]}
                        })}
                    })})
                    break

                case "field":
                    var text = JSON.parse(line.text)

                case "plugin":
                    var id = "plugin"+page.ID()
                    list.push({view: [text.view+" item", "fieldset", "", "field"], data: {id: id, Run: cb}, list: [
                        {text: [text.name+"("+text.help+")", "legend"]},
                        {view: ["option", "form", "", "option"], list: [{type: "input", style: {"display": "none"}}]},
                        {view: ["output", "div", "", "output"]},
                        {script: ""+id+".Script="+(text.init||"{}")},
                    ]})
                    break
            }

            parent.DisplayUser && (list = [{view: ["user", "div", line.create_nick||line.create_user]}, {view: ["text"], list:list}])
            !parent.DisplayRaw && (list = [{view: ["item"], list:list}])
            ui = kit.AppendChild(parent, list)
            ui.field && (ui.field.Meta = text)
            return ui
        },
        Include: function(src, cb) {
            kit.AppendChild(document.body, [{include: [src[0], function(event) {
                src.length == 1? cb(event): page.Include(src.slice(1), cb)
            }]}])
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
        oninput: function(event, local) {
            var target = event.target
            kit.History.add("key", (event.ctrlKey? "Control+": "")+(event.shiftKey? "Shift+": "")+event.key)

            if (event.ctrlKey) {
                if (typeof local == "function" && local(event)) {
                    event.stopPropagation()
                    event.preventDefault()
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
                        event.stopPropagation()
                        return true
                    }
            }
            return false
        },
        onscroll: function(event, target, action) {
            switch (event.key) {
                case "h":
                    if (event.ctrlKey) {
                        target.scrollBy(-conf.scroll_x*10, 0)
                    } else {
                        target.scrollBy(-conf.scroll_x, 0)
                    }
                    break
                case "H":
                    target.scrollBy(-document.body.scrollWidth, 0)
                    break
                case "l":
                    if (event.ctrlKey) {
                        target.scrollBy(conf.scroll_x*10, 0)
                    } else {
                        target.scrollBy(conf.scroll_x, 0)
                    }
                    break
                case "L":
                    target.scrollBy(document.body.scrollWidth, 0)
                    break
                case "j":
                    if (event.ctrlKey) {
                        target.scrollBy(0, conf.scroll_y*10)
                    } else {
                        target.scrollBy(0, conf.scroll_y)
                    }
                    break
                case "J":
                    target.scrollBy(0, document.body.scrollHeight)
                    break
                case "k":
                    if (event.ctrlKey) {
                        target.scrollBy(0, -conf.scroll_y*10)
                    } else {
                        target.scrollBy(0, -conf.scroll_y)
                    }
                    break
                case "K":
                    target.scrollBy(0, -document.body.scrollHeight)
                    break
            }
        },

        initHeader: function(page, field, option, output) {
            var state = {}, list = [], cb = function(event, item, value) {}
            field.onclick = function(event) {
                page.pane && page.pane.scrollTo(0,0)
            }
            return {
                Order: function(value, order, cbs) {
                    state = value, list = order, cb = cbs || cb, this.Show()
                },
                Show: function() {
                    output.innerHTML = "", kit.AppendChild(output, [
                        {"view": ["title", "div", "shycontext"], click: function(event) {
                            cb(event, "title", "shycontext")
                        }},
                        {"view": ["state"], list: list.map(function(item) {return {text: [state[item], "div"], click: function(event) {
                            cb(event, item, state[item])
                        }}})},
                    ])
                },
                State: function(name, value) {
                    if (value != undefined) {
                        state[name] = value, this.Show()
                    }
                    if (name != undefined) {
                        return state[name]
                    }
                    return state
                },
            }
        },
        initFooter: function(page, field, option, output) {
            var state = {}, list = [], cb = function(event, item, value) {}
            field.onclick = function(event) {
                page.pane.scrollTo(0,page.pane.scrollHeight)
            }
            return {
                Order: function(value, order, cbs) {
                    state = value, list = order, cb = cbs || cb, this.Show()
                },
                Show: function() {
                    output.innerHTML = "", kit.AppendChild(output, [
                        {"view": ["title", "div", "<a href='mailto:shylinux@163.com'>shylinux@163.com</>"]},
                        {"view": ["state"], list: list.map(function(item) {return {text: [item+":"+state[item], "div"], click: function(item) {
                            cb(event, item, state[item])
                        }}})},
                    ])
                },
                State: function(name, value) {
                    if (value != undefined) {
                        state[name] = value, this.Show()
                    }
                    if (name != undefined) {
                        return state[name]
                    }
                    return state
                },
            }
        },
        initLogin: function(page, field, option, output) {
            var ui = kit.AppendChild(option, [
                {label: "username"}, {input: ["username"]}, {type: "br"},
                {label: "password"}, {password: ["password"]}, {type: "br"},
                {button: ["login", function(event) {
                    if (!ui.username.value) {
                        ui.username.focus()
                        return
                    }
                    if (!ui.password.value) {
                        ui.password.focus()
                        return
                    }
                    field.Pane.Run([ui.username.value, ui.password.value], function(msg) {
                        if (msg.result && msg.result[0]) {
                            field.Pane.ShowDialog(1, 1)
                            ctx.Cookie("sessid", msg.result[0])
                            kit.reload()
                            return
                        }
                        kit.alert("用户或密码错误")
                    })
                }]},
                {button: ["scan", function(event) {
                    scan(event, function(text) {
                        kit.alert(text)
                    })
                }]},
                {type: "br"},
                {type: "img", data: {"src": "/chat/qrcode?text=hi"}}
            ])
            return {
                Exit: function() {
                    ctx.Cookie("sessid", "")
                    kit.reload()

                },
            }
        },
        Pane: Pane,
    }
    window.onload = function() {
        document.querySelectorAll("body>fieldset").forEach(function(field) {
            page.Pane(page, field)
        })
        page.init(page)
        window.onresize = function(event) {
            page.onlayout && page.onlayout(event)
        }
        // document.body.onkeydown = function(event) {
            // page.onscroll && page.onscroll(event, window, "scroll")
        // }
        document.body.onkeydown = function(event) {
            page.oncontrol && page.oncontrol(event, document.body, "control")
        }
    }
    return page
}
function Pane(page, field) {
    var option = field.querySelector("form.option")
    var action = field.querySelector("div.action")
    var output = field.querySelector("div.output")

    var cache = []
    var timer = ""
    var list = [], last = -1
    var conf = {}, conf_cb = {}
    var name = option.dataset.componet_name
    var pane = (page[field.dataset.init] || function() {
    })(page, field, option, output) || {}; pane.__proto__ = {
        __proto__: page,
        Conf: function(key, value, cb) {
            if (key == undefined) {
                return conf
            }
            if (cb != undefined) {
                conf_cb[key] = cb
            }
            if (value != undefined) {
                var old = conf[key]
                conf[key] = value
                conf_cb[key] && conf_cb[key](value, old)
            }
            return conf[key]
        },
        ShowDialog: function(width, height) {
            if (field.style.display != "block") {
                page.dialog && page.dialog != field && page.dialog.style.display == "block" && page.dialog.Show()
                page.dialog = field, field.style.display = "block", kit.ModifyView(field, {window: [width||80, height||200]})
                return true
            }
            field.style.display = "none"
            delete(page.dialog)
            return false
        },
        Size: function(width, height) {
            field.style.display = (width<=0 || height<=0)? "none": "block"
            field.style.width = width+"px"
            field.style.height = height+"px"
        },
        View: function(parent, type, line, key, cb) {
            var ui = page.View(parent, type, line, key, cb)
            if (type == "plugin" || type == "field") {
                pane.Plugin(page, pane, ui.field)
            }
            return ui
        },
        Run: function(cmds, cb) {
            ctx.Run(page, option.dataset, cmds, cb||this.ondaemon)
        },
        Runs: function(cmds, cb) {
            pane.Run(cmds, function(msg) {
                ctx.Table(msg, function(line, index) {
                    (cb||this.ondaemon)(line, index, msg)
                })
            })
        },
        Time: function(time, cmds, cb) {
            function loop() {
                ctx.Run(page, option.dataset, cmds, cb)
                setTimeout(loop, time)
            }
            setTimeout(loop, time)
        },
        Times: function(time, cmds, cb) {
            timer && clearTimeout(timer)
            function loop() {
                !pane.Stop() && ctx.Run(page, option.dataset, cmds, function(msg) {
                    ctx.Table(msg, function(line, index) {
                        cb(line, index, msg)
                    })
                })
                timer = setTimeout(loop, time)
            }
            time && (timer = setTimeout(loop, 10))
        },

        Clear: function() {
            output.innerHTML = "", list = [], last = -1
        },
        Select: function(index) {
            -1 < last && last < list.length && (list[last].className = "item")
            last = index, list[index] && (list[index].className = "item select")
        },
        Append: function(type, line, key, which, cb) {
            var index = list.length, ui = pane.View(output, line.type || type, line, key, function(event, cmds, cbs) {
                pane.Select(index), pane.which.set(line[which])
                typeof cb == "function" && cb(line, index, event, cmds, cbs)
            })
            list.push(ui.last), field.scrollBy(0, field.scrollHeight+100)
            return ui
        },
        Update: function(cmds, type, key, which, first, cb) {
            pane.Clear(), pane.Runs(cmds, function(line, index, msg) {
                var ui = pane.Append(type, line, key, which, cb)
                if (typeof first == "string") {
                    (line.key == first || line.name == first || line[which] == first) && ui.first.click()
                } else {
                    first && index == 0 && ui.first.click()
                }
            })
        },
        Share: function(objs) {
            objs = objs || {}
            objs.componet_name = option.dataset.componet_name
            objs.componet_group = option.dataset.componet_group
            return ctx.Share(objs)
        },
        Save: function(name, output) {
            var temp = document.createDocumentFragment()
            while (output.childNodes.length>0) {
                var item = output.childNodes[0]
                item.parentNode.removeChild(item)
                temp.appendChild(item)
            }
            cache[name] = temp
            return name
        },
        Back: function(name, output) {
            if (!cache[name]) {
                return
            }
            while (cache[name].childNodes.length>0) {
                item = cache[name].childNodes[0]
                item.parentNode.removeChild(item)
                output.appendChild(item)
            }
            delete(cache[name])
            return name
        },
        which: page.Sync(name), Listen: {},
        Action: {}, Button: [], Plugin: Plugin,
    }

    for (var k in pane.Listen) {
        page.Sync(k).change(pane.Listen[k])
    }
    kit.InsertChild(field, output, "div", pane.Button.map(function(value) {
        return typeof value == "object"? {className: value[0], select: [value.slice(1), function(event) {
            value = event.target.value
            typeof pane.Action == "function"? pane.Action(value, event): pane.Action[value](event, value)
        }]}: value == "br"? {"type": "br"}: {"button": [value, function(event) {
            typeof pane.Action == "function"? pane.Action(value, event): pane.Action[value](event, value)
        }]}
    })).className="action "+name
    option.onsubmit = function(event) {
        event.preventDefault()
    };
    return page[name] = field, pane.Field = field, field.Pane = pane
}
function Plugin(page, pane, field) {
    var option = field.querySelector("form.option")
    var output = field.querySelector("div.output")

    var count = 0
    var wait = false
    var plugin = field.Script || {}; plugin.__proto__ = {
        __proto__: pane,
        Append: function(item, name) {
            name = name || item.name

            item.onfocus = function(event) {
                page.pane = pane.Field, page.plugin = field, page.input = event.target
            }
            item.onkeydown = function(event) {
                page.oninput(event, function(event) {
                    switch (event.key) {
                        case "p":
                            action.Back()
                            break
                        case "i":
                            var next = field.nextSibling;
                            next && next.Plugin.Select()
                            break
                        case "o":
                            var prev = field.previousSibling;
                            prev && prev.Plugin.Select()
                            break
                        case "c":
                            output.innerHTML = ""
                            break
                        case "r":
                            output.innerHTML = ""
                        case "j":
                            plugin.Runs(event)
                            break
                        case "l":
                            page.action.scrollTo(0, field.offsetTop)
                            break
                        case "b":
                            plugin.Append(item).focus()
                            break
                        case "m":
                            plugin.Clone().plugin.Plugin.Select()
                            break
                        default:
                            return false
                    }
                    event.stopPropagation()
                    event.preventDefault()
                    return true
                })
                event.key == "Enter" && plugin.Check(action)
            }

            var input = {type: "input", name: name, data: item}
            switch (item.type) {
                case "button":
                    item.onclick = function(event) {
                        action[item.click]? action[item.click](event, item, option, field):
                            plugin[item.click]? plugin[item.click](event, item, option, field): plugin.Runs(event)
                    }
                    break

                case "select":
                    input.type = "select", input.list = item.values.map(function(value) {
                        return {type: "option", value: value, inner: value}
                    }), item.onchange = function(event) {
                        plugin.Check(action)
                    }

                default:
                    args && count < args.length && (item.value = args[count++]||item.value||"")
                    item.className = "args"
            }

            var ui = kit.AppendChild(option, [{view: [item.view||""], list: [{type: "label", inner: item.label||""}, input]}])
            var action = ui[name] || {}

            action.History = [""], action.Goto = function(value) {
                action.History.push(action.value = value)
                plugin.Check(action)
                return value
            }, action.Back = function() {
                action.History.pop(), action.History.length > 0 && action.Goto(action.History.pop())
            };

            (typeof item.imports == "object"? item.imports: typeof item.imports == "string"? [item.imports]: []).forEach(function(imports) {
                page.Sync(imports).change(action.Goto)
            })
            return action
        },
        Prepend: function() {
            var list = option.querySelectorAll(".args")
            list.length > 0 && option.removeChild(list[list.length-1].parentNode)
        },
        Select: function() {
            option.querySelectorAll("input")[1].focus()
        },
        Format: function() {
            field.Meta.args = arguments.length > 0? kit.List(arguments):
                kit.Selector(option, ".args", function(item) {return item.value})
            return JSON.stringify(field.Meta)
        },
        Reveal: function(msg) {
            return msg.append && msg.append[0]? ["table", JSON.stringify(ctx.Tables(msg))]: ["code", msg.result.join("")]
        },
        Remove: function() {
            field.parentNode.removeChild(field)
        },
        Clone: function() {
            field.Meta.args = kit.Selector(option, "input.args", function(item, index) {
                return item.value
            })
            return pane.View(field.parentNode, "plugin", field.Meta, [], field.Run).field.Plugin
        },
        Check: function(target) {
            option.querySelectorAll(".args").forEach(function(item, index, list) {
                item == target && (index == list.length-1? plugin.Runs(event): page.plugin == field && list[index+1].focus())
            })
        },
        Runs: function(event, cb) {
            event.Plugin = plugin, field.Run(event, kit.Selector(option, ".args", function(item, index) {
                return item.value
            }), function(msg) {
                typeof cb == "function" && cb(msg)
                plugin.ondaemon[display.deal||"table"](msg)
            })
        },
        Location: function(event) {
            output.className = "output long"
            page.getLocation(function(res) {
                field.Run(event, [parseInt(res.latitude*1000000+1400)/1000000.0, parseInt(res.longitude*1000000+6250)/1000000.0].concat(
                    kit.Selector(option, ".args", function(item) {return item.value}))
                , plugin.ondaemon)
            })
        },

        Clear: function() {
            output.innerHTML = ""
        },
        ondaemon: {
            table: function(msg) {
                output.innerHTML = ""
                if (display.map) {
                    kit.AppendChild(output, [{img: ["https://gss0.bdstatic.com/8bo_dTSlRMgBo1vgoIiO_jowehsv/tile/?qt=vtile&x=25310&y=9426&z=17&styles=pl&scaler=2&udt=20190622"]}])
                    return
                }
                output.innerHTML = ""
                !display.hide_append && msg.append && kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), ctx.Table(msg), msg.append), exports[1], function(event, value, name, line) {
                    if (line["latitude"]) {
                        page.openLocation(line.latitude, line.longitude, line.location)
                    }
                    page.Sync("plugin_"+exports[0]).set(plugin.onexport[exports[2]||""](value, name))
                });
                (display.show_result || !msg.append) && msg.result && kit.AppendChild(output, [{view: ["code", "div", msg.Results()]}])
            },
        },
        onexport: {
            "": function(value, name) {
                return value
            },
            "pod": function(value, name) {
                if (option[exports[0]].value) {
                    return option[exports[0]].value+"."+value
                }
                return value
            },
            "dir": function(value, name) {
                if (value.endsWith("/")) {
                    return option[exports[0]] + value
                }
            },
        },
        init: function() {},
    }

    var meta = field.Meta
    var args = meta.args || []
    var display = JSON.parse(meta.display||'{}')
    var exports = JSON.parse(meta.exports||'["",""]')
    JSON.parse(meta.inputs || "[]").map(plugin.Append)

    plugin.init(page, pane, field, option, output)
    return page[field.id] = pane[field.id] = plugin.Field = field, field.Plugin = plugin
}
