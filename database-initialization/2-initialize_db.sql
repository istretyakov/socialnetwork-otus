BEGIN;

DROP TABLE IF EXISTS public.users;

CREATE TABLE IF NOT EXISTS public.users
(
    id uuid,
    first_name character varying(255) COLLATE pg_catalog."default",
    second_name character varying(255) COLLATE pg_catalog."default",
    biography text COLLATE pg_catalog."default",
    city character varying(255) COLLATE pg_catalog."default",
    birthdate date,
    password character varying(255) COLLATE pg_catalog."default"
)
TABLESPACE pg_default;

COMMIT;

ALTER TABLE IF EXISTS public.users
    OWNER to postgres;