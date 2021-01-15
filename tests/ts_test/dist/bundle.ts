{
 function utilFN(a: string, b: number): string {
    console.log(a+b)
    return a+"ret"
}

 function utilFN2(a: number): void {
    console.log("util fn 2 ", a)
}
}
{
const ret: string = utilFN("test", 10)
}
{

        utilFN2(10)

}
