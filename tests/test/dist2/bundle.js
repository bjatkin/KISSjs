{
    MSGtwo = "This is another pate";
} {
    MSGone = "this is one page";
} {
    function SinglePage() {
        console.log(MSGone);
        console.log(MSGtwo)
    };
} {
    function observeIT() {
        console.log("I watched it and it was good")
    };
} {
    let mainMsgMan = "This is the main script";
    mainMsgMan += "!! :)";
    console.log(mainMsgMan);
} {
    function doKiss() {
        console.log("did the kiss");
        observeIT();
        SinglePage()
    };
    doKiss();
    externalFN();
    setTimeout(() => {
        let url = document.getElementById("test").getAttribute("src");
        html = fetch(url, {
            method: "GET",
        }).then(resp => resp.text()).then(resp => {
            console.log(resp);
            document.getElementById("test").innerHTML = resp
        })
    }, 2500);
} {
    console.log("THIS MEANS AN INLINE COMPONENT WAS ADDED");
} {
    let item = document.getElementById("kiss_list");
    console.log("Mouse Over Event for kiss_list Created");
    item.addEventListener("mouseover", () => {
        item.style.color = "red"
    });
} {
    let item = document.getElementById("opt4");
    console.log("Mouse Over Event for opt4 Created");
    item.addEventListener("mouseover", () => {
        item.style.color = "red"
    });
} {
    let item = document.getElementById("opt3");
    console.log("Mouse Over Event for opt3 Created");
    item.addEventListener("mouseover", () => {
        item.style.color = "red"
    });
} {
    let item = document.getElementById("opt2");
    console.log("Mouse Over Event for opt2 Created");
    item.addEventListener("mouseover", () => {
        item.style.color = "red"
    });
} {
    let item = document.getElementById("opt1");
    console.log("Mouse Over Event for opt1 Created");
    item.addEventListener("mouseover", () => {
        item.style.color = "red"
    });
} {
    let elm = document.getElementById("boo_star");
    elm.addEventListener('click', () => {
        alert("button with id: boo_star was pressed");
        console.log("the pressed label was")
    });
} {
    let elm = document.getElementById("meh_star");
    elm.addEventListener('click', () => {
        alert("button with id: meh_star was pressed");
        console.log("the pressed label was")
    });
} {
    let elm = document.getElementById("ya_star");
    elm.addEventListener('click', () => {
        alert("button with id: ya_star was pressed");
        console.log("the pressed label was")
    });
} {
    console.log("Hey this Biggest is cool");
}