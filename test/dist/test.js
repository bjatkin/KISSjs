{MSGtwo = "This is another pate"}
{

MSGone = "this is one page"}
{function observeIT() {
    console.log("I watched it and it was good")
}}
{


function SinglePage() {
    console.log(MSGone)
    console.log(MSGtwo)
}}
{
    console.log("Hey this Biggest is cool");
}
{
    let elm = document.getElementById("ya_star");
    elm.addEventListener('click', () => {
        alert("button with id: ya_star was pressed");
        console.log("the pressed label was $label$");
    });
}
{
    let elm = document.getElementById("meh_star");
    elm.addEventListener('click', () => {
        alert("button with id: meh_star was pressed");
        console.log("the pressed label was $label$");
    });
}
{
    let elm = document.getElementById("boo_star");
    elm.addEventListener('click', () => {
        alert("button with id: boo_star was pressed");
        console.log("the pressed label was boo starwars");
    });
}
{
    let item = document.getElementById("kiss_list");
    console.log("Mouse Over Event for kiss_list Created");
    item.addEventListener("mouseover", () => {
        item.style.color = "red";
    });
}
{
    let item = document.getElementById("opt1");
    console.log("Mouse Over Event for opt1 Created");
    item.addEventListener("mouseover", () => {
        item.style.color = "red";
    });
}
{
    let item = document.getElementById("opt2");
    console.log("Mouse Over Event for opt2 Created");
    item.addEventListener("mouseover", () => {
        item.style.color = "red";
    });
}
{
    let item = document.getElementById("opt3");
    console.log("Mouse Over Event for opt3 Created");
    item.addEventListener("mouseover", () => {
        item.style.color = "red";
    });
}
{
                console.log("THIS MEANS AN INLINE COMPONENT WAS ADDED")
            }
{
                console.log("THIS MEANS AN INLINE COMPONENT WAS ADDED")
            }
{





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
}, 2500);}
{
            let mainMsgMan = "This is the main script";
            mainMsgMan += "!! :)";

            console.log(mainMsgMan);
        }
