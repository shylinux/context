Page({
    check: true,
    conf: {refresh: 1000, border: 4, layout: {header:30, river:120, action:180, source:60, storm:100, footer:30}},
    onlayout: function(event, sizes) {
        var page = this
        kit.isWindows && (document.body.style.overflow = "hidden")

        var height = document.body.clientHeight-page.conf.border
        var width = document.body.clientWidth-page.conf.border-2
        page.conf.height = height
        page.conf.width = width

        sizes = sizes || {}
        sizes.header == undefined && (sizes.header = page.header.clientHeight)
        sizes.footer == undefined && (sizes.footer = page.footer.clientHeight)
        page.header.Pane.Size(width, sizes.header)
        page.footer.Pane.Size(width, sizes.footer)

        sizes.river == undefined && (sizes.river = page.river.clientWidth)
        sizes.storm == undefined && (sizes.storm = page.storm.clientWidth)
        height -= page.header.offsetHeight+page.footer.offsetHeight
        page.river.Pane.Size(sizes.river, height)
        page.storm.Pane.Size(sizes.storm, height)

        sizes.action == undefined && (sizes.action = page.action.clientHeight)
        sizes.source == undefined && (sizes.source = page.source.clientHeight);
        (sizes.action == -1 || sizes.source == 0) && (sizes.action = height, sizes.source = 0)
        width -= page.river.offsetWidth+page.storm.offsetWidth
        page.action.Pane.Size(width, sizes.action)
        page.source.Pane.Size(width, sizes.source)

        height -= page.source.offsetHeight+page.action.offsetHeight
        page.target.Pane.Size(width, height)
        kit.History.add("lay", sizes)
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
        var ui = kit.AppendChild(field, [{view: ["create ocean"], list: [
            {input: ["name", function(event) {
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
                            var pre = ui.list.querySelector("pre")
                            pre && pre.click()
                            return true
                        case "=":
                            var td = table.querySelector("tr.normal td")
                            td && td.click()
                            return true
                    }
                })
                event.key == "Enter" && this.nextSibling.click()

            }]}, {button: ["create", function(event) {
                if (!ui.name.value) {ui.name.focus(); return}

                var list = kit.Selector(ui.list, "pre", function(item) {return item.innerText})
                if (list.length == 0) {kit.alert("请添加组员"); return}

                field.Pane.Create(ui.name.value, list)

            }]}, {name: "list", view: ["list"]},
        ]}])
        return {
            Append: function(msg) {
                kit.AppendTable(table, ctx.Table(msg), ["key", "user.route"], function(value, key, row, i, tr, event) {
                    tr.className = "hidden"
                    var uis = kit.AppendChild(ui.list, [{text: [row.key], click: function(event) {
                        tr.className = "normal", uis.last.parentNode.removeChild(uis.last)
                    }}])
                })
            },
            Clear: function(name) {
                table.innerHTML = "", ui.list.innerHTML = "", ui.name.value = name, ui.name.focus()
            },
            Create: function(name, list) {
                field.Pane.Run(["spawn", "", name].concat(list), function(msg) {
                    page.river.Pane.Show()
                    field.Pane.Show()
                })
            },
            Show: function() {var pane = field.Pane
                pane.Dialog() && (pane.Clear("good"), pane.Run([], pane.Append))
            },
            Action: {
                "取消": function(event) {
                    field.Pane.Show()
                },
                "全选": function(event) {
                    table.querySelectorAll("tr.normal").forEach(function(item) {
                        item.firstChild.click()
                    })
                },
                "清空": function(event) {
                    ui.list.querySelectorAll("pre").forEach(function(item) {
                        item.click()
                    })
                },
            },
            Button: ["取消", "全选", "清空"],
        }
    },
    initRiver: function(page, field, option, output) {
        return {
            Show: function() {var pane = field.Pane
                output.innerHTML = "", pane.Update([], "text", ["nick", "count"], "key", ctx.Search("river")||true)
            },
            Action: {
                "创建": function(event) {
                    page.ocean.Pane.Show()
                },
            },
            Button: ["创建"],
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
            event.key == "Enter" && !event.shiftKey && page.target.Pane.Send("text", event.target.value, field.Pane.Clear)
        }, "onkeydown": function(event) {
            event.key == "Enter" && !event.shiftKey && event.preventDefault()
        }}}])
        return {
            Select: function() {
                ui.first.focus()
            },
            Clear: function(value) {
                ui.first.value = ""
            },
            Size: function(width, height) {
                field.style.display = (width<=0 || height<=0)? "none": "block"
                field.style.width = width+"px"
                field.style.height = height+"px"
                ui.first.style.width = (width-7)+"px"
                ui.first.style.height = (height-7)+"px"
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
                var plugin = event.Plugin || {}, engine = {
                    share: function(args) {
                        typeof cbs == "function" && cbs(ctx.Share({"group": option.dataset.group, "name": option.dataset.name, "cmds": [
                            river, line.group, line.index,  args[1]||"",
                        ]}))
                        return true
                    },
                    echo: function(one, two) {
                        kit.Log(one, two)
                    },
                    help: function() {
                        var args = kit.List(arguments)
                        if (args.length > 1 && page[args[0]] && page[args[0]].Pane[args[1]]) {
                            return kit._call(page[args[0]].Pane[args[1]].Plugin.Help, args.slice(2))
                        }
                        if (args.length > 0 && page[args[0]]) {
                            return kit._call(page[args[0]].Pane.Help, args.slice(1))
                        }
                        return kit._call(page.Help, args)
                    },
                    _split: function(str) {return str.trim().split(" ")},
                    _cmd: function(arg) {
                        var args = engine._split(arg[1]);
                        if (typeof engine[args[0]] == "function") {
                            return kit._call(engine[args[0]], args.slice(1))
                        }
                        if (page.plugin && typeof page.plugin[args[0]] == "function") {
                            return kit._call(page.plugin[args[0]], args.slice(1))
                        }

                        if (page.dialog && page.dialog.Pane.Jshy(event, args)) {return true}
                        if (page.pane && page.pane.Pane.Jshy(event, args)) {return true}
                        if (page.storm && page.storm.Pane.Jshy(event, args)) {return true}
                        if (page.river && page.river.Pane.Jshy(event, args)) {return true}

                        if (page && page.Jshy(event, args)) {return true}
                        if (page.plugin && page.plugin.Plugin.Jshy(event, args)) {return true}
                        kit.Log("not find", arg[1])
                        return true
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
                        field.Pane.Run([meta.river||river, meta.storm||storm, meta.action||index].concat(args), function(msg) {
                            engine._msg(msg), typeof cbs == "function" && cbs(msg)
                        })
                    },
                }
                if (args.length > 0 && engine[args[0]] && engine[args[0]](args)) {return}
                event.shiftKey? engine._msg(): engine._run()
            },
            Show: function() {var pane = field.Pane
                if (field.Pane.Back(river+storm, output)) {return}

                pane.Clear(), pane.Update([river, storm], "plugin", ["node", "name"], "index", false, function(line, index, event, args, cbs) {
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
                    field.Pane.Save(river+storm, output)
                    storm = value, field.Pane.Show()
                },
                source: function(value, old) {input = value},
                target: function(value, old) {share = value},
            },
            Action: {
                "聊天": function(event, value) {
                    page.onlayout(event, page.conf.layout)
                },
                "办公": function(event, value) {
                    page.onlayout(event, page.conf.layout)
                    page.onlayout(event, {river: 0, action:300, source:60})
                },
                "工作": function(event, value) {
                    page.onlayout(event, page.conf.layout)
                    page.onlayout(event, {river:0, action:-1, source:60})
                },
                "最高": function(event, value) {
                    page.onlayout(event, {action: -1})
                },
                "最宽": function(event, value) {
                    page.onlayout(event, {river:0, storm:0})
                },
                "最大": function(event, value) {
                    page.onlayout(event, {header:0, footer:0, river:0, action: -1, storm:0})
                },

                "刷新": function(event, value) {
                    output.innerHTML = "", field.Pane.Show()
                },
                "清空": function(event, value) {
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

                "添加": function(event, value) {
                    page.plugin && page.plugin.Plugin.Clone().Select()
                },
                "删除": function(event, value) {
                    page.input && page.plugin.Plugin.Delete()
                },
                "加参": function(event, value) {
                    page.plugin && page.plugin.Plugin.Append({className: "args temp"})
                },
                "减参": function(event, value) {
                    page.plugin && page.plugin.Plugin.Remove()
                },

                "表格": function(event, value) {
                    page.plugin && page.plugin.Plugin.display("table")
                },
                "编辑": function(event, value) {
                    page.plugin && page.plugin.Plugin.display("editor")
                },
                "绘图": function(event, value) {
                    page.plugin && page.plugin.Plugin.display("canvas")
                },
            },
            Button: [["layout", "聊天", "办公", "工作", "最高", "最宽", "最大"],
                "", "刷新", "清空", "并行", "串行",
				"", ["display", "表格", "编辑", "绘图"],
                "", "添加", "删除", "加参", "减参",
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
            },
            Button: ["创建"],
        }
    },
    initSteam: function(page, field, option, output) {
        var river = ""
        var table = kit.AppendChild(output, "table")
        var device = kit.AppendChild(field, [{"view": ["device", "table"]}]).last
        var ui = kit.AppendChild(field, [{view: ["create steam"], list: [
            {input: ["name", function(event) {
                page.oninput(event, function(event) {
                    switch (event.key) {
                        case "i":
                            var prev = table.querySelector("tr.select").previousSibling
                            prev && prev.childNodes[0].click()
                            return true
                        case "o":
                            var next = table.querySelector("tr.select").nextSibling
                            next && next.childNodes[0].click()
                            return true
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
                    }
                }), event.key == "Enter" && this.nextSibling.click()

            }]}, {button: ["create", function(event) {
                if (!ui.name.value) {ui.name.focus(); return}

                var list = []
                kit.Selector(ui.list, "tr", function(item) {
                    list.push(item.dataset.pod)
                    list.push(item.dataset.group)
                    list.push(item.dataset.index)
                    list.push(item.dataset.name)
                })
                if (list.length == 0) {kit.alert("请添加命令"); return}

                field.Pane.Create(ui.name.value, list)

            }]}, {name: "list", view: ["list", "table"]},
        ]}])

        return {
            Append: function(com, pod) {var pane = field.Pane
                var last = kit.AppendChild(ui.list, [{
                    dataset: {pod: pod.node, group: com.key, index: com.index, name: com.name},
                    row: [com.key, com.index, com.name, com.help],
                    click: function(event) {last.parentNode.removeChild(last)},
                }]).last
            },
            Update: function(list, pod) {var pane = field.Pane
                device.innerHTML = "", kit.AppendTable(device, list, ["key", "index", "name", "help"], function(value, key, com, i, tr, event) {
                    pane.Append(com, pod)
                })
            },
            Select: function(list) {var pane = field.Pane
                table.innerHTML = "", kit.AppendTable(table, list, ["user", "node"], function(value, key, pod, i, tr, event) {
                    var old = table.querySelector("tr.select")
                    tr.className = "select", old && (old.className = "normal"), pane.Run([river, pod.user, pod.node], function(msg) {
                        pane.Update(ctx.Table(msg), pod)
                    })
                }), table.querySelector("td").click()
                ui.name.value = "nice", ui.name.focus()
            },
            Create: function(name, list) {
                field.Pane.Run([river, "spawn", name].concat(list), function(msg) {
                    field.Pane.Show(), page.storm.Pane.Show(name)
                })
            },
            Show: function() {var pane = field.Pane
                pane.Dialog() && pane.Run([river], function(msg) {
                    pane.Select(ctx.Table(msg))
                })
            },
            Listen: {
                river: function(value, old) {river = value},
            },
            Action: {
                "取消": function(event) {field.Pane.Show()},
                "清空": function(event) {ui.list.innerHTML = ""},
                "全选": function(event) {
                    ui.list.innerHTML = "", device.querySelectorAll("tr").forEach(function(item) {
                        item.firstChild.click()
                    })
                },
            },
            Button: ["取消", "清空", "全选"],
        }
    },
    init: function(page) {
        page.footer.Pane.Order({"ncmd": "", "ntxt": ""}, ["ncmd", "ntxt"], function(event, item, value) {})
        page.header.Pane.Order({"logout": "logout", "user": ""}, ["logout", "user"], function(event, item, value) {
            page.onaction[item] && page.onaction[item](event, item, value, page)
        })
        page.river.Pane.Show(), page.pane = page.action, page.plugin = kit.Selector(page.action, "fieldset")[0]
        page.action.Pane.Layout(ctx.Search("layout")? ctx.Search("layout"): kit.isMobile? "办公": "工作")
    },
})
