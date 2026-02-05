import Testing
@testable import NitterKit

@Suite("HTMLParser Tests")
struct HTMLParserTests {
    let parser = HTMLParser()

    // MARK: - Tweet Parsing

    static let timelineHTML = """
    <div class="timeline-item">
        <div class="tweet-body">
            <a class="tweet-link" href="/elonmusk/status/1234567890#m"></a>
            <div class="fullname">Elon Musk</div>
            <div class="username">@elonmusk</div>
            <div class="tweet-content media-body">Building an interstellar civilization</div>
            <span class="tweet-date"><a href="/elonmusk/status/1234567890#m" title="Feb 5, 2026 · 3:45 PM UTC">2h</a></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-comment"></span> 7,545</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-retweet"></span> 5,562</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-heart"></span> 38,790</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-views"></span> 50.2M</div></span>
        </div>
    </div>
    <div class="timeline-item">
        <div class="tweet-body">
            <a class="tweet-link" href="/elonmusk/status/1234567891#m"></a>
            <div class="fullname">Elon Musk</div>
            <div class="username">@elonmusk</div>
            <div class="tweet-content media-body">China now generates 33.2% of the world's electricity</div>
            <span class="tweet-date"><a href="/elonmusk/status/1234567891#m" title="Feb 5, 2026 · 1:00 PM UTC">5h</a></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-comment"></span> 3,100</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-retweet"></span> 3,690</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-heart"></span> 21,429</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-views"></span> 6.1M</div></span>
        </div>
    </div>
    """

    @Test("Parse multiple tweets from timeline HTML")
    func parseTweets() throws {
        let tweets = try parser.parseTweets(html: Self.timelineHTML, instanceName: "xcancel.com")
        #expect(tweets.count == 2)
    }

    @Test("Parse tweet author and handle")
    func parseTweetAuthor() throws {
        let tweets = try parser.parseTweets(html: Self.timelineHTML, instanceName: "xcancel.com")
        #expect(tweets[0].author == "Elon Musk")
        #expect(tweets[0].handle == "@elonmusk")
    }

    @Test("Parse tweet content")
    func parseTweetContent() throws {
        let tweets = try parser.parseTweets(html: Self.timelineHTML, instanceName: "xcancel.com")
        #expect(tweets[0].content == "Building an interstellar civilization")
        #expect(tweets[1].content == "China now generates 33.2% of the world's electricity")
    }

    @Test("Parse tweet ID from link")
    func parseTweetID() throws {
        let tweets = try parser.parseTweets(html: Self.timelineHTML, instanceName: "xcancel.com")
        #expect(tweets[0].id == "1234567890")
        #expect(tweets[1].id == "1234567891")
    }

    @Test("Parse tweet stats")
    func parseTweetStats() throws {
        let tweets = try parser.parseTweets(html: Self.timelineHTML, instanceName: "xcancel.com")
        #expect(tweets[0].replies == "7,545")
        #expect(tweets[0].retweets == "5,562")
        #expect(tweets[0].likes == "38,790")
        #expect(tweets[0].views == "50.2M")
    }

    @Test("Parse tweet timestamps")
    func parseTweetTimestamp() throws {
        let tweets = try parser.parseTweets(html: Self.timelineHTML, instanceName: "xcancel.com")
        #expect(tweets[0].relativeTime == "2h")
        #expect(tweets[0].timestamp.contains("Feb 5, 2026"))
    }

    @Test("Parse tweet link")
    func parseTweetLink() throws {
        let tweets = try parser.parseTweets(html: Self.timelineHTML, instanceName: "xcancel.com")
        #expect(tweets[0].link == "https://x.com/elonmusk/status/1234567890")
    }

    // MARK: - Retweet / Pinned

    @Test("Detect retweet")
    func detectRetweet() throws {
        let html = """
        <div class="timeline-item">
            <div class="retweet-header">Retweeted</div>
            <a class="tweet-link" href="/someone/status/999#m"></a>
            <div class="fullname">Someone</div>
            <div class="username">@someone</div>
            <div class="tweet-content media-body">A retweeted post</div>
            <span class="tweet-date"><a title="Jan 1, 2026">1d</a></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-comment"></span> 10</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-retweet"></span> 20</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-heart"></span> 30</div></span>
        </div>
        """
        let tweets = try parser.parseTweets(html: html, instanceName: "test")
        #expect(tweets[0].isRetweet == true)
    }

    @Test("Detect pinned tweet")
    func detectPinned() throws {
        let html = """
        <div class="timeline-item">
            <div class="pinned">Pinned</div>
            <a class="tweet-link" href="/user/status/888#m"></a>
            <div class="fullname">User</div>
            <div class="username">@user</div>
            <div class="tweet-content media-body">A pinned tweet</div>
            <span class="tweet-date"><a title="Dec 25, 2025">30d</a></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-comment"></span> 5</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-retweet"></span> 10</div></span>
            <span class="tweet-stat"><div class="icon-container"><span class="icon-heart"></span> 15</div></span>
        </div>
        """
        let tweets = try parser.parseTweets(html: html, instanceName: "test")
        #expect(tweets[0].isPinned == true)
    }

    // MARK: - Empty / Malformed

    @Test("Empty HTML returns no tweets")
    func emptyHTML() throws {
        let tweets = try parser.parseTweets(html: "<html></html>", instanceName: "test")
        #expect(tweets.isEmpty)
    }

    @Test("Skip timeline items with empty content")
    func skipEmptyContent() throws {
        let html = """
        <div class="timeline-item">
            <a class="tweet-link" href="/user/status/111#m"></a>
            <div class="fullname">User</div>
            <div class="username">@user</div>
            <div class="tweet-content media-body"></div>
        </div>
        """
        let tweets = try parser.parseTweets(html: html, instanceName: "test")
        #expect(tweets.isEmpty)
    }

    @Test("Skip show-more items")
    func skipShowMore() throws {
        let html = """
        <div class="timeline-item show-more">
            <a href="/user?cursor=abc">Show more</a>
        </div>
        """
        let tweets = try parser.parseTweets(html: html, instanceName: "test")
        #expect(tweets.isEmpty)
    }

    // MARK: - Profile Parsing

    static let profileHTML = """
    <div class="profile-card">
        <a class="profile-card-fullname" href="/jack">Jack</a>
        <a class="profile-card-username" href="/jack">@jack</a>
        <div class="profile-bio">Just setting up my twttr</div>
        <ul class="profile-statlist">
            <li class="profile-stat">
                <span class="profile-stat-header">Tweets</span>
                <span class="profile-stat-num">29,347</span>
            </li>
            <li class="profile-stat">
                <span class="profile-stat-header">Following</span>
                <span class="profile-stat-num">4,573</span>
            </li>
            <li class="profile-stat">
                <span class="profile-stat-header">Followers</span>
                <span class="profile-stat-num">6.6M</span>
            </li>
            <li class="profile-stat">
                <span class="profile-stat-header">Likes</span>
                <span class="profile-stat-num">35,152</span>
            </li>
        </ul>
    </div>
    """

    @Test("Parse profile name and handle")
    func parseProfileIdentity() throws {
        let profile = try parser.parseProfile(html: Self.profileHTML)
        #expect(profile.fullName == "Jack")
        #expect(profile.handle == "@jack")
    }

    @Test("Parse profile bio")
    func parseProfileBio() throws {
        let profile = try parser.parseProfile(html: Self.profileHTML)
        #expect(profile.bio == "Just setting up my twttr")
    }

    @Test("Parse profile stats")
    func parseProfileStats() throws {
        let profile = try parser.parseProfile(html: Self.profileHTML)
        #expect(profile.tweets == "29,347")
        #expect(profile.following == "4,573")
        #expect(profile.followers == "6.6M")
        #expect(profile.likes == "35,152")
    }

    // MARK: - Cursor Parsing

    @Test("Parse cursor from show-more link")
    func parseCursor() {
        let html = """
        <div class="show-more"><a href="/user?cursor=DAABCgABGQO_0dYAAAA">Show more</a></div>
        """
        let cursor = parser.parseCursor(html: html)
        #expect(cursor == "DAABCgABGQO_0dYAAAA")
    }

    @Test("No cursor when no show-more link")
    func noCursor() {
        let cursor = parser.parseCursor(html: "<html></html>")
        #expect(cursor == nil)
    }

    // MARK: - Challenge Detection

    @Test("Detect browser verification challenge", arguments: [
        "Verifying your browser",
        "Just a moment",
        "Making sure you're not a bot",
    ])
    func detectChallenge(phrase: String) {
        let html = "<html><body><p>\(phrase)</p></body></html>"
        #expect(HTMLParser.isChallengePage(html) == true)
    }

    @Test("Normal HTML is not a challenge page")
    func normalPageNotChallenge() {
        #expect(HTMLParser.isChallengePage(Self.timelineHTML) == false)
    }
}
