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
