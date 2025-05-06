import Foundation

/// A pending request holds a continuation that will be resumed when the request finishes.
/// It also tracks the number of retries left.
struct PendingRequest {
    
    let id: UUID
    let request: URLRequest
    let continuation: CheckedContinuation<Data, Error>
    var retriesLeft: Int
    
    /// Creates a new PendingRequest with specified properties.
    /// - Parameters:
    ///   - id: Unique identifier for the request.
    ///   - request: The URLRequest to be executed.
    ///   - continuation: Continuation that will be resumed when the request completes.
    ///   - retriesLeft: Number of retries remaining for this request.
    init(
        id: UUID,
        request: URLRequest,
        continuation: CheckedContinuation<Data, Error>,
        retriesLeft: Int
    ) {
        self.id = id
        self.request = request
        self.continuation = continuation
        self.retriesLeft = retriesLeft
    }
    
    /// Creates a new PendingRequest with one fewer retry remaining.
    /// - Returns: A new PendingRequest with decremented retriesLeft.
    func withDecrementedRetries() -> PendingRequest {
        PendingRequest(
            id: id,
            request: request,
            continuation: continuation,
            retriesLeft: retriesLeft - 1
        )
    }
}
