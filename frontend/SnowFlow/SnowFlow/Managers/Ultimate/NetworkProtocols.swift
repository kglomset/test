import Foundation

/// An async version of URLSession for dependency injection.
protocol URLSessionProtocol {
    func data(from request: URLRequest) async throws -> (Data, URLResponse)
    func dataTask(with request: URLRequest) -> URLSessionDataTask
}

extension URLSession: URLSessionProtocol {
    func data(from request: URLRequest) async throws -> (Data, URLResponse) {
        return try await self.data(for: request)
    }
    
    func dataTask(with request: URLRequest) -> URLSessionDataTask {
        // dummy
        return self.dataTask(with: request, completionHandler: { _, _, _ in })
    }
}

/// Network manager interface for fetching data and managing requests.
protocol NetworkManaging {
    
    func fetch(
        url: URL,
        method: String,
        timeout: TimeInterval?,
        headers: [String: String]?,
        body: Data?,
        maxRetries: Int?,
        retryDelay: TimeInterval?
    ) -> (token: NetworkRequestToken, task: Task<Data, Error>)
    
    func fetchWithCache(
        url: URL,
        method: String,
        timeout: TimeInterval?,
        requestHeaders: [String: String]?,
        expectedContentType: ContentType,
        maxRetries: Int?,
        retryDelay: TimeInterval?
    ) -> (token: NetworkRequestToken, task: Task<Data, Error>)
    
    func cancelRequest(with token: NetworkRequestToken) async
}

/// Manages request queuing, execution scheduling, and cancellation.
protocol RequestQueueManaging: AnyObject {
    func enqueueRequest(
        _ pendingRequest: PendingRequest
    ) async
    
    func cancelRequest(with id: UUID) async
}
