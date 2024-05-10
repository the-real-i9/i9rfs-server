--
-- PostgreSQL database dump
--

-- Dumped from database version 16.1
-- Dumped by pg_dump version 16.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: i9rfs_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.i9rfs_user (
    id integer NOT NULL,
    email character varying NOT NULL,
    username character varying NOT NULL,
    password character varying NOT NULL
);


ALTER TABLE public.i9rfs_user OWNER TO postgres;

--
-- Name: i9rfs_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.i9rfs_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.i9rfs_user_id_seq OWNER TO postgres;

--
-- Name: i9rfs_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.i9rfs_user_id_seq OWNED BY public.i9rfs_user.id;


--
-- Name: ongoing_signup; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ongoing_signup (
    session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying NOT NULL,
    verification_code integer NOT NULL,
    verified boolean DEFAULT false NOT NULL
);


ALTER TABLE public.ongoing_signup OWNER TO postgres;

--
-- Name: i9rfs_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9rfs_user ALTER COLUMN id SET DEFAULT nextval('public.i9rfs_user_id_seq'::regclass);


--
-- Name: i9rfs_user i9rfs_user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9rfs_user
    ADD CONSTRAINT i9rfs_user_pkey PRIMARY KEY (id);


--
-- Name: ongoing_signup ongoing_signup_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ongoing_signup
    ADD CONSTRAINT ongoing_signup_pkey PRIMARY KEY (session_id);


--
-- PostgreSQL database dump complete
--

