## Change log

### 0.0.8-dev
* Updated to Tendermint 0.9.0.
* Improved Tendermint node stability.
* Added batch support to stores.
* Added key-values support to stores.
* Cleaned error message stack trace.
* Created homebrew formula.

### 0.0.7-dev
* Made single binary for Indigo node and Tendermint core.
* Improved `tmstore` websocket.
* Changed license to MPL-2.0.

### 0.0.6-dev
* Fixed `tmstore` `didSave` channel.

### 0.0.5-dev
* Added support for proof-of-process over a Tendermint blockchain network. Each
  node can use any of the available store adapters to save its data.
* Added the `filetmpop` command, which is a Tendermint application that uses a
  filestore to save segments.
* Added the `tmstore` command, which starts a store that communicates with a
  Tendermint network to save segments.
* Added web sockets to stores.
* Added a `getenv` function to generators so they can get the value of
  environment variables.

### 0.0.4-dev
* The `strat` command will now use standard flags, using the `cobra` library.
* Added a `--generators-path` flag to `strat` to specify the location of the
  generators on the local file system. This makes it easier to develop
  generators.

### 0.0.3-dev
* Added a -stdin flag to all `strat` commands that execute scripts. It is set to
  false by default. If true, the standard input will be attached to the process
  executing the shell command. Hopefully this will fix `strat test` not working
  on Windows.

### 0.0.2-dev
* Reduced size of Docker images.
* Improved Docker images security.
* Locked Docker base images to specific versions.
* Added a `STRATUMN_CONFIG` environment variable to override the default
  Stratumn configuration path.
* Fixed `strat generate` trying to call `init` script if there is no project
  file.

### 0.0.1-dev
* Reset version to `0.0.1-dev`.
* Added the `strat deploy` command. It expects an environment name as the first
  argument, and will try to execute the `deploy:{env}` script of the project.
* Minor fixes and improvements.

### 0.24.1-alpha
* The command `strat update` should now work properly even if temporary files
  are on a different file system.

### 0.24.0-alpha
* The update mechanism for `strat update` should now work properly on all
  supported platforms.

### 0.23.0-alpha
* `strat update` will now check the cryptographic signature of downloaded
  binaries.

### 0.22.0-alpha
* All releases binaries are now signed.

### 0.21.0-alpha
* Added `strat` command `down` that executes a project script that should stop
  running services.
* Added `strat` command `pull` that executes a project script that should pull
  updates.
* Added `strat` command `push` that executes a project script that should push
  updates.

### 0.20.0-alpha
* Use `logrus` as the logger.

### 0.19.0-alpha
* Fixed `strat` `down:test` not working.
* Made `strat` support private Github repositories on repository commands via a
  flag or environment variable.

### 0.18.0-alpha
* Added support for `strat` `down:test` script. This addition makes `strat test`
  execute a `down:test` script if present after the tests. It will handle the
  exit code properly.
* Added ability for `strat` scripts to accept extra arguments. This change makes
  it possible to pass arguments (both flags and parameters) to a script when
  executing it.
* Improved code documentation.

### 0.17.0-alpha
* Scripts executed by `strat` can now be OS and architecture specific. It will
  look fist for a script named `{name}:{os}:{arch}`, then `{name}:{os}`, then
  `{name}:{arch}`, and finally `{name}`.
* Added the `strat build` command that executes the `build` script of a project.

### 0.16.0-alpha
* Scripts executed by `strat` are now passed to a shell instead of being
  executed directly. On*nix, the shell is `sh`. On Windows it is `cmd` (needs to
  be tested). This is to allow executing multiple commands in a script, for
  instance using `&&`.
* Scripts executed by `strat` are now passed a working directory and can
  optionally return a success status if the script doesn't exist.
* Project files `stratumn.json` can now specify an `init` script that will be
  executed after `strat generate` generates the project files.
* Changed the way the CLI updates itself in order to make it work on Windows
  (hopefully). The old binary will be renamed instead of attempting to override
  it. If an old binary is found during launch, it will attempt to remove it (it
  will fail if it doesn't have the right permissions, but it will not stop
  execution of the command).
* CLI and generators must now be updated separately using `strat update` and
  `strat update -generators` respectively. This is because the two operations
  might require different user permissions depending on the environment.
* Added a `-force` flag to `strat` to force an update. This is mostly designed
  to test the update mechanism easily.
* Added a function called `secret` to generator templates to create random
  strings.
* Fixed an issue when updating generators.
* Fixed an issue when downloading generators on Windows.
* Fixed some Windows compatibility issues.
* Fixed a problem when updating generators that would result in the generators
  not being saved to the correct directory.

### 0.15.0-alpha
* Command `strat update` will now update the CLI. It will update to the latest
  published release by default, or the latest prerelease if `-prerelease` is
  set.
* Command `strat update` will now update all known generators instead of just
  the default repository.
* Generator repositories now refer to Github references (usually a branch)
  instead of release tags. This approach is more flexible and makes maintaining
  generators easier.
* Commands `strat up` and `strat test` will now properly output errors.
* Commands `strat up` and `strat test` will now properly handle user input.
* Commands `strat up` and `strat test` will now shutdown cleanly and wait for
  the script to terminate.
* Added command `strat run script` that executes a project script.
