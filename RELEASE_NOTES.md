## Release notes

This is an early version of Stratumn Go packages and tools for experimental testing.

[Change log](CHANGE_LOG.md)

### Installing the CLI tools

#### Prerequisite

The only requirements are Docker and Docker Compose.

#### MacOS 64bit

Download `strat-darwin-amd64.zip`.

After downloading, unzip the archive, and move the `strat` binary somewhere in your executables `$PATH`.

```bash
$ unzip strat-darwin-amd64.zip
$ mv strat/strat /usr/local/bin/strat
```

#### Linux 64bit

Download `strat-linux-amd64.zip`.

After downloading, unzip the archive, and move the `strat` binary somewhere in your executables `$PATH`.

```bash
$ unzip strat-linux-amd64.zip
$ mv strat/strat /usr/local/bin/strat
```

#### Windows 64bit

Download `strat-windows-amd64.zip`.

After downloading, unzip the archive, and move `strat.exe` somewhere in your `%PATH%`.

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

To update the generators and the CLI, run: 

```bash
$ strat update
```
