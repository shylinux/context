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
                    list.push({view: ["code", "div", key.length>1? line[key[0]]+"("+line[key[1]]+")":
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

            var item = []
            parent.DisplayUser && item.push({view: ["user", "div", line.create_nick||line.create_user]})
            parent.DisplayTime && (item.push({text: [line.create_time, "div", "time"]}))
            item.push({view: ["text"], list:list})

            !parent.DisplayRaw && (list = [{view: ["item"], list:item}])
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
                display: "block", dialog: [args.width||text.length*10+100, args.height||60], padding: 10,
            })
            if (!text) {
                toast.style.display = "none"
                return
            }

            var list = [{text: [title||"", "div", "title"]}, {text: [args.text||"", "div", "content"]}]
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
            list.push({view: ["tick"], name: "tick"})

            // kit.ModifyNode(toast.querySelector("legend"), args.title||"tips")
            var ui = kit.AppendChild(kit.ModifyNode(toast.querySelector("div.output"), ""), list)
            var tick = 1
            var begin = kit.time(0,"%H:%M:%S")
            var timer = args.duration ==- 1? setTimeout(function() {
                function ticker() {
                    toast.style.display != "none" && (ui.tick.innerText = begin+" ... "+(tick++)+"s") && setTimeout(ticker, 1000)
                }
                ticker()
            }, 10): setTimeout(function(){toast.style.display = "none"}, args.duration||3000)
            page.toast = toast
            return true
        },
        ondebug: function() {
            if (!this.debug) {
                var pane = Pane(page)
                pane.Field.style.position = "absolute"
                pane.Field.style["background-color"] = "#ffffff00"
                pane.Field.style["color"] = "red"
                pane.ShowDialog(400, 400)
                this.debug = pane
            }
            kit.AppendChild(this.debug.Field, [{text: [JSON.stringify(arguments.length==1? arguments[0]: arguments)]}])
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
                var his = target.History
                var pos = target.Current
                switch (event.key) {
                    case "p":
                        if (!his) { break }
                        pos = (pos-1+his.length+1) % (his.length+1)
                        target.value = pos < his.length? his[pos]: ""
                        target.Current = pos
                        break
                    case "n":
                        if (!his) { break }
                        pos = (pos+1) % (his.length+1)
                        target.value = pos < his.length? his[pos]: ""
                        target.Current = pos
                        break
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
            if (page.localMap && page.localMap(event)) {
                return
            }
            page.oncontrol && page.oncontrol(event, document.body, "control")
        }
    }
    return page
}
function Pane(page, field) {
    field = field || kit.AppendChild(document.body, [{type: "fieldset", list: [{view: ["option", "form"]}, {view: ["output"]}]}]).last
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
    pane.Button && pane.Button.length > 0 && (kit.InsertChild(field, output, "div", pane.Button.map(function(value) {
        return typeof value == "object"? {className: value[0], select: [value.slice(1), function(value, event) {
            value = event.target.value
            typeof pane.Action == "function"? pane.Action(value, event): pane.Action[value](event, value)
        }]}: value == ""? {view: ["space"]} :value == "br"? {type: "br"}: {button: [value, function(value, event) {
            typeof pane.Action == "function"? pane.Action(value, event): pane.Action[value](event, value)
        }]}
    })).className="action "+name)
    option.onsubmit = function(event) {
        event.preventDefault()
    };
    return page[name] = field, pane.Field = field, field.Pane = pane
}
function Plugin(page, pane, field) {
    var option = field.querySelector("form.option")
    var output = field.querySelector("div.output")

    var count = 0
    var plugin = field.Script || {}; plugin.__proto__ = {
        __proto__: pane,
        Append: function(item, name) {
            name = item.name || ""

            item.onfocus = function(event) {
                page.pane = pane.Field, page.plugin = field, page.input = event.target
            }
            item.onkeyup = function(event) {
                page.oninput(event, function(event) {
                    switch (event.key) {
                        case "w":
                            break
                        default:
                            return false
                    }
                    event.stopPropagation()
                    event.preventDefault()
                    return true
                })
            }
            item.onkeydown = function(event) {
                page.oninput(event, function(event) {
                    switch (event.key) {
                        case "w":
                            break
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
                item.type != "textarea" && event.key == "Enter" && plugin.Check(action)
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

                case "textarea":
                    if (item.type == "textarea") {
                        input.type = "textarea"
                        item.style = "height:300px;"+"width:"+(option.clientWidth-20)+"px"
                    }

                default:
                    if (item.type == "text") {
                        item.onclick = function(event) {
                            if (event.ctrlKey) {
                                action.value = kit.History.get("txt", -1).data.trim()
                            }
                        }
                        item.ondblclick = function(event) {
                            action.value = kit.History.get("txt", -1).data.trim()
                        }
                        item.autocomplete = "off"

                    }
                    args && count < args.length && (item.value = args[count++]||item.value||"")
                    item.className = "args"
            }

            var ui = kit.AppendChild(option, [{view: [item.view||""], list: [{type: "label", inner: item.label||""}, input]}])
            var action = ui[name] || {}

            action.History = [""], action.Goto = function(value, cb) {
                action.History.push(action.value = value)
                plugin.Check(action, cb)
                return value
            }, action.Back = function() {
                action.History.pop(), action.History.length > 0 && action.Goto(action.History.pop())
            };

            (typeof item.imports == "object"? item.imports: typeof item.imports == "string"? [item.imports]: []).forEach(function(imports) {
                page.Sync(imports).change(function(value) {
                    (action.value = value) && item.action == "auto" && plugin.Runs(window.event)
                })
            })
            item.type == "button" && item.action == "auto" && plugin.Runs(window.event, function() {
                var td = output.querySelector("td")
                td && td.click()
            })
            return action
        },
        Remove: function() {
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
            return msg.append && msg.append[0]? ["table", JSON.stringify(ctx.Tables(msg))]: ["code", msg.result? msg.result.join(""): ""]
        },
        Delete: function() {
            page.plugin = field.previousSibling
            field.parentNode.removeChild(field)
        },
        Clone: function() {
            field.Meta.args = kit.Selector(option, "input.args", function(item, index) {
                return item.value
            })
            return pane.View(field.parentNode, "plugin", field.Meta, [], field.Run).field.Plugin
        },

        Check: function(target, cb) {
            option.querySelectorAll(".args").forEach(function(item, index, list) {
                item == target && (index == list.length-1? plugin.Runs(event, cb): page.plugin == field && list[index+1].focus())
            })
        },
        Run: function(event, args, cb) {
            var show = true
            setTimeout(function() {
                show && page.ontoast(kit.Format(args||["running..."]), meta.name, -1)
            }, 1000)
            event.Plugin = plugin, field.Run(event, args, function(msg) {
                plugin.msg = msg, show = false, page.ontoast("")
                plugin.ondaemon[display.deal||"table"](msg, cb)
            })
        },
        Runs: function(event, cb) {
            page.footer.Pane.State("ncmd", kit.History.get("cmd").length)
            var args = kit.Selector(option, ".args", function(item, index) {
                return item.value
            })
            this.Run(event, args, cb)
        },
        Delay: function(time, event, text) {
            page.ontoast(text, "", -1)
            setTimeout(function() {
                plugin.Runs(event)
                page.ontoast("")
            }, time)
            return time
        },
        Clear: function() {
            output.innerHTML = ""
        },

        ondaemon: {
            table: function(msg, cb) {
                output.innerHTML = ""
                !display.hide_append && msg.append && kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), ctx.Table(msg), msg.append), exports[1], function(event, value, name, line) {
                    page.Sync("plugin_"+exports[0]).set(plugin.onexport[exports[2]||""](value, name, line))
                });
                (display.show_result || !msg.append) && msg.result && kit.OrderCode(kit.AppendChild(output, [{view: ["code", "div", msg.Results()]}]).first)
                typeof cb == "function" && cb(msg)
            },
            editor: function(msg, cb) {
                (output.innerHTML = "", Editor(plugin, option, output, output.clientWidth-40, 400, 10, msg))
            },
            canvas: function(msg, cb) {
                typeof cb == "function" && !cb(msg) || (output.innerHTML = "", Canvas(plugin, option, output, output.clientWidth-40, 400, 10, msg))
            },
            trend: function(msg, cb) {
                typeof cb == "function" && !cb(msg) || (output.innerHTML = "", Canvas(plugin, output, output.clientWidth-40, 400, 10, msg))
            },
            point: function(msg) {
                var id = "canvas"+page.ID()
                var canvas = kit.AppendChild(output, [{view: ["draw", "canvas"], data: {id: id, width: output.clientWidth-15}}]).last.getContext("2d")
                ctx.Table(msg, function(line) {
                    var meta = JSON.parse(line.meta||"{}")
                    switch (line.type) {
                        case "begin":
                            canvas.beginPath()
                            break

                        case "circle":
                            canvas.arc(parseInt(meta.x), parseInt(meta.y), parseInt(meta.r), 0, Math.PI*2, true)
                            break

                        case "stroke":
                            canvas.strokeStyle = meta.color
                            canvas.lineWidth = parseInt(meta.width)
                            canvas.stroke()
                            break
                    }
                })
            },
            map: function(msg) {
                kit.AppendChild(output, [{img: ["https://gss0.bdstatic.com/8bo_dTSlRMgBo1vgoIiO_jowehsv/tile/?qt=vtile&x=25310&y=9426&z=17&styles=pl&scaler=2&udt=20190622"]}])
            },
        },
        onexport: {
            "": function(value, name) {
                return value
            },
            see: function(value, name, line) {
                return value.split("/")[0]
            },
            you: function(value, name, line) {
                window.event.Plugin = plugin
                line.you && name == "status" && (line.status == "start"? function() {
                    plugin.Delay(3000, window.event, line.you+" stop...") && field.Run(window.event, [line.you, "stop"])
                }(): field.Run(window.event, [line.you], function(msg) {
                    plugin.Delay(3000, window.event, line.you+" start...")
                }))
                return name == "status" || line.status == "stop" ? undefined: line.you
            },
            pod: function(value, name, line) {
                if (option[exports[0]].value) {
                    return option[exports[0]].value+"."+line.pod
                }
                return line.pod
            },
            dir: function(value, name, line) {
                if (name != "path") {
                    value = line.path
                }
                if (value.endsWith("/")) {
                    option.dir.Goto(value)
                    return value
                }

                option.dir.value = value
                plugin.Runs(window.event)
            },
        },
        display: function(arg) {
            display.deal = arg
            plugin.ondaemon[display.deal||"table"](plugin.msg)
        },

        Location: function(event) {
            output.className = "output long"
            page.getLocation(function(res) {
                field.Run(event, [parseInt(res.latitude*1000000+1400)/1000000.0, parseInt(res.longitude*1000000+6250)/1000000.0].concat(
                    kit.Selector(option, ".args", function(item) {return item.value}))
                , plugin.ondaemon)
            })
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
function Editor(plugin, option, output, width, height, space, msg) {
    exports = ["dir", "path", "dir"]
    msg.append && kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), ctx.Table(msg), msg.append), exports[1], function(event, value, name, line) {
        page.Sync("plugin_"+exports[0]).set(plugin.onexport[exports[2]||""](value, name, line))
    });

    var args = [option.pod.value, option.dir.value]

    if (msg.file) {
        var action = kit.AppendAction(kit.AppendChild(output, [{view: ["action"]}]).last, [
            "追加", "提交", "取消",
        ], function(value, event) {
            switch (value) {
                case "追加":
                    field.Run(event, args.concat(["dir_sed", "add"]))
                    break
                case "提交":
                    field.Run(event, args.concat(["dir_sed", "put"]))
                    break
                case "取消":
                    break
            }
        })

        kit.AppendChild(output, [{view: ["edit", "table"], list: msg.result.map(function(value, index) {
            return {view: ["line", "tr"], list: [{view: ["num", "td", index+1]}, {view: ["txt", "td"], list: [{value: value, style: {width: width+"px"}, input: [value, function(event) {
                if (event.key == "Enter") {
                    field.Run(event, args.concat(["dir_sed", "set", index, event.target.value]))
                }
            }]}]}]}
        })}])
    }
}
function Canvas(plugin, option, output, width, height, space, msg) {
    var keys = [], data = {}, max = {}, nline = 0
    var nrow = msg[msg.append[0]].length
    var step = width / (nrow - 1)
    msg.append.forEach(function(key, index) {
        var list = []
        msg[key].forEach(function(value, index) {
            var v = parseInt(value)
            !isNaN(v) && (list.push((value.indexOf("-") == -1)? v: value), v > (max[key]||0) && (max[key] = v))
        })
        list.length == nrow && (keys.push(key), data[key] = list, nline++)
    })

    var conf = {
        font: "monospace", text: "hi", tool: "stroke", style: "black",
        type: "trend", shape: "drawText", means: "drawPoint",
        limits: {scale: 3, drawPoint: 1, drawPoly: 3},

        axies: {style: "black", width: 2},
        xlabel: {style: "red", width: 2, height: 5},
        plabel: {style: "red", font: "16px monospace", offset: 10, height: 20, length: 20},
        data: {style: "black", width: 1},

        mpoint: 10,
        play: 500,
    }

    var view = [], ps = [], point = [], now = {}, index = 0
    var trap = false, label = false

    var what = {
        reset: function(x, y) {
            canvas.resetTransform()
            canvas.setTransform(1, 0, 0, -1, space+(x||0), height+space-(y||0))
            canvas.strokeStyle = conf.data.style
            canvas.fillStyle = conf.data.style
            return what
        },
        clear: function() {
            var p0 = what.transform({x:-width, y:-height})
            var p1 = what.transform({x:2*width, y:2*height})
            canvas.clearRect(p0.x, p0.y, p1.x-p0.x, p1.y-p0.y)
            return what
        },

        move: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save(), what.clear().drawLine(meta)
            canvas.translate(p1.x-p0.x, p1.y-p0.y)
            what.drawData().drawView()
            meta.ps.length < 2 && canvas.restore()
        },
        scale: function(meta) {
            var ps = meta.ps
            var p0 = ps[0] || {x:0, y:0}
            var p1 = ps[1] || now
            var p2 = ps[2] || now

            if (ps.length > 1) {
                canvas.save(), what.clear()
                what.drawLine({ps: [p0, {x: p1.x, y: p0.y}]})
                what.drawLine({ps: [{x: p1.x, y: p0.y}, p1]})
                what.drawLine({ps: [p0, {x: p2.x, y: p0.y}]})
                what.drawLine({ps: [{x: p2.x, y: p0.y}, p2]})
                canvas.scale((p2.x-p0.x)/(p1.x-p0.x), (p2.y-p0.y)/(p1.y-p0.y))
                what.drawData().drawView()
                meta.ps.length < 3 && canvas.restore()
            }
        },
        rotate: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save(), what.clear().drawLine(meta)
            canvas.rotate(Math.atan2(p1.y-p0.y, p1.x-p0.x))
            what.drawData().drawView()
            meta.ps.length < 2 && canvas.restore()
        },

        draw: function(meta) {
            function trans(value) {
                if (value == "random") {
                    return ["black", "red", "green", "yellow", "blue", "purple", "cyan", "white"][parseInt(Math.random()*8)]
                }
                return value
            }
            canvas.strokeStyle = trans(meta.style || conf.style)
            canvas.fillStyle = trans(meta.style || conf.style)
            canvas[meta.tool||conf.tool]()
            return meta
        },
        drawText: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            var t = meta.text||status.cmd.value||conf.text

            canvas.save()
            canvas.translate(p0.x, p0.y)
            canvas.scale(1, -1)
            canvas.rotate(-Math.atan2(p1.y-p0.y, p1.x-p0.x))
            what.draw(meta)
            canvas.font=kit.distance(p0.x, p0.y, p1.x, p1.y)/t.length*2+"px "+conf.font
            canvas[(meta.tool||conf.tool)+"Text"](t, 0, 0)
            canvas.restore()
            return meta
        },
        drawPoint: function(meta) {
            meta.ps.concat(now).forEach(function(p) {
                canvas.save()
                canvas.translate(p.x, p.y)
                canvas.beginPath()
                canvas.moveTo(-conf.mpoint, 0)
                canvas.lineTo(conf.mpoint, 0)
                canvas.moveTo(0, -conf.mpoint)
                canvas.lineTo(0, conf.mpoint)
                what.draw(meta)
                canvas.restore()
            })
            return meta
        },
        drawLine: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save()
            canvas.beginPath()
            canvas.moveTo(p0.x, p0.y)
            canvas.lineTo(p1.x, p1.y)
            what.draw(meta)
            canvas.restore()
            return meta
        },
        drawPoly: function(meta) {
            var ps = meta.ps
            canvas.save()
            canvas.beginPath()
            canvas.moveTo(ps[0].x, ps[0].y)
            for (var i = 1; i < ps.length; i++) {
                canvas.lineTo(ps[i].x, ps[i].y)
            }
            ps.length < conf.limits.drawPoly && canvas.lineTo(now.x, now.y)
            what.draw(meta)
            canvas.restore()
            return meta
        },
        drawRect: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save()
            what.draw(meta)
            canvas[(meta.tool||conf.tool)+"Rect"](p0.x, p0.y, p1.x-p0.x, p1.y-p0.y)
            canvas.restore()
            return meta
        },
        drawCircle: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save()
            canvas.beginPath()
            canvas.arc(p0.x, p0.y, kit.distance(p0.x, p0.y, p1.x, p1.y), 0, Math.PI*2, true)
            what.draw(meta)
            canvas.restore()
            return meta
        },
        drawEllipse: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            var r0 = Math.abs(p1.x-p0.x)
            var r1 = Math.abs(p1.y-p0.y)

            canvas.save()
            canvas.beginPath()
            canvas.translate(p0.x, p0.y)
            r1 > r0? (canvas.scale(r0/r1, 1), r0 = r1): canvas.scale(1, r1/r0)
            canvas.arc(0, 0, r0, 0, Math.PI*2, true)
            what.draw(meta)
            canvas.restore()
            return meta
        },

        drawAxies: function() {
            canvas.beginPath()
            canvas.moveTo(-space, 0)
            canvas.lineTo(width+space, 0)
            canvas.moveTo(0, -space)
            canvas.lineTo(0, height+space)
            canvas.strokeStyle = conf.axies.style
            canvas.lineWidth = conf.axies.width
            canvas.stroke()
            return what
        },
        drawXLabel: function(step) {
            canvas.beginPath()
            for (var pos = step; pos < width; pos += step) {
                canvas.moveTo(pos, 0)
                canvas.lineTo(pos, -conf.xlabel.height)
            }
            canvas.strokeStyle = conf.xlabel.style
            canvas.lineWidth = conf.xlabel.width
            canvas.stroke()
            return what
        },

        figure: {
            trend: {
                draw: function() {
                    what.drawAxies().drawXLabel(step)
                    canvas.beginPath()
                    for (var key in data) {
                        data[key].forEach(function(value, i) {
                            i == 0? canvas.moveTo(0, value/max[key]*height): canvas.lineTo(step*i, value/max[key]*height)
                            i == index && (canvas.moveTo(step*i, 0), canvas.lineTo(step*i, value/max[key]*height))
                        })
                    }
                    canvas.strokeStyle = conf.data.style
                    canvas.lineWidth = conf.data.width
                    canvas.stroke()
                },
                show: function(p) {
                    index = parseInt(p.x/step)
                    canvas.moveTo(p.x, -space)
                    canvas.lineTo(p.x, height)
                    canvas.moveTo(-space, p.y)
                    canvas.lineTo(width, p.y)
                    return p
                },
            },
            ticket: {
                draw: function() {
                    what.drawAxies().drawXLabel(step)
                    if (keys.length < 3) {
                        return
                    }
                    canvas.beginPath()

                    var sum = 0, total = 0
                    for (var i = 0; i < nrow; i++) {
                        sum += data[keys[1]][i]
                        sum > total && (total = sum)
                        sum -= data[keys[2]||keys[1]][i]
                    }
                    if (!data["sum"]) {
                        var sum = 0, max = 0, min = 0, end = 0
                        keys = keys.concat(["sum", "max", "min", "end"])
                        data["sum"] = []
                        data["max"] = []
                        data["min"] = []
                        data["end"] = []
                        for (var i = 0; i < nrow; i++) {
                            max = sum + data[keys[1]][i]
                            min = sum - data[keys[2||keys[1]]][i]
                            end = sum + data[keys[1]][i] - data[keys[2]||keys[1]][i]
                            data["sum"].push(sum)
                            data["max"].push(max)
                            data["min"].push(min)
                            data["end"].push(end)
                            sum = end
                        }
                        msg.append.push("sum")
                        msg.sum = data.sum
                        msg.append.push("max")
                        msg.max = data.max
                        msg.append.push("min")
                        msg.min = data.min
                        msg.append.push("end")
                        msg.end = data.end
                    }

                    for (var i = 0; i < nrow; i++) {
                        if (data["sum"][i] < data["end"][i]) {
                            canvas.moveTo(step*i, data["min"][i]/total*height)
                            canvas.lineTo(step*i, data["sum"][i]/total*height)

                            canvas.moveTo(step*i, data["max"][i]/total*height)
                            canvas.lineTo(step*i, data["end"][i]/total*height)
                        } else {
                            canvas.moveTo(step*i, data["min"][i]/total*height)
                            canvas.lineTo(step*i, data["end"][i]/total*height)

                            canvas.moveTo(step*i, data["max"][i]/total*height)
                            canvas.lineTo(step*i, data["sum"][i]/total*height)
                        }
                    }
                    canvas.strokeStyle = conf.data.style
                    canvas.lineWidth = conf.data.width
                    canvas.stroke()
                },
                show: function(p) {
                    index = parseInt(p.x/step)
                    canvas.moveTo(p.x, -space)
                    canvas.lineTo(p.x, height)
                    canvas.moveTo(-space, p.y)
                    canvas.lineTo(width, p.y)
                    return p
                },
            },
            stick: {
                draw: function() {
                    what.drawAxies().drawXLabel(step)
                    canvas.beginPath()

                    var total = 0
                    for (var key in max) {
                        total += max[key]
                    }

                    for (var i = 0; i < nrow; i++) {
                        canvas.moveTo(step*i, 0)
                        for (var key in data) {
                            canvas.lineTo(step*i, data[key][i]/total*height)
                            canvas.moveTo(step*i-step/2, data[key][i]/total*height)
                            canvas.lineTo(step*i+step/2, data[key][i]/total*height)
                            canvas.moveTo(step*i, data[key][i]/total*height)
                        }
                    }
                    canvas.strokeStyle = conf.data.style
                    canvas.lineWidth = conf.data.width
                    canvas.stroke()
                },
                show: function(p) {
                    index = parseInt(p.x/step)
                    canvas.moveTo(p.x, -space)
                    canvas.lineTo(p.x, height)
                    canvas.moveTo(-space, p.y)
                    canvas.lineTo(width, p.y)
                    return p
                },
            },
            weight: {
                conf: {
                    space: 20,
                    focus: "white",
                    style: "black",
                    width: 1,
                    least: 0.01,
                },
                draw: function() {
                    var that = this
                    var space = width / (nline+1)

                    canvas.translate(0, height/2)
                    for (var key in data) {
                        var total = 0
                        data[key].forEach(function(value) {
                            total += value
                        })

                        var sum = 0
                        canvas.translate(space, 0)
                        data[key].forEach(function(value, i) {
                            if (value/total < that.conf.least) {
                                return
                            }

                            var a = sum/total*Math.PI*2
                            var b = (sum+value)/total*Math.PI*2
                            sum+=value

                            canvas.beginPath()
                            canvas.moveTo(0, 0)
                            canvas.arc(0, 0, (space/2)-that.conf.space, a, b, false)
                            canvas.closePath()

                            if (i == index) {
                                canvas.fillStyle = that.conf.focus
                                canvas.fill()
                            } else {
                                canvas.strokeStyle = that.conf.style
                                canvas.lineWidth = that.conf.width
                                canvas.stroke()
                            }
                        })
                    }
                },
                show: function(p) {
                    var nspace = width / (nline+1)
                    var which = parseInt((p.x-nspace/2)/nspace)
                    which >= nline && (which = nline-1), which < 0 && (which = 0)

                    var q = what.reverse(p)
                    canvas.translate((which+1)*nspace, height/2)
                    var p = what.transform(q)

                    var a = Math.atan2(p.y, p.x)
                    a < 0 && (a += Math.PI*2)
                    var pos = a/2/Math.PI

                    var total = 0
                    data[keys[which]].forEach(function(value) {
                        total += value
                    })
                    var sum = 0, weight = 0
                    data[keys[which]].forEach(function(value, i) {
                        sum += value, sum / total < pos && (index = i+1)
                        index == i && (weight = parseInt(value/total*100))
                    })

                    canvas.moveTo(0, 0)
                    canvas.lineTo(p.x, p.y)
                    canvas.lineTo(p.x+conf.plabel.length, p.y)

                    canvas.scale(1, -1)
                    canvas.fillText("weight: "+weight+"%", p.x+conf.plabel.offset, -p.y+conf.plabel.offset)
                    canvas.scale(1, -1)
                    return p
                },
            },
        },

        drawData: function() {
            canvas.save()
            what.figure[conf.type].draw()
            canvas.restore()
            return what
        },
        drawView: function() {
            view.forEach(function(view) {
                view.meta && what[view.type](view.meta)
            })
            return what
        },
        drawLabel: function() {
            if (!label) { return what }

            index = 0
            canvas.save()
            canvas.font = conf.plabel.font || conf.font
            canvas.fillStyle = conf.plabel.style || conf.style
            canvas.strokeStyle = conf.plabel.style || conf.style
            var p = what.figure[conf.type].show(now)
            canvas.stroke()

            canvas.scale(1, -1)
            p.x += conf.plabel.offset
            p.y -= conf.plabel.offset

            if (width - p.x < 200) {
                p.x -= 200
            }
            canvas.fillText("index: "+index, p.x, -p.y+conf.plabel.height)
            msg.append.forEach(function(key, i) {
                msg[key][index] && canvas.fillText(key+": "+msg[key][index], p.x, -p.y+(i+2)*conf.plabel.height)
            })
            canvas.restore()
            return what
        },
        drawShape: function() {
            point.length > 0 && (what[conf.shape]({ps: point}), what[conf.means]({ps: point, tool: "stroke", style: "red"}))
            return what
        },

        refresh: function() {
            return what.clear().drawData().drawView().drawLabel().drawShape()
        },
        cancel: function() {
            point = [], what.refresh()
            return what
        },
        play: function() {
            function cb() {
                view[i] && what[view[i].type](view[i].meta) && (t = kit.Delay(view[i].type == "drawPoint"? 10: conf.play, cb))
                i++
                status.nshape.innerText = i+"/"+view.length
            }
            var i = 0
            what.clear().drawData()
            kit.Delay(10, cb)
            return what
        },
        back: function() {
            view.pop(), status.nshape.innerText = view.length
            return what.refresh()
        },
        push: function(item) {
            item.meta && item.meta.ps < (conf.limits[item.type]||2) && ps.push(item)
            status.nshape.innerText = view.push(item)
            return what
        },
        wait: function() {
            status.cmd.focus()
            return what
        },
        trap: function(value, event) {
            event.target.className = (trap = !trap)? "trap": "normal"
            page.localMap = trap? what.input: undefined
        },
        label: function(value, event) {
            event.target.className = (label = !label)? "trap": "normal"
        },

        movePoint: function(p) {
            now = p, status.xy.innerHTML = p.x+","+p.y;
            (point.length > 0 || ps.length > 0 || label) && what.refresh()
        },
        pushPoint: function(p) {
            if (ps.length > 0) {
                ps[0].meta.ps.push(p) > 1 && ps.pop(), what.refresh()
                return
            }

            point.push(p) >= (conf.limits[conf.shape]||2) && what.push({type: conf.shape,
                meta: what[conf.shape]({ps: point, text: status.cmd.value||conf.text, tool: conf.tool, style: conf.style}),
            }) && (point = [])
            conf.means == "drawPoint" && what.push({type: conf.means, meta: what[conf.means]({ps: [p], tool: "stroke", style: "red"})})
        },
        transform: function(p) {
            var t = canvas.getTransform()
            return {
                x: (p.x-t.c/t.d*p.y+t.c*t.f/t.d-t.e)/(t.a-t.c*t.b/t.d),
                y: (p.y-t.b/t.a*p.x+t.b*t.e/t.a-t.f)/(t.d-t.b*t.c/t.a),
            }
        },
        reverse: function(p) {
            var t = canvas.getTransform()
            return {
                x: t.a*p.x+t.c*p.y+t.e,
                y: t.b*p.x+t.d*p.y+t.f,
            }
        },

        check: function() {
            view.forEach(function(item, index, view) {
                item && item.send && plugin.Run(window.event||{}, item.send.concat(["type", item.type]), function(msg) {
                    msg.text && msg.text[0] && (item.meta.text = msg.text[0])
                    msg.style && msg.style[0] && (item.meta.style = msg.style[0])
                    msg.ps && msg.ps[0] && (item.meta.ps = JSON.parse(msg.ps[0]))
                    what.refresh()
                })
                index == view.length -1 && kit.Delay(1000, what.check)
            })
        },
        parse: function(txt) {
            var meta = {}, cmds = [], rest = -1, send = []
            txt.trim().split(" ").forEach(function(item) {
                switch (item) {
                    case "stroke":
                    case "fill":
                        meta.tool = item
                        break
                    case "black":
                    case "white":
                    case "red":
                    case "yellow":
                    case "green":
                    case "cyan":
                    case "blue":
                    case "purple":
                        meta.style = item
                        break
                    case "cmds":
                        rest = cmds.length
                    default:
                        cmds.push(item)
                }
            }), rest != -1 && (send = cmds.slice(rest+1), cmds = cmds.slice(0, rest))

            var cmd = {
                "t": "drawText",
                "l": "drawLine",
                "p": "drawPoly",
                "r": "drawRect",
                "c": "drawCircle",
                "e": "drawEllipse",
            }[cmds[0]] || cmds[0]
            cmds = cmds.slice(1)

            var args = []
            switch (cmd) {
                case "send":
                    plugin.Run(window.event, cmds, function(msg) {
                        kit.Log(msg)
                    })
                    return
                default:
                    meta.ps = []
                    for (var i = 0; i < cmds.length; i+=2) {
                        var x = parseInt(cmds[i])
                        var y = parseInt(cmds[i+1])
                        !isNaN(x) && !isNaN(y) && meta.ps.push({x: x, y: y}) || (args.push(cmds[i]), i--)
                    }
            }
            meta.args = args

            switch (cmd) {
                case "drawText":
                    meta.text = args.join(" "), delete(meta.args)
                case "drawLine":
                case "drawPoly":
                case "drawRect":
                case "drawCircle":
                case "drawEllipse":
                    what.push({type: cmd, meta: what[cmd](meta), send:send})
            }

            return (what[cmd] || function() {
                return what
            })(meta)
        },
        input: function(event) {
            var map = what.trans[event.key]
            map && action[map[0]] && (action[map[0]].value = map[1])
            map && what.trans[map[0]] && (map = what.trans[map[1]]) && (conf[map[0]] && (conf[map[0]] = map[1]) || what[map[0]] && what[map[0]]())
            what.refresh()
        },
        trans: {
            "折线图": ["type", "trend"],
            "股价图": ["type", "ticket"],
            "柱状图": ["type", "stick"],
            "饼状图": ["type", "weight"],

            "移动": ["shape", "move"],
            "旋转": ["shape", "rotate"],
            "缩放": ["shape", "scale"],

            "文本": ["shape", "drawText"],
            "直线": ["shape", "drawLine"],
            "折线": ["shape", "drawPoly"],
            "矩形": ["shape", "drawRect"],
            "圆形": ["shape", "drawCircle"],
            "椭圆": ["shape", "drawEllipse"],

            "辅助点": ["means", "drawPoint"],
            "辅助线": ["means", "drawRect"],

            "画笔": ["tool", "stroke"],
            "画刷": ["tool", "fill"],

            "黑色": ["style", "black"],
            "红色": ["style", "red"],
            "绿色": ["style", "green"],
            "黄色": ["style", "yellow"],
            "蓝色": ["style", "blue"],
            "紫色": ["style", "purple"],
            "青色": ["style", "cyan"],
            "白色": ["style", "white"],
            "随机色": ["style", "random"],
            "默认色": ["style", "default"],

            "清屏": ["clear"],
            "刷新": ["refresh"],
            "取消": ["cancel"],
            "播放": ["play"],
            "回退": ["back"],
            "输入": ["wait"],

            "标签": ["label"],
            "快捷键": ["trap"],

            "x": ["折线图", "折线图"],
            "y": ["折线图", "饼状图"],

            "a": ["移动", "旋转"],
            "m": ["移动", "移动"],
            "z": ["移动", "缩放"],

            "t": ["文本", "文本"],
            "l": ["文本", "直线"],
            "v": ["文本", "折线"],
            "r": ["文本", "矩形"],
            "c": ["文本", "圆形"],
            "e": ["文本", "椭圆"],

            "s": ["画笔", "画笔"],
            "f": ["画笔", "画刷"],

            "0": ["黑色", "黑色"],
            "1": ["黑色", "红色"],
            "2": ["黑色", "绿色"],
            "3": ["黑色", "黄色"],
            "4": ["黑色", "蓝色"],
            "5": ["黑色", "紫色"],
            "6": ["黑色", "青色"],
            "7": ["黑色", "白色"],
            "8": ["黑色", "随机色"],
            "9": ["黑色", "默认色"],

            "j": ["刷新", "刷新"],
            "g": ["播放", "播放"],
            "b": ["回退", "回退"],
            "q": ["清空", "清空"],

            "Escape": ["取消", "取消"],
            " ": ["输入", "输入"],
        },
    }

    var action = kit.AppendAction(kit.AppendChild(output, [{view: ["action"]}]).last, [
        ["折线图", "股价图", "柱状图", "饼状图"],
        ["移动", "旋转", "缩放"],
        ["文本", "直线", "折线", "矩形", "圆形", "椭圆"],
        ["辅助点", "辅助线"],
        ["画笔", "画刷"],
        ["黑色", "红色", "绿色", "黄色", "蓝色", "紫色", "青色", "白色", "随机色", "默认色"],
        "", "清屏", "刷新", "播放", "回退",
        "", "标签", "快捷键",
    ], function(value, event) {
        var map = what.trans[value]
        conf[map[0]] && (conf[map[0]] = map[1]) || what[map[0]] && what[map[0]](value, event)
        what.refresh()
    })

    var canvas = kit.AppendChild(output, [{view: ["draw", "canvas"], data: {width: width+20, height: height+20,
        onclick: function(event) {
            what.pushPoint(what.transform({x: event.offsetX, y: event.offsetY}), event.clientX, event.clientY)
        }, onmousemove: function(event) {
            what.movePoint(what.transform({x: event.offsetX, y: event.offsetY}), event.clientX, event.clientY)
        },
    }}]).last.getContext("2d")

    var status = kit.AppendStatus(kit.AppendChild(output, [{view: ["status"]}]).last, [{name: "nshape"}, {"className": "cmd", style: {width: (output.clientWidth - 100)+"px"}, data: {autocomplete: "off"}, input: ["cmd", function(event) {
        var target = event.target
        event.type == "keyup" && event.key == "Enter" && what.parse(target.value) && (!target.History && (target.History=[]),
            target.History.push(target.value), target.Current=target.History.length, target.value = "")
        event.type == "keyup" && page.oninput(event), event.stopPropagation()

    }]}, {name: "xy"}], function(value, name, event) {

    })

    return what.reset().refresh()
}

