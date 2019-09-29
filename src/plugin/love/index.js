{init: function(run, field, option, output) {
var id
return {
    initMap: function() {var plugin = field.Plugin
        var width = field.parentNode.clientWidth-40
        !id && (id = plugin.id+"map"+plugin.ID())
        kit.AppendChilds(output, [{type: "div", data: {id: id}, style: {width: width+"px", height: width*3/5}}])
        map = new BMap.Map(id)
        map.addControl(new BMap.NavigationControl())
        map.addControl(new BMap.ScaleControl())
        map.addControl(new BMap.OverviewMapControl())
        map.addControl(new BMap.MapTypeControl())
    },
    show: function(event) {
        run(event, ["", "", "cmd", "ssh.data", "show", option.table.value], function(msg) {ctx.Table(msg, function(value) {
            kit.Selector(output, ".s"+value[option.when.value].split(" ")[0].split("-").join(""), function(item) {
                kit.classList.add(item.parentNode, "select")
                item.parentNode.title = value[option.where.value]
            })
        })})
    },
    show_after: function(msg) {
        kit.Selector(output, ".s"+ kit.format_date().split(" ")[0].split("-").join(""), function(item) {
            kit.classList.add(item.parentNode, "today")
        })
    },
    play: function(event) {
        kit.AppendChilds(output, [{type: "video", data: {src: option.url.value, autoplay: ""}}])
    },
    Quick: function(event) {
        var msg = field.Plugin.msg
        var now = new Date()
        function show(t, cb, cbs) {
            kit.List(ctx.Table(msg).concat([{when: "9999-01-08", what: "最后一次爱你"}]), function(line, index, array) {
                var day = new Date(line.when)
                var mis = parseInt((day.getTime() - now.getTime()) / 1000 / 3600 / 24)
                if (index == array.length-1) {
                    mis = 999999
                }

                var list = []
                list.span = function(value, style) {
                    for (var i = 0; i < arguments.length; i++) {
                        if (typeof arguments[i] == "string") {
                            list.push(arguments[i])
                        } else {
                            list.push('<span class="'+arguments[i][1]+'">', arguments[i][0], "</span>")
                        }
                    }
                    list.push("<br/>")
                    return list
                }

                list.span(["距", "day"], line.when.split(" ")[0]).span([line.what, "what"])
                list.span(mis>0? "还有": "过去", [mis, mis>0? "day1": "day0"], "天")

                var elm = cb(output, [{text: [list.join(""), "div", "day"]}]).last
                kit.List([0.2, 0.4, 0.6, 0.8, 1.0], function(value) {
                    elm.style.opacity = value
                }, 150)
            }, t, cbs)
        }
        show(3000, function(output, list) {
            return kit.AppendChilds(output, list)
        }, function() {
            show(1000, function(output, list) {
                return kit.AppendChild(output, list)
            })
        })
    },
    Search: function() {var plugin = field.Plugin
        plugin.initMap()
        var g = new BMap.Geocoder()
        g.getPoint(option.place.value, function(p) {
            kit.Log("place", p)
            if (!p) {alert("not found"); return}
            map.centerAndZoom(p, option.scale.value)
        }, option.city.value)
    },
    Record: function() {var plugin = field.Plugin
        var l = map.getCenter()
        run(event, [option.table.value, option.when.value, option.what.value, option.city.value, option.place.value, l.lng, l.lat, map.getZoom()], function(msg) {
            console.log("ok")
        })
    },
    Plays: function() {var plugin = field.Plugin
        plugin.initMap(), run(event, [option.table.value], function(msg) {
            kit.List(ctx.Table(msg), function(line) {
                output.style.opacity = 0
                var p = new BMap.Point(line.longitude, line.latitude)
                map.centerAndZoom(p, line.scale)

                var info = new BMap.InfoWindow(line.when+"<br/>"+line.place, {width: 200, height: 100, title: line.what})
                map.openInfoWindow(info, map.getCenter())

                var scales = []
                for (var i = 0; i < 10; i++) {
                    scales.push(i*0.1)
                }
                kit.List(scales, function(scale) {
                    output.style.opacity = scale
                }, 100)
            }, 3000)
        })
    },

}}}
