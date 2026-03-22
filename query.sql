-- name: GetUserByUserName :one
select * from users
WHERE username = $1;

-- name: InsertUser :exec
INSERT into users (username, password) values ($1, $2);


