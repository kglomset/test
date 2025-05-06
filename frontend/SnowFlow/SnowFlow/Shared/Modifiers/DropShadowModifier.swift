import SwiftUI

struct DropShadowModifier: ViewModifier {
    func body(content: Content) -> some View {
        content
            .shadow(color: Color(hex: "#7D828D").opacity(0.08), radius: 8.8, x: 0, y: 2)
    }
}

extension View {
    func applyDropShadow() -> some View {
        self.modifier(DropShadowModifier())
    }
}
