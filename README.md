# KISSjs a simple framework

# Terminology
   Tokenize - Extract and mark each token in the HTML, CSS or JS file
   Parse - Expand or convert a lower level representation into a higher level representation
   Convert - Generate a final, correct structure from the raw HTML, JS and CSS files
   Instance - Replace template fields with passed parameters
   Remote - A file that is not available for compilation or bunding locally
   Render - Take the datastructure and convert it into plain text HTML, JS, and CSS

# TODO
 - [ ] Add a default component so any unpassed parameter can still be set
 - [ ] If a property has no value the value defaults to it's name e.g. class="{long}" and then < p long > makes < p class="long" >
 - [ ] Add in a CSS Parser to make the CSS bundling more robust
 - [ ] Move the JS Parse to it's own module to clean up the naming
 - [ ] It would be nice to make certian sections of HTML render differently depending on if a property is provided (e.g. if you provide and href have a link icon otherwise don't)

# BUGS
 - [ ] Index out of range[0] with length 0 error when paramter node has no child nodes