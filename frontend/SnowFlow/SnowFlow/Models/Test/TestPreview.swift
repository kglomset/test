import Foundation

struct TestPreview: Identifiable {
    let id: UUID
    let title: String
    let date: String
    let productCount: Int
    let temperature: String
    let location: String
    let weatherIcon: String
    let isPrivate: Bool
    
    init(
        id: UUID = UUID(),
        title: String,
        date: String,
        productCount: Int,
        temperature: String,
        location: String,
        weatherIcon: String,
        isPrivate: Bool
    ) {
        self.id = id
        self.title = title
        self.date = date
        self.productCount = productCount
        self.temperature = temperature
        self.location = location
        self.weatherIcon = weatherIcon
        self.isPrivate = isPrivate
    }
}

