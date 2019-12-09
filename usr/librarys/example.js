function Meta(zone, target, obj) {
    // 级连对象
    var a = obj
    for (var i = 3; i < arguments.length; i++) {
        a.__proto__ = arguments[i], a = arguments[i]
    }

    // 构造对象
    var id = 1
    var conf = {}, conf_cb = {}, old
    var sync = {}
    var cache = {}, datas = {}
    var history = []
    var meta = {__proto__: obj, target: target,
        ID: shy("单一序列", function() {return id++}),
        Conf: shy("配置变量", function(key, value, cb) {
            if (kit.isNone(key)) {return conf}
            kit.notNone(cb) && (conf_cb[key] = cb)
            if (kit.notNone(value)) {
                old = conf[key], conf[key] = value
                kit.Log("config", key, value, old)
                kit._call(conf_cb[key], [value, old])
            }
            return kit.isNone(conf[key]) && obj && obj.Conf? obj.Conf(key): conf[key]
        }),
        Sync: shy("同步变量", function(m) {
            var meta = m, data = "", list = []
            return sync[m] || (sync[m] = {
                change: function(cb) {list.push(cb); return list.length-1},
                eq: function(value) {return data == value},
                neq: function(value) {return data != value},
                get: function() {return data},
                set: function(value, force) {
                    if (kit.isNone(value)) {return}
                    if (value == data && !force) {return}
                    old = data, data = value
                    meta && kit.Log("key", meta, value, old)
                    kit.List(list, function(cb) {cb(value, old)})
                    return value
                },
            })
        }),
        Save: shy("保存视图", function(name, output, data) {if (name === "") {return cache = {}}
            kit.Log("view", "save", meta.Zone(name).join("."))
            var temp = document.createDocumentFragment()
            while (output.childNodes.length>0) {
                var item = output.childNodes[0]
                item.parentNode.removeChild(item)
                temp.appendChild(item)
            }
            cache[name] = temp
            datas[name] = data || {}
            return name
        }),
        Load: shy("恢复视图", function(name, output) {if (kit.isNone(cache[name])) {return}
            kit.Log("view", "load", meta.Zone(name).join("."))
            while (cache[name].childNodes.length>0) {
                var item = cache[name].childNodes[0]
                item.parentNode.removeChild(item)
                output.appendChild(item)
            }
            delete(cache[name])
            return datas[name]
        }),
        View: shy("添加视图", function(output, type, line, key) {
            var text = line, list = [], item = false, style = ""
            switch (type) {
                case "icon":
                    list.push({img: [line[key[0]]]})
                    break

                case "text":
                    list.push({text: [key.length>1? line[key[0]]+"("+line[key[1]]+")":
                        (key.length>0? line[key[0]]: "null"), "span"]})
                    break

                case "code":
                    list.push({view: ["code", "div", key.length>1? line[key[0]]+"("+line[key[1]]+")":
                        (key.length>0? line[key[0]]: "null")]})
                    break

                case "table":
                    list.push({type: "table", list: JSON.parse(line.text || "[]").map(function(item, index) {
                        return {type: "tr", list: item.map(function(value) {
                            return {text: [value, index == 0? "th": "td"]}
                        })}
                    })})
                    break

                case "input":
                    style = " "+line.type
                    list.push(line)
                    break

                case "field": text = JSON.parse(line.text)
                case "plugin": if (!text.name) {return {}}
                    var id = "plugin"+meta.ID()
                    list.push({view: ["item", "fieldset"], data: {id: id, draggable: false}, list: [
                        {text: [text.name+"("+text.help+")", "legend"]},
                        {view: ["option", "form"], list: [{type: "input", style: {"display": "none"}}]},
                        {view: ["output", "div"]},
                    ]}), item = true
                    break
            }

            var ui = kit.AppendChild(output, item? list: [{view: ["item"+style], data: {id: "item"+meta.ID(), draggable: false}, list:list}])
            return ui["item"+style].Meta = text, ui
        }),
        Include: shy("加载脚本", function(src, cb) {src = kit.List(src)
            function next(event) {src.length > 1? meta.Include(src.slice(1), cb): cb(event)}
            kit.AppendChild(target, [file.endsWith(".css")? {require: [src[0], next]}: {include: [src[0], next]}])
        }),
        History: shy("操作历史", function(value, target) {var item
            return kit.isNone(value)? (item = history.pop()) && (item.target.value = item.value):
                history.push({value: value, target: target})
        }),
        Event: shy("事件入口", {name: zone}, function(event, msg, proto) {
            return ctx.Event(event, msg, proto||arguments.callee.meta)
        }),
        _call: function(key) {
            kit._call(meta, typeof key == "function"? key: meta[key], kit.List(arguments).slice(1))
            return true
        },
        Zones: function(name) {return zone.concat(kit.List(arguments)).join(".")},
        Zone: function(name) {return zone.concat(kit.List(arguments))},
    }

    // 注册事件
    meta.onaction && kit.Item(meta.onaction.meta, function(key, cb) {target[key] = function(event) {
        meta.onaction(event, key, cb)
    }})
    return meta
}
function Page(page) {
    kit.device.isWeiXin =  true
    var script = {}, record = ""
    page = Meta([document.title], document.body, page, {check: true,
        onload: function(event) {
            // Event入口 0
            if (page.Event(event, {}) && page.check && !ctx.Cookie("sessid")) {
                // 用户登录
                kit.Selector(document, "body>fieldset.Login", function(field) {
                    page.Pane(page, field)
                }), page.login.Pane.Dialog(1, 1)
            } else {
                // 登录成功
                document.querySelectorAll("body>fieldset").forEach(function(field) {
                    page.Pane(page, field)
                }), page.check? page.login.Pane.Run(event, [], function(msg) {
                    msg.result && msg.result[0]? (page.who.set(msg.nickname[0]), page.init(page))
                        :page.login.Pane.Dialog(1, 1)
                }): page.init(page)
                page.Button()
                page.Status()
            }

            // 事件回调
            window.onresize = function(event) {
                page.onlayout(event)
            }, document.body.onkeydown = function(event) {
                // page.oncontrol(event) || page.onscroll(event)
            }, document.body.onkeyup = function(event) {
            }, document.body.oncontextmenu = function(event) {
            }, document.body.onmousewheel = function(event) {
            }, document.body.onmousedown = function(event) {
            }, document.body.onmouseup = function(event) {
            }
        },
        oninput: function(event, local) {var target = event.target
            kit.History("key", -1, (event.ctrlKey? "Control+": "")+(event.shiftKey? "Shift+": "")+event.key)

            if (event.ctrlKey) {
                if (local && kit._call(local, [event])) {
                    event.stopPropagation()
                    event.preventDefault()
                    return true
                }

                var his = target.History || []
                var pos = target.Current || -1
                switch (event.key) {
                    case "p":
                        pos = (pos-1+his.length+1) % (his.length+1)
                        target.value = pos < his.length? his[pos]: ""
                        target.Current = pos
                        break
                    case "n":
                        pos = (pos+1) % (his.length+1)
                        target.value = pos < his.length? his[pos]: ""
                        target.Current = pos
                        break
                    case "a":
                    case "e":
                    case "f":
                    case "b":
                        break
                    case "h":
                        kit.DelText(target, target.selectionStart-1, target.selectionStart)
                        break
                    case "d":
                        kit.DelText(target, 0, target.selectionStart)
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
            } else {
                switch (event.key) {
                    case " ":
                        event.stopPropagation()
                        return true
                    case "Escape":
                        target.blur()
                        break
                    default:
                        if (kit.HitText(target, "jk")) {
                            kit.DelText(target, target.selectionStart-2, 2)
                            target.blur()
                            break
                        }
                        return false
                }
            }

            event.stopPropagation()
            event.preventDefault()
            return true
        },
        onlayout: function() {},
        oncontrol: function() {},
        onscroll: function(event, target, action) {
            switch (event.key) {
                case " ":
                    page.footer.Pane.Select()
                    break
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
        title: function(title) {
            document.title = title
        },

        Require: function(file, cb) {
            if (!file || Plugin[file]) {return kit._call(cb, [Plugin[file]])}
            file.endsWith(".css")? kit.AppendChild(document.body, [{require: ["/require/"+file, function(event) {
                return Plugin[file] = file, kit._call(cb, [Plugin[file]])
            }]}]): kit.AppendChild(document.body, [{include: ["/require/"+file, function(event) {
                return kit._call(cb, [Plugin[file]])
            }]}])
        },
        script: function(action, name, time) {
            switch (action) {
                case "create":
                    record = name, script[name] = []
                    kit.Log("script", action, name)
                    break
                case "record":
                    record && kit.Log("script", action, record, name)
                    record && script[record].push(name)
                    break
                case "finish":
                    kit.Log("script", action, record)
                    record = ""
                    break
                case "replay":
                    kit.Log("script", action, name)
                    record = ""

                    var event = window.event
                    kit.List(script[name], function(item) {
                        kit.Log("script", action, name, item)
                        page.action.Pane.Core(event, {}, ["_cmd", item]);
                    }, time||1000, function() {
                        page.toast.Pane.Show("run "+name+" done")
                    })
                    break
                default:
                    return script
            }
            return true
        },
        Help: function(pane, type, action) {return []},
        Jshy: shy("本地命令", function(event, args) {var msg = ctx.Event(event)
            // 面板命令
            if (page[args[0]] && page[args[0]].type == "fieldset") {
                return page[args[0]].Pane.Jshy(event, args.slice(1)) || page.Zone("select", args[0])
            }
            // 控件命令
            if (typeof page.Action[args[0]] == "function") {
                return kit._call(page.Action[args[0]], [event, args[0]]) || page.Zone("action", args[0])
            }

            // 脚本命令
            if (script[args[0]]) {return page.script("replay", args[0])}

            // 内部命令
            if (typeof page[args[0]] == "function") {
                return kit._call(page[args[0]], args.slice(1)) || page.Zone("function", args[0])
            }
        }),
        WSS: function(cb, onerror, onclose, onopen) {
            return page.socket || (page.socket = ctx.WSS(cb || (function(event, m) {
                if (m.detail) {
                    page.action.Pane.Core(event, m, ["_cmd", m.detail], m.Reply)
                } else {
                    page.toast.Pane.Show(m.result.join(""))
                }

            }), onerror || (function() {
                page.socket.close()

            }), onclose || (function() {
                page.toast.Pane.Show("WSS Close")
                delete(page.socket), setTimeout(function() {
                    page.WSS(cb, onerror, onclose)
                }, 1000)
            }), onopen || (function() {
                page.toast.Pane.Show("WSS Open")
            })))
        },

        initToast: function(page, field, option, output) {
            return {
                Dialog: function(width, height) {
                    kit.ModifyView(field, {display: "block", dialog: [width, height], padding: 10})
                },
                Ticker: function(text, duration) {
                    var tick = 1
                    var begin = kit.time(0, "%H:%M:%S")
                    function ticker() {
                        field.style.display != "none" && (text.innerText = begin+" ... "+(tick++)+"s") && setTimeout(ticker, 1000)
                    }
                    return duration == -1? setTimeout(ticker, 10): setTimeout(field.Pane.Hide, duration||3000)
                },
                Show: function(text, title, duration) {if (!text) {return field.Pane.Hide()}
                    var args = typeof text == "object"? text: {text: text, title: title, duration: duration}

                    var list = [{text: [args.title||"", "div", "title"]},
                        typeof args.text == "string"? {text: [args.text||"", "div", "content"]}: args.text]

                    kit.List(args.inputs, function(input) {
                        typeof input == "string"? list.push({label: input}, {input: [input, page.oninput]}):
                            list.push({label: input[0]}, {select: input.slice(1)})
                        list.push({type: "br"})
                    })
                    kit.List(args.button, function(input) {
                        list.push({button: [input, function(event) {
                            var values = {}
                            output.querySelectorAll("input").forEach(function(input) {
                                values[input.name] = input.value
                            })
                            output.querySelectorAll("select").forEach(function(input) {
                                values[input.name] = input.value
                            })
                            kit._call(args.cb, [input, values]) && field.Pane.Hide()
                        }]})
                    })
                    list.push({view: ["tick"], name: "tick"})

                    field.Pane.Dialog(args.width||text.length*10+100, args.height||80)
                    return field.Pane.Ticker(kit.AppendChilds(output, list).tick, args.button? -1: args.duration || 3000)
                },
            }
        },
        initCarte: function(page, field, option, output) {
            field.onmouseleave = function(event) {field.Pane.Hide()}
            return {
                Show: function(event, cb) {if (!cb.list || cb.list.length == 0) {return}
                    output.innerHTML = ""
                    kit.AppendActions(output, cb.list, function(event, value) {
                        kit._call(cb, [event, value, cb.meta]) && field.Pane.Hide()
                    }, true)

                    var pos = {display: "block", left: event.x, top: event.y}
                    if (document.body.clientWidth - event.x < 60) {
                        var pos = {display: "block", right: event.x, top: event.y}
                    }

                    kit.ModifyView(field, pos)
                    event.stopPropagation()
                    event.preventDefault()
                },
            }
        },
        initDebug: function(page, field, option, output) {
            var table = kit.AppendChilds(output, "table")
            var caption = kit.AppendChild(table, [{type: "caption"}]).last
            var head = kit.AppendChild(table, [{type: "thead", list: [{row: ["index", "time", "type", "order", "action", "target", "args"], sub: "th"}]}]).tr
            var list = kit.AppendChild(table, "tbody")
            kit.OrderTable(table)
            var last, types = {all: 0, event: 0, run: 0, key: 0}


            kit.Log.meta.call.push(function(time, type, order, action, target) {var Choice = field.Pane && field.Pane.Choice || []
                if (kit.isNone(types[type])) {types[type] = 0, Choice.push(type)}
                types[type]++

                last = kit.AppendChild(list, [{className: type, row: [types.all++, time, type, order, action||"", target||"", kit.List(arguments, function(item) {
                    return typeof item == "object"? "{...}": item
                }).slice(5).join(" ")]}]).last
                field.Pane && field.Pane.Head()

                kit.AppendChilds(caption, kit.List(Choice.slice(1), function(item) {return {text: [item+": "+types[item], "span"], click: function(event) {
                    field.Pane.Action(event, item)
                }}}))
            })
            var layout
            return {
                Head: function() {if (kit.isNone(last)) {return}
                    caption.style.width = last.offsetWidth+"px"
                    kit.Selector(last, "td", function(item, index) {
                        head.childNodes[index] && (head.childNodes[index].style.width = item.offsetWidth-14+"px")
                    })
                },
                Show: function() {layout || (layout = field.Pane.Action.meta["最大"](), 1)
                    kit.ModifyView(field, {display: field.style.display != "block"? "block": "none"})
                    field.Pane.Head()
                },
                clear: function() {
                    var th = kit.AppendChilds(list, [{row: ["time", "type", "main", "arg", "args"], sub: "th"}]).last
                },
                Action: shy({
                    "关闭": function() {
                        field.Pane.Show()
                    },
                    "最大": function() {
                        kit.size(output, document.body.clientWidth, document.body.clientHeight-160)
                    },
                    "最小": function() {
                        kit.size(output, document.body.clientWidth/2, document.body.clientHeight/2)
                    },
                    "左边": function() {
                        field.style.left = "0px"
                        kit.size(output, document.body.clientWidth/2, document.body.clientHeight)
                    },
                    "右边": function() {
                        field.style.left = document.body.clientWidth/2+"px"
                        kit.size(output, document.body.clientWidth/2, document.body.clientHeight)
                    },
                }, function(event, type) {last = null
                    kit.Selector(list, "tr", function(item, index) {
                        type == "all" || kit.classList.has(item, type)?
                            (kit.classList.del(item, "hide"), last = item): kit.classList.add(item, "hide")
                    })
                    field.Pane.Head()
                }),
                Choice: ["关闭", "all", "event", "run", "key"],
                Button: ["关闭", "最大", "最小", "左边", "右边"],
            }
        },
        initLogin: function(page, field, option, output) {
            var ui = kit.AppendChilds(option, [
                {label: "username"}, {input: ["username"], data: {autocomplete: "username"}}, {type: "br"},
                {label: "password"}, {password: ["password"], data: {autocomplete: "current-password"}}, {type: "br"},
                {button: ["login", function(value, event) {
                    if (!ui.username.value) {ui.username.focus(); return}
                    if (!ui.password.value) {ui.password.focus(); return}

                    field.Pane.Login(ui.username.value, ui.password.value, function(sessid) {
                        if (!sessid) {kit.alert("用户或密码错误"); return}
                        page.login.Pane.Dialog(1, 1), page.onload(event)
                    })
                }]}, {type: "br"},
            ])

            return {
                WeiXin: function() {
                    // 微信接口
                    kit.device.isWeiXin && page.login.Pane.Run(event, ["weixin"], function(msg) {
                        msg.appid[0] && page.Include(["https://res.wx.qq.com/open/js/jweixin-1.4.0.js"], function(event) {
                            wx.error(function(res){})
                            wx.ready(function(){
                                page.scanQRCode = function(cb) {

                                }
                                page.getLocation = function(cb) {
                                    wx.getLocation({success: cb})
                                }
                                page.openLocation = function(latitude, longitude, name) {
                                    wx.openLocation({latitude: parseFloat(latitude), longitude: parseFloat(longitude), name:name||"here"})
                                }
                            }), wx.config({jsApiList: ["closeWindow", "scanQRCode", "getLocation", "openLocation"],
                                appId: msg.appid[0], nonceStr: msg.nonce[0], timestamp: msg.timestamp[0], signature: msg.signature[0]})
                        })
                    })
                },
                Login: function(username, password, cb) {
                    field.Pane.Run(event, [username, password], function(msg) {cb(msg.result && msg.result[0] || "")})
                },
                Exit: function() {ctx.Cookie("sessid", ""), kit.reload()},
            }
        },
        initHeader: function(page, field, option, output) {
            var cb = function(event, item, value) {kit._call(page.Action[item], [event, item, value, page])}
            field.onclick = function(event) {page.pane && page.pane.scrollTo(0, 0)}
            page.who.change(function(value, old) {page.Button("user", value)})

            return {
                Show: function() {var meta = page.Button.meta, list = page.Button.list
                    kit.AppendChilds(output, [
                        {"view": ["title", "div", meta.title], click: function(event) {
                            cb(event, "title", meta.title)
                        }},
                        {"view": ["state"], list: list.map(function(item) {return {text: [meta[item], "div", "item"], click: function(event) {
                            cb(event, item, meta[item])
                        }}})},
                    ])
                },
				Help: function() {return []},
            }
        },
        initFooter: function(page, field, option, output) {
            var cb = function(event, item, value) {kit._call(page.Action[item], [event, item, value, page])}
            var ui = kit.AppendChilds(output, [
                {"view": ["title", "div", page.Status.meta.title]},
                {"view": ["state"]},
                {"view": ["magic"], list: [{label: "0", name: "count"}, {input: ["magics", function(event) {
                    if (event.key == "Enter" || event.ctrlKey && event.key == "j") {
                        page.action.Pane.Core(event, {}, ["_cmd", event.target.value], function(msg) {
                            page.toast.Pane.Show(JSON.stringify(msg.result), event.target.value)
                        });

                        (ui.magic.History.length == 0 || ui.magic.History[ui.magic.History.length-1] != event.target.value) && ui.magic.History.push(event.target.value)
                        ui.magic.Current = ui.magic.History.length
                        ui.count.innerHTML = ui.magic.Current
                        event.target.value = ""
                    } else {
                        page.oninput(event, function(event) {
                            switch (event.key) {
                                case "Enter":
                                    kit.Log(event.target.value)
                                    break
                                default:
                                    return false
                            }
                            return true
                        })
                    }
                    ui.count.innerHTML = ui.magic.Current || 0
                    field.Pane.Show()
                }]}]},
            ])

            ui.magic.History = []

            return {
                Select: function() {ui.magics.focus()},
                Size: function(width, height) {
                    kit.size(field, width, height)
                    ui && kit.size(ui.magics, (width - ui.count.offsetWidth - ui.first.offsetWidth - ui.state.offsetWidth - 40), height-6)
                },
                Show: function() {var meta = page.Status.meta, list = page.Status.list
                    kit.AppendChilds(ui.state, list.map(function(item) {return {text: [item+":"+meta[item], "div", "item"], click: function(event) {
                        cb(event, item, meta[item])
                    }}}))
                    field.Pane.Size(field.clientWidth, field.clientHeight)
                },
                Help: function() {return []},
            }
        },
        Pane: Pane,
    }), page.which = page.Sync("layout"), page.who = page.Sync("username")

    kit.Log("init", "page", page)
    return window.onload = page.onload, page
}
function Pane(page, field) {
    var option = field.querySelector("form.option")
    var action = field.querySelector("div.action")
    var output = field.querySelector("div.output")

    var timer = ""
    var name = option.dataset.names
    var itemkey = "fieldset.item, div.item"
    var itemkeys = "fieldset.item.select, div.item.select"

    var pane = Meta(page.Zone(name), field, (page[field.dataset.init] || function() {})(page, field, option, output) || {}, {
        Appends: shy("添加列表", function(cmds, type, key, which, first, cb, cbs) {
            var list = []
            pane.Runs(event, cmds, function(line, index, msg) {
                var ui = pane.Append(type, line, key, which, cb)
                list.push(ui)
                if (typeof first == "string") {
                    (line.key == first || line.name == first || line[which] == first || line[key[0]] == first) && ui.item.click()
                } else {
                    first && index == 0 && ui.item.click()
                }
                if (index == msg[msg.append[0]].length-1) {
                    kit.Selector(output, itemkeys).length == 0 && pane.Select(0)
                    kit._call(cbs, [msg, list])
                }
            })
        }),
        Append: shy("添加列表", function(type, line, key, which, cb) {type = type || line.type
            var ui = pane.View(output, type, line, key)
            if (!ui.item) {return}

            ui.item.onclick = function(event) { if (pane.which.get() == line[which]) {return}
                pane.Event(event, {}, {name: pane.Zone("select", line[key[0]])})
                page.script("record", [name, line[key[0]]])
                pane.Select(ui.item)

                pane.which.set(line[which])
                kit._call(cb, [event, line])
            }
            ui.item.ondragstart = function(event) {if (event.target != ui.item) {return}
                event.dataTransfer.setData("item", event.target.id)
                event.stopPropagation()
                event.preventDefault()
            }
            ui.item.ondragover = function(event) {if (event.target != ui.item) {return}
                event.stopPropagation()
                event.preventDefault()
                pane.Select(ui.item)
            }
            ui.item.ondrop = function(event) {if (event.target != ui.item) {return}
                var item = pane[event.dataTransfer.getData("item")]
                output.insertBefore(item, event.target)
                event.stopPropagation()
            }
            ui.item.oncontextmenu = function(event) {
                pane.Detail && page.carte.Pane.Show(event, shy({}, pane.Detail, function(event, value, meta) {
                    pane.Check(event, value, meta, line)
                }))
            }

            if (type == "plugin" && line.name || type == "field") {
                page.Require(line.init? line.group+"/"+line.init: "", function(init) {
                    page.Require(line.view? line.group+"/"+line.view: "", function(view) {
                        can.user.carte = page.carte.Pane.Show;
                        can.user.toast = page.toast.Pane.Show;
                        (
                            line.group == "docker" ||
                            line.group == "story" ||
                            line.group == "favor"
                        ) && can.Plugin(can.action||can, ui.item.Meta, "state", function(event, cmds, cbs) {
                            var msg = ctx.Event(event)
                            kit.Selector(action, "input", function(item, index) {
                                item.value && msg.Option(item.name, item.value)
                            })

                            ctx.Run(event, option.dataset, [line.river, line.storm, line.action].concat(cmds), cbs)
                        }, ui.item) ||
                        pane.Plugin(page, pane, ui.item, init, function(event, cmds, cbs) {
                            kit._call(cb||pane.Core, [event, line, cmds, cbs])
                        })
                    })
                })
            }
            return pane[ui.item.id] = ui.item, ui
        }),
        Select: shy("选择列表", function(index) {
            kit.Selector(output, itemkey, function(item, i) {if (item != index && i != index) {return}
                kit.Selector(output, itemkeys, function(item) {kit.classList.del(item, "select")})
                kit.classList.add(item, "select")
            })
        }),

        Next: function() {
            var next = output.querySelector(itemkeys).nextSibling
            next? next.click(): output.firstChild.click()
        },
        Prev: function() {
            var prev = output.querySelector(itemkeys).previousSibling
            prev? prev.click(): output.lastChild.click()
        },

        Show: function() {
            kit.ModifyView(field, {display: "block"})
        },
        Hide: function() {
            kit.ModifyView(field, {display: "none"})
        },
        Size: function(width, height) {
            if (width > 0) {
                field.style.width = width+"px"
                field.style.display = "block"
            } else if (width === "") {
                field.style.width = ""
                field.style.display = "block"
            } else {
                field.style.display = "none"
                return
            }

            if (height > 0) {
                field.style.height = height+"px"
                field.style.display = "block"
            } else if (height === "") {
                field.style.height = ""
                field.style.display = "block"
            } else {
                field.style.display = "none"
                return
            }
        },
        Dialog: function(width, height) {
            if (field.style.display != "block") {
                page.dialog && page.dialog != field && page.dialog.style.display == "block" && page.dialog.Show()
                page.dialog = field, field.style.display = "block", kit.ModifyView(field, {window: [width||80, height||200]})
                return true
            }
            field.style.display = "none"
            delete(page.dialog)
            return false
        },

        Help: function(type, action) {
            var text = []
            switch (type) {
                case "name":
                case undefined:
                    text = [name]
                    break
                case "list":
                    var list = []
                    for (var k in pane) {list.push(k)}
                    list.sort(), text = text.concat(list.map(function(item) {return "func: "+item+"\n"}))

                    var list = []
                    for (var k in pane.Action) {list.push(k)}
                    list.sort(), text = text.concat(list.map(function(item) {return "action: "+item+"\n"}))
                    break
            }
            return text
        },
        Jshy: shy("本地命令", function(event, args) {var msg = ctx.Event(event)
            if (kit.isNone(args) || args.length == 0) {
                return kit.classList.has(field, "dialog") && pane.Show() || pane.Zone("show", args[0])
            }

            // 插件命令
            if (pane[args[0]] && pane[args[0]].type == "fieldset") {
                return pane[args[0]].Plugin.Jshy(event, args.slice(1)) || pane.Zone("plugin", args[0])
            }
            // 控件命令
            if (typeof pane.Action[args[0]] == "function") {
                return kit._call(pane.Action[args[0]], [event, args[0]]) || pane.Zone("action", args[0])
            }
            // 列表命令
            var list = kit.Selector(output, itemkey, function(item) {
                if (item.Meta.key == args[0] || item.Meta.name == args[0]) {return item}
            })
            if (list.length > 0) {list[0].click(); return pane.Zone("select", args[0])}

            // 内部命令
            if (typeof pane[args[0]] == "function") {
                return kit._call(pane[args[0]], args.slice(1)) || pane.Zone("function", args[0])
            }
        }),

        Tickers: shy("定时刷新", function(time, cmds, cb) {
            pane.Ticker(time, cmds, function(msg) {msg.Table(function(line, index) {
                cb(line, index, msg)
            })})
        }),
        Ticker: shy("定时刷新", function(time, cmds, cb) {timer && clearTimeout(timer)
            function loop() {
                event = document.createEvent("Event")
                pane.Event(event, {}, {name: pane.Zone("ticker")})
                !pane.Stop() && pane.Run(event, cmds, function(msg) {
                cb(msg), timer = setTimeout(loop, time)
            })}
            time && (timer = setTimeout(loop, 10))
        }),
        Check: shy("执行操作", function (event, value) {
            // Event入口 1.1
            pane.Event(event, {}, {name: pane.Zone("click", value)})
            page.script("record", [name, value])

            var cb = pane.Action.option || function() {}
            if (pane.Action && pane.Action.meta && typeof pane.Action.meta[value] == "function") {
                cb = pane.Action.meta[value]
            } else if (pane.Action && typeof pane.Action[value] == "function") {
                cb = pane.Action[value]
            } else if (typeof pane.Action == "function") {
                cb = pane.Action
            } else if (typeof pane[value] == "function") {
                cb = pane[value]
            }

            if (page.Action && page.Action.meta && typeof page.Action.meta[value] == "function") {
                cb = page.Action.meta[value]
            } else if (page.Action && typeof page.Action[value] == "function") {
                cb = page.Action[value]
            } else if (typeof page.Action == "function") {
                cb = page.Action
            } else if (typeof page[value] == "function") {
                cb = page[value]
            }
            kit._call(cb, arguments)
        }),
        Tutor: function() {var pane = field.Pane
            var event = window.event
            function loop(list, index) {
                if (index >= list.length) {return}
                kit.Log(index, list[index])
                pane.Core(event, {}, ["_cmd", list[index]])
                setTimeout(function() {loop(list, index+1)}, 1000)
            }
            loop([
                "聊天", "help", "最高", "最大", "聊天",
                "工作", "串行", "清空", "并行", "help storm", "help storm list", "help action", "help action list",
                "聊天", "help target", "help target list",
            ], 0)
        },
        Core: shy("命令分发", {
            wssid: function(id) {return id && (page.wssid = id)},
        }, function(event, line, args, cbs) {var msg = pane.Event(event), meta = arguments.callee.meta
            if (kit.isNone(args)) {return}
            var plugin = event.Plugin || page.plugin && page.plugin.Plugin || {}, engine = {
                share: function(args) {
                    return ctx.Share({"group": option.dataset.group, "names": option.dataset.names, "cmds": [
                        line.river, line.storm, line.action,  args[1]||"",
                    ]})
                },
                pwd: function(name, value) {
                    name && kit.Selector(page.action, "fieldset.item."+name, function(item) {
                        item.Plugin.Select()
                    })
                    if (value) {return engine.set(value)}
                    return [page.river.Pane.which.get(), page.storm.Pane.which.get(), page.plugin && page.plugin.Meta.name, page.input && page.input.name, page.input && page.input.value]
                },
                set: function(value, name) {
                    try {
                        if (value == undefined) {
                            msg.append = ["name", "value"]
                            msg.name = [], msg.value = []
                            return kit.Selector(page.plugin, ".args", function(item) {
                                msg.Push("name", item.name)
                                msg.Push("value", item.value)
                                return item.name+":"+item.value
                            })

                        } else if (name == undefined) {
                            kit.Selector(page.plugin, "input[type=button]", function(item) {
                                if (item.value == value) {item.click(); return value}
                            }).length > 0 || (page.action.Pane.Action[value]?
                                page.action.Pane.Action[value](event, value): (page.input.value = value))
                        } else {
                            page.plugin.Plugin.Inputs[name].value = value
                        }
                    } catch (e) {
                        engine._cmd("_cmd", [value, name])
                    }
                },
                dir: function(rid, sid, pid, uid) {
                    if (!rid) {
                        return kit.Selector(page.river, "div.output>div.item>div.text>span", function(item) {
                            return item.innerText
                        })
                    }
                    if (!sid) {
                        return kit.Selector(page.storm, "div.output>div.item>div.text>span", function(item) {
                            return item.innerText
                        })
                    }
                    if (!pid) {
                        return kit.Selector(page.action, "fieldset.item>legend", function(item) {
                            msg.Push("name", item.parentNode.Meta.name)
                            msg.Push("help", item.parentNode.Meta.help)
                            return item.innerText
                        })
                    }
                    if (!uid) {
                        return kit.Selector(page.plugin, "input", function(item) {
                            msg.Push("name", item.name)
                            msg.Push("value", item.value)
                            return item.name+":"+item.value
                        })
                    }
                    return [river, storm, page.plugin && page.plugin.Meta.name, page.input && page.input.name]
                },
                echo: function(one, two) {
                    kit.Log(one, two)
                },
                helps: function() {
                    engine.help("river")
                    engine.help("action")
                    engine.help("storm")
                },
                help: function() {
                    var args = kit.List(arguments), cb, target
                    if (args.length > 0 && page.pane && page.pane.Pane[args[0]] && page.pane.Pane[args[0]].Plugin) {
                        cb = page.pane.Pane[args[0]].Plugin.Help, target = page.pane.Pane[args[0]], args = args.slice(1)
                    } else if (args.length > 1 && page[args[0]] && page[args[0]].Pane[args[1]]) {
                        cb = page[args[0]].Pane[args[1]].Plugin.Help, target = page[args[0]].Pane[args[1]], args = args.slice(2)
                    } else if (args.length > 0 && page[args[0]]) {
                        cb = page[args[0]].Pane.Help, target = page[args[0]], args = args.slice(1)
                    } else {
                        cb = page.Help, target = document.body, args
                    }

                    if (kit.Selector(target, "div.Help", function(help) {
                        target.removeChild(help)
                        return help
                    }).length > 0) {return}

                    var text = kit._call(cb, args)
                    var ui = kit.AppendChild(target, [{view: ["Help"], list: [{text: [text.join(""), "div"]}]}])
                    setTimeout(function() {target.removeChild(ui.last)}, 30000)
                },
                _split: function(str) {return str.trim().split(" ")},
                _cmd: function(arg) {
                    var args = typeof arg[1] == "string"? engine._split(arg[1]): arg[1];
                    page.script("record", args)

                    if (typeof meta[args[0]] == "function") {
                        return kit._call(meta[args[0]], args.slice(1))
                    }
                    if (typeof engine[args[0]] == "function") {
                        return kit._call(engine[args[0]], args.slice(1))
                    }

                    // if (page.output && typeof page.output.Output[args[0]] == "function") {
                    //     return kit._call(page.output.Output[args[0]], args.slice(1))
                    // }
                    // if (page.input && typeof page.input.Input[args[0]] == "function") {
                    //     return kit._call(page.input.Input[args[0]], args.slice(1))
                    // }
                    if (page.plugin && typeof page.plugin.Plugin[args[0]] == "function") {
                        return kit._call(page.plugin.Plugin[args[0]], args.slice(1))
                    }

                    if (page.dialog && (res = page.dialog.Pane.Jshy(event, args))) {return res}
                    if (page.pane && (res = page.pane.Pane.Jshy(event, args))) {return res}
                    if (page.action && (res = page.action.Pane.Jshy(event, args))) {return res}
                    if (page.storm && (res = page.storm.Pane.Jshy(event, args))) {return res}
                    if (page.river && (res = page.river.Pane.Jshy(event, args))) {return res}

                    if (page && (res = page.Jshy(event, args))) {return res}

                    if (page.plugin && (res = page.plugin.Plugin.Jshy(event, args))) {return res}
                    return kit.Log(["warn", "not", "find"].concat(args))
                },
                _msg: function(msg) {
                    event.ctrlKey? kit._call(page.target.Pane.Send, msg.Format()):
                        event.shiftKey && page.target.Pane.Send("field", plugin.Reveal())
                },
                _run: function(msg) {
                    pane.Run(event, [Meta.river, Meta.storm, Meta.action].concat(args), function(msg) {
                        kit._call(cbs, [msg]), engine._msg(msg)
                    })
                },
            }

            var Meta = plugin && plugin.target && plugin.target.Meta || line || {}
            kit.Log(["cmd"].concat(kit.List([Meta.river, Meta.storm, Meta.action])).concat(args[0] == "_cmd"? args[1]: args))

            page.Status("ncmd", kit.History("cmd", -1, {args: args, meta: Meta}))
            return args.length > 0 && meta[args[0]]? kit._call(cbs, [msg.Echo(meta[args[0]](args))]):
                    args.length > 0 && engine[args[0]]? kit._call(cbs, [msg.Echo(engine[args[0]](args))]):
                        event.shiftKey? engine._msg(msg): engine._run(msg)
        }),
        Runs: shy("执行命令", function(event, cmds, cb) {
            pane.Run(event, cmds, function(msg) {msg.Table(function(line, index) {
                (cb||pane.ondaemon)(line, index, msg)
            })})
        }),
        Run: shy("执行命令", function(event, cmds, cb) {
            var msg = pane.Event(event, null, {name: pane.Zone(cmds[0])})
            kit.Selector(action, "input", function(item, index) {
                msg.Option(item.name, item.value)
            })
            ctx.Run(event, option.dataset, cmds, cb||pane.ondaemon)
        }),

        onchoice: shy("菜单列表", {
            "删除": "_table",
            "修改": "_canvas",
        }, ["删除", "修改"], function(event, value, meta) {
            return pane._call(meta[value], event)
        }),
        onaction: shy("事件列表", {
            oncontextmenu: function(event) {
                page.carte.Pane.Show(event, pane.Choice? shy({}, pane.Choice, function(event, value, meta) {
                    pane.Check(event, value)
                }): pane.onchoice)
            },
        }, function(event, key, cb) {cb(event)}),
        which: page.Sync(name), Listen: {}, Action: {}, Button: [], Choice: [],
        Plugin: Plugin,
    })

    kit.Log("init", "pane", name, pane)
    kit.AppendAction(action, pane.Button, pane.Check)
    kit.Item(pane.Listen, function(key, cb) {page.Sync(key).change(cb)})
    return page[name] = field, field.Pane = pane
}

function Plugin(page, pane, field, inits, runs) {
    var option = field.querySelector("form.option")
    var action = field.querySelector("div.action")
    var output = field.querySelector("div.output")

    var meta = field.Meta
    var name = meta.name, args = meta.args || []
    var inputs = JSON.parse(meta.inputs || "[]")
    var feature = JSON.parse(meta.feature||'{}')
    kit.classList.add(field, meta.group, name, feature.style)

    var plugin = Meta(pane.Zone(name), field, inits && inits(field, option, output) || {}, {Inputs: {}, Outputs: {},
        Appends: shy("添加控件", function(event, inputs) {
            if (inputs) {return inputs.map(function(item) {plugin.Append(item)})}

            var name = "args"+kit.Selector(option, "input.args.temp").length
            plugin.Append({type: "text", name: name, className: "args temp"}).focus()
        }),
        Append: shy("添加控件", function(item, name, value) {
            var count = kit.Selector(option, ".args").length
            args && count < args.length && (value = value || args[count] || "")

            item.before && kit.AppendChild(option, item.before)
            item && item.width && (item.style = {}, item.style.width = item.width)
            item.view == "full" && kit.AppendChild(option, "br")
            var input = {plug: meta.name, type: "input", name: name || item.name || item.value || "input", data: item}
            switch (item.type) {
                case "upfile": item.type = "file"; break
                case "select":
                    item.values = kit.Trans(item.values)
                    input.type = "select", input.list = item.values.map(function(value) {
                        return {type: "option", value: value, inner: value}
                    })
                    item.value = value || item.value || item.values[0]
                    kit.classList.add(item, "args")
                    break
                case "textarea":
                    var half = parseFloat(item.half||"1")||1
                    input.type = "textarea", item.style = "height:"+(item.height||"50px")+";width:"+parseInt(((pane.target.clientWidth-35)/half))+"px"
                    // no break
                case "text":
                    item.value = value || item.value || ""
                    kit.classList.add(item, "args")
                    item.autocomplete = "off"
                    break
            }
            kit.List((item.clist||"").split(" "), function(value) {
                kit.classList.add(item, value)
            })
            kit.classList.add(item, item.view)
            return Inputs(plugin, input, item, plugin.View(option, "input", input)[input.name], option).target
        }),
        Select: shy("选择控件", function(target, focus) {field.onclick(event)
            page.plugin = field, page.input = target = target || option.querySelectorAll("input")[1]
            plugin.which.set(page.input.name)
            focus && page.input.focus()
        }),
        Remove: shy("删除控件", function() {
            var list = option.querySelectorAll("input.temp")
            list.length > 0 && (option.removeChild(list[list.length-1].parentNode))
        }),
        Rename: shy("命名", function() {
            kit.prompt("控件名称", function(name) {
                meta["help"] = name
                kit.Selector(field, "legend", function(legend) {
                    legend.innerHTML = meta.name+"("+meta.help+")"
                })
            })
        }),

        Delete: shy("删除插件", function() {
            plugin.Prev().Plugin.Select(), field.parentNode.removeChild(field)
        }),
        Reveal: shy("导出插件", function() {
            field.Meta.args = arguments.length > 0? kit.List(arguments):
                kit.Selector(option, ".args", function(item) {return item.value})
            return JSON.stringify(field.Meta)
        }),
        Clone: shy("复制插件", function() {
            return pane.Append("field", {text: plugin.Reveal(), init: meta.init, view: meta.view, group: meta.group}, [], "").item.Plugin.Select(null, true)
        }),
        clear: function() {
            plugin.Save(""), output.innerHTML = ""
        },

        Move: function() {
            return field.nextSibling || field.parentNode.firstChild
        },
        Next: function() {
            return field.nextSibling || field.parentNode.firstChild
        },
        Prev: function() {
            return field.previousSibling || field.parentNode.lastChild
        },

        Help: function(type, action) {
            var text = []
            switch (type) {
                case "name":
                case undefined:
                    text = [meta.name]
                    break
                case "list":
                    var list = []
                    for (var k in plugin) {list.push(k)}
                    list.sort(), text = text.concat(list.map(function(item) {return "func: "+item+"\n"}))

                    var list = []
                    for (var k in plugin.ondaemon) {list.push(k)}
                    list.sort(), text = text.concat(list.map(function(item) {return "daemon: "+item+"\n"}))

                    var list = []
                    for (var k in plugin.onexport) {list.push(k)}
                    list.sort(), text = text.concat(list.map(function(item) {return "export: "+item+"\n"}))

                    var list = []
                    for (var k in plugin.onaction) {list.push(k)}
                    list.sort(), text = text.concat(list.map(function(item) {return "action: "+item+"\n"}))
                    break
            }
            return text
        },
        Jshy: function(event, args) {
            plugin.Select(), field.scrollIntoView()
            // 显示命令
            if (plugin.Outputs[args[0]]) {
                return plugin.Outputs[args[0]].Output.Jshy(event, args.slice(1)) || plugin.Zone("output", args[0])
            }
            // 控件命令
            if (plugin.Inputs[args[0]]) {
                return plugin.Inputs[args[0]].Input.Jshy(event, args.slice(1)) || plugin.Zone("input", args[0])
            }
            // 内部命令
            if (typeof plugin[args[0]] == "function") {
                return kit._call(plugin[args[0]], args.slice(1)) || plugin.Zone("function", args[0])
            }
            // 远程命令
            return kit._call(plugin.Runs, [event]) || plugin.Zone("remote", args[0])
        },

        ontoast: function() {kit._call(page.toast.Pane.Show, arguments)},
        oncarte: function() {kit._call(page.carte.Pane.Show, arguments)},

        Option: function(key, value) {
            kit.notNone(value) && option[key] && (option[key].value = value)
            return kit.notNone(key)? (option[key]? option[key].value: ""):
                kit.Selector(option, ".args", function(item, index) {return item.value})
        },
        upload: function(event) {
            ctx.Upload({river: meta.river, table: plugin.Option("table")}, option.upload.files[0], function(event, msg) {
                Output(plugin, "table", msg, null, output, option)
                plugin.ontoast("上传成功")
            }, function(event) {
                plugin.ontoast("上传进度 "+parseInt(event.loaded*100/event.total)+"%")
            })
        },
        Delay: shy("延时命令", function(time, event, text) {plugin.ontoast(text, "", -1)
            return setTimeout(function() {plugin.Runs(event)}, time)
        }),
        Check: shy("检查命令", function(event, target, cb) {
            kit.Selector(option, ".args", function(item, index, list) {
                kit.isNone(target)? index == list.length-1 && plugin.Runs(event, cb):
                    item == target && (index == list.length-1? plugin.Runs(event, cb): page.plugin == field && list[index+1].focus())
                return item
            }).length == 0 && plugin.Runs(event, cb)
        }),
        Last: shy("历史命令", function(event) {kit.notNone(plugin.History()) && plugin.Check(event)}),
        Runs: shy("执行命令", function(event, cb) {plugin.Run(event, plugin.Option(), cb)}),
        Run: shy("执行命令", function(event, args, cb, silent) {var show = true
            page.script("record", ["action", name].concat(args))
            setTimeout(function() {show && plugin.ontoast(kit.Format(args||["running..."]), meta.name, -1)}, 1000)
            event.Plugin = plugin, runs(event, args, function(msg) {
                silent? kit._call(cb, [msg]): plugin.ondaemon(msg, cb), show = false, plugin.ontoast()
            })
        }),

        ondaemon: shy("接收数据", function(msg, cb) {
            plugin.msg = msg, plugin.Save(""), plugin.onfigure(plugin.onfigure.meta.type, msg, cb)
        }),
        onfigure: shy("显示数据", {type: feature.display||"table",
            max: function(output) {
                output.style.maxWidth = pane.target.clientWidth-30+"px"
                output.style.maxHeight = pane.target.clientHeight-60+"px"
            },
            size: function(cb) {
                kit._call(cb, [pane.target.clientWidth, pane.target.clientHeight])
            },
        }, function(type, msg, cb) {var meta = arguments.callee.meta
            type != meta.type && plugin.Save(meta.type, output), meta.type = type
            if (type == "chart") {
                can.chat.Output(can, plugin, type || feature.display, msg || plugin.msg, cb, output, option)
                return
            } 
            !plugin.Load(type, output) && (type=="chart"? can.chat.Output: Output)(plugin, type || feature.display, msg || plugin.msg, cb, output, option)
        }),
        onexport: shy("导出数据", {
            "": function(value, name, line) {
                return value
            },
            see: function(value, name, line) {
                return value.split("/")[0]
            },
            you: function(value, name, line) {
                event.Plugin = plugin

                line.you && name == "status" && (line.status == "start"? function() {
                    plugin.Delay(3000, event, line.you+" stop...") && plugin.Run(event, [option.pod.value, line.you, "stop"])
                }(): plugin.Run(event, [option.pod.value, line.you], function(msg) {
                    plugin.Delay(3000, event, line.you+" start...")
                }))
                return name == "status" || line.status == "stop" ? undefined: line.you
            },
            pod: function(value, name, line, list) {
                return (option[list[0]].value? option[list[0]].value+".": "")+value
            },
            dir: function(value, name, line) {
                name != "path" && (value = line.path)
                return value
            },
            tip: function(value, name, line) {
                return option.tip.value + value
            },
        }, JSON.parse(meta.exports||'["",""]'), function(event, value, name, line) {
            var meta = arguments.callee.meta
            var list = arguments.callee.list

            for (var i = 0; i < list.length; i += 3) {if (list[i+1] == name || list[i+2]) {
                plugin.ontoast(list[i+1]? line[list[i+1]]: value, list[i+1]||name)
                for (var i = 0; i < list.length; i += 3) {
                    page.Sync("plugin_"+list[i]).set(meta[list[i+2]||""](list[i+1]? line[list[i+1]]: value, list[i+1]||name, line, list))
                }
                break
            }}
        }),
        onchoice: shy("菜单列表", {
            "返回": "Last",
            "清空": "clear",
            "克隆": "Clone",
            "重命名": "Rename",
            "删除": "Delete",
        }, ["返回", "清空", "重命名", "克隆", "删除"], function(event, value, meta) {
            return plugin._call(meta[value], event)
        }),
        onaction: shy("事件列表", {
            oncontextmenu: function(event) {
                plugin.oncarte(event, plugin.onchoice)
            },
        }, function(event, key, cb) {cb(event)}),
    })

    kit.Log("init", "plugin", name, plugin)
    plugin.which = plugin.Sync("input")
    page[field.id] = pane[field.id] = pane[name] = field, field.Plugin = plugin
    plugin.Appends(null, inputs)
    return plugin
}
function Inputs(plugin, meta, item, target, option) {
    var plug = meta.plug, name = meta.name, type = item.type
    var input = Meta(plugin.Zone(name), target, item, {
        getLocation: function(event) {
            var x = parseFloat(option.x.value)
            var y = parseFloat(option.y.value)
            page.getLocation && page.getLocation(function(res) {
                plugin.ondaemon({
                    append: ["longitude", "latitude", "accuracy", "speed"],
                    longitude: [res.longitude+x+""],
                    latitude: [res.latitude+y+""],
                    accuracy: [res.accuracy+""],
                    speed: [res.speed+""],
                })
            })
        },
        openLocation: function(event) {
            var x = parseFloat(option.x.value)
            var y = parseFloat(option.y.value)
            page.getLocation && page.getLocation(function(res) {
                page.openLocation && page.openLocation(res.latitude+y, res.longitude+x, option.pos.value)
            })
        },
        clear: function() {
            target.value = ""
        },

        Jshy: function(event, args) {
            // 内部命令
            if (typeof plugin[args[0]] == "function") {
                return kit._call(plugin[args[0]], args.slice(1)) || plugin.Zone("function", args[0])
            }
            // 内部命令
            return (target.value = args[0]) || plugin.Zone("value", args[0])
        },

        onformat: shy("数据转换", {
            none: function(value) {return value||""},
            date: function(value) {return kit.time()},
            month: function(value) {return kit.time().slice(0, 7)},
        }, function(which, value) {var meta = arguments.callee.meta
            return (meta[which||"none"]||meta["none"])(value)
        }),
        onimport: shy("导入数据", {}, kit.Trans(item.imports), function() {
            kit.List(arguments.callee.list, function(imports) {
                page.Sync(imports).change(function(value) {
                    plugin.History(target.value, target), target.value = value.trim()
                    input.Event(event = document.createEvent("Event"))
                    item.action == "auto" && plugin.Runs(event)
                })
            }), item.type == "button" && item.action == "auto" && target.click()
        }),
        onchoice: shy("菜单列表", {
            "返回": "Last",
            "清空": "clear",
            "加参": "Appends",
            "减参": "Remove",
        }, ["返回", "清空", "加参", "减参"], function(event, value, meta) {
            return input._call(meta[value], event)
        }),
        onaction: shy("事件列表", {
            oncontextmenu: function(event) {
                item.type != "button" && plugin.oncarte(event, input.onchoice)
            },
            onfocus: function(event) {plugin.Select(target)},
            onblur: function(event) {type == "text" && input.which.set(target.value)},
            onclick: function(event) {plugin.Select()
                // Event入口 2.0
                type == "button" && input.Event(event, {}) && kit.Value(input[item.cb], plugin[item.cb], function(event, target) {
                    plugin.Check(event)
                })(event, target)
            },
            onchange: function(event) {
                // Event入口 2.1
               type == "select" && input.Event(event, {}) && plugin.History(target.value, target) && plugin.Check(event, item.action == "auto"? undefined: target)
            },
            ondblclick: function(event) {
                var txt = kit.History("txt", -1);
                (type == "text" || type == "textarea") && txt && (target.value = txt.data.trim())
            },
            onselect: function(event) {
                kit.CopyText()
            },
            onkeydown: function(event) {
                switch (event.key) {
                    case " ":
                        event.stopPropagation()
                        return true
                }

                page.oninput(event, function(event) {
                    switch (event.key) {
                        case " ":
                        case "w":
                            break
                        case "p":
                            action.Last()
                            break
                        case "i":
                            plugin.Next().Plugin.Select()
                            break
                        case "o":
                            plugin.Prev().Plugin.Select()
                            break
                        case "c":
                            plugin.clear()
                            break
                        case "r":
                            plugin.clear()
                        case "j":
                            plugin.Runs(event)
                            break
                        case "l":
                            target.scrollTo(0, plugin.target.offsetTop)
                            break
                        case "b":
                            plugin.Appends(event)
                            break
                        case "m":
                            plugin.Clone()
                            break
                        default:
                            return false
                    }
                    return true
                })

                // Event入口 2.1
                if (event.key == "Enter" && (event.ctrlKey || item.type == "text")) {
                    input.which.set(target.value) != undefined && plugin.History(target.value, target)
                    input.Event(event, {}) && event.ctrlKey? plugin.Runs(event): plugin.Check(event, target)
                }
            },
            onkeyup: function(event) {
                switch (event.key) {
                    case " ":
                        event.stopPropagation()
                        return true
                }

                page.oninput(event, function(event) {
                    switch (event.key) {
                        case " ":
                        case "w":
                            break
                        default:
                            return false
                    }
                    return true
                })
            },
        }, function(event, key, cb) {cb(event)}),
        which: plugin.Sync(name),
    }, plugin)

    kit.Log("init", "input", input.Zones(), input)

    input.onimport()
    target.value = input.onformat(item.init, item.value);
    (type == "text" || type == "textarea") && !target.placeholder && (target.placeholder = item.name)
    type == "text" && !target.title && (target.title = item.placeholder || item.name || "")
    return plugin.Inputs[item.name] = target, target.Input = input
}
function Output(plugin, type, msg, cb, target, option) {
    var name = plugin.target.Meta.name+"."+type
    var feature = JSON.parse(plugin.target.Meta.feature||'{}')
    var exports = JSON.parse(plugin.target.Meta.exports||'{}')

    var output = Meta(plugin.Zone(type), target, {
        Format: function() {
            var ext = ".csv", txt = kit.Selector(target, "tr", function(tr) {
                return kit.Selector(tr, "td,th", function(td) {
                    return td.innerText
                }).join(",")
            }).join("\n");
            type == "editor" && msg.file && (ext = ".txt", txt = msg.result.join("\n"));
            !txt && (ext = ".txt", txt = msg.result.join(""))
            return [name, ext, txt]
        },
        Download: function() {
            var ps = output.Format()

            plugin.ontoast({text:'<a href="'+URL.createObjectURL(new Blob([ps[2]]))+'" target="_blank" download="'+ps[0]+ps[1]+'">'+ps[0]+ps[1]+'</a>', title: "下载中...", width: 200})
            kit.Selector(page.toast, "a", function(item) {item.click()})
        },
        Copy: function(event) {
            kit.CopyText(output.Format()[2])
        },

        Jshy: function(event, args) {
            // 内部命令
            if (typeof plugin[args[0]] == "function") {
                return kit._call(plugin[args[0]], args.slice(1)) || plugin.Zone("function", args[0])
            }
            // 内部命令
            return (target.value = args[0]) || plugin.Zone("value", args[0])
        },

        size: function(cb, ncol, nrow) {ncol = ncol || 1, nrow = nrow || 1
            plugin.onfigure.meta.size(function(width, height) {
                cb(width/ncol-10, height/nrow-10)
            })
        },
        onimport: shy("导入数据", {
            _video: function(msg, line, width, height) {
                kit.AppendChild(target, [{type: "video", style: {width: width},
                    data: {controls: "controls", autoplay: false, loop: true, src: line.file}}])
            },
            _image: function(msg, line, width, height) {
                kit.AppendChild(target, [{type: "img", data: {src: line.file}, style: {width: width}}])
            },
            _table: function(msg, list) {
                return list && list.length > 0 && kit.OrderTable(kit.AppendTable(kit.AppendChild(target, "table"), msg.Table(), list), "", output.onexport, function(event, value, name, line, index) {
                    var td = event.target
                    plugin.oncarte(event, shy("菜单列表", {
                            "选择": "select",
                            "修改": "modify",
                            "复制": function(event, text) {kit.CopyText(text), plugin.ontoast(text, "Copy to ClipBoard!")},
                            "下载": "Download",
                            "删除": "delete",
                        }, feature.detail||["选择", "修改", "复制", "下载", "删除"], function(event, item, meta) {
                            var text = td.innerText.trim()
                            if (typeof meta[item] == "function") {meta[item](event, text); return}

                            // 选项
                            var msg = plugin.Event(event)
                            kit.Selector(option, ".args", function(item) {
                                msg.Option(item.name, item.value||"")
                            })

                            // 参数
                            var cmd = [], id = ""
                            option.dream && cmd.push(option.dream.value)
                            option.table && cmd.push(option.table.value)
                            if (td.dataset && td.dataset.id) {
                                id = td.dataset.id
                            } else if (name == "value") {
                                id = option.index.value, name = line.key
                            } else {
                                for (var i = 0; i < exports.length-1; i += 3) {
                                    line[exports[i+1]] && msg.Option(exports[i], line[exports[i+1]].trim())
                                    id = (id || line[exports[i+1]] || "").trim()
                                }
                            }
                            cmd.push(id, meta[item]||item)

                            // 命令
                            if (item == "修改") {
                                // 交互命令
                                text = kit.AppendChilds(td, [{type: "input", value: text, style: {width: td.clientWidth+"px"}, data: {onkeydown: function(event) {
                                    if (event.key != "Enter") {return}

                                    // 选项
                                    var msg = plugin.Event(event)
                                    kit.Selector(option, ".args", function(item) {
                                        msg.Option(item.name, item.value)
                                    })

                                    // 后端命令
                                    cmd.push(name, event.target.value)
                                    plugin.Run(event, cmd, function(msg) {
                                        td.innerHTML = event.target.value
                                        plugin.ontoast("修改成功")
                                    }, true)
                                }}}]).input, text.focus(), text.setSelectionRange(0, -1)

                            } else if (output[meta[item]]) {
                                // 前端命令
                                output[meta[item]](event)

                            } else {
                                // 后端命令
                                cmd.push(name, text)
                                plugin.Run(event, cmd, function(msg) {
                                    console.log(msg)
                                })
                            }
                            return true
                        })
                    )
                })
            },
            _code: function(msg, line) {
                return msg.result && msg.result.length > 0 && kit.OrderCode(kit.AppendChild(target, [{view: ["code", "div", msg.Results()]}]).first)
            },
            _svg: function(width, height, src) {
                kit.AppendChild(target, '<embed src="'+src+'" width="'+width+'" height="'+height+'" type="image/svg+xml" pluginspage="http://www.adobe.com/svg/viewer/install/" />')
            },
            _svg2: function(msg, width, height) {
                target.innerHTML = msg.result.join("")
            },

            svg: function(msg, cb) {var meta = output.onimport.meta
                plugin.onfigure.meta.size(function(width, height) {
                    meta._svg2(msg, width, height*1.6)
                })
            },
            code: function(msg, cb) {var meta = output.onimport.meta
                target.innerHTML = "", output.onimport.meta._code(msg)
                typeof cb == "function" && cb(msg)
            },
            table: function(msg, cb) {var meta = output.onimport.meta
                target.innerHTML = ""
                output.onimport.meta._table(msg, msg.append) || output.onimport.meta._code(msg)
                typeof cb == "function" && cb(msg)
            },
            album: function(msg, cb) {var meta = output.onimport.meta
                target.innerHTML = "", msg.Table(function(line) {meta._image(msg, line)})
            },
            video: function(msg, cb) {var meta = output.onimport.meta
                target.innerHTML = "", msg.Table(function(line) {meta._video(msg, line)})
            },
            audio: function(msg, cb) {var meta = output.onimport.meta
                target.innerHTML = "", msg.Table(function(line) {meta._audio(msg, line)})
            },
            media: function(msg, cb) {var meta = output.onimport.meta
                output.size(function(width, height) {
                    target.innerHTML = "", msg.Table(function(line) {var ls = (line.file||line.path).split(".")
                        line.file = line.file? line.file: "/download/"+line.path
                        switch (ls[ls.length-1].toLowerCase()) {
                            case "mp4": meta._video(msg, line, width, height); break
                            case "png": meta._image(msg, line, width, height); break
                            case "svg": meta._svg(msg, line, width, height); break
                        }
                    })
                }, 3)
            },
            inner: function(msg, cb) {var meta = output.onimport.meta
                target.innerHTML = "", plugin.onfigure.meta.max(target)
                output.onimport.meta._table(msg, msg.append) || kit.OrderCode(kit.ModifyNode(target, msg.result.join("")))
                kit._call(cb, [msg])
                kit.Selector(target, ".story", function(item) {
                    var data = item.dataset
                    item.oncontextmenu = function(event) {
                        plugin.oncarte(event, shy("", {
                            "提交": function(event) {
                                plugin.Run(event, [option.dream.value, "commit", option.story.value, data.scene, data.enjoy, data.happy], function(msg) {
                                    plugin.ontoast(msg.Result())
                                }, true)
                            },
                            "复制": function(event) {
                                plugin.ontoast(kit.CopyText(data.happy))
                            },
                            "收藏": function(event) {
                                kit.prompt("收藏到", function(table) {
                                    plugin.Run(event, [option.dream.value, "favor", table, option.story.value, data.scene, data.enjoy, data.happy], function(msg) {
                                    }, true)
                                })
                            },
                            "共享": function(event) {
                                plugin.Run(event, [option.dream.value, "share", option.story.value, data.scene, data.enjoy, data.happy], function(msg) {
                                }, true)
                            },
                        }, ["提交", "复制", "收藏", "共享"], function(event, value, meta) {
                            meta[value](event)
                        }))
                    }
                })
            },
            editor: function(msg, cb) {var meta = output.onimport.meta
                output.onimport.meta.table(msg, cb)

                var current = page.Sync("plugin_editor_index").get()
                target.style.maxHeight = ""
                if (msg.file) {
                    target.style.maxHeight = "500px"
                    var code = kit.AppendChild(target, [{view: ["code", "table"], list: kit.List(msg.result, function(line, index) {
                        return {view: ["line"+ (current == index+1? " select": ""), "tr"],
                        list: [{view: ["number", "td", index+1]}, {view: ["code", "td", kit.Origin(kit.Color(line))], style: {"white-space": "pre"}}], click: function(event) {
                            page.Sync("plugin_editor_file").set(msg.file[0])
                            page.Sync("plugin_editor_index").set(index+1)
                            page.Sync("plugin_editor_line").set(line)
                            page.Sync("plugin_editor_word").set(kit.CopyText())
                        }}
                    })}])
                    kit.Selector(target, "tr.line.select", function(tr) {
                        tr.scrollIntoView()
                    })
                }
            },
            canvas: function(msg, cb) {var meta = output.onimport.meta
                target.innerHTML = "", plugin.onfigure.meta.size(function(width, height) {
                    Canvas(plugin, option, target, width-45, height-175, 10, msg)
                })
            },
        }, function(type, msg, cb) {var meta = output.onimport.meta
            page.output = target, meta[type](msg, cb)
        }),
        ondetail: shy("菜单列表", {
            "删除": "_table",
            "修改": "_canvas",
        }, ["删除", "修改"], function(event, value, meta) {
            kit._call(output, output[meta[value]], [event])
            return true
        }),
        onchoice: shy("菜单列表", {
            "返回": "Last",
            "清空": "clear",
            "复制": "Copy",
            "下载": "Download",
            "表格": function() {plugin.onfigure("table")},
            "媒体": function() {plugin.onfigure("media")},
            "绘图": function() {plugin.onfigure("canvas")},
        }, ["返回", "清空", "复制", "下载", "表格", "媒体", "绘图"], function(event, value, meta) {
            return output._call(meta[value], event)
        }),
        onaction: shy("事件列表", {
            oncontextmenu: function(event) {
                plugin.oncarte(event, output.onchoice)
            },
        }, function(event, key, cb) {cb(event)}),
    }, plugin)
    kit.Log("init", "output", output.Zones(), output)
    output.onimport(type, msg, cb)
    return plugin.Outputs[type] = target, target.Output = output
}


var Config = {main: "chat",
    list: ["toast", "carte"],
}
