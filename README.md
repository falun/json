# json

This project's aim is to build a drop-in replacement for [encoding/json][json-lib]
in the standard lib that adds simple validation to the JSON objects as they're
unmarshalled. This validation should be configured near the place that the JSON
structs are defined.

[json-lib]: https://golang.org/pkg/encoding/json/

## Why

Primarily I want something that is less heavyweight than a fully defined schema
and that is built into my data object's definition. An example use case would be
that I have an API which takes a JSON encoded payload and it absolutely must have
some string value set:

```go
type PostSerachFilter struct {
  // Username specifies the user whose content we're searching; an empty string
  // searches all users
  Username string `json:"username"`

  // TagMatches indicates a collection of post tags that results must match
  TagMatches []string `json:"tag_matches"`

  // PostedBefore is the latest post date any result should have
  PostedBefore *time.Time `json:"posted_before"`

  // PostedAfter is the first post date any result should have
  PostedAfter *time.Time `json:"posted_after"`
}
```

In this example `"username":""` / `"username":null` is totally valid input but
makes the query much more expensive and we want to ensure that the decision to
not constrain the query over a specific user account is deliberate. As such we
can't rely on `json.Unmarshal` to provide a default. Additionally, because,
null is a valid input we can't even switch to `Username *string`.

There are several ways to solve this but I've written them enough that I'd like
a general solution.

## What works
I put a proof of concept together to verify that I could do decode-time
validation and generate reasonable errors. This works on a single layer of
struct decoding:

```go
err := json.UnmarshalX(data, &dest, json.Options{Required: []string{"username"}})
```

Will decode a paylooad and produce an error if `username` was not explicitly set.
And if you decide that `null` isn't sufficient and that the empty string must be
passed you could change your unmarshal call to:

```go
err := json.UnmarshalX(data, &dest, json.Options{
  NullNotPresent: []string{"username"},
  Required:       []string{"username"},
})
```

Now when decoding a JSON struct will fail if the required keys are not sent
or, in the updated example, are sent as `null`.

## What needs love

As mentioned this currently only works for one layer of struct and requires
calling `UnmarshalX`. To make this properly useful I will be adding (or you
can if I'm too slow!):

1. processing `jsonx` struct tags that take `required`, `forbidden`, and
   `not-null` and handle the struct accordingly when `Unmarshal` is called;
2. updating my `Unmarshal` implementation to fully handle embedded structs,
   as well as maps, arrays, slices, pointers, and any other neat edge cases;
3. a more complete set of tests (or at least something beyond the absolute
   minimum).
   
I'd really like to avoid reimplementing `json.Unmarshal` completely but maybe
that's not possible _shrug_.

Pull requests appreciated :grinning:
