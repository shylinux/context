var code = {
    showmap: false,
    keymap: [],
    inputs: [],
    ninput: 0,

    ntext: 1,
    quick_txt: false,

    ncommand: 1,
    current_cmd: "",

    show_result: false,
    show_height: "30px",
    hide_height: "14px",
    scroll_x: 50,
    scroll_y: 50,
}

// 粘贴板
function save_clipboard(item) {
    var txt = []
    var li = item.parentElement.children
    for (var i = 0; i < li.length; i++) {
        if (li[i].dataset["text"]) {
            txt.push(li[i].dataset["text"])
        }
    }

    context.GET("", {
        "componet_bench": context.Search("bench"),
        "componet_group": "index",
        "componet_name": "command",
        "cmd": "bench "+context.Search("bench")+".clipstack"+" '"+JSON.stringify(txt)+"'"
    }, function(msg) {
        alert("保存成功")
    })
}
function copy_to_clipboard(text, skip_blur, skip_docker) {
    var clipboard = modify_node(".clipboard", {"value": text})
    if (skip_blur) {
        document.execCommand("copy")
    } else {
        clipboard.select()
        document.execCommand("copy")
        clipboard.blur()
    }

    var clipstack = document.querySelector("#clipstack")
    insert_child(clipstack, "option").value = text
    clipstack.childElementCount > 3 && clipstack.removeChild(clipstack.lastElementChild)

    if (skip_docker) {
        return
    }

    var txt = document.querySelector("div.workflow>ul>li>ul.txt")
    var target = append_child(txt, "li", {
        "innerText": format_date(new Date())+" "+(code.ntext++)+": "+text,
        "dataset": {
            "text": text,
            "action": "copy_txt",
        },
        "onclick": function(event) {
            if (event.altKey) {
                target.parentElement.removeChild(target)
                return
            }

            if (event.shiftKey) {
                var cmd = document.querySelector("form.option.command"+code.current_cmd+" input[name=cmd]")
                cmd && (cmd.value += " "+text)
                return
            }
            copy_to_clipboard(text, false, true)
        },
    })
}

function add_keymap(input) {
    if (code.ninput < code.keymap.length && input.style.display != "none") {
        input.title = "keymap: "+code.keymap[code.ninput]
        input.dataset["keymap"] = code.keymap[code.ninput]
        insert_before(input, "label", {
            "innerText": "("+code.keymap[code.ninput]+")",
            "className": "keymap" + (code.showmap? " show": " hide"),
        })
        code.inputs[code.keymap[code.ninput++]] = input
    }
    return input
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

                    var has = document.querySelector("td.clip")
                    has && (has.className = "")
                    target.className = "clip"

                    if (event.shiftKey) {
                        copy_to_clipboard(target.innerText, false, false)
                        return
                    }

                    var text = window.getSelection().toString()
                    if (text) {
                        copy_to_clipboard(text, true, !event.shiftKey)
                    }
                }
            }
        }
    }
}

// 命令行
function del_command(target) {
    var can_remove = false
    var order = -1
    for (var p = target; p.parentElement && p.tagName != "FIELDSET"; p = p.parentElement) {
        if (p.tagName == "FORM" && p.dataset["componet_name_alias"]) {
            can_remove = true
            order = p.dataset["componet_name_order"]
        }
    }
    if (p && can_remove) {
        p.parentElement.removeChild(p)
    }

    for (;order < code.ncommand; order++) {
        var input = document.querySelector("form.option.command"+order+" input[name=cmd]")
        if (input) {
            input.focus()
            return
        }
    }
    for (;order >= 0; order--) {
        var input = document.querySelector("form.option.command"+(order? order: "")+" input[name=cmd]")
        code.ncommand = order+1
        if (input) {
            input.focus()
            return
        }
    }
}
function shrink_command_result() {
    code.show_result = !code.show_result
    document.querySelectorAll("form.option input[name=cmd]").forEach(function(input) {
        for (var command = input; command.tagName != "FIELDSET"; command = command.parentElement) {}
        var append = command.querySelector("table.append")
        var result = command.querySelector("code.result pre")
        // append.style.display = (code.show_result||!append.querySelector("tr"))? "": "none"
        result.style.height = (code.show_result||result.innerText=="")? "": code.show_height
    })
}
function add_command(init) {
    var order = code.ncommand
    var name = "command"+code.ncommand++

    var fieldset = append_child(document.querySelector("body"), "fieldset")
    append_child(fieldset, "legend", {"innerText": name})

    var option = append_child(fieldset, "form", {
        "className": "option "+name,
        "dataset": {
            "componet_group": "index",
            "componet_name": "command",
            "componet_bench": context.Search("bench"),
            "componet_name_alias": name,
            "componet_name_order": order,
        }
    })

    append_child(option, "input", {"style": {"display": "none"}})
    add_keymap(append_child(option, "input", {
        "name": "cmd", "className": "cmd",
        "onkeyup": function(event) {onaction(event, "input")},
        "onfocus": function(event) {code.current_cmd=order}
    })).focus()

    add_sort(append_child(fieldset, "table", {"className": "append "+name}))
    append_child(append_child(fieldset, "code", {"className": "result "+name}), "pre")

    if (init) {
        return fieldset
    }

    var cmds = document.querySelector("div.workflow ul.cmd")
    var cmd = append_child(cmds, "li", {
        "innertText": order+": ",
        "className": name,
        "dataset": {
            "cmd": order,
        }
    })
    return fieldset
}
function send_command(form, cb) {
    var data = {}
    for (var key in form.dataset) {
        data[key] = form.dataset[key]
    }
    for (var i = 0; i < form.length; i++) {
        if (form[i].name) {
            data[form[i].name] = form[i].value
        }
    }

    var order = (data["componet_name_order"]||"")
    var cmd = document.querySelector("div.workflow>ul>li>ul>li.command"+order)
    cmd && (cmd.innerText = format_date(new Date())+" "+order+": "+data["cmd"])

    context.GET("", data, function(msg) {
        msg = msg[0]

        var name = data["componet_name_alias"] || data["componet_name"]
        var result = document.querySelector("code.result."+name+" pre")
        var append = document.querySelector("table.append."+name)

        if (append && msg) {
            append.innerHTML = ""
            if (msg.append) {
                var tr = append_child(append, "tr")
                for (var i = 0; i < msg.append.length; i++) {
                    append_child(tr, "th", msg.append[i]+"("+(i+1)+")")
                }

                var ncol = msg.append.length
                var nrow = msg[msg.append[0]].length
                for (var i = 0; i < nrow; i ++) {
                    var tr = append_child(append, "tr")
                    for (var k in msg.append) {
                        append_child(tr, "td", msg[msg.append[k]][i])
                    }
                }

                result && ( result.style.height = code.show_result? "": code.show_height)
            }
        }

        if (result && msg) {
            if (msg["Content-Type"] && msg["Content-Type"].join("") == "text/html") {
                append_child(result, "iframe").innerHTML = (msg.result || []).join("")
            } else {
                result.innerHTML = (msg.result || []).join("")
            }
            init_result()
        }

        typeof(cb) == "function" && cb(msg)
    })
}
function check_option(form, target, cb) {
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
    send_command(form, cb)
}

function add_history(input, cmd) {
    var dataset = event.target.dataset
    var history = JSON.parse(input.dataset["history"] || "[]")
    if (history.length == 0 || cmd != history[history.length-1]) {
        history.push(cmd)
    }
    dataset["history_last"] = history.length-1
    dataset["history"] = JSON.stringify(history)
}
function get_history(input, index, cmd) {
    var history = JSON.parse(input.dataset["history"] || "[]")
    var last = input.dataset["history_last"]
    if (last >= 0 && last < history.length) {
        last = (parseInt(last)+index+history.length) % history.length
        input.dataset["history_last"] = last
        cmd = history[last]
    }
    return cmd
}

function onaction(event, action, arg) {
    var target = event.target
    var dataset = target.dataset

    switch (action) {
        case "scroll":
            var body = document.getElementsByTagName("body")[0]
            if (target.tagName == "BODY") {
                switch (event.key) {
                    case "h":
                        if (event.ctrlKey) {
                            window.scrollBy(-code.scroll_x*10, 0)
                        } else {
                            window.scrollBy(-code.scroll_x, 0)
                        }
                        break
                    case "H":
                        window.scrollBy(-body.scrollWidth, 0)
                        break
                    case "l":
                        if (event.ctrlKey) {
                            window.scrollBy(code.scroll_x*10, 0)
                        } else {
                            window.scrollBy(code.scroll_x, 0)
                        }
                        break
                    case "L":
                        window.scrollBy(body.scrollWidth, 0)
                        break
                    case "j":
                        if (event.ctrlKey) {
                            window.scrollBy(0, code.scroll_y*10)
                        } else {
                            window.scrollBy(0, code.scroll_y)
                        }
                        break
                    case "J":
                        window.scrollBy(0, body.scrollHeight)
                        break
                    case "k":
                        if (event.ctrlKey) {
                            window.scrollBy(0, -code.scroll_y*10)
                        } else {
                            window.scrollBy(0, -code.scroll_y)
                        }
                        break
                    case "K":
                        window.scrollBy(0, -body.scrollHeight)
                        break
                }
            }
            return
        case "keymap":
            if (target.tagName == "INPUT" && target.type == "text") {
                return
            }
            if (event.ctrlKey) {
                switch (event.key) {
                    case "b":
                        var item = document.querySelector("div.workflow>div")
                        item.onclick()
                        break
                    case "p":
                        var target = document.querySelector("div.workflow>ul>li>ul>li[data-action=quick_txt]")
                        code.quick_txt = !code.quick_txt
                        target.className= code.quick_txt? "stick": ""
                        break
                    case "y":
                        copy_to_clipboard(prompt("text"))
                        break
                    case "s":
                        save_clipboard(document.querySelector("div.workflow>ul>li>ul.txt>li[data-action=save_txt"))
                        break
                    case "m":
                        add_command()
                        break
                    case "z":
                        shrink_command_result()
                        var target = document.querySelector("div.workflow>ul>li>ul>li[data-action=shrink_cmd]")
                        target.className = code.show_result? "": "stick"
                        break
                    case "r":
                        location.reload()
                        break
                    case "t":
                        location.search = ""
                        break
                    case "g":
                        document.querySelectorAll("form.option label.keymap").forEach(function(item) {
                            code.showmap = !(item.className == "keymap show")
                            item.className = code.showmap? "keymap show": "keymap hide"
                        })
                        break
                    case "0":
                        document.querySelector("form.option.command input[name=cmd]").focus()
                        break
                    case "1":
                    case "2":
                    case "3":
                    case "4":
                    case "5":
                    case "6":
                    case "7":
                    case "8":
                    case "9":
                        document.querySelector("form.option.command"+event.key+" input[name=cmd]").focus()
                        break
                }
                return
            }

            switch (event.key) {
                default:
                    if (code.inputs[event.key]) {
                        code.inputs[event.key].focus()
                    }
                    break
            }
            break
        case "input":
            if (event.key == "Escape") {
                target.blur()
                break
            }
            if (event.key == "Enter") {
                check_option(target.form, target)
                add_history(target, target.value)
                break
            }

            for (var command = target; command.tagName != "FIELDSET"; command = command.parentElement) {}
            var option = command.querySelector("form.option")
            var append = command.querySelector("table.append")
            var result = command.querySelector("code.result pre")

            if (event.ctrlKey) {
                switch (event.key) {
                    case "1":
                    case "2":
                    case "3":
                    case "4":
                    case "5":
                    case "6":
                    case "7":
                    case "8":
                    case "9":
                        if (code.quick_txt) {
                            var item = document.querySelectorAll("div.workflow>ul>li>ul.txt>li[data-text]")
                            target.value += item[parseInt(event.key)-1].dataset["text"]
                        } else {
                            var item = document.querySelectorAll("table.append.command"+(parseInt(option.dataset.componet_name_order)-1)+" td")
                            if (event.shiftKey) {
                                var item = document.querySelectorAll("table.append.command1 td")
                            }
                            target.value += item[parseInt(event.key)-1].innerText
                        }
                        break
                    case "0":
                        var pre_pre = document.querySelector("code.result.command"+(parseInt(option.dataset.componet_name_order)-1)+" pre")
                        target.value += pre_pre.innerText
                        break
                    case "a":
                    case "e":
                    case "b":
                    case "f":
                        break
                    case "h":
                    case "d":
                    case "k":
                        break
                    case "w":
                        var value = target.value
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

                        target.dataset["old_string"] = value.substr(i+1, value.length)
                        target.value = value.substr(0, i+1)
                        break
                    case "u":
                        if (target.value != "") {
                            target.dataset["old_string"] = target.value
                        }
                        target.value = ""
                        break
                    case "v":
                        target.value += target.dataset["old_string"] || ""
                        break
                    case "g":
                        var value = target.value
                        var search = []
                        document.querySelectorAll("form.option input[name=cmd]").forEach(function(input) {
                            if (input.value.startsWith(value)) {
                                search.push(input.value)
                            }
                        })
                        if (search) {
                            target.value = search[0]
                        }
                        break
                    case "j":
                        check_option(target.form, target)
                        add_history(target, target.value)
                        break
                    case "p":
                        target.value = get_history(target, -1, target.value)
                        break
                    case "n":
                        target.value = get_history(target, 1, target.value)
                        break
                    case "c":
                        append.innerHTML = ""
                        result.innerHTML = ""
                        break
                    case "z":
                        // append.style.display = (result.style.height||!append.querySelector("tr"))? "": "none"
                        result.style.height = result.style.height? "": code.show_height
                        break
                    case "x":
                        result.style.height = result.style.height? "": code.hide_height
                        break
                    case "s":
                        copy_to_clipboard(result.innerText)
                        break
                    case "r":
                        append.innerHTML = ""
                        result.innerHTML = ""
                        check_option(option)
                        break
                    case "l":
                        window.scrollTo(0, command.offsetTop)
                        break
                    case "i":
                        for (var order = (parseInt(option.dataset["componet_name_order"])||0)+1; order < code.ncommand; order++) {
                            var input = document.querySelector("form.option.command"+order+" input[name=cmd]")
                            if (input) {
                                input.focus()
                                return
                            }
                        }
                        break
                    case "o":
                        for (var order = parseInt(option.dataset["componet_name_order"])-1; order >= 0; order--) {
                            var input = document.querySelector("form.option.command"+(order? order: "")+" input[name=cmd]")
                            if (input) {
                                input.focus()
                                return
                            }
                        }
                        break
                    case "m":
                        add_command()
                        break
                    case "q":
                        del_command(event.target)
                        break
                }
            }

            if (dataset["last_char"] == "j" && event.key == "k") {
                target.value = target.value.substr(0, target.value.length-2)
                target.blur()
            }

            dataset["last_char"] = event.key
            break
        case "click":
            if (target.nodeName == "INPUT" && event.altKey) {
                var board = document.querySelector(".clipboard")
                target.value = board.value
                check_option(target.form, target)
            }
            break
        case "command":
            check_option(target.form, target, function(msg) {
                if (target["value"] == "login") {
                    location.reload()
                }
            })
            break
        case "toolkit":
            if (event.key == "Enter") {
                context.GET("", {"toolkit": dataset["cmd"], "argument": target.value}, function(msg) {
                    target.nextElementSibling.innerText = msg.result
                    event
                })
            }
            break
    }
}

function init_option() {
    code.inputs = {}
    code.ninput = 0
    code.keymap =[]
    for (var i = 97; i < 123; i++) {
        switch (i) {
            case "j".charCodeAt(0):
            case "k".charCodeAt(0):
            case "h".charCodeAt(0):
            case "l".charCodeAt(0):
                continue
        }
        code.keymap.push(String.fromCharCode(i))
    }
    document.querySelectorAll("form.option input").forEach(add_keymap)
}
function init_append(event) {
    var append = document.querySelectorAll("table.append").forEach(add_sort)
}
function init_result(event) {
    var result = document.querySelectorAll("code.result pre").forEach(function(item) {
        item.onclick = function(event) {
            var text = window.getSelection().toString()
            if (text) {
                copy_to_clipboard(text, true, !event.shiftKey)
            }
        }
    })
}

function init_userinfo() {
    var option = document.querySelector("form.option.userinfo")
    var append = document.querySelector("table.append.userinfo")
    if (!append) {return}

    insert_button(append, "logout", function(event) {
        context.Cookie("sessid", "")
        location.reload()
    })
}
function init_download(event) {
    var option = document.querySelector("form.option.dir")
    var append = document.querySelector("table.append.dir")
    if (!append) {return}

    function change(dir, show, run) {
        if ((dir.endsWith(".sh") || dir.endsWith(".shy") || dir.endsWith(".py")) && !show) {
            var command = document.querySelector("form.option.command")
            var cmd = command["cmd"]
            cmd.value = "run "+ dir.split("/").pop()
            cmd.focus()
            if (run) {
                check_option(command)
            }
            return
        }

        option["dir"].value = dir
        if (dir == "" || dir.endsWith("/")) {
            context.Cookie("download_dir", option["dir"].value)
        }
        send_command(option)
        option["dir"].value = context.Cookie("download_dir")

    }
    insert_button(append, "root", function(event) {
        change("")
    })
    insert_button(append, "back", function(event) {
        var path = option["dir"].value.split("/")
        while (path.pop() == "") {}
        change(path.join("/")+(path.length? "/": ""))
    })

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

    add_sort(append, "filename", function(event) {
        var dir = event.target.innerText
        if (option["dir"].value && !option["dir"].value.endsWith("/")) {
            change(option["dir"].value+"/"+dir, event.altKey, event.shiftKey)
        } else {
            change(option["dir"].value+dir, event.altKey, event.shiftKey)
        }
    })
}

function init_context() {
    var option = document.querySelector("form.option.ctx")
    var append = document.querySelector("table.append.ctx")
    if (!append) {return}

    function change(ctx) {
        option["ctx"].value = ctx
        send_command(option)
        context.Cookie("current_ctx", option["ctx"].value)
        return ctx
    }

    var contexts = ["ctx", "shy", "web", "mdb"]
    for (var i = 0; i < contexts.length; i++) {
        insert_button(append, contexts[i], function(event) {
            change(event.target.value)
        })
    }
    add_sort(append, "name", function(event) {
        change(event.target.innerText.trim())
    })
}
function init_command() {
    if (bench.commands[""]) {
        var option = document.querySelector("form.option.command")
        var cmd = option.querySelector("input[name=cmd]")
        cmd.value = bench.commands[""].cmd.join(" ")
        check_option(option)
    }

    var max = 0
    for (var k in bench.commands) {
        if (parseInt(k) > max) {
            max = parseInt(k)
        }
    }

    for (var i = 1; i <= max; i++) {
        var fieldset = add_command(true)
        if (bench.commands[i]) {
            var option = fieldset.querySelector("form.option")
            var cmd = option.querySelector("input[name=cmd]")
            cmd.value = bench.commands[i].cmd.join(" ")
            check_option(option)
        }
    }
}
function init_docker() {
    text = JSON.parse(bench.clipstack || "[]")
    for (var i = 0; i < text.length; i++) {
        copy_to_clipboard(text[i])
    }

    document.querySelectorAll("div.workflow").forEach(function(workflow) {
        // 移动面板
        workflow.style.left = context.Cookie("docker_left")
        workflow.style.top = context.Cookie("docker_top")
        var moving = false, left0 = 0, top0 = 0, x0 = 0, y0 = 0
        workflow.onclick = function(event) {
            if (event.target != workflow) {
                return
            }
            moving = !moving
            if (moving) {
                left0 = workflow.offsetLeft
                top0 = workflow.offsetTop
                x0 = event.clientX
                y0 = event.clientY
            }
        }
        workflow.onmousemove = function(event) {
            if (moving) {
                workflow.style.left = (left0+(event.clientX-x0))+"px"
                workflow.style.top = (top0+(event.clientY-y0))+"px"
                context.Cookie("docker_left", workflow.style.left)
                context.Cookie("docker_top", workflow.style.top)
            }
        }

        // 固定面板
        if (context.Cookie("docker_class")) {
            workflow.className = context.Cookie("docker_class")
        }
        var head = workflow.querySelector("div")
        head.onclick = function(event) {
            head.dataset["show"] = !right(head.dataset["show"])
            workflow.className = right(head.dataset["show"])? "workflow max": "workflow"
            context.Cookie("docker_class", workflow.className)
        }

        // 折叠目录
        var docker = workflow.querySelector("ul.docker")
        docker.querySelectorAll("li>div").forEach(function(menu) {
            menu.onclick = function(event) {
                menu.dataset["hide"] = !right(menu.dataset["hide"])
                menu.nextElementSibling.style.display = right(menu.dataset["hide"])? "none": ""
            }
        })

        // 事件
        docker.querySelectorAll("li>ul>li").forEach(function(item) {
            if (bench["key"] == item.dataset["key"]) {
                item.className = "stick"
            }

            item.onclick = function(event) {
                var target = event.target
                var data = item.dataset
                switch (data["action"]) {
                    case "quick_txt":
                        code.quick_txt = !code.quick_txt
                        target.className= code.quick_txt? "stick": ""
                        break
                    case "copy_txt":
                        if (event.altKey) {
                            target.parentElement.removeChild(target)
                            return
                        }
                        if (event.shiftKey) {
                            var cmd = document.querySelector("form.option.command"+code.current_cmd+" input[name=cmd]")
                            cmd && (cmd.value += " "+text)
                            return
                        }
                        copy_to_clipboard(data["text"], true)
                        break
                    case "save_txt":
                        save_clipboard(item)
                        return
                    case "create_txt":
                        var text = prompt("text")
                        text && copy_to_clipboard(text)
                        return
                    case "shrink_cmd":
                        shrink_command_result()
                        var target = document.querySelector("div.workflow>ul>li>ul>li[data-action=shrink_cmd]")
                        target.className = code.show_result? "": "stick"
                        return
                    case "create_cmd":
                        add_command()
                        return
                    case "refresh_fly":
                        location.reload()
                        return
                    case "create_fly":
                        location.search = ""
                        return
                    case "rename_fly":
                        context.GET("", {
                            "componet_bench": context.Search("bench"),
                            "componet_group": "index",
                            "componet_name": "command",
                            "cmd": "bench "+context.Search("bench")+".comment"+" "+prompt("name"),
                        })
                        location.reload()
                        return
                    case "remove_fly":
                        context.GET("", {
                            "componet_bench": context.Search("bench"),
                            "componet_group": "index",
                            "componet_name": "command",
                            "cmd": "~code bench delete "+context.Search("bench"),
                        })
                        var b = ""
                        document.querySelectorAll("div.workflow>ul.docker>li>ul.fly>li[data-key]").forEach(function(item){
                            if (!b && item.dataset["key"] != context.Search("bench")) {
                                b = item.dataset["key"]
                            }
                        })
                        context.Search("bench", b)
                        return
                }

                // 切换工作流
                if (data["key"] && data["key"] != bench["key"]) {
                    context.Search("bench", data["key"])
                    return
                }

                // 切换命令行
                var cmd = document.querySelector("form.option.command"+data["cmd"]+" input[name=cmd]")
                cmd && cmd.focus()
            }
        })
    })
}

window.onload = function() {
    init_option()
    init_append()
    init_result()

    init_userinfo()
    init_download()

    init_context()
    init_command()
    init_docker()
}

