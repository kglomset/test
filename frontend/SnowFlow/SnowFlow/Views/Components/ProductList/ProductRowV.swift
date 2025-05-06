import SwiftUI

// MARK: - Product row view
struct ProductRowV: View, Equatable {
    let product: ProductPreviewM
    let status: ProductSelectionStatus?
    let onTap: ((ProductPreviewM) -> Void)?
    
    let cellHeight: CGFloat = 70 // not including padding and border
    let imageWidth: CGFloat = 56 // 4/5 aspect ratio
    
    static func == (lhs: ProductRowV, rhs: ProductRowV) -> Bool {
        return lhs.product.productId == rhs.product.productId && lhs.status == rhs.status
    }
    
    init(
        product: ProductPreviewM,
        status: ProductSelectionStatus? = nil,
        onTap: ((ProductPreviewM) -> Void)? = nil
    ) {
        self.product = product
        self.status = status
        self.onTap = onTap
    }
    
    var body: some View {
        HStack(spacing: Theme.Spacing.small) {
            
            // product image
            Group {
                if let imageUrl = product.imageUrl,
                   let data = try? Data(contentsOf: imageUrl),
                   let uiImage = UIImage(data: data) {
                    Image(uiImage: uiImage)
                        .resizable()
                } else {
                    Image(systemName: "photo")
                        .resizable()
                        .foregroundColor(Theme.Colors.placeholder)
                }
            }
            .aspectRatio(contentMode: .fill)
            .frame(width: imageWidth)
            .cornerRadius(Theme.CornerRadius.medium)
            
            // meta info
            VStack(alignment: .leading, spacing: 0) {
                HStack(alignment: .top, spacing: 0) {
                    Text(product.name)
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                        .fontWeight(.medium)
                        .lineLimit(2)
                        .truncationMode(.tail)
                        .frame(maxWidth: .infinity, alignment: .leading)
                    
                    Text(product.ean ?? "")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                        .fontWeight(.light)
                        .lineLimit(1)
                        .truncationMode(.tail)
                }
                
                Spacer()
                
                HStack {
                    Text(product.brand ?? "")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                        .fontWeight(.light)
                        .lineLimit(1)
                        .truncationMode(.tail)
                }
            }
        }
        .frame(height: cellHeight)
        .padding(Theme.Spacing.small)
        .background(Theme.Colors.backgroundLight)
        .cornerRadius(Theme.CornerRadius.medium)
        .overlay(selectionOverlay)
        .padding(2) // account for border overlay
        .opacity(status == .unavailable ? 0.5 : 1.0)
        //.contentShape(Rectangle()) // unsure if needed - supposed to ensure tapable area
        .onTapGesture {
            if let onTap = onTap {
                onTap(product)
            }
        }
    }
    
    @ViewBuilder
    private var selectionOverlay: some View {
        if let status = status {
            RoundedRectangle(cornerRadius: Theme.CornerRadius.medium)
                .stroke(borderColor(for: status), lineWidth: 2)
        }
    }
    
    private func borderColor(for status: ProductSelectionStatus) -> Color {
        switch status {
        case .selectable: return Theme.Colors.backgroundLight
        case .selected: return Theme.Colors.selected
        case .unavailable: return Theme.Colors.error
        }
    }
}

// MARK: - Preview
#Preview {
    @Previewable @State var selectedProductId: Int? = nil
    @Previewable @State var selectedProduct: ProductPreviewM? = nil
    @Previewable @State var selectedProductIds: [Int] = []
    
    let products = [
        PredefinedProductsPreview.getProductPreviewById(id: 1) ?? ProductPreviewM(productId: -1, name: "not found"),
        PredefinedProductsPreview.getProductPreviewById(id: 2) ?? ProductPreviewM(productId: -1, name: "not found"),
        PredefinedProductsPreview.getProductPreviewById(id: 3) ?? ProductPreviewM(productId: -1, name: "not found")
    ]
    
    ScrollView{
        VStack {
            Text("Normal rows").padding(.top, 24)
            LazyVStack {
                ProductRowV(product: products[0])
                ProductRowV(product: products[1])
                ProductRowV(product: products[2])
            }
            
            Text("Normal rows - detail view").padding(.top, 24)
            LazyVStack {
                ForEach(products, id: \.productId) { product in
                    ProductRowV(
                        product: product,
                        onTap: { tappedProduct in
                            selectedProduct = tappedProduct
                        }
                    )
                }
            }
            .fullScreenCover(item: $selectedProduct, onDismiss: {
                selectedProduct = nil
            }) { productToShow in
                DetailedProductView(productId: productToShow.productId)
            }
            
            Text("Selection rows").padding(.top, 24)
            LazyVStack {
                ProductRowV(product: products[0], status: .selectable)
                ProductRowV(product: products[1], status: .selected)
                ProductRowV(product: products[2], status: .unavailable)
            }
            
            Text("Selection rows - selectable").padding(.top, 24)
            LazyVStack {
                ForEach(products, id: \.productId) { product in
                    
                    // determine status based on selection
                    let status: ProductSelectionStatus =
                    product.productId == 2 ? .unavailable :
                    product.productId == selectedProductId ? .selected : .selectable
                    
                    ProductRowV(
                        product: product,
                        status: status,
                        onTap: { tappedProduct in
                            
                            if status == .unavailable {
                                return
                            }
                            
                            if selectedProductId == tappedProduct.productId {
                                selectedProductId = nil
                            } else {
                                selectedProductId = tappedProduct.productId
                            }
                        }
                    )
                }
            }
            
            Text("Selection rows - multi selectable").padding(.top, 24)
            LazyVStack {
                ForEach(products, id: \.productId) { product in
                    
                    // determine status based on selection
                    let status: ProductSelectionStatus =
                    product.productId == 2 ? .unavailable :
                    selectedProductIds.contains(product.productId) ? .selected : .selectable
                    
                    ProductRowV(
                        product: product,
                        status: status,
                        onTap: { tappedProduct in
                            
                            if status == .unavailable {
                                return
                            }
                            
                            if selectedProductIds.contains(tappedProduct.productId) {
                                selectedProductIds.removeAll(where: { $0 == tappedProduct.productId })
                            } else {
                                selectedProductIds.append(tappedProduct.productId)
                            }
                        }
                    )
                }
            }
        }
        .frame(maxHeight: .infinity)
        .padding()
        .background(Theme.Colors.backgroundGray)
    }
}
