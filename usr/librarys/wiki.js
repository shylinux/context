function action(event, cmd) {
    switch (cmd) {
        case "toggle_nav":
            var nav = document.querySelector("nav")
            nav.hidden = !nav.hidden
            var article = document.querySelector("article")
            if (!context.isMobile) {
                article.style.width = nav.hidden? "80%": "calc(100% - 250px)"
            }
            break
        case "toggle_list":
            var list = document.querySelector(".list")
            list.hidden = !list.hidden
            break
        case "toggle_menu":
            var menu = document.querySelector(".menu")
            menu.hidden = !menu.hidden
            break
        case "toggle_link":
            var link = document.querySelector(".link")
            link.hidden = !link.hidden
            break
    }
}

function init_menu() {
    var max = 0;
    var min = 1000;
    var list = [];
    var hs = ["h2", "h3", "h4"];
    for (var i = 0; i < hs.length; i++) {
        var head = document.getElementsByTagName(hs[i]);
        for (var j = 0; j < head.length; j++) {
            head[j].id = "head"+head[j].offsetTop;
            head[j].onclick = function(event) {}
            list.push({"level": hs[i], "position": head[j].offsetTop, "title": head[j].innerText, "hash": head[j].id})
            if (head[j].offsetTop > max) {
                max = head[j].offsetTop;
            }
            if (head[j].offsetTop < min) {
                min = head[j].offsetTop;
            }
        }
    }

    max = max - min;
    for (var i = 0; i < list.length-1; i++) {
        for (var j = i+1; j < list.length; j++) {
            if (list[j].position < list[i].position) {
                var a = list[i];
                list[i] = list[j];
                list[j] = a;
            }
        }
    }

    var index = [-1, 0, 0]
    for (var i = 0; i < list.length; i++) {
        if (list[i].level == "h2") {
            index[0]++;
            index[1]=0;
            index[2]=0;
        } else if (list[i].level == "h3") {
            index[1]++;
            index[2]=0;
        } else {
            index[2]++;
        }

        list[i].index4 = index[2];
        list[i].index3 = index[1];
        list[i].index2 = index[0];
    }

    var m = document.getElementsByClassName("menu");
    for (var i = 0; i < m.length; i++) {
        for (var j = 0; j < list.length; j++) {
            var text = list[j].index2+"."
            if (list[j].level == "h3") {
                text += list[j].index3
            } else if (list[j].level == "h4") {
                text += list[j].index3+"."+list[j].index4
            }

            text += " "
            text += list[j].title;

            var h = document.getElementById(list[j].hash)
            h.innerText = text

            var one = append_child(m[i], "li")
            var a = append_child(one, "a")
            a.href = "#"+list[j].hash;
            a.innerText = text+" ("+parseInt((list[j].position-min)/max*100)+"%)";

            one.className = list[j].level;
        }
    }
}

function init_link() {
    var link = [];
    var a = document.querySelectorAll(".text a");
    for (var i = 0; i < a.length; i++) {
        link.push({href: a[i].href, title: a[i].innerText});
    }
    var m = document.getElementsByClassName("link");
    for (var i = 0; i < m.length; i++) {
        var one = append_child(m[i], "li")
        var a = one.appendChild(document.createTextNode("相关链接: "));

        for (var j = 0; j < link.length; j++) {
            var one = append_child(m[i], "li")
            var a = one.appendChild(document.createTextNode(link[j].title+": "));
            var a = append_child(one, "a");
            a.href = link[j].href
            a.innerText = a.href
        }
    }
}

function init_code() {
    var m = document.getElementsByTagName("pre");
    for (var i = 0; i < m.length; i++) {
        var line = (m[i].clientHeight-10)/15
        // if (line < 3) {
        // 	continue
        // }
        console.log(m[i].clientHeight)
        var nu = m[i].parentElement.insertBefore(document.createElement("div"), m[i]);
        nu.className = "number1"

        for (var j = 1; j <= line; j++) {
            console.log(j)
            var li = nu.appendChild(document.createElement("div"));
            li.appendChild(document.createTextNode(""+j));
        }
    }
}

window.onload = function() {
    init_menu()
    init_link()
    init_code()
    var article = document.querySelector("article")
    var mav = document.querySelector("nav")
    alert(context.isMobile)
    if (context.isMobile) {
        article.style.width = "100%"
    } else {
        article.style.maxHeight = "calc(100% - 80px)"
        mav.style.maxHeight = "calc(100% - 80px)"
    }
}
