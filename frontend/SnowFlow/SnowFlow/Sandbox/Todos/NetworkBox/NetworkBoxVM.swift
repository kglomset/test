import SwiftUI
import Combine

class NetworkCancellationViewModel: ObservableObject {
    
    private let service: NetworkService
    
    @Published var logs: [String] = []
    @Published var isLoading = false
    @Published var activeRequests: [UUID: String] = [:]
    
    init() {
        service = NetworkService()
        log("network cancellation viewmodel initialized")
    }
    
    // MARK: - test methods
    
    /// case 1 cancel an in-progress request
    func testCancelInProgressRequest() {
        let (token, task) = service.fetchTodoNoRetry(id: 1, delay: 3)
        addActiveRequest(id: token.id, url: "todos/1 with 3s delay")
        log("starting slow request \(token.id.uuidString)")
        
        Task {
            try await Task.sleep(nanoseconds: 1_000_000_000)
            log("requesting cancellation \(token.id.uuidString)")
            await service.cancelRequest(with: token)
        }
        
        // await the task result
        Task {
            do {
                let data = try await task.value
                log("\(token.id.uuidString) completed, received \(data.count) bytes")
                removeActiveRequest(id: token.id)
            } catch {
                log("request failed \(token.id.uuidString) \(error.localizedDescription)")
                removeActiveRequest(id: token.id)
            }
        }
    }
    
    /// case 2 cancel a request during retry cooldown
    func testCancelDuringRetryCooldown() {
        let (token, task) = service.fetchTodoWithCustomRetry(
            id: 1,
            errorRate: 100,
            maxRetries: 3,
            retryDelay: 20
        )
        addActiveRequest(id: token.id, url: "todos/1 with error rate 100%")
        log("starting error prone request \(token.id.uuidString)")
        
        Task {
            try await Task.sleep(nanoseconds: 2_500_000_000)
            log("requesting cancellation during retry cooldown \(token.id.uuidString)")
            await service.cancelRequest(with: token)
        }
        
        Task {
            do {
                let data = try await task.value
                log("\(token.id.uuidString) completed, received \(data.count) bytes")
                removeActiveRequest(id: token.id)
            } catch {
                log("request failed \(token.id.uuidString) \(error.localizedDescription)")
                removeActiveRequest(id: token.id)
            }
        }
    }
    
    /// case 3 cancel a request waiting in queue
    func testCancelQueuedRequest() {
        // create 2 long running requests to fill concurrency slots
        let (token1, task1) = service.fetchTodo(id: 1, delay: 5)
        let (token2, task2) = service.fetchTodo(id: 2, delay: 5)
        let (token3, task3) = service.fetchTodo(id: 3)
        
        // start first blocking request
        log("starting blocking request 1 \(token1.id.uuidString)")
        addActiveRequest(id: token1.id, url: "todos/1 with 5s delay")
        Task {
            do {
                _ = try await task1.value
                log("blocking request 1 completed \(token1.id.uuidString)")
                removeActiveRequest(id: token1.id)
            } catch {
                log("blocking request 1 failed \(token1.id.uuidString) \(error.localizedDescription)")
                removeActiveRequest(id: token1.id)
            }
        }
        
        // start second blocking request
        log("starting blocking request 2 \(token2.id.uuidString)")
        addActiveRequest(id: token2.id, url: "todos/2 with 5s delay")
        Task {
            do {
                _ = try await task2.value
                log("blocking request 2 completed \(token2.id.uuidString)")
                removeActiveRequest(id: token2.id)
            } catch {
                log("blocking request 2 failed \(token2.id.uuidString) \(error.localizedDescription)")
                removeActiveRequest(id: token2.id)
            }
        }
        
        // wait briefly then start queued request and cancel it
        Task {
            try await Task.sleep(nanoseconds: 1_500_000_000)
            log("starting queued request \(token3.id.uuidString)")
            addActiveRequest(id: token3.id, url: "todos/3")
            
            try await Task.sleep(nanoseconds: 1_000_000_000)
            log("requesting cancellation of queued request \(token3.id.uuidString)")
            await service.cancelRequest(with: token3)
            
            Task {
                do {
                    let data = try await task3.value
                    log("\(token3.id.uuidString) completed, received \(data.count) bytes")
                    removeActiveRequest(id: token3.id)
                } catch {
                    log("queued request failed \(token3.id.uuidString) \(error.localizedDescription)")
                    removeActiveRequest(id: token3.id)
                }
            }
        }
    }
    
    /// case 4 multiple requests with selective cancellation
    func testMultipleRequestsWithCancellation() {
        var tokens: [NetworkRequestToken] = []
        
        let requests = [
            (id: 1, delay: 2, errorRate: 0),
            (id: 2, delay: 4, errorRate: 0),
            (id: 3, delay: 0, errorRate: 50),
            (id: 4, delay: 0, errorRate: 0)
        ]
        
        // start all requests
        for (index, req) in requests.enumerated() {
            let (token, task) = service.fetchTodoWithCustomRetry(
                id: req.id,
                delay: req.delay,
                errorRate: req.errorRate,
                maxRetries: 1,
                retryDelay: 1
            )
            tokens.append(token)
            
            // create a descriptive URL for UI display
            let urlDescription = "todos/\(req.id)" +
                (req.delay > 0 ? " with \(req.delay)s delay" : "") +
                (req.errorRate > 0 ? " error rate \(req.errorRate)%" : "")
            
            log("starting request \(index + 1) \(token.id.uuidString)")
            addActiveRequest(id: token.id, url: urlDescription)
            Task {
                do {
                    let data = try await task.value
                    log("\(token.id.uuidString) completed request \(index + 1), received \(data.count) bytes")
                    removeActiveRequest(id: token.id)
                } catch {
                    log("request \(index + 1) failed \(token.id.uuidString) \(error.localizedDescription)")
                    removeActiveRequest(id: token.id)
                }
            }
        }
        
        // cancel second and third requests after delays
        Task {
            try await Task.sleep(nanoseconds: 1_000_000_000)
            log("requesting cancellation of request 2 \(tokens[1].id.uuidString)")
            await service.cancelRequest(with: tokens[1])
            
            try await Task.sleep(nanoseconds: 1_000_000_000)
            log("requesting cancellation of request 3 \(tokens[2].id.uuidString)")
            await service.cancelRequest(with: tokens[2])
        }
    }
    
    /// case 5 cancel an already completed request
    func testCancelCompletedRequest() {
        let (token, task) = service.fetchTodo(id: 1)
        
        addActiveRequest(id: token.id, url: "todos/1")
        log("starting fast request \(token.id.uuidString)")
        
        Task {
            do {
                let data = try await task.value
                log("\(token.id.uuidString) completed, received \(data.count) bytes")
                removeActiveRequest(id: token.id)
                
                try await Task.sleep(nanoseconds: 500_000_000)
                log("attempting cancellation on completed request \(token.id.uuidString)")
                await service.cancelRequest(with: token)
            } catch {
                log("fast request failed \(token.id.uuidString) \(error.localizedDescription)")
                removeActiveRequest(id: token.id)
            }
        }
    }
    
    /// case 6 cancel a non-existent request
    func testCancelNonExistentRequest() {
        let nonExistentToken = NetworkRequestToken()
        log("attempting cancellation of non existent request \(nonExistentToken.id.uuidString)")
        Task {
            await service.cancelRequest(with: nonExistentToken)
            log("cancel non existent request call completed")
        }
    }
    
    // MARK: - helper methods
    
    // add active request to ui
    private func addActiveRequest(id: UUID, url: String) {
        Task { @MainActor in
            self.activeRequests[id] = url
            self.isLoading = !self.activeRequests.isEmpty
        }
    }

    // remove active request from ui
    private func removeActiveRequest(id: UUID) {
        Task { @MainActor in
            self.activeRequests.removeValue(forKey: id)
            self.isLoading = !self.activeRequests.isEmpty
        }
    }

    // clear logs in ui
    func clearLogs() {
        Task { @MainActor in
            self.logs = []
        }
    }

    // log a message with timestamp
    func log(_ message: String) {
        let timestamp = DateFormatter.localizedString(from: Date(), dateStyle: .none, timeStyle: .medium)
        Task { @MainActor in
            self.logs.append("[\(timestamp)] \(message)")
        }
    }
}
