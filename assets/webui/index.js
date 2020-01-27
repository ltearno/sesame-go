const el = document.getElementById.bind(document)

const elFromHtml = html => {
    const e = document.createElement('div')
    e.innerHTML = html
    return e.children.item(0)
}

const memoize = f => {
    const cache = new Map()
    return arg => {
        if (cache.has(arg))
            return cache.get(arg)
        let value = f(arg)
        cache.set(arg, value)
        return value
    }
}