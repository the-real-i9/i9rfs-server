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
	username character varying,
	password character varying
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
-- Name: find_user_by_email_or_username(character varying); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.find_user_by_email_or_username(email_or_username character varying) RETURNS SETOF public.i9rfs_user_t
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT id, username, password FROM i9rfs_user 
  WHERE email_or_username = ANY(ARRAY[email, username]);
  
  RETURN;
END;
$$;


ALTER FUNCTION public.find_user_by_email_or_username(email_or_username character varying) OWNER TO i9;

--
-- Name: find_user_by_id(uuid); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.find_user_by_id(user_id uuid) RETURNS SETOF public.i9rfs_user_t
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT id, username, password FROM i9rfs_user 
  WHERE user_id = id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.find_user_by_id(user_id uuid) OWNER TO i9;

--
-- Name: mkdir(text, text[], uuid); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.mkdir(in_parent_dir_path text, new_dir_tree text[], user_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
  parent_dir_id uuid;
  parent_dir_path text;

  new_dir_node text;
BEGIN
  -- retrieve the parent directory's id and path from the database
  -- the parent directory is one whose path is in_parent_dir_path
  -- if the in_parent_dir_path is "/", we won't find a parent directory, and parent_dir_* values above will be empty,
  -- this new directory, hence, will have no parent i.e it will conceptually be located at the root dir
  SELECT id, path INTO parent_dir_id, parent_dir_path 
  FROM fs_object 
  WHERE path = in_parent_dir_path;

  -- since the user is able to specify a directory path separated by "/" to create a directory (degenerate) tree
  -- each directory in the (degenerate) tree will be the parent of the next
  -- the first directory in the (degenerate) tree will have parent_dir(_id) above, as its parent
  FOREACH new_dir_node IN ARRAY new_dir_tree
  LOOP

    DECLARE
      new_dir_name text := trim('"' from new_dir_node);
	  new_dir_path text := concat(parent_dir_path, '/', new_dir_name);
	  new_dir_date timestamp := now();

	  existing_dir_id uuid;
	  existing_dir_path text;
	BEGIN
	  SELECT id, path INTO existing_dir_id, existing_dir_path 
	  FROM fs_object
	  WHERE path = new_dir_path;

	  -- if a directory along the tree path already exists, rather than raising an error, we just go ahead and use it,
      -- make it our parent_dir_* for the next directory in the tree and skip creating a duplicate.
	  -- otherwise, we create it
	  IF existing_dir_id IS NULL THEN
	    -- if we have no parent directory,
	    -- (i.e. our starting in_parent_dir_path is "/", and, of course, our parent_dir_* above values are empty)
        -- this new directory is going to be directly in the root, since
        -- its parent_directory_id attribute will be NULL and its path attribute will be '/new_dir_name'
	    -- otherwise, we give this new directory as a child to
        -- the previous directory in the tree, which is currently the parent
	    INSERT INTO fs_object (owner_user_id, parent_directory_id, path, object_type, properties)
	    VALUES (user_id, parent_dir_id, new_dir_path, 'directory', jsonb_build_object('name', new_dir_name, 'date_created', new_dir_date, 'date_modified', new_dir_date))
		-- setting this new directory to the parent of the next in the tree
	    RETURNING id, path INTO parent_dir_id, parent_dir_path;
      ELSE
	    -- setting this existing directory to the parent of the next in the tree
  	    parent_dir_id := existing_dir_id;
	    parent_dir_path := existing_dir_path;
	  END IF;
	END;
	
  END LOOP;

  RETURN true;
END;
$$;


ALTER FUNCTION public.mkdir(in_parent_dir_path text, new_dir_tree text[], user_id uuid) OWNER TO i9;

--
-- Name: new_user(character varying, character varying, character varying); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.new_user(in_email character varying, in_username character varying, in_password character varying) RETURNS SETOF public.i9rfs_user_t
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY INSERT INTO i9rfs_user (email, username, password) 
  VALUES (in_email, in_username, in_password)
  RETURNING id, username, password;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.new_user(in_email character varying, in_username character varying, in_password character varying) OWNER TO i9;

--
-- Name: rm(text, boolean); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.rm(fs_object_path text, rflag boolean, OUT status boolean, OUT err_msg text) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  fs_object_id uuid;
  fs_object_type text;
  fs_object_name text;
BEGIN

  SELECT id, object_type, properties ->> 'name' INTO fs_object_id, fs_object_type, fs_object_name FROM fs_object WHERE path = fs_object_path;
  
  -- if fs_object_path doesn't exist at all in fs objects, return error: no such file or directory
  IF fs_object_id IS NULL THEN
    status := false;
	err_msg := 'no such file or directory';
	
	RETURN;
  END IF;

  -- if fs_object_type is 'directory' AND the recursive flag is not set
  IF fs_object_type = 'directory' AND rflag = false THEN
    status := false;
	err_msg := concat('cannot remove ', quote_literal(fs_object_name), ': Is a directory');

	RETURN;
  END IF;

  -- if (fs_object type is 'directory' AND rflag is set) OR fs_object_type is 'file'
  -- actually, this is the only possible condition at this point so there's no need to check
  -- if fs object is a directory this will remove the entire tree (ON DELETE CASCADE)
  DELETE FROM fs_object WHERE id = fs_object_id;

  status := true;
  err_msg := '';

  RETURN;
END;
$$;


ALTER FUNCTION public.rm(fs_object_path text, rflag boolean, OUT status boolean, OUT err_msg text) OWNER TO i9;

--
-- Name: rmdir(text); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.rmdir(dir_path text, OUT status boolean, OUT err_msg text) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  fs_object_id uuid;
  fs_object_type text;
  fs_object_name text;
BEGIN

  SELECT id, object_type, properties ->> 'name' INTO fs_object_id, fs_object_type, fs_object_name FROM fs_object WHERE path = dir_path;
  
  -- if dir_path path doesn't exist at all in fs object, return error: no such file or directory
  IF fs_object_id IS NULL THEN
    status := false;
	err_msg := 'no such file or directory';
	
	RETURN;
  END IF;

  -- if fs object type is not a directory, return error: failed to remove '{object name}': Not a directory
  IF fs_object_type <> 'directory' THEN
    status := false;
	err_msg := concat('failed to remove ', quote_literal(fs_object_name), ': Not a directory');

	RETURN;
  END IF;

  -- if directory is the parent of any other fs object, return error: failed to remove '{object name}': Directory not empty
  IF EXISTS(SELECT 1 FROM fs_object WHERE parent_directory_id = fs_object_id) THEN
    status := false;
	err_msg := concat('failed to remove ', quote_literal(fs_object_name), ': Directory not empty');

	RETURN;
  END IF;

  -- remove directory
  DELETE FROM fs_object WHERE id = fs_object_id;

  status := true;
  err_msg := '';

  RETURN;
END;
$$;


ALTER FUNCTION public.rmdir(dir_path text, OUT status boolean, OUT err_msg text) OWNER TO i9;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: fs_object; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.fs_object (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_user_id uuid NOT NULL,
    parent_directory_id uuid,
    path text NOT NULL,
    object_type text NOT NULL,
    properties jsonb NOT NULL,
    CONSTRAINT fs_object_object_type_check CHECK ((object_type = ANY (ARRAY['directory'::text, 'file'::text])))
);


ALTER TABLE public.fs_object OWNER TO i9;

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
    session_data json NOT NULL
);


ALTER TABLE public.ongoing_signup OWNER TO i9;

--
-- Name: fs_object fs_object_path_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.fs_object
    ADD CONSTRAINT fs_object_path_key UNIQUE (path);


--
-- Name: fs_object fs_object_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.fs_object
    ADD CONSTRAINT fs_object_pkey PRIMARY KEY (id);


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
-- Name: fs_object fs_object_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.fs_object
    ADD CONSTRAINT fs_object_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.i9rfs_user(id) ON DELETE CASCADE;


--
-- Name: fs_object fs_object_parent_directory_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.fs_object
    ADD CONSTRAINT fs_object_parent_directory_id_fkey FOREIGN KEY (parent_directory_id) REFERENCES public.fs_object(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

