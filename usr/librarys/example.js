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
        ctx.Runs(page, option)
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
    reload: function() {
        location.reload()
    },
}

function Page(page) {
    window.onload = function() {
        page.init(exp._init(page))
    }
    return page
}
