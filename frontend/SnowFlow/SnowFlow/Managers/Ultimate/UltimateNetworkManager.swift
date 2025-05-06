import Foundation
import OSLog

struct NetworkRequestToken: Sendable, Equatable {
    public let id: UUID
    public init(id: UUID = UUID()) {
        self.id = id
    }
}

class UltimateNetworkManager: NetworkManaging {
    
    private let logger = Logger(subsystem: "ntnu.stud.SnowFlow.Network", category: "UltimateNetworkManager")
    
    let config: UltimateNetworkConfig
    
    actor InFlightTasksStore {
        // private tasks storage
        private var tasks: [UUID: URLSessionDataTask] = [:]
        
        func store(task: URLSessionDataTask, for id: UUID) {
            tasks[id] = task
        }
        
        func removeTask(for id: UUID) {
            tasks.removeValue(forKey: id)
        }
        
        func task(for id: UUID) -> URLSessionDataTask? {
            return tasks[id]
        }
    }
    
    private let inFlightTasksStore = InFlightTasksStore()
    private let session: URLSessionProtocol
    
    private lazy var queueManager: RequestQueueManager = RequestQueueManager(networkManager: self)
    
    /// creates a new network manager with the specified configuration.
    /// - parameters:
    ///   - config: configuration for network operations.
    ///   - session: urlsession protocol implementation for executing network requests.
    init(
        config: UltimateNetworkConfig,
        session: URLSessionProtocol
    ) {
        self.config = config
        self.session = session
    }
    
    func fetch(
        url: URL,
        method: String = "GET",
        timeout: TimeInterval? = nil,
        headers: [String: String]? = nil,
        body: Data? = nil,
        maxRetries: Int? = nil,
        retryDelay: TimeInterval? = nil
    ) -> (token: NetworkRequestToken, task: Task<Data, Error>) {
        let token = NetworkRequestToken()  // create a token with a new uuid
        let task = Task<Data, Error> {
            // build urlrequest
            var request = URLRequest(url: url)
            request.httpMethod = method
            request.timeoutInterval = timeout ?? config.defaultTimeout
            
            // set the provided body if available
            request.httpBody = body
            
            let finalHeaders = config.defaultHeaders.merging(headers ?? [:]) { _, new in new }
            for (key, value) in finalHeaders {
                request.setValue(value, forHTTPHeaderField: key)
            }
            
            let retriesLeft = maxRetries ?? config.maxRetries
            
            return try await withCheckedThrowingContinuation { continuation in
                let pendingRequest = PendingRequest(
                    id: token.id,
                    request: request,
                    continuation: continuation,
                    retriesLeft: retriesLeft
                )
                // enqueue the request
                Task {
                    await queueManager.enqueueRequest(pendingRequest)
                }
            }
        }
        return (token, task)
    }
    
    /// fetches data with caching. checks the cache first (using the url as the key) and, if the cached entry is valid, returns it.
    /// otherwise, performs a network request and caches the new data.
    func fetchWithCache(
        url: URL,
        method: String = "GET",
        timeout: TimeInterval? = nil,
        requestHeaders: [String: String]? = nil,
        expectedContentType: ContentType,
        maxRetries: Int? = nil,
        retryDelay: TimeInterval? = nil
    ) -> (token: NetworkRequestToken, task: Task<Data, Error>) {
        let cacheKey = url.absoluteString
        
        // check cache first
        if let cachedData = CacheManager.shared.loadValidData(forKey: cacheKey, contentType: expectedContentType) {
            return (NetworkRequestToken(), Task { cachedData })
        }
        
        // if cache miss or entry is stale, perform a network request
        let (token, networkTask) = fetch(
            url: url,
            method: method,
            timeout: timeout,
            headers: requestHeaders,
            maxRetries: maxRetries,
            retryDelay: retryDelay
        )
        
        // cache the network response once received
        let cachingTask = Task<Data, Error> {
            let data = try await networkTask.value
            CacheManager.shared.saveData(data, forKey: cacheKey, contentType: expectedContentType)
            return data
        }
        
        return (token, cachingTask)
    }
    
    func cancelRequest(with token: NetworkRequestToken) async {
        logger.log("attempting to cancel request: \(token.id.uuidString)")
        if let task = await inFlightTasksStore.task(for: token.id) {
            logger.log("found in-flight task for \(token.id.uuidString); calling cancel()")
            task.cancel()
        } else {
            logger.log("no in-flight task found for \(token.id.uuidString)")
        }
        await queueManager.cancelRequest(with: token.id)
    }
    
    /// performs the network request once.
    /// - parameters:
    ///   - id: the id of the request.
    ///   - request: the request to perform.
    ///   - retriesLeft: number of retries remaining.
    /// - returns: the fetched data.
    func performRequest(
        id: UUID,
        request: URLRequest,
        retriesLeft: Int
    ) async throws -> Data {
        logger.log("started fetching: \(request.url?.absoluteString ?? "unknown")")
        let tasksStore = self.inFlightTasksStore
        
        return try await withCheckedThrowingContinuation { continuation in
            Task {
                // capture logger locally to avoid capturing self in the Task
                let localLogger = self.logger
                
                let task = (session as! URLSession).dataTask(with: request, completionHandler: { (data: Data?, response: URLResponse?, error: Error?) in
                    Task {
                        await tasksStore.removeTask(for: id)
                        
                        if let error = error {
                            let nsError = error as NSError
                            if nsError.code == NSURLErrorCancelled {
                                localLogger.log("datatask \(id.uuidString) was cancelled (nsurlerrorcancelled).")
                            } else {
                                localLogger.log("datatask \(id.uuidString) error: \(error)")
                            }
                            continuation.resume(throwing: NetworkErrorMapper.mapUrlErrorToNetworkError(error))
                            return
                        }
                        
                        if let data = data, let response = response {
                            if let httpResponse = response as? HTTPURLResponse {
                                let statusCode = httpResponse.statusCode
                                if !(200...299).contains(statusCode) {
                                    // todo add a check for expected contenttype?
                                    let networkError = NetworkErrorMapper.mapHttpStatusToNetworkError(statusCode)
                                    localLogger.log("datatask \(id.uuidString) received http error: \(statusCode)")
                                    continuation.resume(throwing: networkError)
                                    return
                                }
                            }
                            localLogger.log("done fetching: \(request.url?.absoluteString ?? "unknown") for \(id.uuidString)")
                            continuation.resume(returning: data)
                        } else {
                            localLogger.log("datatask \(id.uuidString) received unexpected data/response.")
                            continuation.resume(throwing: NetworkError.unexpectedData)
                        }
                    }
                })
                
                // ensure the task is stored before starting
                await tasksStore.store(task: task, for: id)
                logger.log("stored in-flight task for \(id.uuidString); starting task.")
                task.resume()
            }
        }
    }
    
    private func storeTask(_ task: URLSessionDataTask, for id: UUID) {
        Task {
            await inFlightTasksStore.store(task: task, for: id)
        }
    }
    
    private func removeTask(for id: UUID) {
        Task {
            await inFlightTasksStore.removeTask(for: id)
        }
    }
    
    /// returns true for errors that should be retried.
    /// - parameter error: the error to check.
    /// - returns: whether the error should be retried.
    func shouldRetryForError(_ error: Error) -> Bool {
        if let networkError = error as? NetworkError {
            return networkError.isRetriable
        }
        return false
    }
    
    /// tears down the network manager by invalidating its underlying urlsession.
    /// - parameter cancelInProgressTasks: when true, immediately cancels all in-progress tasks.
    ///   when false, allows in-progress tasks to complete before invalidating.
    /// - returns: true if teardown was successful.
    func tearDown(cancelInProgressTasks: Bool = false) -> Bool {
        guard let urlSession = session as? URLSession else {
            logger.log("teardown failed: urlsession could not be accessed")
            return false
        }
        
        logger.log("teardown: \(cancelInProgressTasks ? "cancelling all tasks" : "finishing tasks before invalidation")")
        
        if cancelInProgressTasks {
            urlSession.invalidateAndCancel()
        } else {
            urlSession.finishTasksAndInvalidate()
        }
        
        return true
    }
}
