{init: function(run, field, option, output) {
    return {
        Run: function(event, args, cb) {
            run(event, ["share", args[0]], function(url) {
                kit.AppendChilds(output, [{img: [url]}])
                typeof cb == "function" && cb({})
            })
        },
    }
}}
