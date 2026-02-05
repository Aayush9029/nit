import Foundation
#if canImport(FoundationNetworking)
import FoundationNetworking
#endif

public actor NitterClient {
    private let instances: [NitterInstance]
    private let parser = HTMLParser()
    private let session: URLSession
    private let verbose: Bool

    public init(customInstance: String? = nil, verbose: Bool = true) {
        if let custom = customInstance {
            let url = custom.hasSuffix("/") ? String(custom.dropLast()) : custom
            self.instances = [NitterInstance(baseURL: url, name: url)]
        } else {
            self.instances = NitterInstance.defaultInstances
        }
        self.verbose = verbose

        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 15
        config.timeoutIntervalForResource = 30
        config.httpAdditionalHeaders = [
            "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
            "Accept-Language": "en-US,en;q=0.9",
        ]
        self.session = URLSession(configuration: config)
    }

    // MARK: - Public API

    public func fetchTimeline(username: String, count: Int? = nil, cursor: String? = nil) async throws -> TimelineResult {
        var path = "/\(username)"
        if let cursor {
            path += "?cursor=\(cursor)"
        }

        let (html, instance) = try await fetchHTML(path: path)

        // Check for user not found
        if html.contains("User \"") && html.contains("\" not found") {
            throw NitterError.userNotFound(username)
        }

        var tweets = try parser.parseTweets(html: html, instanceName: instance.name)
        let profile = try? parser.parseProfile(html: html)
        let nextCursor = parser.parseCursor(html: html)

        if let count {
            tweets = Array(tweets.prefix(count))
        }

        return TimelineResult(
            profile: profile,
            tweets: tweets,
            instance: instance.name,
            cursor: nextCursor
        )
    }

    public func fetchProfile(username: String) async throws -> (Profile, String) {
        let path = "/\(username)"
        let (html, instance) = try await fetchHTML(path: path)

        if html.contains("User \"") && html.contains("\" not found") {
            throw NitterError.userNotFound(username)
        }

        let profile = try parser.parseProfile(html: html)
        return (profile, instance.name)
    }

    public func search(query: String, count: Int? = nil) async throws -> ([Tweet], String) {
        let encoded = query.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? query
        let path = "/search?f=tweets&q=\(encoded)"
        let (html, instance) = try await fetchHTML(path: path)

        var tweets = try parser.parseTweets(html: html, instanceName: instance.name)
        if let count {
            tweets = Array(tweets.prefix(count))
        }

        return (tweets, instance.name)
    }

    // MARK: - Internal Fetch with Fallback

    private func fetchHTML(path: String) async throws -> (String, NitterInstance) {
        var lastError: Error = NitterError.allInstancesFailed

        for instance in instances {
            let urlString = instance.baseURL + path
            guard let url = URL(string: urlString) else { continue }

            if verbose {
                printDim("  Trying \(instance.name)...")
            }

            do {
                let (data, response) = try await session.data(from: url)

                guard let httpResponse = response as? HTTPURLResponse else {
                    continue
                }

                // Check for empty response
                if data.isEmpty {
                    if verbose { printDim("  ↳ Empty response, skipping") }
                    continue
                }

                guard let html = String(data: data, encoding: .utf8) else {
                    continue
                }

                // Check for challenge pages
                if HTMLParser.isChallengePage(html) {
                    if verbose { printDim("  ↳ Challenge page detected, skipping") }
                    continue
                }

                // Accept 200 responses
                guard httpResponse.statusCode == 200 else {
                    if verbose { printDim("  ↳ HTTP \(httpResponse.statusCode), skipping") }
                    continue
                }

                return (html, instance)

            } catch {
                lastError = error
                if verbose { printDim("  ↳ \(error.localizedDescription), skipping") }
                continue
            }
        }

        throw lastError
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
