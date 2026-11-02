# Tasks

## 1. Implementation

- [x] 1.1 In backend/internal/handler/user_handler_test.go, add a test that parses the 200 response of GET by id. It checks success=true, data.id, data.email and data.firstName, and that there is no error field.
- [x] 1.2 In the same file, extend the 400 and 404 coverage to parse the body. It checks success is false, error.code is VALIDATION_ERROR or NOT_FOUND, and error.message is "Invalid user ID." or "User not found.". Include a whitespace id (%20) and a wrong-length hex id.
- [x] 1.3 In the same file, add a 500 test. Set repo.Err = errors.New("boom db failure"), which avoids the words "connection" and "server selection". GET a valid ObjectID hex and expect 500, code INTERNAL_ERROR, and a body that does not contain the raw error text.
- [x] 1.4 In the same file, add a test for the missing-id request GET /api/v1/users/. It asserts the status the router actually returns, then resolves open question 2.
- [x] 1.5 In backend/internal/service/user_service_test.go, add TestGet_UnexpectedError. With repo.Err set to a generic error, Get returns an apperror.Error of kind KindInternal. Add TestGet_EmptyID, which expects KindValidation.
- [x] 1.6 In backend/internal/server/integration_test.go, extend TestMongoUnavailableReturns503 to GET /api/v1/users/5f8d0d55b54764421b7156c3 and expect 503 with no leaked host. This runs without MongoDB. In the real-MongoDB lifecycle test, add an assertion on the 200 body shape and a 400 for a bad id.
- [x] 1.7 Only if a new test fails, make the smallest fix in user_handler.go, user_service.go or apperror.go. Otherwise leave production code unchanged.
- [x] 1.8 If the behavior from task 4 deserves documenting, update the README.md:52 row to match. Otherwise leave README.md alone.
