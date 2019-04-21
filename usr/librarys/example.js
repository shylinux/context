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
    initHeader: function(field, option, output) {
        return [{"text": ["shylinux", "div", "title"]}]
    },
    initBanner: function(field, option, output) {
        return [{"text": ["shylinux", "div", "title"]}]
    },
    initFooter: function(field, option) {
        return [{"text": ["shycontext", "div", "title"]}]
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
