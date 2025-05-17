-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id UUID PRIMARY KEY,
    service TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
    CONSTRAINT users_service_len CHECK (char_length(service) <= 100)
);

CREATE TABLE students (
    id UUID PRIMARY KEY,
    login TEXT NOT NULL,
    password TEXT NOT NULL,
    CONSTRAINT students_login_len     CHECK (char_length(login)    <= 100),
    CONSTRAINT students_password_len  CHECK (char_length(password) <= 100),
    CONSTRAINT unique_login_password UNIQUE (login, password)
);

CREATE TABLE user_students (
    user_id UUID,
    student_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, student_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_students;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS students;
-- +goose StatementEnd
