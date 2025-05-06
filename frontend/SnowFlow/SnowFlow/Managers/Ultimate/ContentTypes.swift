enum ContentType: String {
    case json = "application/json"
    case plain = "text/plain"
    case pdf = "application/pdf"
    case image = "image/jpeg, image/png" // add heic?
    
    var description: String {
        return self.rawValue
    }
}
