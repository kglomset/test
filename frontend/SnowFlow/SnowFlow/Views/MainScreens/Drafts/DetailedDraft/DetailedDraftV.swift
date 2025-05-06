import SwiftUI

struct DetailedDraftView: View {
    @ObservedObject var vm: DetailedDraftVM
    @Environment(\.dismiss) private var dismiss
    
    var body: some View {
        VStack(spacing: Theme.Spacing.extra_large){
            Buttons(vm: vm)
                .padding(.horizontal, Theme.Spacing.medium)
            ScrollView{
                VStack(spacing: Theme.Spacing.extra_large){
                    Meta(vm: vm)
                    Products(vm: vm)
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
                NavButton(systemName: "chevron.left") {
                    if vm.hasUnsavedChanges() {
                        showConfirmationAlert = true
                    } else {
                        dismiss()
                    }
                }
                .alert("Unsaved changes", isPresented: $showConfirmationAlert) {
                    Button("Discard changes", role: .destructive) { dismiss() }
                    Button("Keep editing", role: .cancel) { }
                } message: {
                    Text("You have unsaved changes. Are you sure you want to leave without saving?")
                }
                
                // might be to close to the back button, can accidently click it when going back
                NavButton(text: "Start") { print("Start button tapped") }
                NavButton(text: "Save") { print("Save button tapped") }
            }
        }
    }
    
    
    // MARK: - Meta
    private struct Meta: View {
        @ObservedObject var vm: DetailedDraftVM
        
        var body: some View {
            VStack(spacing: Theme.Spacing.small){
                
                TextInput(
                    text: $vm.title,
                    label: "Title",
                    placeholder: "Enter title",
                    hasBorder: true,
                    // add max range, and also the possibility validate on unfocus
                    externalError: vm.titleError
                )
                
                HStack(alignment: .top, spacing: Theme.Spacing.small){
                    DatePickerView(selectedDate: $vm.date)
                    
                    SelectTextIcon(
                        title: "Weather",
                        items: weatherConditions,
                        placeholder: "select weather",
                        selectedOption: $vm.weatherIcon
                    )
                }
                
                HStack(alignment: .top, spacing: Theme.Spacing.small){
                    
                    NumberInput(
                        numberText: $vm.airTemperature,
                        label: "Air temperature",
                        placeholder: "Enter number",
                        hasBorder: true,
                        externalError: nil,
                        min: -40.0,
                        max: 20.0
                    )
                    
                    NumberInput(
                        numberText: $vm.airHumidity,
                        label: "Air humidity",
                        placeholder: "Enter number",
                        hasBorder: true,
                        externalError: nil,
                        min: 0.0,
                        max: 100.0
                    )
                    
                }
                
                HStack(alignment: .top, spacing: Theme.Spacing.small){
                    
                    SelectText(
                        title: "Snow type",
                        options: ["test1", "test2", "test3"],
                        placeholder: "Select",
                        selectedOption: $vm.snowType
                    )
                    
                    SelectText(
                        title: "Snow humidity",
                        options: ["test3", "test4", "test5"],
                        placeholder: "Select",
                        selectedOption: $vm.snowHardness
                    )
                    
                }
                
                HStack(alignment: .top, spacing: Theme.Spacing.small){
                    
                    NumberInput(
                        numberText: $vm.snowTemp,
                        label: "Snow temperature",
                        placeholder: "Enter number",
                        hasBorder: true,
                        externalError: nil, //$vm.snowTempError,
                        min: -40.0,
                        max: 20.0
                    )
                    
                    NumberInput(
                        numberText: $vm.snowHumidity,
                        label: "Snow humidity",
                        placeholder: "Enter number",
                        hasBorder: true,
                        externalError: nil, //$vm.snowHumidityError,
                        min: 0.0,
                        max: 100.0
                    )
                }
                
                TextInput(
                    text: $vm.comment,
                    label: "Comment",
                    placeholder: "Enter comment",
                    hasBorder: true,
                    // add max range, and also the possibility validate on unfocus
                    externalError: vm.commentError
                )
                
            }
        }
    }
    
    // MARK: - Products
    private struct Products: View {
        
        @ObservedObject var vm: DetailedDraftVM
        @State private var showEditDraftView = false
        @StateObject var productListVM: ProductListVM
        
        init(vm: DetailedDraftVM) {
            self.vm = vm
            _productListVM = StateObject(wrappedValue: ProductListVM(specificProducts: vm.products))
        }
        
        var body: some View {
            VStack {
                HStack {
                    VStack(alignment: .leading) {
                        Text("Total products: \(vm.productsCount)")
                            .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                            .fontWeight(.medium)
                        Text("Tests needed: \(vm.testCount)")
                            .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                            .fontWeight(.light)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    
                    SecondaryButton(title: "Edit", expand: false) {
                        showEditDraftView.toggle()
                        print(vm.products.count)
                    }
                    .fullScreenCover(isPresented: $showEditDraftView, onDismiss: {
                        productListVM.products = vm.products
                        productListVM.filterProducts()
                    }) {
                        EditDraftProductsView(vm: vm)
                    }
                }
                DetailProductListV(vm: productListVM)
            }
        }
    }
}

#Preview {
    DetailedDraftView(vm: DetailedDraftVM(id: 1))
}
