import Foundation

struct SampleM: Identifiable {
    let id = UUID()
    var skiName: String?
    var productPreview: ProductPreviewM?
    
    var isValid: Bool {
        skiName != nil && productPreview?.productId != nil
    }
}
