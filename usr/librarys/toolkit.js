kit = toolkit = {
    isMobile: navigator.userAgent.indexOf("Mobile") > -1,
    isSpace: function(c) {
        return c == " " || c == "Enter"
    },
    History: {dir: [], pod: [], ctx: [], cmd: [], txt: [], key: [],
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

    ScrollPage: function(event, conf) {
        switch (event.key) {
            case "h":
                if (event.ctrlKey) {
                    window.scrollBy(-conf.scroll_x*10, 0)
                } else {
                    window.scrollBy(-conf.scroll_x, 0)
                }
                break
            case "H":
                window.scrollBy(-document.body.scrollWidth, 0)
                break
            case "l":
                if (event.ctrlKey) {
                    window.scrollBy(conf.scroll_x*10, 0)
                } else {
                    window.scrollBy(conf.scroll_x, 0)
                }
                break
            case "L":
                window.scrollBy(document.body.scrollWidth, 0)
                break
            case "j":
                if (event.ctrlKey) {
                    window.scrollBy(0, conf.scroll_y*10)
                } else {
                    window.scrollBy(0, conf.scroll_y)
                }
                break
            case "J":
                window.scrollBy(0, document.body.scrollHeight)
                break
            case "k":
                if (event.ctrlKey) {
                    window.scrollBy(0, -conf.scroll_y*10)
                } else {
                    window.scrollBy(0, -conf.scroll_y)
                }
                break
            case "K":
                window.scrollBy(0, -document.body.scrollHeight)
                break
        }
        return true
    },
    ModifyView: function(target, args) {
        var width = document.body.offsetWidth-10
        var height = document.body.offsetHeight-10
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
                    break
                case "window":
                    var w = h = args[k]
                    if (typeof(args[k]) == "object") {
                        w = args[k][0]
                        h = args[k][1]
                    }

                    args["top"] = h/2
                    args["left"] = w/2
                    args["width"] = width-w
                    args["height"] = height-h
                    break
            }
        }

        for (var k in args) {
            switch (k) {
                case "top":
                case "left":
                case "width":
                case "height":
                    target.style[k] = args[k]+"px"
                    break
            }
        }
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
                            node[k][d] = html[k][d]
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

        // include require styles // 加载文件
        // name
        // click
        // style
        // button
        // tree, fork, leaf // 树状结构
        // code, text, view // 普通视图
        // type, data, list // 基本结构

        var kit = this

        subs = subs || {}
        children.forEach(function(child, i) {
            child.data = child.data || {}
            child.type = child.type || "div"

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
            if (child.click) {
                child.data["onclick"] = child.click
            }

            if (child.include) {
                child.data["src"] = child.include[0]
                child.data["type"] = "text/javascript"
                child.include.length > 1 && (child.data["onload"] = child.include[1])
                child.type = "script"

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

            } else if (child.button) {
                child.type = "button"
                child.data["innerText"] = child.button[0]
                child.data["onclick"] = child.button[1]

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
    InsertChild: function (parent, position, element, children) {
        var elm = this.CreateNode(element)
        this.AppendChild(elm, children)
        return parent.insertBefore(elm, position || parent.firstElementChild)
    },

    AppendTable: function(table, data, fields, cb) {
        if (!data || !fields) {
            return
        }
        var kit = this
        var tr = kit.AppendChild(table, "tr")
        fields.forEach(function(key, j) {
            var td = kit.AppendChild(tr, "th", key)
        })
        data.forEach(function(row, i) {
            var tr = kit.AppendChild(table, "tr")
            fields.forEach(function(key, j) {
                var td = kit.AppendChild(tr, "td", row[key])
                if (typeof cb == "function") {
                    td.onclick = function(event) {
                        cb(row[key], key, row, i, event)
                    }
                }
            })
        })
    },
    RangeTable: function(table, index, sort_asc) {
        var list = table.querySelectorAll("tr")
        var new_list = []

        var is_time = true
        var is_number = true
        for (var i = 1; i < list.length; i++) {
            var value = Date.parse(list[i].childNodes[index].innerText)
            if (!(value > 0)) {
                is_time = false
            }

            var value = parseInt(list[i].childNodes[index].innerText)
            if (!(value >= 0 || value <= 0)) {
                is_number = false
            }

            new_list.push(list[i])
        }

        var sort_order = ""
        if (is_time) {
            if (sort_asc) {
                method = function(a, b) {return Date.parse(a) > Date.parse(b)}
                sort_order = "time"
            } else {
                method = function(a, b) {return Date.parse(a) < Date.parse(b)}
                sort_order = "time_r"
            }
        } else if (is_number) {
            if (sort_asc) {
                method = function(a, b) {return parseInt(a) > parseInt(b)}
                sort_order = "int"
            } else {
                method = function(a, b) {return parseInt(a) < parseInt(b)}
                sort_order = "int_r"
            }
        } else {
            if (sort_asc) {
                method = function(a, b) {return a > b}
                sort_order = "str"
            } else {
                method = function(a, b) {return a < b}
                sort_order = "str_r"
            }
        }

        list = new_list
        new_list = []
        for (var i = 0; i < list.length; i++) {
            list[i].parentElement && list[i].parentElement.removeChild(list[i])
            for (var j = i+1; j < list.length; j++) {
                if (typeof method == "function" && method(list[i].childNodes[index].innerText, list[j].childNodes[index].innerText)) {
                    var temp = list[i]
                    list[i] = list[j]
                    list[j] = temp
                }
            }
            new_list.push(list[i])
        }

        for (var i = 0; i < new_list.length; i++) {
            table.appendChild(new_list[i])
        }
        return sort_order
    },
    OrderTable: function(table, field, cb) {
        if (!table) {return}
        var kit = this
        table.onclick = function(event) {
            var target = event.target
            var dataset = target.dataset
            var nodes = target.parentElement.childNodes
            for (var i = 0; i < nodes.length; i++) {
                if (nodes[i] == target) {
                    if (target.tagName == "TH") {
                        dataset["sort_asc"] = (dataset["sort_asc"] == "1") ? 0: 1
                        kit.RangeTable(table, i, dataset["sort_asc"] == "1")
                    } else if (target.tagName == "TD") {
                        var tr = target.parentElement.parentElement.querySelector("tr")
                        if (tr.childNodes[i].innerText.startsWith(field)) {
                            typeof cb == "function" && cb(event)
                        }
                        kit.CopyText()
                    }
                }
            }
        }
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
                    page.Runs(page, option)
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
                                        ctx.Search("componet_group", "")
                                    }
                                })
                                return
                            }

                            page.Runs(page, option)
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
                            page.Runs(page, option)
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
}

function right(arg) {
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
}
function format_date(arg) {
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
}

