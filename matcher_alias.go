package httpmock

import "go.nhat.io/matcher/v2"

// Match returns a matcher according to its type.
var Match = matcher.Match

// JSON matches two json strings with <ignore-diff> support.
var JSON = matcher.JSON

// Exact matches two objects by their exact values.
var Exact = matcher.Exact

// Exactf matches two strings by the formatted expectation.
var Exactf = matcher.Exactf

// Regex matches two strings by using regex.
var Regex = matcher.Regex

// RegexPattern matches two strings by using regex.
var RegexPattern = matcher.RegexPattern

// Len matches by the length of the value.
var Len = matcher.Len

// IsEmpty checks whether the value is empty.
var IsEmpty = matcher.IsEmpty

// IsNotEmpty checks whether the value is not empty.
var IsNotEmpty = matcher.IsNotEmpty
