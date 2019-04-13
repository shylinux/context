function scan(event) {
    alert("begin scan")
    wx.scanQRCode({
        needResult: 0,
        scanType: ["qrCode", "barCode"],
        desc: "what",
        success: function(res) {
            alert(res.resultStr)
        },
        fail: function(res) {
            alert(res.errMsg)
        },
    })
}
function close() {
    wx.closeWindow({
        success: function(res) {
            alert(res.resultStr)
        },
        fail: function(res) {
            alert(res.errMsg)
        },
    })
}

function choose() {
    wx.chooseImage({
        count: 1, // 默认9
        sizeType: ['original', 'compressed'], // 可以指定是原图还是压缩图，默认二者都有
        sourceType: ['album', 'camera'], // 可以指定来源是相册还是相机，默认二者都有
        success: function (res) {
            var localIds = res.localIds; // 返回选定照片的本地ID列表，localId可以作为img标签的src属性显示图片
        },
        fail: function(res) {
            alert(res.errMsg)
        },
    });
}
function wopen(event) {
    wx.openAddress({success: function(res) {
        context.Command("show", res)
    }})
}

