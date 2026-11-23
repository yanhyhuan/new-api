package controller

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/MoeForge/moego/base64"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
)

const maxUserAvatarBytes = 2 << 20

func UploadUserAvatar(c *gin.Context) {
	data, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxUserAvatarBytes))
	if err != nil {
		var sizeError *http.MaxBytesError
		if errors.As(err, &sizeError) {
			common.ApiErrorStatus(c, http.StatusRequestEntityTooLarge, common.NewMessage("Avatar must be 2 MB or smaller"), gin.H{"message": "头像文件不能超过 2 MB"})
		} else {
			common.ApiErrorStatus(c, http.StatusBadRequest, common.NewMessage("Unable to read avatar file"), gin.H{"message": "无法读取头像文件"})
		}
		return
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || !slices.Contains([]string{"png", "jpeg", "webp", "gif"}, format) {
		common.ApiErrorStatus(c, http.StatusUnsupportedMediaType, common.NewMessage("Choose a valid PNG, JPG, WebP, or GIF image"), gin.H{"message": "请选择有效的 PNG、JPG、WebP 或 GIF 图片"})
		return
	}

	// A new, unguessable version grants native image requests access to this upload only.
	// Keeping it in a MIME parameter makes the image and its version one atomic column update.
	var nonce [16]byte
	rand.Read(nonce[:])
	version := hex.EncodeToString(nonce[:])
	dataURL := "data:image/" + format + ";v=" + version + ";base64," + base64.StdEncoding.EncodeToString(data)
	if err := model.SetUserAvatar(c.GetInt("id"), dataURL); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"avatar_version": version})
}

func DeleteUserAvatar(c *gin.Context) {
	if err := model.SetUserAvatar(c.GetInt("id"), ""); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"avatar_version": ""})
}

func GetUserAvatar(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	id, err := strconv.Atoi(c.Param("id"))
	version := c.Query("v")
	if err != nil || id <= 0 || len(version) != 32 {
		c.Status(http.StatusNotFound)
		return
	}
	if _, err := hex.DecodeString(version); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	dataURL, err := model.GetUserAvatar(id, version)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	header, encoded, ok := strings.Cut(dataURL, ";v="+version+";base64,")
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Cache-Control", "private, max-age=3600, immutable")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	c.Data(http.StatusOK, strings.TrimPrefix(header, "data:"), data)
}
