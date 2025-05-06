import Foundation

// MARK: - Predefined products
struct PredefinedProducts {
    static let products: [Int: Product] = [
        1: Product(
            productId: 1,
            name: "SG 10",
            ean: "1234567890123",
            brand: "Start",
            warmTemp: -10,
            coldTemp: -30,
            type: "Solid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=SG+10"),
            buyUrl: URL(string: "https://shop.example.com/sg10"),
            comment: """
            Start SG10 is a green, fluorine-free glide wax that you should consider buying.
            Despite becoming extremely hard and glossy, it is relatively easy to work with.
            Start SG10 is widely used in racing as a base glide wax or on its own on the coldest days. Iron 155.
            """,
            isOwner: true,
            isPrivate: false
        ),
        2: Product(
            productId: 2,
            name: "UHX Polar",
            ean: "2345678901234",
            brand: "HWK",
            warmTemp: -7,
            coldTemp: -25,
            type: "Powder",
            imageUrl: URL(string: "https://placehold.co/400x500?text=UHX+Polar"),
            buyUrl: URL(string: "https://shop.example.com/uhx_polar"),
            comment: "Great for extreme cold races.",
            isOwner: false,
            isPrivate: true
        ),
        3: Product(
            productId: 3,
            name: "Race Pink",
            ean: "3456789012345",
            brand: "Ulla",
            warmTemp: 4,
            coldTemp: -4,
            type: "Liquid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Race+Pink"),
            buyUrl: URL(string: "https://shop.example.com/race_pink"),
            comment: "Fast glide in mixed conditions.",
            isOwner: false,
            isPrivate: false
        ),
        4: Product(
            productId: 4,
            name: "Rex Gold Liquid Spray",
            ean: "5678901234567",
            brand: "Rex",
            warmTemp: nil,
            coldTemp: nil,
            type: "Liquid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Gold+Liquid+Spray"),
            buyUrl: URL(string: "https://shop.example.com/rex_gold_liquid"),
            comment: "Rex Gold Liquid Spray is a high-performance liquid glide wax for racing conditions.",
            isOwner: false,
            isPrivate: false
        ),
        5: Product(
            productId: 5,
            name: "Ulla Liquid Race Speed Yellow Black",
            ean: "6789012345678",
            brand: "Ulla",
            warmTemp: 6,
            coldTemp: -2,
            type: "Liquid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Speed+Yellow+Black"),
            buyUrl: URL(string: "https://shop.example.com/ulla_race_speed"),
            comment: "Ulla Liquid Race Speed Yellow Black provides excellent speed in varied race conditions.",
            isOwner: false,
            isPrivate: false
        ),
        6: Product(
            productId: 6,
            name: "Red Creek Fluor Free Racing Liquid Old Fashion Green",
            ean: "7890123456789",
            brand: "Red Creek",
            warmTemp: -3,
            coldTemp: -14,
            type: "Liquid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Old+Fashion+Green"),
            buyUrl: URL(string: "https://shop.example.com/red_creek_green"),
            comment: "Fluor-free liquid wax optimized for colder conditions, ensuring a smooth glide.",
            isOwner: false,
            isPrivate: false
        ),
        7: Product(
            productId: 7,
            name: "Red Creek Fluor Free Racing Liquid Old Fashion Blue",
            ean: "8901234567890",
            brand: "Red Creek",
            warmTemp: -1,
            coldTemp: -5,
            type: "Liquid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Old+Fashion+Blue"),
            buyUrl: URL(string: "https://shop.example.com/red_creek_blue"),
            comment: "Perfect for cold and slightly humid conditions, providing great durability.",
            isOwner: false,
            isPrivate: false
        ),
        8: Product(
            productId: 8,
            name: "Red Creek Fluor Free Racing Silver",
            ean: "9012345678901",
            brand: "Red Creek",
            warmTemp: 15,
            coldTemp: 0,
            type: "Solid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Racing+Silver"),
            buyUrl: URL(string: "https://shop.example.com/red_creek_silver"),
            comment: "Excellent in warmer temperatures, offering a consistent glide.",
            isOwner: false,
            isPrivate: false
        ),
        9: Product(
            productId: 9,
            name: "Red Creek Fluor Free Racing Old Fashion Blue",
            ean: "0123456789012",
            brand: "Red Creek",
            warmTemp: -1,
            coldTemp: -5,
            type: "Solid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Old+Fashion+Blue"),
            buyUrl: URL(string: "https://shop.example.com/red_creek_blue_solid"),
            comment: "Solid wax for cold and stable conditions, great durability.",
            isOwner: false,
            isPrivate: false
        ),
        10: Product(
            productId: 10,
            name: "Red Creek Fluor Free Old Fashion Green",
            ean: "1234567890123",
            brand: "Red Creek",
            warmTemp: -3,
            coldTemp: -14,
            type: "Solid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Old+Fashion+Green"),
            buyUrl: URL(string: "https://shop.example.com/red_creek_green_solid"),
            comment: "Designed for extreme cold conditions, ensuring a long-lasting and smooth glide.",
            isOwner: false,
            isPrivate: false
        ),
        11: Product(
            productId: 11,
            name: "Red Creek Fluor Free Racing Old Fashion Violet",
            ean: "2345678901234",
            brand: "Red Creek",
            warmTemp: 2,
            coldTemp: -12,
            type: "Solid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Old+Fashion+Violet"),
            buyUrl: URL(string: "https://shop.example.com/red_creek_violet"),
            comment: "Great for mixed conditions where temperatures vary between mild and cold.",
            isOwner: false,
            isPrivate: false
        ),
        12: Product(
            productId: 12,
            name: "Vauhti Pure One Polar",
            ean: "3456789012345",
            brand: "Vauhti",
            warmTemp: -5,
            coldTemp: -25,
            type: "Solid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Pure+One+Polar"),
            buyUrl: URL(string: "https://shop.example.com/vauhti_polar"),
            comment: "Ideal for polar conditions, offering excellent glide even in extreme cold.",
            isOwner: false,
            isPrivate: false
        ),
        13: Product(
            productId: 13,
            name: "Rex NF41 Liquid Glider",
            ean: "4567890123456",
            brand: "Rex",
            warmTemp: 5,
            coldTemp: -20,
            type: "Liquid",
            imageUrl: URL(string: "https://placehold.co/400x500?text=NF41+Liquid+Glider"),
            buyUrl: URL(string: "https://shop.example.com/rex_nf41"),
            comment: "Versatile liquid glide wax performing well across a wide range of temperatures.",
            isOwner: false,
            isPrivate: false
        ),
        14: Product(
            productId: 14,
            name: "Toko High Performance Powder Yellow",
            ean: "5678901234567",
            brand: "Toko",
            warmTemp: 10,
            coldTemp: -4,
            type: "Powder",
            imageUrl: URL(string: "https://placehold.co/400x500?text=Powder+Yellow"),
            buyUrl: URL(string: "https://shop.example.com/toko_hp_yellow"),
            comment: "High-performance powder wax optimized for warm-to-mild conditions.",
            isOwner: false,
            isPrivate: false
        )
    ]

    /// Fetch a full product by its `productId`
    static func getProduct(by id: Int) -> Product? {
        return products[id]
    }
}

// MARK: - Predefined products preview
struct PredefinedProductsPreview {
    static let previews: [Int: ProductPreviewM] = PredefinedProducts.products.mapValues { product in
        ProductPreviewM(
            productId: product.productId,
            name: product.name,
            ean: product.ean,
            brand: product.brand,
            warmTemp: product.warmTemp,
            coldTemp: product.coldTemp,
            type: product.type
        )
    }
    
    /// Get a product preview by id
    static func getProductPreviewById(id: Int) -> ProductPreviewM? {
        return previews[id]
    }
    
    /// Get all product previews as an array
    static func getAllProductPreviews() -> [ProductPreviewM] {
        return Array(previews.values)
    }
}

// MARK: - Predefined drafts
struct PredefinedDrafts {
    static let drafts: [Int: DraftM] = [
        0: DraftM(
            draftId: 0,
            title: "Winter sports prep",
            date: Date(),
            testSamples: [
                SampleM(skiName: "X", productPreview: PredefinedProductsPreview.getProductPreviewById(id: 10)!),
                SampleM(skiName: "Y", productPreview: PredefinedProductsPreview.getProductPreviewById(id: 11)!)
            ],
            track: nil,
            airTemp: -12,
            airHumidity: 80,
            snowTemp: -14,
            snowType: "Powder",
            snowHardness: nil,
            snowMoisture: nil,
            location: "Alps",
            weatherIcon: "lightsnow",
            isPrivate: false,
            comment: "Testing new skis in soft snow."
        ),
        
        1: DraftM(
            draftId: 1,
            title: "Ski competition",
            date: Calendar.current.date(from: DateComponents(year: 2025, month: 03, day: 12)),
            testSamples: [
                SampleM(skiName: "Z", productPreview: PredefinedProductsPreview.getProductPreviewById(id: 14)!),
                SampleM(skiName: "A", productPreview: PredefinedProductsPreview.getProductPreviewById(id: 10)!)
            ],
            track: "Olympic Trail",
            airTemp: -6,
            airHumidity: 75,
            snowTemp: nil,
            snowType: "Ice",
            snowHardness: "Very Hard",
            snowMoisture: 10,
            location: "Norway",
            weatherIcon: "fog",
            isPrivate: true,
            comment: nil
        ),

        2: DraftM(
            draftId: 2,
            title: "Casual ski session",
            date: nil,
            testSamples: [],
            track: nil,
            airTemp: nil,
            airHumidity: nil,
            snowTemp: nil,
            snowType: nil,
            snowHardness: nil,
            snowMoisture: nil,
            location: nil,
            weatherIcon: nil,
            isPrivate: nil,
            comment: nil
        ),

        3: DraftM(
            draftId: 3,
            title: "Extreme cold test",
            date: Calendar.current.date(from: DateComponents(year: 2024, month: 12, day: 24)),
            testSamples: [
                SampleM(skiName: "B", productPreview: PredefinedProductsPreview.getProductPreviewById(id: 4)!)
            ],
            track: "High Altitude Run",
            airTemp: -30,
            airHumidity: 60,
            snowTemp: -35,
            snowType: "Dry Powder",
            snowHardness: "Soft",
            snowMoisture: 5,
            location: "Switzerland",
            weatherIcon: "clearsky_day",
            isPrivate: false,
            comment: "Testing at extreme cold conditions."
        ),

        4: DraftM(
            draftId: 4,
            title: "Full data entry",
            date: Calendar.current.date(from: DateComponents(year: 2025, month: 02, day: 10)),
            testSamples: [
                SampleM(skiName: "C", productPreview: PredefinedProductsPreview.getProductPreviewById(id: 8)!),
                SampleM(skiName: "D", productPreview: PredefinedProductsPreview.getProductPreviewById(id: 2)!),
                SampleM(skiName: "E", productPreview: PredefinedProductsPreview.getProductPreviewById(id: 7)!)
            ],
            track: "Pro Championship Track",
            airTemp: -8,
            airHumidity: 85,
            snowTemp: -10,
            snowType: "Packed Snow",
            snowHardness: "Medium",
            snowMoisture: 20,
            location: "Finland",
            weatherIcon: "fair_day",
            isPrivate: true,
            comment: "Complete dataset for professional race testing."
        )
    ]

    /// Get a single draft by id
    static func getDraft(by id: Int) -> DraftM {
        return drafts[id] ?? DraftM(draftId: -1) // creates a new draft if not found ... for now
    }
}

// MARK: - Predefined drafts preview
struct PredefinedDraftsPreview {
    static let previews: [Int: DraftPreviewM] = PredefinedDrafts.drafts.mapValues { draft in
        DraftPreviewM(
            id: UUID(),
            draftId: draft.draftId,
            title: draft.title,
            date: draft.date,
            productCount: draft.testSamples.count,
            temperature: draft.airTemp.map { "\($0)°C" },
            location: draft.location,
            weatherIcon: draft.weatherIcon,
            isPrivate: draft.isPrivate ?? false
        )
    }
    
    /// Get a draft preview by id
    static func getDraftPreviewById(id: Int) -> DraftPreviewM? {
        return previews[id]
    }

    /// Get all draft previews as an array
    static func getAllDraftPreviews() -> [DraftPreviewM] {
        return Array(previews.values)
    }
}

// MARK: - Predefined tests
struct PredefinedTest {
    static let tests: [Int: Test] = [
        0: Test(
            title: "Ski Test - Morning Glide",
            date: "2025-03-04",
            productCount: 2,
            temperature: "-5°C",
            location: "Aspen, USA",
            weatherIcon: "heavysnow",
            isPrivate: false,
            comment: "Testing new fluor-free wax in icy conditions.",
            tournement: nil
        ),
        1: Test(
            title: "Race Day Preparation",
            date: "2025-03-05",
            productCount: 3,
            temperature: "-10°C",
            location: "Oslo, Norway",
            weatherIcon: "cloudy",
            isPrivate: true,
            comment: "Final glide wax selection for competition.",
            tournement: nil
        ),
        2: Test(
            title: "Casual Skiing with Friends",
            date: "2025-03-02",
            productCount: 1,
            temperature: "0°C",
            location: "Whistler, Canada",
            weatherIcon: "clearsky_day",
            isPrivate: false,
            comment: "Just a fun day on the slopes!",
            tournement: nil
        ),
        3: Test(
            title: "Extreme Cold Wax Trial",
            date: "2025-03-06",
            productCount: 2,
            temperature: "-25°C",
            location: "Lapland, Finland",
            weatherIcon: "fair_day",
            isPrivate: true,
            comment: "Testing endurance in sub-zero conditions.",
            tournement: nil
        ),
        4: Test(
            title: "Professional Waxing Seminar",
            date: "2025-03-10",
            productCount: 4,
            temperature: "-2°C",
            location: "Zermatt, Switzerland",
            weatherIcon: "fog",
            isPrivate: false,
            comment: "Demonstration of new race wax formulas.",
            tournement: nil
        )
    ]

    /// Get a test by its id
    static func getTest(by id: Int) -> Test? {
        return tests[id]
    }

    /// Get all tests as an array
    static func getAllTests() -> [Test] {
        return Array(tests.values)
    }
}

// MARK: - Predefined tests preview
struct PredefinedTestPreview {
    static let previews: [Int: TestPreview] = PredefinedTest.tests.mapValues { test in
        TestPreview(
            id: test.id,
            title: test.title,
            date: test.date,
            productCount: test.productCount,
            temperature: test.temperature,
            location: test.location,
            weatherIcon: test.weatherIcon,
            isPrivate: test.isPrivate
        )
    }

    /// Get a test preview by id
    static func getTestPreviewById(id: Int) -> TestPreview? {
        return previews[id]
    }

    /// Get all test previews as an array
    static func getAllTestPreviews() -> [TestPreview] {
        return Array(previews.values)
    }
}

