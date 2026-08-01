// Package domain holds the vocabulary shared by every domain package's Run
// contract.
package domain

// Argument is one positional command-line argument as delivered to a domain
// Run. It is a type alias, not a defined type, so Run signatures written with
// it still match go-app's Runner contract (variadic string) exactly while
// naming the domain concept.
//
// It is declared once, here, rather than per package. Both spellings compile
// and both satisfy the runner contract, so nothing but this file keeps them
// from drifting into the same concept under a different name in every package
// that uses it.
type Argument = string
