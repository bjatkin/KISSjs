# KISSjs a simple framework

# Terminology
   Compile - the process of resolving all imports into a file, inlining all data, and then templating add data
   Parameters - values passed into a component
   Proccess - the proccess of taking parameters and templating them into a node or nodes
   Hydrate - the actuall proccess of replacing strings in a node. Should always be a child process of 'Process'
   Inline - the proccess of copying all the component nodes from child nodes into the parent nodes
   Outline - the proccess of removing nodes from the parent and breaking them back into a child
   Parse - the process of converting text into a datastructure (e.g. html file into html nodes)
   Resolve - the proccess of exploring/ retriving all the imports to a file

# TODO
 - [x] JS should be scoped when added to the bundle js to prevent leaky variables
 - [x] Is the new inline method working?
 - [x] How should we handle script bundling?(<script compile="true"...> + {KISSimport:'path', bundle: true})
 - [x] bundle css in the inline method
 - [x] bundle js in the inline method
 - [x] remove KISSimport statements from the JS code
 - [x] the global config file should be an html file and should only pass values to the main view. From their they will have to be manually passed to deeper components
 - [x] js script KISSimports should support both the 'compile' and 'bundle' keywords
   - [x] the compile keywords means the script will be searched for import statments
   - [x] the bundle keyword means the script will be added to the bundle js file
   - [x] maybe this should be inverted 'nocompile' and 'nobundle' so that they are on by default
 - [x] Inline component deffinitions should be allowed so you don't always have to link an external file
 - [x] support 'complie' and 'bundle' keywords for the KISS html components
   - [x] compile means the component will be templated and serached for deeper imports
   - [x] bundle means the component will be added to the main html file
   - [x] no_bundle should create a div with a src to the component file which may still be compiled
 - [ ] There is a bug where not all the JS scripts are being bundled correctly (missing observer.js)
 - [ ] Script tag src attribute is being re-written when listed as no-compile no-bundle (see jquery import)
 - [ ] Remove component outer tag from lazy components
 - [ ] Add a default component so any unpassed parameter can still be set.
 - [ ] Start working on the JS poriton of the framework
    - [ ] Observer.js for double binding the view
    - [ ] SPWA.js for requesting a new view as a tmp view rather than reloading the whole page
    - [ ] SPWA.js may be unessisary depending on how 'no_bundle' html components end up working.
    - [ ] What else would be important here?
 - [ ] Can I build materialized or other fancy components in this framework?

 # IDEA (probably super overkill but fun to think about)
 html script:
   add variables with:
      <var name="a" val=10/>
      <var name="b" val="test"/>
      <var name="c">
         ...
      </var>
   add control flow with:
      <if exp="<var:a> + <var:b/> == <var:c/>">
         ...
         <elif exp="<var:a/> + <var:b/> == <var:d/>">
            ...
            <else>
               ...
            </else>
         </elif>
      </if>
         or
      <for start="a = 0" end="a < 10" each="a++">
         ...
      </for>
         or
      <while exp="<var:a> < 0">
         ...
      </while>
   
   access properties on variables:
      <var name="withProp">
         <prop>
            myProp
         </prop>
      </var>
      ...
      <var name="prop" val="<var:withProp:prop>"> <- prop  now has the value "myProp"
      ...
      <var name="newWithProp">
         <prop1>
            myNewProp
         </prop1>
         <prop2>
            <var withProp:prop>
         </prop2>
      </var>
   
   composition:
      <var name="innner">
         <p> I'm an inner variable</p>
      </var>
      <var name="outer">
         <p> I'm the outer variable </p>
         <var inner>
      </var>
   
   scope:
      <div>
         <var name="topScope" val=100/>
         <div>
            <var name="scoped" val=100/>
            <p> I can use the variable here <var scoped> </p>
            <p> I can also use variables in parent scopes <var scoped> </p>
         </div>
         <p> But I cant use inner scope varables out here <var scoped> </p> <- this will throw a compilation error
      </div>
   
   Also, we should let text nodes have attributes:
      <textNode prop1="val1" prop2="val2">
         <p> this is a child of the text node </p>
         this is the text of the text node and get's stored as the xml tag data
         <div> <p> this is another child of the text node </p> </div>
      </textNode>

   Maybe indexing as well:
      <var name="index" val="A long string"/>
      <var:index[:10]/>
      
   All this added to the component system that allows for bundling and compiling

   Also, maybe allow inline components:
   <component tag="inlineComponent">
      <style>
         h1 { color: blue; }
      </style>
         <h1> Title Component </h1>
      <script>
         alert("This is a script");
      </script>
   </component>

   indexOf counter all in html
   <component tag="getIndex">
      <var name="source" var="{source}"/>
      <var name="target" var="{target}"/>
      <var name="i" var=0/>
      <while>
         <exp>
            <index src="<var:source>">
               <index>
                  [<var:i>:
                  <exp><var:i> + <var:target:len></exp>]
               </index>
            </index>
            !=
            <var:target/>
         </exp>
         <exp><var:index/> + 1</exp>
      </while>
      <length>
         <var:i>
      </length>
   </component>

   <var name="a" val="this is a test"/>
   <var name="len">
      <getIndex source="<var:a>" target="test"/>
   </var>

   alternate declaration syntax?
   <var:name val="test">
   this makes creating the variable indistiquiable from setting/ changing it's value
   <var:newVal>
      starting value
   </var:newVal>
   <var:newVal>
      changed value
   </var:newVal>

   you could maybe get rid of needing the <exp> node a bunch by parsing the expressions first
   <var:a> + <var:b>
   parses to 
   <add>
      <a>
         <var:a>
      </a>
      <b>
         <var:b>
      </b>
   </add>