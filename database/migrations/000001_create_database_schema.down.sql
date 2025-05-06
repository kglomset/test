-- Drop tables in reverse order to avoid foreign key constraints
DROP TABLE IF EXISTS public.product_bundles;
DROP TABLE IF EXISTS public.test_ranks;
DROP TABLE IF EXISTS public.tests;
DROP TABLE IF EXISTS public.snow_conditions;
DROP TABLE IF EXISTS public.track_conditions;
DROP TABLE IF EXISTS public.air_conditions;
DROP TABLE IF EXISTS public.products;
DROP TABLE IF EXISTS public.sessions;
DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.team;

-- Drop sequences
DROP SEQUENCE IF EXISTS public.air_conditions_id_seq;
DROP SEQUENCE IF EXISTS public.products_id_seq;
DROP SEQUENCE IF EXISTS public.sessions_id_seq;
DROP SEQUENCE IF EXISTS public.snow_conditions_id_seq;
DROP SEQUENCE IF EXISTS public.team_id_seq;
DROP SEQUENCE IF EXISTS public.tests_id_seq;
DROP SEQUENCE IF EXISTS public.track_conditions_id_seq;
DROP SEQUENCE IF EXISTS public.users_id_seq;

-- Drop types
DROP TYPE IF EXISTS public.wind_levels;
DROP TYPE IF EXISTS public.cloud_level;
DROP TYPE IF EXISTS public.snow_type_level;
DROP TYPE IF EXISTS public.humidity_levels;
DROP TYPE IF EXISTS public.hardness_level;
DROP TYPE IF EXISTS public.type_level;
DROP TYPE IF EXISTS public.product_type;
DROP TYPE IF EXISTS public.statustype;

-- Drop schema
DROP SCHEMA IF EXISTS public CASCADE;
