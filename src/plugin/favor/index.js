Plugin["favor/index.js"] = function(field, option, output) {return {
    share: function(event) {var plugin = field.Plugin
        plugin.Run(event, ["share", option.txt.value], function(msg) {
            kit.AppendChilds(output, [{img: [msg.result.join("")]}])
            typeof cb == "function" && cb({})
        })
    },
}}
