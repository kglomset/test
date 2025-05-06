import Foundation

/// Configuration for network operations.
struct UltimateNetworkConfig {
    
    let maxConcurrentLoads: Int
    let maxRetries: Int
    let retryDelay: TimeInterval
    let defaultHeaders: [String: String]
    let defaultTimeout: TimeInterval
    
    /// Default configuration with sensible defaults.
    static let `default` = UltimateNetworkConfig(
        maxConcurrentLoads: 5,
        maxRetries: 2,
        retryDelay: 10,
        defaultHeaders: [:],
        defaultTimeout: 20.0
    )
}
