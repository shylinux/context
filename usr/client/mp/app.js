App({
    log: function(type, args) {
        switch (type) {
            case "info":
                console.log(args)
                break
            default:
                console.log(type, args)
        }
    },
    toast: function(text) {
        wx.showToast(text)
    },
    sheet: function(list, cb) {
        wx.showActionSheet({itemList: list, success(res) {
            typeof cb == "function" && cb(list[res.tapIndex])
        }})
    },
    confirm: function(content, confirm, cancel) {
        wx.showModal({
            title: "context", content: content, success: function(res) {
                res.confirm && typeof confirm == "function" && confirm()
                res.cancel && typeof cancel == "function" && cancel()
            }
        })
    },
    place: function(cb) {
        var app = this
        wx.authorize({scope: "scope.userLocation"})

        wx.chooseLocation({success: function(res) {
            app.log(res)
            typeof cb == "function" && cb(res)
        }})
    },
    stoprefresh: function(cb) {
        wx.stopPullDownRefresh()
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
                app.log(what)
                typeof done == "function" && done(res.data)
            },
            fail: function(res) {
                what.res = res
                app.log(what)
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
    table: function(res, cb) {
        if (res.append) {
            for (var i = 0; i < res[res.append[0]].length; i++) {
                var obj = {}
                var line = []
                for (var j = 0; j < res.append.length; j++) {
                    line.push(res[res.append[j]][i])
                    obj[res.append[j]] = res[res.append[j]][i]
                }
                typeof cb == "function" && cb(i, obj, line)
            }
        }
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
                        var view = JSON.parse(res["view"][i] || "{}")
                        var data = JSON.parse(res["data"][i] || "[]")
                        data.unshift({"name": "model", "type": "text", "value": res["name"][i]})
                        data.unshift({"name": "name", "type": "text"})
                        if (view.edit) {
                            for (var j = 0; j < data.length; j++) {
                                data[j].view = view.edit[data[j].name]
                            }
                        }

                        app.data.model[res["name"][i]] = {name: res["name"][i], data: data, view: view}
                    }
                    typeof cb == "function" && cb(app.data.model)
                })
                break
            case "list":
                if (app.data.length > 0 && app.data.list.length > 0) {
                    typeof cb == "function" && cb(app.data.list)
                    return
                }

                var cmd = {"cmd": ["note", "show", "username", "username", "full"]}

                app.command(cmd, function(res) {
                    if (!res || !res.append) {
                        return
                    }

                    var list = []
                    var ncol = res.append.length
                    var nrow = res[res.append[0]].length
                    for (var i = 0; i < nrow; i++) {
                        var value = JSON.parse(res["value"][i] || "[]")
                        value.unshift({"type": "text", "name": "model", "value": res["model"][i]})
                        value.unshift({"type": "text", "name": "name", "value": res["name"][i]})
                        value.unshift({"type": "text", "name": "create_date", "value": res["create_time"][i].split(" ")[0].replace("-", "/").replace("-", "/")})

                        var line = {
                            create_time: res["create_time"][i],
                            model: res["model"][i], value: value,
                            view: JSON.parse(res["view"][i] || "{}"), data: {},
                        }

                        for (var v in line.view) {
                            var view = line.view[v]

                            var data = []
                            for (var k in view) {
                                if (k in line) {
                                    data.push({name: k, view: view[k], value: line[k]})
                                }
                            }
                            for (var j = 0; j < value.length; j++) {
                                var k = value[j]["name"]
                                if (((v == "edit") || (k in view)) && !(k in line)) {
                                    data.push({name: k, view: view[k] || "", value: value[j]["value"]})
                                }
                            }
                            line.data[v] = data
                        }

                        list.push(line)
                    }

                    app.data.list = list
                    typeof cb == "function" && cb(list)
                })
                break
        }
    },
})
