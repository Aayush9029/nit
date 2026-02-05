import Foundation
import NitterKit

enum Formatter {
    // MARK: - ANSI Colors

    private static let isTTY = isatty(STDOUT_FILENO) != 0

    private static func cyan(_ text: String) -> String {
        isTTY ? "\u{1B}[36m\(text)\u{1B}[0m" : text
    }

    private static func dim(_ text: String) -> String {
        isTTY ? "\u{1B}[2m\(text)\u{1B}[0m" : text
    }

    private static func bold(_ text: String) -> String {
        isTTY ? "\u{1B}[1m\(text)\u{1B}[0m" : text
    }

    private static func gray(_ text: String) -> String {
        isTTY ? "\u{1B}[90m\(text)\u{1B}[0m" : text
    }

    private static let separator = String(repeating: "─", count: 48)

    // MARK: - Timeline Formatting

    static func formatTimeline(_ result: TimelineResult) -> String {
        var lines: [String] = []

        if let profile = result.profile, !profile.fullName.isEmpty {
            lines.append("")
            lines.append("  \(bold(profile.fullName)) \(gray(profile.handle))")
            lines.append("  \(dim("\(profile.tweets) tweets · \(profile.following) following · \(profile.followers) followers"))")
            lines.append("  \(dim(separator))")
        }

        if result.tweets.isEmpty {
            lines.append("")
            lines.append("  No tweets found.")
            return lines.joined(separator: "\n")
        }

        lines.append("")

        for (i, tweet) in result.tweets.enumerated() {
            let num = "\(i + 1)."
            let prefix = tweet.isPinned ? "📌 " : (tweet.isRetweet ? "🔁 " : "")

            // First line: number + content (truncated if needed)
            let contentLines = wrapText(tweet.content, width: 50)
            let firstLine = "  \(dim(num)) \(prefix)\(contentLines[0])"
            let timePad = "  \(dim(tweet.relativeTime))"
            lines.append("\(firstLine)\(timePad)")

            // Remaining content lines
            let indent = String(repeating: " ", count: num.count + 3)
            for line in contentLines.dropFirst() {
                lines.append("  \(indent)\(line)")
            }

            // Stats line
            let stats = "  \(indent)\(dim("💬 \(tweet.replies)  🔁 \(tweet.retweets)  ❤️ \(tweet.likes)"))"
            if !tweet.views.isEmpty {
                lines.append("\(stats)  \(dim("👁 \(tweet.views)"))")
            } else {
                lines.append(stats)
            }

            lines.append("")
        }

        lines.append("  \(dim(separator))")
        lines.append("  \(dim("via \(result.instance)"))")
        lines.append("")

        return lines.joined(separator: "\n")
    }

    // MARK: - Profile Formatting

    static func formatProfile(_ profile: Profile, instance: String) -> String {
        var lines: [String] = []

        lines.append("")
        lines.append("  \(bold(profile.fullName)) \(gray(profile.handle))")
        lines.append("  \(dim(separator))")

        if !profile.bio.isEmpty {
            lines.append("")
            let bioLines = wrapText(profile.bio, width: 50)
            for line in bioLines {
                lines.append("  \(line)")
            }
        }

        lines.append("")
        lines.append("  \(dim("Tweets:"))    \(profile.tweets)")
        lines.append("  \(dim("Following:")) \(profile.following)")
        lines.append("  \(dim("Followers:")) \(profile.followers)")
        lines.append("  \(dim("Likes:"))     \(profile.likes)")
        lines.append("")
        lines.append("  \(dim(separator))")
        lines.append("  \(dim("via \(instance)"))")
        lines.append("")

        return lines.joined(separator: "\n")
    }

    // MARK: - Search Formatting

    static func formatSearchResults(_ tweets: [Tweet], instance: String) -> String {
        let result = TimelineResult(profile: nil, tweets: tweets, instance: instance, cursor: nil)
        return formatTimeline(result)
    }

    // MARK: - Helpers

    private static func wrapText(_ text: String, width: Int) -> [String] {
        guard text.count > width else { return [text] }

        var lines: [String] = []
        var current = ""

        for word in text.split(separator: " ", omittingEmptySubsequences: true) {
            if current.isEmpty {
                current = String(word)
            } else if current.count + 1 + word.count <= width {
                current += " \(word)"
            } else {
                lines.append(current)
                current = String(word)
            }
        }
        if !current.isEmpty { lines.append(current) }
        return lines.isEmpty ? [text] : lines
    }
}
