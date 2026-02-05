import Foundation
import SwiftSoup

public struct HTMLParser: Sendable {

    public init() {}

    // MARK: - Timeline Parsing

    public func parseTweets(html: String, instanceName: String) throws -> [Tweet] {
        let doc = try SwiftSoup.parse(html)
        let items = try doc.select(".timeline-item")
        var tweets: [Tweet] = []

        for item in items.array() {
            // Skip "show more" items and thread connectors
            if (try? item.classNames().contains("show-more")) == true { continue }

            let author = (try? item.select(".fullname").first()?.text()) ?? ""
            let handle = (try? item.select(".username").first()?.text()) ?? ""
            let content = (try? item.select(".tweet-content.media-body").first()?.text()) ?? ""

            // Tweet link / ID
            let tweetLink = (try? item.select(".tweet-link").first()?.attr("href")) ?? ""
            let id = extractTweetID(from: tweetLink)

            // Timestamp
            let dateEl = try? item.select(".tweet-date a").first()
            let timestamp = (try? dateEl?.attr("title")) ?? ""
            let relativeTime = (try? dateEl?.text()) ?? ""

            // Stats: replies, retweets, likes, views (in order)
            // HTML structure: <span class="tweet-stat"><div class="icon-container"><span class="icon-comment"></span> 7,621</div></span>
            let stats = (try? item.select(".tweet-stat .icon-container")) ?? Elements()
            let statValues = stats.array().compactMap { el -> String? in
                guard let text = try? el.text() else { return nil }
                let trimmed = text.trimmingCharacters(in: .whitespaces)
                return trimmed.isEmpty ? "0" : trimmed
            }
            let replies = statValues.count > 0 ? statValues[0] : "0"
            let retweets = statValues.count > 1 ? statValues[1] : "0"
            let likes = statValues.count > 2 ? statValues[2] : "0"
            let views = statValues.count > 3 ? statValues[3] : ""

            // Flags
            let isRetweet = (try? item.select(".retweet-header").first()) != nil
            let isPinned = (try? item.select(".pinned").first()) != nil

            let link = tweetLink.isEmpty ? "" : "https://x.com\(tweetLink.replacingOccurrences(of: "#m", with: ""))"

            guard !content.isEmpty else { continue }

            tweets.append(Tweet(
                id: id, author: author, handle: handle, content: content,
                timestamp: timestamp, relativeTime: relativeTime,
                replies: replies, retweets: retweets, likes: likes, views: views,
                isRetweet: isRetweet, isPinned: isPinned, link: link
            ))
        }

        return tweets
    }

    // MARK: - Profile Parsing

    public func parseProfile(html: String) throws -> Profile {
        let doc = try SwiftSoup.parse(html)

        let fullName = (try? doc.select(".profile-card-fullname").first()?.text()) ?? ""
        let handle = (try? doc.select(".profile-card-username").first()?.text()) ?? ""
        let bio = (try? doc.select(".profile-bio").first()?.text()) ?? ""

        // Stats from profile stat list
        let statEls = try doc.select(".profile-statlist .profile-stat-num")
        let statValues = statEls.array().compactMap { try? $0.text() }

        let tweets = statValues.count > 0 ? statValues[0] : "0"
        let following = statValues.count > 1 ? statValues[1] : "0"
        let followers = statValues.count > 2 ? statValues[2] : "0"
        let likes = statValues.count > 3 ? statValues[3] : "0"

        return Profile(
            fullName: fullName, handle: handle, bio: bio,
            tweets: tweets, following: following, followers: followers, likes: likes
        )
    }

    // MARK: - Cursor Parsing

    public func parseCursor(html: String) -> String? {
        guard let doc = try? SwiftSoup.parse(html),
              let showMore = try? doc.select(".show-more a").last(),
              let href = try? showMore.attr("href") else {
            return nil
        }
        // Extract cursor param from href like ?cursor=xxx
        guard let url = URLComponents(string: href),
              let cursor = url.queryItems?.first(where: { $0.name == "cursor" })?.value else {
            return nil
        }
        return cursor
    }

    // MARK: - Challenge Detection

    public static func isChallengePage(_ html: String) -> Bool {
        let challenges = [
            "Verifying your browser",
            "Just a moment",
            "Making sure you're not a bot",
        ]
        return challenges.contains { html.contains($0) }
    }

    // MARK: - Helpers

    private func extractTweetID(from href: String) -> String {
        // href format: /user/status/1234567890#m
        let parts = href.split(separator: "/")
        guard let statusIdx = parts.firstIndex(of: "status"),
              statusIdx + 1 < parts.count else {
            return ""
        }
        return String(parts[statusIdx + 1]).replacingOccurrences(of: "#m", with: "")
    }
}
