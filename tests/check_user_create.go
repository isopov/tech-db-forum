package tests

import (
	"github.com/bozaro/tech-db-forum/generated/client"
	"github.com/bozaro/tech-db-forum/generated/client/operations"
	"github.com/bozaro/tech-db-forum/generated/models"
	"github.com/go-openapi/strfmt"
	"reflect"
	"sort"
	"strings"
)

func init() {
	Register(Checker{
		Name:        "user_create_simple",
		Description: "",
		FnCheck:     CheckUserCreateSimple,
	})
	Register(Checker{
		Name:        "user_create_unicode",
		Description: "",
		FnCheck:     CheckUserCreateUnicode,
		Deps: []string{
			"user_get_one_simple",
		},
	})
	Register(Checker{
		Name:        "user_create_conflict",
		Description: "",
		FnCheck:     Modifications(CheckUserCreateConflict),
		Deps: []string{
			"user_create_simple",
		},
	})
}

func CreateUser(c *client.Forum, user *models.User) *models.User {
	if user == nil {
		user = RandomUser()
	}

	request := *user
	request.Nickname = ""

	_, err := c.Operations.UserCreate(operations.NewUserCreateParams().
		WithNickname(user.Nickname).
		WithProfile(&request).
		WithContext(Expected(201, user, nil)))
	CheckNil(err)

	return user
}

func CheckUser(c *client.Forum, user *models.User) {
	_, err := c.Operations.UserGetOne(operations.NewUserGetOneParams().
		WithNickname(user.Nickname).
		WithContext(Expected(200, user, nil)))
	CheckNil(err)
}

func CheckUserCreateSimple(c *client.Forum) {
	CreateUser(c, nil)
}

func CheckUserCreateUnicode(c *client.Forum) {
	user := RandomUser()
	user.Fullname = "Маркиз О-де-Колóн"
	user.About = "Бездельник третьего разряда 😋"
	CreateUser(c, user)
	CheckUser(c, user)
}

func CheckUserCreateConflict(c *client.Forum, m *Modify) {
	user1 := CreateUser(c, nil)
	user2 := CreateUser(c, nil)

	expected := []models.User{}
	conflict_user := RandomUser()

	// Email
	switch m.Int(4) {
	case 1:
		conflict_user.Email = user1.Email
		expected = append(expected, *user1)
	case 2:
		conflict_user.Email = strfmt.Email(strings.ToLower(user1.Email.String()))
		expected = append(expected, *user1)
	case 3:
		conflict_user.Email = strfmt.Email(strings.ToUpper(user1.Email.String()))
		expected = append(expected, *user1)
	}

	// Nickname
	switch m.Int(5) {
	case 1:
		conflict_user.Nickname = user2.Nickname
		expected = append(expected, *user2)
	case 2:
		conflict_user.Nickname = strings.ToLower(user2.Nickname)
		expected = append(expected, *user2)
	case 3:
		conflict_user.Nickname = strings.ToUpper(user2.Nickname)
		expected = append(expected, *user2)
	case 4:
		conflict_user.Nickname = user1.Nickname
		if len(expected) == 0 {
			expected = append(expected, *user1)
		}
	}

	// Check
	if len(expected) != 0 {
		nickname := conflict_user.Nickname
		conflict_user.Nickname = ""
		c.Operations.UserCreate(operations.NewUserCreateParams().
			WithNickname(nickname).
			WithProfile(conflict_user).
			WithContext(Expected(409, &expected, func(users interface{}) interface{} {
				result := UserByNickname(reflect.ValueOf(users).Elem().Interface().([]models.User))
				sort.Sort(result)
				return result
			})))
	}
}