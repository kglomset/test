-- Add sample entries to database
SET session_replication_role = 'replica';

INSERT INTO team (name, team_role) VALUES ('SB', 1), ('WaxMax', 2);

INSERT INTO users (email, password, user_role, team_id) VALUES
                                                            ('oskar@pokemon.com', 'pokemongo', 'admin', 1),
                                                            ('petter@northug.com', 'barneskirenn', 'member', 2);

INSERT INTO products (name, brand, ean_code, image_url, comment, is_public, type,
                      high_temperature, low_temperature, testing_team, version, status)
VALUES
    ('SkiFast 2000', 'SnowMax', '1234567890123', '', 'High performance wax for cold conditions', TRUE, 'solid', -10, -30, '1', '2025-03-30T10:30:00Z', 'active'),
    ('InsaneGel', 'GlidePro', '2345678901234', '', 'Eco-friendly gel for moderate temperatures', TRUE, 'gel', 0, -15, '1', '2025-03-30T10:30:00Z', 'active'),
    ('Pokemon wax', '', '', '', 'Durable wax for high-speed skiing', FALSE, 'bundle', 5, -20, '1', '2025-03-30T10:30:00Z', 'development'),
    ('SuperFast', 'IceGrip', '4567890123456', '', 'Affordable liquid wax for icy conditions', TRUE, 'liquid', -5, -25, '1', '2025-03-30T10:30:00Z', 'active'),
    ('Pokemon powder', 'PeakPerformance', '5678901234567', '', 'Premium powder for all temperatures', FALSE, 'powder', 10, -10, '1', '2025-03-30T10:30:00Z', 'development'),
    ('RapidGo', 'RapidGlide', '6789012345678', '', 'Best seller spray for quick application', TRUE, 'spray', 15, -5, '1', '2025-03-30T10:30:00Z', 'active'),
    ('SuperGo Bundle', '', '', '', 'Super duper bundle', TRUE, 'bundle', -1, -5, '1', '2025-03-30T10:30:00Z', 'active');

INSERT INTO product_bundles (bundle_id, product_id, layer_no) VALUES
                                                                  (1, 1, 1),
                                                                  (1, 2, 2),
                                                                  (2, 4, 1),
                                                                  (2, 5, 2),
                                                                  (2, 6, 3);

INSERT INTO air_conditions (temperature, humidity, wind, cloud) VALUES (-10, 75, 'L', '1');
INSERT INTO snow_conditions (temperature, snow_type, snow_humidity) VALUES (-2, 'FS', 'W2');
INSERT INTO track_conditions (track_hardness, track_type) VALUES ('H1', 'D1');

INSERT INTO tests (location, comment, sc_id, tc_id, ac_id, testing_team)
VALUES ('Beito', 'Test på kaldt føre', 1, 1, 1, 1);

INSERT INTO test_ranks (test_id, product_id, rank, distance_behind, version, is_rank_public)
VALUES (1, 3, 1, 0, '2025-03-30T10:30:00Z', true),
       (1, 7, 2, 15, '2025-03-30T10:30:00Z', true);

SET session_replication_role = 'origin';
