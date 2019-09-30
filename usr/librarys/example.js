function Meta(target, obj) {
    var a = obj
    for (var i = 2; i < arguments.length; i++) {
        a.__proto__ = arguments[i], a = arguments[i]
    }

    var id = 1
    var conf = {}, conf_cb = {}
    var sync = {}
    var cache = []
    return {
        __proto__: obj,
        target: target,
        ID: function() {return id++},
        Conf: function(key, value, cb) {
            if (key == undefined) {return conf}

            cb != undefined && (conf_cb[key] = cb)

            if (value != undefined) {
                var old = conf[key]
                conf[key] = value
                conf_cb[key] && conf_cb[key](value, old)
            }
            return conf[key] == undefined && obj && obj.Conf? obj.Conf(key): conf[key]
        },
        Sync: function(m) {
            var meta = m, data = "", list = []
            return sync[m] || (sync[m] = {
                change: function(cb) {list.push(cb); return list.length-1},
                eq: function(value) {return data == value},
                neq: function(value) {return data != value},
                get: function() {return data},
                set: function(value, force) {
                    if (value == undefined) {return}
                    if (value == data && !force) {return value}

                    old_value = data, data = value
                    meta && kit.Log("key", meta, value, old_value)
                    for (var i = 0; i < list.length; i++) {
                        list[i](value, old_value)
                    }
                    return value
                },
            })
        },
        View: function(output, type, line, key, cb) {
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
                        text.view?{styles: text.view}:null, text.init?{script: ""+id+".Script = "+text.init}:null,
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
            if (!cache[name]) {return}

            while (cache[name].childNodes.length>0) {
                item = cache[name].childNodes[0]
                item.parentNode.removeChild(item)
                output.appendChild(item)
            }
            delete(cache[name])
            return name
        },
        Include: function(src, cb) {
            typeof src == "string" && (src = [src])
            kit.AppendChild(target, [{include: [src[0], function(event) {
                src.length == 1? cb(event): page.Include(src.slice(1), cb)
            }]}])
        },
    }
}
function Page(page) {
    var script = {}, record = ""
    page = Meta(document.body, page, {__proto__: ctx,
        onload: function() {
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
        ontoast: function(text, title, duration) {
            // {text, title, duration, inputs, buttons}
            if (!text) {page.toast.style.display = "none"; return}

            var args = typeof text == "object"? text: {text: text, title: title, duration: duration}
            var toast = kit.ModifyView("fieldset.toast", {
                display: "block", dialog: [args.width||text.length*10+100, args.height||80], padding: 10,
            })
            if (!args.duration && args.button) {args.duration = -1}

            var list = [{text: [args.title||"", "div", "title"]}, {text: [args.text||"", "div", "content"]}]
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

            var ui = kit.AppendChild(kit.ModifyNode(toast.querySelector("div.output"), ""), list)
            var tick = 1
            var begin = kit.time(0,"%H:%M:%S")
            var timer = args.duration ==- 1? setTimeout(function() {
                function ticker() {
                    toast.style.display != "none" && (ui.tick.innerText = begin+" ... "+(tick++)+"s") && setTimeout(ticker, 1000)
                }
                ticker()
            }, 10): setTimeout(function(){toast.style.display = "none"}, args.duration||3000)
            return page.toast = toast
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
                var pos = target.Current || -1
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
                        page.ontoast("run "+name+" done")
                    })
                    break
                default:
                    return script
            }
            return true
        },
        Help: function(pane, type, action) {
            return []
        },
        Jshy: function(event, args) {
            var msg = event.msg || {}
            if (page[args[0]] && page[args[0]].type == "fieldset") {
                if (args.length > 1) {
                    return page[args[0]].Pane.Jshy(event, args.slice(1))
                } else {
                    msg.result = ["pane", args[0]]
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
                    page.ontoast(m.result.join(""))
                }

            }), onerror || (function() {
                page.socket.close()

            }), onclose || (function() {
                page.socket = undefined, setTimeout(function() {
                    page.WSS(cb, onerror, onclose)
                }, 1000)
            })))
        },

        initToast: function() {},
        initLogin: function(page, field, option, output) {
            var ui = kit.AppendChilds(option, [
                {label: "username"}, {input: ["username"]}, {type: "br"},
                {label: "password"}, {password: ["password"]}, {type: "br"},
                {button: ["login", function(event) {
                    if (!ui.username.value) {ui.username.focus(); return}
                    if (!ui.password.value) {ui.password.focus(); return}

                    field.Pane.Login(ui.username.value, ui.password.value, function(sessid) {
                        if (!sessid) {kit.alert("用户或密码错误"); return}
                        // ctx.Cookie("sessid", sessid),
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
            var state = {}, list = [], cb = function(event, item, value) {}
            field.onclick = function(event) {page.pane && page.pane.scrollTo(0,0)}
            return {
                Order: function(value, order, cbs) {
                    state = value, list = order, cb = cbs || cb, field.Pane.Show()
                },
                State: function(name, value) {
                    value != undefined && (state[name] = value, field.Pane.Show())
                    return name == undefined? state: state[name]
                },
                Show: function() {
                    output.innerHTML = "", kit.AppendChild(output, [
                        {"view": ["title", "div", "github.com/shylinux/context"], click: function(event) {
                            cb(event, "title", "shycontext")
                        }},
                        {"view": ["state"], list: list.map(function(item) {return {text: [state[item], "div"], click: function(event) {
                            cb(event, item, state[item])
                        }}})},
                    ])
                },
				Help: function() {return []},
            }
        },
        initFooter: function(page, field, option, output) {
            var state = {}, list = [], cb = function(event, item, value) {}
            var ui = kit.AppendChild(output, [
                {"view": ["title", "div", "<a href='mailto:shylinux@163.com'>shylinux@163.com</>"]},
                {"view": ["magic"], style: {"margin-top": "-4px"}, list: [{text: ["0", "label"], name: "count"}, {input: ["magic", function(event) {
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

                }], style: {"margin-top": "-2px", "font-size": "16px"}}]},
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
                    ui.last.innerHTML = "", kit.AppendChild(ui.last, list.map(function(item) {return {text: [item+":"+state[item], "div"], click: function(item) {
                        cb(event, item, state[item])
                    }}}))
                    field.Pane.Size(field.clientWidth, field.clientHeight)
                },
                Help: function() {return []},
            }
        },
        Pane: Pane,
    })
    page.which = page.Sync("layout")
    kit.Log("init", "page", page)
    return window.onload = page.onload, page
}
function Pane(page, field) {
    var option = field.querySelector("form.option")
    var action = field.querySelector("div.action")
    var output = field.querySelector("div.output")

    var timer = ""
    var list = [], last = -1, member = {}
    var name = option.dataset.names
    var pane = Meta(field, (page[field.dataset.init] || function() {
    })(page, field, option, output) || {}, {
        Append: function(type, line, key, which, cb) {
            type = type || line.type
            var index = list.length, ui = pane.View(output, type, line, key, function(event, cmds, cbs) {
                (type != "plugin" && type != "field") && pane.Select(index, line[which])
                page.script("record", [name, line[key[0]]])
                typeof cb == "function" && cb(line, index, event, cmds, cbs)
            })

            list.push(ui.last), field.scrollBy(0, field.scrollHeight+100)
			key && key.length > 0 && (member[line[which]] = member[line[key[0]]] = {index:index, key:line[which]});
            (type == "plugin" && line.name || type == "field") && pane.Plugin(page, pane, ui.field, function(event, cmds, cbs) {
                typeof cb == "function" && cb(line, index, event, cmds, cbs)
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
                    typeof cbs == "function" && cbs(msg)
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
        clear: function() {
            output.innerHTML = "", list = [], last = -1
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
        Jshy: function(event, args) {
            var msg = event.msg || {}
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
            pane.Ticker(time, cmds, function(msg) {
                ctx.Table(msg, function(line, index) {
                    cb(line, index, msg)
                })
            })
        },
        Ticker: function(time, cmds, cb) {
            timer && clearTimeout(timer)
            function loop() {
                !pane.Stop() && pane.Run(cmds, function(msg) {
                    cb(msg), timer = setTimeout(loop, time)
                })
            }
            time && (timer = setTimeout(loop, 10))
        },
        Runs: function(cmds, cb) {
            pane.Run(cmds, function(msg) {
                ctx.Table(msg, function(line, index) {
                    (cb||pane.ondaemon)(line, index, msg)
                })
            })
        },
        Run: function(cmds, cb) {
            ctx.Run(option.dataset, cmds, cb||pane.ondaemon)
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
    })

    for (var k in pane.Listen) {
        page.Sync(k).change(pane.Listen[k])
    }
    pane.Button && pane.Button.length > 0 && (kit.InsertChild(field, output, "div", pane.Button.map(function(value) {
        function call(value, event) {
            // Event入口 1.1
            ctx.Event(event, {}, {name: name+"."+value})
            page.script("record", [name, value])
            typeof pane.Action == "function"? pane.Action(value, event): pane.Action[value](event, value)
        }
        return typeof value == "object"? {className: value[0], select: [value.slice(1), call]}:
            value == ""? {view: ["space"]} :value == "br"? {type: "br"}: {button: [value, call]}
    })).className="action")
    option.onsubmit = function(event) {
        event.preventDefault()
    };
    kit.Log("init", "pane", name, pane)
    return page[name] = field, pane.Field = field, field.Pane = pane
}
function Plugin(page, pane, field, runs) {
    var option = field.querySelector("form.option")
    var action = field.querySelector("div.action")
    var output = field.querySelector("div.output")

    var meta = field.Meta
    var name = meta.name
    var args = meta.args || []
    var inputs = JSON.parse(meta.inputs || "[]")
    var feature = JSON.parse(meta.feature||'{}')
    var display = JSON.parse(meta.display||'{}')
    var exports = JSON.parse(meta.exports||'["",""]')
    var deal = (feature && feature.display) || "table"
    kit.classList.add(field, meta.group, name, feature.style)

    var history = []
    var run = function(event, cmds, cbs) {
        ctx.Event(event, null, {name: name, Plugin: plugin})
        runs(event, cmds, cbs)
    }
    var plugin = Meta(field, (field.Script && field.Script.init || function() {
    })(run, field, option, output)||{}, {
        Inputs: {},
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
        Append: function(item, name, value) {
            kit.Item(plugin.onaction, function(k, cb) {
                item[k] == undefined && (item[k] = typeof cb == "function"? function(event) {
                    cb(event, action, item.type, input.name, item)
                }: cb)
            });
            item.value = plugin.onformat(item.init)(item.value)

            !item.title && item.name && (item.title = item.name)
            !item.placeholder && item.title && (item.placeholder = item.title)

            var input = {type: "input", name: name || item.name || "input", data: item}
            switch (item.type) {
                case "button":
                    input.name = item.name || item.value || name
                    break
                case "upfile":
                    item.type = "file"
                    break
                case "select":
                    kit.classList.add(item, "args")
                    input.type = "select", input.list = item.values.map(function(value) {
                        return {type: "option", value: value, inner: value}
                    })
                    input.value = item.value
                    break
                case "textarea":
                    kit.classList.add(item, "args")
                    input.type = "textarea", item.style = "height:100px;"+"width:"+(pane.target.clientWidth-30)+"px"
                    // no break
                case "text":
                    kit.classList.add(item, "args")
                    item.autocomplete = "off"

                    var count = kit.Selector(option, ".args").length
                    args && count < args.length && (item.value = value||args[count++]||item.value||"");
                    break
            }

            var ui = kit.AppendChild(option, [{view: [item.view||""], list: [{type: "label", inner: item.label||""}, input]}])
            var action = Meta(ui[input.name] || {}, item, plugin.onaction, plugin);

            (typeof item.imports == "object"? item.imports: typeof item.imports == "string"? [item.imports]: []).forEach(function(imports) {
                page.Sync(imports).change(function(value) {
                    history.push({target: action.target, value: action.target.value});
                    (action.target.value = value) && item.action == "auto" && plugin.Runs(document.createEvent("Event"))
                })
            })
            item.type == "button" && item.action == "auto" && action.target.click()

            item.which = plugin.Sync(input.name)
            plugin.Inputs[input.name] = ui[input.name]
            return action.target
        },
        Remove: function() {
            var list = option.querySelectorAll("input.temp")
            list.length > 0 && (option.removeChild(list[list.length-1].parentNode))
        },
        Delete: function() {
            field.previousSibling.Plugin.Select()
            field.parentNode.removeChild(field)
        },
        Select: function(silent) {
            page.plugin && (kit.classList.del(page.plugin, "select"))
            page.plugin = field, kit.classList.add(field, "select")
            !silent && (option.querySelectorAll("input")[1].focus())
			name && pane.which.set(name)
        },
        Reveal: function(msg) {
            return msg.append && msg.append[0]? ["table", JSON.stringify(ctx.Tables(msg))]: ["code", msg.result? msg.result.join(""): ""]
        },
        Format: function() {
            field.Meta.args = arguments.length > 0? kit.List(arguments):
                kit.Selector(option, ".args", function(item) {return item.value})
            return JSON.stringify(field.Meta)
        },
        Clone: function() {
            return pane.Append("field", {text: plugin.Format()}, [], "", function(line, index, event, cmds, cbs) {
                run(event, cmds, cbs)
            }).field.Plugin
        },
        Last: function() {
            var list = history.pop()
            list? (list.target.value = list.value): inputs.map(function(item) {
                option[item.name].value = item.value
            }), plugin.Check()
        },

        getLocation: function(event) {
            var x = parseFloat(option.x.value)
            var y = parseFloat(option.y.value)
            page.getLocation && page.getLocation(function(res) {
                plugin.msg = {
                    append: ["longitude", "latitude", "accuracy", "speed"],
                    longitude: [res.longitude+x+""],
                    latitude: [res.latitude+y+""],
                    accuracy: [res.accuracy+""],
                    speed: [res.speed+""],
                }
                plugin.display("table")
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

        Delay: function(time, event, text) {
            page.ontoast(text, "", -1)
            return setTimeout(function() {
                plugin.Runs(event), page.ontoast("")
            }, time)
        },
        Check: function(target, cb) {
            plugin.Select(true), kit.Selector(option, ".args", function(item, index, list) {
                target == undefined && index == list.length-1 && plugin.Runs(window.event, cb)
                item == target && (index == list.length-1? plugin.Runs(window.event, cb): page.plugin == field && list[index+1].focus())
                return item
            }).length == 0 && plugin.Runs(window.event, cb)
        },
        Runs: function(event, cb) {
            plugin.Run(event, kit.Selector(option, ".args", function(item, index) {return item.value}), cb)
        },
        Run: function(event, args, cb) {
            page.script("record", ["action", name].concat(args))
            var show = true
            setTimeout(function() {
                show && page.ontoast(kit.Format(args||["running..."]), meta.name, -1)
            }, 1000)
            event.Plugin = plugin, run(event, args, function(msg) {
                page.footer.Pane.State("ncmd", kit.History.get("cmd").length)
                plugin.msg = msg, plugin.display(deal, cb)
                show = false, page.ontoast("")
            })
        },

        clear: function() {
            output.innerHTML = ""
        },
        display: function(arg, cb) {
            deal = arg, plugin.ondaemon[deal||"table"](plugin.msg, cb)
            plugin.show_after(plugin.msg)
        },
        Download: function() {
            var text = kit.Selector(output, "tr", function(tr) {
                return kit.Selector(tr, "td,th", function(td) {
                    return td.innerText
                }).join(",")
            }).join("\n"), type = ".csv"

            !text && (text = plugin.msg.result.join(""), type = ".txt")
            page.ontoast({text:'<a href="'+URL.createObjectURL(new Blob([text]))+'" target="_blank" download="'+name+type+'">'+name+type+'</a>', title: "下载中...", width: 200})
            kit.Selector(page.toast, "a", function(item) {
                item.click()
            })
        },
        show_after: function(msg) {},
        upload: function(event) {
            ctx.Upload(option.upload.files[0], function(event, msg) {
                kit.OrderTable(kit.AppendTable(kit.AppendChilds(output, "table"), ctx.Table(msg), msg.append))
                page.ontoast("上传成功")
            }, function(event) {
                page.ontoast(), page.ontoast("上传进度 "+parseInt(event.loaded*100/event.total)+"%")
            })
        },
        onformat: Wrap(function(which) {
            var meta = arguments.callee
            return meta[which]||meta["none"]
        }, {
            none: function(value) {return value},
            date: function(value) {return kit.format_date(new Date())},
        }),
        ondaemon: {
            inner: function(msg, cb) {
                output.style.maxWidth = pane.target.clientWidth-20+"px"
                output.style.maxHeight = pane.target.clientHeight-60+"px"
                output.innerHTML = "", msg.append? kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), ctx.Table(msg), msg.append), exports[1]||"", function(event, value, name, line) {
                    page.Sync("plugin_"+exports[0]).set(plugin.onexport[exports[2]||""](value, name, line))
                }): (output.innerHTML = msg.result.join(""))
                typeof cb == "function" && cb(msg)
            },
            table: function(msg, cb) {
                output.innerHTML = ""
                !display.hide_append && msg.append && kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), ctx.Table(msg), msg.append), exports[1]||"", function(event, value, name, line) {
                    page.Sync("plugin_"+exports[0]).set(plugin.onexport[exports[2]||""](value, name, line))
                });
                (display.show_result || !msg.append) && msg.result && kit.OrderCode(kit.AppendChild(output, [{view: ["code", "div", msg.Results()]}]).first)
                typeof cb == "function" && cb(msg)
            },
            editor: function(msg, cb) {
                (output.innerHTML = "", Editor(run, plugin, option, output, output.clientWidth-40, 400, 10, msg))
            },
            canvas: function(msg, cb) {
                typeof cb == "function" && !cb(msg) || (output.innerHTML = "", Canvas(plugin, option, output, pane.target.clientWidth-45, pane.target.clientHeight-175, 10, msg))
            },
        },
        onexport: {
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
                    plugin.Delay(3000, event, line.you+" stop...") && field.Run(event, [option.pod.value, line.you, "stop"])
                }(): field.Run(event, [option.pod.value, line.you], function(msg) {
                    plugin.Delay(3000, event, line.you+" start...")
                }))
                return name == "status" || line.status == "stop" ? undefined: line.you
            },
            pod: function(value, name, line) {
                return (option[exports[0]].value? option[exports[0]].value+".": "")+line.pod
            },
            dir: function(value, name, line) {
                name != "path" && (value = line.path)
                return value
            },
            tip: function(value, name, line) {
                return option.tip.value + value
            },
        },
        onaction: {
            onfocus: function(event, action, type, name, item) {
                plugin.Select(true)
                page.input = event.target, plugin.which.set(name)
            },
            onblur: function(event, action, type, name, item) {
                item.which.set(action.target.value)
                // page.input = undefined
            },
            onclick: function(event, action, type, name, item) {
                switch (type) {
                    case "button":
                        // Event入口 2.0
                        ctx.Event(event, {}, {name: meta.name+"."+name})

                        action[item.cb]? action[item.cb](event, item, option, field):
                            plugin[item.cb]? plugin[item.cb](event, item, option, field): plugin.Check()
                        break
                    case "text":
                        if (event.ctrlKey) {
                            action.value = kit.History.get("txt", -1).data.trim()
                        }
                        break
                }
            },
            ondblclick: function(event, action, type, name, item) {
                type == "text" && (action.target.value = kit.History.get("txt", -1).data.trim())
            },
            onchange: function(event, action, type, name, item) {
                type == "select" && ctx.Event(event, {}, {name: meta.name+"."+name}) && plugin.Check(item.action == "auto"? undefined: action)
            },
            onkeydown: function(event, action, type, name, item) {
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
                            plugin.Appends()
                            break
                        case "m":
                            plugin.Clone().Select()
                            break
                        default:
                            return false
                    }
                    return true
                })
                if (item.type != "textarea" && event.key == "Enter") {
                    // Event入口 2.1
                    ctx.Event(event, {}, {name: meta.name+"."+name, pane: pane, plugin: plugin, input: item})
                    item.which.set(action.target.value)
                    history.push({target: action.target, value: action.target.value});
                    plugin.Check(action.target)
                }
            },
            onkeyup: function(event, action, type, name, item) {
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
        },
        exports: JSON.parse(meta.exports||'["",""]'),
    })

    plugin.which = plugin.Sync("input")
	page[field.id] = pane[field.id] = pane[name] = field, field.Plugin = plugin
    inputs.map(function(item) {plugin.Append(item)})
    kit.Log("init", "plugin", name, plugin)
    return plugin
}
