// swift-tools-version: 6.0

import PackageDescription

let package = Package(
    name: "nit",
    platforms: [.macOS(.v13)],
    products: [
        .executable(name: "nit", targets: ["CLI"]),
    ],
    dependencies: [
        .package(url: "https://github.com/apple/swift-argument-parser.git", from: "1.3.0"),
    ],
    targets: [
        .target(
            name: "NitterKit"
        ),
        .executableTarget(
            name: "CLI",
            dependencies: [
                "NitterKit",
                .product(name: "ArgumentParser", package: "swift-argument-parser"),
            ]
        ),
        .testTarget(
            name: "NitterKitTests",
            dependencies: ["NitterKit"]
        ),
    ]
)
