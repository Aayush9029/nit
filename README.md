<p align="center">
  <img src="assets/icon.png" width="128" alt="nit">
  <h1 align="center">nit</h1>
  <p align="center">Personal Twitter feed reader for the terminal</p>
</p>

<p align="center">
  <a href="https://github.com/Aayush9029/nit/releases/latest"><img src="https://img.shields.io/github/v/release/Aayush9029/nit" alt="Release"></a>
  <a href="https://github.com/Aayush9029/nit/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Aayush9029/nit" alt="License"></a>
</p>

<p align="center">
  <img src="assets/demo.gif" alt="nit demo" width="800">
</p>

## Install

```bash
brew install aayush9029/tap/nit
```

Or tap first:

```bash
brew tap aayush9029/tap
brew install nit
```

Requires `TWITTER_BEARER_TOKEN` environment variable.

## Usage

```bash
nit elonmusk               # latest tweets from @elonmusk
nit add AnthropicAI         # add to your feed
nit add OpenAI              # add to your feed
nit fetch                   # get new tweets from your feed
nit list                    # show feed accounts
nit remove OpenAI           # remove from feed
nit reset                   # clear tracking, show all on next fetch
```

## Options

| Option | Description |
|--------|-------------|
| `--count, -c <n>` | Number of tweets (5-100, default: 5) |
| `--json, -j` | Output raw JSON |
| `--all, -a` | Show all recent tweets (ignore tracking) |

## License

MIT
