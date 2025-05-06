import Foundation

// Product and product preview should be different to optimize for memory.
// In lists a minimal product is needed, but in detailed products the view-model should fetch more data.
// Caching must be considered.

// MARK: - Product model
struct Product: Codable, Identifiable {
    let id: UUID
    let productId: Int
    let name: String
    let ean: String?
    let brand: String?
    let warmTemp: Double?
    let coldTemp: Double?
    let type: String?
    let imageUrl: URL?
    let buyUrl: URL?
    let comment: String?
    let isOwner: Bool
    let isPrivate: Bool

    init(
        id: UUID = UUID(),
        productId: Int,
        name: String,
        ean: String? = nil,
        brand: String? = nil,
        warmTemp: Double? = nil,
        coldTemp: Double? = nil,
        type: String? = nil,
        imageUrl: URL? = nil,
        buyUrl: URL? = nil,
        comment: String? = nil,
        isOwner: Bool = false,
        isPrivate: Bool = false
    ) {
        self.id = id
        self.productId = productId
        self.name = name
        self.ean = ean
        self.brand = brand
        self.warmTemp = warmTemp
        self.coldTemp = coldTemp
        self.type = type
        self.imageUrl = imageUrl
        self.buyUrl = buyUrl
        self.comment = comment
        self.isOwner = isOwner
        self.isPrivate = isPrivate
    }
}
