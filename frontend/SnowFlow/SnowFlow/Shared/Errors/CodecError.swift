import Foundation

enum CodecError: LocalizedError {
    case jsonEncodingFailed
    case jsonDecodingFailed
    case utf8EncodingFailed
    case utf8DecodingFailed
    case imageEncodingFailed
    case imageDecodingFailed

    var errorDescription: String? {
        switch self {
        case .jsonEncodingFailed:
            return NSLocalizedString("error_json_encoding_failed", comment: "JSON encoding failed error")
        case .jsonDecodingFailed:
            return NSLocalizedString("error_json_decoding_failed", comment: "JSON decoding failed error")
        case .utf8EncodingFailed:
            return NSLocalizedString("error_utf8_encoding_failed", comment: "UTF-8 encoding failed error")
        case .utf8DecodingFailed:
            return NSLocalizedString("error_utf8_decoding_failed", comment: "UTF-8 decoding failed error")
        case .imageEncodingFailed:
            return NSLocalizedString("error_image_encoding_failed", comment: "Image encoding failed error")
        case .imageDecodingFailed:
            return NSLocalizedString("error_image_decoding_failed", comment: "Image decoding failed error")
        }
    }
}
