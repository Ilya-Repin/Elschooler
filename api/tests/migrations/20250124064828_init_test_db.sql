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

INSERT INTO users (id, service) VALUES ('e5e79b0d-1e2c-49f0-97db-4930c9ea8c43', 'testsuite_student');
INSERT INTO users (id, service) VALUES ('d1b07c6e-9091-4f90-b13a-e3247245f1b5', 'testsuite_student');
INSERT INTO students (id, login, password) VALUES ('70d0ed1a-25e1-40f7-877d-9fb9ce28969f', 'existedStudent', 'existedPassword');
INSERT INTO user_students (user_id, student_id) VALUES ('d1b07c6e-9091-4f90-b13a-e3247245f1b5', '70d0ed1a-25e1-40f7-877d-9fb9ce28969f');


INSERT INTO users (id, service) VALUES ('531bce8d-43b9-4ac5-bbc3-e47977189bf1', 'testsuite_student');
INSERT INTO users (id, service) VALUES ('7e889d15-8dda-4d72-82ae-70d380b580ff', 'testsuite_student');

INSERT INTO students (id, login, password) VALUES ('62f486db-a0b8-4ce9-8101-d718bedea98f', 'testStudentUpdate1', 'testPasswordUpdate1');
INSERT INTO user_students (user_id, student_id) VALUES ('7e889d15-8dda-4d72-82ae-70d380b580ff', '62f486db-a0b8-4ce9-8101-d718bedea98f');
INSERT INTO students (id, login, password) VALUES ('103dfbac-e79e-42ff-aef7-5627d1badc63', 'testStudentUpdateExisted', 'testPasswordUpdateExisted');
INSERT INTO user_students (user_id, student_id) VALUES ('531bce8d-43b9-4ac5-bbc3-e47977189bf1', '103dfbac-e79e-42ff-aef7-5627d1badc63');

INSERT INTO students (id, login, password) VALUES ('964810cf-984a-41aa-8050-9d3c6dab81bf', 'testStudentUpdate2', 'testPasswordUpdate2');
INSERT INTO user_students (user_id, student_id) VALUES ('531bce8d-43b9-4ac5-bbc3-e47977189bf1', '964810cf-984a-41aa-8050-9d3c6dab81bf');

INSERT INTO students (id, login, password) VALUES ('4aaa94b8-7c71-4371-a2b2-5a91a44e18fa', 'testStudentDelete', 'testPasswordDelete');
INSERT INTO user_students (user_id, student_id) VALUES ('531bce8d-43b9-4ac5-bbc3-e47977189bf1', '4aaa94b8-7c71-4371-a2b2-5a91a44e18fa');

INSERT INTO students (id, login, password) VALUES ('7000f792-8145-4157-af2c-149d26b647a5', 'popularStudentDelete', 'popularPasswordDelete');
INSERT INTO user_students (user_id, student_id) VALUES ('531bce8d-43b9-4ac5-bbc3-e47977189bf1', '7000f792-8145-4157-af2c-149d26b647a5');
INSERT INTO user_students (user_id, student_id) VALUES ('7e889d15-8dda-4d72-82ae-70d380b580ff', '7000f792-8145-4157-af2c-149d26b647a5');


INSERT INTO users (id, service) VALUES ('599ce889-11be-4abf-89fb-2940a5b3bfec', 'testsuite_marks');

INSERT INTO students (id, login, password) VALUES ('a7b7de0e-d637-41ef-a26a-3a02294ade44', 'testStudent', 'testPassword');
INSERT INTO user_students (user_id, student_id) VALUES ('599ce889-11be-4abf-89fb-2940a5b3bfec', 'a7b7de0e-d637-41ef-a26a-3a02294ade44');


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_students;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS students;
-- +goose StatementEnd
