# Package Organization

This documents the package organization of the fatd repo.

### fatd
Top level package includes the main function.

Imports:
- fatd/flag
- fatd/state

### fatd/flag
Defines all CLI args, environment variables and CLI completion that are used by
other parts of the program. Nearly all other packages include the flag package
to be able to access their settings.

### fatd/state
Defines the goroutine that compute the state of various FAT tokens.

Imports:
- fatd/flag
- fatd/db
- fatd/factom

External imports:
- ../fat
- ../fat/fat0
- ../fat/fat1

### fatd/db
All database related functions. Defines CRUD operations for types defined in
other packages like ../fat.

External imports:
- ../fat
- ../fat/fat0
- ../fat/fat1

### fatd/factom
Defines functions and types for querying the state of Factom through the
factomd api.
