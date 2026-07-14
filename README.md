# grl

A terminal-based HTTP client

<p align="center">
  <img src=".github/assets/grl.gif" alt="grl demo" width="720">
</p>

## Installation

### Go

```sh
go install github.com/nxdir-s/grl/cmd/grl@latest
```

## Usage

Launch the TUI:

```sh
grl
```

Type a URL, press `ctrl+s`, and the response renders in the right panel

### Common keybindings

| Key                 | Action                     |
| ------------------- | -------------------------- |
| `ctrl+s`            | Send request               |
| `esc`               | Cycle focus between panels |
| `ctrl+w`            | Save request to collection |
| `ctrl+o`            | Open collection / history  |
| `ctrl+x` / `ctrl+q` | Cycle between tabs         |
| `ctrl+f`            | Search response            |
| `?`                 | Open help modal            |
| `ctrl+c`            | Quit                       |

## Configuration

gRL stores its state under `~/.config/grl/`:

```
~/.config/grl/
├── collections/     # saved requests (one JSON file per collection)
├── environments/    # environment variable sets
├── history.json     # request history (last 100 by default)
└── config.json      # active env, default method, timeout, history limit
```

## Authorization

Export the following environment variables to enable OAuth2 authorized and authenticated requests

```
export CLIENT_ID=clientid
export CLIENT_SECRET=secret
export OAUTH_URL=https://example.com/oauth2/token
export OAUTH_SCOPES=resource_read,resource_write // optional
```

## Development

```sh
# build
go build -o grl cmd/grl/main.go

# run
go run cmd/grl/main.go

# test
./.github/unit_tests.sh

# bench
go test -run='^$' -bench=. -benchmem ./...

# vet
go vet ./...
```

### Profiling

```sh
# write CPU and/or heap profiles for a session
GRL_CPUPROFILE=cpu.out GRL_MEMPROFILE=mem.out grl
go tool pprof cpu.out

# verbose storage logging in ~/.config/grl/out.log
GRL_LOG=debug grl
```

## Contributing

Issues and pull requests are welcome. For larger changes, please open an issue first to discuss the approach.

## License

Licensed under the [Apache License 2.0](LICENSE).
