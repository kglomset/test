import Foundation

/// Network errors that can occur during network operations.
enum NetworkError: LocalizedError {
    case invalidURL
    case httpError(Int)
    case noInternetConnection
    case serverUnavailable
    case timeout
    case unsupportedContentType(String)
    case unexpectedData
    case unknown(Error)
    case cancelled
    case queueFull
    
    /// Indicates whether the error is retriable.
    var isRetriable: Bool {
        switch self {
        case .serverUnavailable, .timeout, .noInternetConnection:
            return true
        case .invalidURL, .httpError, .unsupportedContentType, .unexpectedData, .unknown, .cancelled, .queueFull:
            return false
        }
    }
    
    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return NSLocalizedString("error_invalid_url", comment: "Invalid URL error")
        case .httpError(let status):
            let errorFormat = NSLocalizedString("error_http", comment: "HTTP error with status code")
            return String(format: errorFormat, status)
        case .noInternetConnection:
            return NSLocalizedString("error_no_internet", comment: "No internet connection error")
        case .serverUnavailable:
            return NSLocalizedString("error_server_unavailable", comment: "Server unavailable error")
        case .timeout:
            return NSLocalizedString("error_timeout", comment: "Request timeout error")
        case .unsupportedContentType(let type):
            let errorFormat = NSLocalizedString("error_unsupported_content", comment: "Unsupported content type error")
            return String(format: errorFormat, type)
        case .unexpectedData:
            return NSLocalizedString("error_unexpected_data", comment: "Unexpected data error")
        case .unknown:
            return NSLocalizedString("error_unknown_network", comment: "Unknown network error")
        case .cancelled:
            return NSLocalizedString("error_cancelled", comment: "Request cancelled")
        case .queueFull:
            return NSLocalizedString("error_queue_full", comment: "Request queue is full")
        }
    }
}

/// Functions to map various error types to NetworkError, primarily to enable custom localization.
struct NetworkErrorMapper {
    /// Maps URLError to NetworkError.
    static func mapUrlErrorToNetworkError(_ error: Error) -> NetworkError {
        if let urlError = error as? URLError {
            switch urlError.code {
            case .notConnectedToInternet:
                return .noInternetConnection
            case .timedOut:
                return .timeout
            case .cannotFindHost, .cannotConnectToHost, .dnsLookupFailed:
                return .serverUnavailable
            case .cancelled:
                return .cancelled
            case .badURL, .unsupportedURL:
                return .invalidURL
            default:
                return .unknown(urlError)
            }
        }
        return .unknown(error)
    }
    
    /// Maps HTTP status code to NetworkError.
    static func mapHttpStatusToNetworkError(_ statusCode: Int) -> NetworkError {
        let retriableServerErrors: Set<Int> = [500, 502, 503, 504, 429]
        
        if retriableServerErrors.contains(statusCode) {
            return .serverUnavailable
        }
        
        if (400...499).contains(statusCode) {
            return .httpError(statusCode)
        }
        
        return .httpError(statusCode)
    }
    
    /// Maps content type to NetworkError if unsupported or unexpected.
    static func validateContentType(_ contentType: String?, expected: ContentType) -> NetworkError? {
        guard let contentType = contentType else {
            return .unexpectedData
        }
        
        let normalizedContentType = contentType
            .split(separator: ";")
            .first?
            .trimmingCharacters(in: .whitespacesAndNewlines)
            .lowercased() ?? ""
        
        if normalizedContentType != expected.rawValue {
            return .unsupportedContentType(normalizedContentType)
        }
        
        return nil
    }
}

extension NetworkError: Equatable {
    public static func == (lhs: NetworkError, rhs: NetworkError) -> Bool {
        switch (lhs, rhs) {
        case (.invalidURL, .invalidURL),
            (.noInternetConnection, .noInternetConnection),
            (.serverUnavailable, .serverUnavailable),
            (.timeout, .timeout),
            (.unexpectedData, .unexpectedData),
            (.cancelled, .cancelled),
            (.queueFull, .queueFull):
            return true
        case (.httpError(let code1), .httpError(let code2)):
            return code1 == code2
        case (.unsupportedContentType(let ct1), .unsupportedContentType(let ct2)):
            return ct1 == ct2
        case (.unknown(let err1), .unknown(let err2)):
            return err1.localizedDescription == err2.localizedDescription
        default:
            return false
        }
    }
}
