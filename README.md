# nit

Browse tweets from the terminal via [Nitter](https://github.com/zedeus/nitter) instances.

## Install

### Homebrew

```bash
brew install Aayush9029/tap/nit
```

### Build from source

```bash
git clone https://github.com/Aayush9029/nit.git
cd nit
swift build -c release
cp .build/release/nit /usr/local/bin/
```

## Usage

```bash
# Show a user's timeline (default command)
nit elonmusk
nit timeline jack

# Show profile info
nit profile elonmusk

# Search tweets
nit search "swift programming"

# Options
nit elonmusk --count 5                              # limit tweets
nit elonmusk --json                                  # JSON output
nit elonmusk --instance https://my-nitter.example.com  # custom instance
```

## How it works

`nit` tries a chain of public Nitter instances in order. If one fails (HTTP error, JS challenge, timeout), it automatically falls back to the next. You can also point it at a self-hosted instance with `--instance`.

## Instances

1. xcancel.com
2. nitter.poast.org
3. nitter.privacyredirect.com
4. lightbrd.com
5. nitter.space
6. nitter.tiekoetter.com
7. nuku.trabun.org
8. nitter.catsarch.com
9. nitter.net

## License

MIT
