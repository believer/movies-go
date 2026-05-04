package handlers

import (
	"believer/movies/db"
	"believer/movies/mocks"
	"believer/movies/types"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type productionCountyHandlerTest struct {
	name         string
	url          string
	mockSetup    func(*mocks.MockProductionCountryQuerier)
	mockAssert   func(*mocks.MockProductionCountryQuerier)
	expectedCode int
}

func setupProductionCountryApp(h *ProductionCountryHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/production-country", h.ListProductionCountries)
	app.Get("/production-country/stats", h.GetProductionCountryStats)
	app.Get("/production-country/:id", h.GetProductionCountry)

	return app
}

func runProductionCountryHandlerTests(t *testing.T, tests []productionCountyHandlerTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockProductionCountryQuerier(t)
			tt.mockSetup(repo)

			app := setupProductionCountryApp(NewProductionCountryHandler(repo))
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.mockAssert != nil {
				tt.mockAssert(repo)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestListProductionCountries(t *testing.T) {
	countries := []db.ProductionCountry{{Name: "Test"}}

	runProductionCountryHandlerTests(t, []productionCountyHandlerTest{
		{
			name: "returns companies by default",
			url:  "/production-country",
			mockSetup: func(m *mocks.MockProductionCountryQuerier) {
				m.On("ListProductionCountries").Return(countries, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns error if request fails",
			url:  "/production-country",
			mockSetup: func(m *mocks.MockProductionCountryQuerier) {
				m.On("ListProductionCountries").Return(countries, fmt.Errorf("Fail"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	})
}

func TestGetProductionCountry(t *testing.T) {
	company := db.TableName{Name: "Test"}
	movies := types.Movies{{Title: "Test"}}

	runProductionCountryHandlerTests(t, []productionCountyHandlerTest{
		{
			name: "returns OK",
			url:  "/production-country/1",
			mockSetup: func(m *mocks.MockProductionCountryQuerier) {
				m.On("GetProductionCountryName", "1").Return(company, nil)
				m.On("GetProductionCountryMovies", "1", "user-123", 0).Return(movies, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns error if company breaks",
			url:  "/production-country/1",
			mockSetup: func(m *mocks.MockProductionCountryQuerier) {
				m.On("GetProductionCountryName", "1").Return(company, fmt.Errorf("Fail"))
			},
			mockAssert: func(m *mocks.MockProductionCountryQuerier) {
				m.AssertNotCalled(t, "GetProductionCountryMovies", "1", "user-123", 0)
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "returns error if movies breaks",
			url:  "/production-country/1",
			mockSetup: func(m *mocks.MockProductionCountryQuerier) {
				m.On("GetProductionCountryName", "1").Return(company, nil)
				m.On("GetProductionCountryMovies", "1", "user-123", 0).Return(movies, fmt.Errorf("Fail"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	})
}

func TestGetProductionCountryStats(t *testing.T) {
	items := []types.ListItem{{Name: "Item"}}

	runProductionCountryHandlerTests(t, []productionCountyHandlerTest{
		{
			name: "returns OK",
			url:  "/production-country/stats",
			mockSetup: func(m *mocks.MockProductionCountryQuerier) {
				m.On("GetProductionCountryStats", "user-123", "All").Return(items, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns OK for specific year",
			url:  "/production-country/stats?year=2025",
			mockSetup: func(m *mocks.MockProductionCountryQuerier) {
				m.On("GetProductionCountryStats", "user-123", "2025").Return(items, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns error",
			url:  "/production-country/stats",
			mockSetup: func(m *mocks.MockProductionCountryQuerier) {
				m.On("GetProductionCountryStats", "user-123", "All").Return(items, fmt.Errorf("Fail"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	})
}
