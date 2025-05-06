import SwiftUI

struct Theme {
    
    struct Colors {
        // Text colors
        static let primary = Color(hex: "#000814")
        static let compPrimary = Color(hex: "#FDFEFF")
        static let secondary = Color(hex: "#3B434F")
        static let placeholder = Color(hex: "#68707C")
        
        static let white = Color(hex: "#FFFFFF")

        // Background colors
        static let backgroundLight = Color(hex: "#FFFFFF")
        static let backgroundGray = Color(hex: "#F0F1F2")

        // Borders & dividers
        static let border = Color(hex: "#68707C")
        static let divider = Color(hex: "#F0F1F2")

        // Feedback colors
        static let error = Color(hex: "#B00020")
        static let selected = Color.blue
        static let warning = Color(hex: "#000814")
        static let success = Color(hex: "#000814")
    }
    
    struct Spacing {
        static let extra_small: CGFloat = 4
        static let small: CGFloat = 8
        static let medium: CGFloat = 16
        static let large: CGFloat = 24
        static let extra_large: CGFloat = 32
    }
    
    struct CornerRadius {
        static let small: CGFloat = 4
        static let medium: CGFloat = 8
        static let large: CGFloat = 12
    }
    
    struct Fonts {
        static let bodyName = "Bricolage Grotesque"
        static let smallHeadlineSize: CGFloat = 16
        static let bodySize: CGFloat = 14
        static let captionSize: CGFloat = 12
    }
}
