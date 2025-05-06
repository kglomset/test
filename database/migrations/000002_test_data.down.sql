-- Remove sample entries from database
SET session_replication_role = 'replica';

DELETE FROM test_ranks WHERE test_id = 1;

DELETE FROM tests WHERE id = 1;

DELETE FROM track_conditions WHERE id = 1;
DELETE FROM snow_conditions WHERE id = 1;
DELETE FROM air_conditions WHERE id = 1;

DELETE FROM product_bundles WHERE bundle_id IN (1, 2);

DELETE FROM products WHERE id IN (1, 2, 3, 4, 5, 6, 7);

DELETE FROM users WHERE id IN (1, 2);
DELETE FROM team WHERE id IN (1, 2);

SET session_replication_role = 'origin';
