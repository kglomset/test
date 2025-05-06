import SwiftUI

// MARK: - Detailed products view model {
class DetailedProductsViewModel: ObservableObject {
    
    @Published private(set) var product: Product
    @Published var productId: Int {
        didSet { product = PredefinedProducts.getProduct(by: productId) ?? Self.defaultProduct }
    }
    
    lazy var tests: [TestPreview] = PredefinedTestPreview.getAllTestPreviews()
    
    var hasBuyUrl: Bool {
        product.buyUrl != nil
    }
    
    var productImage: Image {
        //Image(product?.imageUrl ?? "sg10")
        Image("sg10")
    }
    
    init(id: Int) {
        self.productId = id
        self.product = PredefinedProducts.getProduct(by: id) ?? Self.defaultProduct
    }
    
    private static let defaultProduct = Product(productId: -1, name: "Failed to get product")
}
