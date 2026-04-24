# gRL

A terminal-based HTTP client

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
| `Tab` / `Shift+Tab` | Cycle focus between panels |
| `ctrl+s`            | Save request to collection |
| `ctrl+o`            | Open collection / history  |
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

## Development

```sh
# build
go build -o grl cmd/grl/main.go

# run
go run cmd/grl/main.go

# test
./.github/unit_tests.sh

# vet
go vet ./...
```

## Contributing

Issues and pull requests are welcome. For larger changes, please open an issue first to discuss the approach.

## License

Licensed under the [Apache License 2.0](LICENSE).
