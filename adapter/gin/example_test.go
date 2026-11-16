package ginadapter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"github.com/DarkoKlisuric/godebugbar"
	ginadapter "github.com/DarkoKlisuric/godebugbar/adapter/gin"
)

func ExampleAttach() {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	godebugbar.New(
		godebugbar.WithEnabled(true),
		ginadapter.Attach(router),
	)

	router.GET("/users/:id", func(c *gin.Context) {
		c.String(http.StatusOK, "user %s", c.Param("id"))
	})

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/users/42", nil))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/_debugbar/api/profiles?limit=1", nil))

	var profiles []struct {
		Path  string `json:"path"`
		Route string `json:"route"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&profiles)
	fmt.Println(profiles[0].Path, profiles[0].Route)
	// Output: /users/42 /users/:id
}
