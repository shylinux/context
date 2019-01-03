var wiki = {
    layout: {
        header: {
            height: 40,
        },
        nav: {
            min_width: 240,
        },
        article: {
            padding: 20,
        },
        footer: {
            height: 40,
        },
    },

    show_result: false,
    show_height: "30px",
    hide_height: "14px",
    scroll_x: 50,
    scroll_y: 50,
}

function set_layout() {
    var nav = document.querySelector("nav")
    var article = document.querySelector("article")

    if (window.innerWidth > 600) {
        nav.className = "fixed"
        wiki.layout.article.width = window.innerWidth - nav.offsetWidth- 2*wiki.layout.article.padding
        article.style.width = wiki.layout.article.width+"px"
    } else {
        nav.className = ""
        article.style.width = ""
    }
}

function action(event, cmd) {
    var target = event.target
    var dataset = target.dataset

    switch (cmd) {
        case "toggle_nav":
            var nav = document.querySelector("nav")
            nav.hidden = !nav.hidden
            set_layout(event)
            break
        case "toggle_list":
            var list = event.target.nextElementSibling
            list.hidden = !list.hidden
            break
        case "scroll":
            if (target.tagName == "BODY") {
                scroll_page(event, wiki)
            }
            break
    }
}
function init_layout() {
    var header = document.querySelector("header")
    var nav = document.querySelector("nav")
    var article = document.querySelector("article")
    var footer = document.querySelector("footer")

    wiki.layout.nav.height = window.innerHeight - wiki.layout.header.height - wiki.layout.footer.height
    wiki.layout.article.min_height = window.innerHeight - wiki.layout.header.height - wiki.layout.footer.height - 2*wiki.layout.article.padding

    header.style.height = wiki.layout.header.height+"px"
    footer.style.height = wiki.layout.footer.height+"px"
    nav.style.height = wiki.layout.nav.height+"px"
    nav.style.minWidth = wiki.layout.nav.min_width+"px"
    nav.style.marginTop = wiki.layout.header.height+"px"
    article.style.minHeight = wiki.layout.article.min_height+"px"
    article.style.marginTop = wiki.layout.header.height+"px"
    article.style.padding = wiki.layout.article.padding+"px"

    set_layout()
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
    var link = document.querySelector("nav .link");
    document.querySelectorAll("article a").forEach(function(item) {
        append_child(append_child(link, "li", {"innertText": item.innerText}), "a", {
            "href": item.href,
            "innerText": item.href,
        })
    })
}
function init_code() {
    var fuck = context.isMobile? 22: 16

    document.querySelectorAll("article pre").forEach(function(item, i) {
        var nu = insert_before(item, "div", {"className": "number1"})

        var line = (item.clientHeight-10)/fuck
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

        item.onclick = function(event) {
            window.getSelection().toString() && document.execCommand("copy")
        }
    })
}
function init_table(event) {
    var append = document.querySelectorAll("article table").forEach(add_sort)
}

window.onresize = function (event) {
    init_layout()
}
window.onload = function(event) {
    init_layout()
    init_menu()
    init_link()
    init_code()
    init_table()
    if (!context.isMobile) {
        var nav = document.querySelector("nav")
        nav.hidden = false
        set_layout()
    }
}
