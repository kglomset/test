import Foundation
import Combine

// MARK: - View model
class ProductsViewModel: ObservableObject {
    @Published var products: [ProductPreviewM] = []
    @Published var searchValue: String = ""
    
    var filteredProducts: [ProductPreviewM] {
        guard !searchValue.isEmpty else { return products }
        return products.filter { product in
            product.name.localizedCaseInsensitiveContains(searchValue) ||
            (product.brand ?? "").localizedCaseInsensitiveContains(searchValue) ||
            (product.type ?? "").localizedCaseInsensitiveContains(searchValue) ||
            (product.ean ?? "").contains(searchValue)
        }
    }
    
    init() {
        loadSampleData()
    }
    
    private func loadSampleData() {
        self.products = PredefinedProductsPreview.getAllProductPreviews()
    }
}
