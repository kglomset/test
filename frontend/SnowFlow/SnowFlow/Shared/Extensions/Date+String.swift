import Foundation

extension Date {
    /// Format a date to "MMM d, yyyy"
    func toString(format: String = "MMM d, yyyy") -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = format
        return formatter.string(from: self)
    }

    /// Format a timestamp for logging "HH:mm:ss.SSS"
    func toLogTimeString() -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm:ss.SSS"
        return formatter.string(from: self)
    }
    
    /// ISO8601 formatter for consistency across the app
    private static let iso8601Formatter: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter
    }()

    /// Convert an ISO8601 string to a Date
    static func dateFromISO8601String(_ string: String) -> Date? {
        return iso8601Formatter.date(from: string)
    }

    /// Convert a Date to an ISO8601 string
    static func iso8601StringFromDate(_ date: Date) -> String {
        return iso8601Formatter.string(from: date)
    }
}
