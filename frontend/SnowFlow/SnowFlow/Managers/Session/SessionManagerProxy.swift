import Combine
import Foundation

// the class prioritizes authentication, it is auth optimistic
// this will help ensure offline compatibility in the future

@MainActor
class SessionManagerProxy: ObservableObject {
    
    /// Published authentication state for UI observation.
    @Published var isAuthenticated: Bool = true
    
    static let shared = SessionManagerProxy()
    private init() {
        Task {
            if await !SessionManager.shared.isSessionActive() {
                isAuthenticated = false
            }
        }
    }
}
