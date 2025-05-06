import SwiftUI

// MARK: - Core list component
fileprivate struct CoreProductListV<RowContent: View>: View {
    @ObservedObject var vm: ProductListVM
    let showCount: Bool
    let rowContent: (ProductPreviewM) -> RowContent
    
    init(
        vm: ProductListVM,
        showCount: Bool = false,
        @ViewBuilder rowContent: @escaping (ProductPreviewM) -> RowContent
    ) {
        self.vm = vm
        self.showCount = showCount
        self.rowContent = rowContent
    }
    
    var body: some View {
        VStack(alignment: .leading, spacing: Theme.Spacing.extra_small) {
            if showCount {
                Text("Showing \(vm.filteredProducts.count) of \(vm.products.count) products")
                    .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                    .fontWeight(.light)
            }
            
            LazyVStack {
                ForEach(vm.filteredProducts, id: \.productId) { preview in
                    rowContent(preview)
                }
            }
        }
    }
}

// MARK: - Standard product list
struct ProductListV: View {
    @ObservedObject var vm: ProductListVM
    let showCount: Bool
    
    init(
        vm: ProductListVM,
        showCount: Bool = false
    ) {
        self.vm = vm
        self.showCount = showCount
    }
    
    var body: some View {
        CoreProductListV(vm: vm, showCount: showCount) { product in
            ProductRowV(product: product)
        }
    }
}

// MARK: - Detail view product List
struct DetailProductListV: View {
    @ObservedObject var vm: ProductListVM // why cannot this be a state object?
    let showCount: Bool
    
    @State private var selectedProduct: ProductPreviewM?
    
    init(
        vm: ProductListVM,
        showCount: Bool = false
    ) {
        self.vm = vm
        self.showCount = showCount
    }
    
    var body: some View {
        CoreProductListV(vm: vm, showCount: showCount) { product in
            ProductRowV(
                product: product,
                onTap: { tappedProduct in
                    selectedProduct = tappedProduct
                }
            )
        }
        .fullScreenCover(item: $selectedProduct, onDismiss: {
            selectedProduct = nil
        }) { productToShow in
            DetailedProductView(productId: productToShow.productId)
        }
    }
}

// MARK: - Single selection product list
struct SingleSelectionProductListV: View {
    @ObservedObject var vm: ProductListVM
    @Binding var selectedPreview: ProductPreviewM?
    let unavailableProductIds: Set<Int>
    let showCount: Bool
    
    init(
        vm: ProductListVM,
        selectedPreview: Binding<ProductPreviewM?>,
        unavailableProductIds: Set<Int> = [],
        showCount: Bool = false
    ) {
        self.vm = vm
        self._selectedPreview = selectedPreview
        self.unavailableProductIds = unavailableProductIds
        self.showCount = showCount
    }
    
    var body: some View {
        CoreProductListV(vm: vm, showCount: showCount) { product in
            let status = statusFor(product: product)
            
            ProductRowV(
                product: product,
                status: status,
                onTap: { tappedProduct in
                    if status == .unavailable {
                        return
                    }
                    
                    if selectedPreview?.productId == tappedProduct.productId {
                        selectedPreview = nil
                    } else {
                        selectedPreview = tappedProduct
                    }
                }
            )
        }
    }
    
    private func statusFor(product: ProductPreviewM) -> ProductSelectionStatus {
        if unavailableProductIds.contains(product.productId) {
            return .unavailable
        }
        return product.productId == selectedPreview?.productId ? .selected : .selectable
    }
}

// MARK: - Multi selection product list
struct MultiSelectionProductListV: View {
    @ObservedObject var vm: ProductListVM
    @Binding var selectedIds: [Int]
    let unavailableProductIds: Set<Int>
    let showCount: Bool
    
    init(
        vm: ProductListVM,
        selectedIds: Binding<[Int]>,
        unavailableProductIds: Set<Int> = [],
        showCount: Bool = false
    ) {
        self.vm = vm
        self._selectedIds = selectedIds
        self.unavailableProductIds = unavailableProductIds
        self.showCount = showCount
    }
    
    var body: some View {
        CoreProductListV(vm: vm, showCount: showCount) { product in
            let status = statusFor(product: product)
            
            ProductRowV(
                product: product,
                status: status,
                onTap: { tappedProduct in
                    if status == .unavailable {
                        return
                    }
                    
                    if selectedIds.contains(tappedProduct.productId) {
                        selectedIds.removeAll(where: { $0 == tappedProduct.productId })
                    } else {
                        selectedIds.append(tappedProduct.productId)
                    }
                }
            )
        }
    }
    
    private func statusFor(product: ProductPreviewM) -> ProductSelectionStatus {
        if unavailableProductIds.contains(product.productId) {
            return .unavailable
        }
        return selectedIds.contains(product.productId) ? .selected : .selectable
    }
}

// MARK: - Preview - standard
#Preview {
    @Previewable @StateObject var vm = ProductListVM()
    
    VStack {
        TextInput(
            text: $vm.searchValue,
            label: nil,
            placeholder: "Search products...",
            hasBorder: false,
            externalError: nil
        )
        
        ScrollView {
            ProductListV(vm: vm)
        }
    }
    .padding()
    .background(Color(Theme.Colors.backgroundGray))
}

// MARK: - Preview - detailed
#Preview {
    @Previewable @StateObject var vm = ProductListVM()
    
    VStack {
        TextInput(
            text: $vm.searchValue,
            label: nil,
            placeholder: "Search products...",
            hasBorder: false,
            externalError: nil
        )
        
        ScrollView {
            DetailProductListV(vm: vm, showCount: true)
        }
    }
    .padding()
    .background(Color(Theme.Colors.backgroundGray))
}

// MARK: - Preview - single selection
#Preview {
    @Previewable @StateObject var vm = ProductListVM()
    @Previewable @State var selectedProduct: ProductPreviewM? = nil
    let unavailableProductIds: Set<Int> = [1, 2, 3]
    
    VStack {
        TextInput(
            text: $vm.searchValue,
            label: nil,
            placeholder: "Search products...",
            hasBorder: false,
            externalError: nil
        )
        
        ScrollView {
            SingleSelectionProductListV(
                vm: vm,
                selectedPreview: $selectedProduct,
                unavailableProductIds: unavailableProductIds,
                showCount: true
            )
        }
    }
    .padding()
    .background(Color(Theme.Colors.backgroundGray))
}

// MARK: - Preview - multi selection
#Preview {
    @Previewable @StateObject var vm = ProductListVM()
    @Previewable @State var selectedProductIds: [Int] = []
    let unavailableProductIds: Set<Int> = [4, 5, 6]
    
    VStack {
        TextInput(
            text: $vm.searchValue,
            label: nil,
            placeholder: "Search products...",
            hasBorder: false,
            externalError: nil
        )
        
        ScrollView {
            MultiSelectionProductListV(
                vm: vm,
                selectedIds: $selectedProductIds,
                unavailableProductIds: unavailableProductIds,
                showCount: true
            )
        }
    }
    .padding()
    .background(Color(Theme.Colors.backgroundGray))
}
