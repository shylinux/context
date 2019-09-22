{init: function(run, field, option, output) {return {
    show: function(event) {
        run(event, ["", "", "cmd", "ssh.data", "show", option.table.value], function(msg) {ctx.Table(msg, function(value) {
            kit.Selector(output, ".s"+value[option.when.value].split(" ")[0].split("-").join(""), function(item) {
                kit.classList.add(item.parentNode, "select")
                item.parentNode.title = value[option.where.value]
            })
        })})
    },
    show_after: function(msg) {
        kit.Selector(output, ".s"+ kit.format_date().split(" ")[0].split("-").join(""), function(item) {
            kit.classList.add(item.parentNode, "today")
        })
    },
    play: function(event) {
        kit.AppendChilds(output, [{type: "video", data: {src: option.url.value, autoplay: ""}}])
    },
}}}
