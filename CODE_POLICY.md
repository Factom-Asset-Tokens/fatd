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
things. Common or util packages and the like are prohibited.


### Types

#### Pointer/Value semantics


### Goroutines

You probably don't need to use a goroutine to solve the problem. Always write a
serialized solution first, evaluate performance, and only then can a concurrent
solution be considered.


### Logging


### Documentation


### Testing


