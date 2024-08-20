--
-- PostgreSQL database dump
--

-- Dumped from database version 16.3 (Ubuntu 16.3-1.pgdg22.04+1)
-- Dumped by pg_dump version 16.3 (Ubuntu 16.3-1.pgdg22.04+1)

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

--
-- Name: i9rfs_user_t; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.i9rfs_user_t AS (
	id integer,
	username character varying
);


ALTER TYPE public.i9rfs_user_t OWNER TO postgres;

--
-- Name: account_exists(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.account_exists(email_or_username character varying, OUT exist boolean) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT EXISTS(SELECT 1 FROM i9rfs_user WHERE email_or_username = ANY(ARRAY[email, username])) INTO exist;
END;
$$;


ALTER FUNCTION public.account_exists(email_or_username character varying, OUT exist boolean) OWNER TO postgres;

--
-- Name: end_signup_session(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.end_signup_session(in_session_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  DELETE FROM ongoing_signup 
  WHERE session_id = in_session_id;
  
  RETURN true;
END;
$$;


ALTER FUNCTION public.end_signup_session(in_session_id uuid) OWNER TO postgres;

--
-- Name: get_user(anycompatible); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user(unique_identifier anycompatible) RETURNS SETOF public.i9rfs_user_t
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT id, username FROM i9rfs_user 
  WHERE unique_identifier::varchar = ANY(ARRAY[id::varchar, email, username]);
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user(unique_identifier anycompatible) OWNER TO postgres;

--
-- Name: get_user_password(anycompatible); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user_password(unique_identifier anycompatible, OUT password character varying) RETURNS character varying
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT i9rfs_user.password FROM i9rfs_user 
  WHERE unique_identifier::varchar = ANY(ARRAY[id::varchar, email, username]) 
  INTO "password";
END;
$$;


ALTER FUNCTION public.get_user_password(unique_identifier anycompatible, OUT password character varying) OWNER TO postgres;

--
-- Name: new_signup_session(character varying, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.new_signup_session(in_email character varying, in_verification_code integer, OUT session_id uuid) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
BEGIN
  DELETE FROM ongoing_signup WHERE email = in_email;  
  
  INSERT INTO ongoing_signup (email, verification_code)
  VALUES (in_email, in_verification_code)
  RETURNING ongoing_signup.session_id INTO session_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.new_signup_session(in_email character varying, in_verification_code integer, OUT session_id uuid) OWNER TO postgres;

--
-- Name: new_user(character varying, character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.new_user(in_email character varying, in_username character varying, in_password character varying) RETURNS SETOF public.i9rfs_user_t
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY INSERT INTO i9rfs_user (email, username, password) 
  VALUES (in_email, in_username, in_password)
  RETURNING id, username;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.new_user(in_email character varying, in_username character varying, in_password character varying) OWNER TO postgres;

--
-- Name: verify_email(uuid, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.verify_email(in_session_id uuid, in_verf_code integer, OUT is_success boolean) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN    
  IF (SELECT verification_code FROM ongoing_signup WHERE session_id = in_session_id) = in_verf_code THEN
    UPDATE ongoing_signup SET verified = true 
        WHERE session_id = in_session_id;
    is_success := true;
  ELSE 
    is_success := false;
  END IF;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.verify_email(in_session_id uuid, in_verf_code integer, OUT is_success boolean) OWNER TO postgres;

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

