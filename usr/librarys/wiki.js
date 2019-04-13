var wiki = {
    layout: {
        header: {
            height: 40,
        },
        nav: {
            min_width: 240,
            border_width: 2,
        },
        article: {
            padding: 20,
            max_width: 1000,
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
        var space = wiki.layout.article.width - wiki.layout.article.max_width
        article.style["margin-right"] = (space>0 ? space/2: 0) + "px"
    } else {
        nav.className = ""
        article.style.width = ""

        var space = wiki.layout.article.width - article.style.maxWidth
        if (space > 0) {
            article.style.marginRight = space / 2
        }
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
    nav.style.height = wiki.layout.nav.height-wiki.layout.nav.border_width+"px"
    nav.style.minWidth = wiki.layout.nav.min_width+"px"
    nav.style.marginTop = wiki.layout.header.height+"px"
    article.style.minHeight = wiki.layout.article.min_height+"px"
    article.style.marginTop = wiki.layout.header.height+"px"
    article.style.padding = wiki.layout.article.padding+"px"
    article.style.maxWidth = wiki.layout.article.max_width+"px"

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
    var fuck = kit.isMobile? 22: 16

    document.querySelectorAll("article pre").forEach(function(item, i) {
        var nu = insert_before(item, "div", {"className": "number1"})

        var line = (item.clientHeight-10)/fuck
        for (var j = 1; j <= line; j++) {
            append_child(nu, "div", {
                "style": {
                    "fontSize": kit.isMobile?"20px":"14px",
                    "lineHeight": kit.isMobile?"22px":"16px",
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
function adjust() {
    window.setTimeout(function(){
        window.scrollBy(0, -80)
    }, 100)
}

window.onresize = function (event) {
    init_layout()
}
window.onload = function(event) {
    init_menu()
    init_link()
    init_code()
    init_table()
    init_layout()
}

function modify_node(which, html) {
    var node = which
    if (typeof which == "string") {
        node = document.querySelector(which)
    }

    html && typeof html == "string" && (node.innerHTML = html)
    if (html && typeof html == "object") {
        for (var k in html) {
            if (typeof html[k] == "object") {
                for (var d in html[k]) {
                    node[k][d] = html[k][d]
                }
                continue
            }
            node[k] = html[k]
        }
    }
    return node
}
function create_node(element, html) {
    var node = document.createElement(element)
    return modify_node(node, html)
}
function append_child(parent, element, html) {
    var elm = create_node(element, html)
    parent.append(elm)
    return elm
}
function insert_before(self, element, html) {
    var elm = create_node(element, html)
    return self.parentElement.insertBefore(elm, self)
}

function format(str) {
    if (str.indexOf("http") == 0 && str.indexOf("<a href") == -1) {
        return "<a href='"+str+"' target='_blank'>"+str+"</a>"
    }
    return str
}
function sort_table(table, index, sort_asc) {
    var list = table.querySelectorAll("tr")
    var new_list = []

    var is_time = true
    var is_number = true
    for (var i = 1; i < list.length; i++) {
        var value = Date.parse(list[i].childNodes[index].innerText)
        if (!(value > 0)) {
            is_time = false
        }

        var value = parseInt(list[i].childNodes[index].innerText)
        if (!(value >= 0 || value <= 0)) {
            is_number = false
        }

        new_list.push(list[i])
    }

    var sort_order = ""
    if (is_time) {
        if (sort_asc) {
            method = function(a, b) {return Date.parse(a) > Date.parse(b)}
            sort_order = "time"
        } else {
            method = function(a, b) {return Date.parse(a) < Date.parse(b)}
            sort_order = "time_r"
        }
    } else if (is_number) {
        if (sort_asc) {
            method = function(a, b) {return parseInt(a) > parseInt(b)}
            sort_order = "int"
        } else {
            method = function(a, b) {return parseInt(a) < parseInt(b)}
            sort_order = "int_r"
        }
    } else {
        if (sort_asc) {
            method = function(a, b) {return a > b}
            sort_order = "str"
        } else {
            method = function(a, b) {return a < b}
            sort_order = "str_r"
        }
    }

    list = new_list
    new_list = []
    for (var i = 0; i < list.length; i++) {
        list[i].parentElement && list[i].parentElement.removeChild(list[i])
        for (var j = i+1; j < list.length; j++) {
            if (typeof method == "function" && method(list[i].childNodes[index].innerText, list[j].childNodes[index].innerText)) {
                var temp = list[i]
                list[i] = list[j]
                list[j] = temp
            }
        }
        new_list.push(list[i])
    }

    for (var i = 0; i < new_list.length; i++) {
        table.appendChild(new_list[i])
    }
    return sort_order
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
function scroll_page(event, page) {
    var body = document.querySelector("body")

    switch (event.key) {
        case "h":
            if (event.ctrlKey) {
                window.scrollBy(-page.scroll_x*10, 0)
            } else {
                window.scrollBy(-page.scroll_x, 0)
            }
            break
        case "H":
            window.scrollBy(-body.scrollWidth, 0)
            break
        case "l":
            if (event.ctrlKey) {
                window.scrollBy(page.scroll_x*10, 0)
            } else {
                window.scrollBy(page.scroll_x, 0)
            }
            break
        case "L":
            window.scrollBy(body.scrollWidth, 0)
            break
        case "j":
            if (event.ctrlKey) {
                window.scrollBy(0, page.scroll_y*10)
            } else {
                window.scrollBy(0, page.scroll_y)
            }
            break
        case "J":
            window.scrollBy(0, body.scrollHeight)
            break
        case "k":
            if (event.ctrlKey) {
                window.scrollBy(0, -page.scroll_y*10)
            } else {
                window.scrollBy(0, -page.scroll_y)
            }
            break
        case "K":
            window.scrollBy(0, -body.scrollHeight)
            break
    }
    return
}

