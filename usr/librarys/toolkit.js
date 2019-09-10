kit = toolkit = {
    isMobile: navigator.userAgent.indexOf("Mobile") > -1,
    isWeiXin: navigator.userAgent.indexOf("MicroMessenger") > -1,
    isMacOSX: navigator.userAgent.indexOf("Mac OS X") > -1,
    isWindows: navigator.userAgent.indexOf("Windows") > -1,
    isIPhone: navigator.userAgent.indexOf("iPhone") > -1,
    isSpace: function(c) {
        return c == " " || c == "Enter"
    },
    History: {dir: [], pod: [], ctx: [], cmd: [], txt: [], key: [], lay: [],
        add: function(type, data) {
            var list = this[type] || []
            data && list.push({time: Date.now(), data: data})
            this[type] = list
            return list.length-1
        },
        get: function(type, index) {
            var list = this[type] || []
            var len = list.length
            return index == undefined? this[type]: this[type][(index+len)%len]
        },
    },
    Log: function() {
        var args = []
        for (var i = 0; i < arguments.length; i++) {
            args.push(arguments[i])
        }
        console.log(arguments.length == 1? args[0]: args)
        return args
    },
    Delay: function(time, cb) {
        return setTimeout(cb, time)
    },

    ModifyView: function(which, args) {
        var height = document.body.clientHeight-4
        var width = document.body.clientWidth-4
        for (var k in args) {
            switch (k) {
                case "dialog":
                    var w = h = args[k]
                    if (typeof(args[k]) == "object") {
                        w = args[k][0]
                        h = args[k][1]
                    }
                    if (w > width) {
                        w = width
                    }
                    if (h > height) {
                        h = height
                    }

                    args["top"] = (height-h)/2
                    args["left"] = (width-w)/2
                    args["width"] = w
                    args["height"] = h
                    args[k] = undefined
                    break
                case "window":
                    var w = h = args[k]
                    if (typeof(args[k]) == "object") {
                        w = args[k][0]
                        h = args[k][1]
                    }

                    args["top"] = h/2
                    args["left"] = w/2
                    args["width"] = width-w-20
                    args["height"] = height-h-20
                    args[k] = undefined
                    break
            }
        }

        for (var k in args) {
            switch (k) {
                case "top":
                case "left":
                case "width":
                case "height":
                case "padding":
                    args[k] = args[k]+"px"
                    break
            }
        }
        return kit.ModifyNode(which, {style: args})
    },
    ModifyNode: function(which, html) {
        var node = typeof which == "string"? document.querySelector(which): which
        switch (typeof html) {
            case "string":
                node.innerHTML = html
                break
            case "object":
                for (var k in html) {
                    if (typeof html[k] == "object") {
                        for (var d in html[k]) {
                            node[k] && (node[k][d] = html[k][d])
                        }
                        continue
                    }
                    node[k] = html[k]
                }
                break
        }
        return node
    },
    CreateNode: function(element, html) {
        return this.ModifyNode(document.createElement(element), html)
    },
    AppendChild: function(parent, children, subs) {
        if (typeof children == "string") {
            var elm = this.CreateNode(children, subs)
            parent.append(elm)
            return elm
        }

        // 基本属性: name value inner
        // 基本样式: style className
        // 基本事件: dataset click
        //
        // 按键: button select
        // 输入: input password
        // 输出: label img row
        //
        // 树状结构: tree fork leaf
        // 普通视图: view text code
        // 加载文件: include require styles
        //
        // 基本结构: type data list

        var kit = this

        subs = subs || {}
        children.forEach(function(child, i) {
            if (!child) {
                return
            }
            child.data = child.data || {}
            child.type = child.type || "div"

            if (child.name) {
                child.data["name"] = child.name
            }
            if (child.value) {
                child.data["value"] = child.value
            }
            if (child.inner) {
                child.data["innerHTML"] = child.inner
            }
            if (typeof(child.style) == "object") {
                var str = []
                for (var k in child.style) {
                    str.push(k)
                    str.push(":")
                    str.push(child.style[k] + (typeof child.style[k] == "number"? "px": ""))
                    str.push(";")
                }
                child.data["style"] = str.join("")
            }
            if (child.className) {
                child.data["className"] = child.className
            }
            if (child.dataset) {
                child.data["dataset"] = child.dataset
            }
            if (child.click) {
                child.data["onclick"] = child.click
            }

            if (child.button) {
                child.type = "button"
                child.data["onclick"] = function(event) {
                    child.button[1](child.button[0], event)
                }
                child.data["innerText"] = child.button[0]
                child.name = child.name || child.button[0]

            } else if (child.select) {
                child.type = "select"
                child.name = child.select[0][0]
                child.list = child.select[0].map(function(value) {
                    return {type: "option", value: value, inner: value}
                })
                child.data["onchange"] = function(event) {
                    child.select[1](event.target.value, event)
                }

            } else if (child.input) {
                child.type = "input"
                child.data["name"] = child.input[0]
                // child.data["onkeyup"] = child.input[1]
                child.data["onkeydown"] = child.input[1]
                child.name = child.name || child.input[0]

            } else if (child.password) {
                child.type = "input"
                child.data["onkeyup"] = child.password[1]
                child.data["name"] = child.password[0]
                child.data["type"] = "password"
                child.name = child.name || child.password[0]

            } else if (child.label) {
                child.type = "label"
                child.data["innerText"] = child.label

            } else if (child.img) {
                child.type = "img"
                child.data["src"] = child.img[0]
                child.img.length > 1 && (child.data["onload"] = child.img[1])

            } else if (child.row) {
                child.type = "tr"
                child.list = child.row.map(function(item) {return {text: [item, "td"]}})

            } else if (child.tree) {
                child.type = "ul"
                child.list = child.tree

            } else if (child.fork) {
                child.type = "li"
                child.list = [
                    {"text": [child.fork[0], "div"], "click": (child.fork.length>2? child.fork[2]: "")},
                    {"type": "ul", "list": child.fork[1]},
                ]

            } else if (child.leaf) {
                child.type = "li"
                child.list = [{"text": [child.leaf[0], "div"]}]
                if (child.leaf.length > 1 && typeof child.leaf[1] == "function") {
                    child.data["onclick"] = function(event) {
                        child.leaf[1](event, node)
                    }
                }

            } else if (child.view) {
                child.data["className"] = child.view[0]
                child.type = child.view.length > 1? child.view[1]: "div"
                child.view.length > 2 && (child.data["innerHTML"] = child.view[2])
                child.view.length > 3 && (child.name = child.view[3])

            } else if (child.text) {
                child.data["innerText"] = child.text[0]
                child.type = child.text.length > 1? child.text[1]: "pre"
                child.text.length > 2 && (child.data["className"] = child.text[2])

            } else if (child.code) {
                child.type = "code"
                child.list = [{"type": "pre" ,"data": {"innerText": child.code[0]}, "name": child.code[1]}]
                child.code.length > 2 && (child.data["className"] = child.code[2])

            } else if (child.script) {
                child.type = "script"
                child.data.innerHTML = child.script

            } else if (child.include) {
                child.type = "script"
                child.data["src"] = child.include[0]
                child.data["type"] = "text/javascript"
                child.include.length > 1 && (child.data["onload"] = child.include[1])

            } else if (child.require) {
                child.data["href"] = child.require[0]
                child.data["rel"] = child.require.length > 1? child.require[1]: "stylesheet"
                child.data["type"] = child.require.length > 2? child.require[2]: "text/css"
                child.type = "link"

            } else if (child.styles) {
                var str = []
                for (var key in child.styles) {
                    str.push(key)
                    str.push(" {")
                    for (var k in child.styles[key]) {
                        str.push(k)
                        str.push(":")
                        str.push(child.styles[key][k] + (typeof child.styles[key][k] == "number"? "px": ""))
                        str.push(";")
                    }
                    str.push("}\n")
                }
                child.data["innerHTML"] = str.join("")
                child.data["type"] = "text/css"
                child.type = "style"

            }

            var node = kit.CreateNode(child.type, child.data)
            child.list && kit.AppendChild(node, child.list, subs)
            child.name && (subs[child.name] = node)
            subs.first || (subs.first = node)
            subs.last, subs.last = node
            parent.append(node)
        })
        return subs
    },
    AppendChilds: function(parent, children, subs) {
        return parent.innerHTML = "", this.AppendChild(parent, children, subs)
    },
    InsertChild: function (parent, position, element, children) {
        var elm = this.CreateNode(element)
        this.AppendChild(elm, children)
        return parent.insertBefore(elm, position || parent.firstElementChild)
    },

    AppendAction: function(parent, list, cb) {
        var result = []
        list.forEach(function(item, index) {
            if (item == "") {
                result.push({view: ["space"]})
            } else if (typeof item == "string") {
                result.push({button: [item, cb]})
            } else if (item.forEach) {
                result.push({select: [item, cb]})
            } else {
                result.push(item)
            }
        })
        return kit.AppendChild(parent, result)
    },
    AppendStatus: function(parent, list, cb) {
        var result = []
        list.forEach(function(item, index) {
            if (item == "") {
                result.push({view: ["space"]})
            } else if (typeof item == "string") {
                result.push({button: [item, cb]})
            } else if (item.forEach) {
                result.push({select: [item, cb]})
            } else {
                result.push(item)
            }
        })
        return kit.AppendChild(parent, result)
    },
    AppendTable: function(table, data, fields, cb) {
        if (!data || !fields) {
            return
        }
        var kit = this
        var tr = kit.AppendChild(table, "tr")
        fields.forEach(function(key, j) {
            var td = kit.AppendChild(tr, "th", kit.Color(key))
        })
        data.forEach(function(row, i) {
            var tr = kit.AppendChild(table, "tr", {className: "normal"})
            tr.Meta = row
            fields.forEach(function(key, j) {
                var td = kit.AppendChild(tr, "td", kit.Color(row[key]))
                if (row[key].startsWith("http")) {
                    td.innerHTML = "<a href='"+row[key]+"' target='_blank'>"+row[key]+"</a>"
                }

                if (typeof cb == "function") {
                    td.onclick = function(event) {
                        cb(row[key], key, row, i, tr, event)
                    }
                }
            })
        })
        return table
    },
    RangeTable: function(table, index, sort_asc) {
        var list = table.querySelectorAll("tr")

        var is_time = true, is_number = true
        for (var i = 1; i < list.length; i++) {
            var text = list[i].childNodes[index].innerText
            var value = Date.parse(text)
            if (!(value > 0)) {
                is_time = false
            }

            var value = parseInt(text)
            if (text != "" && !(value >= 0 || value <= 0)) {
                is_number = false
            }
        }

        var num_list = [], new_list = []
        for (var i = 1; i < list.length; i++) {
            var text = list[i].childNodes[index].innerText
            if (is_time) {
                num_list.push(Date.parse(text))
            } else if (is_number) {
                num_list.push(parseInt(text) || 0)
            } else {
                num_list.push(text)
            }
            new_list.push(list[i])
        }

        for (var i = 0; i < new_list.length; i++) {
            for (var j = i+1; j < new_list.length; j++) {
                if (sort_asc? num_list[i] < num_list[j]: num_list[i] > num_list[j]) {
                    var temp = num_list[i]
                    num_list[i] = num_list[j]
                    num_list[j] = temp
                    var temp = new_list[i]
                    new_list[i] = new_list[j]
                    new_list[j] = temp
                }
            }
            new_list[i].parentElement && new_list[i].parentElement.removeChild(new_list[i])
            table.appendChild(new_list[i])
        }
    },
    OrderTable: function(table, field, cb) {
        if (!table) {return}
        table.onclick = function(event) {
            var index = 0
            var target = event.target
            var dataset = target.dataset
            var head = target.parentElement.parentElement.querySelector("tr")
            kit.Selector(table, "tr.select", function(item) {item.className = ""})
            kit.Selector(table, "td.select", function(item) {item.className = ""})
            kit.Selector(table, "tr", function(item, i) {item == target.parentElement && (index = i)})

            target.parentElement.childNodes.forEach(function(item, i) {
                if (item != target) {return}

                if (target.tagName == "TH") {
                    dataset["sort_asc"] = (dataset["sort_asc"] == "1") ? 0: 1
                    kit.RangeTable(table, i, dataset["sort_asc"] == "1")
                    return
                }
                var name = head.childNodes[i].innerText
                if (name.startsWith(field)) {
                    item.className = "select"
                    item.parentElement.className = "select"
                    typeof cb == "function" && cb(event, item.innerText, name, item.parentNode.Meta, index)
                }
                kit.CopyText()
            })
        }
    },

    OrderCode: function(code) {
        if (!code) {return}

        var kit = this
        code.onclick = function(event) {
            kit.CopyText()
        }
    },
    OrderLink: function(link) {
        link.target = "_blank"
    },
    OrderText: function(pane, text) {
        text.querySelectorAll("a").forEach(function(value, index, array) {
            kit.OrderLink(value, pane)
        })
        text.querySelectorAll("table").forEach(function(value, index, array) {
            kit.OrderTable(value)
        })

        var i = 0, j = 0, k = 0
        var h0 = [], h2 = [], h3 = []
        text.querySelectorAll("h2,h3,h4").forEach(function(value, index, array) {
            var id = ""
            var text = value.innerText
            var ratio = parseInt(value.offsetTop/pane.scrollHeight*100)
            if (value.tagName == "H2") {
                j=k=0
                h2 = []
                id = ++i+"."
                text = id+" "+text
                h0.push({"fork": [text+" ("+ratio+"%)", h2, function(event) {
                    location.hash = id
                }]})
            } else if (value.tagName == "H3") {
                k=0
                h3 = []
                id = i+"."+(++j)
                text = id+" "+text
                h2.push({"fork": [text+" ("+ratio+"%)", h3, function(event) {
                    location.hash = id
                }]})
            } else if (value.tagName == "H4") {
                id = i+"."+j+"."+(++k)
                text = id+" "+text
                h3.push({"leaf": [text+" ("+ratio+"%)", function(event) {
                    location.hash = id
                }]})
            }
            value.innerText = text
            value.id = id
        })
        return h0

        text.querySelectorAll("table.wiki_list").forEach(function(value, index, array) {
            kit.OrderTable(value, "path", function(event) {
                var text = event.target.innerText
                ctx.Search({"class": text})
            })
        })
    },
    Position: function(which) {
        return (parseInt((which.scrollTop + which.clientHeight) / which.scrollHeight * 100)||0)+"%"
    },
    Color: function(s) {
        s = s.replace(/\033\[1m/g, "<span style='font-weight:bold'>")
        s = s.replace(/\033\[36m/g, "<span style='color:#0ff'>")
        s = s.replace(/\033\[33m/g, "<span style='color:#ff0'>")
        s = s.replace(/\033\[32m/g, "<span style='color:#0f0'>")
        s = s.replace(/\033\[32;1m/g, "<span style='color:#0f0'>")
        s = s.replace(/\033\[31m/g, "<span style='color:#f00'>")
        s = s.replace(/\033\[0m/g, "</span>")
        s = s.replace(/\033\[m/g, "</span>")
        return s
    },

    CopyText: function(text) {
        text = window.getSelection().toString()
        if (text == "") {return}
        kit.History.add("txt", text)
        document.execCommand("copy")
    },
    DelText: function(target, start, count) {
        target.value = target.value.substring(0, start)+target.value.substring(start+(count||target.value.length), target.value.length)
        target.setSelectionRange(start, start)
    },
    HitText: function(target, text) {
        var start = target.selectionStart
        for (var i = 1; i < text.length+1; i++) {
            var ch = text[text.length-i]
            if (target.value[start-i] != ch && kit.History.get("key", -i).data != ch) {
                return false
            }
        }
        return true
    },

    Selector: function(obj, item, cb) {
        var list = []
        obj.querySelectorAll(item).forEach(function(item, index, array) {
            if (typeof cb == "function") {
                var value = cb(item, index, array)
                value != undefined && list.push(value)
            } else {
                list.push(item)
            }
        })
        for (var i = list.length-1; i >= 0; i--) {
            if (list[i] == "") {
                list.pop()
            } else {
                break
            }
        }
        return list
    },
    Format: function(objs) {
        return JSON.stringify(objs)
         wa
    },
    List: function(obj, cb, interval, cbs) {
        if (interval) {
            function loop(i) {
                if (i >= obj.length) {typeof cbs == "function" && cbs(); return}
                typeof cb == "function" && cb(obj[i], i)
                setTimeout(function() {loop(i+1)}, interval)
            }
            obj.length > 0 && setTimeout(function() {loop(0)}, interval)
            return obj
        }
        var list = []
        for (var i = 0; i < obj.length; i++) {
            list.push(typeof cb == "function"? cb(obj[i], i): obj[i])
        }
        return list
    },
    Item: function(obj, cb) {
        var list = []
        for (var k in obj) {
            list.push(typeof cb == "function"? cb(k, obj[k]): k)
        }
        return list
    },
    alert: function(text) {
        alert(JSON.stringify(text))
    },
    prompt: function(text) {
        return prompt(text)
    },
    confirm: function(text) {
        return confirm(text)
    },
    reload: function() {
        location.reload()
    },

    distance: function(x0, y0, x1, y1) {
        return Math.sqrt(Math.pow(x1-x0, 2)+Math.pow(y1-y0, 2))
    },
    number: function(d, n) {
        var result = []
        while (d>0) {
            result.push(d % 10)
            d = parseInt(d / 10)
            n--
        }
        while (n > 0) {
            result.push("0")
            n--
        }
        result.reverse()
        return result.join("")
    },
    right: function(arg) {
        if (arg == "true") {
            return true
        }
        if (arg == "false") {
            return false
        }
        if (arg) {
            return true
        }
        return false
    },
    time: function(t, fmt) {
        var now = t? new Date(t): new Date()
        fmt = fmt || "%y-%m-%d %H:%M:%S"
        fmt = fmt.replace("%y", now.getFullYear())
        fmt = fmt.replace("%m", kit.number(now.getMonth()+1, 2))
        fmt = fmt.replace("%d", kit.number(now.getDate(), 2))
        fmt = fmt.replace("%H", kit.number(now.getHours(), 2))
        fmt = fmt.replace("%M", kit.number(now.getMinutes(), 2))
        fmt = fmt.replace("%S", kit.number(now.getSeconds(), 2))
        return fmt
    },
    format_date: function(arg) {
        var date = arg.getDate()
        if (date < 10) {
            date = "0"+date
        }
        var month = arg.getMonth()+1
        if (month < 10) {
            month = "0"+month
        }
        var hour = arg.getHours()
        if (hour < 10) {
            hour = "0"+hour
        }
        var minute = arg.getMinutes()
        if (minute < 10) {
            minute = "0"+minute
        }
        var second = arg.getSeconds()
        if (second < 10) {
            second = "0"+second
        }
        return arg.getFullYear()+"-"+month+"-"+date+" "+hour+":"+minute+":"+second
    },
    size: function(obj, width, height) {
        obj.style.width = width+"px"
        obj.style.height = height+"px"
    },
    _call: function(cb, arg) {
        var res
        switch (arg.length) {
            case 0: res = cb(); break
            case 1: res = cb(arg[0]); break
            case 2: res = cb(arg[0], arg[1]); break
            case 3: res = cb(arg[0], arg[1], arg[2]); break
            case 4: res = cb(arg[0], arg[1], arg[2], arg[3]); break
            case 5: res = cb(arg[0], arg[1], arg[2], arg[3], arg[4]); break
            case 6: res = cb(arg[0], arg[1], arg[2], arg[3], arg[4], arg[5]); break
            case 7: res = cb(arg[0], arg[1], arg[2], arg[3], arg[4], arg[5], arg[6]); break
        }
        return res || true
    },

    OrderForm: function(page, field, option, append, result) {
        if (!option) {return}
        option.ondaemon = option.ondaemon || function(msg) {
            append.innerHTML = ""
            msg && msg.append && kit.AppendTable(append, ctx.Table(msg), msg.append, function(value, key, row, index, event) {
                typeof option.daemon_action[key] == "function" && option.daemon_action[key](value, key, row, index, event)
            })
            result && (result.innerText = (msg && msg.result)? msg.result.join(""): "")
        }

        option.querySelectorAll("select").forEach(function(select, index, array) {
            select.onchange = select.onchange || function(event) {
                if (index == array.length-1) {
                    ctx.Runs(page, option)
                    return
                }
                if (array[index+1].type == "button") {
                    array[index+1].click()
                    return
                }
                array[index+1].focus()
            }
        })
        option.querySelectorAll("input").forEach(function(input, index, array) {
            switch (input.type) {
                case "button":
                    input.onclick = input.onclick || function(event) {
                        if (index == array.length-1) {
                            if (input.value == "login") {
                                ctx.Runs(page, option, function(msg) {
                                    if (document.referrer) {
                                        location.href = document.referrer
                                    } else {
                                        ctx.Search("group", "")
                                    }
                                })
                                return
                            }

                            ctx.Runs(page, option)
                            return
                        }
                        if (array[index+1].type == "button") {
                            array[index+1].click()
                            return
                        }
                        array[index+1].focus()
                    }
                default:
                    input.onkeyup = input.onkeyup || function(event) {
                        if (event.key != "Enter") {
                            return
                        }
                        if (index == array.length-1) {
                            ctx.Runs(page, option)
                            return
                        }
                        if (array[index+1].type == "button") {
                            array[index+1].click()
                            return
                        }
                        array[index+1].focus()
                    }
            }
        })
    },
}

function Editor(run, plugin, option, output, width, height, space, msg) {
    exports = ["dir", "path", "dir"]
    msg.append && kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), ctx.Table(msg), msg.append), exports[1], function(event, value, name, line) {
        page.Sync("plugin_"+exports[0]).set(plugin.onexport[exports[2]||""](value, name, line))
    });

    var args = [option.pod.value, option.dir.value]

    if (msg.file) {
        var action = kit.AppendAction(kit.AppendChild(output, [{view: ["action"]}]).last, [
            "追加", "提交", "取消",
        ], function(value, event) {
            switch (value) {
                case "追加":
                    run(event, args.concat(["dir_sed", "add"]))
                    break
                case "提交":
                    run(event, args.concat(["dir_sed", "put"]))
                    break
                case "取消":
                    break
            }
        })

        kit.AppendChild(output, [{view: ["edit", "table"], list: (msg.result||[]).map(function(value, index) {
            return {view: ["line", "tr"], list: [{view: ["num", "td", index+1]}, {view: ["txt", "td"], list: [{value: value, style: {width: width+"px"}, input: [value, function(event) {
                if (event.key == "Enter") {
                    field.Run(event, args.concat(["dir_sed", "set", index, event.target.value]))
                }
            }]}]}]}
        })}])
    }
}
function Canvas(plugin, option, output, width, height, space, msg) {
    var keys = [], data = {}, max = {}, nline = 0
    var nrow = msg[msg.append[0]].length
    var step = width / (nrow - 1)
    msg.append.forEach(function(key, index) {
        var list = []
        msg[key].forEach(function(value, index) {
            var v = parseInt(value)
            !isNaN(v) && (list.push((value.indexOf("-") == -1)? v: value), v > (max[key]||0) && (max[key] = v))
        })
        list.length == nrow && (keys.push(key), data[key] = list, nline++)
    })

    var conf = {
        font: "monospace", text: "hi", tool: "stroke", style: "black",
        type: "trend", shape: "drawText", means: "drawPoint",
        limits: {scale: 3, drawPoint: 1, drawPoly: 3},

        axies: {style: "black", width: 2},
        xlabel: {style: "red", width: 2, height: 5},
        plabel: {style: "red", font: "16px monospace", offset: 10, height: 20, length: 20},
        data: {style: "black", width: 1},

        mpoint: 10,
        play: 500,
    }

    var view = [], ps = [], point = [], now = {}, index = 0
    var trap = false, label = false

    var what = {
        reset: function(x, y) {
            canvas.resetTransform()
            canvas.setTransform(1, 0, 0, -1, space+(x||0), height+space-(y||0))
            canvas.strokeStyle = conf.data.style
            canvas.fillStyle = conf.data.style
            return what
        },
        clear: function() {
            var p0 = what.transform({x:-width, y:-height})
            var p1 = what.transform({x:2*width, y:2*height})
            canvas.clearRect(p0.x, p0.y, p1.x-p0.x, p1.y-p0.y)
            return what
        },

        move: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save(), what.clear().drawLine(meta)
            canvas.translate(p1.x-p0.x, p1.y-p0.y)
            what.drawData().drawView()
            meta.ps.length < 2 && canvas.restore()
        },
        scale: function(meta) {
            var ps = meta.ps
            var p0 = ps[0] || {x:0, y:0}
            var p1 = ps[1] || now
            var p2 = ps[2] || now

            if (ps.length > 1) {
                canvas.save(), what.clear()
                what.drawLine({ps: [p0, {x: p1.x, y: p0.y}]})
                what.drawLine({ps: [{x: p1.x, y: p0.y}, p1]})
                what.drawLine({ps: [p0, {x: p2.x, y: p0.y}]})
                what.drawLine({ps: [{x: p2.x, y: p0.y}, p2]})
                canvas.scale((p2.x-p0.x)/(p1.x-p0.x), (p2.y-p0.y)/(p1.y-p0.y))
                what.drawData().drawView()
                meta.ps.length < 3 && canvas.restore()
            }
        },
        rotate: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save(), what.clear().drawLine(meta)
            canvas.rotate(Math.atan2(p1.y-p0.y, p1.x-p0.x))
            what.drawData().drawView()
            meta.ps.length < 2 && canvas.restore()
        },

        draw: function(meta) {
            function trans(value) {
                if (value == "random") {
                    return ["black", "red", "green", "yellow", "blue", "purple", "cyan", "white"][parseInt(Math.random()*8)]
                }
                return value
            }
            canvas.strokeStyle = trans(meta.style || conf.style)
            canvas.fillStyle = trans(meta.style || conf.style)
            canvas[meta.tool||conf.tool]()
            return meta
        },
        drawText: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            var t = meta.text||status.cmd.value||conf.text

            canvas.save()
            canvas.translate(p0.x, p0.y)
            canvas.scale(1, -1)
            canvas.rotate(-Math.atan2(p1.y-p0.y, p1.x-p0.x))
            what.draw(meta)
            canvas.font=kit.distance(p0.x, p0.y, p1.x, p1.y)/t.length*2+"px "+conf.font
            canvas[(meta.tool||conf.tool)+"Text"](t, 0, 0)
            canvas.restore()
            return meta
        },
        drawPoint: function(meta) {
            meta.ps.concat(now).forEach(function(p) {
                canvas.save()
                canvas.translate(p.x, p.y)
                canvas.beginPath()
                canvas.moveTo(-conf.mpoint, 0)
                canvas.lineTo(conf.mpoint, 0)
                canvas.moveTo(0, -conf.mpoint)
                canvas.lineTo(0, conf.mpoint)
                what.draw(meta)
                canvas.restore()
            })
            return meta
        },
        drawLine: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save()
            canvas.beginPath()
            canvas.moveTo(p0.x, p0.y)
            canvas.lineTo(p1.x, p1.y)
            what.draw(meta)
            canvas.restore()
            return meta
        },
        drawPoly: function(meta) {
            var ps = meta.ps
            canvas.save()
            canvas.beginPath()
            canvas.moveTo(ps[0].x, ps[0].y)
            for (var i = 1; i < ps.length; i++) {
                canvas.lineTo(ps[i].x, ps[i].y)
            }
            ps.length < conf.limits.drawPoly && canvas.lineTo(now.x, now.y)
            what.draw(meta)
            canvas.restore()
            return meta
        },
        drawRect: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save()
            what.draw(meta)
            canvas[(meta.tool||conf.tool)+"Rect"](p0.x, p0.y, p1.x-p0.x, p1.y-p0.y)
            canvas.restore()
            return meta
        },
        drawCircle: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            canvas.save()
            canvas.beginPath()
            canvas.arc(p0.x, p0.y, kit.distance(p0.x, p0.y, p1.x, p1.y), 0, Math.PI*2, true)
            what.draw(meta)
            canvas.restore()
            return meta
        },
        drawEllipse: function(meta) {
            var p0 = meta.ps[0] || {x:0, y:0}
            var p1 = meta.ps[1] || now
            var r0 = Math.abs(p1.x-p0.x)
            var r1 = Math.abs(p1.y-p0.y)

            canvas.save()
            canvas.beginPath()
            canvas.translate(p0.x, p0.y)
            r1 > r0? (canvas.scale(r0/r1, 1), r0 = r1): canvas.scale(1, r1/r0)
            canvas.arc(0, 0, r0, 0, Math.PI*2, true)
            what.draw(meta)
            canvas.restore()
            return meta
        },

        drawAxies: function() {
            canvas.beginPath()
            canvas.moveTo(-space, 0)
            canvas.lineTo(width+space, 0)
            canvas.moveTo(0, -space)
            canvas.lineTo(0, height+space)
            canvas.strokeStyle = conf.axies.style
            canvas.lineWidth = conf.axies.width
            canvas.stroke()
            return what
        },
        drawXLabel: function(step) {
            canvas.beginPath()
            for (var pos = step; pos < width; pos += step) {
                canvas.moveTo(pos, 0)
                canvas.lineTo(pos, -conf.xlabel.height)
            }
            canvas.strokeStyle = conf.xlabel.style
            canvas.lineWidth = conf.xlabel.width
            canvas.stroke()
            return what
        },

        figure: {
            trend: {
                draw: function() {
                    what.drawAxies().drawXLabel(step)
                    canvas.beginPath()
                    for (var key in data) {
                        data[key].forEach(function(value, i) {
                            i == 0? canvas.moveTo(0, value/max[key]*height): canvas.lineTo(step*i, value/max[key]*height)
                            i == index && (canvas.moveTo(step*i, 0), canvas.lineTo(step*i, value/max[key]*height))
                        })
                    }
                    canvas.strokeStyle = conf.data.style
                    canvas.lineWidth = conf.data.width
                    canvas.stroke()
                },
                show: function(p) {
                    index = parseInt(p.x/step)
                    canvas.moveTo(p.x, -space)
                    canvas.lineTo(p.x, height)
                    canvas.moveTo(-space, p.y)
                    canvas.lineTo(width, p.y)
                    return p
                },
            },
            ticket: {
                draw: function() {
                    what.drawAxies().drawXLabel(step)
                    if (keys.length < 3) {
                        return
                    }

                    var sum = 0, total = 0
                    for (var i = 0; i < nrow; i++) {
                        sum += data[keys[1]][i]
                        sum > total && (total = sum)
                        sum -= data[keys[2]||keys[1]][i]
                    }
                    if (!data["sum"]) {
                        var sum = 0, max = 0, min = 0, end = 0
                        keys = keys.concat(["sum", "max", "min", "end"])
                        data["sum"] = []
                        data["max"] = []
                        data["min"] = []
                        data["end"] = []
                        for (var i = 0; i < nrow; i++) {
                            max = sum + data[keys[1]][i]
                            min = sum - data[keys[2||keys[1]]][i]
                            end = sum + data[keys[1]][i] - data[keys[2]||keys[1]][i]
                            data["sum"].push(sum)
                            data["max"].push(max)
                            data["min"].push(min)
                            data["end"].push(end)
                            sum = end
                        }
                        msg.append.push("sum")
                        msg.sum = data.sum
                        msg.append.push("max")
                        msg.max = data.max
                        msg.append.push("min")
                        msg.min = data.min
                        msg.append.push("end")
                        msg.end = data.end
                    }

                    for (var i = 0; i < nrow; i++) {
                        canvas.beginPath()
                        canvas.moveTo(step*i, data["min"][i]/total*height)
                        if (data["sum"][i] < data["end"][i]) {
                            canvas.strokeStyle = "white", canvas.lineTo(step*i, data["sum"][i]/total*height), canvas.stroke()
                            canvas.fillStyle = "white", canvas.fillRect(step*i-step/3, data["sum"][i]/total*height, step/3*2, (data["end"][i]-data["sum"][i])/total*height)
                            canvas.moveTo(step*i, data["end"][i]/total*height)
                        } else {
                            canvas.strokeStyle = "black", canvas.lineTo(step*i, data["end"][i]/total*height), canvas.stroke()
                            canvas.fillStyle = "black", canvas.fillRect(step*i-step/3, data["sum"][i]/total*height, step/3*2, (data["end"][i]-data["sum"][i])/total*height)
                            canvas.moveTo(step*i, data["sum"][i]/total*height)
                        }
                        canvas.lineTo(step*i, data["max"][i]/total*height), canvas.stroke()
                    }
                    // canvas.strokeStyle = conf.data.style
                    // canvas.lineWidth = conf.data.width
                    // canvas.stroke()
                },
                show: function(p) {
                    index = parseInt(p.x/step)
                    canvas.moveTo(p.x, -space)
                    canvas.lineTo(p.x, height)
                    canvas.moveTo(-space, p.y)
                    canvas.lineTo(width, p.y)
                    return p
                },
            },
            stick: {
                draw: function() {
                    what.drawAxies().drawXLabel(step)
                    canvas.beginPath()

                    var total = 0
                    for (var key in max) {
                        total += max[key]
                    }

                    for (var i = 0; i < nrow; i++) {
                        canvas.moveTo(step*i, 0)
                        for (var key in data) {
                            canvas.lineTo(step*i, data[key][i]/total*height)
                            canvas.moveTo(step*i-step/2, data[key][i]/total*height)
                            canvas.lineTo(step*i+step/2, data[key][i]/total*height)
                            canvas.moveTo(step*i, data[key][i]/total*height)
                        }
                    }
                    canvas.strokeStyle = conf.data.style
                    canvas.lineWidth = conf.data.width
                    canvas.stroke()
                },
                show: function(p) {
                    index = parseInt(p.x/step)
                    canvas.moveTo(p.x, -space)
                    canvas.lineTo(p.x, height)
                    canvas.moveTo(-space, p.y)
                    canvas.lineTo(width, p.y)
                    return p
                },
            },
            weight: {
                conf: {
                    space: 20,
                    focus: "white",
                    style: "black",
                    width: 1,
                    least: 0.01,
                },
                draw: function() {
                    var that = this
                    var space = width / (nline+1)

                    canvas.translate(0, height/2)
                    for (var key in data) {
                        var total = 0
                        data[key].forEach(function(value) {
                            total += value
                        })

                        var sum = 0
                        canvas.translate(space, 0)
                        data[key].forEach(function(value, i) {
                            if (value/total < that.conf.least) {
                                return
                            }

                            var a = sum/total*Math.PI*2
                            var b = (sum+value)/total*Math.PI*2
                            sum+=value

                            canvas.beginPath()
                            canvas.moveTo(0, 0)
                            canvas.arc(0, 0, (space/2)-that.conf.space, a, b, false)
                            canvas.closePath()

                            if (i == index) {
                                canvas.fillStyle = that.conf.focus
                                canvas.fill()
                            } else {
                                canvas.strokeStyle = that.conf.style
                                canvas.lineWidth = that.conf.width
                                canvas.stroke()
                            }
                        })
                    }
                },
                show: function(p) {
                    var nspace = width / (nline+1)
                    var which = parseInt((p.x-nspace/2)/nspace)
                    which >= nline && (which = nline-1), which < 0 && (which = 0)

                    var q = what.reverse(p)
                    canvas.translate((which+1)*nspace, height/2)
                    var p = what.transform(q)

                    var a = Math.atan2(p.y, p.x)
                    a < 0 && (a += Math.PI*2)
                    var pos = a/2/Math.PI

                    var total = 0
                    data[keys[which]].forEach(function(value) {
                        total += value
                    })
                    var sum = 0, weight = 0
                    data[keys[which]].forEach(function(value, i) {
                        sum += value, sum / total < pos && (index = i+1)
                        index == i && (weight = parseInt(value/total*100))
                    })

                    canvas.moveTo(0, 0)
                    canvas.lineTo(p.x, p.y)
                    canvas.lineTo(p.x+conf.plabel.length, p.y)

                    canvas.scale(1, -1)
                    canvas.fillText("weight: "+weight+"%", p.x+conf.plabel.offset, -p.y+conf.plabel.offset)
                    canvas.scale(1, -1)
                    return p
                },
            },
        },

        drawData: function() {
            canvas.save()
            what.figure[conf.type].draw()
            canvas.restore()
            return what
        },
        drawView: function() {
            view.forEach(function(view) {
                view.meta && what[view.type](view.meta)
            })
            return what
        },
        drawLabel: function() {
            if (!label) { return what }

            index = 0
            canvas.save()
            canvas.font = conf.plabel.font || conf.font
            canvas.fillStyle = conf.plabel.style || conf.style
            canvas.strokeStyle = conf.plabel.style || conf.style
            var p = what.figure[conf.type].show(now)
            canvas.stroke()

            canvas.scale(1, -1)
            p.x += conf.plabel.offset
            p.y -= conf.plabel.offset

            if (width - p.x < 200) {
                p.x -= 200
            }
            canvas.fillText("index: "+index, p.x, -p.y+conf.plabel.height)
            msg.append.forEach(function(key, i) {
                msg[key][index] && canvas.fillText(key+": "+msg[key][index], p.x, -p.y+(i+2)*conf.plabel.height)
            })
            canvas.restore()
            return what
        },
        drawShape: function() {
            point.length > 0 && (what[conf.shape]({ps: point}), what[conf.means]({ps: point, tool: "stroke", style: "red"}))
            return what
        },

        refresh: function() {
            return what.clear().drawData().drawView().drawLabel().drawShape()
        },
        cancel: function() {
            point = [], what.refresh()
            return what
        },
        play: function() {
            function cb() {
                view[i] && what[view[i].type](view[i].meta) && (t = kit.Delay(view[i].type == "drawPoint"? 10: conf.play, cb))
                i++
                status.nshape.innerText = i+"/"+view.length
            }
            var i = 0
            what.clear().drawData()
            kit.Delay(10, cb)
            return what
        },
        back: function() {
            view.pop(), status.nshape.innerText = view.length
            return what.refresh()
        },
        push: function(item) {
            item.meta && item.meta.ps < (conf.limits[item.type]||2) && ps.push(item)
            status.nshape.innerText = view.push(item)
            return what
        },
        wait: function() {
            status.cmd.focus()
            return what
        },
        trap: function(value, event) {
            event.target.className = (trap = !trap)? "trap": "normal"
            page.localMap = trap? what.input: undefined
        },
        label: function(value, event) {
            event.target.className = (label = !label)? "trap": "normal"
        },

        movePoint: function(p) {
            now = p, status.xy.innerHTML = p.x+","+p.y;
            (point.length > 0 || ps.length > 0 || label) && what.refresh()
        },
        pushPoint: function(p) {
            if (ps.length > 0) {
                ps[0].meta.ps.push(p) > 1 && ps.pop(), what.refresh()
                return
            }

            point.push(p) >= (conf.limits[conf.shape]||2) && what.push({type: conf.shape,
                meta: what[conf.shape]({ps: point, text: status.cmd.value||conf.text, tool: conf.tool, style: conf.style}),
            }) && (point = [])
            conf.means == "drawPoint" && what.push({type: conf.means, meta: what[conf.means]({ps: [p], tool: "stroke", style: "red"})})
        },
        transform: function(p) {
            var t = canvas.getTransform()
            return {
                x: (p.x-t.c/t.d*p.y+t.c*t.f/t.d-t.e)/(t.a-t.c*t.b/t.d),
                y: (p.y-t.b/t.a*p.x+t.b*t.e/t.a-t.f)/(t.d-t.b*t.c/t.a),
            }
        },
        reverse: function(p) {
            var t = canvas.getTransform()
            return {
                x: t.a*p.x+t.c*p.y+t.e,
                y: t.b*p.x+t.d*p.y+t.f,
            }
        },

        check: function() {
            view.forEach(function(item, index, view) {
                item && item.send && plugin.Run(window.event||{}, item.send.concat(["type", item.type]), function(msg) {
                    msg.text && msg.text[0] && (item.meta.text = msg.text[0])
                    msg.style && msg.style[0] && (item.meta.style = msg.style[0])
                    msg.ps && msg.ps[0] && (item.meta.ps = JSON.parse(msg.ps[0]))
                    what.refresh()
                })
                index == view.length -1 && kit.Delay(1000, what.check)
            })
        },
        parse: function(txt) {
            var meta = {}, cmds = [], rest = -1, send = []
            txt.trim().split(" ").forEach(function(item) {
                switch (item) {
                    case "stroke":
                    case "fill":
                        meta.tool = item
                        break
                    case "black":
                    case "white":
                    case "red":
                    case "yellow":
                    case "green":
                    case "cyan":
                    case "blue":
                    case "purple":
                        meta.style = item
                        break
                    case "cmds":
                        rest = cmds.length
                    default:
                        cmds.push(item)
                }
            }), rest != -1 && (send = cmds.slice(rest+1), cmds = cmds.slice(0, rest))

            var cmd = {
                "t": "drawText",
                "l": "drawLine",
                "p": "drawPoly",
                "r": "drawRect",
                "c": "drawCircle",
                "e": "drawEllipse",
            }[cmds[0]] || cmds[0]
            cmds = cmds.slice(1)

            var args = []
            switch (cmd) {
                case "send":
                    plugin.Run(window.event, cmds, function(msg) {
                        kit.Log(msg)
                    })
                    return
                default:
                    meta.ps = []
                    for (var i = 0; i < cmds.length; i+=2) {
                        var x = parseInt(cmds[i])
                        var y = parseInt(cmds[i+1])
                        !isNaN(x) && !isNaN(y) && meta.ps.push({x: x, y: y}) || (args.push(cmds[i]), i--)
                    }
            }
            meta.args = args

            switch (cmd) {
                case "drawText":
                    meta.text = args.join(" "), delete(meta.args)
                case "drawLine":
                case "drawPoly":
                case "drawRect":
                case "drawCircle":
                case "drawEllipse":
                    what.push({type: cmd, meta: what[cmd](meta), send:send})
            }

            return (what[cmd] || function() {
                return what
            })(meta)
        },
        input: function(event) {
            var map = what.trans[event.key]
            map && action[map[0]] && (action[map[0]].value = map[1])
            map && what.trans[map[0]] && (map = what.trans[map[1]]) && (conf[map[0]] && (conf[map[0]] = map[1]) || what[map[0]] && what[map[0]]())
            what.refresh()
        },
        trans: {
            "折线图": ["type", "trend"],
            "股价图": ["type", "ticket"],
            "柱状图": ["type", "stick"],
            "饼状图": ["type", "weight"],

            "移动": ["shape", "move"],
            "旋转": ["shape", "rotate"],
            "缩放": ["shape", "scale"],

            "文本": ["shape", "drawText"],
            "直线": ["shape", "drawLine"],
            "折线": ["shape", "drawPoly"],
            "矩形": ["shape", "drawRect"],
            "圆形": ["shape", "drawCircle"],
            "椭圆": ["shape", "drawEllipse"],

            "辅助点": ["means", "drawPoint"],
            "辅助线": ["means", "drawRect"],

            "画笔": ["tool", "stroke"],
            "画刷": ["tool", "fill"],

            "黑色": ["style", "black"],
            "红色": ["style", "red"],
            "绿色": ["style", "green"],
            "黄色": ["style", "yellow"],
            "蓝色": ["style", "blue"],
            "紫色": ["style", "purple"],
            "青色": ["style", "cyan"],
            "白色": ["style", "white"],
            "随机色": ["style", "random"],
            "默认色": ["style", "default"],

            "清屏": ["clear"],
            "刷新": ["refresh"],
            "取消": ["cancel"],
            "播放": ["play"],
            "回退": ["back"],
            "输入": ["wait"],

            "标签": ["label"],
            "快捷键": ["trap"],

            "x": ["折线图", "折线图"],
            "y": ["折线图", "饼状图"],

            "a": ["移动", "旋转"],
            "m": ["移动", "移动"],
            "z": ["移动", "缩放"],

            "t": ["文本", "文本"],
            "l": ["文本", "直线"],
            "v": ["文本", "折线"],
            "r": ["文本", "矩形"],
            "c": ["文本", "圆形"],
            "e": ["文本", "椭圆"],

            "s": ["画笔", "画笔"],
            "f": ["画笔", "画刷"],

            "0": ["黑色", "黑色"],
            "1": ["黑色", "红色"],
            "2": ["黑色", "绿色"],
            "3": ["黑色", "黄色"],
            "4": ["黑色", "蓝色"],
            "5": ["黑色", "紫色"],
            "6": ["黑色", "青色"],
            "7": ["黑色", "白色"],
            "8": ["黑色", "随机色"],
            "9": ["黑色", "默认色"],

            "j": ["刷新", "刷新"],
            "g": ["播放", "播放"],
            "b": ["回退", "回退"],
            "q": ["清空", "清空"],

            "Escape": ["取消", "取消"],
            " ": ["输入", "输入"],
        },
    }

    var action = kit.AppendAction(kit.AppendChild(output, [{view: ["action"]}]).last, [
        ["折线图", "股价图", "柱状图", "饼状图"],
        ["移动", "旋转", "缩放"],
        ["文本", "直线", "折线", "矩形", "圆形", "椭圆"],
        ["辅助点", "辅助线"],
        ["画笔", "画刷"],
        ["黑色", "红色", "绿色", "黄色", "蓝色", "紫色", "青色", "白色", "随机色", "默认色"],
        "", "清屏", "刷新", "播放", "回退",
        "", "标签", "快捷键",
    ], function(value, event) {
        var map = what.trans[value]
        conf[map[0]] && (conf[map[0]] = map[1]) || what[map[0]] && what[map[0]](value, event)
        what.refresh()
    })

    var canvas = kit.AppendChild(output, [{view: ["draw", "canvas"], data: {width: width+20, height: height+20,
        onclick: function(event) {
            what.pushPoint(what.transform({x: event.offsetX, y: event.offsetY}), event.clientX, event.clientY)
        }, onmousemove: function(event) {
            what.movePoint(what.transform({x: event.offsetX, y: event.offsetY}), event.clientX, event.clientY)
        },
    }}]).last.getContext("2d")

    var status = kit.AppendStatus(kit.AppendChild(output, [{view: ["status"]}]).last, [{name: "nshape"}, {"className": "cmd", style: {width: (output.clientWidth - 100)+"px"}, data: {autocomplete: "off"}, input: ["cmd", function(event) {
        var target = event.target
        event.type == "keyup" && event.key == "Enter" && what.parse(target.value) && (!target.History && (target.History=[]),
            target.History.push(target.value), target.Current=target.History.length, target.value = "")
        event.type == "keyup" && page.oninput(event), event.stopPropagation()

    }]}, {name: "xy"}], function(value, name, event) {

    })

    return what.reset().refresh()
}


