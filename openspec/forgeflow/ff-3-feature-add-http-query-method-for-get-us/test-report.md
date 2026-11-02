# Test report

Tested commit `42b2d7c93eb070043604b277f58ed69b9c8c1891` on branch `forgeflow/3-feature-add-http-query-method-for-get-us`.

Commands from: approved plan v3 (Commands)

53 test cases: 49 passed, 0 failed, 4 skipped.

| Command | Exit | Seconds |
|---|---|---|
| `cd backend && go build ./... && go vet ./...` | 0 | 0.4 |
| `cd frontend && npm run build` | 0 | 1.6 |
| `cd backend && go test -v -count=1 ./...` | 0 | 0.8 |
| `cd frontend && npx vitest run --reporter=verbose` | 0 | 5.7 |
