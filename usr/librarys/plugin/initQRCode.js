{
    init: function(page, pane, plug, form, output) {
        form.Runs = function(event) {
            var url = "/chat/?componet_group=index&componet_name=login&cmds=qrcode&cmds="+form.content.value
            output.innerHTML = "", kit.AppendChild(output, [{img: [url]}])
            event.ctrlKey? page.target.Send("icon", url): form.Run(event, [form.content.value])
        }
    },
}
