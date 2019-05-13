exp = example = {
    __proto__: ctx,
    _init: function(page) {
        page.__proto__ = this

        var body = document.body
        body.onkeydown = function(event) {
            page.onscroll && page.onscroll(event, body, "scroll")
        }
        return this
    },
    initHeader: function(page, field, option, output) {
        return [{"text": ["shycontext", "div", "title"]}]
    },
    initField: function(page, field, option, output) {
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
    initFooter: function(page, field, option) {
        return [{"view": ["title", "div", "<a href='mailto:shylinux@163.com'>shylinux@163.com</>"]}]
    },

    onscroll: function(event, target, action) {
        var page = this
        switch (action) {
            case "scroll":
                if (event.target == document.body) {
                    kit.ScrollPage(event, page.conf)
                }
                break
        }
    },
    onresize: function(event) {},
    reload: function() {
        location.reload()
    },

    eachField: function(page, cb) {
        document.querySelectorAll("body>fieldset").forEach(function(pane) {
            // pane init
            pane.ShowDialog = function(width, height) {
                return kit.ShowDialog(this, width, height)
            }
            pane.Size = function(width, height) {
                pane.style.display = width==0? "none": "block"
                pane.style.width = width+"px"
                pane.style.height = height+"px"
            }


            // form init
            var form = pane.querySelector("form.option")
            form.Run = function(cmds, cb) {
                ctx.Run(page, form.dataset, cmds, cb)
            }
            form.Runs = function(cmds, cb) {
                ctx.Run(page, form.dataset, cmds, function(msg) {
                    ctx.Table(msg, function(line, index) {
                        cb(line, index, msg)
                    })
                })
            }

            page[form.dataset.componet_name] = pane
            typeof cb == "function" && cb(page[pane.dataset.init], pane, form)
        })
    },
}

function Page(page) {
    window.onload = function() {
        page.init(exp._init(page))
    }
    return page
}
