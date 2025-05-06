import SwiftUI

// MARK: - Product list view
struct ProductsView: View {
    
    @ObservedObject var viewModel: ProductsViewModel
    
    // track the currently selected product for full-screen presentation
    @State private var selectedProduct: ProductPreviewM? = nil
    
    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: Theme.Spacing.large) {
                    
                    // Search bar
                    TextInput(
                        text: $viewModel.searchValue,
                        label: "Search",
                        placeholder: "Search products...",
                        hasBorder: true,
                        externalError: nil
                    )
                    
                    VStack(alignment: .leading, spacing: Theme.Spacing.extra_small) {
                        
                        Text("Showing \(viewModel.filteredProducts.count) of \(viewModel.products.count) products")
                            .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                            .fontWeight(.light)
                        
                        // Product list
                        LazyVStack(spacing: 8) {
                            ForEach(viewModel.filteredProducts, id: \.productId) { productPreview in
                                ProductRowV(product: productPreview)
                                    .onTapGesture {
                                        selectedProduct = productPreview
                                    }
                            }
                        }
                    }
                }
                .padding(.horizontal, 8)
                .padding(.bottom, 8)
            }
            .frame(maxWidth: .infinity)
            .background(Color(Theme.Colors.backgroundGray))
            
            // for now it is ok to create a view-model each time,
            // but should consider to optimize in the future
            .fullScreenCover(item: $selectedProduct) { productToShow in
                DetailedProductView(productId: productToShow.productId)
            }
        }
    }
}

// MARK: - Preview
#Preview {
    ContentView(initialTab: .products)
}
