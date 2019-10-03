Script = {}
function Meta(target, obj) {
    var a = obj
    for (var i = 2; i < arguments.length; i++) {
        a.__proto__ = arguments[i], a = arguments[i]
    }

    var id = 1
    var conf = {}, conf_cb = {}
    var sync = {}
    var cache = {}
    return {__proto__: obj, target: target,
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
                    if (value == data && !force) {return}

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
            if (name === "") {return cache = {}}

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
        Require: function(file, cb) {
            if (!file || Script[file]) {return kit._call(cb, [Script[file]])}
            file.endsWith(".css")? kit.AppendChild(document.body, [{require: ["/require/"+file, function(event) {
                return Script[file] = file, kit._call(cb, [Script[file]])
            }]}]): kit.AppendChild(document.body, [{data: {what: id++}, include: ["/require/"+file, function(event) {
                return kit._call(cb, [Script[file]])
            }]}])
        },
        History: shy("操作历史", {}, [], function(value, target) {
            var list = arguments.callee.list, item
            return value == undefined? (item = list.pop()) && (item.target.value = item.value):
                list.push({value: value, target: target})
        }),
    }
}
function Page(page) {
    var script = {}, record = ""
    var carte = document.querySelector("fieldset.carte")
    carte.onmouseleave = function(event) {
        kit.ModifyView(carte, {display: "none"})
    }
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
        oncarte: function(event, cb) {
            kit.Selector(carte, "div.output", function(output) {if (!cb.list || cb.list.length == 0) {return}
                kit.AppendChilds(output, kit.List(cb.list, function(item) {
                    return item === ""? {view: ["line"]}: {text: [item, "div", "item"], click: function(event) {
                        kit._call(cb, [item, cb.meta, event]) && kit.ModifyView(carte, {display: "none"})
                    }}
                }))
                kit.ModifyView(carte, {display: "block", left: event.x, top: event.y})
                event.stopPropagation()
                event.preventDefault()
            })
        },
        ontoast: function(text, title, duration) {
            // {text, title, duration, inputs, buttons}
            if (!text) {page.toast.style.display = "none"; return}

            var args = typeof text == "object"? text: {text: text, title: title, duration: duration}
            var toast = kit.ModifyView("fieldset.toast", {
                display: "block", dialog: [args.width||text.length*10+100, args.height||80], padding: 10,
            })
            if (!args.duration && args.button) {args.duration = -1}

            var main = typeof args.text == "string"? {text: [args.text||"", "div", "content"]}: args.text

            var list = [{text: [args.title||"", "div", "title"]}, main]
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
                kit._call(cb, [line, index, event, cmds, cbs])
            })

            list.push(ui.last), field.scrollBy(0, field.scrollHeight+100)
			key && key.length > 0 && (member[line[which]] = member[line[key[0]]] = {index:index, key:line[which]});

            (type == "plugin" && line.name || type == "field") && page.Require(line.init? line.group+"/"+line.init: "", function(init) {
                page.Require(line.view? line.group+"/"+line.view: "", function(view) {
                    pane.Plugin(page, pane, ui.field, init, function(event, cmds, cbs) {
                        kit._call(cb, [line, index, event, cmds, cbs])
                    })
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
    function call(value, event) {
        // Event入口 1.1
        ctx.Event(event, {}, {name: name+"."+value})
        page.script("record", [name, value])
        typeof pane.Action == "function"? pane.Action(value, event): pane.Action[value](event, value)
    }
    pane.Button && pane.Button.length > 0 && (kit.InsertChild(field, output, "div", pane.Button.map(function(value) {
        return typeof value == "object"? {className: value[0], select: [value.slice(1), call]}:
            value == ""? {view: ["space"]} :value == "br"? {type: "br"}: {button: [value, call]}
    })).className="action")
    field.oncontextmenu = function(event) {
        page.oncarte(event, pane.Choice, function(event, value) {
            call(value, event)
            return true
        }) && (event.stopPropagation(), event.preventDefault())
    }
    option.onsubmit = function(event) {
        event.preventDefault()
    };
    kit.Log("init", "pane", name, pane)
    return page[name] = field, pane.Field = field, field.Pane = pane
}
function Plugin(page, pane, field, inits, runs) {
    var option = field.querySelector("form.option")
    var action = field.querySelector("div.action")
    var output = field.querySelector("div.output")

    var meta = field.Meta
    var name = meta.name
    var args = meta.args || []
    var inputs = JSON.parse(meta.inputs || "[]")
    var feature = JSON.parse(meta.feature||'{}')
    var display = JSON.parse(meta.display||'{}')
    var deal = (feature && feature.display) || "table"
    kit.classList.add(field, meta.group, name, feature.style)

    var plugin = Meta(field, (inits || function() {
    })(field, option, output)||{}, {Inputs: {},
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
            var input = {type: "input", name: name || item.name || "input", data: item}
            switch (item.type) {
                case "button":
                    input.name = name || item.name || item.value
                    break
                case "upfile":
                    item.type = "file"
                    break
                case "select":
                    kit.classList.add(item, "args")
                    input.type = "select", input.list = item.values.map(function(value) {
                        return {type: "option", value: value, inner: value}
                    })
                    input.value = item.value || item.values[0]
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
            item.plug = meta.name, item.name = input.name
            return Inputs(plugin, item, kit.AppendChild(option, [{view: [item.view||""], list: [{type: "label", inner: item.label||""}, input]}])[input.name]).target
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
            return msg.append && msg.append[0]? ["table", JSON.stringify(ctx.Tables(msg))]: ["code", msg.result? msg.result.join(""): ""]
        },
        Format: function() {
            field.Meta.args = arguments.length > 0? kit.List(arguments):
                kit.Selector(option, ".args", function(item) {return item.value})
            return JSON.stringify(field.Meta)
        },
        Clone: function() {
            return pane.Append("field", {text: plugin.Format()}, [], "", function(line, index, event, cmds, cbs) {
                plugin.Run(event, cmds, cbs, true)
            }).field.Plugin
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
            if (value != undefined) {
                option[key] && (option[key].value = value)
            }
            return option[key]? option[key].value: ""
        },
        Check: function(target, cb) {
            kit.Selector(option, ".args", function(item, index, list) {
                target == undefined && index == list.length-1 && plugin.Runs(window.event, cb)
                item == target && (index == list.length-1? plugin.Runs(window.event, cb): page.plugin == field && list[index+1].focus())
                return item
            }).length == 0 && plugin.Runs(window.event, cb)
        },
        Delay: function(time, event, text) {
            page.ontoast(text, "", -1)
            return setTimeout(function() {
                plugin.Runs(event), page.ontoast("")
            }, time)
        },
        Last: function() {
            plugin.History() != undefined && plugin.Check()
        },
        Runs: function(event, cb) {
            plugin.Run(event, kit.Selector(option, ".args", function(item, index) {return item.value}), cb)
        },
        Run: function(event, args, cb, silent) {
            page.script("record", ["action", name].concat(args))
            var show = true
            setTimeout(function() {
                show && page.ontoast(kit.Format(args||["running..."]), meta.name, -1)
            }, 1000)
            event.Plugin = plugin, runs(event, args, function(msg) {
                page.footer.Pane.State("ncmd", kit.History.get("cmd").length)
                silent? kit._call(cb, [msg]): plugin.ondaemon(msg, cb)
                show = false, page.ontoast("")
            })
        },
        clear: function() {
            output.innerHTML = ""
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
        upload: function(event) {
            ctx.Upload({river: meta.river, table: plugin.Option("table")}, option.upload.files[0], function(event, msg) {
                kit.OrderTable(kit.AppendTable(kit.AppendChilds(output, "table"), ctx.Table(msg), msg.append))
                page.ontoast("上传成功")
            }, function(event) {
                page.ontoast(), page.ontoast("上传进度 "+parseInt(event.loaded*100/event.total)+"%")
            })
        },

        ontoast: page.ontoast,
        onformat: shy("数据转换", {
            none: function(value) {return value||""},
            date: function(value) {return kit.format_date(new Date())},
        }, function(which, value) {var meta = arguments.callee.meta
            return (meta[which||"none"]||meta["none"])(value)
        }),

        ondaemon: shy("接收数据", function(msg, cb) {
            plugin.msg = msg, plugin.Save(""), plugin.onfigure.meta.type = "", plugin.onfigure(deal, msg, cb)
        }),
        onfigure: shy("显示数据", {type: "",
            max: function(output) {
                output.style.maxWidth = pane.target.clientWidth-30+"px"
                output.style.maxHeight = pane.target.clientHeight-60+"px"
            },
        }, function(type, msg, cb) {var meta = arguments.callee.meta
            meta.type && plugin.Save(meta.type, output), meta.type = type
            !plugin.Back(type, output) && Output(plugin, type, msg, cb, output, option)
        }),
        onchoice: shy("菜单列表", {
            "添加": "Clone",
            "删除": "Delete",
            "加参": "Appends",
            "减参": "Remove",
        }, ["添加", "删除", "加参", "减参"], function(value, meta, event) {
            kit._call(plugin, plugin[meta[value]])
            return true
        }),
        onaction: shy("事件列表", {
            oncontextmenu: function(event) {
                page.oncarte(event, plugin.onchoice)
            },
        }, function() {
            kit.Item(arguments.callee.meta, function(k, cb) {field[k] = function(event) {
                cb(event)
            }})
        }),
    })

    plugin.onaction()
    plugin.which = plugin.Sync("input")
	page[field.id] = pane[field.id] = pane[name] = field, field.Plugin = plugin
    inputs.map(function(item) {plugin.Append(item)})
    kit.Log("init", "plugin", name, plugin)
    return plugin
}
function Inputs(plugin, item, target) {
    var plug = item.plug, name = item.name, type = item.type
    var input = Meta(target, item, {
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
                var txt = kit.History.get("txt", -1)
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
                            plugin.Clone().Select()
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
        }, function() {
            kit.Item(arguments.callee.meta, function(k, cb) {target[k] = function(event) {
                cb(event)
            }})
        }),
        Event: shy("事件入口", {name: plug+"."+name}, function(event, msg) {
            return ctx.Event(event, msg||{}, arguments.callee.meta)
        }),
        which: plugin.Sync(name),
    }, plugin)
    input.onaction()
    input.onimport()
    target.value = plugin.onformat(item.init, item.value)
    plugin.Inputs[item.name] = target, target.Input = input
    !target.placeholder && item.name && (target.placeholder = item.name)
    !target.title && item.placeholder && (target.title = item.placeholder)
    kit.Log("init", "input", plug+"."+name, input)
    return input
}
function Output(plugin, type, msg, cb, target, option) {
    var exports = plugin.target.Meta.exports
    var output = Meta(target, {
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
                    plugin.Delay(3000, event, line.you+" stop...") && field.Run(event, [option.pod.value, line.you, "stop"])
                }(): field.Run(event, [option.pod.value, line.you], function(msg) {
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
                return list && list.length > 0 && kit.OrderTable(kit.AppendTable(kit.AppendChild(target, "table"), ctx.Table(msg), list), "", output.onexport)
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
                typeof cb == "function" && !cb(msg) || (target.innerHTML = "", Canvas(plugin, option, target, target.parentNode.clientWidth-45, target.parentNode.clientHeight-175, 10, msg))
            },
        }, function(type, msg, cb) {var meta = arguments.callee.meta
            meta[type](msg, cb)
        }),
        onchoice: shy("菜单列表", {
            "返回": "Last",
            "清空": "clear",
            "下载": "Download",
        }, ["返回", "清空", "下载"], function(value, meta, event) {
            kit._call(plugin, plugin[meta[value]])
            return true
        }),
        onaction: shy("事件列表", {
            oncontextmenu: function(event) {
                page.oncarte(event, output.onchoice)
            },
        }, function() {
            kit.Item(arguments.callee.meta, function(k, cb) {target[k] = function(event) {
                cb(event)
            }})
        }),
    }, plugin)
    output.onaction()
    output.onimport(type, msg, cb)
    return output
}
