var page = Page({check: true,
    conf: {refresh: 1000, border: 4, first: "工作", mobile: "工作", layout: {header:30, footer:30, river:100, storm:100, action:180, source:45}},
    onlayout: function(event, sizes) {
        var page = this
        var height = document.body.clientHeight-page.conf.border
        var width = document.body.clientWidth-page.conf.border
        kit.device.isWindows && (document.body.style.overflow = "hidden")

        sizes = sizes || {}
        sizes.header == undefined && (sizes.header = page.header.clientHeight)
        sizes.footer == undefined && (sizes.footer = page.footer.clientHeight)
        page.header.Pane.Size(width, sizes.header)
        page.footer.Pane.Size(width, sizes.footer)
        height -= page.header.offsetHeight+page.footer.offsetHeight

        sizes.river == undefined && (sizes.river = page.river.clientWidth)
        sizes.storm == undefined && (sizes.storm = page.storm.clientWidth)
        page.river.Pane.Size(sizes.river, height)
        page.storm.Pane.Size(sizes.storm, height)
        width -= page.river.offsetWidth+page.storm.offsetWidth

        sizes.action == -1 && (sizes.action = kit.device.isMobile? "": height, sizes.target = 0, sizes.source = 0)
        sizes.action == undefined && (sizes.action = page.action.offsetHeight-page.conf.border)
        sizes.source == undefined && (sizes.source = page.source.clientHeight)
        sizes.target == undefined && (sizes.target = page.target.clientHeight)
		sizes.source == 0 && sizes.target == 0 && !kit.device.isMobile && (sizes.action = height)
        page.action.Pane.Size(width, sizes.action)
        page.source.Pane.Size(width, sizes.source)
        height -= sizes.target==0? height: page.source.offsetHeight+page.action.offsetHeight

        page.target.Pane.Size(width, height)
    },
    oncontrol: function(event, target, action) {
        var page = this
        switch (action) {
            case "control":
                if (event.ctrlKey) {
                    switch (event.key) {
                        case "0":
                            page.source.Pane.Select()
                            break
                        case "1":
                        case "2":
                        case "3":
                        case "4":
                        case "5":
                        case "6":
                        case "7":
                        case "8":
                        case "9":
                            page.action.Pane.Select(parseInt(event.key))
                            break
                        case "n":
                            page.ocean.Pane.Show()
                            break
                        case "m":
                            page.steam.Pane.Show()
                            break
                        case "i":
                            page.storm.Pane.Next()
                            break
                        case "o":
                            page.storm.Pane.Prev()
                            break
                        case "b":
                            page.action.Action["最大"](event)

                    }
                    break
                } else {
                    switch (event.key) {
                        case " ":
                            page.footer.Pane.Select()
                            break
                        case "Escape":
                            page.dialog && page.dialog.Pane.Show()
                            break
                    }
                }
                break
        }
    },
    onaction: {
        title: function(event, item, value, page) {
            ctx.Search({"river": page.river.Pane.which.get(), "storm": page.storm.Pane.which.get(), "layout": page.action.Pane.Layout()})
        },
        user: function(event, item, value, page) {
            var name = kit.prompt("new name")
            name && page.login.Pane.Run(["rename", name], function(msg) {
                page.header.Pane.State("user", name)
            })
        },
        logout: function(event, item, value, page) {
            kit.confirm("logout?") && page.login.Pane.Exit()
        },
    },

    initOcean: function(page, field, option, output) {
        var table = kit.AppendChild(output, "table")
        var ui = kit.AppendChild(field, [{view: ["create"], list: [
            {title: "群聊名称", input: ["name", function(event) {
                page.oninput(event, function(event) {
                    switch (event.key) {
                        case "1":
                        case "2":
                        case "3":
                        case "4":
                        case "5":
                        case "6":
                        case "7":
                        case "8":
                            var tr = table.querySelectorAll("tr.normal")[parseInt(event.key)-1]
                            tr && tr.childNodes[0].click()
                            return true
                        case "9":
                            field.Pane.Action["全选"](event)
                            return true
                        case "0":
                            field.Pane.Action["清空"](event)
                            return true

                        case "-":
                            var tr = ui.list.querySelectorAll("tr")[1]
                            tr && tr.childNodes[0].click()
                            return true
                        case "=":
                            var td = table.querySelector("tr.normal td")
                            td && td.click()
                            return true
                    }
                }), event.key == "Enter" && this.nextSibling.click()

            }]}, {button: ["创建群聊", function(event) {
                if (!ui.name.value) {ui.name.focus(); return}

                var list = kit.Selector(ui.list, "tr", function(item) {
                    return item.dataset.user
                })
                if (list.length == 0) {kit.alert("请添加组员"); return}


                field.Pane.Create(ui.name.value, list)

            }]}, {name: "list", view: ["list", "table"], list: [{text: ["2. 已选用户列表", "caption"]}]},
        ]}])
        return {
            Append: function(msg) {
                kit.AppendChilds(table, [{text: ["1. 选择用户节点 ->", "caption"]}])
                kit.AppendTable(table, msg.Table(), ["key", "user.route"], function(value, key, row, i, tr, event) {
                    tr.className = "hidden"
                    var uis = kit.AppendChild(ui.list, [{row: [row.key, row["user.route"]], dataset: {user: row.key}, click: function(event) {
                        tr.className = "normal", uis.last.parentNode.removeChild(uis.last)
                    }}])
                })
            },
            Create: function(name, list) {
                field.Pane.Run(["spawn", "", name].concat(list), function(msg) {
                    field.Pane.Show(), page.river.Pane.Show(name)
                })
            },
            Show: function(name) {var pane = field.Pane
                pane.Dialog(), ui.name.focus(), ui.name.value = name||"good", pane.Run([], pane.Append)
            },
            Action: {
                "取消": function(event) {field.Pane.Show()},
                "清空": function(event) {
                    kit.Selector(ui.list, "tr", function(item) {item.click()})
                },
                "全选": function(event) {
                    kit.Selector(table, "tr.normal", function(item) {item.firstChild.click()})
                },
            },
            Button: ["取消", "清空", "全选"],
        }
    },
    initRiver: function(page, field, option, output) {
        return {
            Show: function(which) {var pane = field.Pane
                ctx.Event(event, {}, {name: "river.show"})
                output.innerHTML = "", pane.Update([], "text", ["nick", "count"], "key", which||ctx.Search("river")||true)
            },
            Action: {
                "创建": function(event) {
                    page.ocean.Pane.Show()
                },
                "共享": function(event) {
                    page.login.Pane.Run(["relay", "river", "username", kit.prompt("分享给用户"), "url", ctx.Share({
                        "river": page.river.Pane.which.get(),
                        "layout": page.action.Pane.Layout(),
                    })], function(msg) {
                        page.toast.Pane.Show({text: location.origin+location.pathname+"?relay="+msg.result.join(""), title: "共享链接", button: ["确定"], cb: function(which) {
                            page.toast.Pane.Show()
                        }})
                    })
                },
            },
            Button: ["创建", "共享"],
        }
    },
    initTarget: function(page, field, option, output) {
        var river = "", which = {}
        output.DisplayUser = true
        output.DisplayTime = true
        return {
            Send: function(type, text, cb) {var pane = field.Pane
                pane.Run([river, "flow", type, text], function(msg) {
                    pane.Show(), typeof cb == "function" && cb(msg)
                })
            },
            Stop: function() {
                return field.style.display == "none"
            },
            Show: function(i) {var pane = field.Pane
                field.Pane.Back(river, output)

                var foot = page.footer.Pane, cmds = [river, "brow", i||which[river]||0]
                cmds[2] || (output.innerHTML = ""), pane.Tickers(page.conf.refresh, cmds, function(line, index, msg) {
                    pane.Append("", line, ["text"], "index", function(line, index, event, args, cbs) {
                        page.action.Pane.Core(event, line, args, cbs)
                    })
                    foot.State("ntxt", which[river] = cmds[2] = parseInt(line.index)+1)
                })
            },
            Listen: {
                river: function(value, old) {
                    field.Pane.Save(river, output)
                    river = value, field.Pane.Show()
                },
            },
        }
    },
    initSource: function(page, field, option, output) {
        var ui = kit.AppendChild(field, [{"view": ["input", "textarea"], "data": {"onkeyup": function(event){
            page.oninput(event), kit.isSpace(event.key) && field.Pane.which.set(event.target.value)
            event.key == "Enter" && !event.shiftKey && page.target.Pane.Send("text", event.target.value, field.Pane.clear)
            event.stopPropagation()
        }, "onkeydown": function(event) {
            event.key == "Enter" && !event.shiftKey
            event.stopPropagation()
        }}}])
        return {
            Select: function() {
                ui.first.focus()
            },
            clear: function(value) {
                ui.first.value = ""
            },
            Size: function(width, height) {
                if (width > 0) {
                    field.style.width = width+"px"
                    ui.first.style.width = (width-7)+"px"
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
                    ui.first.style.height = (height-7)+"px"
                    field.style.display = "block"
                } else if (height === "") {
                    field.style.height = ""
                    field.style.display = "block"
                } else {
                    field.style.display = "none"
                    return
                }
            },
        }
    },
    initAction: function(page, field, option, output) {
        var river = "", storm = 0, input = "", share = ""
        output.DisplayRaw = true
        return {
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
            Core: function(event, line, args, cbs) {
                var msg = ctx.Event(event)
                var plugin = event.Plugin || page.plugin && page.plugin.Plugin || {}, engine = {
                    share: function(args) {
                        return ctx.Share({"group": option.dataset.group, "names": option.dataset.names, "cmds": [
                            river, line.storm, line.action,  args[1]||"",
                        ]})
                    },
                    wssid: function(id) {
                        return id && (page.wssid = id)
                    },
                    pwd: function(name, value) {
                        name && kit.Selector(page.action, "fieldset.item."+name, function(item) {
                            item.Plugin.Select()
                        })
                        if (value) {return engine.set(value)}
                        return [river, storm, page.plugin && page.plugin.Meta.name, page.input && page.input.name, page.input && page.input.value]
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
                        kit.Log(["cmd"].concat(args))

                        if (typeof engine[args[0]] == "function") {
                            return kit._call(engine[args[0]], args.slice(1))
                        }
                        if (page.plugin && typeof page.plugin.Plugin[args[0]] == "function") {
                            return kit._call(page.plugin.Plugin[args[0]], args.slice(1))
                        }

                        if (page.dialog && (res = page.dialog.Pane.Jshy(event, args))) {return res}
                        if (page.pane && (res = page.pane.Pane.Jshy(event, args))) {return res}
                        if (page.storm && (res = page.storm.Pane.Jshy(event, args))) {return res}
                        if (page.river && (res = page.river.Pane.Jshy(event, args))) {return res}


                        if (page && (res = page.Jshy(event, args))) {return res}
                        if (page.plugin && (res = page.plugin.Plugin.Jshy(event, args))) {return res}
                        return kit.Log(["warn", "not", "find"].concat(args))
                    },
                    _msg: function(msg) {
                        if (msg) {
                            var text = plugin? plugin.Reveal(msg): ""
                            text && event.ctrlKey && page.target.Pane.Send(text[0], text[1])
                        } else {
                            page.target.Pane.Send("field", plugin.Format())
                        }
                    },
                    _run: function() {
                        var meta = plugin && plugin.target && plugin.target.Meta || {}
                        field.Pane.Run([meta.river||river, meta.storm||storm, meta.action].concat(args), function(msg) {
                            engine._msg(msg), typeof cbs == "function" && cbs(msg)
                        })
                    },
                }
                if (args.length > 0 && engine[args[0]] && msg.Echo(engine[args[0]](args))) {typeof cbs == "function" && cbs(msg); return}
                event.shiftKey? engine._msg(): engine._run()
            },
            Show: function() {var pane = field.Pane
                if (field.Pane.Back(river+storm, output)) {return}

                ctx.Event(event, {}, {name: "action.show"})
                pane.clear(), pane.Update([river, storm], "plugin", ["node", "name"], "index", false, function(line, index, event, args, cbs) {
                    pane.Core(event, line, args, cbs)
                })
            },
            Layout: function(name) {var pane = field.Pane
                var layout = field.querySelector("select.layout")
                name && pane.Action[layout.value = name](window.event, layout.value)
                return layout.value
            },
            Listen: {
                river: function(value, old) {river = value},
                storm: function(value, old) {
                    field.Pane.Save(river+"."+storm, output)
                    storm = value, field.Pane.Show()
                },
                source: function(value, old) {input = value},
                target: function(value, old) {share = value},
            },
            Action: {
                "聊天": function(event, value) {
                    page.which.set(value)
                    page.onlayout(event, page.conf.layout)
                },
                "办公": function(event, value) {
                    page.which.set(value)
                    page.onlayout(event, page.conf.layout)
                    page.onlayout(event, {river: 0, action:300})
                },
                "工作": function(event, value) {
                    page.which.set(value)
                    page.onlayout(event, page.conf.layout)
                    page.onlayout(event, {river:0, action:-1})
                },
                "最高": function(event, value) {
                    page.which.set(value)
                    page.onlayout(event, {action: -1})
                },
                "最宽": function(event, value) {
                    page.which.set(value)
                    page.onlayout(event, {river:0, storm:0})
                },
                "最大": function(event, value) {
                    page.which.set(value)
                    page.onlayout(event, {header:0, footer:0, river:0, storm:0, action: -1})
                },

                "刷新": function(event, value) {
                    output.innerHTML = "", field.Pane.Show()
                },
                "清屏": function(event, value) {
                    kit.Selector(output, "fieldset>div.output", function(item) {
                        item.innerHTML = ""
                    })
                },
                "并行": function(event, value) {
                    kit.Selector(output, "fieldset", function(item) {
                        item.Plugin.Runs(event)
                    })
                },
                "串行": function(event, value) {
                    var list = kit.Selector(output, "fieldset")
                    function run(list) {
                        list.length > 0? list[0].Plugin.Runs(event, function() {
                            field.Pane.Conf("running", true), setTimeout(function() {
                                run(list.slice(1))
                            }, 100)
                        }): field.Pane.Conf("running", false)
                    }
                    run(list)
                },

                "表格": function(event, value) {
                    page.plugin && page.plugin.Plugin.onfigure("table")
                },
                "编辑": function(event, value) {
                    page.plugin && page.plugin.Plugin.onfigure("editor")
                },
                "绘图": function(event, value) {
                    page.plugin && page.plugin.Plugin.onfigure("canvas")
                },

                "添加": function(event, value) {
                    page.plugin && page.plugin.Plugin.Clone().Select()
                },
                "删除": function(event, value) {
                    page.plugin && page.plugin.Plugin.Delete()
                },
                "加参": function(event, value) {
                    page.plugin && page.plugin.Plugin.Appends()
                },
                "减参": function(event, value) {
                    page.plugin && page.plugin.Plugin.Remove()
                },
                "执行": function(event, value) {
                    page.plugin && page.plugin.Plugin.Check()
                },
                "下载": function(event, value) {
                    page.plugin && page.plugin.Plugin.Download()
                },
                "清空": function(event, value) {
                    page.plugin && page.plugin.Plugin.clear()
                },
                "返回": function(event, value) {
                    page.plugin && page.plugin.Plugin.Last()
                },
            },
            Button: [["layout", "聊天", "办公", "工作", "最高", "最宽", "最大"],
                "", "刷新", "清屏", "并行", "串行",
				"", ["display", "表格", "编辑", "绘图"],
                "", "添加", "删除", "加参", "减参",
                "", "执行", "下载", "清空", "返回",
            ],
            Choice: [
                ["layout", "聊天", "办公", "工作"],
                "刷新", "清屏", "并行", "串行",
            ],
        }
    },
    initStorm: function(page, field, option, output) {
        var river = ""
        return {
            Next: function() {
                var next = output.querySelector("div.item.select").nextSibling
                next? next.click(): output.firstChild.click()
            },
            Prev: function() {
                var prev = output.querySelector("div.item.select").previousSibling
                prev? prev.click(): output.lastChild.click()
            },
            Show: function(which) {var pane = field.Pane
                ctx.Event(event, {}, {name: "storm.show"})
                pane.which.get("") == which && page.action.Pane.Show()
                output.innerHTML = "", pane.Update([river], "text", ["key", "count"], "key", which||ctx.Search("storm")||true)
            },
            Listen: {
                river: function(value, old) {
                    field.Pane.which.set(""), river = value, field.Pane.Show()
                },
            },
            Action: {
                "创建": function(event) {
                    page.steam.Pane.Show()
                },
                "共享": function(event) {
                    var user = kit.prompt("分享给用户")
                    if (user == null) {return}

                    page.login.Pane.Run(["relay", "storm", "username", user, "url", ctx.Share({
                        "river": page.river.Pane.which.get(),
                        "storm": page.storm.Pane.which.get(),
                        "layout": page.action.Pane.Layout(),
                    })], function(msg) {
                        var url = location.origin+location.pathname+"?relay="+msg.result.join("")
                        page.toast.Pane.Show({text: "<img src=\""+ctx.Share({"group": "index", "names": "login", cmds: ["share", url]})+"\">", height: 320, width: 320, title: url, button: ["确定"], cb: function(which) {
                            page.toast.Pane.Show()
                        }})
                    })
                },
            },
            Button: ["创建", "共享"],
            Choice: ["创建", "共享"],
        }
    },
    initSteam: function(page, field, option, output) {
        var table = kit.AppendChild(output, "table")
        var device = kit.AppendChild(field, [{"view": ["device", "table"]}]).last
        var ui = kit.AppendChild(field, [{view: ["create"], list: [
            {title: "应用名称", input: ["name", function(event) {
                page.oninput(event, function(event) {
                    switch (event.key) {
                        case "1":
                        case "2":
                        case "3":
                        case "4":
                        case "5":
                        case "6":
                        case "7":
                        case "8":
                            var tr = device.querySelectorAll("tr.normal")[parseInt(event.key)-1]
                            tr && tr.childNodes[0].click()
                            return true
                        case "9":
                            field.Pane.Action["全选"](event)
                            return true
                        case "0":
                            field.Pane.Action["清空"](event)
                            return true

                        case "-":
                            var tr = ui.list.querySelectorAll("tr")[1]
                            tr && tr.childNodes[0].click()
                            return true
                        case "=":
                            var td = device.querySelector("tr.normal td")
                            td && td.click()
                            return true

                        case "i":
                            var prev = table.querySelector("tr.select").previousSibling
                            prev && prev.childNodes[0].click()
                            return true
                        case "o":
                            var next = table.querySelector("tr.select").nextSibling
                            next && next.childNodes[0].click()
                            return true
                    }
                }), event.key == "Enter" && this.nextSibling.click()

            }]}, {button: ["创建应用", function(event) {
                if (!ui.name.value) {ui.name.focus(); return}

                var list = []
                kit.Selector(ui.list, "tr", function(item) {
                    list.push(item.dataset.pod)
                    list.push(item.dataset.group)
                    list.push(item.dataset.index)
                    list.push(item.dataset.name)
                })

                field.Pane.Create(ui.name.value, list)

            }]}, {name: "list", view: ["list", "table"], list: [{text: ["3. 已选命令列表", "caption"]}]},
        ]}])

        var river = ""
        return {
            Select: function(com, pod) {var pane = field.Pane
                var last = kit.AppendChild(ui.list, [{
                    dataset: {pod: pod.node, group: com.key, index: com.index, name: com.name},
                    row: [com.key, com.index, com.name, com.help],
                    click: function(event) {last.parentNode.removeChild(last)},
                }]).last
            },
            Update: function(list, pod) {var pane = field.Pane
                kit.AppendChilds(device, [{text: ["2. 选择模块命令 ->", "caption"]}])
                kit.AppendTable(device, list, ["key", "index", "name", "help"], function(value, key, com, i, tr, event) {
                    pane.Select(com, pod)
                }, function(value, key, com, i, tr, event) {
                    page.carte.Pane.Show(event, ["创建"], function(event, item) {
                        pane.Create(com.key)
                    })
                })
            },
            Append: function(msg) {var pane = field.Pane
                kit.AppendChilds(table, [{text: ["1. 选择用户节点 ->", "caption"]}])
                kit.AppendTable(table, msg.Table(), ["user", "node"], function(value, key, pod, i, tr, event) {
                    kit.Selector(table, "tr.select", function(item) {item.className = "normal"})
                    tr.className = "select", pane.Run([river, pod.user, pod.node], function(msg) {
                        pane.Update(msg.Table(), pod)
                    })
                }), table.querySelector("td").click()
            },
            Create: function(name, list) {
                field.Pane.Run([river, "spawn", name].concat(list||[]), function(msg) {
                    field.Pane.Show(), page.storm.Pane.Show(name)
                })
            },
            Show: function(name) {var pane = field.Pane
                pane.Dialog(), ui.name.focus(), ui.name.value = name||"nice", pane.Run([river], pane.Append)
            },
            Listen: {
                river: function(value, old) {river = value},
            },
            Action: {
                "取消": function(event) {field.Pane.Show()},
                "清空": function(event) {
                    kit.Selector(ui.list, "tr", function(item) {item.click()})
                },
                "全选": function(event) {
                    kit.Selector(device, "tr.normal", function(item) {item.firstChild.click()})
                },
            },
            Button: ["取消", "清空", "全选"],
        }
    },
    init: function(page) {
		page.action.Pane.Layout(ctx.Search("layout")? ctx.Search("layout"): kit.device.isMobile? page.conf.first: page.conf.mobile)
        page.footer.Pane.Order({"ncmd": "0", "ntxt": "0"}, ["ncmd", "ntxt"], function(event, item, value) {})
        page.header.Pane.Order({"logout": "logout", "user": "", "title": "github.com/shylinux/context"}, ["logout", "user"], function(event, item, value) {
            page.onaction[item] && page.onaction[item](event, item, value, page)
        })
        page.river.Pane.Show()
        page.WSS()
    },
})
