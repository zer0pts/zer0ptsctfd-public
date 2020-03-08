CREATE TABLE login.users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(256) UNIQUE KEY NOT NULL,
    password VARCHAR(256) NOT NULL
)Engine=InnoDB DEFAULT CHARSET utf8mb4;

INSERT INTO login.users(username, password) VALUES ('admin', '<flag>');
