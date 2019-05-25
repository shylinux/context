var page = Page({
    conf: {border: 4, banner: 105, layout: {river:160, source:60, action:60, storm:160}},
    onlayout: function(event, sizes) {
        var width = document.body.offsetWidth
        var height = document.body.offsetHeight-page.conf.banner
        sizes = sizes || {}

        sizes.river == undefined && (sizes.river = page.river.offsetWidth-page.conf.border)
        sizes.storm == undefined && (sizes.storm = page.storm.offsetWidth-page.conf.border)
        sizes.width = width - sizes.river - sizes.storm-5*page.conf.border
        page.river.Size(sizes.river, height)
        page.storm.Size(sizes.storm, height)

        sizes.action == undefined && (sizes.action = page.action.offsetHeight-page.conf.border)
        sizes.source == undefined && (sizes.source = page.source.offsetHeight-page.conf.border)
        sizes.target = height - sizes.action - sizes.source - 2*page.conf.border
        if (sizes.action == -1) {
            sizes.action = height
            sizes.target = 0
            sizes.source = 0
        }
        page.target.Size(sizes.width, sizes.target)
        page.source.Size(sizes.width, sizes.source)
        page.action.Size(sizes.width, sizes.action)
        kit.History.add("lay", sizes)
    },
    oncontrol: function(event, target, action) {
        switch (action) {
            case "control":
                if (event.ctrlKey) {
                    switch (event.key) {
                        case "n":
                            page.ocean.Show()
                            break
                        case "m":
                            page.steam.Show()

                    }
                    break
                }
                break
        }
    },

    initOcean: function(page, pane, form, output) {
        var table = kit.AppendChild(output, "table")
        var ui = kit.AppendChild(pane, [{view: ["create ocean"], list: [
            {input: ["name", function(event) {
                if (event.ctrlKey) {
                    switch (event.key) {
                        case "a":
                            pane.Action["全选"](event)
                            break
                        case "c":
                            pane.Action["清空"](event)
                            break
                    }
                    return
                }

                if (event.key == "Enter") {
                    ui.name.nextSibling.click()
                }
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
            pane.ShowDialog() && (table.innerHTML = "", ui.list.innerHTML = "", ui.name.value = "good", form.Run([], function(msg) {
                kit.AppendTable(table, ctx.Table(msg), ["key", "user.route"], function(value, key, row, i, tr, event) {
                    tr.style.display = "none"
                    var uis = kit.AppendChild(ui.list, [{text: [row.key], click: function(event) {
                        tr.style.display = "", uis.last.parentNode.removeChild(uis.last)
                    }}])
                })
            }))
        }
        pane.Action = {
            "取消": function(event) {
                pane.Show()
            },
            "全选": function(event) {
                ui.list.innerHTML = ""
                table.querySelectorAll("tr").forEach(function(item) {
                    item.firstChild.click()
                })
            },
            "清空": function(event) {
                ui.list.innerHTML = ""
                table.querySelectorAll("tr").forEach(function(item) {
                    item.style.display = ""
                })

            },
        }
        return {"button": ["取消", "全选", "清空"], "action": pane.Action}
    },
    initRiver: function(page, pane, form, output) {
        pane.Show = function() {
            output.Update([], "text", ["name", "count"], "key", true)
        }
        pane.Show()
        pane.Action = {
            "创建": function(event) {
                page.ocean.Show()
            },
        }
		return {"button": ["创建"], "action": pane.Action}
    },
    initTarget: function(page, pane, form, output) {
        var river = ""
        pane.Listen = {
            river: function(value, old) {
                river = value, pane.Show()
            },
        }

        function fun(line, index, event, args, cbs) {
            var data = JSON.parse(line.text)
            var cmds = ["wave", river, data.node, data.group, data.index].concat(args)
            form.Run(cmds, cbs)
        }

        pane.Show = function() {
            output.Update(["flow", river], "text", ["text"], "index", false, fun)
        }

        pane.postion = page.Sync()
        pane.onscroll = function(event) {
            pane.postion.set({top: event.target.scrollTop, height: event.target.clientHeight, bottom: event.target.scrollHeight})
        }

        pane.Send = function(type, text, cb) {
            form.Run(["flow", river, type, text], function(msg) {
                output.Append(type, {text:text, index: msg.result[0]}, ["text"], "index", fun), typeof cb == "function" && cb()
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
            pane.style.display = (width==0 || height==0)? "none": "block"
            pane.style.width = width+"px"
            pane.style.height = height+"px"
            ui.first.style.width = (width-7)+"px"
            ui.first.style.height = (height-7)+"px"
        }

        pane.Clear = function(value) {
            ui.first.value = value || ""
        }
        return
    },
    initAction: function(page, pane, form, output) {
        var river = "", input = "", water = 0
        pane.Listen = {
            river: function(value, old) {
                river = value
            },
            source: function(value, old) {
                input = value, kit.Log(value)
            },
            storm: function(value, old) {
                water = value, pane.Show()
            },
        }
        pane.Show = function() {
            output.Update([river, water], "plugin", ["node", "name"], "index", false, function(line, index, event, args, cbs) {
                var cmds = [river, water, index].concat(args)

                // event.shiftKey? page.target.Send("field", JSON.stringify({
                //     componet_group: "index", componet_name: "river",
                //     cmds: ["wave", river, line.node, line.group, line.index], input: [{type: "input", data: {name: "hi", value: line.cmd}}]
                //
                event.shiftKey? page.target.Send("field", JSON.stringify({
                    name: line.name, view: line.view, init: line.init,
                    node: line.node, group: line.group, index: line.index,
                    inputs: line.inputs,
                })): form.Run(cmds, function(msg) {
                    event.ctrlKey && (msg.append && msg.append[0]?
                        page.target.Send("table", JSON.stringify(ctx.Table(msg))):
                        page.target.Send("text", msg.result.join("")))
                    cbs(msg)
                })
            })
        }
        pane.Action = {
            "恢复": function(event) {
                page.onlayout(event, page.conf.layout)
            },
            "放大": function(event) {
                page.onlayout(event, {action:300})
            },
            "最宽": function(event) {
                page.onlayout(event, {river:0, storm:0})
            },
            "最大": function(event) {
                page.onlayout(event, {river:0, action: -1, storm:0})
            },
        }
		return {"button": ["恢复", "放大", "最宽", "最大"], "action": pane.Action}
    },
    initStorm: function(page, pane, form, output) {
        var river = ""
        pane.Listen = {
            river: function(value, old) {
                river = value, pane.Show()
                pane.which.set("")
            },
        }
        pane.Show = function() {
            output.Update([river], "text", ["key", "count"], "key", true)
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
                if (event.key == "Enter") {
                    ui.name.nextSibling.click()
                }
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
                    page.storm.Show()
                    pane.Show()
                })
            }]}, {name: "list", view: ["list", "table"]},
        ]}])

        pane.Show = function() {
            pane.ShowDialog() && (table.innerHTML = "", ui.name.value = "nice", form.Run([river], function(msg) {
                kit.AppendTable(table, ctx.Table(msg), ["key", "user.route"], function(value, key, pod, i, tr, event) {
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
            }))
        }

        return [{"text": ["steam"]}]
    },
    init: function(page) {
        page.initField(page, function(init, pane, form) {
            var output = pane.querySelector("div.output")

            var list = [], last = -1
            output.Clear = function() {
                output.innerHTML = "", list = [], last = -1
            }
            output.Append = function(type, line, key, which, cb) {
                var index = list.length
                type = line.type || type
                var ui = page.View(output, type, line, key, function(event, cmds, cbs) {
                    output.Select(index), pane.which.set(line[which])
                    typeof cb == "function" && cb(line, index, event, cmds, cbs)
                })
                if (type == "table") {
                    kit.OrderTable(ui.last)
                }
                // if (type == "field") {
                //     kit.OrderForm(page, ui.last, ui.form, ui.table, ui.code)
                // }
                list.push(ui.last)
                pane.scrollBy(0, pane.scrollHeight)
                return ui
            }
            output.Select = function(index) {
                -1 < last && last < list.length && (list[last].className = "item")
                last = index, list[index].className = "item select"
            }
            output.Update = function(cmds, type, key, which, first, cb) {
                output.Clear(), form.Runs(cmds, function(line, index, msg) {
                    var ui = output.Append(type, line, key, which, cb)
                    first && index == 0 && ui.first.click()
                })
            }

            if (typeof init == "function") {
                var conf = init(page, pane, form, output)
                if (conf && conf["button"]) {
                    var buttons = []
                    conf.button.forEach(function(value, index) {
                        buttons.push({"button": [value, function(event) {
                            typeof conf["action"] == "function" && conf["action"](value, event)
                            typeof conf["action"] == "object" && conf["action"][value](event)
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
    },
})
