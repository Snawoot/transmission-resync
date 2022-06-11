# transmission-resync

Updates torrents in transmission, pulling them from upstream tracker site. It invokes external scripts defined in configuration to ask if any of them have updated torrent URL. If such URL found, torrent will be replaced in transmission.

External scripts get fed via stdin with JSON describing torrent. These scripts are expected to return new torrent URL via stdout or empty string (nothing). In both cases successful exit code is mandatory, otherwise transmission-resync will stop processing of chain of scripts.

May be a great automation addition to [transmission-monitor](https://github.com/Snawoot/transmission-monitor).

Remote RPC must be enabled in Transmission for this program to work.

## Installation

#### Binaries

Pre-built binaries are available [here](https://github.com/Snawoot/transmission-resync/releases/latest).

#### Build from source

Alternatively, you may install transmission-resync from source. Run the following within the source directory:

```
make install
```

## Configuration

Configuration example:

#### /home/user/.config/transmission-resync.yaml

```yaml
rpc:
  user: transmissionuser
  password: transmissionpassword
chain:
  - command:
    - /home/user/.config/rutracker-resync.sh
```

Please consult [source](cmd/transmission-resync/defaults.go) for all available configuration options.

#### /home/user/.config/rutracker-resync.sh

To be done.

## Synopsis

```
$ ./bin/transmission-resync -h
Usage of transmission-resync:
  -conf string
    	path to configuration file (default "/home/user/.config/transmission-resync.yaml")
  -hash string
    	target torrent hash
  -version
    	show program version and exit
```
