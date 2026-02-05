import ArgumentParser

@main
struct Nit: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        commandName: "nit",
        abstract: "Browse tweets via Nitter instances",
        version: "0.1.0",
        subcommands: [TimelineCommand.self, ProfileCommand.self, SearchCommand.self],
        defaultSubcommand: TimelineCommand.self
    )
}
