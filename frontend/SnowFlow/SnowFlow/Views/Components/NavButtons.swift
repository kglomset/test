import SwiftUI

// MARK: - Nav button
struct NavButton: View {
    var text: String?
    var systemName: String?
    var action: () -> Void
    
    init(text: String? = nil, systemName: String? = nil, action: @escaping () -> Void) {
        self.text = text
        self.systemName = systemName
        self.action = action
    }
    
    var body: some View {
        Button(action: action) {
            if let systemName = systemName {
                Image(systemName: systemName)
                    .font(.system(size: Theme.Fonts.bodySize, weight: .regular, design: .monospaced))
                    .frame(width: 50)
            } else if let text = text {
                Text(text)
                    .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                    .fontWeight(.regular)
                    .frame(maxWidth: .infinity)
                    .padding(.horizontal, Theme.Spacing.medium)
            }
        }
        .frame(height: 38)
        .background(Theme.Colors.backgroundLight)
        .cornerRadius(Theme.CornerRadius.medium)
    }
}
