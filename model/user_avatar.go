package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type userAvatarColumn LongText

func (userAvatarColumn) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return LongText("").GormDBDataType(db, field)
}

func (userAvatarColumn) QueryClauses(*schema.Field) []clause.Interface {
	return []clause.Interface{omitUserAvatar{}}
}

type omitUserAvatar struct{ clause.Select }

// Normal User reads, including lists and login lookups, must not transfer a
// multi-megabyte image. Explicit projections such as GetUserAvatar opt in.
func (omitUserAvatar) ModifyStatement(stmt *gorm.Statement) {
	switch stmt.Dest.(type) {
	case *User, *[]User, *[]*User:
		stmt.Omits = append(stmt.Omits, "avatar")
	}
}

// SetUserAvatar updates the migration-only User.Avatar column without a read/modify/write of the user.
func SetUserAvatar(id int, dataURL string) error {
	return DB.Table("users").Where("id = ? AND deleted_at IS NULL", id).UpdateColumn("avatar", dataURL).Error
}

// GetUserAvatar limits the comparison to the data URL header, so a stale or unknown
// version never transfers the image from the database.
func GetUserAvatar(id int, version string) (string, error) {
	var result struct{ Avatar string }
	err := DB.Model(&User{}).Select("avatar").Where("id = ? AND SUBSTR(avatar, 1, 64) LIKE ?", id, "%;v="+version+";base64,%").Take(&result).Error
	return result.Avatar, err
}
