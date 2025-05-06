import SwiftUI

struct EditDraftProductsView: View {
    @ObservedObject var vm: DetailedDraftVM
    @Environment(\.dismiss) private var dismiss
    
    var body: some View {
        VStack(spacing: Theme.Spacing.extra_large){
            Buttons(vm: vm)
                .padding(.horizontal, Theme.Spacing.medium)
            ScrollView{
                VStack(spacing: Theme.Spacing.extra_large){
                    Meta(vm: vm)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
                .padding(.horizontal, Theme.Spacing.medium)
            }
        }
        .background(Theme.Colors.backgroundGray)
    }
    
    // MARK: - Buttons
    private struct Buttons: View {
        @State private var showConfirmationAlert = false
        @ObservedObject var vm: DetailedDraftVM
        @Environment(\.dismiss) private var dismiss
        
        var body: some View {
            HStack {
                NavButton(systemName: "chevron.left") {dismiss()}
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }
    
    
    // MARK: - Meta
    private struct Meta: View {
        @ObservedObject var vm: DetailedDraftVM
        
        var body: some View {
            VStack(spacing: Theme.Spacing.small){
                
                VStack(alignment: .leading){
                    Text("Total products: \(vm.productsCount)")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                        .fontWeight(.medium)
                    Text("Tests needed: \(vm.testCount)")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                        .fontWeight(.light)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                
                
                VStack {
                    
                    ForEach(vm.samples.indices, id: \.self) { index in
                        ChooseSkiProduct(
                            label: "Ski \(index + 1)",
                            skiName: $vm.samples[index].skiName,
                            productPreview: $vm.samples[index].productPreview,
                            //alreadySelectedProducts: vm.products,
                            
                            deleteAction: {
                                vm.removeSample(vm.samples[index])
                            }
                        )
                    }
                    
                    SecondaryButton(
                        title: "Add product", expand: true, action: {vm.addSample()})
                    .padding(.top, Theme.Spacing.medium)
                }
            }
        }
        
        
        private struct ChooseSkiProduct: View {
            
            let label: String
            @Binding var skiName: String?
            
            // ugly ass
            var skiNameBinding: Binding<String> {
                Binding(
                    get: { skiName ?? "" },  // default value when nil
                    set: { skiName = $0 }    // allow updates
                )
            }
            
            @Binding var productPreview: ProductPreviewM?
            //var alreadySelectedProducts: [Int]
            
            let deleteAction: () -> Void
            
            var body: some View {
                VStack(spacing: Theme.Spacing.small){
                    
                    HStack(alignment: .bottom, spacing: Theme.Spacing.small){
                        
                        TextInput(
                            text: skiNameBinding,
                            label: label,
                            placeholder: "Enter name",
                            hasBorder: true,
                            externalError: nil // ski name cant be blank and must be unique within test
                        )
                        
                        ChooseProductButton(label: "Choose product", selectedPreview: $productPreview, unavailableProductIds: [])
                        
                        Button(action: deleteAction) {
                            Image(systemName: "trash")
                                .foregroundColor(.red)
                                .padding(8)
                        }
                    }
                }
            }
        }
    }
}

#Preview {
    EditDraftProductsView(vm: DetailedDraftVM(id: 1))
}
