import Foundation

class DraftsViewModel: ObservableObject {
    @Published var drafts: [DraftPreviewM] = PredefinedDraftsPreview.getAllDraftPreviews()
    
    /// Adds a new empty draft and returns it
    func addNewDraft() -> DraftPreviewM {
        let newDraft = DraftPreviewM(draftId: Int.random(in: 100...1000)) // Assuming `draftId` is an Int
        
        drafts.append(newDraft) // Append to the list
        
        return newDraft // Return the new draft
    }
}
