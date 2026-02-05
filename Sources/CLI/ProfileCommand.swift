import ArgumentParser
import Foundation
import NitterKit

struct ProfileCommand: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        commandName: "profile",
        abstract: "Show a user's profile info"
    )

    @Argument(help: "Twitter username (without @)")
    var username: String

    @Flag(name: .long, help: "Output as JSON")
    var json = false

    @Option(name: .long, help: "Custom Nitter instance URL")
    var instance: String?

    func run() async throws {
        let client = NitterClient(customInstance: instance, verbose: !json)
        let (profile, instanceName) = try await client.fetchProfile(
            username: username.hasPrefix("@") ? String(username.dropFirst()) : username
        )

        if json {
            let encoder = JSONEncoder()
            encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
            let data = try encoder.encode(profile)
            print(String(data: data, encoding: .utf8)!)
        } else {
            print(Formatter.formatProfile(profile, instance: instanceName))
        }
    }
}
