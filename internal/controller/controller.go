package controller

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	Err "Avito/internal/errors"
	"Avito/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type IController interface {
	Balance(userID uuid.UUID) (*model.User, error)
	Enrollment(userID uuid.UUID, funds float64) error
	Transfer(senderID, recipientID uuid.UUID, funds float64) error
	Order(userID, serviceID, orderID uuid.UUID, serviceName string, cost float64) error
	OrderSuccess(userID, serviceID, orderID uuid.UUID, serviceName string, cost float64) error
	OrderFailed(userID, serviceID, orderID uuid.UUID, serviceName string, cost float64) error
	Report(year, month string) (string, error)
	History(userID uuid.UUID, offset, limit int) ([]model.History, error)
}

type controller struct {
	repository IRepository
}

func NewController(repository IRepository) (IController, error) {
	if repository == nil {
		return nil, Err.ErrNoRepository
	}
	return &controller{repository: repository}, nil
}

//go:generate minimock -g -i
type IRepository interface {
	Balance(userID uuid.UUID) (*model.User, error)
	Enrollment(user model.User, funds float64) error
	Transfer(sender, recipient model.User, funds float64) error
	AddUser(user model.User) error
	Order(user model.User, order model.Order) error
	GetOrder(orderID uuid.UUID) (*model.Order, error)
	OrderSuccess(order model.Order) error
	OrderFailed(user model.User, order model.Order) error
	Report(time.Time) ([]model.Report, error)
	History(userID uuid.UUID, limit, offset int) ([]model.History, error)
}

func (c *controller) Balance(userID uuid.UUID) (*model.User, error) {
	logrus.Infoln("Starting controller.Balance")

	user, err := c.repository.Balance(userID)
	if err != nil {
		logrus.Infoln("Ending controller.Balance")
		return nil, err
	}

	logrus.Infoln("Ending controller.Balance")
	return user, err
}

func (c *controller) Enrollment(userID uuid.UUID, funds float64) error {
	logrus.Infoln("Starting controller.Accrual")

	user := model.User{ID: userID, Funds: funds}

	balance, err := c.repository.Balance(userID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logrus.Infoln("Ending controller.Enrollment")
			return err
		}

		user.DateCreate = time.Now()
		user.LastUpdate = time.Now()

		err = c.repository.AddUser(user)

		logrus.Infoln("Ending controller.Enrollment")
		return err
	}

	user.Funds += balance.Funds
	user.LastUpdate = time.Now()

	err = c.repository.Enrollment(user, funds)

	logrus.Infoln("Ending controller.Enrollment")
	return err
}

func (c *controller) Transfer(senderID, recipientID uuid.UUID, funds float64) error {
	logrus.Infoln("Starting controller.Transfer")

	sender, err := c.repository.Balance(senderID)
	if err != nil {
		logrus.Infoln("Ending controller.Transfer")
		return err
	}

	if sender.Funds < funds {
		logrus.Errorf("%s sender.Fuds: %s, funds: %s\n", Err.ErrInsufficientFunds, sender.Funds, funds)
		logrus.Infoln("Ending controller.Transfer")
		return Err.ErrInsufficientFunds
	}

	recipient, err := c.repository.Balance(recipientID)
	if err != nil {
		logrus.Infoln("Ending controller.Transfer")
		return err
	}

	sender.Funds -= funds
	sender.LastUpdate = time.Now()

	recipient.Funds += funds
	recipient.LastUpdate = time.Now()

	err = c.repository.Transfer(*sender, *recipient, funds)

	logrus.Infoln("Ending controller.Transfer")
	return err
}

func (c *controller) Order(userID, serviceID, orderID uuid.UUID, serviceName string, funds float64) error {
	logrus.Infoln("Starting controller.Order")

	user, err := c.repository.Balance(userID)
	if err != nil {
		logrus.Infoln("Ending controller.Order")
		return err
	}

	if funds > user.Funds {
		logrus.Errorf("%s user.Funds: %s, cost: %s\n", Err.ErrInsufficientFunds, user.Funds, funds)
		logrus.Infoln("Ending controller.Order")
		return Err.ErrInsufficientFunds
	}

	user.Funds -= funds
	user.LastUpdate = time.Now()

	order := model.Order{ID: orderID, UserID: userID, ServiceID: serviceID, ServiceName: serviceName, DateCreate: time.Now(), Funds: funds}

	err = c.repository.Order(*user, order)

	logrus.Infoln("Ending controller.Order")
	return err
}

func (c *controller) OrderSuccess(userID, serviceID, orderID uuid.UUID, serviceName string, cost float64) error {
	logrus.Infoln("Starting controller.OrderSuccess")

	order, err := c.repository.GetOrder(orderID)
	if err != nil {
		logrus.Infoln("Ending controller.OrderSuccess")
		return err
	}

	if userID != order.UserID || serviceID != order.ServiceID || orderID != order.ID || serviceName != order.ServiceName || cost != order.Funds {
		logrus.Errorln(Err.ErrBadRequest)
		return Err.ErrBadRequest
	}

	err = c.repository.OrderSuccess(model.Order{ID: orderID, UserID: userID, ServiceID: serviceID, ServiceName: serviceName, DateCreate: order.DateCreate, Funds: order.Funds})

	logrus.Infoln("Ending controller.OrderSuccess")
	return err
}

func (c *controller) OrderFailed(userID, serviceID, orderID uuid.UUID, serviceName string, cost float64) error {
	logrus.Infoln("Starting controller.OrderFailed")

	order, err := c.repository.GetOrder(orderID)
	if err != nil {
		logrus.Infoln("Ending controller.OrderFailed")
		return err
	}

	if userID != order.UserID || serviceID != order.ServiceID || orderID != order.ID || serviceName != order.ServiceName || cost != order.Funds {
		logrus.Errorln(Err.ErrBadRequest)
		return Err.ErrBadRequest
	}

	user, err := c.repository.Balance(userID)
	if err != nil {
		logrus.Infoln("Ending controller.OrderFailed")
		return err
	}

	user.Funds += cost
	user.LastUpdate = time.Now()

	err = c.repository.OrderFailed(*user, *order)

	logrus.Infoln("Ending controller.OrderFailed")
	return err
}

func (c *controller) Report(year, month string) (string, error) {
	logrus.Infoln("Starting controller.Report")

	date := fmt.Sprintf("%s-%s-01", year, month)
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		logrus.Errorf("Parse %s: %s\n", date, err)
		logrus.Infoln("Ending controller.Report")
		return "", Err.ErrBadRequest
	}

	rep, err := c.repository.Report(t)
	if err != nil {
		logrus.Infoln("Ending controller.Report")
		return "", err
	}

	var report [][]string
	for _, r := range rep {
		var slice []string
		slice = append(slice, r.ServiceName, strconv.FormatFloat(r.Revenue, 'f', -1, 64))
		report = append(report, slice)
	}

	id := uuid.New().String()
	csvFile, err := os.Create("./reports/" + id + ".csv")
	if err != nil {
		logrus.Errorln("Create: ", err)
		logrus.Infoln("Ending controller.Report")
		return "", err
	}

	csvWriter := csv.NewWriter(csvFile)
	for _, row := range report {
		_ = csvWriter.Write(row)
	}
	csvWriter.Flush()

	logrus.Infoln("Ending controller.Report")
	return id, nil
}

func (c *controller) History(userID uuid.UUID, limit, offset int) ([]model.History, error) {
	logrus.Infoln("Starting controller.History")

	if _, err := c.repository.Balance(userID); err != nil {
		logrus.Infoln("Ending controller.History")
		return nil, err
	}

	report, err := c.repository.History(userID, limit, offset)
	if err != nil {
		logrus.Infoln("Ending controller.History")
		return nil, err
	}

	logrus.Infoln("Ending controller.History")
	return report, nil
}
