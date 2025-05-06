import Foundation

/// factory for creating ultimenetworkmanager instances.
struct NetworkManagerFactory {
    /// creates a new ultimenetworkmanager with the specified configuration.
    /// - parameters:
    ///   - config: configuration for network operations. defaults to .default.
    ///   - sessionconfiguration: configuration for the urlsession. defaults to .default.
    /// - returns: a configured ultimenetworkmanager.
    static func create(
        config: UltimateNetworkConfig = .default,
        sessionConfiguration: URLSessionConfiguration = .default
    ) -> UltimateNetworkManager {
        // configure the session with values from the config.
        sessionConfiguration.timeoutIntervalForRequest = config.defaultTimeout
        sessionConfiguration.httpMaximumConnectionsPerHost = config.maxConcurrentLoads
        
        let session = URLSession(configuration: sessionConfiguration)
        
        return UltimateNetworkManager(
            config: config,
            session: session
        )
    }
}
