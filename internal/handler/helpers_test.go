package handler_test

import (
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
)

const testUserID = "user-1"

// authed marks the request as made by the test user, as RequireSession
// would in production.
func authed(req *http.Request) *http.Request {
	return req.WithContext(handler.ContextWithUser(req.Context(), model.User{ID: testUserID}))
}
