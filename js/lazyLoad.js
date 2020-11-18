let loaded = [];

var lazyLoad = (id, wait) => {
    for (let i = 0; i < loaded.length; i++) {
        if (loaded[i] == id) {
            return Promise.resolve(null);
        }
    }
    loaded.push(id)

    let elm = document.getElementById(id);
    let src = elm.getAttribute("src");
    let css = elm.getAttribute("css");
    let js = elm.getAttribute("js");

    return fetch(src, { method: "GET" }).
    then(resp => resp.text()).
    then(resp => {
        elm.innerHTML = resp;

        if (css != "") {
            let cssTag = document.createElement('link')
            cssTag.setAttribute("rel", "stylesheet")
            cssTag.setAttribute("href", css)
            document.head.appendChild(cssTag)
        }

        if (js != "") {
            let jsTag = document.createElement('script')
            jsTag.setAttribute("type","text/javascript")
            jsTag.setAttribute("src", js)
            document.head.appendChild(jsTag)
        }
    }).
    then(() => {
        return new Promise(resolve => setTimeout(resolve, wait));
    });
}

var hideComponent = (id) => {
    document.getElementById(id).style.display = "none";
}

var showComponent = (id) => {
    document.getElementById(id).style.display = "unset";
}