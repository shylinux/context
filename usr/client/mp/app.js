//app.js
App({
    request: function(data, done, fail) {
        var app = this
        data = data || {}
        data.sessid = app.sessid || ""

        wx.request({url: "https://shylinux.com/chat/mp", data: data,
            success: function(res) {
                typeof done == "function" && done(res.data)
            },
            fail: function(res) {
                typeof done == "function" && done(res.data)
            },
        })
    },
    login: function(cb) {
        var app = this
        wx.login({success: res => {
            app.request({code: res.code}, function(sessid) {
                app.sessid = sessid

                wx.getSetting({success: res => {
                    if (res.authSetting['scope.userInfo']) {
                        wx.getUserInfo({success: res => {
                            app.userInfo = res.userInfo
                            app.request(res, cb)
                        }})
                    }
                }})
            })
        }})
    },

    onLaunch: function () {
        this.login()
    },
})
