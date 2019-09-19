{init: function(run, field, option, output) {
    return {
        show: function(event) {
            run(event, ["", "", "cmd", "ssh.data", "show", "love"], function(msg) {
                ctx.Table(msg, function(value) {
                    var ts = ".s"+value.when.split(" ")[0].split("-").join("")
                    kit.Selector(output, ts, function(item) {
                        item.parentNode.style.backgroundColor = "red"
                        item.parentNode.title = value.where
                    })
                })
            })
        },
    }
}}
