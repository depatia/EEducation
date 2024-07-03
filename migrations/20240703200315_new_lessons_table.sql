-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS service.lessons
(
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),
    lesson_name character varying COLLATE pg_catalog."default" NOT NULL,
    teacher_id integer NOT NULL,
    classname character varying(3) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT lessons_pk PRIMARY KEY (id),
    CONSTRAINT lessons_unique UNIQUE (lesson_name, teacher_id, classname)
)-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS service.lessons;
-- +goose StatementEnd
