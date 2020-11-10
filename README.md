# KISSjs a simple framework

# TODO
[] JS that does not change should be linked, not added to the js bundle (does not apply to component scripts)
[] JS that does not change should not be added multiple times (only applies to component scripts)
[] CSS that does not change after being hydrated should be scoped by the import class, not the component class
[] CSS that does not change after being hydrated should only be added once
[] How should middle ware be incorporated? (SASS compilers? TS compiler?)
[] JS should be scoped when added to the bundle js to prevent leaky variables
[] Start working on the JS poriton of the framework
    [] Observer.js for double binding the view
    [] SPWA.js for requesting a new view as a tmp view rather than reloading the whole page
    [] What else would be important here?
[] Can I build materialized components in this framework?

# BUGS
[] 'Compiled' script tags should be removed
[] CSS is not being hydrated correctly
[] CSS rules are not rendering into the bundle css file correctly
[] Components that are inside components are not being rendered or hydrated correctly
[] JS that is not added to the bundle should be linked in the head