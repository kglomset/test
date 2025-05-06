import Foundation

struct Test: Identifiable {
    let id: UUID
    let title: String
    let date: String
    let productCount: Int
    let temperature: String
    let location: String
    let weatherIcon: String
    let isPrivate: Bool
    let comment: String?
    let tournement: Tournament?

    init(
        id: UUID = UUID(),
        title: String,
        date: String,
        productCount: Int,
        temperature: String,
        location: String,
        weatherIcon: String,
        isPrivate: Bool,
        comment: String? = nil,
        tournement: Tournament? = nil
    ) {
        self.id = id
        self.title = title
        self.date = date
        self.productCount = productCount
        self.temperature = temperature
        self.location = location
        self.weatherIcon = weatherIcon
        self.isPrivate = isPrivate
        self.comment = comment
        self.tournement = tournement
    }
}
