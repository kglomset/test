import OSLog
import Foundation

final class SnowFlowService {
    private let networkManager: NetworkManaging
    private let baseURL = "http://localhost:9988"
    private let logger = Logger(subsystem: "ntnu.stud.SnowFlow.snowflow", category: "snowFlowService")
    
    init() {
        let config = UltimateNetworkConfig(
            maxConcurrentLoads: 5,
            maxRetries: 2,
            retryDelay: 1.8,
            defaultHeaders: ["Service-Class": "SnowFlow/Service"],
            defaultTimeout: 12
        )
        networkManager = NetworkManagerFactory.create(config: config)
    }
    
    /// Generic perform request that listens for a 401 error and logs responses.
    private func performRequest<T: Decodable>(
        url: URL,
        method: String,
        timeout: TimeInterval? = nil,
        headers: [String: String] = [:],
        body: Data? = nil,
        maxRetries: Int? = nil,
        retryDelay: TimeInterval? = nil
    ) async throws -> T {
        
        let (_, dataTask) = networkManager.fetch(
            url: url,
            method: method,
            timeout: timeout,
            headers: headers,
            body: body,
            maxRetries: maxRetries,
            retryDelay: retryDelay
        )
        
        do {
            let data = try await dataTask.value
            return try CodecManager.decode(data)
        } catch let error as NetworkError {
            if case .httpError(let statusCode) = error, statusCode == 401 {
                await SessionManager.shared.clearSession()
                throw AuthError.invalidCredentials
            }
            throw error
        }
    }
    
    // MARK: - Login
    
    struct LoginRequest: Encodable {
        let email: String
        let password: String
    }
    
    struct LoginResponse: Decodable {
        let expiresAt: String
        let sessionToken: String
        
        enum CodingKeys: String, CodingKey {
            case expiresAt = "expires_at"
            case sessionToken = "session_token"
        }
    }
    
    /// Login method: Sends email and password to the server to authenticate.
    /// - Parameters:
    ///   - email: User's email.
    ///   - password: User's password.
    /// - Returns: A `LoginResponse` on success.
    /// - Throws: Underlying errors such as `NetworkError`, `AuthError`, or `CodecError`.
    func login(email: String, password: String) async throws -> LoginResponse {
        guard let url = URL(string: "\(baseURL)/login") else {
            throw NetworkError.invalidURL
        }
        
        // Prepare the login request payload.
        let loginRequest = LoginRequest(email: email, password: password)
        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        let requestBody: Data
        do {
            requestBody = try encoder.encode(loginRequest)
            if let jsonString = String(data: requestBody, encoding: .utf8) {
                logger.log("sending login request with body: \(jsonString, privacy: .public)")
            }
        } catch {
            logger.error("error encoding login request: \(String(describing: error), privacy: .public)")
            throw error
        }
        
        let headers = ["Content-Type": "application/json"]
        
        // use the middleware performRequest
        return try await performRequest(
            url: url,
            method: "POST",
            timeout: 20,
            headers: headers,
            body: requestBody,
            maxRetries: 0,
            retryDelay: 0
        )
    }
    
    // MARK: - Session active check
    
    struct SessionActiveResponse: Codable {
        let active: Bool
    }
    
    /// Calls the /is-session-active endpoint and returns the active status.
    func checkSessionActive() async throws -> Bool {
        guard let url = URL(string: "\(baseURL)/is-session-active") else {
            throw NetworkError.invalidURL
        }
        
        let session = try await SessionManager.shared.getSession()
        let headers = [
            "Authorization": "Bearer \(session.token)"
        ]
        
        // use the middleware performRequest
        let response: SessionActiveResponse = try await performRequest(
            url: url,
            method: "GET",
            timeout: 10,
            headers: headers,
            body: nil,
            maxRetries: 0,
            retryDelay: 0
        )
        return response.active
    }
    
    // MARK: - temp
    
    struct TempTodo: Codable {
        let id: Int
        let title: String
        let completed: Bool
        
        enum CodingKeys: String, CodingKey {
            case id
            case title
            case completed
        }
    }
    
    /// Calls the /is-session-active endpoint and returns the active status.
    func getTempTodo() async throws -> Bool {
        guard let url = URL(string: "\(baseURL)/auth/todos/1") else {
            throw NetworkError.invalidURL
        }
        
        let session = try await SessionManager.shared.getSession()
        let headers = [
            "Authorization": "Bearer \(session.token)"
        ]
        
        // use the middleware performRequest
        let response: SessionActiveResponse = try await performRequest(
            url: url,
            method: "GET",
            timeout: 10,
            headers: headers,
            body: nil,
            maxRetries: 2,
            retryDelay: 4
        )
        return response.active
    }
}
