# KISSjs a simple framework

# TODO
 - [ ] Add a default component so any unpassed parameter can still be set
 - [ ] I'm adding an excessive number of classes to deeply nested elements, this will probably negatively impact performance
 - [ ] If a property has no value the value defaults to it's name e.g. class="{long}" and then < p long > makes < p class="long" >
 - [ ] Add in a CSS Parser to make the CSS bundling more robust
 - [ ] Move the JS Parse to it's own module to clean up the naming
 - [ ] It would be nice to make certian sections of HTML render differently depending on if a property is provided (e.g. if you provide and href have a link icon otherwise don't)
            Potential syntax is < if:my_var >...</> ... < else:my_var >...</>
            Potential syntax < my_var:href >...</> ... < __my_var:href >...</>
 - [ ] Support for CSS animations and the ULR function 
 - [ ] If you only pass in a single text node as a parameter default to filling all template values with that value
            e.g. (< chip > TEST </> is the same as < chip label="TEST"></> )
 - [x] Having some way to import a component and then instantly instance it would be nice
            Potential syntax < component src="only_once.html" atrib="test" ></>
 - [x] Fix the docstrings