package matcher

import "github.com/nhatthm/go-matcher"

// Matcher determines if the actual matches the expectation.
type Matcher = matcher.Matcher

// ExactMatcher matches by exact string.
type ExactMatcher = matcher.ExactMatcher

// Callback matches by calling a function.
type Callback = matcher.Callback

// Match returns a matcher according to its type.
var Match = matcher.Match

// JSON matches two json strings with <ignore-diff> support.
var JSON = matcher.JSON

// Regex matches two strings by using regex.
var Regex = matcher.Regex

// RegexPattern matches two strings by using regex.
var RegexPattern = matcher.RegexPattern

// Exact matches two objects by their exact values.
var Exact = matcher.Exact

// IsNotEmpty checks whether the value is not empty.
var IsNotEmpty = matcher.IsNotEmpty
