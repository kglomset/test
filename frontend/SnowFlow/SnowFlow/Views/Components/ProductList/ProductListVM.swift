import Foundation
import Combine

// MARK: - ProductListVM
class ProductListVM: ObservableObject {
    @Published var products: [ProductPreviewM] = []
    @Published var searchValue: String = ""
    @Published var selectedProductPreview: ProductPreviewM? = nil
    @Published var filteredProducts: [ProductPreviewM] = []
    
    private var cancellables = Set<AnyCancellable>()
    private let unselectableProductIds: [Int]
    private let specificProductIds: [Int]?
    private let specificProducts: [ProductPreviewM]?
    
    // Init with specific product IDs
    init(
        selectedProductPreview: ProductPreviewM? = nil,
        unselectableProductIds: [Int] = [],
        specificProductIds: [Int]? = nil
    ) {
        self.selectedProductPreview = selectedProductPreview
        self.unselectableProductIds = unselectableProductIds
        self.specificProductIds = specificProductIds
        self.specificProducts = nil
        loadProducts()
        filterProducts()
        
        setupSearchPublisher()
    }
    
    // Init with specific product previews
    init(
        selectedProductPreview: ProductPreviewM? = nil,
        unselectableProductIds: [Int] = [],
        specificProducts: [ProductPreviewM]
    ) {
        self.selectedProductPreview = selectedProductPreview
        self.unselectableProductIds = unselectableProductIds
        self.specificProductIds = nil
        self.specificProducts = specificProducts
        loadProducts()
        filterProducts()
        
        setupSearchPublisher()
    }
    
    // Setup the search publisher (extracted to avoid duplication)
    private func setupSearchPublisher() {
        $searchValue
            .dropFirst()
            .debounce(for: .milliseconds(180), scheduler: RunLoop.main)
            .sink { [weak self] _ in self?.filterProducts() }
            .store(in: &cancellables)
    }
    
    // methods
    private func loadProducts() {
        if let specificProducts = specificProducts {
            // Use provided product previews directly
            self.products = specificProducts
        } else {
            // Load all products
            var allProducts = PredefinedProductsPreview.getAllProductPreviews()
            
            // If specific product IDs are provided, filter the products
            if let specificIds = specificProductIds, !specificIds.isEmpty {
                allProducts = allProducts.filter { specificIds.contains($0.productId) }
            }
            
            self.products = allProducts
        }
    }
    
    func filterProducts() {
        filteredProducts = searchValue.isEmpty ? products : products.filter {
            $0.name.localizedCaseInsensitiveContains(searchValue) ||
            ($0.brand ?? "").localizedCaseInsensitiveContains(searchValue) ||
            ($0.type ?? "").localizedCaseInsensitiveContains(searchValue) ||
            ($0.ean ?? "").contains(searchValue)
        }
    }
}
