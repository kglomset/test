import Foundation

enum AuthError: LocalizedError {
    case invalidCredentials
    case sessionExpired

    var errorDescription: String? {
        switch self {
        case .invalidCredentials:
            return NSLocalizedString("error_invalid_credentials", comment: "Invalid user credentials error")
        case .sessionExpired:
            return NSLocalizedString("error_session_expired", comment: "User session expired error")
        }
    }
}
