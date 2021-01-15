{
    function utilFN(a, b) {
        console.log(a + b);
        return a + "ret";
    }
    function utilFN2(a) {
        console.log("util fn 2 ", a);
    }
}
{
    var ret = utilFN("test", 10);
}
{
    utilFN2(10);
}
