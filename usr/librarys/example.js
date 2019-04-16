exp = example = {
    __proto__: ctx,
    _init: function(page) {
        page.__proto__ = this
        return this
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
