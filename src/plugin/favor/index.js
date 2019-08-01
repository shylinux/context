
{init: function(page, pane, plugin, field, option, output) {
    kit.Log("hello world")
    plugin.Run = function(event, args, cb) {
        field.Run(event, ["share", args[0]], function(url) {
            kit.AppendChild(output, [{img: [url]}])
        })
    }
}}
