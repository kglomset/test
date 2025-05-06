import Foundation
import Security
import OSLog

@globalActor
actor SessionManager {
    
    // constants
    private let sessionTokenKey = "sessionToken"
    private let expiresAtKey = "expiresAt"
    
    // singleton
    static let shared = SessionManager()
    private init() {} // prevent external instantiation
    
    // logger
    private let logger = Logger(subsystem: "ntnu.stud.SnowFlow.Session", category: "SessionManager")
    
    // service
    private var service = SnowFlowService()
    
    // in-memory access to session
    private var _session: SessionM?
    
    // MARK: - Session persistence methods
    
    /// Save session to memory & keychain.
    func saveSession(token: String, expiration: String) async {
        guard let expirationDate = Date.dateFromISO8601String(expiration) else {
            logger.error("failed to parse expiration date from string: \(expiration)")
            return
        }
        _session = SessionM(token: token, expiresAt: expirationDate)
        do {
            try await KeychainHelper.save(key: sessionTokenKey, value: token)
            try await KeychainHelper.save(key: expiresAtKey, value: expiration)
            logger.log("successfully saved session to memory and keychain")
        } catch {
            logger.error("error saving session to keychain: \(error.localizedDescription)")
        }
    }
    
    /// Clear session from memory & keychain.
    func clearSession() async {
        await MainActor.run {
            SessionManagerProxy.shared.isAuthenticated = false
        }
        _session = nil
        logger.log("session cleared from memory")
        do {
            try await KeychainHelper.delete(key: sessionTokenKey)
            try await KeychainHelper.delete(key: expiresAtKey)
            logger.log("session cleared from keychain")
        } catch {
            logger.error("failed to clear session from keychain: \(error.localizedDescription)")
        }
    }
    
    // MARK: - Session active
    
    /// Validates the current session by checking local expiration and calling the network endpoint.
    func isSessionActive() async -> Bool {
        do {
            _ = try await getSession()
            return try await service.checkSessionActive()
        } catch {
            logger.error("session validation failed with error: \(error.localizedDescription)")
            return false
        }
    }
    
    // MARK: - Session access
    
    /// Retrieves a non-optional session. If no valid session exists, it throws a SessionError.
    func getSession() async throws -> SessionM {
        // if the in-memory session is nil, attempt to load it from the keychain.
        if _session == nil {
            await loadSessionFromKeychain()
        }
        
        // if loading still didn't yield a session, throw a missing session error.
        guard let session = _session else {
            throw SessionError.missing
        }
        
        // ensure that the session has not expired
        guard Date() < session.expiresAt else {
            throw SessionError.expired
        }
        
        return session
    }
    
    // MARK: - Helpers
    
    /// Loads the session from the keychain and updates the in-memory session.
    private func loadSessionFromKeychain() async {
        do {
            let token = try await KeychainHelper.retrieve(key: sessionTokenKey)
            let expiration = try await KeychainHelper.retrieve(key: expiresAtKey)
            guard let expirationDate = Date.dateFromISO8601String(expiration) else {
                logger.error("date parsing issue encountered for expiration: \(expiration)")
                return
            }
            _session = SessionM(token: token, expiresAt: expirationDate)
            logger.log("successfully loaded session from keychain")
        } catch {
            logger.error("error loading session from keychain: \(error.localizedDescription)")
        }
    }
}
