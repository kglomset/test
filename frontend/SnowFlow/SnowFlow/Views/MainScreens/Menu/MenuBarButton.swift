import SwiftUI

struct TabBarButton: View {
    let icon: String
    let tab: MenuBar.Tab
    @Binding var selectedTab: MenuBar.Tab
    var animationNamespace: Namespace.ID
    
    var body: some View {
        Button(action: {
            withAnimation(.spring(response: 0.4, dampingFraction: 0.6, blendDuration: 0.2)) {
                selectedTab = tab
            }
        }) {
            ZStack {
                if selectedTab == tab {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.white.opacity(0.2))
                        .frame(width: 36, height: 36)
                        .matchedGeometryEffect(id: "tabIndicator", in: animationNamespace)
                }
                
                Image(systemName: icon)
                    .font(.system(size: 18))
                    .foregroundColor(selectedTab == tab ? .white : .gray)
            }
            .frame(maxWidth: .infinity)
        }
    }
}

#Preview {
    @Previewable @State var selectedTab: MenuBar.Tab = .analytics
    @Previewable @Namespace var animationNamespace
    
    HStack {
        TabBarButton(
            icon: "chart.bar",
            tab: .analytics,
            selectedTab: $selectedTab,
            animationNamespace: animationNamespace
        )
        
        TabBarButton(
            icon: "triangle",
            tab: .products,
            selectedTab: $selectedTab,
            animationNamespace: animationNamespace
        )
    }
    .background(Color.black)
}
