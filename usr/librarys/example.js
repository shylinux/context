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
    var cache = {}
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
        Save: shy("保存视图", function(name, output) {if (name === "") {return cache = {}}
            kit.Log("view", "save", zone, name)
            var temp = document.createDocumentFragment()
            while (output.childNodes.length>0) {
                var item = output.childNodes[0]
                item.parentNode.removeChild(item)
                temp.appendChild(item)
            }
            cache[name] = temp
            return name
        }),
        Back: shy("恢复视图", function(name, output) {if (kit.isNone(cache[name])) {return}
            kit.Log("view", "back", zone, name)
            while (cache[name].childNodes.length>0) {
                var item = cache[name].childNodes[0]
                item.parentNode.removeChild(item)
                output.appendChild(item)
            }
            delete(cache[name])
            return name
        }),
        View: shy("添加视图", function(output, type, line, key, cb) {
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
                    if (!text.name) {return {}}

                    var id = "plugin"+this.ID()
                    list.push({view: ["item", "fieldset", "", "field"], data: {id: id, Run: cb}, list: [
                        {text: [text.name+"("+text.help+")", "legend"]},
                        {view: ["option", "form", "", "option"], list: [{type: "input", style: {"display": "none"}}]},
                        {view: ["output", "div", "", "output"]},
                    ]})
                    break
            }

            var item = []
            output.DisplayUser && item.push({view: ["user", "div", line.create_nick||line.create_user]})
            output.DisplayTime && (item.push({text: [line.create_time, "div", "time"]}))
            item.push({view: ["text"], list:list})

            !output.DisplayRaw && (list = [{view: ["item"], list:item}])
            ui = kit.AppendChild(output, list)
            ui.field && (ui.field.Meta = text)
            return ui
        }),
        Include: shy("加载脚本", function(src, cb) {src = kit.List(src)
            function next(event) {src.length > 1? meta.Include(src.slice(1), cb): cb(event)}
            kit.AppendChild(target, [file.endsWith(".css")? {require: [src[0], next]}: {include: [src[0], next]}])
        }),
        History: shy("操作历史", function(value, target) {var item
            return kit.isNone(value)? (item = history.pop()) && (item.target.value = item.value):
                history.push({value: value, target: target})
        }),
    }

    // 注册事件
    meta.onaction && kit.Item(meta.onaction.meta, function(key, cb) {target[key] = function(event) {
        meta.onaction(event, key, cb)
    }})
    return meta
}
function Page(page) {
    var script = {}, record = ""
    page = Meta(document.title, document.body, page, {__proto__: ctx,
        onload: function(event) {
            // Event入口 0
            ctx.Event(event, {}, {name: document.title})
            if (page.check && !ctx.Cookie("sessid")) {
                // 用户登录
                document.querySelectorAll("body>fieldset.Login").forEach(function(field) {
                    page.Pane(page, field)
                }), page.login.Pane.Dialog(1, 1)
            } else {
                // 登录成功
                document.querySelectorAll("body>fieldset").forEach(function(field) {
                    page.Pane(page, field)
                }), page.check? page.login.Pane.Run([], function(msg) {
                    msg.result && msg.result[0]? (page.init(page), page.header.Pane.State("user", msg.nickname[0]))
                        :page.login.Pane.Dialog(1, 1)
                }): page.init(page)
            }

            // 微信接口
            kit.device.isWeiXin && page.login.Pane.Run(["weixin"], function(msg) {
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

            // 事件回调
            window.onresize = function(event) {
                page.onlayout && page.onlayout(event)
            }, document.body.onkeydown = function(event) {
                if (page.localMap && page.localMap(event)) {return}
                page.oncontrol && page.oncontrol(event, document.body, "control")

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
        Jshy: function(event, args) {var msg = ctx.Event(event)
            if (page[args[0]] && page[args[0]].type == "fieldset") {
                if (args.length > 1) {
                    return page[args[0]].Pane.Jshy(event, args.slice(1))
                } else {
                    msg.Echo("pane", args[0])
                    return page[args[0]].Pane.Show()
                }
            }
            if (script[args[0]]) {return page.script("replay", args[0])}

            return typeof page[args[0]] == "function" && kit._call(page[args[0]], args.slice(1))
        },
        WSS: function(cb, onerror, onclose) {
            return page.socket || (page.socket = ctx.WSS(cb || (function(m) {
                if (m.detail) {
                    page.action.Pane.Core(event, m, ["_cmd", m.detail], m.Reply)
                } else {
                    page.toast.Pane.Show(m.result.join(""))
                }

            }), onerror || (function() {
                page.socket.close()

            }), onclose || (function() {
                delete(page.socket), setTimeout(function() {
                    page.WSS(cb, onerror, onclose)
                }, 1000)
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
                            toast.querySelectorAll("input").forEach(function(input) {
                                values[input.name] = input.value
                            })
                            toast.querySelectorAll("select").forEach(function(input) {
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
                    kit.AppendActions(output, cb.list, function(value, event) {
                        kit._call(cb, [value, cb.meta, event]) && field.Pane.Hide()
                    }, true)
                    kit.ModifyView(field, {display: "block", left: event.x, top: event.y})
                    event.stopPropagation()
                    event.preventDefault()
                },
            }
        },
        initLogin: function(page, field, option, output) {
            var ui = kit.AppendChilds(option, [
                {label: "username"}, {input: ["username"]}, {type: "br"},
                {label: "password"}, {password: ["password"]}, {type: "br"},
                {button: ["login", function(event) {
                    if (!ui.username.value) {ui.username.focus(); return}
                    if (!ui.password.value) {ui.password.focus(); return}

                    field.Pane.Login(ui.username.value, ui.password.value, function(sessid) {
                        if (!sessid) {kit.alert("用户或密码错误"); return}
                        page.login.Pane.Dialog(1, 1), page.onload()
                    })
                }]}, {type: "br"},
            ])

            if (kit.device.isWeiXin) {
                kit.AppendChild(output, [])
            }
            return {
                Login: function(username, password, cb) {
                    field.Pane.Run([username, password], function(msg) {cb(msg.result && msg.result[0] || "")})
                },
                Exit: function() {ctx.Cookie("sessid", ""), kit.reload()},
            }
        },
        initHeader: function(page, field, option, output) {
            var state = {title: "github.com/shylinux/context"}, list = [], cb = function(event, item, value) {}
            field.onclick = function(event) {page.pane && page.pane.scrollTo(0, 0)}
            return {
                Order: function(value, order, cbs) {
                    state = value, list = order, cb = cbs || cb, field.Pane.Show()
                },
                State: function(name, value) {
                    kit.notNone(value) && (state[name] = value, field.Pane.Show())
                    return kit.isNone(name)? state: state[name]
                },
                Show: function() {
                    kit.AppendChilds(output, [
                        {"view": ["title", "div", state.title], click: function(event) {
                            cb(event, "title", state.title)
                        }},
                        {"view": ["state"], list: list.map(function(item) {return {text: [state[item], "div", "item"], click: function(event) {
                            cb(event, item, state[item])
                        }}})},
                    ])
                },
				Help: function() {return []},
            }
        },
        initFooter: function(page, field, option, output) {
            var state = {title: "<a href='mailto:shylinux@163.com'>shylinux@163.com</>"}, list = [], cb = function(event, item, value) {}
            var ui = kit.AppendChilds(output, [
                {"view": ["title", "div", state.title]},
                {"view": ["magic"], list: [{label: "0", name: "count"}, {input: ["magic", function(event) {
                    if (event.key == "Enter" || event.ctrlKey && event.key == "j") {
                        page.action.Pane.Core(event, {}, ["_cmd", event.target.value]);
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
                {"view": ["state"]},
            ])

            ui.magic.History = []

            return {
                Select: function() {ui.magic.focus()},
                Order: function(value, order, cbs) {
                    state = value, list = order, cb = cbs || cb, field.Pane.Show()
                },
                State: function(name, value) {
                    value != undefined && (state[name] = value, field.Pane.Show())
                    return name == undefined? state: state[name]
                },
                Size: function(width, height) {
                    kit.size(field, width, height)
                    ui && kit.size(ui.magic, (width - ui.count.offsetWidth - ui.first.offsetWidth - ui.last.offsetWidth - 20), height-6)
                },
                Show: function() {
                    kit.AppendChilds(ui.last, list.map(function(item) {return {text: [item+":"+state[item], "div"], click: function(item) {
                        cb(event, item, state[item])
                    }}}))
                    field.Pane.Size(field.clientWidth, field.clientHeight)
                },
                Help: function() {return []},
            }
        },
        Pane: Pane,
    }), page.which = page.Sync("layout")

    kit.Log("init", "page", page)
    return window.onload = page.onload, page
}
function Pane(page, field) {
    var option = field.querySelector("form.option")
    var action = field.querySelector("div.action")
    var output = field.querySelector("div.output")

    var timer = ""
    var name = option.dataset.names
    var list = [], last = -1, member = {}
    var pane = Meta(name, field, (page[field.dataset.init] || function() {
    })(page, field, option, output) || {}, {
        Append: function(type, line, key, which, cb) {type = type || line.type
            var index = list.length, ui = pane.View(output, type, line, key, function(event, cmds, cbs) {
                (type != "plugin" && type != "field") && pane.Select(index, line[which])
                page.script("record", [name, line[key[0]]])
                kit._call(cb, [line, index, event, cmds, cbs])
            })

            list.push(ui.last), field.scrollBy(0, field.scrollHeight+100)
			key && key.length > 0 && (member[line[which]] = member[line[key[0]]] = {index:index, key:line[which]});

            (type == "plugin" && line.name || type == "field") && page.Require(line.init? line.group+"/"+line.init: "", function(init) {
                page.Require(line.view? line.group+"/"+line.view: "", function(view) {
                    var p = pane.Plugin(page, pane, ui.field, init, function(event, cmds, cbs) {
                        kit._call(cb, [line, index, event, cmds, cbs])
                    })
                    index == 0 && p.Select()
                })
            })
            return ui
        },
        Update: function(cmds, type, key, which, first, cb, cbs) {
            pane.Runs(cmds, function(line, index, msg) {
                var ui = pane.Append(type, line, key, which, cb)
                if (typeof first == "string") {
                    (line.key == first || line.name == first || line[which] == first || line[key[0]] == first) && ui.first.click()
                } else {
                    first && index == 0 && ui.first.click()
                }
                if (index == msg[msg.append[0]].length-1) {
                    kit._call(cbs, [msg])
                }
            })
        },
        Select: function(index, key) {
            -1 < last && last < list.length && (kit.classList.del(list[last], "select"))
            last = index, list[index] && (kit.classList.add(list[index], "select"))
            // Event入口 1.0
            ctx.Event(event, {}, {name: name+"."+key})
			key && pane.which.set(key)
        },
        clear: function() {output.innerHTML = "", list = [], last = -1},

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
        Jshy: function(event, args) {var msg = ctx.Event(event)
            if (pane[args[0]] && pane[args[0]].type == "fieldset") {
                msg.result = ["plugin", args[0]]
                pane[args[0]].scrollIntoView(), pane[args[0]].Plugin.Select(true)
                return pane[args[0]].Plugin.Jshy(event, args.slice(1))
            }
            if (typeof pane.Action[args[0]] == "function") {
                msg.result = ["action", args[0]]
                return kit._call(pane.Action[args[0]], [event, args[0]])
            }
			if (member[args[0]] != undefined) {
                msg.result = ["item", args[0]]
				pane.Select(member[args[0]].index, member[args[0]].key)
				return true
			}
            return typeof pane[args[0]] == "function" && kit._call(pane[args[0]], args.slice(1))
        },

        Tickers: function(time, cmds, cb) {
            pane.Ticker(time, cmds, function(msg) {msg.Table(function(line, index) {
                cb(line, index, msg)
            })})
        },
        Ticker: function(time, cmds, cb) {timer && clearTimeout(timer)
            function loop() {!pane.Stop() && pane.Run(cmds, function(msg) {
                cb(msg), timer = setTimeout(loop, time)
            })}
            time && (timer = setTimeout(loop, 10))
        },
        Runs: function(cmds, cb) {
            pane.Run(cmds, function(msg) {msg.Table(function(line, index) {
                (cb||pane.ondaemon)(line, index, msg)
            })})
        },
        Run: function(cmds, cb) {ctx.Run(option.dataset, cmds, cb||pane.ondaemon)},

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
        which: page.Sync(name), Listen: {}, Action: {}, Button: [],
        Plugin: Plugin,

        onaction: shy("事件列表", {
            oncontextmenu: function(event) {
                pane.Choice && page.carte.Pane.Show(event, shy({}, pane.Choice, function(value, meta, event) {
                    kit._call(pane.Action[value], [event])
                }))
            },
        }, function(event, key, cb) {cb(event)}),
    })

    function call(value, event) {
        // Event入口 1.1
        ctx.Event(event, {}, {name: name+"."+value})
        page.script("record", [name, value])
        typeof pane.Action == "function"? pane.Action(value, event): pane.Action[value](event, value)
    }
    kit.AppendAction(action, pane.Button, call)

    kit.Item(pane.Listen, function(key, cb) {page.Sync(key).change(cb)})
    option.onsubmit = function(event) {event.preventDefault()};
    kit.Log("init", "pane", name, pane)
    return page[name] = field, pane.Field = field, field.Pane = pane
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

    var plugin = Meta(meta.group+"."+name, field, inits && inits(field, option, output) || {}, {Inputs: {},
        Appends: function() {
            var name = "args"+kit.Selector(option, "input.args.temp").length
            plugin.Append({type: "text", name: name, className: "args temp"}).focus()
        },
		Change: function(target, cb) {
			var value = target.value
			function reset(event) {
				value != event.target.value && kit._call(cb, [event.target.value, value])
				target.innerHTML = event.target.value
			}
			kit.AppendChilds(target, [{type: "input", value: target.innerText, data: {
				onblur: reset,
				onkeydown: function(event) {
					page.oninput(event), event.key == "Enter" && reset(event)
				},
			}}]).last.focus()
		},
        Append: shy("添加控件", function(item, name, value) {
            var count = kit.Selector(option, ".args").length
            args && count < args.length && (value = value || args[count] || "")

            var input = {plug: meta.name, type: "input", name: name || item.name || item.value || "input", data: item}
            switch (item.type) {
                case "upfile": item.type = "file"; break
                case "select":
                    input.type = "select", input.list = item.values.map(function(value) {
                        return {type: "option", value: value, inner: value}
                    })
                    item.value = value || item.value || item.values[0]
                    kit.classList.add(item, "args")
                    break
                case "textarea":
                    input.type = "textarea", item.style = "height:100px;"+"width:"+(pane.target.clientWidth-30)+"px"
                    // no break
                case "text":
                    item.value = value || item.value || ""
                    kit.classList.add(item, "args")
                    item.autocomplete = "off"
                    break
            }
            return Inputs(plugin, input, item, kit.AppendChild(option, [{view: [item.view||""], list: [{type: "label", inner: item.label||""}, input]}])[input.name]).target
        }),
        Remove: function() {
            var list = option.querySelectorAll("input.temp")
            list.length > 0 && (option.removeChild(list[list.length-1].parentNode))
        },
        Delete: function() {
            field.previousSibling.Plugin.Select()
            field.parentNode.removeChild(field)
        },
        Select: function(target) {
            page.plugin && (kit.classList.del(page.plugin, "select"))
            page.plugin = field, kit.classList.add(field, "select")
            pane.which.set(name)

            page.input = target || option.querySelectorAll("input")[1]
            plugin.which.set(page.input.name)
            page.input.focus()
        },
        Reveal: function(msg) {
            return msg.append && msg.append[0]? ["table", JSON.stringify(msg.Table())]: ["code", msg.result? msg.result.join(""): ""]
        },
        Format: function() {
            field.Meta.args = arguments.length > 0? kit.List(arguments):
                kit.Selector(option, ".args", function(item) {return item.value})
            return JSON.stringify(field.Meta)
        },
        Clone: function() {
            return pane.Append("field", {text: plugin.Format()}, [], "", function(line, index, event, cmds, cbs) {
                plugin.Run(event, cmds, cbs, true)
            }).field.Plugin.Select()
        },
        Next: function() {
            return field.nextSibling || field.parentNode.firstChild
        },
        Prev: function() {
            return field.previousSibling || field.parentNode.lastChild
        },

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
            if (typeof plugin[args[0]] == "function") {
                return kit._call(plugin[args[0]], args.slice(1))
            }
            if (args.length > 0) {
                kit.Selector(option, ".args", function(item, index) {
                    index < args.length && (item.value = args[index])
                })
            }
            return kit._call(plugin.Runs, [event])
        },

        Option: function(key, value) {
            kit.notNone(value) && option[key] && (option[key].value = value)
            return option[key]? option[key].value: ""
        },
        Check: function(target, cb) {
            kit.Selector(option, ".args", function(item, index, list) {
                kit.isNone(target)? index == list.length-1 && plugin.Runs(window.event, cb):
                    item == target && (index == list.length-1? plugin.Runs(window.event, cb): page.plugin == field && list[index+1].focus())
                return item
            }).length == 0 && plugin.Runs(window.event, cb)
        },
        Delay: function(time, event, text) {
            plugin.ontoast(text, "", -1)
            return setTimeout(function() {
                plugin.Runs(event), plugin.ontoast("")
            }, time)
        },
        Last: function() {kit.notNone(plugin.History()) && plugin.Check()},
        Runs: function(event, cb) {plugin.Run(event, kit.Selector(option, ".args", function(item, index) {return item.value}), cb)},
        Run: function(event, args, cb, silent) {var show = true
            page.script("record", ["action", name].concat(args))
            setTimeout(function() {show && plugin.ontoast(kit.Format(args||["running..."]), meta.name, -1)}, 1000)
            event.Plugin = plugin, runs(event, args, function(msg) {
                silent? kit._call(cb, [msg]): plugin.ondaemon(msg, cb), show = false, plugin.ontoast()
            })
        },
        clear: function() {output.innerHTML = ""},
        Download: function() {
            var type = ".csv", text = kit.Selector(output, "tr", function(tr) {
                return kit.Selector(tr, "td,th", function(td) {
                    return td.innerText
                }).join(",")
            }).join("\n")
            !text && (type = ".txt", text = plugin.msg.result.join(""))

            plugin.ontoast({text:'<a href="'+URL.createObjectURL(new Blob([text]))+'" target="_blank" download="'+name+type+'">'+name+type+'</a>', title: "下载中...", width: 200})
            kit.Selector(page.toast, "a", function(item) {item.click()})
        },
        upload: function(event) {
            ctx.Upload({river: meta.river, table: plugin.Option("table")}, option.upload.files[0], function(event, msg) {
                Output(plugin, "table", msg, null, output, option)
                plugin.ontoast("上传成功")
            }, function(event) {
                plugin.ontoast("上传进度 "+parseInt(event.loaded*100/event.total)+"%")
            })
        },

        ontoast: function() {kit._call(page.toast.Pane.Show, arguments)},
        oncarte: function() {kit._call(page.carte.Pane.Show, arguments)},
        onformat: shy("数据转换", {
            none: function(value) {return value||""},
            date: function(value) {return kit.time()},
        }, function(which, value) {var meta = arguments.callee.meta
            return (meta[which||"none"]||meta["none"])(value)
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
            !plugin.Back(type, output) && Output(plugin, type || feature.display, msg || plugin.msg, cb, output, option)
        }),
        onchoice: shy("菜单列表", {
            "添加": "Clone",
            "加参": "Appends",
            "减参": "Remove",
            "删除": "Delete",
        }, ["添加", "加参", "减参", "删除"], function(value, meta, event) {
            kit._call(plugin, plugin[meta[value]])
            return true
        }),
        onaction: shy("事件列表", {
            oncontextmenu: function(event) {
                plugin.oncarte(event, plugin.onchoice)
            },
        }, function(event, key, cb) {cb(event)}),
    })

    plugin.which = plugin.Sync("input")
	page[field.id] = pane[field.id] = pane[name] = field, field.Plugin = plugin
    inputs.map(function(item) {plugin.Append(item)})
    kit.Log("init", "plugin", name, plugin)
    return plugin
}
function Inputs(plugin, meta, item, target) {
    var plug = meta.plug, name = meta.name, type = item.type
    var input = Meta(plug+"."+name, target, item, {
        onimport: shy("导入数据", {}, [item.imports], function() {
            kit.List(arguments.callee.list, function(imports) {
                page.Sync(imports).change(function(value) {
                    plugin.History(target.value, target), target.value = value
                    item.action == "auto" && plugin.Runs(document.createEvent("Event"))
                })
            }), item.type == "button" && item.action == "auto" && target.click()
        }),
        onaction: shy("事件列表", {
            onfocus: function(event) {plugin.Select(target)},
            onblur: function(event) {input.which.set(target.value)},
            onclick: function(event) {
                // Event入口 2.0
                type == "button" && input.Event(event) && kit.Value(input[item.cb], plugin[item.cb], function() {
                    plugin.Check()
                })(event, input)
            },
            onchange: function(event) {
                // Event入口 2.1
                type == "select" && ctx.Event(event) && plugin.Check(item.action == "auto"? undefined: target)
            },
            ondblclick: function(event) {
                var txt = kit.History("txt", -1)
                type == "text" && txt && (target.value = txt.data.trim())
            },
            oncontextmenu: function(event) {
                type == "text" && event.stopPropagation()
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
                            plugin.Appends()
                            break
                        case "m":
                            plugin.Clone()
                            break
                        default:
                            return false
                    }
                    return true
                })

                if (event.key == "Enter" && (event.ctrlKey || item.type == "text")) {
                    // Event入口 2.1
                    input.which.set(target.value) != undefined && plugin.History(target.value, target)
                    input.Event(event) && plugin.Check(target)
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
        Event: shy("事件入口", {name: plug+"."+name}, function(event, msg) {
            return ctx.Event(event, msg||{}, arguments.callee.meta)
        }),
        which: plugin.Sync(name),
    }, plugin)

    input.onimport()
    target.value = plugin.onformat(item.init, item.value)
    plugin.Inputs[item.name] = target, target.Input = input
    type == "text" && !target.placeholder && item.name && (target.placeholder = item.name)
    !target.title && item.placeholder && (target.title = item.placeholder)
    kit.Log("init", "input", plug+"."+name, input)
    return input
}
function Output(plugin, type, msg, cb, target, option) {
    var exports = plugin.target.Meta.exports
    var output = Meta(type, target, {
        _table: function() {plugin.onfigure("table")},
        _canvas: function() {plugin.onfigure("canvas")},
        onexport: shy("导出数据", {
            "": function(value, name, line) {
                return value
            },
            see: function(value, name, line) {
                return value.split("/")[0]
            },
            you: function(value, name, line) {
                var event = window.event
                event.Plugin = plugin

                line.you && name == "status" && (line.status == "start"? function() {
                    plugin.Delay(3000, event, line.you+" stop...") && plugin.Run(event, [option.pod.value, line.you, "stop"])
                }(): plugin.Run(event, [option.pod.value, line.you], function(msg) {
                    plugin.Delay(3000, event, line.you+" start...")
                }))
                return name == "status" || line.status == "stop" ? undefined: line.you
            },
            pod: function(value, name, line, list) {
                return (option[list[0]].value? option[list[0]].value+".": "")+line.pod
            },
            dir: function(value, name, line) {
                name != "path" && (value = line.path)
                return value
            },
            tip: function(value, name, line) {
                return option.tip.value + value
            },
        }, JSON.parse(exports||'["",""]'), function(event, value, name, line) {
            var meta = arguments.callee.meta
            var list = arguments.callee.list
            ;(!list[1] || list[1] == name) && page.Sync("plugin_"+list[0]).set(meta[list[2]||""](value, name, line))
        }),
        onimport: shy("导入数据", {
            _table: function(msg, list) {
                return list && list.length > 0 && kit.OrderTable(kit.AppendTable(kit.AppendChild(target, "table"), msg.Table(), list), "", output.onexport)
            },
            _code: function(msg) {
                return msg.result && msg.result.length > 0 && kit.OrderCode(kit.AppendChild(target, [{view: ["code", "div", msg.Results()]}]).first)
            },
            inner: function(msg, cb) {
                target.innerHTML = "", plugin.onfigure.meta.max(target)
                output.onimport.meta._table(msg, msg.append) || (target.innerHTML = msg.result.join(""))
                typeof cb == "function" && cb(msg)
            },
            table: function(msg, cb) {
                target.innerHTML = ""
                output.onimport.meta._table(msg, msg.append) || output.onimport.meta._code(msg)
                typeof cb == "function" && cb(msg)
            },
            editor: function(msg, cb) {
                (target.innerHTML = "", Editor(plugin.Run, plugin, option, target, target.clientWidth-40, 400, 10, msg))
            },
            canvas: function(msg, cb) {
                target.innerHTML = "", plugin.onfigure.meta.size(function(width, height) {
                    Canvas(plugin, option, target, width-45, height-175, 10, msg)
                })
            },
        }, function(type, msg, cb) {var meta = arguments.callee.meta
            meta[type](msg, cb)
        }),
        onchoice: shy("菜单列表", {
            "返回": "Last",
            "清空": "clear",
            "下载": "Download",
            "表格": "_table",
            "绘图": "_canvas",
        }, ["返回", "表格", "绘图", "清空", "下载"], function(value, meta, event) {
            kit._call(output, output[meta[value]])
            return true
        }),
        onaction: shy("事件列表", {
            oncontextmenu: function(event) {
                plugin.oncarte(event, output.onchoice)
            },
        }, function(event, key, cb) {cb(event)}),
    }, plugin)
    output.onimport(type, msg, cb)
    return output
}
