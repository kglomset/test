import Foundation

/// Centralized configuration for all caching operations
public struct CacheConfig {
    // MARK: - memory cache settings
    
    /// Maximum memory for JSON items (in bytes)
    let maxMemoryCacheSizeJSON: Int
    
    /// Maximum memory for image items (in bytes)
    let maxMemoryCacheSizeImages: Int
    
    /// Maximum memory for other items (in bytes)
    let maxMemoryCacheSizeOther: Int

    // MARK: - disk cache settings
    
    /// Maximum disk space for JSON cache (in bytes)
    let maxDiskCacheSizeJSON: Int64
    
    /// Maximum disk space for image cache (in bytes)
    let maxDiskCacheSizeImages: Int64
    
    /// Maximum disk space for other cache (in bytes)
    let maxDiskCacheSizeOther: Int64

    // MARK: - cache expiry settings
    
    /// Time after which cached JSON items expire (in seconds)
    let cacheExpiryJSON: TimeInterval
    
    /// Time after which cached images expire  (in seconds)
    let cacheExpiryImages: TimeInterval
    
    /// Time after which cached other items expire (in seconds)
    let cacheExpiryOther: TimeInterval

    // MARK: - debugging
    
    /// Enable verbose logging for cache operations
    let logEnabled: Bool

    // MARK: - standard configuration
    
    /// Standard cache settings with optimized performance
    static let standard = CacheConfig(
        // memory settings
        maxMemoryCacheSizeJSON: 50_000_000, // 50MB
        maxMemoryCacheSizeImages: 200_000_000, // 200MB
        maxMemoryCacheSizeOther: 50_000_000, // 50MB for other types
        
        // disk settings
        maxDiskCacheSizeJSON: 100_000_000, // 100MB for JSON
        maxDiskCacheSizeImages: 500_000_000, // 500MB for images
        maxDiskCacheSizeOther: 100_000_000, // 100MB for other types
        
        // expiry settings
        cacheExpiryJSON: 120, // 2 minutes
        cacheExpiryImages: 60 * 60 * 24 * 2, // 2 days
        cacheExpiryOther: 60 * 60, // 1 hour
        
        // debugging
        logEnabled: false
    )
}
