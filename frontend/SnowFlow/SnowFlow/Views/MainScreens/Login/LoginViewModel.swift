import Foundation

@MainActor
final class LoginViewModel: ObservableObject {
    
    @Published var email: String = ""
    @Published var password: String = ""
    @Published var errorMessage: String? = nil
    
    private let service = SnowFlowService()
    private let session = SessionManager.shared
    
    /// Initiates the login process.
    /// On success, saves the session via the SessionManager and calls onSuccess.
    // in an offline scenario (managed by global state), we will skip login altogether i think
    func login(onSuccess: @escaping () -> Void) {
        errorMessage = nil
        
        // todo check for offline mode before attempting login
        // e.g. if GlobalAppState.shared.isOffline { return }
        Task {
            do {
                let response = try await service.login(email: email, password: password)
                await session.saveSession(token: response.sessionToken, expiration: response.expiresAt)
                onSuccess()
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }
}

