{init: function(field, option, output) {
    return {
        share: function() {var plugin = field.Plugin
            plugin.Run(event, ["share", args[0]], function(msg) {
                kit.AppendChilds(output, [{img: [msg.result.join("")]}])
                typeof cb == "function" && cb({})
            })
        },
    }
}}
