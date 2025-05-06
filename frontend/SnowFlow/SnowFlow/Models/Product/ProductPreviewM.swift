import Foundation

// MARK: - Product preview model
struct ProductPreviewM: Codable, Identifiable {
    let id: UUID
    let productId: Int
    let name: String
    let ean: String?
    let brand: String?
    let warmTemp: Double?
    let coldTemp: Double?
    let type: String?
    let imageUrl: URL?
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
        self.isPrivate = isPrivate
    }
}
