import Foundation
import UIKit
import OSLog

/// Unified Cache Manager handling in‑memory and disk caches for different content types.
class CacheManager {
    
    // MARK: - singleton
    static let shared = CacheManager(config: .standard, directoryName: "UnifiedCache")
    
    private let logger = Logger(subsystem: "ntnu.stud.SnowFlow.Cache", category: "CacheManager")
    
    // MARK: - properties
    let config: CacheConfig
    let fileManager = FileManager.default
    
    // memory caches (NSCache is thread safe)
    private let jsonMemoryCache = NSCache<NSString, NSData>()
    private let imageMemoryCache = NSCache<NSString, NSData>()
    private let otherMemoryCache = NSCache<NSString, NSData>()
    
    // disk directories for each content type
    let jsonCacheDirectory: URL
    let imageCacheDirectory: URL
    let otherCacheDirectory: URL
    
    // in‑memory access timestamps for expiration (not thread safe by default)
    private var jsonAccessTimestamps: [String: Date] = [:]
    private var imageAccessTimestamps: [String: Date] = [:]
    private var otherAccessTimestamps: [String: Date] = [:]
    
    // synchronization queue for mutable state
    private let syncQueue = DispatchQueue(label: "com.cacheManager.syncQueue")
    
    private var isCleanupScheduled = false
    
    // MARK: - initialization
    init(config: CacheConfig, directoryName: String) {
        self.config = config
        
        // configure memory cache limits
        jsonMemoryCache.totalCostLimit = config.maxMemoryCacheSizeJSON
        imageMemoryCache.totalCostLimit = config.maxMemoryCacheSizeImages
        otherMemoryCache.totalCostLimit = config.maxMemoryCacheSizeOther
        
        // set up disk directories for caches
        let cacheBaseURL = fileManager.urls(for: .cachesDirectory, in: .userDomainMask).first!
        jsonCacheDirectory = cacheBaseURL.appendingPathComponent("\(directoryName)Json")
        imageCacheDirectory = cacheBaseURL.appendingPathComponent("\(directoryName)Images")
        otherCacheDirectory = cacheBaseURL.appendingPathComponent("\(directoryName)Other")
        
        // create directories if they don't exist
        try? fileManager.createDirectory(at: jsonCacheDirectory, withIntermediateDirectories: true)
        try? fileManager.createDirectory(at: imageCacheDirectory, withIntermediateDirectories: true)
        try? fileManager.createDirectory(at: otherCacheDirectory, withIntermediateDirectories: true)
        
        // register for memory warnings
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleMemoryWarning),
            name: UIApplication.didReceiveMemoryWarningNotification,
            object: nil
        )
        
        scheduleCleanupIfNeeded()
    }
    
    deinit {
        NotificationCenter.default.removeObserver(self)
    }
    
    // MARK: - public cache methods
    
    func saveData(_ data: Data, forKey key: String, contentType: ContentType) {
        let now = Date()
        switch contentType {
            
        case .json:
            jsonMemoryCache.setObject(data as NSData, forKey: key as NSString)
            syncQueue.async { self.jsonAccessTimestamps[key] = now }
            
            let fileURL = cacheFileURL(for: key, in: jsonCacheDirectory)
            try? data.write(to: fileURL)
            logger.log("cache saved (json): \(key)")
            
        case .image:
            imageMemoryCache.setObject(data as NSData, forKey: key as NSString)
            
            syncQueue.async { self.imageAccessTimestamps[key] = now }
            
            let fileURL = cacheFileURL(for: key, in: imageCacheDirectory)
            
            try? data.write(to: fileURL)
            logger.log("cache saved (image): \(key)")
            
        default:
            otherMemoryCache.setObject(data as NSData, forKey: key as NSString)
            syncQueue.async { self.otherAccessTimestamps[key] = now }
            
            let fileURL = cacheFileURL(for: key, in: otherCacheDirectory)
            try? data.write(to: fileURL)
            logger.log("cache saved (other): \(key)")
        }
    }
    
    func loadValidData(forKey key: String, contentType: ContentType) -> Data? {
        guard let data = loadData(forKey: key, contentType: contentType) else {
            logger.log("cache miss: \(key) for content type \(contentType.description)")
            return nil
        }
        if isExpired(forKey: key, contentType: contentType) {
            logger.log("cache stale: \(key) for content type \(contentType.description)")
            return nil
        }
        logger.log("cache hit: \(key) for content type \(contentType.description)")
        return data
    }
    
    func deleteCache(forKey key: String, contentType: ContentType) {
        switch contentType {
        case .json:
            jsonMemoryCache.removeObject(forKey: key as NSString)
            syncQueue.async { self.jsonAccessTimestamps.removeValue(forKey: key) }
            let fileURL = cacheFileURL(for: key, in: jsonCacheDirectory)
            try? fileManager.removeItem(at: fileURL)
            logger.log("cache deleted (json): \(key)")
            
        case .image:
            imageMemoryCache.removeObject(forKey: key as NSString)
            syncQueue.async { self.imageAccessTimestamps.removeValue(forKey: key) }
            let fileURL = cacheFileURL(for: key, in: imageCacheDirectory)
            try? fileManager.removeItem(at: fileURL)
            logger.log("cache deleted (image): \(key)")
            
        default:
            otherMemoryCache.removeObject(forKey: key as NSString)
            syncQueue.async { self.otherAccessTimestamps.removeValue(forKey: key) }
            let fileURL = cacheFileURL(for: key, in: otherCacheDirectory)
            try? fileManager.removeItem(at: fileURL)
            logger.log("cache deleted (other): \(key)")
        }
    }
    
    func emptyAllCaches() {
        // clear in-memory caches
        jsonMemoryCache.removeAllObjects()
        imageMemoryCache.removeAllObjects()
        otherMemoryCache.removeAllObjects()
        
        // clear access timestamps on a background thread
        syncQueue.async {
            self.jsonAccessTimestamps.removeAll()
            self.imageAccessTimestamps.removeAll()
            self.otherAccessTimestamps.removeAll()
        }
        
        // define an array of all cache directories
        let directories = [jsonCacheDirectory, imageCacheDirectory, otherCacheDirectory]
        
        // remove all files in each cache directory
        for directory in directories {
            if let fileURLs = try? fileManager.contentsOfDirectory(at: directory, includingPropertiesForKeys: nil) {
                for fileURL in fileURLs {
                    try? fileManager.removeItem(at: fileURL)
                }
            }
        }
        
        logger.log("all caches emptied")
    }
    
    private func loadData(forKey key: String, contentType: ContentType) -> Data? {
        let now = Date()
        
        switch contentType {
        case .json:
            if let data = jsonMemoryCache.object(forKey: key as NSString) as Data? {
                syncQueue.async { self.jsonAccessTimestamps[key] = now }
                return data
            }
            let jsonURL = cacheFileURL(for: key, in: jsonCacheDirectory)
            if let data = try? Data(contentsOf: jsonURL) {
                jsonMemoryCache.setObject(data as NSData, forKey: key as NSString)
                syncQueue.async { self.jsonAccessTimestamps[key] = now }
                return data
            }
            
        case .image:
            if let data = imageMemoryCache.object(forKey: key as NSString) as Data? {
                syncQueue.async { self.jsonAccessTimestamps[key] = now }
                return data
            }
            let imageURL = cacheFileURL(for: key, in: imageCacheDirectory)
            if let data = try? Data(contentsOf: imageURL) {
                imageMemoryCache.setObject(data as NSData, forKey: key as NSString)
                syncQueue.async { self.imageAccessTimestamps[key] = now }
                return data
            }
            
        default:
            if let data = otherMemoryCache.object(forKey: key as NSString) as Data? {
                syncQueue.async { self.otherAccessTimestamps[key] = now }
                return data
            }
            let otherURL = cacheFileURL(for: key, in: otherCacheDirectory)
            if let data = try? Data(contentsOf: otherURL) {
                otherMemoryCache.setObject(data as NSData, forKey: key as NSString)
                syncQueue.async { self.otherAccessTimestamps[key] = now }
                return data
            }
        }
        
        return nil
    }
    
    private func isExpired(forKey key: String, contentType: ContentType) -> Bool {
        let now = Date()
        switch contentType {
        case .json:
            let timestamp = syncQueue.sync { jsonAccessTimestamps[key] }
            guard let timestamp = timestamp else { return true }
            return now.timeIntervalSince(timestamp) > config.cacheExpiryJSON
        case .image:
            let timestamp = syncQueue.sync { imageAccessTimestamps[key] }
            guard let timestamp = timestamp else { return true }
            return now.timeIntervalSince(timestamp) > config.cacheExpiryImages
        default:
            let timestamp = syncQueue.sync { otherAccessTimestamps[key] }
            guard let timestamp = timestamp else { return true }
            return now.timeIntervalSince(timestamp) > config.cacheExpiryOther
        }
    }
    
    // MARK: - disk file utilities
    private func cacheFileURL(for key: String, in directory: URL) -> URL {
        let safeKey = makeSafeFilename(from: key)
        return directory.appendingPathComponent(safeKey)
    }
    
    private func makeSafeFilename(from key: String) -> String {
        return key.replacingOccurrences(of: "[^a-zA-Z0-9_]", with: "_", options: .regularExpression)
    }
    
    // MARK: - cache cleanup
    private func cleanupExpiredEntries() {
        cleanupExpiredJSON()
        cleanupExpiredImages()
        cleanupExpiredOthers()
    }
    
    private func cleanupExpiredJSON() {
        let now = Date()
        syncQueue.sync {
            for (key, timestamp) in jsonAccessTimestamps where now.timeIntervalSince(timestamp) > config.cacheExpiryJSON {
                jsonMemoryCache.removeObject(forKey: key as NSString)
                let fileURL = cacheFileURL(for: key, in: jsonCacheDirectory)
                try? fileManager.removeItem(at: fileURL)
                jsonAccessTimestamps.removeValue(forKey: key)
                logger.log("removed expired JSON cache: \(key)")
            }
        }
    }
    
    private func cleanupExpiredImages() {
        let now = Date()
        syncQueue.sync {
            for (key, timestamp) in imageAccessTimestamps where now.timeIntervalSince(timestamp) > config.cacheExpiryImages {
                imageMemoryCache.removeObject(forKey: key as NSString)
                let fileURL = cacheFileURL(for: key, in: imageCacheDirectory)
                try? fileManager.removeItem(at: fileURL)
                imageAccessTimestamps.removeValue(forKey: key)
                logger.log("removed expired image cache: \(key)")
            }
        }
    }
    
    private func cleanupExpiredOthers() {
        let now = Date()
        syncQueue.sync {
            for (key, timestamp) in otherAccessTimestamps where now.timeIntervalSince(timestamp) > config.cacheExpiryOther {
                otherMemoryCache.removeObject(forKey: key as NSString)
                let fileURL = cacheFileURL(for: key, in: otherCacheDirectory)
                try? fileManager.removeItem(at: fileURL)
                otherAccessTimestamps.removeValue(forKey: key)
                logger.log("removed expired other cache: \(key)")
            }
        }
    }
    
    private func scheduleCleanupIfNeeded() {
        syncQueue.async {
            guard !self.isCleanupScheduled else { return }
            self.isCleanupScheduled = true
            DispatchQueue.global(qos: .background).asyncAfter(deadline: .now() + 300) { [weak self] in
                self?.cleanupExpiredEntries()
                self?.syncQueue.async {
                    self?.isCleanupScheduled = false
                    self?.scheduleCleanupIfNeeded() // reschedule cleanup
                }
            }
        }
    }
    
    // MARK: - memory warning handling
    @objc private func handleMemoryWarning() {
        jsonMemoryCache.removeAllObjects()
        imageMemoryCache.removeAllObjects()
        otherMemoryCache.removeAllObjects()
        syncQueue.async {
            self.jsonAccessTimestamps.removeAll()
            self.imageAccessTimestamps.removeAll()
            self.otherAccessTimestamps.removeAll()
        }
        logger.log("memory warning: cleared all caches")
    }
    
    func printCacheParentFolderPath() {
        let fileManager = FileManager.default
        if let cachesDirectory = fileManager.urls(for: .cachesDirectory, in: .userDomainMask).first {
            print("caches parent folder path: \(cachesDirectory.path)")
        } else {
            print("caches parent folder path not found")
        }
    }
    
    func printMemoryCacheStatus() {
        let now = Date()
        print("\n--- memory cache status ---")
        
        print("json cache:")
        for (key, timestamp) in jsonAccessTimestamps {
            let expiresIn = config.cacheExpiryJSON - now.timeIntervalSince(timestamp)
            let fileSize = ((jsonMemoryCache.object(forKey: key as NSString))! as NSData).length
            print(" * \(key): \(fileSize) bytes, expires in: \(expiresIn) seconds")
        }
        
        print("image cache:")
        for (key, timestamp) in imageAccessTimestamps {
            let expiresIn = config.cacheExpiryImages - now.timeIntervalSince(timestamp)
            let fileSize = ((imageMemoryCache.object(forKey: key as NSString))! as NSData).length
            print(" * \(key): \(fileSize) bytes, expires in: \(expiresIn) seconds")
        }
        
        print("other cache:")
        for (key, timestamp) in otherAccessTimestamps {
            let expiresIn = config.cacheExpiryOther - now.timeIntervalSince(timestamp)
            let fileSize = ((otherMemoryCache.object(forKey: key as NSString))! as NSData).length
            print(" * \(key): \(fileSize) bytes, expires in: \(expiresIn) seconds")
        }
        
        print("--- end memory cache status ---\n")
    }
    
    func printDiskCacheStatus() {
        let now = Date()
        print("\n--- disk cache status ---")
        
        let caches: [(directory: URL, name: String, expiry: TimeInterval)] = [
            (jsonCacheDirectory, "json cache", config.cacheExpiryJSON),
            (imageCacheDirectory, "image cache", config.cacheExpiryImages),
            (otherCacheDirectory, "other cache", config.cacheExpiryOther)
        ]
        
        for cache in caches {
            print("\(cache.name):")
            if let files = try? fileManager.contentsOfDirectory(atPath: cache.directory.path) {
                for file in files {
                    let fileURL = cache.directory.appendingPathComponent(file)
                    let attributes = (try? fileManager.attributesOfItem(atPath: fileURL.path)) ?? [:]
                    let fileSize = attributes[.size] as? Int ?? 0
                    let modificationDate = attributes[.modificationDate] as? Date ?? now
                    let expiresIn = cache.expiry - now.timeIntervalSince(modificationDate)
                    print(" * \(file): \(fileSize) bytes, expires in: \(expiresIn) seconds")
                }
            }
        }
        
        print("--- end disk cache status ---\n")
    }
}
