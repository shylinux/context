App({
    log: function(type, args) {
        console[type](args)
    },
    toast: function(text) {
        wx.showToast()
    },
    place: function(cb) {
        var app = this
        wx.authorize({scope: "scope.userLocation"})

        wx.chooseLocation({success: function(res) {
            app.log("info", res)
            typeof cb == "function" && cb(res)
        }})
    },
    navigate: function(page, args) {
        if (!page) {
            wx.navigateBack()
            return
        }
        var list = []
        for (var k in args) {
            list.push(k+"="+args[k])
        }

        wx.navigateTo({url:"/pages/"+page+"/"+page + (list.length>0? "?"+list.join("&"): "")})
    },

    request: function(data, done, fail) {
        var app = this
        data = data || {}
        data.sessid = app.sessid || ""

        var what = {method: "POST", url: "https://shylinux.com/chat/mp", data: data,
            success: function(res) {
                what.res = res
                app.log("info", what)
                typeof done == "function" && done(res.data)
            },
            fail: function(res) {
                what.res = res
                app.log("info", what)
                typeof done == "function" && done(res.data)
            },
        }

        wx.request(what)
    },

    sessid: "",
    userInfo: {},
    login: function(cb) {
        var app = this
        if (app.sessid) {
            typeof cb == "function" && cb(app.userInfo)
            return
        }

        wx.login({success: function(res) {
            app.request({code: res.code}, function(sessid) {
                app.sessid = sessid

                wx.getSetting({success: function(res) {
                    if (res.authSetting['scope.userInfo']) {
                        wx.getUserInfo({success: function(res) {
                            app.userInfo = res.userInfo
                            app.request(res, function() {
                                typeof cb == "function" && cb(app.userInfo)
                            })
                        }})
                    }
                }})
            })
        }})
    },

    command: function(args, cb) {
        var app = this
        var cmd = args["cmd"]
        if (cmd[0] == "note") {
            cmd = ["context", "ssh", "sh", "node", "note", "context", "mdb"].concat(args["cmd"])
        }

        app.login(function(userInfo) {
            app.request({cmd: cmd}, function(res) {
                app.toast("ok")
                typeof cb == "function" && cb(res)
            })
        })
    },

    model: {},
    data: {model: {}, list: []},
    load: function(type, cb) {
        var app = this
        switch (type) {
            case "model":
                if (app.data.length > 0 && app.data.model.length > 0) {
                    typeof cb == "function" && cb(app.data.model)
                    return
                }

                var cmd = {"cmd": ["note", type]}
                if (type == "note") {
                    cmd.cmd.push("note")
                }

                app.command(cmd, function(res) {
                    var ncol = res.append.length
                    var nrow = res[res.append[0]].length
                    for (var i = 0; i < nrow; i++) {
                        app.data.model[res["name"][i]] = {
                            name: res["name"][i],
                            data: JSON.parse(res["data"][i] || "[]"),
                            view: JSON.parse(res["view"][i] || "{}"),
                        }
                    }
                    typeof cb == "function" && cb(app.data.model)
                })
                break
            case "list":
                if (app.data.length > 0 && app.data.list.length > 0) {
                    typeof cb == "function" && cb(app.data.list)
                    return
                }

                var cmd = {"cmd": ["note", "search", "username"]}

                app.command(cmd, function(res) {
                    if (!res || !res.append) {
                        return
                    }

                    var ncol = res.append.length
                    var nrow = res[res.append[0]].length
                    for (var i = 0; i < nrow; i++) {
                        var args = {}
                        var value = JSON.parse(res["value"][i] || "[]")
                        for (var j = 0; j < value.length; j++) {
                            args[value[j].name] = value[j].value
                        }

                        app.data.list.push({
                            create_date:  res["create_time"][i].split(" ")[0].replace("-", "/").replace("-", "/"),
                            create_time: res["create_time"][i],
                            model: res["model"][i],
                            name: res["name"][i],
                            value: value,
                            args: args,
                            view: JSON.parse(res["view"][i] || "{}"),
                        })
                    }
                    typeof cb == "function" && cb(app.data.list)
                })
                break
        }
    },

    onLaunch: function () {},
})
