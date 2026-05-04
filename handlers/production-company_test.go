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

type productionCompanyHandlerTest struct {
	name         string
	url          string
	mockSetup    func(*mocks.MockProductionCompanyQuerier)
	mockAssert   func(*mocks.MockProductionCompanyQuerier)
	expectedCode int
}

func setupProductionCompanyApp(h *ProductionCompanyHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/production-company", h.ListProductionCompanies)
	app.Get("/production-company/stats", h.GetProductionCompanyStats)
	app.Get("/production-company/:id", h.GetProductionCompany)

	return app
}

func runProductionCompanyHandlerTests(t *testing.T, tests []productionCompanyHandlerTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockProductionCompanyQuerier(t)
			tt.mockSetup(repo)

			app := setupProductionCompanyApp(NewProductionCompanyHandler(repo))
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

func TestListProductionCompanies(t *testing.T) {
	companies := []db.ProductionCompany{{Name: "Test"}}

	runProductionCompanyHandlerTests(t, []productionCompanyHandlerTest{
		{
			name: "returns companies by default",
			url:  "/production-company",
			mockSetup: func(m *mocks.MockProductionCompanyQuerier) {
				m.On("ListProductionCompanies", 1).Return(companies, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns error if request fails",
			url:  "/production-company",
			mockSetup: func(m *mocks.MockProductionCompanyQuerier) {
				m.On("ListProductionCompanies", 1).Return(companies, fmt.Errorf("Fail"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	})
}

func TestGetProductionCompany(t *testing.T) {
	company := db.TableName{Name: "Test"}
	movies := types.Movies{{Title: "Test"}}

	runProductionCompanyHandlerTests(t, []productionCompanyHandlerTest{
		{
			name: "returns OK",
			url:  "/production-company/1",
			mockSetup: func(m *mocks.MockProductionCompanyQuerier) {
				m.On("GetProductionCompanyName", "1").Return(company, nil)
				m.On("GetProductionCompanyMovies", "1", "user-123", 0).Return(movies, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns error if company breaks",
			url:  "/production-company/1",
			mockSetup: func(m *mocks.MockProductionCompanyQuerier) {
				m.On("GetProductionCompanyName", "1").Return(company, fmt.Errorf("Fail"))
			},
			mockAssert: func(m *mocks.MockProductionCompanyQuerier) {
				m.AssertNotCalled(t, "GetProductionCompanyMovies", "1", "user-123", 0)
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "returns error if movies breaks",
			url:  "/production-company/1",
			mockSetup: func(m *mocks.MockProductionCompanyQuerier) {
				m.On("GetProductionCompanyName", "1").Return(company, nil)
				m.On("GetProductionCompanyMovies", "1", "user-123", 0).Return(movies, fmt.Errorf("Fail"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	})
}

func TestGetProductionCompanyStats(t *testing.T) {
	items := []types.ListItem{{Name: "Item"}}

	runProductionCompanyHandlerTests(t, []productionCompanyHandlerTest{
		{
			name: "returns OK",
			url:  "/production-company/stats",
			mockSetup: func(m *mocks.MockProductionCompanyQuerier) {
				m.On("GetProductionCompanyStats", "user-123", "All").Return(items, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns OK for specific year",
			url:  "/production-company/stats?year=2025",
			mockSetup: func(m *mocks.MockProductionCompanyQuerier) {
				m.On("GetProductionCompanyStats", "user-123", "2025").Return(items, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns error",
			url:  "/production-company/stats",
			mockSetup: func(m *mocks.MockProductionCompanyQuerier) {
				m.On("GetProductionCompanyStats", "user-123", "All").Return(items, fmt.Errorf("Fail"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	})
}
