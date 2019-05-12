var page = Page({
    initOcean: function(page, field, option) {
        page.panes.ocean = field
        ctx.Run(page, option.dataset, ["ocean"], function(msg) {
            kit.Log(msg.result)
        })
        return [{"text": ["ocean"]}]
    },
    initRiver: function(page, field, option, output) {
        page.panes.river = field
        page.panes.channel = output
        page.showRiver(page, option)
		return {"button": ["添加"], "action": function(value) {
			ctx.Run(page, option.dataset, ["river", "create", prompt("name")], function(msg) {
				page.showRiver(page, option)
			})
		}}
    },
    showRiver: function(page, option) {
        page.panes.channel.innerHTML = ""
        page.getRiver(page, option, function(line, index) {
            page.conf.river = page.conf.river || page.showTarget(page, option, line.key) || line.key
            kit.AppendChild(page.panes.channel, [{view: ["item", "div", line.name+"("+line.count+")"], click: function(event) {
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
        page.panes.target = field
        page.panes.output = field.querySelector("div.target.output")
        ctx.Run(page, option.dataset, ["river"], function(msg) {
            kit.Log(msg.result)
        })
        return [{"text": ["target"]}]
    },
    showTarget: function(page, option, id) {
        page.panes.output.innerHTML = ""
        page.getTarget(page, option, id, function(line, index) {
            kit.AppendChild(page.panes.output, [{"view": ["item", "div", line.text]}])
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
                kit.AppendChild(page.panes.output, [{"text" :[value, "div"]}])
                page.panes.output.scrollBy(0,100)
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
        page.panes.input = ui.input

        page.panes.source = field
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
        page.panes.storm = field
        ctx.Run(page, option.dataset, ["river"], function(msg) {
            kit.Log(msg.result)
        })
        return [{"text": ["storm"]}]
    },
    initSteam: function(page, field, option) {
        page.panes.steam = field
        ctx.Run(page, option.dataset, ["river"], function(msg) {
            kit.Log(msg.result)
        })
        return [{"text": ["steam"]}]
    },

    panes: {},
    range: function(sizes) {
        sizes = sizes || {}
        var width = document.body.offsetWidth
        var river_width = page.panes.river.offsetWidth
        var storm_width = page.panes.storm.offsetWidth
        var source_width = page.panes.source.offsetWidth
        var source_height = page.panes.source.offsetHeight
        var height = document.body.offsetHeight-80


        page.panes.river.style.height = height+"px"
        page.panes.storm.style.height = height+"px"
        page.panes.target.style.height = (height-source_height)+"px"
        if (sizes.left != undefined) {
            if (sizes.left == 0) {
                page.panes.river.style.display = "none"
            } else {
                page.panes.river.style.display = "block"
                page.panes.river.style.width = sizes.left+"px"
            }
        }
        if (sizes.right != undefined) {
            if (sizes.right == 0) {
                page.panes.storm.style.display = "none"
            } else {
                page.panes.storm.style.display = "block"
                page.panes.storm.style.width = sizes.right+"px"
            }
        }
        if (sizes.middle != undefined) {
            page.panes.source.style.height = sizes.middle+"px"
            page.panes.target.style.height = (height-sizes.middle-4)+"px"
            page.panes.output.style.height = (height-sizes.middle-8)+"px"
            page.panes.input.style.height = (sizes.middle-7)+"px"
        } else {
            var source_height = page.panes.source.offsetHeight-10
            page.panes.input.style.height = source_height+"px"
            page.panes.output.style.height = source_height+"px"
        }

        var source_width = page.panes.source.offsetWidth-10
        page.panes.input.style.width = source_width+"px"
        page.panes.output.style.width = source_width+"px"
    },

    init: function(exp) {
        var page = this

        document.querySelectorAll("body>fieldset").forEach(function(field) {
            var option = field.querySelector("form.option")
            var output = field.querySelector("div.output")

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
