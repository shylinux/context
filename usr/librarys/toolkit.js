kit = toolkit = {
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

        // code, text, view, click
        // type, name, data, list

        subs = subs || {}
        for (var i = 0; i < children.length; i++) {
            var child = children[i]
            child.data = child.data || {}
            child.type = child.type || "div"

            if (child.click) {
                child.type = "button"
                child.data["innerText"] = child.click[0]
                child.data["onclick"] = child.click[1]
            } else if (child.view) {
                child.data["className"] = child.view[0]
                child.type = child.view.length > 1? child.view[1]: "div"
                child.view.length > 2 && (child.data["innerText"] = child.view[2])
            } else if (child.text) {
                child.data["innerText"] = child.text[0]
                child.type = child.text.length > 1? child.text[1]: "pre"
                child.text.length > 2 && (child.data["className"] = child.text[2])
            } else if (child.code) {
                child.type = "code"
                child.list = [{"type": "pre" ,"data": {"innerText": child.code[0]}, "name": child.code[1]}]
                child.code.length > 2 && (child.data["className"] = child.code[2])
            }

            var node = this.CreateNode(child.type, child.data)
            child.list && this.AppendChild(node, child.list, subs)
            child.name && (subs[child.name] = node)
            parent.append(node)
        }
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

function modify_node(which, html) {
    var node = which
    if (typeof which == "string") {
        node = document.querySelector(which)
    }

    html && typeof html == "string" && (node.innerHTML = html)
    if (html && typeof html == "object") {
        for (var k in html) {
            if (typeof html[k] == "object") {
                for (var d in html[k]) {
                    node[k][d] = html[k][d]
                }
                continue
            }
            node[k] = html[k]
        }
    }
    return node
}
function create_node(element, html) {
    var node = document.createElement(element)
    return modify_node(node, html)
}

function insert_child(parent, element, html, position) {
    var elm = create_node(element, html)
    return parent.insertBefore(elm, position || parent.firstElementChild)
}
function append_child(parent, element, html) {
    var elm = create_node(element, html)
    parent.append(elm)
    return elm
}
function insert_before(self, element, html) {
    var elm = create_node(element, html)
    return self.parentElement.insertBefore(elm, self)
}
function insert_button(which, value, callback) {
    insert_before(which, "input", {
        "type": "button", "value": value, "onclick": callback,
    })
}

function format(str) {
    if (str.indexOf("http") == 0 && str.indexOf("<a href") == -1) {
        return "<a href='"+str+"' target='_blank'>"+str+"</a>"
    }
    return str
}
function sort_table(table, index, sort_asc) {
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
}
function add_sort(append, field, cb) {
    append.onclick = function(event) {
        var target = event.target
        var dataset = target.dataset
        var nodes = target.parentElement.childNodes
        for (var i = 0; i < nodes.length; i++) {
            if (nodes[i] == target) {
                if (target.tagName == "TH") {
                    dataset["sort_asc"] = (dataset["sort_asc"] == "1") ? 0: 1
                    sort_table(append, i, dataset["sort_asc"] == "1")
                } else if (target.tagName == "TD") {
                    var tr = target.parentElement.parentElement.querySelector("tr")
                    if (tr.childNodes[i].innerText.startsWith(field)) {
                        typeof cb == "function" && cb(event)
                    }
                }
            }
        }
    }
}
function scroll_page(event, page) {
    var body = document.querySelector("body")

    switch (event.key) {
        case "h":
            if (event.ctrlKey) {
                window.scrollBy(-page.scroll_x*10, 0)
            } else {
                window.scrollBy(-page.scroll_x, 0)
            }
            break
        case "H":
            window.scrollBy(-body.scrollWidth, 0)
            break
        case "l":
            if (event.ctrlKey) {
                window.scrollBy(page.scroll_x*10, 0)
            } else {
                window.scrollBy(page.scroll_x, 0)
            }
            break
        case "L":
            window.scrollBy(body.scrollWidth, 0)
            break
        case "j":
            if (event.ctrlKey) {
                window.scrollBy(0, page.scroll_y*10)
            } else {
                window.scrollBy(0, page.scroll_y)
            }
            break
        case "J":
            window.scrollBy(0, body.scrollHeight)
            break
        case "k":
            if (event.ctrlKey) {
                window.scrollBy(0, -page.scroll_y*10)
            } else {
                window.scrollBy(0, -page.scroll_y)
            }
            break
        case "K":
            window.scrollBy(0, -body.scrollHeight)
            break
    }
    return
}

