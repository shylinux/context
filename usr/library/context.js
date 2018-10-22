context = {
    GET: function(url, form, cb) {
        form = form || {}

        var args = [];
        for (var k in form) {
            if (form[k] instanceof Array) {
                for (i in form[k]) {
                    args.push(k+"="+encodeURIComponent(form[k][i]));
                }
            } else if (form[k] != undefined) {
                args.push(k+"="+encodeURIComponent(form[k]));
            }
        }

        var arg = args.join("&");
        arg && (url += ((url.indexOf("?")>-1)? "&": "?") + arg)
        console.log("GET: "+url);

        var xhr = new XMLHttpRequest();
        xhr.open("GET", url);
        xhr.setRequestHeader("Accept", "application/json")

        xhr.onreadystatechange = function() {
            if (xhr.readyState != 4) {
                return
            }
            if (xhr.status != 200) {
                return
            }

            try {
                var msg = JSON.parse(xhr.responseText||'{"result":[]}');
            } catch (e) {
                var msg = {"result": [xhr.responseText]}
            }

            console.log(msg)
            msg.result && console.log(msg.result.join(""));
            if (msg.page_redirect) {
                location.href = msg.page_redirect.join("")
            }
            typeof cb == "function" && cb(msg)
        }
        xhr.send();
    },
}
