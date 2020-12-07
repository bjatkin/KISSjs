# KISSjs a simple framework

# Terminology
   Tokenize - Extract and mark each token in the HTML, CSS or JS file
   Parse - Expand or convert a lower level representation into a higher level representation
   Convert - Generate a final, correct structure from the raw HTML, JS and CSS files
   Instance - Replace template fields with passed parameters
   Remote - A file that is not available for compilation or bunding locally
   Render - Take the datastructure and convert it into plain text HTML, JS, and CSS

# TODO
 - [ ] Fix the bug where id is being set to {id}
 - [ ] Remove the lazy.js poriton from the framework. Lazy components are not really nessisary right now
 - [ ] Add in a CSS Parser to make the CSS bundling more robust
 - [ ] Move the JS Parse to it's own module to clean up the naming
 - [ ] Add a default component so any unpassed parameter can still be set
 - [ ] Can I build materialized or other fancy components in this framework?