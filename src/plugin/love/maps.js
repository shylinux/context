Plugin["love/maps.js"] = function(field, option, output) {
var id
return {
    initMap: function() {var plugin = field.Plugin
        !id && (id = plugin.id+"map"+plugin.ID())

        var width = field.parentNode.clientWidth-40
        kit.AppendChilds(output, [{type: "div", data: {id: id}, style: {width: width+"px", height: width*(kit.device.isMobile? 7: 3)/5}}])

        map = new BMap.Map(id)
        map.addControl(new BMap.NavigationControl())
        map.addControl(new BMap.ScaleControl())
        map.addControl(new BMap.OverviewMapControl())
        map.addControl(new BMap.MapTypeControl())
		return map
    },
	Current: function() {
		var geo = new BMap.Geolocation()
		geo.getCurrentPosition(function(p) {
			option.city.value = p.address.city
			option.where.value = kit.Value(kit.Value(p.address.street, "")+kit.Value(p.address.street_number, ""), p.address.city)
			map.centerAndZoom(p.point, map.getZoom())
		})
	},
    Search: function() {var plugin = field.Plugin
        plugin.initMap()
        var g = new BMap.Geocoder()
        g.getPoint(option.where.value, function(p) {
            kit.Log("where", p)
            if (!p) {alert("not found"); return}
            map.centerAndZoom(p, 18)
        }, option.city.value)
    },
    Record: function() {var plugin = field.Plugin
		function trunc(value, len) {
			len = kit.Value(len, 1000000)
			return parseInt(value*len)/parseFloat(len)
		}
        var l = map.getCenter()
        plugin.Run(event, [option.table.value, option.when.value, option.what.value, option.city.value, option.where.value,
			"longitude", trunc(l.lng), "latitude", trunc(l.lat), "scale", map.getZoom()])
    },
    Flashs: function() {var plugin = field.Plugin
        plugin.initMap(), plugin.Run(event, [option.table.value], function(msg) {
            kit.List(msg.Table(), plugin.place, 1000)
        }, true)
    },
	place: function(line) {
		var p = new BMap.Point(line.longitude, line.latitude)
		map.centerAndZoom(p, line.scale)

		var info = new BMap.InfoWindow(line.when+"<br/>"+line.where, {width: 200, height: 100, title: line.what})
		map.openInfoWindow(info, map.getCenter())

		output.style.opacity = 0
		kit.Opacity(output)
	},
	onexport: {"": function(value, name, line) {var plugin = field.Plugin
		plugin.initMap(), plugin.place(line)
		return line.id
	}},
}}

