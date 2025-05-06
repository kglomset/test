import Foundation
import UIKit
import OSLog

private let logger = Logger(subsystem: "ntnu.stud.SnowFlow.codec", category: "codecManager")

enum EncodeType {
    case none
    case json(Encodable)
    case plainText(String)
}

class CodecManager {
    
    // MARK: - request encoding
    /// encodes request content into appropriate data format with content type
    /// - Parameter EncodeType: the content to encode (none, json, or plain text)
    /// - Returns: tuple containing encoded data and content type header
    /// - Throws: a codec error if encoding fails
    static func encode(_ requestContent: EncodeType) throws -> (data: Data, contentType: String)? {
        switch requestContent {
        case .none:
            return nil
        case .json(let encodableObj):
            do {
                let jsonData = try JSONEncoder().encode(encodableObj)
                return (jsonData, "application/json")
            } catch {
                logger.error("json encoding failed: \(error.localizedDescription)")
                throw CodecError.jsonEncodingFailed
            }
        case .plainText(let text):
            guard let textData = text.data(using: .utf8) else {
                logger.error("utf-8 encoding failed: unable to encode text")
                throw CodecError.utf8EncodingFailed
            }
            return (textData, "text/plain")
        }
    }
    
    // MARK: - response decoding
    /// decodes response data into the specified type
    /// - Parameter data: raw data from network response
    /// - Returns: the decoded object of type T
    /// - Throws: a codec error if decoding fails
    static func decode<T: Decodable>(_ data: Data) throws -> T {
        // handle data type
        if T.self == Data.self, let result = data as? T {
            return result
        }
        
        // handle string type
        if T.self == String.self {
            if let result = String(data: data, encoding: .utf8) as? T {
                return result
            } else {
                logger.error("utf-8 decoding failed: unable to convert data to string")
                throw CodecError.utf8DecodingFailed
            }
        }
        
        // handle json decodable types - default case
        do {
            let decodedObject = try JSONDecoder().decode(T.self, from: data)
            return decodedObject
        } catch {
            logger.error("json decoding failed: \(error.localizedDescription)")
            throw CodecError.jsonDecodingFailed
        }
    }
    
    /// decodes image from response data
    /// - Parameter data: raw data from network response
    /// - Returns: the decoded uiimage
    /// - Throws: a codec error if image decoding fails
    static func decodeImage(from data: Data) throws -> UIImage {
        if let image = UIImage(data: data) {
            return image
        } else {
            logger.error("image decoding failed: data could not be converted to uiimage")
            throw CodecError.imageDecodingFailed
        }
    }
}
