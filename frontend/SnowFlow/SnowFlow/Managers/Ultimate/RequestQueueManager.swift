import Foundation
import OSLog

/// actor that manages queuing and execution of network requests
actor RequestQueueManager: RequestQueueManaging {
    
    private var activeRequests: Int = 0
    private var fifoQueue = [PendingRequest]()
    private var dynamicQueue = [(request: PendingRequest, readyTime: Date)]()
    private var dynamicQueueTask: Task<Void, Never>?
    
    private let networkManager: UltimateNetworkManager
    
    private let logger = Logger(subsystem: "ntnu.stud.SnowFlow.Network", category: "RequestQueueManager")
    
    /// initialize with logger and network manager
    /// - Parameters:
    ///   - networkManager: network manager to perform requests
    init(
        networkManager: UltimateNetworkManager
    ) {
        self.networkManager = networkManager
    }
    
    /// enqueue a request for execution
    /// - Parameter pendingRequest: the request to enqueue
    func enqueueRequest(_ pendingRequest: PendingRequest) async {
        if activeRequests < networkManager.config.maxConcurrentLoads {
            await executeRequest(pendingRequest)
        } else {
            fifoQueue.append(pendingRequest)
            logger.log("queued \(pendingRequest.request.url?.absoluteString ?? "unknown")")
        }
    }
    
    /// execute a request immediately if within concurrency limits
    /// - Parameter pendingRequest: the request to execute
    private func executeRequest(_ pendingRequest: PendingRequest) async {
        activeRequests += 1

        do {
            let data = try await networkManager.performRequest(
                id: pendingRequest.id,
                request: pendingRequest.request,
                retriesLeft: pendingRequest.retriesLeft
            )
            pendingRequest.continuation.resume(returning: data)
        } catch {
            // if error cancelled propagate immediately
            if let networkError = error as? NetworkError, networkError == .cancelled {
                pendingRequest.continuation.resume(throwing: error)
            } else if networkManager.shouldRetryForError(error) && pendingRequest.retriesLeft > 0 {
                let retryRequest = pendingRequest.withDecrementedRetries()
                await scheduleRetry(retryRequest, delay: networkManager.config.retryDelay)
            } else {
                pendingRequest.continuation.resume(throwing: error)
            }
        }

        await completeRequest()
    }
    
    /// schedule a request for retry after a delay
    /// - Parameters:
    ///   - pendingRequest: the request to retry
    ///   - delay: time interval before retrying
    private func scheduleRetry(_ pendingRequest: PendingRequest, delay: TimeInterval) async {
        let readyTime = Date().addingTimeInterval(delay)
        dynamicQueue.append((request: pendingRequest, readyTime: readyTime))
        logger.log("scheduled retry in \(delay)s \(pendingRequest.request.url?.absoluteString ?? "unknown")")
        
        if dynamicQueueTask == nil {
            dynamicQueueTask = Task { await processRetryQueue() }
        }
    }
    
    /// process the retry queue and execute ready requests
    private func processRetryQueue() async {
        while !dynamicQueue.isEmpty {
            let now = Date()
            let readyItems = dynamicQueue.filter { $0.readyTime <= now }
            
            if !readyItems.isEmpty {
                dynamicQueue.removeAll { $0.readyTime <= now }
                
                // move ready requests to fifo
                for item in readyItems {
                    logger.log("moving \(item.request.request) from dynamic queue to fifo")
                    await enqueueRequest(item.request)
                }
            }
            
            // wait before checking again
            if !dynamicQueue.isEmpty {
                try? await Task.sleep(nanoseconds: 1_000_000_000)
            }
        }
        
        dynamicQueueTask = nil
    }
    
    /// mark a request as complete and process next in fifo if available
    private func completeRequest() async {
        activeRequests -= 1
        
        if !fifoQueue.isEmpty && activeRequests < networkManager.config.maxConcurrentLoads {
            let nextRequest = fifoQueue.removeFirst()
            await executeRequest(nextRequest)
        }
    }
    
    /// cancel a pending request from both queues
    /// - Parameter id: identifier of the request to cancel
    func cancelRequest(with id: UUID) async {
        // cancel and remove pending requests in fifo queue
        fifoQueue.removeAll { pendingRequest in
            if pendingRequest.id == id {
                pendingRequest.continuation.resume(throwing: NetworkError.cancelled)
                logger.log("removed \(pendingRequest.request) from fifo queue cancelled")
                return true
            }
            return false
        }
        
        // cancel and remove pending requests in dynamic queue
        dynamicQueue.removeAll { element in
            if element.request.id == id {
                element.request.continuation.resume(throwing: NetworkError.cancelled)
                logger.log("removed \(element.request.request) from dynamic queue cancelled")
                return true
            }
            return false
        }
    }
}
