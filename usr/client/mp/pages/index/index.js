const app = getApp()

Page({
    data: {
        nodes: [
            ["note", "shy"],
            ["ctx", "cmd"],
            ["note", "show"],
        ],
        shows: [0, 0, 0],
        ctx: "", cmd: "", focus: false,
        append: [], table: [], result: "",
    },
    getPod(cb) {
        var page = this
        app.command({"cmd": ["context", "ssh", "remote"]}, function(res) {
            var pod = [""]
            app.table(res, function(i, obj, line) {
                pod.push(obj.key)
            })
            page.data.nodes[0] = pod
            page.data.shows[0] = 0
            page.setData({nodes: page.data.nodes, shows: page.data.shows})
            typeof cb == "function" && cb(pod)
        })
    },
    getCtx(pod, cb) {
        var page = this
        app.command({"cmd": ["context", "ssh", "sh", pod, "context", "ctx", "context"]}, function(res) {
            var ctx = []
            app.table(res, function(i, obj, line) {
                ctx.push(obj.names)
            })
            page.data.nodes[1] = ctx
            page.data.shows[1] = 0
            page.setData({nodes: page.data.nodes, shows: page.data.shows})
            typeof cb == "function" && cb(ctx)
        })
    },
    getCmd(pod, ctx, cb) {
        var page = this
        app.command({"cmd": ["context", "ssh", "sh", pod, "context", ctx, "command", "all"]}, function(res) {
            var cmd = [""]
            app.table(res, function(i, obj, line) {
                cmd.push(obj.key)
            })
            page.data.nodes[2] = cmd
            page.data.shows[2] = 0
            page.setData({nodes: page.data.nodes, shows: page.data.shows})
            page.setData({ctx: "context ssh sh node '"+pod+"' context "+ctx+" "+cmd[0]})
            typeof cb == "function" && cb(cmd)
        })
    },
    onChange: function(e) {
        var column = e.detail.column
        var value = e.detail.value
        var page = this
        page.data.shows[column] = value

        var pod = page.data.nodes[0][page.data.shows[0]]
        var ctx = page.data.nodes[1][page.data.shows[1]]
        var cmd = page.data.nodes[2][page.data.shows[2]]
        switch (column) {
            case 0:
                page.getCtx(pod, function(ctx) {
                    page.getCmd(pod, ctx[0], function(cmd) {
                        this.onCommand({detail:{value: ""}})
                    })
                })
                break
            case 1:
                page.getCmd(pod, ctx, function(cmd) {
                    this.onCommand({detail:{value: ""}})
                })
                break
            case 2:
                page.setData({ctx: "context ssh sh node '"+pod+"' context "+ctx+" "+cmd})
                this.onCommand({detail:{value: ""}})
                break
        }
    },
    onCommand: function(e) {
        var cmd = e.detail.value
        var page = this
        app.command({"cmd": ["source", page.data.ctx, cmd]}, function(res) {
            var table = []
            app.table(res, function(i, obj, line) {
                table.push(line)
            })

            page.setData({append: res.append || [], table: table, result: res.result? res.result.join("") :res})
            if (page.data.cmd) {
                return
            }
            app.command({"cmd": ["note", cmd, "proxy", page.data.ctx, cmd]})
        })
    },
    onLoad: function (options) {
        app.log({page: this.route, options: options})
        var data = app.data.list[options.index]
        var cmd = ""
        var page = this
        page.getPod(function(pod) {
            page.getCtx(pod[0], function(ctx) {
                page.getCmd(pod[0], ctx[0], function(cmd) {
                })
            })
        })

        if (data) {
            for (var i = 0; i < data.value.length; i++) {
                if (data.value[i].name == "cmd") {
                    cmd = data.value[i].value
                    this.onCommand({detail:{value: cmd}})
                }
            }
        }

        this.setData({cmd: cmd, focus: cmd? false: true})
    },
})
