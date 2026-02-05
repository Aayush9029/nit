import ArgumentParser
import Foundation
import NitterKit

struct SearchCommand: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        commandName: "search",
        abstract: "Search for tweets (best-effort, may be blocked on some instances)"
    )

    @Argument(help: "Search query")
    var query: String

    @Option(name: .long, help: "Maximum number of tweets to display")
    var count: Int?

    @Flag(name: .long, help: "Output as JSON")
    var json = false

    @Option(name: .long, help: "Custom Nitter instance URL")
    var instance: String?

    func run() async throws {
        let client = NitterClient(customInstance: instance, verbose: !json)
        let (tweets, instanceName) = try await client.search(query: query, count: count)

        if json {
            let encoder = JSONEncoder()
            encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
            let data = try encoder.encode(tweets)
            print(String(data: data, encoding: .utf8)!)
        } else {
            print(Formatter.formatSearchResults(tweets, instance: instanceName))
        }
    }
}
