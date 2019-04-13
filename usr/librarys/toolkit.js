kit = toolkit = {
    History: {cmd: [], txt: [], key: [],
        add: function(type, data) {
            var list = this[type] || []
            data && list.push({time: Date.now(), data: data})
            this[type] = list
            return list.length
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
    isMobile: navigator.userAgent.indexOf("Mobile") > -1,

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

        // tree, code, text, view, click
        // type, name, data, list
        var kit = this

        subs = subs || {}
        children.forEach(function(child, i) {
            child.data = child.data || {}
            child.type = child.type || "div"

            if (child.button) {
                child.type = "button"
                child.data["innerText"] = child.button[0]
                child.data["onclick"] = child.button[1]
            } else if (child.click) {
                child.data["onclick"] = child.click[0]
            } else if (child.fork) {
                child.type = "li"
                child.list = [
                    {"text": [child.fork[0], "div"]},
                    {"type": "ul", "list": child.fork[1]},
                ]
            } else if (child.leaf) {
                child.type = "li"
                child.data["innerText"] = child.leaf[0]
                if (child.leaf.length > 1 && typeof child.leaf[1] == "function") {
                    child.data["onclick"] = function(event) {
                        child.leaf[1](event, node)
                    }
                }
            } else if (child.view) {
                child.data["className"] = child.view[0]
                child.type = child.view.length > 1? child.view[1]: "div"
                child.view.length > 2 && (child.data["innerText"] = child.view[2])
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
    OrderCode: function(code) {
        if (!code) {return}

        var kit = this
        code.onclick = function(event) {
            kit.CopyText()
        }
    },
    OrderForm: function(page, form, append, result) {
        if (!form) {return}
        form.onactions = form.onactions || function(msg) {
            if (!msg) {return}
            append.innerHTML = ""
            kit.AppendTable(append, ctx.Table(msg), msg.append)
            result.innerHTML = ""
            result.innerText = msg.result.join("")
        }

        form.querySelectorAll("select").forEach(function(select, index, array) {
            select.onchange = function(event) {
                if (index == array.length-1) {
                    page.Runs(page, form)
                    return
                }
                if (array[index+1].type == "button") {
                    array[index+1].click()
                    return
                }
                array[index+1].focus()
            }
        })
        form.querySelectorAll("input").forEach(function(input, index, array) {
            switch (input.type) {
                case "button":
                    input.onclick = function(event) {
                        if (index == array.length-1) {
                            page.Runs(page, form)
                            return
                        }
                        if (array[index+1].type == "button") {
                            array[index+1].click()
                            return
                        }
                        array[index+1].focus()
                    }
                default:
                    input.onkeyup = function(event) {
                        if (event.key != "Enter") {
                            return
                        }
                        if (index == array.length-1) {
                            page.Runs(page, form)
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

    CopyText: function(text) {
        text = window.getSelection().toString()
        if (text == "") {return}
        kit.Log(text)
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

