<img src="assets/icon.png" width="128" alt="nit">

# nit

Personal Twitter feed reader for the terminal using X API v2.

## Installation

```bash
brew install aayush9029/tap/nit
```

Or tap first:

```bash
brew tap aayush9029/tap
brew install nit
```

## Usage

```bash
nit elonmusk               # Latest tweets from @elonmusk
nit add AnthropicAI         # Add to your feed
nit add OpenAI              # Add to your feed
nit fetch                   # Get new tweets from your feed
nit list                    # Show feed accounts
nit remove OpenAI           # Remove from feed
nit reset                   # Clear tracking, show all on next fetch
```

## Options

| Option | Description |
|--------|-------------|
| `--count, -c <n>` | Number of tweets (5-100, default: 5) |
| `--json, -j` | Output raw JSON |
| `--all, -a` | Show all recent tweets (ignore tracking) |
| `--version, -v` | Show version |
| `--help, -h` | Show help |

## How it works

1. Manages a personal feed of Twitter accounts in `~/.config/nit/feed.txt`
2. Uses X API v2 with bearer token auth to fetch tweets
3. Batch-resolves usernames to minimize API calls (1 call for up to 100 users)
4. Tracks last-seen tweet IDs to only show new content on `nit fetch`
5. Formats tweets with engagement stats and relative timestamps

## Requirements

- macOS
- `TWITTER_BEARER_TOKEN` environment variable

## License

MIT

---

*More CLI tools: [`brew tap aayush9029/tap`](https://github.com/Aayush9029/homebrew-tap)*
