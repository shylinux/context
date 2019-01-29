const app = getApp()

Page({
    data: {
        cmd: "",
        table: [],
        append: [],
        result: "",
    },
    onCommand: function(e) {
        var page = this
        app.request({"cmd": e.detail.value}, function(res) {
            if (res.append) {
                var table = []
                for (var i = 0; i < res[res.append[0]].length; i++) {
                    var line = []
                    for (var j = 0; j < res.append.length; j++) {
                        line.push(res[res.append[j]][i])
                    }
                    table.push(line)
                }
                page.setData({append: res.append, table: table})
            }
            page.setData({result: res.result? res.result.join("") :res})
        })
    },
    onLoad: function () {},
})
