SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

DROP DATABASE IF EXISTS dappctrl;

CREATE DATABASE dappctrl WITH TEMPLATE = template0 ENCODING = 'UTF8'

ALTER DATABASE dappctrl OWNER TO postgres;
