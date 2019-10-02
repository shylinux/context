{init: function(run, field, option, output) {
return {
    onexport: {"": function(value, name, line) {var plugin = field.Plugin
        kit.AppendChilds(output, [{img: ["/download/"+line.hash], data: {width: output.clientWidth, onclick: function() {
            plugin.display("table")
        }}}])
    }},
    show: function() {var plugin = field.Plugin
        var msg = plugin.msg
        var width = output.clientWidth
        output.innerHTML = "", kit.List(ctx.Table(msg), function(line) {
            kit.Opacity(kit.AppendChilds(output, [{img: ["/download/"+line.hash], data: {width: width, onclick: function(event) {
            }}}]).last)
        }, 1000, function() {
            output.innerHTML = "", kit.List(ctx.Table(msg), function(line) {
                kit.Opacity(kit.AppendChild(output, [{img: ["/download/"+line.hash], data: {width: 200, onclick: function(event) {
                    plugin.ontoast({width: width, height: width*3/5+40,
                        text: {img: ["/download/"+line.hash], data: {width: width-20, onclick: function(event) {
                            plugin.ontoast()
                        }}}, button: ["确定"], cb: function() {
                            plugin.ontoast()
                        }
                    })
                }}}]).last)
            }, 500)
        })

    },
}}}

