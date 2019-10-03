ctx = context = {__proto__: kit,
    Event: Wrap(function(event, msg, proto) {
        event = event || document.createEvent("Event")
        if (event.msg && !msg) {return event.msg}

        var meta = arguments.callee
        var order = ++meta.order

        event.msg = msg = msg || {}, proto = proto || {}, msg.__proto__ = proto, proto.__proto__ = {
            Push: function(key, value) {
                msg.append || (msg.append = [])
                msg[key]? msg[key].push(value): (msg[key] = [value], msg.append.push(key))
                return msg
            },
            Echo: function(res) {
                res != null && res != undefined && (msg.result = (msg.result || []).concat(kit.Trans(res)))
                return msg
            },
            Result: function() {
                return msg.result? msg.result.join(""): ""
            },
            Results: function() {
                var s = msg.Result()
                s = s.replace(/</g, "&lt;")
                s = s.replace(/>/g, "&gt;")
                s = kit.Color(s)
                return s
            },
        }, msg.event = event
        kit.Log("event", order, event.type, proto.name, msg)
        return msg
    }, {order: 0}),
    Run: Wrap(function(dataset, cmd, cb) {
        var meta = arguments.callee
        var order = ++meta.order
        var msg = ctx.Event(event||document.createEvent("Event"), null, {name: "ctx.run"})

        var option = {"cmds": cmd}
        msg.option && msg.option.forEach(function(item) {
            msg.option[item] && (option[item] = msg.option[item])
        })
        for (var k in dataset) {
            option[k] = dataset[k].split(",")
        }

        msg.option = []
        for (var k in option) {
            msg.option.push(k)
            msg[k] = option[k]
        }
        msg.detail = ["run", order].concat(option.group).concat(option.names).concat(option.cmds)
        msg.Order = order

        kit.Log(msg.detail.concat([msg]))
        this.POST("", option, function(msg) {
            kit.Log("run", order, "result", msg.result? msg.result[0]: "", msg)
            typeof cb == "function" && cb(msg || {})
        }, msg)
    }, {order: 0}),
    Runs: function(form, cb) {
        var data = {}
        for (var key in form.dataset) {
            data[key] = form.dataset[key]
        }
        for (var i = 0; i < form.length; i++) {
            if (form[i].name) {
                data[form[i].name] = form[i].value
            }
        }
        this.Run(data, [], cb || form.ondaemon)
    },
    Table: function(msg, cb) {
        var ret = []
        if (!msg || !msg.append || !msg.append.length || !msg[msg.append[0]]) {
            return ret
        }

        var ncol = msg.append.length
        var nrow = msg[msg.append[0]].length
        for (var i = 0; i < nrow; i++) {
            var one = {}
            for (var j = 0; j < ncol; j++) {
                one[msg.append[j]] = msg[msg.append[j]][i]
            }
            ret.push(one)
        }

        var list = []
        typeof cb == "function" && ret.forEach(function(value, index, array) {
            var item = cb(value, index, array)
            item && list.push(item)
        })
        if (list.length > 0) {
            return list
        }
        return ret
    },
    Tables: function(msg, cb) {
        var ret = []
        if (!msg || !msg.append || !msg.append.length || !msg[msg.append[0]]) {
            return ret
        }
        ret.push(msg.append)

        var ncol = msg.append.length
        var nrow = msg[msg.append[0]].length
        for (var i = 0; i < nrow; i++) {
            var one = []
            for (var j = 0; j < ncol; j++) {
                one.push(msg[msg.append[j]][i])
            }
            ret.push(one)
        }

        var list = []
        typeof cb == "function" && ret.forEach(function(value, index, array) {
            var item = cb(value, index, array)
            item && list.push(item)
        })
        if (list.length > 0) {
            return list
        }
        return ret
    },
    Upload: function(form, file, cb, detail) {
        var data = new FormData()
        data.append("upload", file)
        for (var k in form) {
            data.append(k, form[k])
        }

        var xhr = new XMLHttpRequest()
        xhr.onload = function(event) {
            var msg = JSON.parse(xhr.responseText||'{"result":[]}')
            typeof cb == "function" && cb(event, msg)
        }

        xhr.onreadystatechange = function() {
            if (xhr.readyState != 4) {
                return
            }
            if (xhr.status != 200) {
                return
            }
        }

        xhr.upload.onprogress = function(event) {
            typeof detail == "function" && detail(event)
        }

        xhr.open("POST", "/upload", true)
        xhr.send(data)
    },
    Share: function(objs) {
        var args = this.Search()
        for (var k in objs) {
            args[k] = objs[k]
        }

        var as = []
        for (var k in args) {
            if (typeof args[k] == "object") {
                for (var i = 0; i < args[k].length; i++) {
                    as.push(k+"="+encodeURIComponent(args[k][i]));
                }
            } else {
                as.push(k+"="+encodeURIComponent(args[k]));
            }
        }
        var arg = as.join("&");
        return location.origin+location.pathname+"?"+arg
    },

    Search: function(key, value) {
        var args = {}
        var search = location.search.split("?")
        if (search.length > 1) {
            var searchs = search[1].split("&")
            for (var i = 0; i < searchs.length; i++) {
                var keys = searchs[i].split("=")
                if (keys[1] == "") {continue}
                args[keys[0]] = decodeURIComponent(keys[1])
            }
        }

        if (key == undefined) {
            return args
        } else if (typeof key == "object") {
            for (var k in key) {
                if (key[k] != undefined) {
                    args[k] = key[k]
                }
            }
        } else if (value == undefined) {
            return args[key] || this.Cookie(key)
        } else {
            args[key] = value
        }

        var arg = []
        for (var k in args) {
            arg.push(k+"="+encodeURIComponent(args[k]))
        }
        location.search = arg.join("&");
        return value
    },
    Cookie: function(key, value, expire) {
        if (key == undefined) {
            cs = {}
            cookies = document.cookie.split("; ")
            for (var i = 0; i < cookies.length; i++) {
                cookie = cookies[i].split("=")
                cs[cookie[0]] = cookie[1]
            }
            return cs
        }
        if (typeof key == "object") {
            for (var k in key) {
                document.cookie = k+"="+key[k];
            }
            return arguments.callee()
        }
        if (value == undefined) {
            var pattern = new RegExp(key+"=([^;]*);?")
            var result = pattern.exec(document.cookie)
            return result && result.length > 0? result[1]: ""
        }
        document.cookie = key+"="+value+";path=/"
        return arguments.callee(key)
    },
    POST: function(url, form, cb, msg) {
        var args = []
        for (var k in form) {
            if (form[k] instanceof Array) {
                for (i in form[k]) {
                    args.push(k+"="+encodeURIComponent(form[k][i]))
                }
            } else if (form[k] != undefined) {
                args.push(k+"="+encodeURIComponent(form[k]))
            }
        }

        var xhr = new XMLHttpRequest()
        xhr.onreadystatechange = function() {
            if (xhr.readyState != 4) {
                return
            }
            if (xhr.status != 200) {
                return
            }

            try {
                var res = JSON.parse(xhr.responseText||'[{"result":[]}]')
                res.length > 0 && res[0] && (res = res[0])
            } catch (e) {
                var res = {"result": [xhr.responseText]}
            }

            res.result && (msg.result = res.result)
            res.append && (msg.append = res.append) && res.append.forEach(function(item) {
                res[item] && (msg[item] = res[item])
            })

            if (msg.download_file) {
                window.open(msg.download_file.join(""))
            } else if (msg.page_redirect) {
                location.href = msg.page_redirect.join("")
            } else if (msg.page_refresh) {
                location.reload()
            }
            typeof cb == "function" && cb(msg || {})
        }

        xhr.open("POST", url)
        xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded")
        xhr.setRequestHeader("Accept", "application/json")
        xhr.send(args.join("&"))
    },
    WSS: Wrap(function(cb, onerror, onclose) {
        var meta = arguments.callee
        var s = new WebSocket(location.protocol.replace("http", "ws")+"//"+location.host+"/wss?wssid="+(page.wssid||""))
        s.onopen = function(event) {
            kit.Log("wss", "open")
            page.ontoast("wss open")
        }
        s.onmessage = function(event) {
            var order = ++meta.order
            try {
                var msg = JSON.parse(event.data||'{}')
            } catch (e) {
                var msg = {"result": [event.data]}
            }

            // Event入口 -1.0
            msg = ctx.Event(event, msg, {
                name: document.title,
                Order: order,
                Reply: function(msg) {
                    kit.Log(["wss", order, "result"].concat(msg.result).concat([msg]))
                    delete(msg.event), s.send(JSON.stringify(msg))
                },
            })

            // try {
                kit.Log(["wss", order].concat(msg.detail).concat([msg]))
                typeof cb == "function" && cb(msg)
            // } catch (e) {
            //     msg.Reply(kit.Log("err", e))
            // }
        }
        s.onerror = function(event) {
            kit.Log("wss", "error", event)
            typeof onerror == "function" && onerror(event)
        }
        s.onclose = function(event) {
            kit.Log("wss", "close")
            page.ontoast("wss close")
            typeof onclose == "function" && onclose(event)
        }
        return s
    }, {order: 0}),
}
