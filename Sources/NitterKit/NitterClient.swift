import Foundation

public actor NitterClient {
    private let verbose: Bool
    private static let syndicationBase = "https://syndication.twitter.com/srv/timeline-profile/screen-name"

    public init(verbose: Bool = true) {
        self.verbose = verbose
    }

    // MARK: - Public API

    public func fetchTimeline(username: String, count: Int? = nil) async throws -> TimelineResult {
        let timeline = try await fetchSyndication(username: username)

        var tweets = timeline.tweets
        if let count {
            tweets = Array(tweets.prefix(count))
        }

        return TimelineResult(
            profile: timeline.profile,
            tweets: tweets,
            instance: "x.com",
            cursor: nil
        )
    }

    public func fetchProfile(username: String) async throws -> (Profile, String) {
        let timeline = try await fetchSyndication(username: username)
        guard let profile = timeline.profile else {
            throw NitterError.userNotFound(username)
        }
        return (profile, "x.com")
    }

    // MARK: - Syndication Fetch

    private func fetchSyndication(username: String) async throws -> TimelineResult {
        let urlString = "\(Self.syndicationBase)/\(username)"

        if verbose {
            printDim("  Fetching from x.com syndication...")
        }

        let html = try curlFetch(urlString)
        return try parseSyndicationHTML(html)
    }

    private func curlFetch(_ urlString: String) throws -> String {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/usr/bin/curl")
        process.arguments = ["-s", "--max-time", "15", "-w", "\n%{http_code}", urlString]

        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = FileHandle.nullDevice

        try process.run()

        let data = pipe.fileHandleForReading.readDataToEndOfFile()
        process.waitUntilExit()

        guard process.terminationStatus == 0 else {
            throw NitterError.networkError("Connection failed")
        }

        guard var output = String(data: data, encoding: .utf8), !output.isEmpty else {
            throw NitterError.networkError("Empty response")
        }

        // Extract HTTP status from last line (added by -w)
        let lines = output.split(separator: "\n", omittingEmptySubsequences: false)
        guard let statusStr = lines.last, let status = Int(statusStr) else {
            throw NitterError.networkError("Could not determine HTTP status")
        }

        guard status == 200 else {
            if status == 404 {
                throw NitterError.userNotFound("unknown")
            }
            if status == 429 {
                throw NitterError.networkError("Rate limited. Try again in a few minutes.")
            }
            throw NitterError.networkError("HTTP \(status)")
        }

        // Remove the status line from output
        output = lines.dropLast().joined(separator: "\n")
        return output
    }

    // MARK: - JSON Parsing

    private func parseSyndicationHTML(_ html: String) throws -> TimelineResult {
        // Extract __NEXT_DATA__ JSON from the HTML
        guard let startRange = html.range(of: "id=\"__NEXT_DATA__\" type=\"application/json\">"),
              let endRange = html.range(of: "</script>", range: startRange.upperBound..<html.endIndex) else {
            throw NitterError.noTweetsFound
        }

        let jsonString = String(html[startRange.upperBound..<endRange.lowerBound])
        guard let jsonData = jsonString.data(using: .utf8) else {
            throw NitterError.networkError("Could not parse response data")
        }

        let root = try JSONDecoder().decode(SyndicationRoot.self, from: jsonData)
        let entries = root.props.pageProps.timeline?.entries ?? []

        var profile: Profile?
        var tweets: [Tweet] = []

        for entry in entries {
            guard entry.type == "tweet", let tweetData = entry.content?.tweet else { continue }

            // Extract profile from first tweet's user data
            if profile == nil {
                let user = tweetData.user
                profile = Profile(
                    fullName: user.name,
                    handle: "@\(user.screen_name)",
                    bio: user.description ?? "",
                    tweets: formatCount(user.statuses_count),
                    following: formatCount(user.friends_count),
                    followers: formatCount(user.followers_count),
                    likes: formatCount(user.favourites_count)
                )
            }

            let isRetweet = tweetData.full_text?.hasPrefix("RT @") ?? tweetData.text.hasPrefix("RT @")
            let link = "https://x.com/\(tweetData.user.screen_name)/status/\(tweetData.id_str)"

            tweets.append(Tweet(
                id: tweetData.id_str,
                author: tweetData.user.name,
                handle: "@\(tweetData.user.screen_name)",
                content: tweetData.full_text ?? tweetData.text,
                timestamp: tweetData.created_at,
                relativeTime: relativeTime(from: tweetData.created_at),
                replies: formatCount(tweetData.reply_count ?? 0),
                retweets: formatCount(tweetData.retweet_count ?? 0),
                likes: formatCount(tweetData.favorite_count ?? 0),
                views: "",
                isRetweet: isRetweet,
                isPinned: false,
                link: link
            ))
        }

        return TimelineResult(profile: profile, tweets: tweets, instance: "x.com", cursor: nil)
    }

    // MARK: - Helpers

    private func formatCount(_ n: Int) -> String {
        if n >= 1_000_000 {
            let m = Double(n) / 1_000_000
            return String(format: "%.1fM", m)
        } else if n >= 1_000 {
            let k = Double(n) / 1_000
            return k >= 10 ? String(format: "%.0fK", k) : String(format: "%.1fK", k)
        }
        return "\(n)"
    }

    private func relativeTime(from dateString: String) -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "EEE MMM dd HH:mm:ss Z yyyy"
        formatter.locale = Locale(identifier: "en_US_POSIX")

        guard let date = formatter.date(from: dateString) else { return "" }
        let interval = Date().timeIntervalSince(date)

        if interval < 60 { return "now" }
        if interval < 3600 { return "\(Int(interval / 60))m" }
        if interval < 86400 { return "\(Int(interval / 3600))h" }
        if interval < 604800 { return "\(Int(interval / 86400))d" }

        let output = DateFormatter()
        output.dateFormat = "MMM d"
        return output.string(from: date)
    }

    private nonisolated func printDim(_ text: String) {
        let isTTY = isatty(STDERR_FILENO) != 0
        if isTTY {
            FileHandle.standardError.write(Data("\u{1B}[2m\(text)\u{1B}[0m\n".utf8))
        } else {
            FileHandle.standardError.write(Data("\(text)\n".utf8))
        }
    }
}

// MARK: - Syndication JSON Models

private struct SyndicationRoot: Decodable {
    let props: SyndicationProps
}

private struct SyndicationProps: Decodable {
    let pageProps: SyndicationPageProps
}

private struct SyndicationPageProps: Decodable {
    let timeline: SyndicationTimeline?
}

private struct SyndicationTimeline: Decodable {
    let entries: [SyndicationEntry]
}

private struct SyndicationEntry: Decodable {
    let type: String
    let content: SyndicationContent?
}

private struct SyndicationContent: Decodable {
    let tweet: SyndicationTweet?
}

private struct SyndicationTweet: Decodable {
    let id_str: String
    let text: String
    let full_text: String?
    let created_at: String
    let favorite_count: Int?
    let retweet_count: Int?
    let reply_count: Int?
    let user: SyndicationUser
}

private struct SyndicationUser: Decodable {
    let name: String
    let screen_name: String
    let description: String?
    let followers_count: Int
    let friends_count: Int
    let statuses_count: Int
    let favourites_count: Int
}
