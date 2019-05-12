var pane = {}
var page = Page({
    pane: pane,
    initOcean: function(page, field, option, output) {
        return {"button": ["关闭"], "action": function(event) {
            pane.ocean.showDialog()
        }}
    },

    initRiver: function(page, field, option, output) {
        pane.channel = output
        page.showRiver(page, option)
		return {"button": ["添加", "查找"], "action": function(value) {
            switch (value) {
                case "添加":
                    var name = prompt("name")
                    if (name) {
                        ctx.Run(page, option.dataset, ["river", "create", name], function(msg) {
                            page.showRiver(page, option)
                        })
                    }
                    break
                case "查找":
                    kit.showDialog(pane.ocean)
                    break
            }
		}}
    },
    showRiver: function(page, option) {
        pane.channel.innerHTML = ""
        page.getRiver(page, option, function(line, index) {
            page.conf.river = page.conf.river || page.showTarget(page, option, line.key) || line.key
            kit.AppendChild(pane.channel, [{view: ["item", "div", line.name+"("+line.count+")"], click: function(event) {
                if (page.conf.river == line.key) {
                    return
                }
                page.conf.river = line.key
                page.showTarget(page, option, line.key)
            }}])
        })
    },
    getRiver: function(page, option, cb) {
        ctx.Run(page, option.dataset, ["river"], function(msg) {
            ctx.Table(msg, function(line, index) {
                cb(line, index)
            })
        })
    },
    initTarget: function(page, field, option) {
        pane.output = field.querySelector("div.target.output")
        ctx.Run(page, option.dataset, ["river"], function(msg) {
            kit.Log(msg.result)
        })
        return [{"text": ["target"]}]
    },
    showTarget: function(page, option, id) {
        pane.output.innerHTML = ""
        page.getTarget(page, option, id, function(line, index) {
            kit.AppendChild(pane.output, [{"view": ["item", "div", line.text]}])
        })
    },
    getTarget: function(page, option, id, cb) {
        ctx.Run(page, option.dataset, ["river", "wave", id], function(msg) {
            ctx.Table(msg, function(line, index) {
                cb(line, index)
            })
        })
    },
    initSource: function(page, field, option) {
        var ui = kit.AppendChild(option, [{"view": ["input", "textarea"], "name": "input", "data": {"onkeyup": function(event){
            if (event.key == "Enter" && !event.shiftKey) {
                var value = event.target.value
                kit.AppendChild(pane.output, [{"text" :[value, "div"]}])
                pane.output.scrollBy(0,100)
                // event.target.value = ""
                ctx.Run(page, option.dataset, ["river", "wave", page.conf.river, "text", value], function(msg) {
                    kit.Log(msg.result)
                })
            }
        }, "onkeydown": function(event) {
            if (event.key == "Enter" && !event.shiftKey) {
                event.preventDefault()
            }
        }}}])
        pane.input = ui.input

        ctx.Run(page, option.dataset, ["river"], function(msg) {
            msg.Table(function(index, line) {
                console.log(index)
                console.log(line)
            })
            kit.Log(msg.result)
        })
        return
    },
    initStorm: function(page, field, option) {
        ctx.Run(page, option.dataset, ["river"], function(msg) {
            kit.Log(msg.result)
        })
        return [{"text": ["storm"]}]
    },
    initSteam: function(page, field, option) {
        ctx.Run(page, option.dataset, ["river"], function(msg) {
            kit.Log(msg.result)
        })
        return [{"text": ["steam"]}]
    },

    panes: {},
    range: function(sizes) {
        sizes = sizes || {}
        var width = document.body.offsetWidth
        var river_width = pane.river.offsetWidth
        var storm_width = pane.storm.offsetWidth
        var source_width = pane.source.offsetWidth
        var source_height = pane.source.offsetHeight
        var height = document.body.offsetHeight-80


        pane.river.style.height = height+"px"
        pane.storm.style.height = height+"px"
        pane.target.style.height = (height-source_height)+"px"
        if (sizes.left != undefined) {
            if (sizes.left == 0) {
                pane.river.style.display = "none"
            } else {
                pane.river.style.display = "block"
                pane.river.style.width = sizes.left+"px"
            }
        }
        if (sizes.right != undefined) {
            if (sizes.right == 0) {
                pane.storm.style.display = "none"
            } else {
                pane.storm.style.display = "block"
                pane.storm.style.width = sizes.right+"px"
            }
        }
        if (sizes.middle != undefined) {
            pane.source.style.height = sizes.middle+"px"
            pane.target.style.height = (height-sizes.middle-4)+"px"
            pane.output.style.height = (height-sizes.middle-8)+"px"
            pane.input.style.height = (sizes.middle-7)+"px"
        } else {
            var source_height = pane.source.offsetHeight-10
            pane.input.style.height = source_height+"px"
            pane.output.style.height = source_height+"px"
        }

        var source_width = pane.source.offsetWidth-10
        pane.input.style.width = source_width+"px"
        pane.output.style.width = source_width+"px"
    },

    init: function(exp) {
        var page = this

        document.querySelectorAll("body>fieldset").forEach(function(field) {
            var option = field.querySelector("form.option")
            var output = field.querySelector("div.output")
            pane[option.dataset.componet_name] = field

            var init = page[field.dataset.init]
            if (typeof init == "function") {
                var conf = page[field.dataset.init](page, field, option, output)
                if (conf && conf["button"]) {
                    var buttons = []
                    conf.button.forEach(function(value, index) {
                        buttons.push({"button": [value, function(event) {
                            typeof conf["action"] == "function" && conf["action"](value, event)
                        }]})
                    })
                    kit.InsertChild(field, output, "div", buttons)
                } else if (conf) {
                    kit.AppendChild(output, conf)
                }
            }
        })
        window.onresize = this.range
        this.range({left:120, middle:60, right:120})
    },
    conf: {},
})
