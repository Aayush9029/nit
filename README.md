# nit

Browse tweets from the terminal via [Nitter](https://github.com/zedeus/nitter) instances.

## Install

```bash
brew install aayush9029/tap/nit
```

## Usage

```bash
# Show a user's timeline (default)
nit elonmusk
nit timeline jack

# Show profile info
nit profile elonmusk

# Search tweets
nit search "swift programming"

# Options
nit elonmusk --count 5                                 # limit tweets
nit elonmusk --json                                    # JSON output
nit elonmusk --instance https://my-nitter.example.com  # custom instance
```
