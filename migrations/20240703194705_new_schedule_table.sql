-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS service.schedule
(
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),
    date date NOT NULL,
    lesson_name character varying COLLATE pg_catalog."default" NOT NULL,
    classroom integer NOT NULL,
    classname character varying(3) COLLATE pg_catalog."default" NOT NULL,
    grade integer,
    homework character varying COLLATE pg_catalog."default",
    CONSTRAINT schedule_pk PRIMARY KEY (id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS service.schedule;
-- +goose StatementEnd
