import SwiftUI

// MARK: - Core Button
struct CoreButton: View {
    var title: String
    var backgroundColor: Color
    var foregroundColor: Color
    var borderColor: Color
    var expand: Bool
    var action: () -> Void
    
    var body: some View {
        Button(action: action) {
            Text(title)
                .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                .fontWeight(.medium)
                .multilineTextAlignment(.center)
                .foregroundColor(foregroundColor)
                .padding(.vertical, 8)
                .padding(.horizontal, 24)
                .frame(minWidth: 64, maxWidth: expand ? .infinity : nil)
        }
        .background(backgroundColor)
        .cornerRadius(24)
        .overlay(RoundedRectangle(cornerRadius: 24).stroke(borderColor, lineWidth: 1))
    }
}

// MARK: - Primary Button
struct PrimaryButton: View {
    var title: String
    var expand: Bool
    var action: () -> Void
    
    var body: some View {
        CoreButton(
            title: title,
            backgroundColor: Theme.Colors.primary,
            foregroundColor: Theme.Colors.compPrimary,
            borderColor: Theme.Colors.primary,
            expand: expand,
            action: action
        )
    }
}

// MARK: - Primary Button Preview
struct PrimaryButton_Previews: PreviewProvider {
    static var previews: some View {
        VStack{
            PrimaryButton(title: "Primary button", expand: true) {
                print("Primary button clicked")
            }
        }
        .frame(width: 220, height: 140)
        .background(Theme.Colors.backgroundGray)
        .background(Color.blue)
        .previewLayout(.sizeThatFits)
    }
}

// MARK: - Secondary Button
struct SecondaryButton: View {
    var title: String
    var expand: Bool
    var action: () -> Void
    
    var body: some View {
        CoreButton(
            title: title,
            backgroundColor: .clear,
            foregroundColor: Theme.Colors.primary,
            borderColor: Theme.Colors.primary,
            expand: expand,
            action: action
        )
    }
}

// MARK: - Secondary Button Preview
struct SecondaryButton_Previews: PreviewProvider {
    static var previews: some View {
        VStack {
            SecondaryButton(title: "Secondary button", expand: true) {
                print("Secondary button clicked")
            }
        }
        .frame(width: 280, height: 140)
        .background(Theme.Colors.backgroundGray)
        .background(Color.blue)
        .previewLayout(.sizeThatFits)
    }
}
