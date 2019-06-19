function Page(page) {
    var id = 1
    var conf = {}
    var conf_cb = {}
    var sync = {}
    page.__proto__ = {
        ID: function() {
            return id++
        },
        Conf: function(key, value, cb) {
            if (value != undefined) {
                var old = conf[key]
                conf[key] = value
                conf_cb[key] && conf_cb[key](value, old)
            }
            if (cb != undefined) {
                conf_cb[key] = cb
            }
            if (key != undefined) {
                return conf[key]
            }
            return conf
        },
        Sync: function(m) {
            var meta = m, data = "", list = []
            return sync[m] || (sync[m] = {
                change: function(cb) {
                    list.push(cb)
                    return list.length-1
                },
                eq: function(value) {
                    return data == value
                },
                neq: function(value) {
                    return data != value
                },
                get: function() {
                    return data
                },
                set: function(value, force) {
                    if (value == data && !force) {
                        return value
                    }
                    old_value = data, data = value
                    meta && kit.Log(meta, value, old_value)
                    for (var i = 0; i < list.length; i++) {
                        list[i](value, old_value)
                    }
                    return value
                },
            })
        },
        View: function(parent, type, line, key, cb) {
            var ui = {}
            var result = []
            var text = line
            switch (type) {
                case "icon":
                    result = [{view: ["item", "div"], list: [{img: [line[key[0]], function(event) {
                        event.target.scrollIntoView()
                    }]}]}]
                    break

                case "text":
                    result = [{text: [key.length>1? line[key[0]]+"("+line[key[1]]+")": (key.length>0? line[key[0]]: "null"), "span"], click: cb}]
                    break

                case "code":
                    result = [{type: "code", list: [{text: [key.length>1? line[key[0]]+"("+line[key[1]]+")": (key.length>0? line[key[0]]: "null")], click: cb}]}]
                    break

                case "table":
                    result = [{view: [""], list: [
                        {view: ["", "table"], list: JSON.parse(line.text || "[]").map(function(item, index) {
                            return {type: "tr", list: item.map(function(value) {
                                return {text: [value, index == 0? "th": "td"]}
                            })}
                        })},
                    ]}]
                    break

                case "field":
                    var text = JSON.parse(line.text)

                case "plugin":
                    var id = "plugin"+page.ID()
                    result = [{name: "field", view: [text.view, "fieldset"], data: {id: id}, list: [
                        {text: [text.name+"("+text.help+")", "legend"]},
                        {name: "option", view: ["option", "form"], data: {Run: cb}, list: [{type: "input", style: {"display": "none"}}]},
                        {name: "output", view: ["output", "div"]},
                        {script: "Plugin("+id+","+JSON.stringify(text)+","+"[\""+(text.args||[]).join("\",\"")+"\"]"+","+(text.init||"")+")"},
                    ]}]
                    break
            }
            if (parent.DisplayUser) {
                ui = kit.AppendChild(parent, [{view: ["item"], list:[
                    {view: ["user", "div", line.create_nick]},
                    {view: ["text"], list:result}
                ]}])
            } else {
                ui = kit.AppendChild(parent, [{view: ["item"], list:result}])
            }

            ui.field && (ui.field.Meta = text)
            return ui
        },
        alert: function(text) {
            alert(text)
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
        oninput: function(event, local) {
            var target = event.target
            kit.History.add("key", (event.ctrlKey? "Control+": "")+(event.shiftKey? "Shift+": "")+event.key)

            if (event.ctrlKey) {
                if (typeof local == "function" && local(event)) {
                    event.stopPropagation()
                    return true
                }
                switch (event.key) {
                    case "a":
                    case "e":
                    case "f":
                    case "b":
                    case "h":
                    case "d":
                        break
                    case "k":
                        kit.DelText(target, target.selectionStart)
                        break
                    case "u":
                        kit.DelText(target, 0, target.selectionEnd)
                        break
                    case "w":
                        var start = target.selectionStart-2
                        var end = target.selectionEnd-1
                        for (var i = start; i >= 0; i--) {
                            if (target.value[end] == " " && target.value[i] != " ") {
                                break
                            }
                            if (target.value[end] != " " && target.value[i] == " ") {
                                break
                            }
                        }
                        kit.DelText(target, i+1, end-i)
                        break
                    default:
                        return false

                }
                event.stopPropagation()
                return true
            }
            switch (event.key) {
                case "Escape":
                    target.blur()
                    event.stopPropagation()
                    return true
                default:
                    if (kit.HitText(target, "jk")) {
                        kit.DelText(target, target.selectionStart-2, 2)
                        target.blur()
                        return true
                    }
            }
            return false
        },
        ontoast: function(text, title, duration) {
            var args = typeof text == "object"? text: {text: text, title: title, duration: duration}
            var toast = kit.ModifyView("fieldset.toast", {
                display: "block", dialog: [args.width||200, args.height||40], padding: 10,
            })

            var list = [{text: [args.text||""]}]
            args.inputs && args.inputs.forEach(function(input) {
                if (typeof input == "string") {
                    list.push({inner: input, type: "label", style: {"margin-right": "5px"}})
                    list.push({input: [input, page.oninput]})
                } else {
                    list.push({inner: input[0], type: "label", style: {"margin-right": "5px"}})
                    var option = []
                    for (var i = 1; i < input.length; i++) {
                        option.push({type: "option", inner: input[i]})
                    }
                    list.push({name: input[0], type: "select", list: option})
                }
                list.push({type: "br"})
            })
            args.button && args.button.forEach(function(input) {
                list.push({type: "button", inner: input, click: function(event) {
                    var values = {}
                    toast.querySelectorAll("input").forEach(function(input) {
                        values[input.name] = input.value
                    })
                    toast.querySelectorAll("select").forEach(function(input) {
                        values[input.name] = input.value
                    })
                    typeof args.cb == "function" && args.cb(input, values)
                    toast.style.display = "none"
                }})
            })

            kit.ModifyNode(toast.querySelector("legend"), args.title||"tips")
            var ui = kit.AppendChild(kit.ModifyNode(toast.querySelector("div.output"), ""), list)
            args.duration !=- 1 && setTimeout(function(){toast.style.display = "none"}, args.duration||3000)
            page.toast = toast
        },
        onscroll: function(event, target, action) {
            switch (action) {
                case "scroll":
                    if (event.target == document.body) {
                        kit.ScrollPage(event, page.conf)
                    }
                    break
            }
        },

        initHeader: function(page, pane, form, output) {
            var state = {}, list = [], cb = function(event, item, value) {
            }
            pane.Order = function(value, order, cbs) {
                state = value, list = order, cb = cbs || cb, pane.Show()
            }
            pane.Show = function() {
                output.innerHTML = "", kit.AppendChild(output, [
                    {"view": ["title", "div", "shycontext"], click: function(event) {
                        cb(event, "title", "shycontext")
                    }},
                    {"view": ["state"], list: list.map(function(item) {return {text: [state[item], "div"], click: function(event) {
                        cb(event, item, state[item])
                    }}})},
                ])
            }
            pane.State = function(name, value) {
                state[name] = value, pane.Show()
            }
            return
        },
        initBanner: function(page, field, option, output) {
            field.querySelectorAll("li").forEach(function(item) {
                item.onclick = function(event) {
                    ctx.Search("componet_group", item.innerText)
                    if (item.innerText == "login") {
                        ctx.Cookie("sessid", "")
                    }
                }
            })
            return [{"text": ["shylinux", "div", "title"]}]
        },
        initFooter: function(page, pane, form, output) {
            var state = {}, list = [], cb = function(event, item, value) {
            }
            pane.Order = function(value, order, cbs) {
                state = value, list = order, cb = cbs || cb, pane.Show()
            }
            pane.State = function(name, value) {
                if (value != undefined) {
                    state[name] = value, pane.Show()
                }
                if (name != undefined) {
                    return state[name]
                }
                return state
            }

            pane.Show = function() {
                output.innerHTML = "", kit.AppendChild(output, [
                    {"view": ["title", "div", "<a href='mailto:shylinux@163.com'>shylinux@163.com</>"]},
                    {"view": ["state"], list: list.map(function(item) {return {text: [item+":"+state[item], "div"]}})},
                ])
            }
            return
        },
        initField: function(page, cb) {
            document.querySelectorAll("body>fieldset").forEach(function(pane) {
                var form = pane.querySelector("form.option")
                page[form.dataset.componet_name] = pane

                // pane init
                pane.which = page.Sync(form.dataset.componet_name)
                pane.ShowDialog = function(width, height) {
                    if (pane.style.display != "block") {
                        page.dialog && page.dialog != pane && page.dialog.style.display == "block" && page.dialog.Show()
                        pane.style.display = "block", page.dialog = pane
                        kit.ModifyView(pane, {window: [width||80, height||200]})
                        return true
                    }
                    pane.style.display = "none"
                    delete(page.dialog)
                    return false
                }
                pane.Size = function(width, height) {
                    pane.style.display = (width<=0 || height<=0)? "none": "block"
                    pane.style.width = width+"px"
                    pane.style.height = height+"px"
                }

                var conf = {}
                var conf_cb = {}
                pane.Conf = function(key, value, cb) {
                    if (value != undefined) {
                        var old = conf[key]
                        conf[key] = value
                        conf_cb[key] && conf_cb[key](value, old)
                    }
                    if (cb != undefined) {
                        conf_cb[key] = cb
                    }
                    if (key != undefined) {
                        return conf[key]
                    }
                    return conf
                }

                // form init
                pane.Run = form.Run = function(cmds, cb) {
                    ctx.Run(page, form.dataset, cmds, cb)
                }
                pane.Runs = form.Runs = function(cmds, cb) {
                    ctx.Run(page, form.dataset, cmds, function(msg) {
                        ctx.Table(msg, function(line, index) {
                            cb(line, index, msg)
                        })
                    })
                }
                pane.Time = form.Time = function(time, cmds, cb) {
                    function loop() {
                        ctx.Run(page, form.dataset, cmds, cb)
                        setTimeout(loop, time)
                    }
                    setTimeout(loop, time)
                }

                var timer = ""
                pane.Times = form.Times = function(time, cmds, cb) {
                    timer && clearTimeout(timer)
                    function loop() {
                        !pane.Stop && ctx.Run(page, form.dataset, cmds, function(msg) {
                            ctx.Table(msg, function(line, index) {
                                cb(line, index, msg)
                            })
                        })
                        timer = setTimeout(loop, time)
                    }
                    time && (timer = setTimeout(loop, time))
                }
                form.onsubmit = function(event) {
                    event.preventDefault()
                }

                cb(page[pane.dataset.init], pane, form)
            })

            document.querySelectorAll("body>fieldset").forEach(function(pane) {
                for (var k in pane.Listen) {
                    page[k].which.change(pane.Listen[k])
                }
            })
        },
    }
    window.onload = function() {
        page.init(page)

        window.onresize = function(event) {
            page.onlayout && page.onlayout(event)
        }
        document.body.onkeydown = function(event) {
            page.onscroll && page.onscroll(event, document.body, "scroll")
        }
        document.body.onkeyup = function(event) {
            page.oncontrol && page.oncontrol(event, document.body, "control")
        }
    }
    return page
}
function Plugin(field, tool, args, plugin) {
    var option = field.querySelector("form.option")
    var output = field.querySelector("div.output")

    var exports = JSON.parse(tool.exports||'["",""]')
    var display = JSON.parse(tool.display||'{}')
    option.Runs = function(event) {
        option.Run(event, kit.Selector(option, ".args", function(item, index) {
            return item.value
        }), function(msg) {
            (option.ondaemon || function(msg) {
                output.innerHTML = ""
                !display.hide_append && msg.append && kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), ctx.Table(msg), msg.append), exports[1], function(event, value) {
                    if (exports.length > 2) {
                        if (value.endsWith("/")) {
                            value = option[exports[2]].value + value
                        } else {
                            return
                        }
                    }
                    page.Sync("plugin_"+exports[0]).set(value)
                });
                (display.show_result || !msg.append) && msg.result && kit.AppendChild(output, [{view: ["code", "div", msg.Results()]}])
            })(msg)
        })
    }

    var total = 0, count = 0
    plugin = plugin || {}, plugin.__proto__ = {
        show: function() {},
        init: function() {},
        Clone: function() {
            field.Meta.args = kit.Selector(option, ".args", function(item, index) {
                return item.value
            })
            page.View(field.parentNode, "plugin", field.Meta, [], option.Run)
        },
        Clear: function() {
            field.parentNode && field.parentNode.removeChild(field)
        },
        Check: function(event, index) {
            index == total-1 || (index == total-2 && event.target.parentNode.nextSibling.childNodes[1].type == "button")?
                option.Runs(event): event.target.parentNode.nextSibling.childNodes[1].focus()
        },
        Remove: function(who) {
            who.parentNode && who.parentNode.removeChild(who)
        },
        Append: function(item, name) {
            var index = total
            total += 1
            name = name || item.name

            item.onfocus = function(event) {
                page.plugin = plugin
                page.input = event.target
                page.footer.State(".", field.id)
                page.footer.State(":", index)
            }
            item.onkeyup = function(event) {
                page.oninput(event, function(event) {
                    switch (event.key) {
                        case "i":
                            var next = field.nextSibling;
                            next && next.Select()
                            break
                        case "o":
                            var prev = field.previousSibling;
                            prev && prev.Select()
                            break
                        case "c":
                            output.innerHTML = ""
                            break
                        case "r":
                            output.innerHTML = ""
                        case "j":
                            run(event)
                            break
                        case "l":
                            page.action.scrollTo(0, option.parentNode.offsetTop)
                            break
                        case "m":
                            plugin.Clone()
                            break
                        case "b":
                            plugin.Append(item, "args"+total).focus()
                            break
                        default:
                            return false
                    }
                    return true
                })
                event.key == "Enter" && plugin.Check(event, index)
            }

            var input = {type: "input", name: name, data: item}
            switch (item.type) {
                case "button":
                    item.onclick = function(event) {
                        action[item.click]? action[item.click](event, item, option, field):
                            plugin[item.click]? plugin[item.click](event, item, option, field): option.Runs(event)
                    }
                    break

                case "select":
                    input = {type: "select", name: name, data: {className: "args", onchange: function(event) {
                        plugin.Check(event, index)

                    }}, list: item.values.map(function(value) {
                        return {type: "option", value: value, inner: value}
                    })}
                    args && count < args.length && (item.value = args[count++])
                    break

                default:
                    args && count < args.length && (item.value = args[count++]||item.value||"")
                    item.className = "args"
            }

            var ui = kit.AppendChild(option, [{view: [item.view||""], list: [{type: "label", inner: item.label||""}, input]}])
            var action = ui[name] || {}

            page.plugin = field
            page.input = action
            index == 0 && action && action.focus && action.focus()

            action.History = []
            action.Goto = function(value) {
                action.value = value;
                (index == total-1 || (index == total-2 && action.parentNode.nextSibling.childNodes[1].type == "button")) && option.Runs(event)
                action.History.push(value)
                plugin.Back = function() {
                    action.History.pop()
                    action.History.length > 0 && action.Goto(action.History.pop())
                }
                return value
            }

            item.imports && typeof item.imports == "object" && item.imports.forEach(function(imports) {
                page.Sync(imports).change(action.Goto)
            })
            item.imports && typeof item.imports == "string" && page.Sync(item.imports).change(action.Goto)
            return action
        },
        Select: function() {
            page.plugin = field
            page.footer.State(".", field.id)
        },
    }

    var inputs = JSON.parse(tool.inputs || "[]")
    inputs.map(function(item, index, inputs) {
        plugin.Append(item)
    })

    plugin.init(page, page.action, field, option, output)
    page[field.id] = plugin
}
