var page = Page({
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
    oncontrol: function(event) {
        var page = this
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
    },

    Action: {
        title: function(event, item, value, page) {
            ctx.Search({"river": page.river.Pane.which.get(), "storm": page.storm.Pane.which.get(), "layout": page.action.Pane.Layout()})
        },
        user: function(event, item, value, page) {
            page.carte.Pane.Show(event, shy({
                "修改昵称": function(event) {
                    var name = kit.prompt("new name")
                    name && page.login.Pane.Run(event, ["rename", name], function(msg) {
                        page.header.Pane.State("user", name)
                    })
                },
                "退出登录": function(event) {
                    kit.confirm("logout?") && page.login.Pane.Exit()
                },
            }, ["修改昵称", "退出登录"], function(event, value, meta) {
                meta[value](event)
            }))
        },

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
    },
    Button: shy({"title": "github.com/shylinux/context", "user": "", "time": ""}, ["time", "user"], function(key, value) {var meta = arguments.callee.meta
        return kit.isNone(key)? meta: kit.isNone(value)? meta[key]: (meta[key] = value, page.header.Pane.Show())
    }),
    Status: shy({title: '<a href="mailto:shylinux@163.com">shylinux@163.com</a>', "ncmd": "0", "ntxt": "0"}, ["ncmd", "ntxt"], function(key, value) {var meta = arguments.callee.meta
        return kit.isNone(key)? meta: kit.isNone(value)? meta[key]: (meta[key] = value, page.footer.Pane.Show())
    }),

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
                field.Pane.Run(event, ["spawn", "", name].concat(list), function(msg) {
                    field.Pane.Show(), page.river.Pane.Show(name)
                })
            },
            Show: function(name) {var pane = field.Pane
                pane.Dialog(), ui.name.focus(), ui.name.value = name||"good", pane.Run(event, [], pane.Append)
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
                pane.Event(event, {}, {name: pane.Zone("show", page.who.get())})
                output.innerHTML = "", pane.Appends([], "text", ["nick", "count"], "key", which||ctx.Search("river")||true)
            },
            Action: {
                "创建": function(event) {
                    page.ocean.Pane.Show()
                },
                "共享": function(event) {
                    page.login.Pane.Run(event, ["relay", "river", "username", kit.prompt("分享给用户"), "url", ctx.Share({
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
            Choice: ["创建", "共享"],
        }
    },
    initTarget: function(page, field, option, output) {
        var river = "", which = {}
        output.DisplayUser = true
        output.DisplayTime = true
        return {
            Send: function(type, text, cb) {var pane = field.Pane
                pane.Run(event, [river, "flow", type, text], function(msg) {
                    pane.Show(), typeof cb == "function" && cb(msg)
                })
            },
            Stop: function() {
                return field.style.display == "none"
            },
            Show: function(i) {var pane = field.Pane
                field.Pane.Load(river, output)

                var foot = page.footer.Pane, cmds = [river, "brow", i||which[river]||0]
                cmds[2] || (output.innerHTML = ""), pane.Tickers(page.conf.refresh, cmds, function(line, index, msg) {
                    pane.Append("", line, ["text"], "index")
                    page.Status("ntxt", which[river] = cmds[2] = parseInt(line.index)+1)
                })
            },
            Listen: {
                river: function(value, old) {
                    field.Pane.Save(river, output)
                    river = value, field.Pane.Show()
                },
            },
            Choice: [
                ["layout", "工作", "聊天", "最高"],
            ],
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
        var temp = ""
        output.DisplayRaw = true
        return {
            Show: function() {var pane = field.Pane
                if (river && storm && field.Pane.Load(river+"."+storm, output)) {return}

                pane.Event(event, {}, {name: pane.Zone("show", river, storm)})
                output.innerHTML = "", pane.Appends([river, storm], "plugin", ["name", "help"], "name", true, null, function() {
                })
            },
            Layout: function(name) {var pane = field.Pane
                var layout = field.querySelector("select.layout")
                name && page.Action[layout.value = name](window.event, layout.value)
                return layout.value
            },
            Listen: {
                river: function(value, old) {temp = value},
                storm: function(value, old) {
                    river && storm && field.Pane.Save(river+"."+storm, output)
                    ;(river = page.river.Pane.which.get(), storm = value) && field.Pane.Show()
                },
                source: function(value, old) {input = value},
                target: function(value, old) {share = value},
            },
            Action: {
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

                "复制": function(event, value) {
                    page.plugin && page.plugin.Plugin.Clone()
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
                "调试": function(event, value) {
                    page.debug.Pane.Show()
                },
            },
            Button: [["layout", "聊天", "办公", "工作", "最高", "最宽", "最大"],
                "", "刷新", "清屏", "并行", "串行",
				"", ["display", "表格", "编辑", "绘图"],
                "", "复制", "删除", "加参", "减参",
                "", "执行", "下载", "清空", "返回",
            ],
            Choice: [
                ["layout", "工作", "聊天", "最高"],
                "", "刷新", "清屏", "并行", "串行", "调试",
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
                var data = river && field.Pane.Load(river, output)
                if (data) {return pane.which.set(data.which)}

                pane.Event(event, {}, {name: pane.Zone("show", river)})
                output.innerHTML = "", pane.Appends([river], "text", ["key", "count"], "key", which||ctx.Search("storm")||true, null)
            },
            Listen: {
                river: function(value, old) {var pane = field.Pane
                    river && pane.Save(river, output, {which: pane.which.get()})
                    river = value, pane.which.set(""), pane.Show()
                },
            },
            Action: {
                "删除": function(event, value, meta, line) {var pane = field.Pane
                    kit.confirm("删除") && pane.Run(event, [river, "delete", line.key], function(msg) {
                        field.Pane.Show()
                    })
                },
                "共享": function(event, value, meta, line) {
                    var user = kit.prompt("分享给用户")
                    if (user == null) {return}

                    page.login.Pane.Run(event, ["relay", "storm", "username", user, "url", ctx.Share({
                        "river": page.river.Pane.which.get(),
                        "storm": line.key,
                        "layout": page.action.Pane.Layout(),
                    })], function(msg) {
                        var url = location.origin+location.pathname+"?relay="+msg.result.join("")
                        page.toast.Pane.Show({text: "<img src=\""+ctx.Share({"group": "index", "names": "login", cmds: ["share", url]})+"\">", height: 320, width: 320, title: url, button: ["确定"], cb: function(which) {
                            page.toast.Pane.Show()
                        }})
                    })
                },
                "复制": function(event, value, meta, line) {var pane = field.Pane
                    var name = kit.prompt("名称")
                    name && pane.Run(event, [river, "clone", name, line.key], function(msg) {
                        field.Pane.Show(name)
                    })
                },
                "恢复": function(event, value, meta, line) {
                    var status = JSON.parse(line.status)
                    kit.Selector(page.action, "fieldset.item", function(field, index) {
                        var args = status[index].args
                        kit.Selector(field, ".args", function(input, index) {
                            input.value = args[index]||""
                        })
                    })
                },
                "保存": function(event, value, meta, line) {var pane = field.Pane
                    field.Pane.Run(event, [river, "save", pane.which.get(),
                        line.status=JSON.stringify(kit.Selector(page.action, "fieldset.item", function(field) {
                            return {name: field.Meta.name, args: kit.Selector(field, ".args", function(input) {
                                return input.value
                            })}
                        })), JSON.stringify(kit.Selector(page.action, "fieldset.item", function(field) {
                            return {group: field.Meta.group, index: field.Meta.index, name: field.Meta.name, node: field.Meta.node}
                        }))], function(msg) {
                            page.toast.Pane.Show("保存成功")
                        })
                },
                "创建": function(event) {
                    page.steam.Pane.Show()
                },
                "刷新": function(event) {var pane = field.Pane
                    pane.Save(""), field.Pane.Show()
                },
            },
            Button: ["刷新", "创建"],
            Choice: ["刷新", "创建"],
            Detail: ["保存", "恢复", "复制", "共享", "删除"],
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

        var river = "", user = "", node = ""
        return {
            Select: function(com, pod) {var pane = field.Pane
                var last = kit.AppendChild(ui.list, [{
                    dataset: {pod: pod.node, group: com.key, index: com.index, name: com.name},
                    row: [com.key, com.index, com.name, com.help],
                    click: function(event) {last.parentNode.removeChild(last)},
                }]).last
            },
            Appends: function(list, pod) {var pane = field.Pane
                kit.AppendChilds(device, [{text: ["2. 选择模块命令 ->", "caption"]}])
                kit.AppendTable(device, list, ["key", "index", "name", "help"], function(value, key, com, i, tr, event) {
                    pane.Select(com, pod)
                }, function(value, key, com, i, tr, event) {
                    page.carte.Pane.Show(event, shy(pane.Action, pane.Detail, function(event, item, meta) {
                        meta[item](event, value, key, com)
                    }))
                })
            },
            Append: function(msg) {var pane = field.Pane
                kit.AppendChilds(table, [{text: ["1. 选择用户节点 ->", "caption"]}])
                kit.AppendTable(table, msg.Table(), ["user", "node"], function(value, key, pod, i, tr, event) {
                    pane.Event(event, {}, {name: pane.Zone("show", river, pod.user, pod.node)})
                    kit.Selector(table, "tr.select", function(item) {item.className = "normal"})

                    node && field.Pane.Save(river+"."+user+"."+node, device)
                    user = pod.user, node = pod.node, tr.className = "select"
                    if (field.Pane.Load(river+"."+user+"."+node, device)) {return}

                    pane.Run(event, [river, pod.user, pod.node], function(msg) {
                        pane.Appends(msg.Table(), pod)
                    })
                }), table.querySelector("td").click()
            },
            Create: function(name, list) {
                field.Pane.Run(event, [river, "spawn", name].concat(list||[]), function(msg) {
                    field.Pane.Show(), page.storm.Pane.Show(name)
                })
            },
            Show: function(name) {var pane = field.Pane
                pane.Event(event, {}, {name: pane.Zone("show", river)})
                pane.Dialog(), ui.name.focus(), ui.name.value = name||"nice", pane.Run(event, [river], pane.Append)
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
                "刷新": function(event) {
                    field.Pane.Save(""), field.Pane.Show(), field.Pane.Show()
                },
                "创建": function(event, value, key, line) {
                    field.Pane.Create(line.key)
                },
            },
            Button: ["取消", "清空", "全选", "刷新"],
            Choice: ["取消", "清空", "全选", "刷新"],
            Detail: ["创建"],
        }
    },
    init: function(page) {
		page.Action[ctx.Search("layout") || (kit.device.isMobile? page.conf.first: page.conf.mobile)]()
        page.river.Pane.Show()
        page.WSS()
        var update = function() {
            page.Button("time", kit.time("", "%H:%M:%S"))
            setTimeout(update, 1000)
        }
        setTimeout(update, 1)
    },
})
