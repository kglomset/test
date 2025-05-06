import SwiftUI

// MARK: - Choose product view
struct ChooseProductV: View {
    
    @StateObject var vm = ProductListVM()
    @Binding var selectedPreview: ProductPreviewM?
    @Binding var unavailableProductIds: Set<Int>
    
    @Environment(\.dismiss) private var dismiss
    
    init(
        selectedPreview: Binding<ProductPreviewM?> = .constant(nil),
        unavailableProductIds: Binding<Set<Int>> = .constant([])
    ) {
        _selectedPreview = selectedPreview
        _unavailableProductIds = unavailableProductIds
    }
    
    var body: some View {
        VStack(spacing: Theme.Spacing.extra_large){
            HStack {
                NavButton(systemName: "chevron.left") { dismiss() }
                
                // search bar
                TextInput(
                    text: $vm.searchValue,
                    label: nil,
                    placeholder: "Search products...",
                    hasBorder: false,
                    externalError: nil
                )
            }
            .padding(.horizontal, Theme.Spacing.medium)
            
            ScrollView {
                SingleSelectionProductListV(
                    vm: vm,
                    selectedPreview: $selectedPreview,
                    unavailableProductIds: unavailableProductIds,
                    showCount: true
                )
                .padding(.horizontal, Theme.Spacing.medium)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .background(Color(Theme.Colors.backgroundGray))
    }
}

// MARK: - Preview
#Preview {
    @Previewable @State var selectedPreview: ProductPreviewM? = nil
    @Previewable @State var unavailableProductIds: Set<Int> = [4, 5, 6]
    
    ChooseProductV(selectedPreview: $selectedPreview, unavailableProductIds: $unavailableProductIds)
}
