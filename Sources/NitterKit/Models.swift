import Foundation

public struct Tweet: Sendable, Codable {
    public let id: String
    public let author: String
    public let handle: String
    public let content: String
    public let timestamp: String
    public let relativeTime: String
    public let replies: String
    public let retweets: String
    public let likes: String
    public let views: String
    public let isRetweet: Bool
    public let isPinned: Bool
    public let link: String

    public init(
        id: String, author: String, handle: String, content: String,
        timestamp: String, relativeTime: String,
        replies: String, retweets: String, likes: String, views: String,
        isRetweet: Bool, isPinned: Bool, link: String
    ) {
        self.id = id
        self.author = author
        self.handle = handle
        self.content = content
        self.timestamp = timestamp
        self.relativeTime = relativeTime
        self.replies = replies
        self.retweets = retweets
        self.likes = likes
        self.views = views
        self.isRetweet = isRetweet
        self.isPinned = isPinned
        self.link = link
    }
}

public struct Profile: Sendable, Codable {
    public let fullName: String
    public let handle: String
    public let bio: String
    public let tweets: String
    public let following: String
    public let followers: String
    public let likes: String

    public init(
        fullName: String, handle: String, bio: String,
        tweets: String, following: String, followers: String, likes: String
    ) {
        self.fullName = fullName
        self.handle = handle
        self.bio = bio
        self.tweets = tweets
        self.following = following
        self.followers = followers
        self.likes = likes
    }
}

public struct TimelineResult: Sendable, Codable {
    public let profile: Profile?
    public let tweets: [Tweet]
    public let instance: String
    public let cursor: String?

    public init(profile: Profile?, tweets: [Tweet], instance: String, cursor: String?) {
        self.profile = profile
        self.tweets = tweets
        self.instance = instance
        self.cursor = cursor
    }
}

public enum NitterError: Error, CustomStringConvertible {
    case allInstancesFailed
    case noTweetsFound
    case userNotFound(String)
    case networkError(String)

    public var description: String {
        switch self {
        case .allInstancesFailed:
            "All Nitter instances failed. Try again later or use --instance with a self-hosted instance."
        case .noTweetsFound:
            "No tweets found."
        case .userNotFound(let user):
            "User '\(user)' not found."
        case .networkError(let msg):
            "Network error: \(msg)"
        }
    }
}
