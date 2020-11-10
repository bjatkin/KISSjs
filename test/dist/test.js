{KISSimport: "observer.js"}
{KISSimport: "sp.js"}
{KISSimport: "sp1.js"}
{KISSimport: "sp2.js"}

function doKiss() {
    console.log("did the kiss")
    observeIT() 
    SinglePage()
}
{KISSimport: "sp1.js"}
{KISSimport: "sp2.js"}

function SinglePage() {
    console.log(MSGone)
    console.log(MSGtwo)
}
function observeIT() {
    console.log("I watched it and it was good")
}
{KISSimport: "sp2.js"}

MSGone = "this is one page"
MSGtwo = "This is another pate"

    item = document.getElementById("opt1");
    item.addEventListener("mouseover", () => {
        item.style.color = "red";
    });


    item = document.getElementById("opt2");
    item.addEventListener("mouseover", () => {
        item.style.color = "red";
    });


    item = document.getElementById("opt3");
    item.addEventListener("mouseover", () => {
        item.style.color = "red";
    });


    elm = document.getElementById("ya_star");
    elm.click(() => {
        alert("button with id: ya_star was pressed");
        console.log("the pressed label was $label$");
    });


    elm = document.getElementById("meh_star");
    elm.click(() => {
        alert("button with id: meh_star was pressed");
        console.log("the pressed label was $label$");
    });


    elm = document.getElementById("boo_star");
    elm.click(() => {
        alert("button with id: boo_star was pressed");
        console.log("the pressed label was boo starwars");
    });


    console.log("Hey this Biggest is cool");


    item = document.getElementById("$id$");
    item.addEventListener("mouseover", () => {
        item.style.color = "red";
    });

