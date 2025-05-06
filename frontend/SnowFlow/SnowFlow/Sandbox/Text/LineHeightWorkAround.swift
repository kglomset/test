import SwiftUI

struct FlowLayout: Layout {
    var spacing: CGFloat = 8
    
    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        var currentRowWidth: CGFloat = 0, currentRowHeight: CGFloat = 0, totalHeight: CGFloat = 0
        let maxWidth = proposal.width ?? .infinity
        
        for subview in subviews {
            let size = subview.sizeThatFits(.unspecified)
            if currentRowWidth + size.width > maxWidth {
                totalHeight += currentRowHeight + spacing
                currentRowWidth = size.width + spacing
                currentRowHeight = size.height
            } else {
                currentRowWidth += size.width + spacing
                currentRowHeight = max(currentRowHeight, size.height)
            }
        }
        totalHeight += currentRowHeight
        return CGSize(width: proposal.width ?? currentRowWidth, height: totalHeight)
    }
    
    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        var x = bounds.minX, y = bounds.minY, currentRowHeight: CGFloat = 0
        
        for subview in subviews {
            let size = subview.sizeThatFits(.unspecified)
            if x + size.width > bounds.maxX {
                x = bounds.minX
                y += currentRowHeight + spacing
                currentRowHeight = 0
            }
            subview.place(at: CGPoint(x: x, y: y), proposal: ProposedViewSize(size))
            x += size.width + spacing + 5 // space width
            currentRowHeight = max(currentRowHeight, size.height)
        }
    }
}

struct FlexBoxView: View {
    let text: String
    let fontSize: CGFloat = 32 // make it adjust based on the space available? with a min max value, font size value?
    
    // split the input string on whitespace.
    var words: [String] {
        text.split(separator: " ").map { String($0) }
    }
    
    var body: some View {
        FlowLayout(spacing: 0) {
            ForEach(words, id: \.self) { word in
                Text(word)
                    .font(.custom(Theme.Fonts.bodyName, size: fontSize))
                    .fontWeight(.medium)
                    .frame(height: (fontSize - 4)) // basicly negative line space
            }
        }
    }
}

#Preview{
    FlexBoxView(text: "This is a sample string that will be split into words and each word will be displayed in its own text box.")
        .padding()
}
