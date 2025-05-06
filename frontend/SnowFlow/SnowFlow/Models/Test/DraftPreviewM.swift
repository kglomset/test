import Foundation

struct DraftPreviewM: Identifiable {
    let id: UUID
    let draftId: Int
    let title: String
    let date: Date?
    let productCount: Int
    let temperature: String?
    let location: String?
    let weatherIcon: String?
    let isPrivate: Bool

    init(
        id: UUID = UUID(),
        draftId: Int,
        title: String = "Untitled draft",
        date: Date? = nil,
        productCount: Int = 0,
        temperature: String? = nil,
        location: String? = nil,
        weatherIcon: String? = nil,
        isPrivate: Bool = false
    ) {
        self.id = id
        self.draftId = draftId
        self.title = title
        self.date = date
        self.productCount = productCount
        self.temperature = temperature
        self.location = location
        self.weatherIcon = weatherIcon
        self.isPrivate = isPrivate
    }
}
