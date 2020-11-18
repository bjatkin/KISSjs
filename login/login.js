({ KISSimport: "../js/lazyLoad.js" })

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
    console.log(lazyLoad("signup_page", 700).then(() => {
        startFadeOut();
        hideComponent("login_page");
        showComponent("signup_page");
    }));
});

({KISSimport: "../js/observe.js"})

var user = observe({
    username: "",
    password: "",
}).
onChange(["username"], (o) => {
    console.log(o.username);
});