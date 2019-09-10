
{init: function(run, field, option, output) {
    var stop = false
    return {
        ondaemon: {
            table: function(msg, cb) {
                if (stop) {return}
                var plugin = field.Plugin
                output.innerHTML = "", msg.append && kit.OrderTable(kit.AppendTable(kit.AppendChild(output, "table"), ctx.Table(msg), msg.append), "", function(event, value, name, line, index) {
                    if (name == "id") {
                        page.Sync("plugin_"+plugin.exports[0]).set(plugin.onexport[plugin.exports[2]||""](value, name, line))
                    } else {
                        var td = event.target
                        function submit(event) {
                            td.innerText = event.target.value
                            if (event.target.value != value) {
                                stop = true
                                plugin.Run(event, [option.title.value, name, index-1, event.target.value], function() {
                                    plugin.Check()
                                    stop = false
                                })
                            }
                        }
                        kit.AppendChilds(td, [{type: "input", value: value, data: {onblur: function(event) {
                            submit(event)
                        }, onkeyup: function(event) {
                            switch (event.key) {
                                case "Enter":
                                    break
                                case "Tab":
                                    break
                                default:
                                    return
                            }
                            event.stopPropagation()
                            event.preventDefault()
                        }, onkeydown: function(event) {
                            console.log(event.key)
                            switch (event.key) {
                                case "Enter":
                                    if (event.shiftKey) {
                                        td.parentNode.previousSibling.querySelector("td").click()
                                    } else {
                                        td.parentNode.nextSibling.querySelector("td").click()
                                    }
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
                })
                typeof cb == "function" && cb(msg)
            },
        },

    }
}}
