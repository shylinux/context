{
    init: function(page, pane, plug, form, output, ui) {
        form.Runs = function(event) {
            var url = "/chat/?componet_group=index&componet_name=login&cmds=qrcode&cmds="+ui.content.value
            output.innerHTML = "", kit.AppendChild(output, [{img: [url]}])
            event.ctrlKey? page.target.Send("icon", url): form.Run(event, [ui.content.value])
        }
    },
}
