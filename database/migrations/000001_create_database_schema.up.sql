--
-- PostgreSQL database dump
--

-- Dumped from database version 17.2 (Debian 17.2-1.pgdg120+1)
-- Dumped by pg_dump version 17.2

-- Started on 2025-03-26 09:12:34

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


CREATE TYPE public.user_role_type AS ENUM (
    'admin',
    'member'
    );


ALTER TYPE public.user_role_type OWNER TO postgres;


--
-- TOC entry 908 (class 1247 OID 22440)
-- Name: cloud_level; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.cloud_level AS ENUM (
    '1',
    '2',
    '3',
    '4'
);


ALTER TYPE public.cloud_level OWNER TO postgres;

--
-- TOC entry 896 (class 1247 OID 22396)
-- Name: hardness_level; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.hardness_level AS ENUM (
    'H1',
    'H2',
    'H3',
    'H4',
    'H5',
    'H6'
);


ALTER TYPE public.hardness_level OWNER TO postgres;

--
-- TOC entry 887 (class 1247 OID 22353)
-- Name: humidity_levels; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.humidity_levels AS ENUM (
    'DS',
    'W1',
    'W2',
    'W3',
    'W4'
);


ALTER TYPE public.humidity_levels OWNER TO postgres;

--
-- Name: product_type
--

CREATE TYPE public.product_type AS ENUM (
    'powder',
    'solid',
    'spray',
    'liquid',
    'gel',
    'bundle'
);


ALTER TYPE public.product_type OWNER TO postgres;

--
-- TOC entry 890 (class 1247 OID 22364)
-- Name: snow_type_level; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.snow_type_level AS ENUM (
    'A1',
    'A2',
    'A3',
    'A4',
    'A5',
    'FS',
    'NS',
    'IN',
    'IT',
    'TR'
);


ALTER TYPE public.snow_type_level OWNER TO postgres;

--
-- TOC entry 920 (class 1247 OID 22542)
-- Name: statustype; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.statustype AS ENUM (
    'active',
    'discontinued',
    'development',
    'retired'
);


ALTER TYPE public.statustype OWNER TO postgres;

--
-- TOC entry 899 (class 1247 OID 22410)
-- Name: type_level; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.type_level AS ENUM (
    'T1',
    'T2',
    'D1',
    'D2'
);


ALTER TYPE public.type_level OWNER TO postgres;

--
-- TOC entry 905 (class 1247 OID 22430)
-- Name: wind_levels; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.wind_levels AS ENUM (
    'S',
    'L',
    'M',
    'ST'
);


ALTER TYPE public.wind_levels OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 231 (class 1259 OID 22450)
-- Name: air_conditions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.air_conditions (
    id bigint NOT NULL,
    temperature integer NOT NULL,
    humidity integer NOT NULL,
    wind public.wind_levels NOT NULL,
    cloud public.cloud_level NOT NULL
);


ALTER TABLE public.air_conditions OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 22449)
-- Name: air_conditions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.air_conditions ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.air_conditions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 225 (class 1259 OID 22338)
-- Name: products; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.products (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    brand character varying (150),
    ean_code character varying(50),
    image_url character varying(255),
    comment text,
    is_public boolean DEFAULT true NOT NULL,
    type public.product_type NOT NULL,
    high_temperature double precision,
    low_temperature double precision,
    testing_team bigint NOT NULL,
    version timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    status public.statustype DEFAULT 'active'::public.statustype NOT NULL
);


ALTER TABLE public.products OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 22337)
-- Name: products_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.products ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 235 (class 1259 OID 22471)
-- Name: test_ranks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.test_ranks (
    test_id bigint NOT NULL,
    product_id bigint NOT NULL,
    rank integer,
    distance_behind integer,
    version timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    is_rank_public boolean DEFAULT true NOT NULL
);


ALTER TABLE public.test_ranks OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 22323)
-- Name: sessions; Type: TABLE; Schema: public; Owner: postgres
--
CREATE TYPE public.session_status_type AS ENUM (
    'active',
    'expired'
    );


CREATE TABLE public.sessions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    session_token text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp without time zone NOT NULL,
    status public.session_status_type DEFAULT 'active'::public.session_status_type NOT NULL
);


ALTER TABLE public.sessions OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 22322)
-- Name: sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.sessions ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 227 (class 1259 OID 22386)
-- Name: snow_conditions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.snow_conditions (
    id bigint NOT NULL,
    temperature double precision NOT NULL,
    snow_type public.snow_type_level NOT NULL,
    snow_humidity public.humidity_levels NOT NULL
);


ALTER TABLE public.snow_conditions OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 22385)
-- Name: snow_conditions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.snow_conditions ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.snow_conditions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 219 (class 1259 OID 22304)
-- Name: team; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.team (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    team_role int NOT NULL DEFAULT 2
);


ALTER TABLE public.team OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 22303)
-- Name: team_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.team ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.team_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 233 (class 1259 OID 22460)
-- Name: tests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tests (
    id bigint NOT NULL,
    test_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    location character varying(50) NOT NULL,
    comment character varying(255),
    sc_id bigint NOT NULL,
    tc_id bigint NOT NULL,
    ac_id bigint NOT NULL,
    version timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    is_public boolean DEFAULT true NOT NULL,
    testing_team bigint
);


ALTER TABLE public.tests OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 22459)
-- Name: tests_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.tests ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.tests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 243 (class 1259 OID 22655)
-- Name: product_bundles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product_bundles (
    bundle_id bigint NOT NULL,
    product_id bigint NOT NULL,
    layer_no integer NOT NULL
);


ALTER TABLE public.product_bundles OWNER TO postgres;


--
-- TOC entry 229 (class 1259 OID 22420)
-- Name: track_conditions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.track_conditions (
    id bigint NOT NULL,
    track_hardness public.hardness_level NOT NULL,
    track_type public.type_level NOT NULL
);


ALTER TABLE public.track_conditions OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 22419)
-- Name: track_conditions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.track_conditions ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.track_conditions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 221 (class 1259 OID 22314)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    user_role public.user_role_type DEFAULT 'member'::public.user_role_type NOT NULL,
    team_id bigint NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 22313)
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.users ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 3322 (class 2606 OID 22454)
-- Name: air_conditions air_conditions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.air_conditions
    ADD CONSTRAINT air_conditions_pkey PRIMARY KEY (id);


--
-- TOC entry 3316 (class 2606 OID 22345)
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- TOC entry 3326 (class 2606 OID 22475)
-- Name: test_ranks test_ranks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.test_ranks
    ADD CONSTRAINT test_ranks_pkey PRIMARY KEY (test_id, product_id);


--
-- TOC entry 3312 (class 2606 OID 22330)
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- TOC entry 3314 (class 2606 OID 22332)
-- Name: sessions sessions_session_token_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_session_token_key UNIQUE (session_token);


--
-- TOC entry 3318 (class 2606 OID 22390)
-- Name: snow_conditions snow_conditions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.snow_conditions
    ADD CONSTRAINT snow_conditions_pkey PRIMARY KEY (id);


--
-- TOC entry 3307 (class 2606 OID 22308)
-- Name: team team_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team
    ADD CONSTRAINT team_pkey PRIMARY KEY (id);


--
-- TOC entry 3324 (class 2606 OID 22465)
-- Name: tests tests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tests
    ADD CONSTRAINT tests_pkey PRIMARY KEY (id);

-- Create the sequence for bundle_id
CREATE SEQUENCE bundle_id_seq
    START WITH 1
    INCREMENT BY 1;

--
-- TOC entry 3334 (class 2606 OID 22659)
-- Name: product_bundles product_bundles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_bundles
    ADD CONSTRAINT product_bundles_pkey PRIMARY KEY (bundle_id, product_id);



--
-- TOC entry 3320 (class 2606 OID 22424)
-- Name: track_conditions track_conditions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.track_conditions
    ADD CONSTRAINT track_conditions_pkey PRIMARY KEY (id);


--
-- TOC entry 3310 (class 2606 OID 22321)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 3308 (class 1259 OID 22687)
-- Name: unique_official_role; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX unique_official_role ON public.team USING btree (team_role) WHERE ((team_role)::text = 'official'::text);


--
-- TOC entry 3342 (class 2606 OID 22500)
-- Name: test_ranks fk_test_ranks_test; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.test_ranks
    ADD CONSTRAINT fk_test_ranks_test FOREIGN KEY (test_id) REFERENCES public.tests(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_test_ranks_products FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- TOC entry 3338 (class 2606 OID 22495)
-- Name: tests fk_test_air_conditions; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tests
    ADD CONSTRAINT fk_test_air_conditions FOREIGN KEY (ac_id) REFERENCES public.air_conditions(id) ON DELETE CASCADE;


--
-- TOC entry 3339 (class 2606 OID 22485)
-- Name: tests fk_test_snow_conditions; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tests
    ADD CONSTRAINT fk_test_snow_conditions FOREIGN KEY (sc_id) REFERENCES public.snow_conditions(id) ON DELETE CASCADE;


--
-- TOC entry 3340 (class 2606 OID 22490)
-- Name: tests fk_test_track_conditions; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tests
    ADD CONSTRAINT fk_test_track_conditions FOREIGN KEY (tc_id) REFERENCES public.track_conditions(id) ON DELETE CASCADE;


--
-- TOC entry 3337 (class 2606 OID 22519)
-- Name: products fk_testing_team; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_testing_team FOREIGN KEY (testing_team) REFERENCES public.team(id) ON DELETE CASCADE;


--
-- TOC entry 3341 (class 2606 OID 22579)
-- Name: tests fk_tests_team; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tests
    ADD CONSTRAINT fk_tests_team FOREIGN KEY (testing_team) REFERENCES public.team(id) ON DELETE CASCADE;


--
-- TOC entry 3335 (class 2606 OID 22480)
-- Name: users fk_user_enterprise; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_user_enterprise FOREIGN KEY (team_id) REFERENCES public.team(id) ON DELETE CASCADE;


--
-- TOC entry 3336 (class 2606 OID 22505)
-- Name: sessions fkey_session_users; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT fkey_session_users FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 3349 (class 2606 OID 22665)
-- Name: product_bundles product_bundles_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_bundles
    ADD CONSTRAINT product_bundles_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;




