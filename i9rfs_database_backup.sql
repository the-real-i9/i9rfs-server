--
-- PostgreSQL database dump
--

-- Dumped from database version 17.0 (Ubuntu 17.0-1.pgdg24.04+1)
-- Dumped by pg_dump version 17.0 (Ubuntu 17.0-1.pgdg24.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: i9rfs_user_t; Type: TYPE; Schema: public; Owner: i9
--

CREATE TYPE public.i9rfs_user_t AS (
	id uuid,
	username character varying
);


ALTER TYPE public.i9rfs_user_t OWNER TO i9;

--
-- Name: account_exists(character varying); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.account_exists(email_or_username character varying, OUT exist boolean) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT EXISTS(SELECT 1 FROM i9rfs_user WHERE email_or_username = ANY(ARRAY[email, username])) INTO exist;
END;
$$;


ALTER FUNCTION public.account_exists(email_or_username character varying, OUT exist boolean) OWNER TO i9;

--
-- Name: end_signup_session(uuid); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.end_signup_session(in_session_id uuid) OWNER TO i9;

--
-- Name: get_user(uuid); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.get_user(user_id uuid) RETURNS SETOF public.i9rfs_user_t
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT id, username FROM i9rfs_user 
  WHERE user_id = id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user(user_id uuid) OWNER TO i9;

--
-- Name: get_user_password(uuid); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.get_user_password(user_id uuid, OUT password character varying) RETURNS character varying
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT i9rfs_user.password FROM i9rfs_user 
  WHERE user_id = id 
  INTO "password";
END;
$$;


ALTER FUNCTION public.get_user_password(user_id uuid, OUT password character varying) OWNER TO i9;

--
-- Name: new_signup_session(character varying, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.new_signup_session(in_email character varying, in_verification_code integer, OUT session_id uuid) OWNER TO i9;

--
-- Name: new_user(character varying, character varying, character varying); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.new_user(in_email character varying, in_username character varying, in_password character varying) OWNER TO i9;

--
-- Name: verify_email(uuid, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.verify_email(in_session_id uuid, in_verf_code integer, OUT is_success boolean) OWNER TO i9;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: directory; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.directory (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_user_id uuid NOT NULL,
    parent_directory_id uuid,
    path character varying NOT NULL,
    name character varying NOT NULL,
    date_modified timestamp without time zone DEFAULT now(),
    date_created timestamp without time zone DEFAULT now()
);


ALTER TABLE public.directory OWNER TO i9;

--
-- Name: file; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.file (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_user_id uuid NOT NULL,
    parent_directory_id uuid,
    path character varying NOT NULL,
    name character varying NOT NULL,
    type character varying NOT NULL,
    date_accessed timestamp without time zone DEFAULT now(),
    date_modified timestamp without time zone DEFAULT now(),
    date_created timestamp without time zone DEFAULT now()
);


ALTER TABLE public.file OWNER TO i9;

--
-- Name: i9rfs_user; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.i9rfs_user (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying NOT NULL,
    username character varying NOT NULL,
    password character varying NOT NULL
);


ALTER TABLE public.i9rfs_user OWNER TO i9;

--
-- Name: ongoing_signup; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.ongoing_signup (
    session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying NOT NULL,
    verification_code integer NOT NULL,
    verified boolean DEFAULT false NOT NULL
);


ALTER TABLE public.ongoing_signup OWNER TO i9;

--
-- Data for Name: directory; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.directory (id, owner_user_id, parent_directory_id, path, name, date_modified, date_created) FROM stdin;
\.


--
-- Data for Name: file; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.file (id, owner_user_id, parent_directory_id, path, name, type, date_accessed, date_modified, date_created) FROM stdin;
\.


--
-- Data for Name: i9rfs_user; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.i9rfs_user (id, email, username, password) FROM stdin;
b6f39c6f-1347-491a-b455-990bdc4c14f4	ken@gmail.com	ken	dode
\.


--
-- Data for Name: ongoing_signup; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.ongoing_signup (session_id, email, verification_code, verified) FROM stdin;
\.


--
-- Name: directory directory_path_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.directory
    ADD CONSTRAINT directory_path_key UNIQUE (path);


--
-- Name: directory directory_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.directory
    ADD CONSTRAINT directory_pkey PRIMARY KEY (id);


--
-- Name: file file_path_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.file
    ADD CONSTRAINT file_path_key UNIQUE (path);


--
-- Name: file file_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.file
    ADD CONSTRAINT file_pkey PRIMARY KEY (id);


--
-- Name: i9rfs_user i9rfs_user_email_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.i9rfs_user
    ADD CONSTRAINT i9rfs_user_email_key UNIQUE (email);


--
-- Name: i9rfs_user i9rfs_user_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.i9rfs_user
    ADD CONSTRAINT i9rfs_user_pkey PRIMARY KEY (id);


--
-- Name: i9rfs_user i9rfs_user_username_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.i9rfs_user
    ADD CONSTRAINT i9rfs_user_username_key UNIQUE (username);


--
-- Name: ongoing_signup ongoing_signup_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.ongoing_signup
    ADD CONSTRAINT ongoing_signup_pkey PRIMARY KEY (session_id);


--
-- Name: directory directory_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.directory
    ADD CONSTRAINT directory_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.i9rfs_user(id) ON DELETE CASCADE;


--
-- Name: directory directory_parent_directory_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.directory
    ADD CONSTRAINT directory_parent_directory_id_fkey FOREIGN KEY (parent_directory_id) REFERENCES public.directory(id) ON DELETE CASCADE;


--
-- Name: file file_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.file
    ADD CONSTRAINT file_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.i9rfs_user(id) ON DELETE CASCADE;


--
-- Name: file file_parent_directory_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.file
    ADD CONSTRAINT file_parent_directory_id_fkey FOREIGN KEY (parent_directory_id) REFERENCES public.directory(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

