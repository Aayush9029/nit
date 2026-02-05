import Foundation
import Testing
@testable import NitterKit

@Suite("Models Tests")
struct ModelsTests {

    // MARK: - Tweet Codable

    @Test("Tweet encodes and decodes to JSON")
    func tweetCodable() throws {
        let tweet = Tweet(
            id: "123", author: "Test", handle: "@test", content: "Hello world",
            timestamp: "Feb 5, 2026", relativeTime: "1h",
            replies: "10", retweets: "20", likes: "30", views: "100",
            isRetweet: false, isPinned: false, link: "https://x.com/test/status/123"
        )

        let encoder = JSONEncoder()
        let data = try encoder.encode(tweet)
        let decoded = try JSONDecoder().decode(Tweet.self, from: data)

        #expect(decoded.id == "123")
        #expect(decoded.author == "Test")
        #expect(decoded.content == "Hello world")
        #expect(decoded.isRetweet == false)
        #expect(decoded.isPinned == false)
    }

    // MARK: - Profile Codable

    @Test("Profile encodes and decodes to JSON")
    func profileCodable() throws {
        let profile = Profile(
            fullName: "Test User", handle: "@test", bio: "A bio",
            tweets: "100", following: "50", followers: "200", likes: "300"
        )

        let data = try JSONEncoder().encode(profile)
        let decoded = try JSONDecoder().decode(Profile.self, from: data)

        #expect(decoded.fullName == "Test User")
        #expect(decoded.handle == "@test")
        #expect(decoded.bio == "A bio")
        #expect(decoded.followers == "200")
    }

    // MARK: - TimelineResult Codable

    @Test("TimelineResult encodes and decodes with optional fields")
    func timelineResultCodable() throws {
        let result = TimelineResult(
            profile: nil,
            tweets: [],
            instance: "xcancel.com",
            cursor: nil
        )

        let data = try JSONEncoder().encode(result)
        let decoded = try JSONDecoder().decode(TimelineResult.self, from: data)

        #expect(decoded.profile == nil)
        #expect(decoded.tweets.isEmpty)
        #expect(decoded.instance == "xcancel.com")
        #expect(decoded.cursor == nil)
    }

    // MARK: - NitterError

    @Test("NitterError descriptions are meaningful")
    func errorDescriptions() {
        #expect(NitterError.allInstancesFailed.description.contains("All Nitter instances failed"))
        #expect(NitterError.noTweetsFound.description.contains("No tweets"))
        #expect(NitterError.userNotFound("bob").description.contains("bob"))
        #expect(NitterError.networkError("timeout").description.contains("timeout"))
    }
}
