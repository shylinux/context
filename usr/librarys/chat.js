var page = Page({
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
            output.innerHTML = "", form.Runs(["river"], function(line, index, msg) {
                index == 0 && page.target.Show(line.key)
                kit.AppendChild(output, [{view: ["item", "div", line.name+"("+line.count+")"], click: function(event) {
                    page.target.Show(line.key)
                }}])
            })
        }
        pane.Show()
		return {"button": ["添加", "查找"], "action": function(value) {
            switch (value) {
                case "添加":
                    var name = prompt("name")
                    name && form.Run(["river", "create", name], pane.Show)
                    break
                case "查找":
                    page.ocean.Show()
                    break
            }
		}}
    },
    initTarget: function(page, pane, form, output) {
        var river = ""
        pane.Show = function(which) {
            which && river != which && (river = which, output.innerHTML = "", form.Runs(["river", "wave", river], function(line, index, msg) {
                kit.AppendChild(output, [{view: ["item", "div", line.text], click: function(event) {}}])
                pane.scrollBy(0,100)
            }))
        }
        pane.Send = function(type, text, cb) {
            form.Run(["river", "wave", river, type, text], function(msg) {
                kit.AppendChild(output, [{"text" :[text, "div"]}])
                pane.scrollBy(0,100)
                typeof cb == "function" && cb(msg)
            })
        }
        return [{"text": ["target"]}]
    },
    initSource: function(page, pane, form, output) {
        var ui = kit.AppendChild(pane, [{"view": ["input", "textarea"], "name": "input", "data": {"onkeyup": function(event){
            event.key == "Enter" && !event.shiftKey && page.target.Send("text", event.target.value, pane.Clear)
        }, "onkeydown": function(event) {
            event.key == "Enter" && !event.shiftKey && event.preventDefault()
        }}}])

        pane.Size = function(width, height) {
            pane.style.display = width==0? "none": "block"
            pane.style.width = width+"px"
            pane.style.height = height+"px"
            ui.input.style.width = (width-7)+"px"
            ui.input.style.height = (height-7)+"px"
        }

        pane.Clear = function() {
            ui.input = ""
        }
        return
    },
    initStorm: function(page, pane, form, output) {
        return [{"text": ["storm"]}]
    },
    initSteam: function(page, pane, form, output) {
        return [{"text": ["steam"]}]
    },

    range: function(sizes) {
        sizes = sizes || {}
        var width = document.body.offsetWidth
        var height = document.body.offsetHeight-80

        sizes.left == undefined && (sizes.left = page.river.offsetWidth-page.conf.border)
        sizes.right == undefined && (sizes.right = page.storm.offsetWidth-page.conf.border)
        sizes.middle = width - sizes.left - sizes.right-5*page.conf.border
        page.river.Size(sizes.left, height)
        page.storm.Size(sizes.right, height)

        if (sizes.top != undefined) {
            sizes.bottom = height-sizes.top-page.conf.border
        } else if (sizes.bottom != undefined) {
            sizes.top = height-sizes.bottom-page.conf.border
        } else {
            sizes.bottom = page.source.offsetHeight-page.conf.border
            sizes.top = height-sizes.bottom-page.conf.border
        }
        kit.Log(sizes)
        page.target.Size(sizes.middle, sizes.top)
        page.source.Size(sizes.middle, sizes.bottom)
    },

    init: function(exp) {
        var page = this
        page.eachField(page, function(init, pane, form) {
            var output = pane.querySelector("div.output")
            if (typeof init == "function") {
                var conf = init(page, pane, form, output)
                if (conf && conf["button"]) {
                    var buttons = []
                    conf.button.forEach(function(value, index) {
                        buttons.push({"button": [value, function(event) {
                            typeof conf["action"] == "function" && conf["action"](value, event)
                        }]})
                    })
                    kit.InsertChild(pane, output, "div", buttons)
                } else if (conf) {
                    kit.AppendChild(output, conf)
                }
            }
        })
        window.onresize = this.range
        this.range({left:160, bottom:60, right:160})
    },
    conf: {border: 4},
})
