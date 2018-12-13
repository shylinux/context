function action(event, cmd) {
    switch (cmd) {
        case "toggle_nav":
            var nav = document.querySelector("nav")
            nav.hidden = !nav.hidden
            set_layout(event)
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
            head[j].id = hs[i]+"_"+j
            head[j].onclick = function(event) {
                location.hash=event.target.id
            }
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
    var fuck = context.isMobile? 22: 16
    var m = document.getElementsByTagName("pre");
    for (var i = 0; i < m.length; i++) {
        var nu = insert_before(m[i], "div", {"className": "number1"})

        var line = (m[i].clientHeight-10)/fuck
        for (var j = 1; j <= line; j++) {
            append_child(nu, "div", {
                "style": {
                    "fontSize": context.isMobile?"20px":"14px",
                    "lineHeight": context.isMobile?"22px":"16px",
                },
                "id": "code"+i+"_"+"line"+j,
                "onclick": function(event) {
                    location.href = "#"+event.target.id
                },
            }).appendChild(document.createTextNode(""+j));
        }
    }
}

function add_sort(append, field, cb) {
    append.onclick = function(event) {
        var target = event.target
        var dataset = target.dataset
        var nodes = target.parentElement.childNodes
        for (var i = 0; i < nodes.length; i++) {
            if (nodes[i] == target) {
                if (target.tagName == "TH") {
                    dataset["sort_asc"] = (dataset["sort_asc"] == "1") ? 0: 1
                    sort_table(append, i, dataset["sort_asc"] == "1")
                } else if (target.tagName == "TD") {
                    var tr = target.parentElement.parentElement.querySelector("tr")
                    if (tr.childNodes[i].innerText.startsWith(field)) {
                        typeof cb == "function" && cb(event)
                    }
                }
            }
        }
    }
}

function init_table(event) {
    var append = document.querySelectorAll("article table").forEach(add_sort)
}

function set_layout() {
    article = document.querySelector("article")
    nav = document.querySelector("nav")
    if (window.innerWidth > 600) {
        article.style.maxWidth = (window.innerWidth - nav.offsetWidth-40)+"px"
        nav.className = "fixed"
    } else {
        article.style.maxWidth = ""
        nav.className = ""
    }
}

window.onresize = function (event) {
    set_layout()
}

window.onload = function() {
    init_menu()
    init_link()
    init_code()
    init_table()
    set_layout()
}
