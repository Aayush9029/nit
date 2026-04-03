package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Aayush9029/nit/internal/api"
	"github.com/Aayush9029/nit/internal/cache"
	"github.com/Aayush9029/nit/internal/config"
	"github.com/Aayush9029/nit/internal/render"
	"github.com/Aayush9029/nit/internal/tui"
	"github.com/Aayush9029/nit/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9_]{1,15}$`)

func main() {
	if err := run(os.Args[1:]); err != nil {
		ui.Fatalf("%s", err)
	}
}

func run(args []string) error {
	if err := config.EnsureDir(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	token := config.BearerToken()
	if token == "" {
		if len(args) > 0 {
			switch args[0] {
			case "--version", "-v", "--help", "-h", "list", "ls":
				// These don't need a token
			default:
				return fmt.Errorf("TWITTER_BEARER_TOKEN not set")
			}
		}
	}

	if len(args) == 0 {
		showHelp()
		return nil
	}

	store := cache.NewStore(config.Dir())

	switch args[0] {
	case "fetch":
		return cmdFetch(args[1:], store, token)
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: nit add <username>")
		}
		return cmdAdd(args[1], store, token)
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("usage: nit remove <username>")
		}
		return cmdRemove(args[1], store)
	case "list", "ls":
		return cmdList(store)
	case "reset":
		return cmdReset(args[1:], store)
	case "--version", "-v":
		fmt.Printf("nit %s\n", version)
		return nil
	case "--help", "-h":
		showHelp()
		return nil
	default:
		return cmdUser(args, store, token)
	}
}

// --- cmdAdd ---

func cmdAdd(username string, store *cache.Store, token string) error {
	username = strings.TrimPrefix(username, "@")
	if !usernameRe.MatchString(username) {
		return fmt.Errorf("invalid username: %s", username)
	}
	if store.FeedContains(username) {
		ui.Status("@" + username + " is already in your feed")
		return nil
	}

	ui.Status("Looking up @" + username + "...")
	client := api.NewClient(token)
	resp, err := client.LookupUser(context.Background(), username)
	if err != nil {
		return err
	}
	if len(resp.Data) == 0 {
		return fmt.Errorf("user @%s not found", username)
	}

	store.PutUsers(resp.Data)
	canonical := resp.Data[0].Username
	if err := store.AppendFeed(canonical); err != nil {
		return err
	}
	ui.Success("Added @" + canonical + " to your feed")
	return nil
}

// --- cmdRemove ---

func cmdRemove(username string, store *cache.Store) error {
	username = strings.TrimPrefix(username, "@")
	if !store.FeedContains(username) {
		return fmt.Errorf("@%s is not in your feed", username)
	}
	store.RemoveUser(username)
	ui.Success("Removed @" + username + " from your feed")
	return nil
}

// --- cmdList ---

func cmdList(store *cache.Store) error {
	feed := store.ReadFeed()
	if len(feed) == 0 {
		ui.Dimf("Your feed is empty. Add accounts with: nit add <username>")
		return nil
	}
	ui.Header("Your feed")
	fmt.Println()
	for _, u := range feed {
		fmt.Printf("  %s@%s%s\n", ui.Cyan, u, ui.Reset)
	}
	fmt.Println()
	ui.Dimf("%d account(s)", len(feed))
	return nil
}

// --- cmdReset ---

func cmdReset(args []string, store *cache.Store) error {
	clearCache := false
	for _, a := range args {
		if a == "--cache" || a == "-c" {
			clearCache = true
		}
	}
	if clearCache {
		store.ResetAll()
		ui.Success("Reset all tracking, tweet cache, and user cache.")
	} else {
		store.ResetTracking()
		ui.Success("Reset tracking and tweet cache.")
		ui.Dimf("Use --cache to also clear user cache.")
	}
	return nil
}

// --- cmdUser ---

func cmdUser(args []string, store *cache.Store, token string) error {
	username := strings.TrimPrefix(args[0], "@")
	count := config.DefaultCount
	jsonOutput := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--count", "-c":
			if i+1 >= len(args) {
				return fmt.Errorf("--count requires a value")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				return fmt.Errorf("invalid count: %s", args[i])
			}
			count = clamp(n, 5, 100)
		case "--json", "-j":
			jsonOutput = true
		default:
			return fmt.Errorf("unknown option: %s", args[i])
		}
	}

	client := api.NewClient(token)
	ctx := context.Background()
	apiCalls := 0
	userLookups := 0

	// Resolve user (cache-first)
	var user api.User
	if cu, ok := store.GetUser(username); ok {
		user = api.User{
			ID:       cu.ID,
			Username: cu.Username,
			Name:     cu.Name,
			Description: cu.Description,
			PublicMetrics: api.UserMetrics{
				FollowersCount: cu.Followers,
				FollowingCount: cu.Following,
				TweetCount:     cu.Tweets,
			},
		}
	} else {
		if !jsonOutput {
			ui.Status("Looking up @" + username + "...")
		}
		resp, err := client.LookupUser(ctx, username)
		if err != nil {
			return err
		}
		if len(resp.Data) == 0 {
			return fmt.Errorf("user @%s not found", username)
		}
		user = resp.Data[0]
		store.PutUsers(resp.Data)
		apiCalls++
		userLookups++
	}

	// Fetch tweets
	tweetsResp, err := client.UserTimeline(ctx, user.ID, api.TimelineOpts{
		MaxResults: count,
		Exclude:    "retweets,replies",
	})
	if err != nil {
		return err
	}
	apiCalls++

	// Expand t.co links
	for i := range tweetsResp.Data {
		tweetsResp.Data[i].Text = api.ExpandTcoLinks(tweetsResp.Data[i].Text)
		if tweetsResp.Data[i].NoteTweet != nil {
			tweetsResp.Data[i].NoteTweet.Text = api.ExpandTcoLinks(tweetsResp.Data[i].NoteTweet.Text)
		}
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(tweetsResp)
	}

	// Display
	color := ui.IsTTY()
	fmt.Print(render.FormatUserProfile(user, color))
	if len(tweetsResp.Data) == 0 {
		ui.Dimf("  No tweets found.")
	} else {
		fmt.Print(render.FormatTweets(tweetsResp.Data, color))
	}

	cost := render.FormatCost(userLookups, len(tweetsResp.Data))
	ui.Dimf("%d API call(s) · ~%s", apiCalls, cost)
	return nil
}

// --- cmdFetch ---

type resolvedUser struct {
	ID        string
	Username  string
	Name      string
	Followers int
}

type fetchResult struct {
	user      resolvedUser
	newTweets []api.Tweet
	cached    []api.Tweet
	err       error
}

func cmdFetch(args []string, store *cache.Store, token string) error {
	useSince := true
	jsonOutput := false
	for _, a := range args {
		switch a {
		case "--all", "-a":
			useSince = false
		case "--json", "-j":
			jsonOutput = true
		default:
			return fmt.Errorf("unknown option: %s", a)
		}
	}

	feed := store.ReadFeed()
	if len(feed) == 0 {
		ui.Dimf("Your feed is empty. Add accounts with: nit add <username>")
		return nil
	}

	client := api.NewClient(token)
	ctx := context.Background()
	var apiCalls atomic.Int32

	// Resolve users (cache-first, batch API for misses)
	var users []resolvedUser
	var needLookup []string

	for _, username := range feed {
		if cu, ok := store.GetUser(username); ok {
			users = append(users, resolvedUser{
				ID: cu.ID, Username: cu.Username,
				Name: cu.Name, Followers: cu.Followers,
			})
		} else {
			needLookup = append(needLookup, username)
		}
	}

	userLookups := 0
	if len(needLookup) > 0 {
		resp, err := client.LookupUsers(ctx, needLookup)
		if err != nil {
			return err
		}
		apiCalls.Add(1)
		userLookups = len(needLookup)
		store.PutUsers(resp.Data)
		for _, u := range resp.Data {
			users = append(users, resolvedUser{
				ID: u.ID, Username: u.Username,
				Name: u.Name, Followers: u.PublicMetrics.FollowersCount,
			})
		}
		for _, e := range resp.Errors {
			ui.Error("@" + e.Value + ": " + e.Detail)
		}
	}

	// Concurrent timeline fetch
	doFetch := func() ([]render.TimelineEntry, int) {
		results := make(chan fetchResult, len(users))
		var wg sync.WaitGroup

		for _, u := range users {
			wg.Add(1)
			go func(u resolvedUser) {
				defer wg.Done()
				res := fetchResult{user: u}

				// Read cached tweets first
				res.cached = store.GetCachedTweets(u.Username)

				opts := api.TimelineOpts{
					MaxResults: config.DefaultCount,
					Exclude:    "retweets,replies",
				}
				if useSince {
					opts.SinceID = store.GetLastSeen(u.Username)
				}

				tweetsResp, err := client.UserTimeline(ctx, u.ID, opts)
				if err != nil {
					var rlErr *api.RateLimitError
					if errors.As(err, &rlErr) {
						ui.Error(rlErr.Error())
					}
					res.err = err
					results <- res
					return
				}
				apiCalls.Add(1)

				// Expand t.co links
				for i := range tweetsResp.Data {
					tweetsResp.Data[i].Text = api.ExpandTcoLinks(tweetsResp.Data[i].Text)
					if tweetsResp.Data[i].NoteTweet != nil {
						tweetsResp.Data[i].NoteTweet.Text = api.ExpandTcoLinks(tweetsResp.Data[i].NoteTweet.Text)
					}
				}

				res.newTweets = tweetsResp.Data

				// Save to cache
				store.SaveTweets(u.Username, tweetsResp.Data)

				// Update since_id
				if tweetsResp.Meta.NewestID != "" {
					store.SetLastSeen(u.Username, tweetsResp.Meta.NewestID)
				}

				results <- res
			}(u)
		}

		go func() { wg.Wait(); close(results) }()

		var entries []render.TimelineEntry
		totalNew := 0

		for res := range results {
			if res.err != nil {
				continue
			}

			// Tag new tweets
			newIDs := make(map[string]bool)
			for _, t := range res.newTweets {
				newIDs[t.ID] = true
				entries = append(entries, render.TimelineEntry{
					Tweet:    t,
					Username: res.user.Username,
					IsNew:    true,
				})
			}
			totalNew += len(res.newTweets)

			// Tag cached tweets (only when using since_id)
			if useSince {
				for _, t := range res.cached {
					if !newIDs[t.ID] {
						entries = append(entries, render.TimelineEntry{
							Tweet:    t,
							Username: res.user.Username,
							IsNew:    false,
						})
					}
				}
			}
		}

		return entries, totalNew
	}

	// JSON output
	if jsonOutput {
		entries, _ := doFetch()
		// Collect just the tweets
		var tweets []api.Tweet
		for _, e := range entries {
			tweets = append(tweets, e.Tweet)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(tweets)
	}

	// TUI mode
	if ui.IsTTY() {
		fetchCmd := func() tea.Msg {
			entries, totalNew := doFetch()
			return tui.FetchDoneMsg{
				Entries:  entries,
				NewCount: totalNew,
				APICalls: int(apiCalls.Load()),
				CostStr:  render.FormatCost(userLookups, totalNew),
			}
		}

		m := tui.NewModel(fetchCmd)
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err := p.Run()
		return err
	}

	// Non-interactive: print plain text
	ui.Header(fmt.Sprintf("Fetching feed (%d accounts)", len(users)))
	fmt.Println()

	entries, totalNew := doFetch()
	fmt.Print(render.FormatTimeline(entries, false))

	calls := int(apiCalls.Load())
	cost := render.FormatCost(userLookups, totalNew)
	if totalNew > 0 {
		ui.Success(fmt.Sprintf("%d new tweet(s)", totalNew))
	} else {
		ui.Dimf("No new tweets")
	}
	ui.Dimf("%d API call(s) · ~%s", calls, cost)
	return nil
}

func clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func showHelp() {
	fmt.Printf(`%snit%s v%s — Personal Twitter feed reader

%sUsage:%s
  nit <username>           Fetch recent tweets from a user
  nit fetch                Fetch new tweets from your feed
  nit add <username>       Add a user to your feed
  nit remove <username>    Remove a user from your feed
  nit list                 Show your feed accounts
  nit reset                Clear tracking (show all on next fetch)

%sOptions:%s
  --count, -c <n>          Number of tweets (5-100, default: 5)
  --json, -j               Output raw JSON
  --all, -a                Fetch all recent (ignore tracking)
  --version, -v            Show version
  --help, -h               Show this help

%sExamples:%s
  nit elonmusk             Latest tweets from @elonmusk
  nit add AnthropicAI      Add @AnthropicAI to your feed
  nit fetch                Get new tweets from all feed accounts

%sRequires TWITTER_BEARER_TOKEN environment variable.%s
`, ui.Bold, ui.Reset, version,
		ui.Bold, ui.Reset,
		ui.Bold, ui.Reset,
		ui.Bold, ui.Reset,
		ui.Dim, ui.Reset)
}
