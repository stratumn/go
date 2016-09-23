## Change log

### 0.15.6-alpha
 * Changed the way the CLI updates itself in order to make it
   work on Windows (hopefully). The old binary will be renamed
   instead of attempting to override it. If an old binary is
   found during launch, it will attempt to remove it (it will
   fail if it doesn't have the right permissions, but it will
   not stop execution of the command).
 * CLI and generators must now be updated separately using
   `strat update` and `strat update -generators` respectively.
   This is because the two operations might require different user
   permissions depending on the environment.
 * Added a `-force` flag to `strat` to force an update. This is
   mostly designed to test the update mechanism easily.

### 0.15.5-alpha
 * Added a function called `secret` to generator templates to create
   random strings.

### 0.15.4-alpha
 * Fixed an issue when updating generators.

### 0.15.3-alpha
 * Fixed an issue when downloading generators on Windows.

### 0.15.2-alpha
 * Fixed some Windows compatibility issues.

### 0.15.1-alpha
 * Fixed a problem when updating generators that would result in the
   generators not being saved to the correct directory.

### 0.15.0-alpha
 * Command `strat update` will now update the CLI. It will update to the
   latest published release by default, or the latest prerelease if
   `-prerelease` is set.
 * Command `strat update` will now update all known generators instead of just
   the default repository.
 * Generator repositories now refer to Github references (usually a branch)
   instead of release tags. This approach is more flexible and makes
   maintaining generators easier.
 * Commands `strat up` and `strat test` will now properly output errors.
 * Commands `strat up` and `strat test` will now properly handle user input.
 * Commands `strat up` and `strat test` will now shutdown cleanly and wait for
   the script to terminate.
 * Added command `strat run script` that executes a project script.
