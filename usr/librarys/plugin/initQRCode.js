{init: function(page, pane, field, option, output) {
    this.Runs = function(event) {
        var value = option.content.value
        var url = page.login.Pane.Share({cmds: ["qrcode", value]})
        kit.AppendChilds(output, [{img: [url]}])
        event.ctrlKey && page.target.Pane.Send("icon", url)
        event.shiftKey && page.target.Pane.Send("field", this.Format(value))
    }
}}
