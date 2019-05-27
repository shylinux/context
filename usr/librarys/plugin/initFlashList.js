{
    show: function(event, item, option, plugin) {
        var args = item.value == "所有"? ["all"]: []
        option.Run(event, args, function(msg) {
            option.ondaemon(msg)
        })
    },
    init: function(page, pane, plugin, option, output) {
        option.ondaemon = function(msg) {
            output.innerHTML = ""
            kit.AppendChild(output, [{type: "code", list: [{text: [msg.result.join(""), "pre"]}]}])
        }
        output.innerHTML = "hello"
    }
}
