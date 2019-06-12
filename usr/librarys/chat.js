var page = Page({
    conf: {border: 4, layout: {header:30, river:180, action:180, source:60, storm:180, footer:30}},
    onlayout: function(event, sizes) {
        kit.isWindows && (document.body.style.overflow = "hidden")

        var height = document.body.clientHeight-page.conf.border
        var width = document.body.clientWidth-page.conf.border
        page.conf.height = height
        page.conf.width = width

        sizes = sizes || {}
        sizes.header == undefined && (sizes.header = page.header.clientHeight)
        sizes.footer == undefined && (sizes.footer = page.footer.clientHeight)
        page.header.Size(width, sizes.header)
        page.footer.Size(width, sizes.footer)

        sizes.river == undefined && (sizes.river = page.river.clientWidth)
        sizes.storm == undefined && (sizes.storm = page.storm.clientWidth)
        height -= page.header.offsetHeight+page.footer.offsetHeight
        page.river.Size(sizes.river, height)
        page.storm.Size(sizes.storm, height)

        sizes.action == undefined && (sizes.action = page.action.clientHeight)
        sizes.source == undefined && (sizes.source = page.source.clientHeight)
        sizes.action == -1 && (sizes.action = height, sizes.source = 0)
        width -= page.river.offsetWidth+page.storm.offsetWidth
        page.action.Size(width, sizes.action)
        page.source.Size(width, sizes.source)

        height -= page.source.offsetHeight+page.action.offsetHeight
        page.target.Size(width, height)
        kit.History.add("lay", sizes)
    },
    oncontrol: function(event, target, action) {
        switch (action) {
            case "control":
                if (event.ctrlKey) {
                    switch (event.key) {
                        case "0":
                            page.source.Select()
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
                            page.action.Select(parseInt(event.key))
                            break
                        case "n":
                            page.ocean.Show()
                            break
                        case "m":
                            page.steam.Show()
                            break
                        case "i":
                            page.storm.Next()
                            break
                        case "o":
                            page.storm.Prev()
                            break
                        case "b":
                            page.action.Action["最大"](event)

                    }
                    break
                } else {
                    switch (event.key) {
                        case "Escape":
                            page.dialog && page.dialog.Show()
                    }
                }
                break
        }
    },

    initLogin: function(page, pane, form, output) {
        var ui = kit.AppendChild(form, [
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
                form.Run([ui.username.value, ui.password.value], function(msg) {
                    if (msg.result && msg.result[0]) {
                        pane.ShowDialog(1, 1)
                        ctx.Cookie("sessid", msg.result[0])
                        location.reload()
                        return
                    }
                    page.alert("用户或密码错误")
                })
            }]},
            {button: ["scan", function(event) {
                scan(event, function(text) {
                    alert(text)
                })
            }]},
            {type: "br"},
            {type: "img", data: {"src": "/chat/qrcode?text=hi"}}
        ])


        if (true||kit.isWeiXin) {
            pane.Run(["weixin"], function(msg) {
                // if (!ctx.Search("state")) {
                //     location.href = msg["auth2.0"][0]
                // }
                // return
                kit.AppendChild(document.body, [{include: ["https://res.wx.qq.com/open/js/jweixin-1.4.0.js", function(event) {
                    kit.AppendChild(document.body, [{include: ["/static/librarys/weixin.js", function(event) {
                        wx.error(function(res){
                        })
                        wx.ready(function(){
                            wx.getNetworkType({success: function (res) {
                            }})
                            return
                            wx.getLocation({
                                success: function (res) {
                                    page.footer.State("site", parseInt(res.latitude*10000)+","+parseInt(res.longitude*10000))
                                },
                            })
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
                            ]
                        })
                    }]}])
                }]}])
            })
        }
        form.Run([], function(msg) {
            if (msg.result && msg.result[0]) {
                page.header.State("user", msg.nickname[0])
                page.river.Show()
                page.footer.State("ip", msg.remote_ip[0])
                return
            }
            pane.ShowDialog(1, 1)
        })
        pane.Exit = function() {
            ctx.Cookie("sessid", "")
            page.reload()

        }
    },
    initOcean: function(page, pane, form, output) {
        var table = kit.AppendChild(output, "table")
        var ui = kit.AppendChild(pane, [{view: ["create ocean"], list: [
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
                            pane.Action["全选"](event)
                            return true
                        case "0":
                            pane.Action["清空"](event)
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
                    alert("请添加组员")
                    return
                }

                form.Run(cmd, function(msg) {
                    page.river.Show()
                    pane.Show()
                })
            }]}, {name: "list", view: ["list"]},
        ]}])

        pane.Show = function() {
            pane.ShowDialog() && (table.innerHTML = "", ui.list.innerHTML = "", ui.name.value = "good", ui.name.focus(), form.Run([], function(msg) {
                kit.AppendTable(table, ctx.Table(msg), ["key", "user.route"], function(value, key, row, i, tr, event) {
                    tr.className = "hidden"
                    var uis = kit.AppendChild(ui.list, [{text: [row.key], click: function(event) {
                        tr.className = "normal", uis.last.parentNode.removeChild(uis.last)
                    }}])
                })
            }))
        }
        pane.Action = {
            "取消": function(event) {
                pane.Show()
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
        }
        return {"button": ["取消", "全选", "清空"], "action": pane.Action}
    },
    initRiver: function(page, pane, form, output) {
        pane.Show = function() {
            output.Update([], "text", ["name", "count"], "key", ctx.Search("river")||true, function(line, index, event) {})
        }
        pane.Action = {
            "创建": function(event) {
                page.ocean.Show()
            },
        }
		return {"button": ["创建"], "action": pane.Action}
    },
    initTarget: function(page, pane, form, output) {
        output.DisplayUser = true
        var river = ""
        pane.Listen = {
            river: function(value, old) {
                river = value, pane.Show()
            },
        }

        pane.Stop = false
        pane.Show = function() {
            var cmds = ["brow", river, 0]
            output.innerHTML = "", pane.Times(1000, cmds, function(line, index, msg) {
                output.Append("", line, ["text"], "index", fun)
                cmds[2] = parseInt(line.index)+1
                page.footer.State("text", cmds[2])
            })
        }

        function fun(line, index, event, args, cbs) {
            var data = JSON.parse(line.text)
            form.Run(["wave", river, data.node, data.group, data.index].concat(args), cbs)
        }

        pane.Send = function(type, text, cb) {
            form.Run(["flow", river, type, text], function(msg) {
                // output.Append(type, {create_user: msg.create_user[0], text:text, index: msg.result[0]}, ["text"], "index", fun)
                typeof cb == "function" && cb()
            })
        }
        return [{"text": ["target"]}]
    },
    initSource: function(page, pane, form, output) {
        var ui = kit.AppendChild(pane, [{"view": ["input", "textarea"], "data": {"onkeyup": function(event){
            kit.isSpace(event.key) && pane.which.set(event.target.value)
            event.key == "Enter" && !event.shiftKey && page.target.Send("text", event.target.value, pane.Clear)
        }, "onkeydown": function(event) {
            event.key == "Enter" && !event.shiftKey && event.preventDefault()
        }}}])

        pane.Size = function(width, height) {
            pane.style.display = (width<=0 || height<=0)? "none": "block"
            pane.style.width = width+"px"
            pane.style.height = height+"px"
            ui.first.style.width = (width-7)+"px"
            ui.first.style.height = (height-7)+"px"
        }
        pane.Select = function() {
            ui.first.focus()
        }

        pane.Clear = function(value) {
            ui.first.value = value || ""
        }
        return
    },
    initAction: function(page, pane, form, output) {
        var cache = {}
        var river = "", storm = 0, input = "", share = ""
        pane.Listen = {
            river: function(value, old) {
                river = value
            },
            storm: function(value, old) {
                var temp = document.createDocumentFragment()
                while (output.childNodes.length>0) {
                    item = output.childNodes[0]
                    item.parentNode.removeChild(item)
                    temp.appendChild(item)
                }
                cache[river+storm] = temp
                storm = value, pane.Show()
            },
            source: function(value, old) {
                input = value, kit.Log(value)
            },
            target: function(value, old) {
                share = value, kit.Log(value)
            },
        }
        pane.Show = function() {
            if (cache[river+storm]) {
                while (cache[river+storm].childNodes.length>0) {
                    item = cache[river+storm].childNodes[0]
                    item.parentNode.removeChild(item)
                    output.appendChild(item)
                }
                cache[river+storm] = undefined
                return
            }

            output.Update([river, storm], "plugin", ["node", "name"], "index", false, function(line, index, event, args, cbs) {
                event.shiftKey? page.target.Send("field", JSON.stringify({
                    name: line.name, help: line.help, view: line.view, init: line.init,
                    node: line.node, group: line.group, index: line.index,
                    inputs: line.inputs, args: args,
                })): form.Run([river, storm, index].concat(args), function(msg) {
                    event.ctrlKey && (msg.append && msg.append[0]?
                        page.target.Send("table", JSON.stringify(ctx.Tables(msg))):
                        page.target.Send("code", msg.result.join("")))
                    cbs(msg)
                })
            })
        }

        pane.Select = function(index) {
            output.querySelectorAll("fieldset")[index-1].Select()
        }

        var toggle = true
        pane.Action = {
            "恢复": function(event, value) {
                page.onlayout(event, page.conf.layout)
            },
            "缩小": function(event, value) {
                page.onlayout(event, {action:60, source:60})
            },
            "放大": function(event, value) {
                page.onlayout(event, {action:300, source:60})
            },
            "最高": function(event, value) {
                page.onlayout(event, {action: -1})
            },
            "最宽": function(event, value) {
                page.onlayout(event, {river:0, storm:0})
            },
            "最大": function(event, value) {
                (toggle = !toggle)? page.onlayout(event, page.conf.layout): page.onlayout(event, {river:0, action:-1, source:60})
                page.target.Stop = !toggle
            },
            "全屏": function(event, value) {
                page.onlayout(event, {header:0, footer:0, river:0, action: -1, storm:0})
            },
        }
        return {"button": ["恢复", "缩小", "放大", "最高", "最宽", "最大", "全屏"], "action": pane.Action}
    },
    initStorm: function(page, pane, form, output) {
        var river = "", index = -1
        pane.Listen = {
            river: function(value, old) {
                pane.which.set(""), river = value, pane.Show()
            },
        }
        pane.Show = function() {
            output.Update([river], "text", ["key", "count"], "key", ctx.Search("storm")||true)
        }
        pane.Next = function() {
            var next = output.querySelector("div.item.select").nextSibling
            next? next.click(): output.firstChild.click()
        }
        pane.Prev = function() {
            var prev = output.querySelector("div.item.select").previousSibling
            prev? prev.click(): output.lastChild.click()
        }
        pane.Action = {
            "创建": function(event) {
                page.steam.Show()
            },
        }
        return {"button": ["创建"], "action": pane.Action}
    },
    initSteam: function(page, pane, form, output) {
        var river = ""
        pane.Listen = {
            river: function(value, old) {
                river = value
            },
        }

        var table = kit.AppendChild(output, "table")
        var device = kit.AppendChild(pane, [{"view": ["device", "table"]}]).last
        var ui = kit.AppendChild(pane, [{view: ["create steam"], list: [
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
                            pane.Action["全选"](event)
                            return true
                        case "0":
                            pane.Action["清空"](event)
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
                    alert("请添加命令")
                    return
                }

                form.Run(cmd, function(msg) {
                    pane.Show()
                    page.storm.Show()
                    page.storm.which.set(ui.name.value, true)
                })
            }]}, {name: "list", view: ["list", "table"]},
        ]}])

        pane.Show = function() {
            pane.ShowDialog() && (table.innerHTML = "", ui.name.value = "nice", form.Run([river], function(msg) {
                kit.AppendTable(table, ctx.Table(msg), ["key", "user.route"], function(value, key, pod, i, tr, event) {
                    var old = table.querySelector("tr.select")
                    tr.className = "select", old && (old.className = "normal")
                    form.Run([river, pod.key], function(msg) {
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
        }

        pane.Action = {
            "取消": function(event) {
                pane.Show()
            },
            "全选": function(event) {
                ui.list.innerHTML = "", device.querySelectorAll("tr").forEach(function(item) {
                    item.firstChild.click()
                })
            },
            "清空": function(event) {
                ui.list.innerHTML = ""
            },
        }
        return {"button": ["取消", "全选", "清空"], "action": pane.Action}
    },
    init: function(page) {
        page.initField(page, function(init, pane, form) {
            var output = pane.querySelector("div.output")

            var list = [], last = -1
            output.Clear = function() {
                output.innerHTML = "", list = [], last = -1
            }
            output.Select = function(index) {
                -1 < last && last < list.length && (list[last].className = "item")
                last = index
                list[index] && (list[index].className = "item select")
            }
            output.Append = function(type, line, key, which, cb) {
                var index = list.length, ui = page.View(output, line.type || type, line, key, function(event, cmds, cbs) {
                    output.Select(index), pane.which.set(line[which])
                    typeof cb == "function" && cb(line, index, event, cmds, cbs)
                })
                list.push(ui.last), pane.scrollBy(0, pane.scrollHeight+100)
                return ui
            }
            output.Update = function(cmds, type, key, which, first, cb) {
                output.Clear(), form.Runs(cmds, function(line, index, msg) {
                    var ui = output.Append(type, line, key, which, cb)
                    if (typeof first == "string") {
                        (line.key == first || line.name == first) && ui.first.click()
                    } else {
                        first && index == 0 && ui.first.click()
                    }
                })
            }

            if (typeof init == "function") {
                var conf = init(page, pane, form, output)
                if (conf && conf["button"]) {
                    var buttons = []
                    conf.button.forEach(function(value, index) {
                        buttons.push({"button": [value, function(event) {
                            typeof conf["action"] == "function" && conf["action"](value, event)
                            typeof conf["action"] == "object" && conf["action"][value](event, value)
                            pane.Button = value
                        }]})
                    })
                    kit.InsertChild(pane, output, "div", buttons).className = "action "+form.dataset.componet_name
                } else if (conf) {
                    kit.AppendChild(output, conf)
                }
            }
            return conf
        })

        page.onlayout(null, page.conf.layout)
        kit.isMobile && page.action.Action["最宽"]()
        ctx.Search("layout") && page.action.Action[ctx.Search("layout")]()

        page.footer.Order({"text": "", "ip": ""}, ["ip", "text"])
        kit.isMobile && page.footer.Order({"text": "", "site": "", "ip": ""}, ["ip", "text", "site"])
        page.header.Order({"user": "", "logout": "logout"}, ["logout", "user"], function(event, item, value) {
            switch (item) {
                case "title":
                    ctx.Search({"river": page.river.which.get(), "storm": page.storm.which.get(), "layout": page.action.Button})
                    break
                case "user":
                    var name = page.prompt("new name")
                    name && page.login.Run(["rename", name], function(msg) {
                        page.header.State("user", name)
                    })
                    break
                case "logout":
                    page.confirm("logout?") && page.login.Exit()
                    break
                default:
            }
        })

    },
})
