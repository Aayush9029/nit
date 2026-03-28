import ArgumentParser
import Foundation
import NitterKit

struct TimelineCommand: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        commandName: "timeline",
        abstract: "Show a user's timeline (default command)"
    )

    @Argument(help: "Twitter/X username (without @)")
    var username: String

    @Flag(name: .long, help: "Output as JSON")
    var json = false

    func run() async throws {
        let client = NitterClient(verbose: !json)
        let result = try await client.fetchTimeline(
            username: username.hasPrefix("@") ? String(username.dropFirst()) : username
        )

        if json {
            let encoder = JSONEncoder()
            encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
            let data = try encoder.encode(result)
            print(String(data: data, encoding: .utf8)!)
        } else {
            print(Formatter.formatTimeline(result))
        }
    }
}
