Plugin["mind/index.js"] = function(field, option, output) {return {
    onfigure: shy(function(type, msg, cb) {if (msg.event && msg.event.type == "blur") {return}
        var plugin = field.Plugin
        output.innerHTML = "", msg.append && kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), msg.Table(), msg.append), "", function(event, value, name, line, index) {
            if (name == "id") {
                plugin.onexport(event, value, name, line)

            } else {
                var td = event.target
                function submit(event) {
                    (td.innerText = event.target.value) != value && plugin.Run(event, [option.title.value, name, index-1, event.target.value], plugin.Check)
                }

                kit.AppendChilds(td, [{type: "input", value: value, data: {onblur: submit, onkeyup: function(event) {
                    switch (event.key) {
                        case "Enter":
                        case "Tab":
                            break
                        default:
                            return
                    }
                    event.stopPropagation()
                    event.preventDefault()

                }, onkeydown: function(event) {
                    switch (event.key) {
                        case "Enter":
                            var s = td.parentNode[event.shiftKey?"previousSibling":"nextSibling"]
                            s && s.querySelector("td").click()
                            break
                        case "Tab":
                            if (event.shiftKey) {
                                if (td.previousSibling) {
                                    td.previousSibling.click()
                                } else {
                                    td.parentNode.previousSibling.querySelector("td").click()
                                }
                            } else {
                                if (td.nextSibling) {
                                    td.nextSibling.click()
                                } else {
                                    td.parentNode.nextSibling.querySelector("td").click()
                                }
                            }
                            break
                        default:
                            return
                    }
                    event.stopPropagation()
                    event.preventDefault()
                }}}]).first.focus()
            }
        }), typeof cb == "function" && cb(msg)
    }),
}}
