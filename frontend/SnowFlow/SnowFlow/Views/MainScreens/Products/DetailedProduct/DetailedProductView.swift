import SwiftUI

// MARK: - Detailed product view
struct DetailedProductView: View {
    @StateObject var vm: DetailedProductsViewModel
    
    // dismiss environment variable to close the view
    @Environment(\.dismiss) private var dismiss
    
    init(productId: Int) {
        _vm = StateObject(wrappedValue: DetailedProductsViewModel(id: productId))
    }
    
    var body: some View {
        VStack(spacing: 0) {
            // the header
            ProductHeaderView(vm: vm)
            
            // scrollable information
            ZStack(alignment: .top) {
                // gradient background at the top
                LinearGradient(
                    gradient: Gradient(colors: [
                        Color(Theme.Colors.secondary).opacity(0.08),
                        Color(Theme.Colors.secondary).opacity(0.0)
                    ]),
                    startPoint: .top,
                    endPoint: .bottom
                )
                .frame(height: Theme.Spacing.extra_large)
                
                // actual information
                DetailedInfoSection(vm: vm)
            }
            .background(Theme.Colors.backgroundGray)
        }
        .ignoresSafeArea(.all)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
    }
    
    // MARK: - Product header
    struct ProductHeaderView: View {
        
        @ObservedObject var vm: DetailedProductsViewModel
        @Environment(\.dismiss) private var dismiss
        
        let imageWidth: CGFloat = 124
        
        var body: some View {
            ZStack(alignment: .bottomTrailing) {
                VStack(spacing: 0) {
                    ZStack(alignment: .bottomLeading) {
                        Image("tracks_mountain_ole")
                            .resizable()
                            .scaledToFill()
                            .offset(y: 148)
                            .frame(height: 212, alignment: .bottom)
                            .clipped()
                        
                        LinearGradient(
                            gradient: Gradient(colors: [
                                Color(hex: "#3B434F").opacity(0.0),
                                Color(hex: "#3B434F").opacity(0.8)
                            ]),
                            startPoint: .top,
                            endPoint: .bottom
                        )
                        .frame(height: 136)
                        
                        HStack(alignment: .bottom, spacing: 2){
                            
                            if vm.product.isPrivate {
                                Image(systemName: "lock")
                                    .font(.system(size: Theme.Fonts.bodySize, weight: .regular, design: .monospaced))
                                    .fontWeight(.medium)
                                    .padding(.bottom, 8)
                            }
                            
                            FlexBoxView(text: vm.product.name)
                                .padding(.trailing, imageWidth + Theme.Spacing.medium)
                                .padding(.bottom, Theme.Spacing.extra_small)
                            //MinimalText(vm.product.name, size: 32, font: .primary, lineSpacing: -14).padding(.trailing, imageWidth + Theme.Spacing.medium)
                            
                        }
                        .padding(.leading, Theme.Spacing.medium)
                        .foregroundColor(Theme.Colors.white)
                    }
                    
                    HStack(spacing: 12) {
                        HeaderButtonView(systemName: "chevron.left") { dismiss() }
                        
                        if vm.product.isOwner {
                            HeaderButtonView(systemName: "square.and.pencil") { print("Edit button tapped") }
                        }
                        
                        Spacer()
                        
                        if vm.hasBuyUrl {
                            HeaderButtonView(systemName: "cart") { print("Cart button tapped") }
                        }
                    }
                    .padding(Theme.Spacing.medium)
                    .padding(.trailing, Theme.Spacing.medium + imageWidth)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Theme.Colors.backgroundGray)
                }
                
                vm.productImage
                    .resizable()
                    .scaledToFit()
                    .frame(width: imageWidth)
                    .cornerRadius(Theme.CornerRadius.large)
                    .frame(maxWidth: .infinity, alignment: .bottomTrailing)
                    .padding(Theme.Spacing.medium)
                    .applyDropShadow()
            }
        }
        
        // MARK: - Header button view
        struct HeaderButtonView: View {
            let systemName: String
            let action: () -> Void
            
            var body: some View {
                Button(action: action) {
                    Image(systemName: systemName)
                        .font(.system(size: 18, weight: .regular, design: .monospaced))
                }
                .frame(width: 50, height: 38)
                .background(Theme.Colors.backgroundLight)
                .cornerRadius(Theme.CornerRadius.medium)
                .applyDropShadow()
            }
        }
    }
    
    
    // MARK: - Detailed info view
    struct DetailedInfoSection: View {
        
        @ObservedObject var vm: DetailedProductsViewModel
        
        var body: some View {
            ScrollView {
                VStack(spacing: Theme.Spacing.extra_large) {
                    ProductAttributesView(vm: vm)
                        .applyDropShadow()
                    
                    Text(vm.product.comment ?? "")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                    
                    ProductStatsView()
                    PreviousTestsListView(vm: vm)
                }
                .padding(.horizontal, Theme.Spacing.medium)
                .padding(.vertical, Theme.Spacing.large)
            }
        }
        
        // MARK: - Product attributes
        struct ProductAttributesView: View {
            
            @ObservedObject var vm: DetailedProductsViewModel
            
            var body: some View {
                VStack(alignment: .leading, spacing: Theme.Spacing.extra_small) {
                    Text("Info")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.smallHeadlineSize))
                        .fontWeight(.semibold)
                    
                    HStack {
                        AttributeCellView(title: "Brand", value: vm.product.brand ?? "-")
                            .frame(maxWidth: .infinity)
                        
                        VerticalDividerView()
                        
                        AttributeCellView(title: "Warm", value: vm.product.warmTemp.map { "\($0)°C" } ?? "-")
                            .frame(maxWidth: .infinity)
                        
                        VerticalDividerView()
                        
                        AttributeCellView(title: "Cold", value: vm.product.coldTemp.map { "\($0)°C" } ?? "-")
                            .frame(maxWidth: .infinity)
                        
                        VerticalDividerView()
                        
                        AttributeCellView(title: "Type", value: vm.product.type ?? "-")
                            .frame(maxWidth: .infinity)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.vertical, Theme.Spacing.medium)
                    .padding(.horizontal, Theme.Spacing.extra_small)
                    .background(Theme.Colors.white)
                    .cornerRadius(Theme.CornerRadius.medium)
                }
                .frame(maxWidth: .infinity)
            }
            
            private struct AttributeCellView: View {
                let title: String
                let value: String
                
                var body: some View {
                    VStack(alignment: .leading, spacing: Theme.Spacing.extra_small) {
                        Text(title)
                            .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                            .fontWeight(.light)
                            .foregroundColor(Theme.Colors.placeholder)
                        
                        Text(value)
                            .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                            .fontWeight(.medium)
                    }
                }
            }
            
            private struct VerticalDividerView: View {
                var body: some View {
                    Rectangle()
                        .frame(width: 1)
                        .foregroundColor(Theme.Colors.backgroundGray)
                }
            }
        }
        
        // MARK: - product stats
        private struct ProductStatsView: View {
            var body: some View {
                Text("Test stats")
                    .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.smallHeadlineSize))
                    .fontWeight(.semibold)
            }
        }
        
        // MARK: - previous tests list
        private struct PreviousTestsListView: View {
            
            @ObservedObject var vm: DetailedProductsViewModel
            
            var body: some View {
                
                VStack(alignment: .leading, spacing: Theme.Spacing.extra_small){
                    
                    Text("Previous tests")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.smallHeadlineSize))
                        .fontWeight(.semibold)
                    
                    LazyVStack(spacing: Theme.Spacing.small) {
                        ForEach(vm.tests) { test in
                            TestListItemView(
                                title: test.title,
                                date: test.date,
                                productCount: test.productCount,
                                temperature: test.temperature,
                                location: test.location,
                                weatherIcon: test.weatherIcon,
                                isPrivate: test.isPrivate
                            )
                        }
                    }
                }
            }
        }
    }
}

// MARK: - preview
#Preview {
    DetailedProductView(productId: 6)
}

