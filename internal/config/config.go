package config

import (
	"os"
	"path/filepath"
)

const (
	APIBase     = "https://api.x.com/2"
	DefaultCount = 5
	CacheMaxAge  = 31536000 // 1 year in seconds

	TweetFields = "created_at,public_metrics,text,note_tweet,referenced_tweets,author_id"
	UserFields  = "name,username,public_metrics,description"

	CostTweet = 0.005
	CostUser  = 0.010
)

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nit")
}

func EnsureDir() error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for _, name := range []string{"feed.txt", "last_seen.txt"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if err := os.WriteFile(p, nil, 0644); err != nil {
				return err
			}
		}
	}
	for _, name := range []string{"cache.json", "tweets.json"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if err := os.WriteFile(p, []byte("{}"), 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func BearerToken() string {
	return os.Getenv("TWITTER_BEARER_TOKEN")
}
