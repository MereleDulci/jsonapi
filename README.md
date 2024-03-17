# JSON:API implementation for Go


## Why another JSON:API serialization library?

There are a few other JSONAPI libraries for Go. Some are rather good but not actively maintaned, other feel to pull
in way more than they should. The goal of this one is to strictly follow the JSON:API specification and to clutter
your codebase with as little implementation details as possible. 

## Target constraints

* Do not require implementation of any specific interface outside of the standard library
* Do not introduce vendor struct tags beyond what's absolutely necessary. Prioritize using `json` tag instead
* Full support for JSON:API 1.1
* Zero dependencies
* Minimize runtime overhead

## Explicitly out of scope

Anything related to the underlying API implementation is out of scope. This includes:

* Filtering
* Sorting
* Pagination
* Support for "schema-less" structures (though you can use map[string]interface{} for that still)

Leaving it entirely up to the user how to structure their API.

However, we can provide some helper methods to make it easier to pair the implementations up.
