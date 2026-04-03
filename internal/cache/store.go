package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Aayush9029/nit/internal/api"
	"github.com/Aayush9029/nit/internal/config"
)

type CachedUser struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Name        string  `json:"name"`
	Followers   int     `json:"followers"`
	Following   int     `json:"following"`
	Tweets      int     `json:"tweets"`
	Description string  `json:"description"`
	Timestamp   float64 `json:"ts"`
}

type Store struct {
	dir string
	mu  sync.RWMutex
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// --- feed.txt ---

func (s *Store) ReadFeed() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(filepath.Join(s.dir, "feed.txt"))
	if err != nil {
		return nil
	}
	var usernames []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			usernames = append(usernames, line)
		}
	}
	return usernames
}

func (s *Store) WriteFeed(usernames []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	content := strings.Join(usernames, "\n") + "\n"
	return atomicWrite(filepath.Join(s.dir, "feed.txt"), []byte(content))
}

func (s *Store) AppendFeed(username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(filepath.Join(s.dir, "feed.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(username + "\n")
	return err
}

func (s *Store) FeedContains(username string) bool {
	for _, u := range s.ReadFeed() {
		if strings.EqualFold(u, username) {
			return true
		}
	}
	return false
}

// --- cache.json (user cache) ---

func (s *Store) GetUser(username string) (*CachedUser, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cache := s.readUserCache()
	lower := strings.ToLower(username)
	for k, v := range cache {
		if strings.ToLower(k) == lower {
			if time.Now().Unix()-int64(v.Timestamp) < config.CacheMaxAge {
				return &v, true
			}
			return nil, false
		}
	}
	return nil, false
}

func (s *Store) PutUsers(users []api.User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cache := s.readUserCacheUnsafe()
	now := float64(time.Now().Unix())
	for _, u := range users {
		cache[u.Username] = CachedUser{
			ID:          u.ID,
			Username:    u.Username,
			Name:        u.Name,
			Followers:   u.PublicMetrics.FollowersCount,
			Following:   u.PublicMetrics.FollowingCount,
			Tweets:      u.PublicMetrics.TweetCount,
			Description: strings.ReplaceAll(u.Description, "\n", " "),
			Timestamp:   now,
		}
	}
	s.writeUserCacheUnsafe(cache)
}

func (s *Store) readUserCache() map[string]CachedUser {
	data, err := os.ReadFile(filepath.Join(s.dir, "cache.json"))
	if err != nil {
		return make(map[string]CachedUser)
	}
	var cache map[string]CachedUser
	if json.Unmarshal(data, &cache) != nil {
		return make(map[string]CachedUser)
	}
	return cache
}

func (s *Store) readUserCacheUnsafe() map[string]CachedUser {
	return s.readUserCache()
}

func (s *Store) writeUserCacheUnsafe(cache map[string]CachedUser) {
	data, _ := json.Marshal(cache)
	_ = atomicWrite(filepath.Join(s.dir, "cache.json"), data)
}

// --- last_seen.txt ---

func (s *Store) GetLastSeen(username string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(filepath.Join(s.dir, "last_seen.txt"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && parts[0] == username {
			return parts[1]
		}
	}
	return ""
}

func (s *Store) SetLastSeen(username, tweetID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := filepath.Join(s.dir, "last_seen.txt")
	data, _ := os.ReadFile(p)

	found := false
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && parts[0] == username {
			lines = append(lines, username+"="+tweetID)
			found = true
		} else {
			lines = append(lines, line)
		}
	}
	if !found {
		lines = append(lines, username+"="+tweetID)
	}
	_ = atomicWrite(p, []byte(strings.Join(lines, "\n")+"\n"))
}

// --- tweets.json (tweet cache) ---

func (s *Store) GetCachedTweets(username string) []api.Tweet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cache := s.readTweetCache()
	return cache[username]
}

func (s *Store) SaveTweets(username string, newTweets []api.Tweet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cache := s.readTweetCacheUnsafe()
	existing := cache[username]

	seen := make(map[string]bool)
	var merged []api.Tweet
	for _, t := range append(newTweets, existing...) {
		if !seen[t.ID] {
			seen[t.ID] = true
			merged = append(merged, t)
			if len(merged) >= 10 {
				break
			}
		}
	}
	cache[username] = merged
	s.writeTweetCacheUnsafe(cache)
}

func (s *Store) readTweetCache() map[string][]api.Tweet {
	data, err := os.ReadFile(filepath.Join(s.dir, "tweets.json"))
	if err != nil {
		return make(map[string][]api.Tweet)
	}
	var cache map[string][]api.Tweet
	if json.Unmarshal(data, &cache) != nil {
		return make(map[string][]api.Tweet)
	}
	return cache
}

func (s *Store) readTweetCacheUnsafe() map[string][]api.Tweet {
	return s.readTweetCache()
}

func (s *Store) writeTweetCacheUnsafe(cache map[string][]api.Tweet) {
	data, _ := json.Marshal(cache)
	_ = atomicWrite(filepath.Join(s.dir, "tweets.json"), data)
}

// --- RemoveUser ---

func (s *Store) RemoveUser(username string) {
	// Remove from feed
	feed := s.ReadFeed()
	var filtered []string
	for _, u := range feed {
		if !strings.EqualFold(u, username) {
			filtered = append(filtered, u)
		}
	}
	_ = s.WriteFeed(filtered)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from user cache
	uc := s.readUserCacheUnsafe()
	for k := range uc {
		if strings.EqualFold(k, username) {
			delete(uc, k)
		}
	}
	s.writeUserCacheUnsafe(uc)

	// Remove from last_seen
	p := filepath.Join(s.dir, "last_seen.txt")
	data, _ := os.ReadFile(p)
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], username) {
			continue
		}
		lines = append(lines, line)
	}
	_ = atomicWrite(p, []byte(strings.Join(lines, "\n")+"\n"))

	// Remove from tweet cache
	tc := s.readTweetCacheUnsafe()
	for k := range tc {
		if strings.EqualFold(k, username) {
			delete(tc, k)
		}
	}
	s.writeTweetCacheUnsafe(tc)
}

// --- Reset ---

func (s *Store) ResetTracking() {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = os.WriteFile(filepath.Join(s.dir, "last_seen.txt"), nil, 0644)
	_ = os.WriteFile(filepath.Join(s.dir, "tweets.json"), []byte("{}"), 0644)
}

func (s *Store) ResetAll() {
	s.ResetTracking()
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = os.WriteFile(filepath.Join(s.dir, "cache.json"), []byte("{}"), 0644)
}

// --- atomicWrite ---

func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
