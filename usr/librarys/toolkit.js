shy = function(help, meta, list, cb) {
    var index = -1, value = "", type = "string", args = arguments
    function next(check) {
        if (++index >= args.length) {return false}
        if (check && check != typeof args[index]) {index--; return false}
        return value = args[index], type = typeof value, value
    }

    var cb = arguments[arguments.length-1] || function() {}
    cb.help = next("string") || "还没有写"
    cb.meta = next("object") || {}
    cb.list = next("object") || {}
    cb.runs = function() {}
    return cb
}
kit = toolkit = (function() {var kit = {__proto__: document,
    // 用户终端
    device: {
        isWeiXin: navigator.userAgent.indexOf("MicroMessenger") > -1,
        isMobile: navigator.userAgent.indexOf("Mobile") > -1,
        isIPhone: navigator.userAgent.indexOf("iPhone") > -1,
        isMacOSX: navigator.userAgent.indexOf("Mac OS X") > -1,
        isWindows: navigator.userAgent.indexOf("Windows") > -1,
    },
    alert: function(text) {alert(JSON.stringify(text))},
    confirm: function(text) {return confirm(text)},
    prompt: function(text, cb) {
        var text = prompt(text)
        text && kit._call(cb, text)
        return text
    },
    reload: function() {location.reload()},
    // 日志调试
    History: shy("历史记录", {lay: [], cmd: [], txt: [], key: []}, function(type, index, data) {var meta = arguments.callee.meta
        if (kit.isNone(index)) {return meta[type]}
        var list = meta[type] || []
        if (kit.isNone(data)) {var len = list.length
            return list[(index+len)%len]
        }
        return meta[type] = list, list.push({time: Date.now(), data: data})-1
    }),
    Debug: shy("调试断点", {why: true, msg: true, config: false}, function(key) {
        if (arguments.callee.meta[key]) {debugger}
    }),
    Log: shy("输出日志", {hide: {"init": true, "wss": false}, call: [],
        func: {debug: "debug", info: "info", parn: "warn", err: "error"},
    }, function(type, arg) {var meta = arguments.callee.meta
        var args = [kit.time().split(" ")[1]].concat(kit.List(kit.isNone(arg)? type: arguments))
        !meta.hide[args[1]] && console[meta.func[args[1]]||"log"](args)
        kit.List(meta.call, function(cb) {kit._call(cb, args)})
        kit.Debug(args[1])
        return args.slice(1)
    }),
    Tip: shy("用户提示", function() {}),

    // HTML节点操作
    classList: {
        has: function(obj, key) {
            var list = obj.className? obj.className.split(" "): []
            for (var i = 1; i < arguments.length; i++) {
                if (list.indexOf(arguments[i]) == -1) {
                    return false
                }
            }
            return true
        },
        add: function(obj, key) {var list = obj.className? obj.className.split(" "): []
            return obj.className = list.concat(kit.List(arguments, function(value, index) {
                return index > 0 && list.indexOf(value) == -1? value: undefined
            })).join(" ")
        },
        del: function(obj, key) {
            var list = kit.List(arguments, function(value, index) {return index > 0? value: undefined})
            return obj.className = kit.List(obj.className.split(" "), function(value) {
                return list.indexOf(value) == -1? value: undefined
            }).join(" ")
        },
    },
    ModifyView: function(which, args) {
        var width = document.body.clientWidth-4
        var height = document.body.clientHeight-4
        kit.Item(args, function(key, value) {var w = h = value
            if (typeof(value) == "object") {w = value[0], h = value[1]}
            switch (key) {
                case "dialog": // 设置宽高
                    if (w > width) {w = width}
                    if (h > height) {h = height}
                    args["top"] = (height-h)/2
                    args["left"] = (width-w)/2
                    args["width"] = w
                    args["height"] = h
                    break
                case "window": // 设置边距
                    args["top"] = h/2
                    args["left"] = w/2
                    args["width"] = width-w-20
                    args["height"] = height-h-20
                    break
                default:
                    return
            }
            delete(args[key])
        })

        var list = ["top", "left", "width", "height", "padding", "margin"]
        kit.Item(args, function(key, value) {
            typeof value == "number" && list.indexOf(key) != -1 && (args[key] = value+"px")
        })
        return kit.ModifyNode(which, {style: args})
    },
    ModifyNode: function(which, html) {
        var node = typeof which == "string"? document.querySelector(which): which
        typeof html == "string"? (node.innerHTML = html): kit.Item(html, function(key, value) {
            typeof value != "object"? (node[key] = value): kit.Item(value, function(item, value) {
                node[key] && (node[key][item] = value)
            })
        })
        return node
    },
    CreateNode: function(element, html) {return kit.ModifyNode(document.createElement(element), html)},
    AppendChild: function(parent, children, subs) {
        if (typeof children == "string") {
            var elm = kit.CreateNode(children, subs)
            return parent.append(elm), elm
        }

        // 基本属性: name value title
        // 基本内容: inner innerHTML
        // 基本样式: style className
        // 基本事件: click dataset
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

        subs = subs || {}
        children.forEach(function(child, i) {if (kit.isNone(child)) {return}
            var type = child.type || "div", data = child.data || {}
            var name = child.name || data.name

            kit.List([
                "name", "value", "title",
                "innerHTML",
                "className",
                "dataset",
            ], function(key) {
                kit.notNone(child[key]) && (data[key] = child[key])
            })
            kit.notNone(child.click) && (data.onclick = child.click)
            kit.notNone(child.inner) && (data.innerHTML = child.inner)
            kit.notNone(child.style) && (data.style = typeof child.style == "string"? child.style: kit.Item(child.style, function(key, value) {
                return [key, ": ", kit.pixs(key, value)].join("")
            }).join("; "))

            if (kit.notNone(child.button)) {var list = kit.List(child.button)
                type = "button", name = name || list[0]
                data.innerText = list[0], data.onclick = function(event) {
                    kit._call(list[1], [event, list[0]])
                }

            } else if (child.select) {var list = child.select
                type = "select", name = name || list[0][0]
                data.onchange = function(event) {
                    kit._call(list[1], [event, event.target.value])
                }
                child.list = list[0].slice(1).map(function(value) {
                    return {type: "option", value: value, inner: value}
                })
                data.className = list[0][0] || ""

            } else if (child.input) {var list = kit.List(child.input)
                type = "input", name = name || list[0]
                data.onkeydown = function(event) {
                    kit._call(list[1], [event])
                }
                data.onkeyup = function(event) {
                    kit._call(list[2], [event])
                }

            } else if (child.password) {var list = kit.List(child.password)
                type = "input", name = name || list[0]
                data.type = "password"

            } else if (child.label) {var list = kit.List(child.label)
                type = "label", data.innerText = list[0]

            } else if (child.img) {var list = kit.List(child.img)
                type = "img", data.src = list[0], data.onload = function(event) {
                    kit._call(list[1], [event])
                }

            } else if (child.row) {
                type = "tr"
                child.list = child.row.map(function(item) {return {text: [item, child.sub||"td"]}})

            } else if (child.tree) {
                type = "ul", child.list = child.tree

            } else if (child.fork) {var list = kit.List(child.fork)
                type = "li", child.list = [
                    {"text": [list[0], "div"], "click": function(event) {
                        kit._call(list[2], [event])
                    }},
                    {"type": "ul", "list": list[1]},
                ]

            } else if (child.leaf) {var list = kit.List(child.leaf)
                type = "li"
                child.list = [{"text": [list[0], "div"]}]
                data.onclick = function(event) {
                    kit._call(list[1], [event])
                }

            } else if (child.view) {var list = kit.List(child.view);
                (list.length > 0 && list[0]) && (data.className = list[0])
                type = list[1] || "div"
                data.innerHTML = list[2] || ""
                name = name || list[3] || ""

            } else if (child.text) {var list = kit.List(child.text)
                data.innerHTML = list[0]
                type = list[1] || "pre"
                list.length > 2 && (data.className = list[2])

            } else if (child.code) {var list = kit.List(child.code)
                type = "code"
                child.list = [{type: "pre" ,data: {innerText: list[0]}, name: list[1]||""}]
                list.length > 2 && (data.className = list[2])

            } else if (child.script) {
                type = "script", data.innerHTML = child.script

            } else if (child.include) {var list = kit.List(child.include)
                type = "script", data.type = "text/javascript"
                data.src = list[0], data.onload = function(event) {
                    kit._call(list[1], [event])
                }

            } else if (child.require) {var list = kit.List(child.require)
                type = "link", data.type = "text/css", data.rel = "stylesheet"
                data.href = list[0], data.onload = function(event) {
                    kit._call(list[1], [event])
                }

            } else if (child.styles) {
                type = "style", data.type = "text/css"
                data.innerHTML = typeof child.styles == "string"? child.styles: kit.Item(child.styles, function(key, value) {
                    return key + " {\n" + kit.Item(value, function(item, value) {
                        return ["  ", item, ": ", kit.pixs(value)].join("")
                    }).join(";\n") + "\n}\n"
                }).join("")
            }

            name = name || data.className || type
            var node = kit.CreateNode(type, data)
            child.list && kit.AppendChild(node, child.list, subs)
            subs.first || (subs.first = node), subs.last = node
            name && (subs[name] = node)
            parent && parent.append && parent.append(node)
        })
        return subs
    },
    AppendChilds: function(parent, children, subs) {
        return parent.innerHTML = "", kit.AppendChild(parent, children, subs)
    },
    InsertChild: function (parent, position, element, children) {
        var elm = kit.CreateNode(element)
        kit.AppendChild(elm, children)
        return parent.insertBefore(elm, position || parent.firstElementChild)
    },
    // HTML控件操作
    AppendActions: function(parent, list, cb, diy) {
        parent.innerHTML = "", kit.AppendAction(parent, list, cb, diy)
    },
    AppendAction: function(parent, list, cb, diy) {
        if (diy) {
            return kit.AppendChild(parent, kit.List(list, function(item, index) {
                return item === ""? {view: ["space"]}:
                    typeof item == "string"? {view: ["item", "div", item], click: function(event) {
                        kit._call(cb, [event, item])
                    }}: item.forEach? {view: item[0], list: kit.List(item.slice(1), function(value) {return {text: [value, "div", "item"], click: function(event) {
                        kit._call(cb, [event, value])
                    }}})}: item
            }))
        }
        return kit.AppendChild(parent, kit.List(list, function(item, index) {
            return item === ""? {view: ["space"]}:
                typeof item == "string"? {button: [item, cb]}:
                    item.forEach? {select: [item, cb]}: item
        }))
    },
    AppendTable: function(table, data, fields, cb, cbs) {if (!data || !fields) {return}
        kit.AppendChild(table, [{row: fields, sub: "th"}])
        data.forEach(function(row, i) {
            var tr = kit.AppendChild(table, "tr", {className: "normal"})
            tr.Meta = row, fields.forEach(function(key, j) {
                var td = kit.AppendChild(tr, "td", kit.Color(row[key]))

				if (key == "when") {td.className = "when"}
                if ((row[key]||"").startsWith("http")) {
                    td.innerHTML = "<a href='"+row[key]+"' target='_blank'>"+row[key]+"</a>"
                }

                cb && (td.onclick = function(event) {
                    kit._call(cb, [row[key], key, row, i, tr, event])
                })
                cbs && (td.oncontextmenu = function(event) {
                    kit._call(cbs, [row[key], key, row, i, tr, event])
                    event.stopPropagation()
                    event.preventDefault()
                })
            })
        })
        return table
    },
    RangeTable: function(table, index, sort_asc) {
        var list = kit.Selector(table, "tr", function(tr) {
            return tr.style.display == "none" || kit.classList.has(tr, "hide")? null: tr
        }).slice(1)

        var is_time = true, is_number = true
        kit.List(list, function(tr) {
            var text = tr.childNodes[index].innerText
            is_time = is_time && Date.parse(text) > 0
            is_number = is_number && !isNaN(parseInt(text))
        })

        var num_list = kit.List(list, function(tr) {
            var text = tr.childNodes[index].innerText
            return is_time? Date.parse(text):
                is_number? parseInt(text): text
        })

        for (var i = 0; i < num_list.length; i++) {
            for (var j = i+1; j < num_list.length; j++) {
                if (sort_asc? num_list[i] < num_list[j]: num_list[i] > num_list[j]) {
                    var temp = num_list[i]
                    num_list[i] = num_list[j]
                    num_list[j] = temp
                    var temp = list[i]
                    list[i] = list[j]
                    list[j] = temp
                }
            }
            var tbody = list[i].parentElement
            list[i].parentElement && tbody.removeChild(list[i])
            tbody.appendChild(list[i])
        }
    },
    OrderTable: function(table, field, cb, cbs) {if (!table) {return}
        table.oncontextmenu = table.onclick = function(event) {var target = event.target
            target.parentElement.childNodes.forEach(function(item, i) {if (item != target) {return}
                if (target.tagName == "TH") {var dataset = target.dataset
                    dataset["sort_asc"] = (dataset["sort_asc"] == "1") ? 0: 1
                    kit.RangeTable(table, i, dataset["sort_asc"] == "1")
                } else if (target.tagName == "TD") {var index = 0
                    kit.Selector(table, "tr", function(item, i) {item == target.parentElement && (index = i)})
                    var name = target.parentElement.parentElement.querySelector("tr").childNodes[i].innerText
                    name.startsWith(field) && kit._call(event.type=="contextmenu"? cbs: cb, [event, item.innerText, name, item.parentNode.Meta, index])
                } else if (target.parentNode.tagName == "TD"){
                    kit.Selector(table, "tr", function(item, i) {item == target.parentElement.parentElement && (index = i)})
                    var name = target.parentElement.parentElement.parentElement.querySelector("tr").childNodes[i].innerText
                    name.startsWith(field) && kit._call(event.type=="contextmenu"? cbs: cb, [event, item.innerText, name, item.parentNode.Meta, index])
                }
            })
            kit.CopyText()
        }
        return true
    },
    Change: function(target, cb) {
        var value = target.value
        function reset(event) {
            value != event.target.value && kit._call(cb, [event.target.value, value])
            target.innerHTML = event.target.value
        }
        kit.AppendChilds(target, [{type: "input", value: target.innerText, data: {
            onblur: reset,
            onkeydown: function(event) {
                page.oninput(event), event.key == "Enter" && reset(event)
            },
        }}]).last.focus()
    },

    // HTML显示文本
    OrderCode: function(code) {if (!code) {return}
        code.onclick = function(event) {kit.CopyText()}
        kit.Selector(code, "a", function(item) {
            item.target = "_blank"
        })
    },
    OrderLink: function(link) {link.target = "_blank"},
    OrderText: function(pane, text) {
        text.querySelectorAll("a").forEach(function(value, index, array) {
            kit.OrderLink(value, pane)
        })
        text.querySelectorAll("code").forEach(function(value, index, array) {
            kit.OrderCode(value)
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
    },
    Position: function(which) {
        return (parseInt((which.scrollTop + which.clientHeight) / which.scrollHeight * 100)||0)+"%"
    },
    // HTML输入文本
    CopyText: function(text) {
        if (text) {
            var input = kit.AppendChild(document.body, [{type: "textarea", inner: text}]).last
            input.focus(), input.setSelectionRange(0, text.length)
        }

        text = window.getSelection().toString()
        if (text == "") {return ""}

        kit.History("txt", -1) && kit.History("txt", -1).data == text || kit.History("txt", -1, text) && document.execCommand("copy")
        input && document.body.removeChild(input)
        return text
    },
    DelText: function(target, start, count) {
        target.value = target.value.substring(0, start)+target.value.substring(start+(count||target.value.length), target.value.length)
        target.setSelectionRange(start, start)
    },
    HitText: function(target, text) {
        var start = target.selectionStart
        for (var i = 1; i < text.length+1; i++) {
            var ch = text[text.length-i]
            if (target.value[start-i] != ch && kit.History("key", -i).data != ch) {
                return false
            }
        }
        return true
    },
    Delay: function(time, cb) {
        return setTimeout(cb, time)
    },

    // 数据容器迭代
    Push: function(list, value) {list = list || []
        return (kit.notNone||check)(value) && list.push(value), list
    },
    List: function(obj, cb, interval, cbs) {obj = typeof obj == "string"? [obj]: (obj || [])
        if (interval > 0) {
            function loop(i) {if (i >= obj.length) {return kit._call(cbs)}
                kit._call(cb, [obj[i], i, obj]), setTimeout(function() {loop(i+1)}, interval)
            }
            obj.length > 0 && setTimeout(function() {loop(0)}, interval/4)
            return obj
        }

        var list = []
        for (var i = 0; i < obj.length; i++) {
            kit.Push(list, kit._call(cb, [obj[i], i, obj]))
        }
        return list
    },
    Item: function(obj, cb) {var list = []
        for (var k in obj) {
            kit.Push(list, kit._call(cb, [k, obj[k]]))
        }
        return list
    },
    Items: function(obj, cb) {var list = []
        for (var key in obj) {
            list = list.concat(kit.List(obj[key], function(value, index, array) {
                return kit._call(cb, [value, index, key, obj])
            }))
        }
        return list
    },
    Span: function(list) {list = list || []
		list.span = function(value, style) {
            return kit.List(arguments, function(item) {
                kit._call(list, list.push, typeof item == "string"? [item]: ['<span class="'+item[1]+'">', item[0], "</span>"])
            }), list.push("<br/>"), list
		}
        return list
    },
    Opacity: function(obj, interval, list) {
		kit.List(kit.Value(list, [0, 0.2, 0.4, 0.6, 1.0]), function(value) {
			obj.style.opacity = value
		}, kit.Value(interval, 150))
	},
    Selector: function(obj, item, cb, interval, cbs) {var list = []
		kit.List(obj.querySelectorAll(item), function(item, index, array) {
            kit.Push(list, kit._call(cb, [item, index, array]))
        }, interval, cbs)

        for (var i = list.length-1; i >= 0; i--) {
            if (list[i] !== "") {break}
            list.pop()
        }
        return list
    },
    // 数据类型转换
    isNone: function(c) {return c === undefined || c === null},
    notNone: function(c) {return !kit.isNone(c)},
    isSpace: function(c) {return c == " " || c == "Enter"},
    Format: function(objs) {return JSON.stringify(objs)},
    Origin: function(s) {
        s = s.replace(/</g, "&lt;")
        s = s.replace(/>/g, "&gt;")
        return s
    },
    Color: function(s) {if (!s) {return s}
        s = s.replace(/\033\[1m/g, "<span style='font-weight:bold'>")
        s = s.replace(/\033\[36m/g, "<span style='color:#0ff'>")
        s = s.replace(/\033\[33m/g, "<span style='color:#ff0'>")
        s = s.replace(/\033\[32m/g, "<span style='color:#0f0'>")
        s = s.replace(/\033\[32;1m/g, "<span style='color:#0f0'>")
        s = s.replace(/\033\[31m/g, "<span style='color:#f00'>")
        s = s.replace(/\033\[0m/g, "</span>")
        s = s.replace(/\033\[m/g, "</span>")
        s = s.replace(/\n/g, "<br/>")
        return s
    },
    Value: function() {
        for (var i = 0; i < arguments.length; i++) {
            switch (arguments[i]) {
                case undefined:
                case null:
                case "":
                    break
                default:
                    return arguments[i]
            }
        }
	},
    distance: function(x0, y0, x1, y1) {return Math.sqrt(Math.pow(x1-x0, 2)+Math.pow(y1-y0, 2))},
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
    size: function(obj, width, height) {obj.style.width = width+"px", obj.style.height = height+"px"},
    pixs: function(key, value) {
        var list = ["top", "left", "width", "height", "padding", "margin"]
        return typeof value == "number" && list.indexOf(key) != -1? value+"px": value
    },
    type: function(obj, type) {return type == undefined? typeof obj: typeof obj == type? obj: null},
    _call: function() {// obj, cb, arg
        var index = 0, obj, cb, arg;
        (obj = kit.type(arguments[index], "object")) && index++
        (cb = kit.type(arguments[index], "function")), index++
        (arg = kit.type(arguments[index], "object")) && index++

        arg = arg || []
        while (index < arguments.length) {
            arg.push(arguments[index++])
        }
        return typeof cb == "function"? cb.apply(obj||window, arg||[]): arg && arg.length > 0? arg[0]: null
    },
}; return kit})()

function Editor(run, plugin, option, output, width, height, space, msg) {
    exports = ["dir", "path", "dir"]
    msg.append && kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), msg.Table(), msg.append), exports[1], function(event, value, name, line) {
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
        type: "ticket", shape: "drawText", means: "drawPoint",
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
            "股价图": ["type", "ticket"],
            "折线图": ["type", "trend"],
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
        ["", "股价图", "折线图", "柱状图", "饼状图"],
        ["", "移动", "旋转", "缩放"],
        ["", "文本", "直线", "折线", "矩形", "圆形", "椭圆"],
        ["", "辅助点", "辅助线"],
        ["", "画笔", "画刷"],
        ["", "黑色", "红色", "绿色", "黄色", "蓝色", "紫色", "青色", "白色", "随机色", "默认色"],
        "", "清屏", "刷新", "播放", "回退",
        "", "标签", "快捷键",
    ], function(event, value) {
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

    var status = kit.AppendAction(kit.AppendChild(output, [{view: ["status"]}]).last, [{name: "nshape"}, {"className": "cmd", style: {width: (output.clientWidth - 100)+"px"}, data: {autocomplete: "off"}, input: ["cmd", function(event) {
        var target = event.target
        event.type == "keyup" && event.key == "Enter" && what.parse(target.value) && (!target.History && (target.History=[]),
            target.History.push(target.value), target.Current=target.History.length, target.value = "")
        event.type == "keyup" && page.oninput(event), event.stopPropagation()

    }]}, {name: "xy"}], function(value, name, event) {

    })

    return what.reset().refresh()
}

