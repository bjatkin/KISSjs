({KISSimport: "observer.js"});
({KISSimport: "sp.js"});
({KISSimport: "sp1.js"});
({KISSimport: "sp2.js"});
({KISSimport: "scripts/external.js", remote: true});

function doKiss() {
    console.log("did the kiss");
    observeIT();
    SinglePage();
}

doKiss();
externalFN();

setTimeout(() => {
    let url = document.getElementById("test").getAttribute("src")
    html = fetch(url, {
        method: "GET",
    }).then(resp => resp.text()).
    then(resp => {
        console.log(resp);
        document.getElementById("test").innerHTML = resp;
    })
}, 2500);