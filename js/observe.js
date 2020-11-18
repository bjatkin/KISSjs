var observe = (obj) => {
    let pxy = new Proxy(obj, {
        set: function(obj, prop, value) {
            obj[prop] = value;
            pxy.fns.forEach((fn) => {
                if (fn.props.indexOf(prop) >= 0) {
                    fn.fu(obj)
                }
            });
            return true;
        }
    });
    pxy.fns = [];
    pxy.onChange = function(props, func) {
        pxy.fns.push({props: props, fu: func})
        return this;
    }

    return pxy;
}