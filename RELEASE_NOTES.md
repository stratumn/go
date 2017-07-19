## Release notes

This is an early version of Stratumn's SDK to create Indigo applications and networks.

[Change log](CHANGE_LOG.md)

### Installing the CLI tools

#### Prerequisite

The only requirements are Docker and Docker Compose.

#### MacOS 64bit

Download `strat-darwin-amd64.zip`.

After downloading, unzip the archive, and move the `strat` binary somewhere in
your executables `$PATH`.

```bash
$ unzip strat-darwin-amd64.zip
$ mv strat/strat /usr/local/bin/strat
```

#### Linux 64bit

Download `strat-linux-amd64.zip`.

After downloading, unzip the archive, and move the `strat` binary somewhere in
your executables `$PATH`.

```bash
$ unzip strat-linux-amd64.zip
$ mv strat/strat /usr/local/bin/strat
```

#### Windows 64bit

Download `strat-windows-amd64.zip`.

After downloading, unzip the archive, and move `strat.exe` somewhere in your
`%PATH%`.

### CLI quickstart

To generate a project in a directory named `demo`, run:

```bash
$ strat generate demo
$ cd demo
```

To launch all services locally, run within the project folder: 

```bash
$ strat up
```

To launch tests, run within the project folder: 

```bash
$ strat test
```

To update the SDK to the latest stable release, run: 

```bash
$ strat update
```

To update the SDK to the latest prerelease, run: 

```bash
$ strat update --prerelease
```

### Configuration

You can generate a configuration directory using:

```bash
$ strat generate /path/to/home/.stratumn
```

Choose `A Stratumn configuration` as the option.

This will set several configuration options, as well as default values for the
generators such as your name and company.

By default, the Stratumn tools look for configuration options in the `.stratumn`
directory within your home directory. You can change this directory by setting
a `STRATUMN_CONFIG` environment variable containing the path to the
configuration directory. This also allows you to manage multiple configurations.

### PGP signatures

Each zip file contains a cryptographic signature of the binary.

The signature can be verified using this public key:

https://keybase.io/stephan83/key.asc

This key is linked to the CTO of Stratumn on Keybase:

https://keybase.io/stephan83

With keybase (replace `BINARY` with the name of the binary):

```
$ keybase follow stephan83
$ keybase pgp verify -d BINARY.sig -i BINARY
```

With OpenPGP (replace `BINARY` with the name of the binary):

```
$ wget https://keybase.io/stephan83/key.asc
$ gpg2 --import key.asc
$ gpg2 --edit-key "Stephan Florquin <stephan@stratumn.com>"
        > trust
        > 5
        > y
        > quit
$ gpg2 --verify BINARY.sig BINARY
```
