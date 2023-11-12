package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"

	"tonexplorer/internal/repo/shards"
	"tonexplorer/internal/repo/transactions"
)

type API struct {
	shardsRepo      *shards.R
	transactionRepo *transactions.R
}

func New(shardsRepo *shards.R, transactionRepo *transactions.R) *API {
	return &API{shardsRepo, transactionRepo}
}

func (a *API) Init() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/debug", a.debug)
	e.GET("/account/:id", a.account)
	e.GET("/tx/all", a.transactions)
	e.GET("/tx/:id", a.transactionByID)

	return e
}

func (a *API) debug(c echo.Context) error {
	return c.JSON(http.StatusOK, "debug")
}

func (a *API) account(c echo.Context) error {
	return c.JSON(http.StatusOK, "Account")
}

func (a *API) transactions(c echo.Context) error {
	return c.JSON(http.StatusOK, "Transaction")
}

func (a *API) transactionByID(c echo.Context) error {
	return c.JSON(http.StatusOK, "TransactionByID")
}
