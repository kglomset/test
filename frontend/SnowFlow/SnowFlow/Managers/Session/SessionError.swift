import Foundation

/// Session errors that can occur during session operations.
enum SessionError: LocalizedError {
    case missing
    case expired
    case invalid

    var errorDescription: String? {
        switch self {
        case .missing:
            return "Session is missing."
        case .expired:
            return "Session has expired."
        case .invalid:
            return "Session data is invalid."
        }
    }
}
