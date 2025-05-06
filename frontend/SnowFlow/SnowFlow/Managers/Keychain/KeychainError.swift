import Foundation

enum KeychainError: LocalizedError {
    case saveFailed(OSStatus)
    case retrieveFailed(OSStatus)
    case deleteFailed(OSStatus)

    var errorDescription: String? {
        switch self {
        case .saveFailed(let status):
            return "Failed to save item to keychain. OSStatus: \(status)"
        case .retrieveFailed(let status):
            return "Failed to retrieve item from keychain. OSStatus: \(status)"
        case .deleteFailed(let status):
            return "Failed to delete item from keychain. OSStatus: \(status)"
        }
    }
}
