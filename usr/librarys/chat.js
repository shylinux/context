page = Page({
    conf: {refresh: 1000, border: 4, layout: {header:30, river:120, action:180, source:60, storm:100, footer:30}},
    onlayout: function(event, sizes) {
        var page = this
        kit.isWindows && (document.body.style.overflow = "hidden")

        var height = document.body.clientHeight-page.conf.border
        var width = document.body.clientWidth-page.conf.border
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
                        case "Escape":
                            page.dialog && page.dialog.Pane.Show()
                    }
                }
                break
        }
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
                if (!ui.name.value) {
                    ui.name.focus()
                    return
                }

                var cmd = ["spawn", "", ui.name.value]
                ui.list.querySelectorAll("pre").forEach(function(item) {
                    cmd.push(item.innerText)
                })
                if (cmd.length == 3) {
                    kit.alert("请添加组员")
                    return
                }

                field.Pane.Run(cmd, function(msg) {
                    page.river.Pane.Show()
                    field.Pane.Show()
                })
            }]}, {name: "list", view: ["list"]},
        ]}])

        return {
            Show: function() {
                this.ShowDialog() && (table.innerHTML = "", ui.list.innerHTML = "", ui.name.value = "good", ui.name.focus(), this.Run([], function(msg) {
                    kit.AppendTable(table, ctx.Table(msg), ["key", "user.route"], function(value, key, row, i, tr, event) {
                        tr.className = "hidden"
                        var uis = kit.AppendChild(ui.list, [{text: [row.key], click: function(event) {
                            tr.className = "normal", uis.last.parentNode.removeChild(uis.last)
                        }}])
                    })
                }))
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
            Show: function() {
                this.Update([], "text", ["name", "count"], "key", ctx.Search("river")||true)
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
        var river = ""
        var which = {}
        output.DisplayUser = true
        return {
            Listen: {
                river: function(value, old) {
                    field.Pane.Save(river, output)
                    river = value, field.Pane.Show()
                },
            },
            Stop: function() {
                return field.style.display == "none"
            },
            Show: function(i) {
                field.Pane.Back(river, output)

                var pane = this, foot = page.footer.Pane
                var cmds = ["brow", river, i||which[river]||0]
                cmds[2] || (output.innerHTML = ""), pane.Times(page.conf.refresh, cmds, function(line, index, msg) {
                    pane.Append("", line, ["text"], "index", function(line, index, event, args, cbs) {
                        var text = JSON.parse(line.text)
                        page.action.Pane.Run([text.river, text.storm, text.action].concat(args), function(msg) {
                            typeof cbs == "function" && cbs(msg)
                        })
                    })
                    foot.State("text", which[river] = cmds[2] = parseInt(line.index)+1)
                })
            },
            Send: function(type, text, cb) {
                var pane = this
                pane.Run(["flow", river, type, text], function(msg) {
                    pane.Show(), typeof cb == "function" && cb(msg)
                })
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
            Size: function(width, height) {
                field.style.display = (width<=0 || height<=0)? "none": "block"
                field.style.width = width+"px"
                field.style.height = height+"px"
                ui.first.style.width = (width-7)+"px"
                ui.first.style.height = (height-7)+"px"
            },
            Select: function() {
                ui.first.focus()
            },
            Clear: function(value) {
                ui.first.value = ""
            },
        }
    },
    initAction: function(page, field, option, output) {
        var river = "", storm = 0, input = "", share = ""

        output.DisplayRaw = true
        return {
            Listen: {
                river: function(value, old) {
                    river = value
                },
                storm: function(value, old) {
                    field.Pane.Save(river+storm, output)
                    storm = value, field.Pane.Show()
                },
                source: function(value, old) {
                    input = value, kit.Log(value)
                },
                target: function(value, old) {
                    share = value, kit.Log(value)
                },
            },
            Show: function() {
                if (field.Pane.Back(river+storm, output)) {
                    return
                }

                this.Update([river, storm], "plugin", ["node", "name"], "index", false, function(line, index, event, args, cbs) {
                    var plugin = event.Plugin
                    event.shiftKey? page.target.Pane.Send("field", plugin.Format()):
                        field.Pane.Run([river, storm, index].concat(args), function(msg) {
                            var text = plugin? plugin.Reveal(msg): ""
                            text && event.ctrlKey && page.target.Pane.Send(text[0], text[1])
                            typeof cbs == "function" && cbs(msg)
                        })
                })
            },
            Layout: function(name) {
                var layout = field.querySelector("select.layout")
                name && this.Action[layout.value = name](null, layout.value)
                return layout.value
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
                    page.plugin && page.plugin.Plugin.Append({})
                },
                "减参": function(event, value) {
                    page.plugin && page.plugin.Plugin.Remove()
                },
            },
            Button: [["layout", "聊天", "办公", "工作", "最高", "最宽", "最大"], "",
                "刷新", "清空", "并行", "串行", "",
                "添加", "删除", "加参", "减参", "",
            ],
        }
    },
    initStorm: function(page, field, option, output) {
        var river = ""
        return {
            Listen: {
                river: function(value, old) {
                    field.Pane.which.set(""), river = value, field.Pane.Show()
                },
            },
            Show: function(which) {
                this.which.get("") == which && page.action.Pane.Show()
                this.Update([river], "text", ["key", "count"], "key", which||ctx.Search("storm")||true)
            },
            Next: function() {
                var next = output.querySelector("div.item.select").nextSibling
                next? next.click(): output.firstChild.click()
            },
            Prev: function() {
                var prev = output.querySelector("div.item.select").previousSibling
                prev? prev.click(): output.lastChild.click()
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
                })
                event.key == "Enter" && this.nextSibling.click()
            }]}, {button: ["create", function(event) {
                if (!ui.name.value) {
                    ui.name.focus()
                    return
                }

                var cmd = [river, "spawn", ui.name.value]
                ui.list.querySelectorAll("tr").forEach(function(item) {
                    cmd.push(item.dataset.pod)
                    cmd.push(item.dataset.group)
                    cmd.push(item.dataset.index)
                    cmd.push(item.dataset.name)
                })

                if (cmd.length == 4) {
                    kit.alert("请添加命令")
                    return
                }

                field.Pane.Run(cmd, function(msg) {
                    field.Pane.Show()
                    page.storm.Pane.Show(ui.name.value)
                })
            }]}, {name: "list", view: ["list", "table"]},
        ]}])

        return {
            Listen: {
                river: function(value, old) {
                    river = value
                },
            },
            Show: function() {
                this.ShowDialog() && (table.innerHTML = "", ui.name.value = "nice", this.Run([river], function(msg) {
                    kit.AppendTable(table, ctx.Table(msg), ["key", "user.route"], function(value, key, pod, i, tr, event) {
                        var old = table.querySelector("tr.select")
                        tr.className = "select", old && (old.className = "normal"), field.Pane.Run([river, pod.key], function(msg) {
                            device.innerHTML = "", kit.AppendTable(device, ctx.Table(msg), ["key", "index", "name", "help"], function(value, key, com, i, tr, event) {
                                var last = kit.AppendChild(ui.list, [{type: "tr", list: [
                                    {text: [com.key, "td"]}, {text: [com.index, "td"]}, {text: [com.name, "td"]}, {text: [com.help, "td"]},
                                ], dataset: {pod: pod["user.route"], group: com.key, index: com.index, name: com.name}, click: function(event) {
                                    last.parentNode.removeChild(last)
                                }}]).last
                            })
                        })
                    })
                    table.querySelector("td").click()
                    ui.name.focus()
                }))
            },
            Action: {
                "取消": function(event) {
                    field.Pane.Show()
                },
                "全选": function(event) {
                    ui.list.innerHTML = "", device.querySelectorAll("tr").forEach(function(item) {
                        item.firstChild.click()
                    })
                },
                "清空": function(event) {
                    ui.list.innerHTML = ""
                },
            },
            Button: ["取消", "全选", "清空"],
        }
    },
    init: function(page) {
        page.onlayout(null, page.conf.layout)
        page.action.Pane.Layout(ctx.Search("layout")? ctx.Search("layout"): kit.isMobile? "办公": "工作")
        page.footer.Pane.Order({"ncmd": "", "text": "", ":":""}, kit.isMobile? ["ncmd", "text"]: ["ncmd", "text", ":"], function(event, item, value) {})
        page.header.Pane.Order({"logout": "logout", "user": ""}, ["logout", "user"], function(event, item, value) {
            switch (item) {
                case "title":
                    ctx.Search({"river": page.river.Pane.which.get(), "storm": page.storm.Pane.which.get(), "layout": page.action.Pane.Layout()})
                    break
                case "user":
                    var name = kit.prompt("new name")
                    name && page.login.Pane.Run(["rename", name], function(msg) {
                        page.header.Pane.State("user", name)
                    })
                    break
                case "logout":
                    kit.confirm("logout?") && page.login.Pane.Exit()
                    break
                default:
            }
        })
        false && kit.isWeiXin && page.login.Pane.Run(["weixin"], function(msg) {
            page.Include([
                "https://res.wx.qq.com/open/js/jweixin-1.4.0.js",
                "/static/librarys/weixin.js",
            ], function(event) {
                wx.error(function(res){})
                wx.ready(function(){
                    page.getLocation = function(cb) {
                        wx.getLocation({success: function (res) {
                            cb(res)
                        }})
                    }
                    page.openLocation = function(latitude, longitude, name) {
                        wx.openLocation({latitude: parseFloat(latitude), longitude: parseFloat(longitude), name:name||"here"})
                    }

                    wx.getNetworkType({success: function (res) {}})
                    wx.getLocation({success: function (res) {
                        page.footer.Pane.State("site", parseInt(res.latitude*10000)+","+parseInt(res.longitude*10000))
                    }})
                })
                wx.config({
                    appId: msg.appid[0],
                    timestamp: msg.timestamp[0],
                    nonceStr: msg.nonce[0],
                    signature: msg.signature[0],
                    jsApiList: [
                        "scanQRCode",
                        "chooseImage",
                        "closeWindow",
                        "openAddress",
                        "getNetworkType",
                        "getLocation",
                        "openLocation",
                    ]
                })
            })
        })
        page.login.Pane.Run([], function(msg) {
            if (msg.result && msg.result[0]) {
                page.header.Pane.State("user", msg.nickname[0])
                page.footer.Pane.State("ip", msg.remote_ip[0])
                page.river.Pane.Show()
                return
            }
            page.login.Pane.ShowDialog(1, 1)
        })
    },
})
