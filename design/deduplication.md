# The problem

One of the most significant issues in google/jsonapi that got me thinking towards reinventing the wheel was that 
recursive data structures in relationships are not supported great. 

Consider the follwoing example

```go

type Post struct {
	ID string `jsonapi:"primary,posts"`
	Comments []*Comment
}


type Comment struct {
    ID string `jsonapi:"primary,comments"`
	Content string
    ReplyTo *Comment
}


post := Post{
	ID: "1",
    Comments: []*Comment{
		{
            ID: "1",
			Content: "first"
        },
        {
            ID: "2",
			Content: "second",
            ReplyTo: &Comment{
                ID: "1",
            },
        },
        {
            ID: "3",
			Content: "third"
            ReplyTo: &Comment{
                ID: "2",
            },
        }
    },
}


```

Here we have Comment ID 2 repeating twice. And this has a nasty side effect when we try to serialize this to JSONAPI.

"included" list is supposed to keep a list of unique values by their "type" and "id". It makes perfect sense as we could
potentially have hundreds of documents referencing the same resource. And we don't want to repeat it in the "included" section
another hundred times as it's ultimately the same data.

The problems start when it's the same resource type and id, but slightly different actual data. 

Comment 2 has two "views" - one with "Content" and another one without. So which one should go in? 

# Considerations

We know the resources are unique by their respective "type" and "id". But we also know that the "attributes" and 
"relationships" could have zero values if the details for them are not present. Also the value of each individual
field can be considered atomic. 

Following the example above, if we have a Comment with ID "2" and content "second" referenced directly from the Post
the same Comment ID "2" below must also have "content" equal to "second" and cannot be anything else. The same would 
apply to the other possible types.


# The solution

Shallow merge the content of the resources across all available records emitted for the "included" section. For the 
example above it's going to mean that the "included" seciton will have the following structure:

```json

{
  "included": [
    {"type":  "comments", "id": "1", "attributes": {"content": "first"}},
    {"type":  "comments", "id": "2", "attributes": {"content": "second"}},
    {"type":  "comments", "id": "3", "attributes": {"content": "third"}}
  ]
}

```

