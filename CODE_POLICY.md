# CODE POLICY

## Core Values

The following policies are oriented to promote the following:
1. Correctness
2. Consistency
3. Simplicity
4. Performance

Note that performance is dead last. We address performance issues only when
they arise, and never before they are actually a problem for someone.
Optimizing for performance generally means increasing complexity, which is a
trade off that must be weighed as part of design.

It is much better to have a simple design that is easy for anyone to debug,
than a performant design that is difficult to debug. It is better to bear the
keystroke cost of clear code upfront, than bear the cost of increased time
spent debugging later.

> DONT MAKE THINGS EASY TO DO, MAKE THEM EASY TO UNDERSTAND.
>
> -- [Bill Kenedy, Ardan Labs](https://twitter.com/goinggodotnet)

## Recommended Reading

- [SOLID Go Design by Dave Cheney](https://dave.cheney.net/2016/08/20/solid-go-design)
- [Data and Semantics by Bill Kenedy](https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html)
- [For Range Semantics by Bill Kenedy](https://www.ardanlabs.com/blog/2017/06/for-range-semantics.html)
- [Interface Values are Valueless by Bill Kenedy](https://www.ardanlabs.com/blog/2018/03/interface-values-are-valueless.html)
- [On Packaging by Bill Kenedy](https://www.ardanlabs.com/blog/2017/02/design-philosophy-on-packaging.html)
- [On Logging by Bill Kenedy](https://www.ardanlabs.com/blog/2017/05/design-philosophy-on-logging.html)
- [Ardan Labs Go Training repo](https://github.com/ardanlabs/gotraining)
- [Go and SQLite by David Crawshaw](https://crawshaw.io/blog/go-and-sqlite)

## Git

### Committing
Do the following before committing:
- Ensure the code builds using `make`.
- Always run `go mod tidy` before committing your code and include changes to
  `go.mod` and `go.sum` to ensure reproducible builds.

### Commit messages
Commit messages allow for developers to quickly and concisely review changes.
Bad commit messages force developers to look at the code and wonder why a
change was made.

Commit messages should follow these conventions:
1. [Separate subject from body with a blank line](https://chris.beams.io/posts/git-commit/#separate)
2. [Limit the subject line to 50 characters (as best as possible)](https://chris.beams.io/posts/git-commit/#limit-50)
3. [Capitalize the subject line](https://chris.beams.io/posts/git-commit/#capitalize)
4. [Do not end the subject line with a period](https://chris.beams.io/posts/git-commit/#end)
5. [Use the imperative mood in the subject line](https://chris.beams.io/posts/git-commit/#imperative)
6. [Wrap the body at 72 characters](https://chris.beams.io/posts/git-commit/#wrap-72)
7. [Use the body to explain what and why vs. how](https://chris.beams.io/posts/git-commit/#why-not-how)

Most of those are self explanatory but it is recommended that everyone at least
review the links for 5 and 7.

### Branches
We generally follow use a [Git
Flow](https://nvie.com/posts/a-successful-git-branching-model/) branching
model. The two major long running branches are:
- `develop` - where the action happens, you probably want to start from here
- `master` - latest official release, you'll only need this when tracking down
  an open bug on the release

These two branches shall never be rebased or `--force` pushed.

Other transient branches shall use the following naming conventions
- `feature/ABC` - long running features under development, generally "owned" by
  one developer and only pushed to share for code reviews, regularly rebased on
`develop` and later deleted
- `release/vX.X.X` - where the next release is prepared and reviewed, based on
  `develop`, deleted after merging into `master` and back into `develop`
- `hotfix/DEF` - where a bugfix for the current release is prepared and
  reviewed, based on `master`, deleted after merging back into `master` and
into `develop`

#### Rebase onto develop/master, don't merge
When working on your own fork, please always rebase onto whatever branch you
intend to make your PR against, and never merge from the base branch (`master`
or `develop`) to your feature.

## Golang

### Imports and modules

Importing external code should be done with care. Imports should be evaluated
on test coverage, recent development activity, documentation, and code quality.
All external code that is called should be read and understood.

Always run `go mod tidy` before committing your code.

### Packages

Packages should *provide* something useful and specific, not just contain
things. Packages named `common`, `utils`, and the like are prohibited.

### Variables

When intentionally declaring a variable with its zero value, use the `var`
declaration syntax.
```golang
var x int
```

Only use the short declaration syntax `:=` when declaring AND initializing a
variable to a non-zero value.
```golang
y := Type{Hello: "World"}
x, err := computeX()
```

Never do this: `x := Type{}`

### Errors
Errors must always be checked, even if it is extremely unlikely, or guaranteed
not to occur. Most errors should cause `fatd` to cleanly exit. Only a few
specific errors should

### Types

#### Interfaces
Interfaces define behavior, not data. Do not use interfaces to represent data.
Interfaces should describe what something *does*, not what it *is*.

As a general rule, you probably don't need an interface. Create the concrete
type first and *discover* the appropriate interfaces later when you refactor to
de-duplicate code that needs to *do* the same thing to more than one type.


#### Pointer/Value semantics

#### Factory functions
Factory functions construct and initialize a type and by convention start with
the word `New`.

Only create factory functions for types where the initialization/set-up is not
possible or obvious from outside the package.

Do not simply create factory functions out of convenience. It is preferred that
types that can be initialized by the user, are left to the user to initialize.

Factory functions must follow the data semantics of the type. See Pointer/Value
semantics above.


### Goroutines

You probably don't need to use a goroutine to solve the problem. Always write a
serialized solution first, evaluate performance, and only then can a concurrent
solution be considered.


### Logging

Only main, engine, and srv get to log. No other package may log. Other packages
must return errors up to the caller.


### Documentation


### Testing


