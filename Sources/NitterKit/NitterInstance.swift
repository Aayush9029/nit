import Foundation

public struct NitterInstance: Sendable {
    public let baseURL: String
    public let name: String

    public init(baseURL: String, name: String) {
        self.baseURL = baseURL
        self.name = name
    }

    public static let defaultInstances: [NitterInstance] = [
        NitterInstance(baseURL: "https://xcancel.com", name: "xcancel.com"),
        NitterInstance(baseURL: "https://nitter.poast.org", name: "nitter.poast.org"),
        NitterInstance(baseURL: "https://nitter.privacyredirect.com", name: "nitter.privacyredirect.com"),
        NitterInstance(baseURL: "https://lightbrd.com", name: "lightbrd.com"),
        NitterInstance(baseURL: "https://nitter.space", name: "nitter.space"),
        NitterInstance(baseURL: "https://nitter.tiekoetter.com", name: "nitter.tiekoetter.com"),
        NitterInstance(baseURL: "https://nuku.trabun.org", name: "nuku.trabun.org"),
        NitterInstance(baseURL: "https://nitter.catsarch.com", name: "nitter.catsarch.com"),
        NitterInstance(baseURL: "https://nitter.net", name: "nitter.net"),
    ]
}
