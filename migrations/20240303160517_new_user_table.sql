-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id        INT(11) NOT NULL PRIMARY KEY AUTO_INCREMENT,
    email     VARCHAR(255) NOT NULL UNIQUE,
    pass_hash BLOB NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
