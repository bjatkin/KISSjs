({KISSimport: "observer.js"})
({KISSimport: "sp.js"})
({KISSimport: "sp1.js"})
({KISSimport: "sp2.js"})
({KISSimport: "scripts/external.js", nocompile: true, nobundle: true})

function doKiss() {
    console.log("did the kiss");
    observeIT();
    SinglePage();
}

doKiss();
externalFN();