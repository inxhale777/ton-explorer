package main

import (
	"context"
	"github.com/blevesearch/bleve/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"tonexplorer/internal/entity"
	"tonexplorer/internal/fetcher"
	"tonexplorer/internal/tclient"
)

func main() {
	ctx := context.Background()
	api, err := tclient.New()
	if err != nil {
		log.Fatalf("API() fail: %s", err)
	}

	var insertedTX int
	txStorage := make(map[string]entity.Transaction)
	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		log.Fatal(err)
		return
	}

	go func() {
		f := fetcher.New(api)
		txCh, errCh, err := f.Run(ctx)
		if err != nil {
			log.Fatal(err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-txCh:
				txStorage[t.Hash] = t
				insertedTX++
				err := index.Index(t.Hash, t)
				if err != nil {
					log.Fatal(err)
					return
				}
			case e := <-errCh:
				log.Fatal(e)
				return
			}
		}
	}()

	go func() {
		e := echo.New()
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())

		e.GET("/debug", func(c echo.Context) error {
			stat := index.StatsMap()
			stat["total_inserted"] = insertedTX

			return c.JSON(http.StatusOK, stat)
		})
		e.GET("/account/:id", func(c echo.Context) error {
			query := bleve.NewMatchQuery(c.Param("id"))
			query.SetField("Account")
			req := bleve.NewSearchRequestOptions(query, 1000, 0, false)

			r, err := index.Search(req)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, err)
			}

			var result []entity.Transaction
			for _, i := range r.Hits {
				result = append(result, txStorage[i.ID])
			}

			return c.JSON(http.StatusOK, result)
		})
		e.GET("/tx/all", func(c echo.Context) error {
			req := bleve.NewSearchRequestOptions(bleve.NewMatchAllQuery(), 1000, 0, false)

			r, err := index.Search(req)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, err)
			}

			var result []entity.Transaction
			for _, i := range r.Hits {
				result = append(result, txStorage[i.ID])
			}

			return c.JSON(http.StatusOK, result)
		})
		e.GET("/tx/:id", func(c echo.Context) error {
			q := bleve.NewSearchRequest(bleve.NewDocIDQuery([]string{c.Param("id")}))
			r, err := index.Search(q)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, err)
			}

			if len(r.Hits) > 0 {
				return c.JSON(http.StatusOK, txStorage[r.Hits[0].ID])
			}

			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "not found",
			})
		})

		e.Logger.Fatal(e.Start(":8083"))
	}()

	select {}
}
