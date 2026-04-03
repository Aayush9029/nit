package api

import (
	"fmt"
	"time"
)

type User struct {
	ID            string      `json:"id"`
	Username      string      `json:"username"`
	Name          string      `json:"name"`
	Description   string      `json:"description,omitempty"`
	PublicMetrics UserMetrics `json:"public_metrics"`
}

type UserMetrics struct {
	FollowersCount int `json:"followers_count"`
	FollowingCount int `json:"following_count"`
	TweetCount     int `json:"tweet_count"`
}

type Tweet struct {
	ID               string            `json:"id"`
	Text             string            `json:"text"`
	CreatedAt        string            `json:"created_at,omitempty"`
	AuthorID         string            `json:"author_id,omitempty"`
	PublicMetrics    TweetMetrics      `json:"public_metrics"`
	NoteTweet        *NoteTweet        `json:"note_tweet,omitempty"`
	ReferencedTweets []ReferencedTweet `json:"referenced_tweets,omitempty"`
}

func (t Tweet) FullText() string {
	if t.NoteTweet != nil && t.NoteTweet.Text != "" {
		return t.NoteTweet.Text
	}
	return t.Text
}

type TweetMetrics struct {
	RetweetCount int `json:"retweet_count"`
	LikeCount    int `json:"like_count"`
	ReplyCount   int `json:"reply_count"`
}

type NoteTweet struct {
	Text string `json:"text"`
}

type ReferencedTweet struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type UsersResponse struct {
	Data   []User     `json:"data,omitempty"`
	Errors []APIError `json:"errors,omitempty"`
}

type TweetsResponse struct {
	Data []Tweet    `json:"data,omitempty"`
	Meta TweetsMeta `json:"meta"`
}

type TweetsMeta struct {
	ResultCount int    `json:"result_count"`
	NewestID    string `json:"newest_id,omitempty"`
}

type APIError struct {
	Value  string `json:"value,omitempty"`
	Detail string `json:"detail,omitempty"`
	Title  string `json:"title,omitempty"`
}

type RateLimitError struct {
	ResetAt time.Time
}

func (e *RateLimitError) Error() string {
	wait := time.Until(e.ResetAt)
	if wait > 0 {
		return fmt.Sprintf("rate limited, resets in %s", wait.Round(time.Second))
	}
	return "rate limited, try again now"
}

type AuthError struct {
	StatusCode int
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed (HTTP %d), check TWITTER_BEARER_TOKEN", e.StatusCode)
}

type RequestError struct {
	Message string
}

func (e *RequestError) Error() string {
	return e.Message
}

type TimelineOpts struct {
	MaxResults int
	SinceID    string
	Exclude    string
}
