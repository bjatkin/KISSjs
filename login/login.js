var hideComponent = (id) => {
    document.getElementById(id).style.display = "none";
}

var showComponent = (id) => {
    document.getElementById(id).style.display = "unset";
}

document.getElementById("forgot-link").addEventListener("click", () => {
    hideComponent("login_page");
    showComponent("forgot_page");
});
document.getElementById("signup-link").addEventListener("click", () => {
    hideComponent("login_page");
    showComponent("signup_page");
});

({KISSimport: "../js/observe.js"});

var user = observe({
    username: "",
    password: "",
}).
onChange(["username"], (o) => {
    console.log(o.username);
});