import Testing
@testable import NitterKit

@Suite("NitterInstance Tests")
struct NitterInstanceTests {

    @Test("Default instances list has 9 entries")
    func defaultInstanceCount() {
        #expect(NitterInstance.defaultInstances.count == 9)
    }

    @Test("First instance is xcancel.com")
    func firstInstance() {
        #expect(NitterInstance.defaultInstances[0].name == "xcancel.com")
        #expect(NitterInstance.defaultInstances[0].baseURL == "https://xcancel.com")
    }

    @Test("Last instance is nitter.net")
    func lastInstance() {
        let last = NitterInstance.defaultInstances.last!
        #expect(last.name == "nitter.net")
        #expect(last.baseURL == "https://nitter.net")
    }

    @Test("All instances use HTTPS")
    func allHTTPS() {
        for instance in NitterInstance.defaultInstances {
            #expect(instance.baseURL.hasPrefix("https://"))
        }
    }

    @Test("Custom instance creation")
    func customInstance() {
        let instance = NitterInstance(baseURL: "https://my-nitter.example.com", name: "my-nitter.example.com")
        #expect(instance.baseURL == "https://my-nitter.example.com")
        #expect(instance.name == "my-nitter.example.com")
    }
}
