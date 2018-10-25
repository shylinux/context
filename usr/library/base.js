function insert_child(parent, element, html, position) {
    var elm = document.createElement(element)
    html && (elm.innerHTML = html)
    return parent.insertBefore(elm, position || parent.firstElementChild)
}
function append_child(parent, element, html) {
    var elm = document.createElement(element)
    html && (elm.innerHTML = html)
    parent.append(elm)
    return elm
}
function insert_before(self, element, html) {
    var elm = document.createElement(element)
    html && (elm.innerHTML = html)
    return self.parentElement.insertBefore(elm, self)
}

function copy_to_clipboard(text) {
    var clipboard = document.querySelector("#clipboard")
    clipboard.value = text
    clipboard.select()
    document.execCommand("copy")
    clipboard.blur()

    var clipstack = document.querySelector("#clipstack")
    insert_child(clipstack, "option").value = clipboard.value
    clipstack.childElementCount > 3 && clipstack.removeChild(clipstack.lastElementChild)
}
function send_command(form, cb) {
    var data = {}
    for (var key in form.dataset) {
        data[key] = form.dataset[key]
    }
    for (var i = 0; i < form.length; i++) {
        data[form[i].name] = form[i].value
    }

    context.GET("", data, function(msg) {
        msg = msg[0]

        var result = document.querySelector("code.result."+data["componet_name"]+" pre")
        result && (result.innerHTML = (msg.result || []).join(""))

        var append = document.querySelector("table.append."+data["componet_name"])
        append && (append.innerHTML = "")

        if (append && msg.append) {
            var tr = append_child(append, "tr")
            for (var i in msg.append) {
                append_child(tr, "th", msg.append[i])
            }

            var ncol = msg.append.length
            var nrow = msg[msg.append[0]].length
            for (var i = 0; i < nrow; i ++) {
                var tr = append_child(append, "tr")
                for (var k in msg.append) {
                    append_child(tr, "td", msg[msg.append[k]][i])
                }
            }
        }

        typeof(cb) == "function" && cb(msg)
    })
}
function check_argument(form, target) {
    for (var i = 0; i < form.length-1; i++) {
        if (form[i] == target) {
            if (form[i+1].type == "button") {
                form[i+1].click()
            } else {
                form[i+1].focus()
            }
            return false
        }
    }
    send_command(form)
}
function onaction(event, action) {
    switch (action) {
        case "click":
            if (event.target.nodeName == "INPUT") {
                if (event.altKey) {
                    var board = document.querySelector("#clipboard")
                    event.target.value = board.value
                    check_argument(event.target.form, event.target)
                }
            }
            break
        case "command":
            send_command(event.target.form)
            break
        case "input":
            switch (event.key) {
                case "Enter":
                    check_argument(event.target.form, event.target)
                    break
                case "Escape":
                    event.target.value = ""
                    event.target.blur()
                    break
            }
            break
        case "keymap":
            switch (event.key) {
                case "g":
                    document.querySelectorAll("form.option label.keymap").forEach(function(item) {
                        item.className = (item.className == "keymap show")? "keymap hide": "keymap show"
                    })
                    break
                default:
                    if (inputs[event.key]) {
                        inputs[event.key].focus()
                    }
                    break
            }
            break
        case "copy":
            copy_to_clipboard(event.target.innerText)
            break
    }
}
var inputs = {}
var ninput = 0
var keymap = ['a', 'b', 'c']
function init_option() {
    inputs = {}
    ninput = 0
    keymap =[]
    for (var i = 97; i < 123; i++) {
        if (i == 103) {
            continue
        }
        keymap.push(String.fromCharCode(i))
    }

    document.querySelectorAll("form.option input").forEach(function(input) {
        if (ninput < keymap.length && input.style.display != "none") {
            input.title = "keymap: "+keymap[ninput]
            input.dataset["keymap"] = keymap[ninput]
            insert_before(input, "label", "("+keymap[ninput]+")").className = "keymap"
            inputs[keymap[ninput++]] = input
        }
    })
}
function init_append(event) {
    var append = document.querySelectorAll("table.append").forEach(function(item) {
        item.onclick = function(event) {
            if (event.target.tagName == "TD") {
                copy_to_clipboard(event.target.innerText)
            }
        }
    })
}
function init_result(event) {
    var result = document.querySelectorAll("code.result pre").forEach(function(item) {
        item.onclick = function(event) {
            copy_to_clipboard(event.target.innerText)
        }
    })
}
function init_download(event) {
    var option = document.querySelector("form.option.dir")
    if (!option) {
        return
    }

    document.querySelector("form.option.dir input[name=dir]").value = context.Search("download_dir")

    option["dir"].value && send_command(option)

    var append = document.querySelector("table.append.dir")
    append.onchange = 
    append.onclick = function(event) {
        console.log(event)
        if (event.target.tagName == "A") {
            if (event.target.dataset.type != "true") {
                location.href = option["dir"].value+"/"+event.target.innerText
                return
            }

            option["dir"].value = option["dir"].value+"/"+event.target.innerText
            send_command(option, function(msg) {
                context.Cookie("download_dir", option["dir"].value = msg.dir.join(""))
            })
        } else if (event.target.tagName == "TD") {
            copy_to_clipboard(event.target.innerText)
        } else if (event.target.tagName == "TH") {
            option["sort_field"].value = event.target.innerText

            var sort_order = option["sort_order"]
            switch (event.target.innerText) {
                case "filename":
                case "is_dir":
                    sort_order.value = (sort_order.value == "str")? "str_r": "str"
                    break
                case "line":
                case "size":
                    sort_order.value = (sort_order.value == "int")? "int_r": "int"
                    break
                case "time":
                    sort_order.value = (sort_order.value == "time")? "time_r": "time"
                    break
            }

            send_command(option)
        }
    }
}

window.onload = function() {
    init_option()
    init_append()
    init_result()
    init_download()
}
