package controller

import (
	Err "Avito/internal/errors"
	"Avito/internal/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestController_Balance(t *testing.T) {
	mRepo := NewIRepositoryMock(t)

	c, err := NewController(mRepo)
	require.NoError(t, err)

	t.Run("failed", func(t *testing.T) {
		mRepo.BalanceMock.Return(nil, pgx.ErrNoRows)

		res, err := c.Balance(uuid.New())
		require.ErrorIs(t, err, pgx.ErrNoRows)
		require.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		m := &model.User{
			ID:         uuid.New(),
			Funds:      10,
			DateCreate: time.Now(),
			LastUpdate: time.Now(),
		}
		mRepo.BalanceMock.Return(m, nil)

		res, err := c.Balance(uuid.New())
		require.NoError(t, err)
		require.Equal(t, res, m)
	})
}

func TestController_Transfer(t *testing.T) {
	mRepo := NewIRepositoryMock(t)

	c, err := NewController(mRepo)
	require.NoError(t, err)

	t.Run("failed", func(t *testing.T) {
		sender := &model.User{
			ID:         uuid.New(),
			Funds:      10,
			DateCreate: time.Time{},
			LastUpdate: time.Time{},
		}
		receiver := &model.User{
			ID:         uuid.New(),
			Funds:      0,
			DateCreate: time.Time{},
			LastUpdate: time.Time{},
		}

		mRepo.BalanceMock.Set(func(userID uuid.UUID) (up1 *model.User, err error) {
			if userID == sender.ID {
				return sender, nil
			}
			return receiver, nil
		})

		err := c.Transfer(sender.ID, receiver.ID, 1000)
		require.ErrorIs(t, err, Err.ErrInsufficientFunds)
	})

	t.Run("success", func(t *testing.T) {
		sender := &model.User{
			ID:         uuid.New(),
			Funds:      10,
			DateCreate: time.Time{},
			LastUpdate: time.Time{},
		}
		receiver := &model.User{
			ID:         uuid.New(),
			Funds:      0,
			DateCreate: time.Time{},
			LastUpdate: time.Time{},
		}

		mRepo.BalanceMock.Set(func(userID uuid.UUID) (up1 *model.User, err error) {
			if userID == sender.ID {
				return sender, nil
			}
			return receiver, nil
		})

		mRepo.TransferMock.Return(nil)

		err := c.Transfer(sender.ID, receiver.ID, 5)
		require.NoError(t, err)
	})
}

func TestController_Report(t *testing.T) {
	mRepo := NewIRepositoryMock(t)

	c, err := NewController(mRepo)
	require.NoError(t, err)

	t.Run("failed", func(t *testing.T) {
		res, err := c.Report("wrong year", "wrong month")
		require.ErrorIs(t, err, Err.ErrBadRequest)
		require.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		reports := []model.Report{
			{
				ServiceName: uuid.New().String(),
				Revenue:     1000,
			},
		}
		mRepo.ReportMock.Return(reports, nil)

		res, err := c.Report("2022", "05")
		require.NoError(t, err)
		require.NotEmpty(t, res)
	})
}

func TestController_Enrollment(t *testing.T) {
	mRepo := NewIRepositoryMock(t)

	c, err := NewController(mRepo)
	require.NoError(t, err)

	t.Run("success: add new user", func(t *testing.T) {
		m := &model.User{
			ID:         uuid.New(),
			Funds:      10,
			DateCreate: time.Time{},
			LastUpdate: time.Time{},
		}

		mRepo.BalanceMock.Return(nil, pgx.ErrNoRows)

		mRepo.AddUserMock.Set(func(user model.User) (err error) {
			require.Equal(t, m.ID, user.ID)
			require.Equal(t, m.Funds, user.Funds)

			return nil
		})

		err := c.Enrollment(m.ID, m.Funds)
		require.NoError(t, err)
	})

	t.Run("success: enrollment balance", func(t *testing.T) {
		var funds float64 = 10

		m := &model.User{
			ID:         uuid.New(),
			Funds:      10,
			DateCreate: time.Time{},
			LastUpdate: time.Time{},
		}

		mRepo.BalanceMock.Return(m, nil)

		mRepo.EnrollmentMock.Set(func(user model.User, funds float64) (err error) {

			require.Equal(t, m.ID, user.ID)
			require.Equal(t, m.Funds+funds, funds)

			return nil
		})

		err := c.Enrollment(m.ID, funds)
		require.NoError(t, err)
	})
}

func TestController_Order(t *testing.T) {
	mRepo := NewIRepositoryMock(t)

	c, err := NewController(mRepo)
	require.NoError(t, err)

	t.Run("failed", func(t *testing.T) {
		m := &model.User{
			ID:         uuid.New(),
			Funds:      10,
			DateCreate: time.Time{},
			LastUpdate: time.Time{},
		}

		mRepo.BalanceMock.Return(m, nil)

		err := c.Order(m.ID, uuid.New(), uuid.New(), uuid.New().String(), 100)
		require.ErrorIs(t, err, Err.ErrInsufficientFunds)
	})

	t.Run("success", func(t *testing.T) {
		m := &model.User{
			ID:         uuid.New(),
			Funds:      1000,
			DateCreate: time.Time{},
			LastUpdate: time.Time{},
		}

		mRepo.BalanceMock.Return(m, nil)

		err := c.Order(m.ID, uuid.New(), uuid.New(), uuid.New().String(), 100)
		require.NoError(t, err)
	})
}
