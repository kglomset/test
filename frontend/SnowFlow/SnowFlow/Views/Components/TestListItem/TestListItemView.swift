import SwiftUI

// MARK: - Test list item
struct TestListItemView: View {
    let title: String
    let date: String
    let productCount: Int
    let temperature: String
    let location: String
    let weatherIcon: String
    let isPrivate: Bool
    
    var body: some View {
        VStack(spacing: Theme.Spacing.extra_small){
            
            // icon, title, date & lock
            HStack(alignment: .top, spacing: Theme.Spacing.small){
                Image(weatherIcon)
                    .resizable()
                    .frame(width: 32, height: 32)
                
                Text(title)
                    .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                    .fontWeight(.medium)
                    .foregroundColor(Theme.Colors.primary)
                    .multilineTextAlignment(.leading)
                    .lineLimit(2)
                    .truncationMode(.tail)
                    .frame(maxWidth: .infinity, alignment: .leading)
                
                HStack(alignment: .top, spacing: Theme.Spacing.extra_small) {
                    Text(date)
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                        .fontWeight(.light)
                        .foregroundColor(Theme.Colors.secondary)
                    
                    if isPrivate {
                        Image(systemName: "lock")
                            .font(.system(size: Theme.Fonts.captionSize, weight: .light, design: .monospaced))
                            .foregroundColor(Theme.Colors.secondary)
                    }
                }
            }
            
            // tags
            HStack(spacing: Theme.Spacing.extra_small){
                tagView(value: "\(productCount) products")
                tagView(value: temperature)
                tagView(value: location)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(Theme.Spacing.small)
        .background(Theme.Colors.backgroundLight)
        .cornerRadius(Theme.CornerRadius.medium)
    }
    
    // MARK: - test item tag
    private struct tagView: View {
        let value: String
        
        var body: some View {
            Text(value)
                .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                .fontWeight(.light)
                .foregroundStyle(Theme.Colors.secondary)
                .padding(.vertical, Theme.Spacing.extra_small)
                .padding(.horizontal, Theme.Spacing.small)
                .background(Theme.Colors.backgroundGray)
                .cornerRadius(Theme.CornerRadius.small)
        }
    }
}
