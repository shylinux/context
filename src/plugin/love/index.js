Plugin["love/index.js"] = function(field, option, output) {return {
	data: function(event) {var plugin = field.Plugin
        plugin.Run(event, [option.table.value], function(msg) {
            plugin.ondaemon(msg)
		})
	},
    show: function(event) {var plugin = field.Plugin
		plugin.Check(event, undefined, function(msg) {
			plugin.Run(event, [option.table.value], function(msg) {
				kit.List(msg.Table(), function(line) {
					kit.Selector(output, ".s"+line.when.split(" ")[0].split("-").join(""), function(item) {
						kit.classList.add(item.parentNode, "select")
						item.parentNode.title = line.what
                        plugin.ontoast(line.what, line.where)
					})
				}, 200)
			}, true)
		})
    },
    show_after: function(msg) {
        kit.Selector(output, ".s"+ kit.time().split(" ")[0].split("-").join(""), function(item) {
            kit.classList.add(item.parentNode, "today")
        })
    },
	onexport: {"": function(value, name, line) {var plugin = field.Plugin
		switch (field.Meta.name) {
			case "days": plugin.flash(line, function(list) {
				return kit.AppendChilds(output, list)
			}); break
			case "date":
				plugin.Check(event, undefined, function(msg) {
					kit.Selector(output, ".s"+line.when.split(" ")[0].split("-").join(""), function(item) {
						kit.classList.add(item.parentNode, "select")
						item.parentNode.title = line.what
					})
				})
				break
			case "detail":
				kit.Change(event.target, function(value) {
					plugin.Run(event, ["update", option.table.value, option.index.value, line.key, value], function(msg) {
						kit.Log("ok")
					})
				})
				break
		}
		return line.id
	}},
	flash: function(line, cb, index, array) {var plugin = field.Plugin
		var now = new Date()
		var day = new Date(line.when)
		var mis = parseInt((day.getTime() - now.getTime()) / 1000 / 3600 / 24)
		if (array && index == array.length-1) {
			mis = 999999
		}

		var list = kit.Span()
		list.span(["距", "day"], line.when.split(" ")[0])
		list.span(["在", "day"], line.where)
		list.span([line.what, "what"])
		list.span(mis>0? "还有": "过去", [mis, mis>0? "day1": "day0"], "天")

		kit.Opacity(cb([{text: [list.join(""), "div", "day"]}]).last)
	},
	Order: function(t, cb, cbs) {var plugin = field.Plugin
		kit.List(plugin.msg.Table().concat([{when: "9999-01-08", what: "最后一次爱你", where: "北京市"}]), function(line, index, array) {
			plugin.flash(line, cb, index, array)
		}, t, cbs)
	},
    Flash: function(event) {var plugin = field.Plugin
        plugin.Order(1000, function(list) {
            return kit.AppendChilds(output, list)
        }, function() {
			output.innerHTML = "", plugin.Order(400, function(list) {
                return kit.AppendChild(output, list)
            })
        })
    },
}}
