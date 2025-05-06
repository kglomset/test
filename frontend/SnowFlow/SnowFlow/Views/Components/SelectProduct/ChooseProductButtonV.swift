import SwiftUI

struct ChooseProductButton: View {
    
    let label: String?
    @Binding var selectedPreview: ProductPreviewM?
    @State var unavailableProductIds: Set<Int>
    
    @State private var isChoosingProduct = false
    
    init(
        label: String? = nil,
        selectedPreview: Binding<ProductPreviewM?>,
        unavailableProductIds: Set<Int> = []
    ) {
        self.label = label
        self._selectedPreview = selectedPreview
        self._unavailableProductIds = State(initialValue: unavailableProductIds)
    }
    
    var body: some View {
        Button(action: { isChoosingProduct.toggle() }) {
            VStack(alignment: .leading, spacing: 2) {
                if let label = label, !label.isEmpty {
                    Text(label)
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                        .fontWeight(.medium)
                }
                HStack {
                    Text(selectedPreview?.name ?? "Choose product")
                        .foregroundStyle(selectedPreview == nil ? Theme.Colors.placeholder : .primary)
                }
                .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                .foregroundColor(Theme.Colors.primary)
                .padding(.horizontal, 16)
                .frame(height: 38)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Theme.Colors.backgroundLight)
                .cornerRadius(Theme.CornerRadius.medium)
                .overlay(
                    RoundedRectangle(cornerRadius: Theme.CornerRadius.medium)
                        .stroke(Theme.Colors.border, lineWidth: 1)
                )
            }
            .fullScreenCover(isPresented: $isChoosingProduct) {
                ChooseProductV(selectedPreview: $selectedPreview, unavailableProductIds: $unavailableProductIds)
            }
        }
        .buttonStyle(PlainButtonStyle())
    }
}

// MARK: - Preview
#Preview {
    @Previewable @State var selectedProduct1: ProductPreviewM? = nil
    @Previewable @State var selectedProduct2: ProductPreviewM? = PredefinedProductsPreview.getProductPreviewById(id: 1) ?? nil
    @Previewable @State var selectedProduct3: ProductPreviewM? = PredefinedProductsPreview.getProductPreviewById(id: 2) ?? nil
    
    let unavailableProductIds: Set<Int> = [4, 5, 6]
    
    VStack {
        ChooseProductButton(selectedPreview: $selectedProduct1, unavailableProductIds: unavailableProductIds)
        ChooseProductButton(selectedPreview: $selectedProduct2, unavailableProductIds: unavailableProductIds)
        ChooseProductButton(label: "Select product", selectedPreview: $selectedProduct3)
    }
    .padding()
    .background(Theme.Colors.backgroundGray)
}
