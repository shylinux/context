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

        // 基本属性: name value inner style
        //
        // dataset click
        // button input label img
        // select
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
            if (child.className) {
                child.data["className"] = child.className
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
            if (child.dataset) {
                child.data["dataset"] = child.dataset
            }
            if (child.click) {
                child.data["onclick"] = child.click
            }

            if (child.button) {
                child.type = "button"
                child.data["onclick"] = child.button[1]
                child.data["innerText"] = child.button[0]
                child.name = child.name || child.button[0]

            } else if (child.select) {
                child.type = "select"
                child.list = child.select[0].map(function(value) {
                    return {type: "option", value: value, inner: value}
                })
                child.data["onchange"] = child.select[1]

            } else if (child.input) {
                child.type = "input"
                child.data["onkeyup"] = child.input[1]
                child.data["name"] = child.input[0]
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
            var tr = kit.AppendChild(table, "tr", {className: "normal"})
            tr.Meta = row
            fields.forEach(function(key, j) {
                var td = kit.AppendChild(tr, "td", row[key])
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
            var target = event.target
            var dataset = target.dataset
            var head = target.parentElement.parentElement.querySelector("tr")
            target.parentElement.childNodes.forEach(function(item, i) {
                if (item != target) {
                    return
                }
                if (target.tagName == "TH") {
                    dataset["sort_asc"] = (dataset["sort_asc"] == "1") ? 0: 1
                    kit.RangeTable(table, i, dataset["sort_asc"] == "1")
                    return
                }
                var name = head.childNodes[i].innerText
                if (name.startsWith(field)) {
                    typeof cb == "function" && cb(event, item.innerText, name,item.parentNode.Meta)
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
                ctx.Search({"wiki_class": text})
            })
        })
    },
    Position: function(which) {
        return (parseInt((which.scrollTop + which.clientHeight) / which.scrollHeight * 100)||0)+"%"
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
        obj.querySelectorAll(item).forEach(function(item, index) {
            if (typeof cb == "function") {
                var value = cb(item)
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
        return json.stringify(objs)
    },
    List: function(obj, cb) {
        var list = []
        for (var i = 0; i < obj.length; i++) {
            list.push(typeof cb == "function"? cb(obj[i]): obj[i])
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
                                        ctx.Search("componet_group", "")
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

