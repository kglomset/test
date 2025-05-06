import SwiftUI

struct DraftRowView: View {
    let draft: DraftPreviewM
    
    var body: some View {
        VStack(spacing: Theme.Spacing.extra_small){
            HStack(alignment: .top, spacing: Theme.Spacing.small){
                Image(draft.weatherIcon ?? "")
                    .resizable()
                    .aspectRatio(1, contentMode: .fit)
                    .frame(height: 32)
                
                Text(draft.title)
                    .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                    .fontWeight(.medium)
                    .foregroundColor(Theme.Colors.primary)
                    .multilineTextAlignment(.leading)
                    .lineLimit(2)
                    .truncationMode(.tail)
                    .frame(maxWidth: .infinity, alignment: .leading)
                
                HStack(alignment: .top, spacing: Theme.Spacing.extra_small) {
                    Text(draft.date?.toString() ?? "")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                        .fontWeight(.light)
                        .foregroundColor(Theme.Colors.secondary)
                    
                    if draft.isPrivate {
                        Image(systemName: "lock")
                            .font(.system(size: Theme.Fonts.captionSize, weight: .light, design: .monospaced))
                            .foregroundColor(Theme.Colors.secondary)
                    }
                }
            }
            
            HStack(spacing: Theme.Spacing.extra_small){
                tagView(value: "\(draft.productCount) products")
                if let temperature = draft.temperature {
                    tagView(value: temperature)
                }
                
                if let location = draft.location {
                    tagView(value: location)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(Theme.Spacing.small)
        .background(Theme.Colors.backgroundLight)
        .cornerRadius(Theme.CornerRadius.medium)
    }
    
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

// MARK: - Preview
#Preview {
    VStack {
        ForEach(PredefinedDraftsPreview.getAllDraftPreviews(), id: \.id) { draft in
            DraftRowView(draft: draft)
        }
    }
    .frame(maxHeight: .infinity)
    .padding()
    .background(Theme.Colors.backgroundGray)
}
