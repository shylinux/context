Plugin["favor/macro.js"] = function(field, option, output) {return {
    Record: function() {
        if (confirm("run script "+option.mac.value)) {
            page.script("record", option.mac.value)
        }
    },
    Replay: function() {
        if (confirm("run script "+option.mac.value)) {
            page.script("replay", option.mac.value)
        }
    },
    all: function() {var plugin = field.Plugin
        option.mac.value = "", plugin.Runs(window.event, function() {
            page.Sync("plugin_"+plugin.exports[0]).set(plugin.onexport[plugin.exports[2]||""]("", "name", {name: ""}))
        })
    },
    Run: function(event, args, cb) {var plugin = field.Plugin
        var script = page.script()
        if (args[0] && !script[args[0]]) {
            return confirm("create script "+args[0]) && page.script("create", args[0])
        }

        plugin.msg = args[0]? ({append: ["index", "script"],
            index: kit.List(script[args[0]], function(item, index) {return index+""}),
            script: kit.List(script[args[0]], function(item) {return item.join(" ")}),

        }): ({append: ["name", "count"],
            name: kit.Item(script),
            count: kit.Item(script, function(key, list) {return list.length+""}),

        }), plugin.display("table", cb)
    },
}}
