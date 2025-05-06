import Foundation

/*
 A test has meta data and content. This is the content.
 A tournement is a H2H format however each H2H has to be done twice
 to count for favorable conditions. This is considered a match and a match
 consits of two battles. The winner of the match moves on to the next round.
 */

// MARK: - Tournament model
struct Tournament: Codable {
    var products: [Product]
    var rounds: [Round]
}

// MARK: - Round model
struct Round: Codable {
    var round: Int
    var matches: [Match]
}

// MARK: - Match model (contains two battles)
struct Match: Codable {
    let matchID: Int
    let product1: ProductReference
    let product2: ProductReference
    let battle1: Int // with reference to product 1 so if product 1 wins +20
    let battle2: Int // with reference to product 1 so if product 1 looses -20
    var result: MatchResult
    
    enum CodingKeys: String, CodingKey {
        case matchID = "match_id"
        case product1, product2, battle1, battle2, result
    }
    
    // perfect to unit test - do it
    mutating func determineMatchResult() {
        let sum = battle1 + battle2
        result = MatchResult(
            winner: sum == 0 ? nil : (sum > 0 ? product1 : product2),
            sum: abs(sum)
        )
    }
}

// MARK: - Product reference
struct ProductReference: Codable, Equatable {
    let id: Int
    
    static func == (lhs: ProductReference, rhs: ProductReference) -> Bool {
        return lhs.id == rhs.id
    }
}

// MARK: - Match result model
struct MatchResult: Codable {
    let winner: ProductReference?  // if nil, it's a draw
    let sum: Int
}
