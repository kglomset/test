import SwiftUI

struct MenuBar: View {
    @Binding var selectedTab: Tab
    @Namespace private var animationNamespace
    
    enum Tab: String, CaseIterable {
        case analytics = "chart.bar"
        case tests = "clipboard"
        case drafts = "flask"
        case products = "triangle"
        case profile = "person.crop.circle"
        
        @ViewBuilder
        func view(
            analyticsVM: AnalyticsViewModel,
            testsVM: TestsViewModel,
            draftsVM: DraftsViewModel,
            productsVM: ProductsViewModel,
            profileVM: ProfileViewModel
        ) -> some View {
            switch self {
            case .analytics:
                AnalyticsView(vm: analyticsVM)
            case .tests:
                TestsView(viewModel: testsVM)
            case .drafts:
                DraftsView(vm: draftsVM)
            case .products:
                ProductsView(viewModel: productsVM)
            case .profile:
                ProfileView(viewModel: profileVM)
            }
        }
    }
    
    var body: some View {
        HStack {
            ForEach(Tab.allCases, id: \.self) { tab in
                TabBarButton(icon: tab.rawValue, tab: tab, selectedTab: $selectedTab, animationNamespace: animationNamespace)
            }
        }
        .padding(.top, 12)
        .background(Theme.Colors.primary)
    }
}

#Preview {
    @Previewable @State var selectedTab: MenuBar.Tab = .analytics
    return MenuBar(selectedTab: $selectedTab)
}
