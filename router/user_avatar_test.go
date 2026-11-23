package router

import (
	"bytes"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/MoeForge/moego/base64"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelfAvatarUploadAndAccess(t *testing.T) {
	setupRelayRouterTestDB(t)
	require.NoError(t, model.DB.AutoMigrate(&model.UserSession{}))
	engine := gin.New()
	SetApiRouter(engine)

	login := func(name string) (*model.User, string) {
		user := &model.User{Username: name, AffCode: name, Role: common.RoleCommonUser}
		require.NoError(t, model.DB.Create(user).Error)
		session, err := service.CreateLoginSession(user.Id, "password", "127.0.0.1", "avatar-test")
		require.NoError(t, err)
		return user, session.AccessToken
	}
	owner, token := login("avatar-owner")
	_, otherToken := login("avatar-other")
	request := func(method, path, bearer string, data []byte) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewReader(data))
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		// A misleading content type must neither admit invalid bytes nor reject a real image.
		req.Header.Set("Content-Type", "image/jpeg")
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, req)
		return recorder
	}
	version := func(response *httptest.ResponseRecorder) string {
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		var envelope struct {
			Success bool
			Data    struct {
				Version string `json:"avatar_version"`
			}
		}
		require.NoError(t, common.Unmarshal(response.Body.Bytes(), &envelope))
		require.True(t, envelope.Success, response.Body.String())
		return envelope.Data.Version
	}
	const writePath = "/api/user/self/avatar"
	readPath := "/api/user/" + strconv.Itoa(owner.Id) + "/avatar"
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		assert.Equal(t, http.StatusUnauthorized, request(method, writePath, "", nil).Code)
	}
	assert.Equal(t, http.StatusNotFound, request(http.MethodGet, readPath, "", nil).Code)

	// Two differently colored frames in each animated fixture.
	photos := []struct{ mime, data string }{
		{"image/gif", "R0lGODlhAQABAIEAAP8AAAAAAAAAAAAAACH/C05FVFNDQVBFMi4wAwEAAAAh+QQACgAAACwAAAAAAQABAAAIBAABBAQAIfkEAQoAAQAsAAAAAAEAAQCBAAD/AAAAAAAAAAAACAQAAQQEADs="},
		{"image/png", "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAACGFjVEwAAAACAAAAAPONk3AAAAAaZmNUTAAAAAAAAAABAAAAAQAAAAAAAAAAAAEACgAAWn8w0AAAAAxJREFUeJxj+M/AAAADAQEAyf6S7wAAABpmY1RMAAAAAQAAAAEAAAABAAAAAAAAAAAAAQAKAADBDNoEAAAAEGZkQVQAAAACeJxjYGD4DwABAwEAaL/t9AAAAABJRU5ErkJggg=="},
		{"image/webp", "UklGRoQAAABXRUJQVlA4WAoAAAACAAAAAAAAAAAAQU5JTQYAAAAAAAAAAABBTk1GKAAAAAAAAAAAAAAAAAAAAGQAAAJWUDhMDwAAAC8AAAAABxD9j/4HIqL/AQBBTk1GKAAAAAAAAAAAAAAAAAAAAGQAAABWUDhMDwAAAC8AAAAABxDR//4HIqL/AQA="},
	}
	var lastURL string
	for _, photo := range photos {
		data, err := base64.StdEncoding.DecodeString(photo.data)
		require.NoError(t, err)
		newVersion := version(request(http.MethodPut, writePath, token, data))
		require.NotEmpty(t, newVersion)
		newURL := readPath + "?v=" + newVersion
		if lastURL != "" {
			assert.Equal(t, http.StatusNotFound, request(http.MethodGet, lastURL, "", nil).Code)
		}
		img := request(http.MethodGet, newURL, "", nil)
		require.Equal(t, http.StatusOK, img.Code)
		assert.Equal(t, photo.mime, img.Header().Get("Content-Type"))
		assert.Contains(t, img.Header().Get("Cache-Control"), "private")
		assert.Equal(t, data, img.Body.Bytes())
		self := request(http.MethodGet, "/api/user/self", token, nil)
		assert.Equal(t, newVersion, version(self))
		assert.NotContains(t, self.Body.String(), photo.data)
		assert.NotContains(t, self.Body.String(), `"avatar":`)
		lastURL = newURL
	}

	// A stale User save and writes by a second authenticated user must leave the owner's avatar intact.
	require.NoError(t, model.DB.Save(owner).Error)
	var jpg bytes.Buffer
	require.NoError(t, jpeg.Encode(&jpg, image.NewRGBA(image.Rect(0, 0, 1, 1)), nil))
	// A JPEG may be followed by padding. The byte limit is inclusive.
	sized := make([]byte, 2<<20)
	copy(sized, jpg.Bytes())
	otherVersion := version(request(http.MethodPut, writePath+"?id="+strconv.Itoa(owner.Id), otherToken, sized))
	require.NotEmpty(t, otherVersion)
	assert.Equal(t, http.StatusNotFound, request(http.MethodGet, readPath+"?v="+otherVersion, "", nil).Code)
	assert.Empty(t, version(request(http.MethodDelete, writePath+"?id="+strconv.Itoa(owner.Id), otherToken, nil)))
	assert.Equal(t, http.StatusOK, request(http.MethodGet, lastURL, "", nil).Code)

	for _, invalid := range []struct {
		name    string
		data    []byte
		status  int
		message string
	}{
		{"empty", nil, http.StatusUnsupportedMediaType, "请选择有效的"},
		{"html with JPEG MIME", []byte("<html></html>"), http.StatusUnsupportedMediaType, "请选择有效的"},
		{"truncated GIF header", []byte("GIF89a"), http.StatusUnsupportedMediaType, "请选择有效的"},
		{"larger than 2 MB", append(sized, 0), http.StatusRequestEntityTooLarge, "不能超过 2 MB"},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			res := request(http.MethodPut, writePath, token, invalid.data)
			assert.Equal(t, invalid.status, res.Code)
			assert.Contains(t, res.Body.String(), invalid.message)
			assert.Equal(t, http.StatusOK, request(http.MethodGet, lastURL, "", nil).Code)
		})
	}
	assert.Empty(t, version(request(http.MethodDelete, writePath, token, nil)))
	assert.Empty(t, version(request(http.MethodGet, "/api/user/self", token, nil)))
	missing := request(http.MethodGet, lastURL, "", nil)
	assert.Equal(t, http.StatusNotFound, missing.Code)
	assert.Equal(t, "no-store", missing.Header().Get("Cache-Control"))
}
