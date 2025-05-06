import SwiftUI

@main
struct SnowFlowApp: App {
    // computed property to check the researcher mode setting
    private var isResearcherModeEnabled: Bool {
        UserDefaults.standard.bool(forKey: "researcher_enabled_preference")
    }
    
    var body: some Scene {
        WindowGroup {
            if isResearcherModeEnabled {
                ContentView()
            } else {
                TodoListView()
            }
        }
    }
}
