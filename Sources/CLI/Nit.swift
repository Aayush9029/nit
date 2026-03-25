import ArgumentParser

@main
struct Nit: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        commandName: "nit",
        abstract: "Browse tweets from X/Twitter via syndication",
        version: "0.2.0",
        subcommands: [TimelineCommand.self, ProfileCommand.self],
        defaultSubcommand: TimelineCommand.self
    )
}
