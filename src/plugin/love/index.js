{init: function(run, field, option, output) {
return {
    show: function(event) {
        run(event, ["", "", "cmd", "ssh.data", "show", option.table.value], function(msg) {
			kit.List(ctx.Table(msg), function(value) {
				kit.Selector(output, ".s"+value[option.when.value].split(" ")[0].split("-").join(""), function(item) {
					kit.classList.add(item.parentNode, "select")
					item.parentNode.title = value[option.where.value]
				})
			}, 500)
		})
    },
    show_after: function(msg) {
        kit.Selector(output, ".s"+ kit.format_date().split(" ")[0].split("-").join(""), function(item) {
            kit.classList.add(item.parentNode, "today")
        })
    },
	Order: function(t, cb, cbs) {var plugin = field.Plugin
        var msg = plugin.msg, now = new Date()
		kit.List(ctx.Table(msg).concat([{when: "9999-01-08", what: "最后一次爱你"}]), function(line, index, array) {
			var day = new Date(line.when)
			var mis = parseInt((day.getTime() - now.getTime()) / 1000 / 3600 / 24)
			if (index == array.length-1) {
				mis = 999999
			}

			var list = kit.Span()
			list.span(["距", "day"], line.when.split(" ")[0]).span([line.what, "what"])
			list.span(mis>0? "还有": "过去", [mis, mis>0? "day1": "day0"], "天")

			kit.Opacity(cb(output, [{text: [list.join(""), "div", "day"]}]).last)
		}, t, cbs)
	},
    Flash: function(event) {var plugin = field.Plugin
        plugin.Order(3000, function(output, list) {
            return kit.AppendChilds(output, list)
        }, function() {
			output.innerHTML = "", plugin.Order(1000, function(output, list) {
                return kit.AppendChild(output, list)
            })
        })
    },
}}}
