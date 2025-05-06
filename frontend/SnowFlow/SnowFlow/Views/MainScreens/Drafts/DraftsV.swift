import SwiftUI

struct DraftsView: View {
    
    @ObservedObject var vm: DraftsViewModel
    @State private var selectedDraft: DraftPreviewM? = nil
    
    var body: some View {
        NavigationView {
            // drafts
            ScrollView {
                VStack(alignment: .leading, spacing: Theme.Spacing.extra_small) {
                    
                    Text("Drafts")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.smallHeadlineSize))
                        .fontWeight(.semibold)
                    
                    LazyVStack(spacing: Theme.Spacing.small) {
                        
                        ForEach(vm.drafts, id: \.draftId) { preview in
                            DraftRowView(draft: preview)
                                .onTapGesture {
                                    selectedDraft = preview
                                }
                        }
                    }
                }
                .padding(.horizontal, Theme.Spacing.medium)
            }
            .background(Theme.Colors.backgroundGray)
            .overlay(
                Button(action: {
                    let newDraftViewModel = vm.addNewDraft()
                    selectedDraft = newDraftViewModel
                }) {
                    Image(systemName: "plus")
                        .font(.system(size: 18, weight: .regular, design: .monospaced))
                }
                    .frame(width: 50, height: 38)
                    .background(Theme.Colors.backgroundLight)
                    .cornerRadius(Theme.CornerRadius.medium)
                    .shadow(color: Color(hex: "#7D828D").opacity(0.64), radius: 6.4, x: 0, y: 1.5)
                    .padding(),
                
                alignment: .bottomTrailing
            )
        }
        .frame(maxWidth: .infinity)
        .background(Color(Theme.Colors.backgroundGray))
        .fullScreenCover(item: $selectedDraft) { draftToShow in
            DetailedDraftView(vm: DetailedDraftVM(id: draftToShow.draftId))
        }
    }
}

// MARK: - Date formatter
private let dateFormatter: DateFormatter = {
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    return formatter
}()

// MARK: - preview
#Preview {
    ContentView(initialTab: .drafts)
}
