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

const maxWidth = 80 // max text width for readability

var (
	greenDiamond  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("◆")
	yellowDiamond = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("◇")
	usernameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cachedText    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// wrapText wraps a line to maxWidth, preserving a 2-space indent.
func wrapText(line string, width int) []string {
	if len(line) <= width {
		return []string{line}
	}
	var lines []string
	for len(line) > width {
		// Find last space before width
		cut := width
		for cut > 0 && line[cut] != ' ' {
			cut--
		}
		if cut == 0 {
			cut = width // no space found, hard break
		}
		lines = append(lines, line[:cut])
		line = line[cut:]
		if len(line) > 0 && line[0] == ' ' {
			line = line[1:]
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

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
			wrapped := wrapText(line, maxWidth-2) // -2 for indent
			for _, wl := range wrapped {
				if color && !e.IsNew {
					b.WriteString("  " + cachedText.Render(wl) + "\n")
				} else {
					b.WriteString("  " + wl + "\n")
				}
			}
		}

		rt := FormatCount(e.Tweet.PublicMetrics.RetweetCount)
		lk := FormatCount(e.Tweet.PublicMetrics.LikeCount)
		stats := fmt.Sprintf("↻ %s  ♥ %s", rt, lk)
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
			for _, wl := range wrapText(line, maxWidth-2) {
				b.WriteString("  " + wl + "\n")
			}
		}
		rel := RelativeTime(t.CreatedAt)
		rt := FormatCount(t.PublicMetrics.RetweetCount)
		lk := FormatCount(t.PublicMetrics.LikeCount)
		stats := fmt.Sprintf("%s · ↻ %s  ♥ %s", rel, rt, lk)
		if color {
			b.WriteString("  " + dimStyle.Render(stats) + "\n")
		} else {
			b.WriteString("  " + stats + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}
