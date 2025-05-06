import Foundation
import OSLog

final class AuthTodoService {
    
    private let networkManager: NetworkManaging
    private let baseURL = "http://localhost:9988"
    private let logger = Logger(subsystem: "ntnu.stud.SnowFlow.AuthBox", category: "AuthBoxService")
    
    init() {
        let config = UltimateNetworkConfig(
            maxConcurrentLoads: 1,
            maxRetries: 0,
            retryDelay: 1,
            defaultHeaders: ["Service-Class": "AuthBox/Service"],
            defaultTimeout: 5
        )
        networkManager = NetworkManagerFactory.create(config: config)
    }
    
    /// Fetches a random todo (id between 400 and 500) from the protected endpoint.
    func fetchRandomTodo() async throws -> TodoM {
        // build url
        let id = Int.random(in: 400...500)
        guard let url = URL(string: "\(baseURL)/auth/todos/\(id)") else {
            throw NetworkError.invalidURL
        }
        
        // retrieve the session token
        let session = try await SessionManager.shared.getSession()
        let token = session.token
        
        // create the headers
        let headers = [
            "Authorization": "Bearer \(token)"
        ]
        
        // fetch
        let (_, dataTask) = networkManager.fetch(
            url: url,
            method: "GET",
            timeout: 10,
            headers: headers,
            body: nil,
            maxRetries: 0,
            retryDelay: 0
        )
        
        // decode and return
        let data = try await dataTask.value
        let todo: TodoM = try CodecManager.decode(data)
        return todo
    }
}
