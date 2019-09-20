{init: function(run, field, option, output) {return {
    show: function(event) {
        run(event, ["", "", "cmd", "ssh.data", "show", option.table.value], function(msg) {ctx.Table(msg, function(value) {
            kit.Selector(output, ".s"+value[option.when.value].split(" ")[0].split("-").join(""), function(item) {
                item.parentNode.style.backgroundColor = "red"
                item.parentNode.title = value[option.where.value]
            })
        })})
    },
    show_after: function(msg) {
        var now = kit.format_date(new Date())
        kit.Selector(output, ".s"+now.split(" ")[0].split("-").join(""), function(item) {
            item.parentNode.style.backgroundColor = "red"
            item.innerText = "TODAY"
            item.parentNode.title = "today"
        })
    },
}}}
