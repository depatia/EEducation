-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS service.grades
(
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),
    student_id integer NOT NULL,
    lesson_name character varying COLLATE pg_catalog."default" NOT NULL,
    grade integer,
    date date NOT NULL,
    is_term boolean,
    CONSTRAINT marks_pk PRIMARY KEY (id),
    CONSTRAINT grades_unique UNIQUE (student_id, lesson_name, grade, date, is_term)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS service.grades;
-- +goose StatementEnd
