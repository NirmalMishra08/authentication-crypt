-- name: GetUserByUserName :one
select * from users
WHERE username = $1;
