INSERT INTO teams (id, teamname, token, country_code, is_hidden)
VALUES (0, "adminers", "0", "", TRUE);

INSERT INTO users (id, username, email, password_hash, icon_path, team_id, is_hidden, is_admin)
VALUES (0, "admin", "admin@example.com", "$2a$10$vs25bQ2vIy4FmGKXdohrF.HXW49xZ0qwuVoTqShbM/Z2cVKmbOOS6	", NULL, 0, TRUE, TRUE); -- adminpassword

INSERT INTO config (ctf_name, start_at, end_at, min_score, easy_solves, medium_solves, lock_second, lock_duration, lock_count)
VALUES ("zer0ptsctf", from_unixtime(0), cast('2038-01-01 00:00:00' AS DATETIME), 100, 100, 50, 0, 0, 999);
