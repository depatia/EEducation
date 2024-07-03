package mysql

import (
	"AuthService/internal/models"
	"AuthService/internal/storage/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type StDb struct {
	db *sql.DB
}

func New(path string) (*StDb, error) {
	db, err := sql.Open("mysql", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open db due to error: %w", err)
	}

	return &StDb{db: db}, nil
}

func (s *StDb) CreateUser(ctx context.Context, email string, hash []byte) (int64, error) {
	stmt, err := s.db.Prepare("INSERT INTO users(email, pass_hash) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement due to error: %w", err)
	}

	res, err := stmt.ExecContext(ctx, email, hash)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				return 0, fmt.Errorf("failed to create user: %w", storage.ErrUserExists)
			}
		}

		return 0, fmt.Errorf("failed to create user due to error: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get ID due to error: %w", err)
	}

	return id, nil
}

func (s *StDb) GetUser(ctx context.Context, email string) (models.User, error) {
	stmt, err := s.db.Prepare("SELECT id, email, pass_hash, permission_level FROM users WHERE email = ?")
	if err != nil {
		return models.User{}, err
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash, &user.PermissionLevel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}

		return models.User{}, err
	}

	return user, nil
}

func (s *StDb) UpdatePassword(ctx context.Context, userID int64, passHash []byte) error {
	stmt, err := s.db.Prepare("UPDATE users SET `pass_hash` = ? WHERE id = ?")
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, passHash, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *StDb) SetPermission(ctx context.Context, userID int64, permissionLevel int64) error {
	count := 0
	stmt, err := s.db.Prepare("UPDATE users SET `permission_level` = ? WHERE id = ?")
	if err != nil {
		return err
	}
	q := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID)
	q.Scan(&count)

	_, err = stmt.ExecContext(ctx, permissionLevel, userID)
	if err != nil {
		return err
	}
	if count == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

func (s *StDb) GetPermission(ctx context.Context, userID int64) (int64, error) {
	stmt, err := s.db.Prepare("SELECT permission_level FROM users WHERE id = ?")
	if err != nil {
		return 0, err
	}

	row := stmt.QueryRowContext(ctx, userID)
	var permissionLevel int64
	err = row.Scan(&permissionLevel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, storage.ErrUserNotFound
		}

		return 0, err
	}

	return permissionLevel, nil
}

func (s *StDb) FillUserInfo(ctx context.Context, user models.UserInfo) error {
	count := 0
	stmt, err := s.db.Prepare("UPDATE users SET `name` = ?, `lastname` = ?, `middlename` = ?, `date_of_birth` = ?, `classname` = ? WHERE id = ? AND is_active = 0")
	if err != nil {
		return err
	}

	q := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", user.ID)
	q.Scan(&count)

	_, err = stmt.ExecContext(ctx, user.Name, user.Lastname, user.Middlename, user.DateOfBirth, user.Classname, user.ID)
	if err != nil {
		return err
	}
	if count == 0 {
		return storage.ErrUserNotFound
	}
	return nil
}

func (s *StDb) ChangeStatus(ctx context.Context, userID int64, isActive bool) error {
	count := 0

	stmt, err := s.db.Prepare("UPDATE users SET `is_active` = ? WHERE id = ?")
	if err != nil {
		return err
	}

	q := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID)
	q.Scan(&count)

	_, err = stmt.ExecContext(ctx, isActive, userID)
	if err != nil {
		return err
	}
	if count == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

func (s *StDb) IsActive(ctx context.Context, userID int64) (bool, error) {
	stmt, err := s.db.Prepare("SELECT is_active FROM users WHERE id = ?")
	if err != nil {
		return false, err
	}

	row := stmt.QueryRowContext(ctx, userID)
	var isActive bool
	err = row.Scan(&isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrUserNotFound
		}

		return false, err
	}

	return isActive, nil
}

func (s *StDb) GetStudentsByClass(ctx context.Context, classname string) ([]*models.UserDTO, error) {
	stmt, err := s.db.Prepare("SELECT name, lastname FROM users WHERE classname = ?")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, classname)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*models.UserDTO
	for rows.Next() {
		user := new(models.UserDTO)
		err = rows.Scan(&user.Name, &user.Lastname)
		if err != nil {
			return nil, fmt.Errorf("failed to scanning rows due to error: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (s *StDb) DelUser(ctx context.Context, userID int64) error {
	count := 0
	stmt, err := s.db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		return err
	}

	q := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID)
	q.Scan(&count)

	_, err = stmt.ExecContext(ctx, userID)
	if err != nil {
		return err
	}
	if count == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

func (s *StDb) Stop() {
	s.db.Close()
}
