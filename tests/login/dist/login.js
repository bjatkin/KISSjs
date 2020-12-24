{let loaded = [];

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
}}
{
            
            hideComponent("forgot_page");
            hideComponent("signup_page");

            let fading = false;
            function startFadeIn() {
                if (fading) {
                    return;
                }
                document.getElementById("fg").style.display = "unset";
                document.getElementById("fg").style.opacity = 0;
                fading = true;
                window.requestAnimationFrame(fadeInFG)
            }

            function startFadeOut() {
                if (fading) {
                    return;
                }
                fading = true
                window.requestAnimationFrame(fadeOutFG)
            }

            function fadeInFG() {
                opacity = parseFloat(document.getElementById("fg").style.opacity);
                if (opacity < 1) {
                    document.getElementById("fg").style.opacity = opacity + 0.1;
                    window.requestAnimationFrame(fadeInFG)
                } else {
                    fading = false;
                }
            }

            function fadeOutFG() {
                opacity = parseFloat(document.getElementById("fg").style.opacity);
                if (opacity > 0) {
                    document.getElementById("fg").style.opacity = opacity - 0.1;
                    window.requestAnimationFrame(fadeOutFG)
                } else {
                    document.getElementById("fg").style.display = "none";
                    document.getElementById("fg").style.opacity = 0;
                    fading = false;
                }
            } 
 
            document.getElementById("forgot-link").addEventListener("click", () => {
                startFadeIn();
                lazyLoad("forgot_page", 700).
                then(() => {
                    startFadeOut();
                    hideComponent("login_page");
                    showComponent("forgot_page");
                })
            });
            document.getElementById("signup-link").addEventListener("click", () => {
                startFadeIn();
                console.log(lazyLoad("signup_page", 700).
                then(() => {
                    startFadeOut();
                    hideComponent("login_page");
                    showComponent("signup_page");
                }));
            });
        }
