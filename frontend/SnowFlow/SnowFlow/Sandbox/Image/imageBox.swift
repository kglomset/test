import SwiftUI

class CacheDemoViewModel: ObservableObject {
    @Published var refreshId = UUID()
    
    private let cacheManager = CacheManager.shared
    
    // define image URLs (using absoluteString as cache key)
    let pngURL = URL(string: "https://placehold.co/400x500.png?text=This+is+a+png+preview")!
    let jpegURL = URL(string: "https://placehold.co/400x500.jpeg?text=This+is+a+jpeg+preview")!
    
    /// Triggers a refresh by changing the view's identity.
    func refresh() {
        refreshId = UUID()
    }
    
    /// Deletes cached images for both PNG and JPEG, then refreshes.
    func deleteCache() {
        cacheManager.emptyAllCaches()
        //cacheManager.deleteCache(forKey: pngURL.absoluteString, contentType: .image)
        //cacheManager.deleteCache(forKey: jpegURL.absoluteString, contentType: .image)
    }
    func checkCache() {
        cacheManager.printMemoryCacheStatus()
        cacheManager.printDiskCacheStatus()
    }
}

struct CacheDemoView: View {
    @StateObject private var vm = CacheDemoViewModel()
    
    var body: some View {
        VStack(spacing: 20) {
            VStack {
                AsyncImageView(url: vm.pngURL)
                    .frame(width: 200, height: 250)
                    .id(vm.refreshId) // forces re-creation on refresh
                AsyncImageView(url: vm.jpegURL)
                    .frame(width: 200, height: 250)
                    .id(vm.refreshId)
            }
            HStack(spacing: 40) {
                Button("Refresh") {
                    vm.refresh()
                }
                Button("Check cache") {
                    vm.checkCache()
                }
                Button("Delete cache") {
                    vm.deleteCache()
                }
            }
            .padding()
        }
    }
}
