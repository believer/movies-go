package handlers

import (
	"believer/movies/db"
	"believer/movies/mocks"
	"believer/movies/types"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type personHandlerTest struct {
	name         string
	url          string
	mockSetup    func(*mocks.MockPersonQuerier)
	mockAssert   func(*mocks.MockPersonQuerier)
	expectedCode int
}

func setupPersonApp(h *PersonHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/person/:id", h.GetPersonByID)

	return app
}

func runPersonHandlerTests(t *testing.T, tests []personHandlerTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockPersonQuerier(t)
			tt.mockSetup(repo)

			app := setupPersonApp(NewPersonHandler(repo))
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetPersonByID(t *testing.T) {
	person := types.Person{Name: "Brad Pitt"}
	academy := db.GroupedAwards{"Test": {{ID: "1234"}}}
	baftas := db.GroupedAwards{"Test": {{ID: "4321"}}}
	order := []string{"Test"}

	runPersonHandlerTests(t, []personHandlerTest{
		{
			name: "returns person and awards",
			url:  "/person/brad-pitt-1",
			mockSetup: func(m *mocks.MockPersonQuerier) {
				m.On("GetPersonByID", "1", "user-123").Return(person, nil)
				m.On("GetGroupedAwards", "1", db.AcademyAward).Return(academy, 4, order, nil)
				m.On("GetGroupedAwards", "1", db.Bafta).Return(baftas, 0, order, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "handles missing person",
			url:  "/person/brad-pitt-1",
			mockSetup: func(m *mocks.MockPersonQuerier) {
				m.On("GetPersonByID", "1", "user-123").Return(person, sql.ErrNoRows)
			},
			mockAssert: func(m *mocks.MockPersonQuerier) {
				m.AssertNotCalled(t, "GetGroupedAwards", "1", db.AcademyAward)
				m.AssertNotCalled(t, "GetGroupedAwards", "1", db.Bafta)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "handles academy award error",
			url:  "/person/brad-pitt-1",
			mockSetup: func(m *mocks.MockPersonQuerier) {
				m.On("GetPersonByID", "1", "user-123").Return(person, nil)
				m.On("GetGroupedAwards", "1", db.AcademyAward).Return(academy, 4, order, fmt.Errorf("Test"))
			},
			mockAssert: func(m *mocks.MockPersonQuerier) {
				m.AssertNotCalled(t, "GetGroupedAwards", "1", db.Bafta)
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "handles bafta error",
			url:  "/person/brad-pitt-1",
			mockSetup: func(m *mocks.MockPersonQuerier) {
				m.On("GetPersonByID", "1", "user-123").Return(person, nil)
				m.On("GetGroupedAwards", "1", db.AcademyAward).Return(academy, 4, order, nil)
				m.On("GetGroupedAwards", "1", db.Bafta).Return(baftas, 0, order, fmt.Errorf("Test"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	})
}
