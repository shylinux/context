var page = Page({
    conf: {border: 4, banner: 105},
    size: function(event, sizes) {
        sizes = sizes || {}
        var width = document.body.offsetWidth
        var height = document.body.offsetHeight-page.conf.banner

        sizes.river == undefined && (sizes.river = page.river.offsetWidth-page.conf.border)
        sizes.storm == undefined && (sizes.storm = page.storm.offsetWidth-page.conf.border)
        sizes.width = width - sizes.river - sizes.storm-5*page.conf.border
        page.river.Size(sizes.river, height)
        page.storm.Size(sizes.storm, height)

        sizes.action == undefined && (sizes.action = page.action.offsetHeight-page.conf.border)
        sizes.source == undefined && (sizes.source = page.source.offsetHeight-page.conf.border)
        sizes.target = height - sizes.action - sizes.source - 2*page.conf.border
        page.target.Size(sizes.width, sizes.target)
        page.source.Size(sizes.width, sizes.source)
        page.action.Size(sizes.width, sizes.action)
    },

    initOcean: function(page, pane, form, output) {
        var table = kit.AppendChild(output, "table")
        pane.Show = function() {
            pane.ShowDialog() && (table.innerHTML = "", form.Run(["ocean"], function(msg) {
                kit.AppendTable(table, ctx.Table(msg), ["key", "user.route"])
            }))
        }
        return {"button": ["关闭"], "action": function(event) {
            pane.Show()
        }}
    },
    initRiver: function(page, pane, form, output) {
        pane.Show = function() {
            output.Update(["river"], "text", ["name", "count"], "key", true)
        }
        pane.Show()
        pane.Action = {
            "添加": function(event) {
                var name = prompt("name")
                name && form.Run(["river", "create", name], pane.Show)
            },
            "查找": function(event) {
                page.ocean.Show()
            },
        }
		return {"button": ["添加", "查找"], "action": pane.Action}
    },
    initTarget: function(page, pane, form, output) {
        var river = ""
        pane.Listen = {
            river: function(value, old) {
                river = value, pane.Show()
            },
        }

        pane.Show = function() {
            output.Update(["river", "wave", river], "text", ["text"], "index")
        }

        pane.postion = page.Sync()
        pane.onscroll = function(event) {
            pane.postion.set({top: event.target.scrollTop, height: event.target.clientHeight, bottom: event.target.scrollHeight})
        }

        pane.Send = function(type, text, cb) {
            form.Run(["river", "wave", river, type, text], function(msg) {
                output.Append(type, {text:text, index: msg.result[0]}, ["text"], "index"), typeof cb == "function" && cb()
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
            pane.style.display = width==0? "none": "block"
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
            output.Update(["river", "tool", river, water], "text", ["cmd"], "cmd", false, function(line, index) {
                form.Run(["river", "tool", river, water, index], function(msg) {
                    msg.append && msg.append[0]?
                        page.target.Send("table", JSON.stringify(ctx.Table(msg))):
                        page.target.Send("text", msg.result.join(""))
                })
            })
        }
        pane.Action = {
            "添加": function(event) {
                var name = prompt("name")
                name && form.Run(["river", "tool", river, water, "add", name], pane.Show)
            },
            "查找": function(event) {
                page.ocean.Show()
            },
        }
		return {"button": ["添加", "查找"], "action": pane.Action}
    },
    initStorm: function(page, pane, form, output) {
        var river = ""
        pane.Listen = {
            river: function(value, old) {
                river = value, pane.Show()
            },
        }
        pane.Show = function() {
            output.Update(["river", "tool", river], "text", ["key", "count"], "key", true)
        }
        pane.Action = {
            "添加": function(event) {
                var name = prompt("name")
                name && form.Run(["river", "tool", river, name, "pwd"], pane.Show)
            },
            "查找": function(event) {
                page.steam.Show()
            },
        }
		return {"button": ["添加", "查找"], "action": pane.Action}
    },
    initSteam: function(page, pane, form, output) {
        pane.Show = function() {
            pane.ShowDialog() && (table.innerHTML = "", form.Run(["ocean"], function(msg) {
                kit.AppendTable(table, ctx.Table(msg), ["key", "user.route"])
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
                var ui = kit.AppendChild(output, page.View(type, line, key, function(event) {
                    output.Select(index), pane.which.set(line[which])
                    typeof cb == "function" && cb(line, index)
                }))
                if (type == "table") {
                    kit.OrderTable(ui.last)
                }
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

        page.size(null, {river:160, source:60, action:60, storm:160})
    },
})
