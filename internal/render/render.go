package render

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Aayush9029/nit/internal/api"
	"github.com/charmbracelet/lipgloss"
)

type TimelineEntry struct {
	Tweet    api.Tweet
	Username string
	IsNew    bool
}

var (
	greenDiamond  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("◆")
	yellowDiamond = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("◇")
	usernameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func FormatCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func RelativeTime(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 02")
	}
}

func FormatCost(userLookups, tweetReads int) string {
	cost := float64(userLookups)*0.01 + float64(tweetReads)*0.005
	if cost == 0 {
		return "free"
	}
	if cost < 0.01 {
		return "<$0.01"
	}
	return fmt.Sprintf("$%.2f", cost)
}

func FormatTimeline(entries []TimelineEntry, color bool) string {
	sort.Slice(entries, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, entries[i].Tweet.CreatedAt)
		tj, _ := time.Parse(time.RFC3339, entries[j].Tweet.CreatedAt)
		return ti.After(tj)
	})

	var b strings.Builder
	for _, e := range entries {
		indicator := greenDiamond
		if !e.IsNew {
			indicator = yellowDiamond
		}

		rel := RelativeTime(e.Tweet.CreatedAt)
		if color {
			b.WriteString(fmt.Sprintf("%s %s %s\n",
				indicator,
				usernameStyle.Render("@"+e.Username),
				dimStyle.Render("· "+rel),
			))
		} else {
			if e.IsNew {
				b.WriteString(fmt.Sprintf("◆ @%s · %s\n", e.Username, rel))
			} else {
				b.WriteString(fmt.Sprintf("◇ @%s · %s\n", e.Username, rel))
			}
		}

		text := e.Tweet.FullText()
		for _, line := range strings.Split(text, "\n") {
			if color && !e.IsNew {
				b.WriteString("  " + dimStyle.Render(line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}

		rt := FormatCount(e.Tweet.PublicMetrics.RetweetCount)
		lk := FormatCount(e.Tweet.PublicMetrics.LikeCount)
		stats := fmt.Sprintf("↻ %s  ♡ %s", rt, lk)
		if color {
			b.WriteString("  " + dimStyle.Render(stats) + "\n")
		} else {
			b.WriteString("  " + stats + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func FormatUserProfile(u api.User, color bool) string {
	var b strings.Builder
	if color {
		b.WriteString(fmt.Sprintf("\n%s · %s\n",
			usernameStyle.Render("@"+u.Username),
			lipgloss.NewStyle().Bold(true).Render(u.Name),
		))
		if u.Description != "" {
			b.WriteString("  " + dimStyle.Render(u.Description) + "\n")
		}
		b.WriteString(fmt.Sprintf("  %s\n",
			dimStyle.Render(fmt.Sprintf("%s followers · %s following · %s tweets",
				FormatCount(u.PublicMetrics.FollowersCount),
				FormatCount(u.PublicMetrics.FollowingCount),
				FormatCount(u.PublicMetrics.TweetCount),
			)),
		))
	} else {
		b.WriteString(fmt.Sprintf("\n@%s · %s\n", u.Username, u.Name))
		if u.Description != "" {
			b.WriteString("  " + u.Description + "\n")
		}
		b.WriteString(fmt.Sprintf("  %s followers · %s following · %s tweets\n",
			FormatCount(u.PublicMetrics.FollowersCount),
			FormatCount(u.PublicMetrics.FollowingCount),
			FormatCount(u.PublicMetrics.TweetCount),
		))
	}
	b.WriteString("──────────────────────────────────────────────────\n")
	return b.String()
}

func FormatTweets(tweets []api.Tweet, color bool) string {
	var b strings.Builder
	for _, t := range tweets {
		text := t.FullText()
		for _, line := range strings.Split(text, "\n") {
			b.WriteString("  " + line + "\n")
		}
		rel := RelativeTime(t.CreatedAt)
		rt := FormatCount(t.PublicMetrics.RetweetCount)
		lk := FormatCount(t.PublicMetrics.LikeCount)
		stats := fmt.Sprintf("%s · ↻ %s  ♡ %s", rel, rt, lk)
		if color {
			b.WriteString("  " + dimStyle.Render(stats) + "\n")
		} else {
			b.WriteString("  " + stats + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}
