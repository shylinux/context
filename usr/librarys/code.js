function copy_to_clipboard(text) {
    var clipboard = document.querySelector(".clipboard")
    clipboard.value = text
    clipboard.select()
    document.execCommand("copy")
    clipboard.blur()

    var clipstack = document.querySelector("#clipstack")
    insert_child(clipstack, "option").value = text
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
        var append = document.querySelector("table.append."+data["componet_name"])
        if (msg && (msg.append || msg.result)) {
            result && (result.innerHTML = (msg.result || []).join(""))
            append && (append.innerHTML = "")
        }

        if (append && msg && msg.append) {
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
        case "submit":
            break
        case "click":
            if (event.target.nodeName == "INPUT") {
                if (event.altKey) {
                    var board = document.querySelector(".clipboard")
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
                    var history = JSON.parse(event.target.dataset["history"] || "[]")
                    if (history.length == 0 || event.target.value != history[history.length-1]) {
                        history.push(event.target.value)
                        var clistack = document.querySelector("#clistack")
                        insert_child(clistack, "option").value = event.target.value
                    }
                    check_argument(event.target.form, event.target)
                    event.target.dataset["history"] = JSON.stringify(history)
                    event.target.dataset["history_last"] = history.length-1
                    console.log(history.length)
                    break
                case "Escape":
                    if (event.target.value) {
                        event.target.value = ""
                    } else {
                        event.target.blur()
                    }
                    break
                case "w":
                    if (event.ctrlKey) {
                        var value = event.target.value
                        var space = value.length > 0 && value[value.length-1] == ' '
                        for (var i = value.length-1; i > -1; i--) {
                            if (space) {
                                if (value[i] != ' ') {
                                    break
                                }
                            } else {
                                if (value[i] == ' ') {
                                    break
                                }
                            }
                        }
                        event.target.value = value.substr(0, i+1)
                        break
                    }
                case "u":
                    if (event.ctrlKey && event.key == "u") {
                        event.target.value = ""
                        break
                    }
                case "p":
                    if (event.ctrlKey && event.key == "p") {
                        var history = JSON.parse(event.target.dataset["history"] || "[]")
                        var last = event.target.dataset["history_last"]
                        console.log(last)
                        event.target.value = history[last--]
                        event.target.dataset["history_last"] = (last + history.length) % history.length
                        return false
                        break
                    }
                case "n":
                    if (event.ctrlKey && event.key == "n") {
                        var history = JSON.parse(event.target.dataset["history"] || "[]")
                        var last = event.target.dataset["history_last"]
                        last = (last +1) % history.length
                        console.log(last)
                        event.target.value = history[last]
                        event.target.dataset["history_last"] = last
                        break
                    }
                case "j":
                    if (event.ctrlKey && event.key == "j") {
                        var history = JSON.parse(event.target.dataset["history"] || "[]")
                        if (history.length == 0 || event.target.value != history[history.length-1]) {
                            history.push(event.target.value)
                            var clistack = document.querySelector("#clistack")
                            insert_child(clistack, "option").value = event.target.value
                        }
                        check_argument(event.target.form, event.target)
                        event.target.dataset["history"] = JSON.stringify(history)
                        event.target.dataset["history_last"] = history.length-1
                        break
                    }
                default:
                    console.log(event)
                    if (event.target.dataset["last_char"] == "j" && event.key == "k") {
                        if (event.target.value) {
                            event.target.value = ""
                        } else {
                            event.target.blur()
                        }
                    }
                    event.target.dataset["last_char"] = event.key
            }
            break
        case "keymap":
            break
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
        item.onselect = function(event) {
            console.log(event)

        }
        item.onclick = function(event) {
            console.log(event)
            return
            copy_to_clipboard(event.target.innerText)
        }
    })
}
function init_download(event) {
    var append = document.querySelector("table.append.dir")
    insert_before(append, "input", {
        "type": "button",
        "value": "root",
        "onclick": function(event) {
            option["dir"].value = ""
            context.Cookie("download_dir", option["dir"].value)
            send_command(option)
            return true
        }
    })
    insert_before(append, "input", {
        "type": "button",
        "value": "back",
        "onclick": function(event) {
            var path = option["dir"].value.split("/")
            while (path.pop() == "") {}
            option["dir"].value = path.join("/")+(path.length? "/": "")
            context.Cookie("download_dir", option["dir"].value)
            send_command(option)
            return true
        }
    })

    var option = document.querySelector("form.option.dir")
    var sort_order = option["sort_order"]
    var sort_field = option["sort_field"]
    sort_field.innerHTML = ""
    sort_field.onchange = function(event) {
        switch (event.target.selectedOptions[0].value) {
            case "filename":
            case "type":
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

    var th = append.querySelectorAll("th")
    for (var i = 0; i < th.length; i++) {
        var value = th[i].innerText.trim()
        var opt = append_child(sort_field, "option", {
            "value": value, "innerText": value,
        })
    }

    (option["dir"].value = context.Search("download_dir")) && send_command(option)

    append.onchange = append.onclick = function(event) {
        console.log(event)
        if (event.target.tagName == "TD") {
            copy_to_clipboard(event.target.innerText.trim())
            var name = event.target.innerText.trim()
            if (option["dir"].value && !option["dir"].value.endsWith("/")) {
                option["dir"].value += "/"+name
            } else {
                option["dir"].value += name
            }
            if (name.endsWith("/")) {
                context.Cookie("download_dir", option["dir"].value)
            }
        } else if (event.target.tagName == "TH") {
            option["sort_field"].value = event.target.innerText.trim()

            switch (event.target.innerText.trim()) {
                case "filename":
                case "type":
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
        }

        send_command(option, function(){
            option["dir"].value = context.Cookie("download_dir")
        })
    }
}

function init_context() {
    var append = document.querySelector("table.append.ctx")
    var option = document.querySelector("form.option.ctx")
    insert_before(append, "input", {
        "type": "button",
        "value": "ctx",
        "onclick": function(event) {
            option["ctx"].value = "ctx"
            send_command(option)
            context.Cookie("current_ctx", option["ctx"].value)
            return true
        }
    })
    insert_before(append, "input", {
        "type": "button",
        "value": "shy",
        "onclick": function(event) {
            option["ctx"].value = "shy"
            send_command(option)
            context.Cookie("current_ctx", option["ctx"].value)
            return true
        }
    })
    insert_before(append, "input", {
        "type": "button",
        "value": "mdb",
        "onclick": function(event) {
            option["ctx"].value = "mdb"
            send_command(option)
            context.Cookie("current_ctx", option["ctx"].value)
            return true
        }
    })

    option["ctx"].value = context.Cookie("current_ctx")
    send_command(option)

    append.onchange = append.onclick = function(event) {
        console.log(event)
        if (event.target.tagName == "TD") {
            var name = event.target.innerText.trim()
            copy_to_clipboard(name)
            option["ctx"].value = name
            context.Cookie("current_ctx", option["ctx"].value)
        } else if (event.target.tagName == "TH") {
        }

        send_command(option)
    }
}

window.onload = function() {
    init_option()
    init_append()
    init_result()
    init_download()
    init_context()
}
