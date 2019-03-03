const app = getApp()

Page({
    data: {
        focus: false,
        cmd: "",
        table: [],
        append: [],
        result: "",
    },
    onCommand: function(e) {
        var page = this
        var cmd = e.detail.value
        app.command({"cmd": cmd}, function(res) {
            if (res.append) {
                var table = []
                for (var i = 0; i < res[res.append[0]].length; i++) {
                    var line = []
                    for (var j = 0; j < res.append.length; j++) {
                        line.push(res[res.append[j]][i])
                    }
                    table.push(line)
                }
                page.setData({append: res.append, table: table})
            } else {
                page.setData({append: [], table: []})
            }
            page.setData({result: res.result? res.result.join("") :res})
            if (page.data.cmd) {
                return
            }
            app.command({"cmd": ["note", cmd, "flow", cmd]}, function(res) {})
        })
    },
    onLoad: function (options) {
        app.log("info", {page: "pages/index/index", options: options})

        var page = this
        app.load("model", function(model) {
            app.log("info", app.data.list[options.index])
            var cmd = app.data.list[options.index]? app.data.list[options.index].args["cmd"]: ""
            page.setData({
                model: model[options.model],
                value: app.data.list[options.index],
                view: model[options.model].view,
                cmd: cmd,
                focus: cmd? false: true,
            })
            if (cmd) {
                page.onCommand({detail:{value:cmd}})
            }
        })
    },
})
