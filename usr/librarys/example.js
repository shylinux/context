exp = example = {
    __proto__: ctx,
    _init: function(page) {
        page.__proto__ = this
        document.querySelectorAll("body>fieldset").forEach(function(field) {
            var init = page[field.dataset.init]
            if (typeof init == "function") {
                var option = field.querySelector("form.option")
                var append = field.querySelector("table.append")
                var result = field.querySelector("div.result")
                var conf = init(page, field, option, append, result)
                if (conf && conf["button"]) {
                    var buttons = []
                    conf.button.forEach(function(value, index) {
                        buttons.push({"click": [value, function(event) {
                            typeof conf["action"] == "function" && conf["action"](value, event)
                        }]})
                    })
                    kit.InsertChild(field, append, "div", buttons)
                }
                if (conf && conf["table"]) {
                    option.onactions = function(msg) {
                        append.innerHTML = ""
                        kit.AppendTable(append, ctx.PackAppend(msg), msg.append, function(value, key, row, index, event) {
                            typeof conf["table"][key] && conf["table"][key](value, key, row, index, event)
                        })
                    }
                    ctx.Runs(page, option)
                }
            }
        })
        return this
    },
    reload: function() {
        location.reload()
    },
    _exit: function(page) {
    },
}

function Page(page) {
    window.onload = function() {
        page.init(exp._init(page))
    }
    return page
}
