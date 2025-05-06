import Testing
@testable import SnowFlow

@Suite("NetworkErrorMapper Tests")
struct NetworkErrorMapperTests {

    @Test("Nil content type returns .unexpectedData")
    func testNilContentType() {
        let error = NetworkErrorMapper.validateContentType(nil, expected: .json)
        #expect(error == .unexpectedData)
    }

    @Test("Valid content type returns nil")
    func testValidContentType() {
        let error = NetworkErrorMapper.validateContentType("application/json", expected: .json)
        #expect(error == nil)
    }

    @Test("Mismatched content type returns .unsupportedContentType")
    func testMismatchedContentType() {
        let error = NetworkErrorMapper.validateContentType("text/plain", expected: .json)
        #expect(error == .unsupportedContentType("text/plain"))
    }
}
