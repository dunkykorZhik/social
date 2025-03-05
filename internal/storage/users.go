package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	Is_Active bool     `json:"is_active"`
	Role_id   int64    `json:"role_id"`
}

type password struct {
	hash []byte
	text *string
}

type UserStorage struct {
	db *sql.DB
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.hash = hash
	p.text = &text

	return nil
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))

}

func (u *UserStorage) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {

		if err := u.create(ctx, tx, user); err != nil {
			return err
		}

		if err := u.createInvitation(ctx, tx, token, exp, user.ID); err != nil {
			return err
		}

		return nil

	})
}

func (u *UserStorage) Activate(ctx context.Context, token string) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		userId, err := u.getUserFromToken(ctx, tx, token)
		if err != nil {
			return err
		}
		if err := u.updateUserActivity(ctx, tx, userId); err != nil {
			return err
		}
		if err := u.deleteInvitation(ctx, tx, userId); err != nil {
			return err
		}

		return nil
	})
}

func (u *UserStorage) Delete(ctx context.Context, id int64) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		if err := u.deleteInvitation(ctx, tx, id); err != nil {
			return err
		}
		if err := u.deleteUser(ctx, tx, id); err != nil {
			return err
		}
		return nil
	})
}

func (u *UserStorage) GetByID(ctx context.Context, userId int64) (*User, error) {
	var user User
	query := `SELECT id, username, email, password, created_at, is_active, role_id FROM users WHERE id = $1 AND is_active = TRUE;`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := u.db.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Is_Active,
		&user.Role_id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}

	}

	return &user, nil
}

func (u *UserStorage) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `SELECT id, username, email, password, created_at, is_active FROM users WHERE email = $1 AND is_active = TRUE;`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := u.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Is_Active)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}

	}

	return &user, nil
}

func (u *UserStorage) Follow(ctx context.Context, fId int64, userId int64) error {
	query := `INSERT INTO followers (user_id, follower_id) VALUES ($1,$2)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)

	defer cancel()

	_, err := u.db.ExecContext(ctx, query, fId, userId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}
	return nil

}

func (u *UserStorage) UnFollow(ctx context.Context, fId int64, userId int64) error {
	query := `DELETE FROM followers WHERE user_id=$1 AND follower_id=$2`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	res, err := u.db.ExecContext(ctx, query, fId, userId)
	if err != nil {
		return err
	}

	rowsCount, err := res.RowsAffected()
	if err != nil {

		return err
	}
	if rowsCount == 0 {
		return ErrNotFound
	} else if rowsCount != 1 {
		return ErrTooMuchChanged
	}
	return nil

}

func (u *UserStorage) GetUserFeed(ctx context.Context, user_id int64, pagQ PaginateQuery) ([]PostForFeed, error) {
	query := `
	SELECT 
    	p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
    	u.username, COUNT(c.id) AS comments_count
	FROM posts p
	LEFT JOIN comments c ON c.post_id = p.id
	LEFT JOIN users u ON p.user_id = u.id
	LEFT JOIN followers f ON f.user_id = p.user_id AND f.follower_id = $1
	WHERE (p.user_id = $1 OR f.user_id IS NOT NULL) 
    AND (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') 
    AND (p.tags @> $5 OR array_length($5, 1) = 0) 
	GROUP BY p.id, u.username
	ORDER BY p.created_at ` + pagQ.Sort + `
	LIMIT $2 OFFSET $3;
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	rows, err := u.db.QueryContext(ctx, query, user_id, pagQ.Limit, pagQ.Offset, pagQ.Search, pagQ.Tags)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feed []PostForFeed
	for rows.Next() {
		var p PostForFeed
		err := rows.Scan(
			&p.Post.ID,
			&p.Post.UserID,
			&p.Post.Title,
			&p.Post.Content,
			&p.Post.CreatedAt,
			&p.Post.Version,
			pq.Array(&p.Post.Tags),
			&p.Post.User.Username,
			&p.CommentCount)
		if err != nil {
			return nil, err
		}
		feed = append(feed, p)

	}

	return feed, nil
}

func (u *UserStorage) create(ctx context.Context, tx *sql.Tx, user *User) error {

	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id, created_at;"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := tx.QueryRowContext(ctx, query, user.Username, user.Email, user.Password.hash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}
func (u *UserStorage) deleteUser(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
func (u *UserStorage) deleteInvitation(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}
	return nil

}

func (u *UserStorage) updateUserActivity(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `UPDATE users SET is_active = TRUE WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil

}

func (u *UserStorage) getUserFromToken(ctx context.Context, tx *sql.Tx, token string) (int64, error) {
	query := `SELECT u.id, u.is_active
		FROM users u
		JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	user := &User{}
	if err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(&user.ID, &user.Is_Active); err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, err
	}

	return user.ID, nil

}

func (u *UserStorage) createInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userId int64) error {
	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userId, time.Now().Add(exp))
	if err != nil {
		return nil
	}

	return nil
}
