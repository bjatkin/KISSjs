# KISSjs a simple framework

# TODO
 - [x] JS should be scoped when added to the bundle js to prevent leaky variables
 - [x] Is the new inline method working?
 - [x] How should we handle script bundling?(<script compile="true"...> + {KISSimport:'path', bundle: true})
 - [x] bundle css in the inline method
 - [x] bundle js in the inline method
 - [ ] How should middle ware be incorporated? (SASS compilers? TS compiler?)
 - [ ] Start working on the JS poriton of the framework
    - [ ] Observer.js for double binding the view
    - [ ] SPWA.js for requesting a new view as a tmp view rather than reloading the whole page
    - [ ] SPWA.js may be unessisary depending on how 'no_bundle' html components end up working.
    - [ ] What else would be important here?
 - [ ] Can I build materialized components in this framework?
 - [ ] remove KISSimport statements from the JS code
 - [ ] js script KISSimports should support both the 'compile' and 'bundle' keywords
   - [ ] the compile keywords means the script will be searched for import statments
   - [ ] the bundle keyword means the script will be added to the bundle js file
   - [ ] maybe this should be inverted 'nocompile' and 'nobundle' so that they are on by default
 - [ ] support 'complie' and 'bundle' keywords for the KISS html components
   - [ ] compile means the component will be templated and serached for deeper imports
   - [ ] bundle means the component will be added to the main html file
   - [ ] no_bundle should create a div with a src to the component file which may still be compiled
 - [ ] currently 'complie' indicates both that the element with be instatiated and that it will be searched for imports. Should these two functions be split?