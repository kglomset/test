import SwiftUI

struct ContentView: View {
    init(initialTab: MenuBar.Tab? = nil) {
        _selectedTab = State(initialValue: initialTab ?? .analytics)
    }
    
    @State private var selectedTab: MenuBar.Tab
    @StateObject private var contentVM = ContentViewModel()
    
    // vms for persistence between tab selections
    @StateObject private var analyticsViewModel = AnalyticsViewModel()
    @StateObject private var testsViewModel = TestsViewModel()
    @StateObject private var draftsViewModel = DraftsViewModel()
    @StateObject private var productsViewModel = ProductsViewModel()
    @StateObject private var profileViewModel = ProfileViewModel()
    
    // state to control the fade-in effect for login
    @State private var isLoginVisible = false
    
    var body: some View {
        VStack(spacing: 0) {
            selectedTab.view(
                analyticsVM: analyticsViewModel,
                testsVM: testsViewModel,
                draftsVM: draftsViewModel,
                productsVM: productsViewModel,
                profileVM: profileViewModel
            )
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            
            MenuBar(selectedTab: $selectedTab)
        }
        .fullScreenCover(isPresented: $contentVM.showLogin) {
            Group {
                if isLoginVisible {
                    LoginView {
                        withAnimation(.linear(duration: 0.8)) {
                            isLoginVisible = false
                        }
                    }
                    .onDisappear {
                        contentVM.showLogin = false
                    }
                }
            }
            .onAppear {
                withAnimation(.linear(duration: 0.8)) {
                    isLoginVisible = true
                }
            }
        }
    }
}

#Preview {
    ContentView(initialTab: .products)
}
