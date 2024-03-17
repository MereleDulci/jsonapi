Prior to writing yet another library, let's reconcile the players we've got at hand and see if the problems they have
can be solved without introducing some other compromises.

## What's wrong with existing libraries?

https://github.com/manyminds/api2go

Very close to what I'd like in principle but requires a ton of interfaces to be implemented which can be easily
replaced with struct tags. It also weirdly keeps referenced structures embedded into `attributes` even when it's
explicitly marked as a relationship and can be safely cleaned out.

Kudos for using plain []byte and input/output instead of hooking

https://github.com/google/jsonapi

Not actively maintained. Perhaps the best implementation so far from usability perspective. But regularly hallucinates
in the output and doesn't support nested structures at all.

It's also guilty of "overwriting" the final "included" results with the last seen resource even if it's empty and there's
a "full" alternative available.

https://github.com/derekdowling/go-json-spec-handler

Outright bakes JSON:API into your structs... Which obviously leads to a super lightweight implementation. `json.Unmarshal`
would just do equally well??

https://github.com/mfcochauxlaberge/jsonapi

Doesn't support nesting structs

https://github.com/pieoneers/jsonapi-go  /  https://github.com/pieoneers/jsonapi-client-go

Requires implementation of vendor interface on all involved structs. Manual "building" of the response through

## Key principles

The principles we should stick to and make sure they aren't violated:

* Open for extension by the users. You don't have to fork the library to extend the base functionality such as to introduce
support for a custom struct handling in the attributes. Keep is as modular as possible. Great if we can swap standalone 
parts in the end. But on the other hand it should not be an excuse to compromise on any of the other points.
* Rock-solid interface. There's only one version - v1. Any need for v2 should outright mean a new library
* Do not expand the amount of code anyone needs to write for the sake of expanding it. There's plenty of standard tools
that can just be a good fit for the job. We don't need to repeat functionality of `json` struct tag.

## Problem definition

There're two main public methods - `Marshal` and `Unmarshal`. In it's simplest form Marshal takes in a struct or a 
slice of structs and yields a `[]byte` representation of the JSON:API document. Unmarshal takes in a `[]byte` and a 
pointer to a struct and fills it in with the data from the JSON:API document.

For convenience we can provide alternate high level methods that would operate on `io.Reader` and `io.Writer` instead.

Additionally it should be able to operate on slices of structs for both `Marshal` and `Unmarshal` methods alternative. 
Both options are equally valid in JSON:API.

## Design constraints

First of all, move as much metadata as possible into struct tags. This immediately includes:

* resource name
* attribute and relationships field names


### Marshal

It should yield a top-level JSON:API document. Under the hood it should be able to recognise the type of the struct
it's dealing with, fill in "data" and "included" parts of the output correctly. 

It should not include document in the "included" part of the output if there's no valid relationship on the provided
struct (e.g. it's nil). One common problem on the other libraries is that they push phantom structs into "included".

A phantom struct should be defined as a struct with no attributes and relationships. However it should be possible for thee
user to make such structs legit for "included" case by case. For instance if the only field a struct has is the ID. Default
values for primitive types could cause some headache here as it's impossible to tell the difference between a zero value
and a missing value case.

Internally `Marshall` should inspect the struct provided to it and identify three key parts:

* What's the `type` of the provided resource is
* What's the `id` of the provided resource is
* What are the `attributes` of the provided resource
* What are the `relationships` of the provided resource

With that settled, it should make sure there can be zero overlap between `attributes` and `relationships` fields.

For relationships it should be able to "emit" the related resource into the "included" part of the output.


Applied to a slice of structs it should iteratively apply the same logic to each element.


### Unmarshal

It should receive the top level document and fill in the provided struct with the data from the "data" part of the document.
In order to support parsing a collection of documents we're going to need a separate method treating collections. 

The struct provided to `Unrmarshal` should be validated for a "type" match with the provided json record. "attributes" is 
getting populated to the struct fields, and "relationships" are created as instances of the related structs. 


## Struct tags

Strictly speaking the libraries are mutually exclusive. One wouldn't use both google/jsonapi and this library in a single 
project. Taking that convention we can safely reuse the `jsonapi` tag for our purposes. However, it would just be nice to
follow the same conventions so that migration to or from `google/jsonapi` library is easy.

* `json:"<fieldname>"` - standard tag for the field name. If provided alone the field is treated as an attribute and 
the value from it is used as a public field name in the JSON:API document.
* `jsonapi:"<kind>,<fieldname>,<specs>"` - extended version of the `json` tag. Following conventions set up by `google/jsonapi`
Ideally a project using `google/jsonapi` would be fully compatible with this library. 
    * `kind` == "attr": indicates the field is an attribute
    * `kind` == "rel": indicates the field is a relationship
    * `kind` == "primary": indicates the field is the primary key of the resource. Field value goes into "id" and <fieldname>
  value becomes the resource type on JSON:API document

By default if struct field is missing `json` or `jsonapi` tags exported fields are included in their camelCase form.



## Benchmarking

As we're building on top of `encoding/json` primarily we should be worried of performance overhead that JSON:API encoding introduces

It would also be interesting to see how it compares to the other JSON:API libraries. So that it ends up either on par
or better than the rest. This one is likely going to be pulled into another repository to keep this one clean. 


