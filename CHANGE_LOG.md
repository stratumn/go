## Change log

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
