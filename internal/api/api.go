package api

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	Err "Avito/internal/errors"
	"Avito/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type IApi interface {
	Balance(c *gin.Context)
	Enrollment(c *gin.Context)
	Transfer(c *gin.Context)
	Order(c *gin.Context)
	OrderSuccess(c *gin.Context)
	OrderFailed(c *gin.Context)
	Report(c *gin.Context)
	CsvReport(c *gin.Context)
	History(c *gin.Context)
}

type api struct {
	controller IController
}

func NewApi(controller IController) (IApi, error) {
	if controller == nil {
		return nil, Err.ErrNoController
	}
	return &api{controller: controller}, nil
}

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

// @Summary      Balance
// @Description  Предоставляет информацию о пользователе
// @Tags         balance
// @Produce      json
// @Param        id   query   string  true "UserID"
// @Success		 200 {object} model.User
// @Failure 	 400 {object} message
// @Failure 	 404 {object} message
// @Failure 	 500 {object} message
// @Router       /balance [get]
func (a *api) Balance(c *gin.Context) {
	logrus.Infoln("Starting api.Balance")

	arg := c.Query("id")
	userID, err := uuid.Parse(arg)
	if err != nil {
		logrus.Errorf("Parse %s: %s\n", arg, err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.Balance")
		return
	}

	user, err := a.controller.Balance(userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.IndentedJSON(http.StatusNotFound, message{Message: "Not found"})
			logrus.Infoln("Ending api.Balance")
			return
		} else {
			c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal error"})
			logrus.Infoln("Ending api.Balance")
			return
		}
	}

	c.IndentedJSON(http.StatusOK, user)
	logrus.Infoln("Ending api.Balance")
}

// @Summary      Enrollment
// @Description  Начисляет пользователю средства, регистрирует его
// @Tags         balance
// @Accept       json
// @Produce      json
// @Success		 200 {object} message
// @Failure 	 400 {object} message
// @Failure 	 500 {object} message
// @Router       /balance [post]
func (a *api) Enrollment(c *gin.Context) {
	logrus.Infoln("Starting api.Enrollment")

	u := user{}
	if err := json.NewDecoder(c.Request.Body).Decode(&u); err != nil {
		logrus.Errorln("Decoding: ", err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.Enrollment")
		return
	}

	if u.Funds <= 0 {
		logrus.Errorf("%s: %s", u.Funds, Err.ErrBadRequest)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.Enrollment")
		return
	}

	err := a.controller.Enrollment(u.ID, u.Funds)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal error"})
		logrus.Infoln("Ending api.Enrollment")
		return
	}

	c.IndentedJSON(http.StatusOK, message{Message: "Success"})
	logrus.Infoln("Ending api.Enrollment")
}

// @Summary      Transfer
// @Description  Перевод средств от пользователя к пользователю
// @Tags         balance
// @Accept       json
// @Produce      json
// @Success		 200 {object} message
// @Failure 	 400 {object} message
// @Failure 	 404 {object} message
// @Failure 	 500 {object} message
// @Router       /transfer [post]
func (a *api) Transfer(c *gin.Context) {
	logrus.Infoln("Starting api.Transfer")

	t := transfer{}
	if err := json.NewDecoder(c.Request.Body).Decode(&t); err != nil {
		logrus.Errorln("Decoding: ", err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.Transfer")
		return
	}

	if t.Funds <= 0 {
		logrus.Errorf("%s: %s", t.Funds, Err.ErrBadRequest)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.Transfer")
		return
	}

	err := a.controller.Transfer(t.SenderID, t.RecipientID, t.Funds)
	if err != nil {
		switch {
		case errors.Is(err, Err.ErrInsufficientFunds):
			c.IndentedJSON(http.StatusBadRequest, message{Message: "Insufficient funds"})
			logrus.Infoln("Ending api.Transfer")
			return
		case errors.Is(err, pgx.ErrNoRows):
			c.IndentedJSON(http.StatusNotFound, message{Message: "Not found"})
			logrus.Infoln("Ending api.Transfer")
			return
		default:
			c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal error"})
			logrus.Infoln("Ending api.Transfer")
			return
		}
	}

	c.IndentedJSON(http.StatusOK, message{Message: "Success"})
	logrus.Infoln("Ending api.Transfer")
}

// @Summary      Order
// @Description  Заказ пользователем услуги
// @Tags         order
// @Accept       json
// @Produce      json
// @Success		 200 {object} message
// @Failure 	 400 {object} message
// @Failure 	 404 {object} message
// @Failure 	 500 {object} message
// @Router       /order [post]
func (a *api) Order(c *gin.Context) {
	logrus.Infoln("Starting api.Order")

	o := order{}
	if err := json.NewDecoder(c.Request.Body).Decode(&o); err != nil {
		logrus.Errorln("Decoding: ", err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.Order")
		return
	}

	if o.Cost <= 0 {
		logrus.Errorf("%s: %s", o.Cost, Err.ErrBadRequest)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.Order")
		return
	}

	if o.ServiceName == "" {
		logrus.Errorln(Err.ErrBadRequest)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.Order")
		return
	}

	err := a.controller.Order(o.UserID, o.ServiceID, o.OrderID, o.ServiceName, o.Cost)
	if err != nil {
		switch {
		case errors.Is(err, Err.ErrInsufficientFunds):
			c.IndentedJSON(http.StatusBadRequest, message{Message: "Insufficient funds"})
			logrus.Infoln("Ending api.Order")
			return
		case errors.Is(err, pgx.ErrNoRows):
			c.IndentedJSON(http.StatusNotFound, message{Message: "Not found"})
			logrus.Infoln("Ending api.Order")
			return
		default:
			c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal error"})
			logrus.Infoln("Ending api.Order")
			return
		}
	}

	c.IndentedJSON(http.StatusOK, message{Message: "Success"})
	logrus.Infoln("Ending api.Order")
}

// @Summary      Success order
// @Description  Успешное выполнение услуги
// @Tags         order
// @Accept       json
// @Produce      json
// @Success		 200 {object} message
// @Failure 	 400 {object} message
// @Failure 	 404 {object} message
// @Failure 	 500 {object} message
// @Router       /order/success [post]
func (a *api) OrderSuccess(c *gin.Context) {
	logrus.Infoln("Starting api.OrderSuccess")

	o := order{}
	if err := json.NewDecoder(c.Request.Body).Decode(&o); err != nil {
		logrus.Errorln("Decoding: ", err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.OrderSuccess")
		return
	}

	err := a.controller.OrderSuccess(o.UserID, o.ServiceID, o.OrderID, o.ServiceName, o.Cost)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.IndentedJSON(http.StatusBadRequest, message{Message: "Not found"})
			logrus.Infoln("Ending api.OrderSuccess")
			return
		case errors.Is(err, Err.ErrBadRequest):
			c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
			logrus.Infoln("Ending api.OrderSuccess")
			return
		default:
			c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal error"})
			logrus.Infoln("Ending api.OrderSuccess")
			return
		}
	}

	c.IndentedJSON(http.StatusOK, message{Message: "Success"})
	logrus.Infoln("Ending api.OrderSuccess")
}

// @Summary      Failed order
// @Description  Услуга не была оказана
// @Tags         order
// @Accept       json
// @Produce      json
// @Success		 200 {object} message
// @Failure 	 400 {object} message
// @Failure 	 404 {object} message
// @Failure 	 500 {object} message
// @Router       /order/failed [post]
func (a *api) OrderFailed(c *gin.Context) {
	logrus.Infoln("Starting api.OrderFailed")

	o := order{}
	if err := json.NewDecoder(c.Request.Body).Decode(&o); err != nil {
		logrus.Errorln("Decoding: ", err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.OrderFailed")
		return
	}

	err := a.controller.OrderFailed(o.UserID, o.ServiceID, o.OrderID, o.ServiceName, o.Cost)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.IndentedJSON(http.StatusBadRequest, message{Message: "Not found"})
			logrus.Infoln("Ending api.OrderFailed")
			return
		case errors.Is(err, Err.ErrBadRequest):
			c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
			logrus.Infoln("Ending api.OrderFailed")
			return
		default:
			c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal error"})
			logrus.Infoln("Ending api.OrderFailed")
			return
		}
	}

	c.IndentedJSON(http.StatusOK, message{Message: "Success"})
	logrus.Infoln("Ending api.OrderFailed")
}

// @Summary      Report
// @Description  Предоставляет ссылку на месячный отчет по пользователям
// @Tags         report
// @Accept       json
// @Produce      json
// @Success		 200 {object} message
// @Failure 	 400 {object} message
// @Failure 	 500 {object} message
// @Router       /report [post]
func (a *api) Report(c *gin.Context) {
	logrus.Infoln("Starting api.Report")

	r := report{}
	if err := json.NewDecoder(c.Request.Body).Decode(&r); err != nil {
		logrus.Errorln("Deconding: ", err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.Report")
		return
	}

	str, err := a.controller.Report(r.Year, r.Month)
	if err != nil {
		if errors.Is(err, Err.ErrBadRequest) {
			c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
			logrus.Infoln("Ending api.Report")
			return
		} else {
			c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal error"})
			logrus.Infoln("Ending api.Report")
			return
		}
	}

	c.IndentedJSON(http.StatusOK, message{Message: "http://localhost:8080/report/csv?id=" + str})
	logrus.Infoln("Ending api.Report")
}

// @Summary      CsvReport
// @Description  Предоставляет месячный отчет по пользователям
// @Tags         report
// @Produce      plain
// @Param        id   query   string  true "file ID"
// @Success      200 {array}  string
// @Failure 	 400 {object} message
// @Failure 	 500 {object} message
// @Router       /report/csv [get]
func (a *api) CsvReport(c *gin.Context) {
	logrus.Infoln("Starting api.CsvReport")

	id := c.Query("id")
	name := fmt.Sprintf("./reports/%s.csv", id)

	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		logrus.Errorf("Stat %s: %s\n", name, err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.CsvReport")
		return
	}

	file, err := os.Open(name)
	if err != nil {
		logrus.Errorf("Open %s: %s\n", name, err)
		c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal Error"})
		logrus.Infoln("Ending api.CsvReport")
		return
	}

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		logrus.Errorln("ReadALL: ", err)
		c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal Error"})
		logrus.Infoln("Ending api.CsvReport")
		return
	}

	for _, str := range records {
		for i, s := range str {
			if i == 0 {
			} else {
				c.Writer.Write([]byte(";"))
			}
			c.Writer.Write([]byte(s))
		}
		c.Writer.Write([]byte("\n"))
	}

	logrus.Infoln("Ending api.CsvReport")
}

// @Summary      History
// @Description  Предоставляет историю заказов пользователя
// @Tags         report
// @Produce      json
// @Success		 200 {array}  model.History
// @Failure 	 400 {object} message
// @Failure 	 404 {object} message
// @Failure 	 500 {object} message
// @Router       /history [post]
func (a *api) History(c *gin.Context) {
	logrus.Infoln("Starting api.History")

	id := c.Query("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		logrus.Errorf("Parse %s: %s\n", id, err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.History")
		return
	}

	l := c.Query("limit")
	limit, err := strconv.Atoi(l)
	if err != nil {
		logrus.Errorf("Atoi %s: %s\n", l, err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.History")
		return
	}

	o := c.Query("offset")
	offset, err := strconv.Atoi(o)
	if err != nil {
		logrus.Errorf("Atoi %s: %s\n", o, err)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.History")
		return
	}

	if limit <= 0 || offset < 0 {
		logrus.Errorf("%s, limit: %s, offset: %s\n", Err.ErrBadRequest, limit, offset)
		c.IndentedJSON(http.StatusBadRequest, message{Message: "Wrong data"})
		logrus.Infoln("Ending api.History")
		return
	}

	report, err := a.controller.History(userID, limit, offset)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.IndentedJSON(http.StatusNotFound, message{Message: "Not found"})
			logrus.Infoln("Ending api.History")
			return
		} else {
			c.IndentedJSON(http.StatusInternalServerError, message{Message: "Internal error"})
			logrus.Infoln("Ending api.History")
			return
		}
	}

	c.IndentedJSON(http.StatusOK, report)
	logrus.Infoln("Ending api.History")
}
