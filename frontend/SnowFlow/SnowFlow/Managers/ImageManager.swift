import SwiftUI
import UIKit
import Foundation

// todo the network manager should come from the called
private let sharedNetworkManager = NetworkManagerFactory.create()

/// Asynchronously fetches an image from the given URL using the shared network manager.
/// - Parameter url: The URL from which to fetch the image.
/// - Returns: A UIImage fetched from cache or network.
/// - Throws: An error if the fetch or image decoding fails.
public func fetchImage(from url: URL) async throws -> UIImage {
    let (_, task) = sharedNetworkManager.fetchWithCache(url: url, expectedContentType: .image)
    let data = try await task.value
    return try CodecManager.decodeImage(from: data)
}

/// A SwiftUI view that displays an image fetched asynchronously.
/// While the image is loading, a placeholder image is shown.
struct AsyncImageView: View {
    let url: URL
    let placeholder: UIImage = UIImage(systemName: "photo")!
    
    @State private var image: UIImage?
    
    var body: some View {
        Group {
            if let loadedImage = image {
                Image(uiImage: loadedImage)
                    .resizable()
            } else {
                Image(uiImage: placeholder)
                    .resizable()
            }
        }
        .task {
            if image == nil {  // ensure we only fetch once
                do {
                    image = try await fetchImage(from: url)
                } catch {
                    print("Error fetching image from \(url): \(error)")
                }
            }
        }
    }
}

#Preview {
    VStack {
        AsyncImageView(url: URL(string: "https://placehold.co/400x500.png?text=This+is+a+preview")!)
            .frame(width: 200, height: 250)
    }
}
