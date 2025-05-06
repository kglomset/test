import Foundation
import Combine

class NetworkService {
    private let networkManager: UltimateNetworkManager
    private let baseUrl = "http://localhost:9988"
    
    init() {
        let config = UltimateNetworkConfig(
            maxConcurrentLoads: 2,
            maxRetries: 2,
            retryDelay: 1.8,
            defaultHeaders: ["Service-Class": "network/box"],
            defaultTimeout: 12
        )
        networkManager = NetworkManagerFactory.create(config: config)
    }
    
    // helper method to construct the URL
    private func constructURL(id: Int, delay: Int, errorRate: Int) -> URL {
        var urlComponents = "\(baseUrl)"
        
        if delay > 0 {
            urlComponents += "/delay/\(delay)"
        }
        
        if errorRate > 0 {
            urlComponents += "/error/\(errorRate)"
        }
        
        urlComponents += "/todos/\(id)"
        
        guard let url = URL(string: urlComponents) else {
            fatalError("Invalid URL: \(urlComponents)")
        }
        return url
    }
    
    func fetchTodo(
        id: Int,
        delay: Int = 0,
        errorRate: Int = 0
    ) -> (NetworkRequestToken, Task<Data, Error>) {
        let url = constructURL(id: id, delay: delay, errorRate: errorRate)
        return networkManager.fetch(
            url: url,
            method: "GET",
            timeout: 10.0,
            headers: ["Content-Type": "application/json"]
        )
    }
    
    func fetchTodoNoRetry(
        id: Int,
        delay: Int = 0,
        errorRate: Int = 0
    ) -> (NetworkRequestToken, Task<Data, Error>) {
        let url = constructURL(id: id, delay: delay, errorRate: errorRate)
        return networkManager.fetch(
            url: url,
            method: "GET",
            timeout: 10.0,
            headers: ["Content-Type": "application/json"],
            maxRetries: 0
        )
    }
    
    func fetchTodoWithCustomRetry(
        id: Int, delay: Int = 0,
        errorRate: Int = 0,
        maxRetries: Int,
        retryDelay: TimeInterval
    ) -> (NetworkRequestToken, Task<Data, Error>) {
        let url = constructURL(id: id, delay: delay, errorRate: errorRate)
        return networkManager.fetch(
            url: url,
            method: "GET",
            timeout: 10.0,
            headers: ["Content-Type": "application/json"],
            maxRetries: maxRetries,
            retryDelay: retryDelay
        )
    }
    
    func cancelRequest(with token: NetworkRequestToken) async {
        await networkManager.cancelRequest(with: token)
    }
}
