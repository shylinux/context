{init: function(run, field, option, output) {
    return {
        Run: function(event, args, cb) {
            run(event, ["share", args[0]], function(msg) {
                kit.AppendChilds(output, [{img: [msg.result.join("")]}])
                typeof cb == "function" && cb({})
            })
        },
    }
}}
