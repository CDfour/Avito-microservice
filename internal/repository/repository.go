package repository

import (
	"context"
	"time"

	Err "Avito/internal/errors"
	"Avito/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type IRepository interface {
	Balance(userID uuid.UUID) (*model.User, error)
	Enrollment(user model.User, funds float64) error
	AddUser(user model.User) error
	Transfer(sender, recipient model.User, funds float64) error
	Order(user model.User, order model.Order) error
	GetOrder(orderID uuid.UUID) (*model.Order, error)
	OrderSuccess(order model.Order) error
	OrderFailed(user model.User, order model.Order) error
	Report(time.Time) ([]model.Report, error)
	History(userID uuid.UUID, limit, offset int) ([]model.History, error)
}

type repository struct {
	dbConnection *pgx.Conn
}

func NewRepository(dbConnection *pgx.Conn) (IRepository, error) {
	if dbConnection == nil {
		return nil, Err.ErrNoConnectionToDb
	}
	return &repository{dbConnection: dbConnection}, nil
}

func (r *repository) Balance(userID uuid.UUID) (*model.User, error) {
	logrus.Infoln("Starting repository.Balance")

	query := `SELECT id, balance, date_create, last_update
			  FROM public.user
			  WHERE id = $1;`
	u := user{}
	if err := r.dbConnection.QueryRow(context.Background(), query, userID).Scan(&u.id, &u.balance, &u.dateCreate, &u.lastUpdate); err != nil {
		logrus.Errorln("Scan: ", err)
		logrus.Infoln("Ending repository.Balance")
		return nil, err
	}

	logrus.Infoln("Ending repository.Balance")
	return &model.User{ID: u.id, Funds: u.balance, DateCreate: u.dateCreate, LastUpdate: u.lastUpdate}, nil
}

func (r *repository) AddUser(user model.User) error {
	logrus.Infoln("Starting repository.AddUser")

	tx, err := r.dbConnection.Begin(context.Background())
	if err != nil {
		logrus.Errorln("Begin: ", err)
		logrus.Infoln("Ending repository.AddUser")
		return err
	}

	query := `INSERT INTO public.user(id, balance, date_create, last_update)
			  VALUES
			  ($1, $2, $3, $4);`
	if _, err = tx.Exec(context.Background(), query, user.ID, user.Funds, user.DateCreate, user.LastUpdate); err != nil {
		logrus.Errorf("Exec %s: %s\n", user, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.AddUser")
		return err
	}

	query = `INSERT INTO public.accounting(user_id,service_name,date_create,funds)
		     VALUES
		     ($1, 'Replenished', $2, $3);`
	if _, err = tx.Exec(context.Background(), query, user.ID, user.LastUpdate, user.Funds); err != nil {
		logrus.Errorln("Exec %s: %s\n", user, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.AddUser")
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		logrus.Errorln("Commit: ", err)
	}

	logrus.Infoln("Ending repository.AddUser")
	return err
}

func (r *repository) Enrollment(user model.User, funds float64) error {
	logrus.Infoln("Starting repository.Enrollment")

	tx, err := r.dbConnection.Begin(context.Background())
	if err != nil {
		logrus.Errorln("Begin: ", err)
		logrus.Infoln("Ending repository.Enrollment")
		return err
	}

	query := `UPDATE public.user
			  SET balance = $1, last_update = $2
			  WHERE id = $3;`
	if _, err := tx.Exec(context.Background(), query, user.Funds, user.LastUpdate, user.ID); err != nil {
		logrus.Errorf("Exec %s: %s", user, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.Enrollment")
		return err
	}

	query = `INSERT INTO public.accounting(user_id,service_name,date_create,funds)
		     VALUES
		     ($1, 'Replenished', $2, $3);`
	if _, err = tx.Exec(context.Background(), query, user.ID, user.LastUpdate, funds); err != nil {
		logrus.Errorf("Exec %s: %s\n", user, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.Enrollment")
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		logrus.Errorln("Commit: ", err)
	}

	logrus.Infoln("Ending repository.Transfer")
	return err
}

func (r *repository) Transfer(sender, recipient model.User, funds float64) error {
	logrus.Infoln("Starting repository.Transfer")

	tx, err := r.dbConnection.Begin(context.Background())
	if err != nil {
		logrus.Errorln("Begin: ", err)
		logrus.Infoln("Ending repository.Transfer")
		return err
	}

	query := `UPDATE public.user
			  SET balance = $1, last_update = $2
			  WHERE id = $3;`
	if _, err := tx.Exec(context.Background(), query, sender.Funds, sender.LastUpdate, sender.ID); err != nil {
		logrus.Errorf("Exec %s: %s", sender, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.Transfer")
		return err
	}

	query = `INSERT INTO public.accounting(user_id, service_name, date_create, funds)
		     VALUES
		     ($1, 'Transferred', $2, $3);`
	if _, err = tx.Exec(context.Background(), query, sender.ID, sender.LastUpdate, funds); err != nil {
		logrus.Errorf("Exec %s %s: %s\n", sender, funds, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.Transfer")
		return err
	}

	query = `UPDATE public.user
			 SET balance = $1, last_update = $2
			 WHERE id = $3;`
	if _, err := tx.Exec(context.Background(), query, recipient.Funds, recipient.LastUpdate, recipient.ID); err != nil {
		logrus.Errorf("Exec %s: %s", sender, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.Transfer")
		return err
	}

	query = `INSERT INTO public.accounting(user_id, service_name, date_create, funds)
		     VALUES
		     ($1, 'Replenished', $2, $3);`
	if _, err = tx.Exec(context.Background(), query, recipient.ID, recipient.LastUpdate, funds); err != nil {
		logrus.Errorf("Exec %s %s: %s\n", sender, funds, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.Transfer")
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		logrus.Errorln("Commit: ", err)
	}

	logrus.Infoln("Ending repository.Transfer")
	return err
}

func (r *repository) Order(user model.User, order model.Order) error {
	logrus.Infoln("Starting repository.Order")

	tx, err := r.dbConnection.Begin(context.Background())
	if err != nil {
		logrus.Errorln("Begin: ", err)
		logrus.Infoln("Ending repository.Order")
		return err
	}

	query := `UPDATE public.user
			  SET balance = $1, last_update = $2
			  WHERE id = $3`
	if _, err := tx.Exec(context.Background(), query, user.Funds, user.LastUpdate, user.ID); err != nil {
		logrus.Errorf("Exec %s: %s\n", user, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.Order")
		return err
	}

	query = `INSERT INTO public.order
		     VALUES
		     ($1, $2, $3, $4, $5, $6);`
	if _, err := tx.Exec(context.Background(), query, order.ID, order.UserID, order.ServiceID, order.ServiceName, order.DateCreate, order.Funds); err != nil {
		logrus.Errorf("Exec %s: %s\n", order, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.Order")
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		logrus.Errorln("Commit: ", err)
	}

	logrus.Infoln("Ending repository.Order")
	return err
}

func (r *repository) GetOrder(orderID uuid.UUID) (*model.Order, error) {
	logrus.Infoln("Starting repository.GetOrder")

	query := `SELECT order_id, user_id, service_id, service_name, date_create, funds
			  FROM public.order
			  WHERE id = $1;`
	o := order{}
	if err := r.dbConnection.QueryRow(context.Background(), query, orderID).Scan(&o.id, &o.userID, &o.serviceID, &o.serviceName, &o.dateCreate, &o.funds); err != nil {
		logrus.Errorf("Scan %s, %s\n", orderID, err)
		logrus.Infoln("Ending repository.GetOrder")
		return nil, err
	}

	return &model.Order{ID: o.id, UserID: o.userID, ServiceID: o.serviceID, ServiceName: o.serviceName, DateCreate: o.dateCreate, Funds: o.funds}, nil
}

func (r *repository) OrderSuccess(order model.Order) error {
	logrus.Infoln("Starting repository.OrderSuccess")

	tx, err := r.dbConnection.Begin(context.Background())
	if err != nil {
		logrus.Errorln("Begin: ", err)
		logrus.Infoln("Ending repository.OrderSuccess")
		return err
	}

	query := `INSERT INTO public.accounting
			  VALUES
			  ($1 ,$2, $3, $4, $5, $6);`
	if _, err := tx.Exec(context.Background(), query, order.ID, order.UserID, order.ServiceID, order.ServiceName, order.DateCreate, order.Funds); err != nil {
		logrus.Errorf("Exec %s: %s\n", order, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.OrderSuccess")
		return err
	}

	query = `DELETE FROM public.order
			 WHERE id = $1;`
	if _, err := tx.Exec(context.Background(), query, order.ID); err != nil {
		logrus.Errorf("Exec %s: %s\n", order, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.OrderSuccess")
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		logrus.Errorln("Commit: ", err)
	}

	logrus.Infoln("Ending repository.OrderSuccess")
	return err
}

func (r *repository) OrderFailed(user model.User, order model.Order) error {
	logrus.Infoln("Starting repository.OrderFailed")

	tx, err := r.dbConnection.Begin(context.Background())
	if err != nil {
		logrus.Errorln("Begin: ", err)
		logrus.Infoln("Ending repository.OrderFailed")
		return err
	}

	query := `UPDATE puiblic.user
			  SET balance = $1, last_update = $2
			  WHERE id = $3;`
	if _, err := tx.Exec(context.Background(), query, user.Funds, user.LastUpdate, user.ID); err != nil {
		logrus.Errorf("Exec %s: %s\n", user, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.OrderFailed")
		return err
	}

	query = `DELETE FROM public.order
			 WHERE id = $1;`
	if _, err := tx.Exec(context.Background(), query, order.ID); err != nil {
		logrus.Errorf("Exec %s: %s\n", order, err)
		if err := tx.Rollback(context.Background()); err != nil {
			logrus.Errorln("Rollback: ", err)
		}
		logrus.Infoln("Ending repository.OrderFailed")
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		logrus.Errorln("Commit: ", err)
	}

	logrus.Infoln("Ending repository.OrderFailed")
	return err
}

func (r *repository) Report(t time.Time) (report []model.Report, err error) {
	logrus.Infoln("Starting repository.Report")

	query := `SELECT public.accounting.service_name, SUM(public.accounting.funds)
			  FROM public.accounting
			  WHERE date_part('year', public.accounting.date_create) = date_part('year', date($1)) AND date_part('month', public.accounting.date_create) = date_part('month', date($1)) AND public.accounting.service_id IS NOT NULL
			  GROUP BY public.accounting.service_name
			  ORDER BY SUM(public.accounting.funds) DESC;`

	rows, err := r.dbConnection.Query(context.Background(), query, t)
	if err != nil {
		logrus.Errorf("Query %s: %s\n", t, err)
		logrus.Infoln("Ending repository.Report")
		return nil, err
	}

	for rows.Next() {
		r := model.Report{}
		if err := rows.Scan(&r.ServiceName, &r.Revenue); err != nil {
			logrus.Errorln("Scan: ", err)
			logrus.Infoln("Ending repository.Report")
			return nil, err
		}
		report = append(report, r)
	}

	logrus.Infoln("Ending repository.Report")
	return report, nil
}

func (r *repository) History(userID uuid.UUID, limit, offset int) ([]model.History, error) {
	logrus.Infoln("Starting repository.History")

	query := `SELECT public.accounting.user_id,public.accounting.service_name, public.accounting.funds, public.accounting.date_create
			  FROM public.accounting			 
			  WHERE public.accounting.user_id = $1
			  ORDER BY public.accounting.funds DESC, public.accounting.date_create 
			  LIMIT $2 OFFSET $3;`
	rows, err := r.dbConnection.Query(context.Background(), query, userID, limit, offset)
	if err != nil {
		logrus.Errorf("Query %s: %s\n", userID, err)
		logrus.Infoln("Ending repository.History")
		return nil, err
	}

	report := []model.History{}

	for rows.Next() {
		h := history{}
		if err := rows.Scan(&h.id, &h.serviceName, &h.cost, &h.date); err != nil {
			logrus.Errorln("Scan: ", err)
			logrus.Infoln("Ending repository.History")
			return nil, err
		}
		report = append(report, model.History{UserID: h.id, ServiceName: h.serviceName, Cost: h.cost, OrderDate: h.date})
	}

	logrus.Infoln("Ending repository.History")
	return report, nil
}
