package test

import (
	"APIGateway/internal/app"
	grade "APIGateway/internal/clients/grade_service"
	"APIGateway/internal/clients/lesson_service"
	schedule "APIGateway/internal/clients/schedule_service"
	user "APIGateway/internal/clients/user_service"
	"APIGateway/internal/config"
	"APIGateway/pkg/tools/jwt"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
)

var email = gofakeit.Email()
var pass = randomFakePassword()
var token = ""
var id = gofakeit.Uint16()

func NewServer() *httprouter.Router {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	app := app.New(context.Background(), slog.New(slog.NewTextHandler(os.Stdout, nil)), cfg)

	return app.Router
}

var r = NewServer()

func GetAdminToken() string {
	user := &user.LoginReq{
		Email:    "test@mail.ru",
		Password: "123123123",
	}
	data, _ := json.Marshal(user)
	buf := bytes.NewBuffer(data)
	req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/login", buf)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	adminToken := w.Result().Header.Get("Authorization")
	return adminToken
}

func TestAuthConvey(t *testing.T) {
	Convey("Auth endpoints should respond correctly", t, func() {
		Convey("Register Should response 201 status", func() {
			user := &user.RegisterReq{
				Email:    email,
				Password: pass,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/reg", buf)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusCreated)
		})
		Convey("Register Should response 409 status", func() {
			user := &user.RegisterReq{
				Email:    email,
				Password: pass,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/reg", buf)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusConflict)
		})
		Convey("Login Should response 200 status && token from body", func() {
			user := &user.LoginReq{
				Email:    email,
				Password: pass,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/login", buf)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Result().Header.Get("Authorization"), ShouldContainSubstring, "Bearer")
		})
		Convey("Login Should response 404 status", func() {
			user := &user.LoginReq{
				Email:    gofakeit.Email(),
				Password: randomFakePassword(),
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/login", buf)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Login Should response 401 status", func() {
			user := &user.LoginReq{
				Email:    email,
				Password: randomFakePassword(),
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/login", buf)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func TestAnotherMethods(t *testing.T) {
	SetToken()
	reqToken := strings.TrimSpace(strings.Split(token, "Bearer ")[1])

	Convey("Endpoints should respond correctly", t, func() {

		Convey("Update password with status 404", func() {
			user := &user.UpdatePasswordReq{
				Email:       gofakeit.Email(),
				OldPassword: randomFakePassword(),
				NewPassword: randomFakePassword(),
				Token:       reqToken,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/profile/change-password", buf)
			req.Header.Add("Authorization", token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Update password with status 401", func() {
			user := &user.UpdatePasswordReq{
				Email:       email,
				OldPassword: randomFakePassword(),
				NewPassword: randomFakePassword(),
				Token:       reqToken,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/profile/change-password", buf)
			req.Header.Add("Authorization", token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})
		Convey("Update password with status 200", func() {
			user := &user.UpdatePasswordReq{
				Email:       email,
				OldPassword: pass,
				NewPassword: randomFakePassword(),
				Token:       reqToken,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/profile/change-password", buf)
			req.Header.Add("Authorization", token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Set permission with status 403", func() {
			user := &user.SetPermissionLevelReq{
				UserID:          1,
				InitiatorID:     20,
				PermissionLevel: 3,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/admin/profiles/permissions", buf)
			req.Header.Add("Authorization", token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusForbidden)
		})
		Convey("Set permission with status 404", func() {
			user := &user.SetPermissionLevelReq{
				UserID:          2,
				InitiatorID:     512,
				PermissionLevel: 3,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/admin/profiles/permissions", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Set permission with status 200", func() {
			user := &user.SetPermissionLevelReq{
				UserID:          1,
				InitiatorID:     512,
				PermissionLevel: int64(gofakeit.Number(1, 9)),
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/admin/profiles/permissions", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Fill profile with status 200", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid

			user := &user.FillUserProfileReq{
				UserID:      id,
				Name:        gofakeit.Name(),
				Lastname:    gofakeit.LastName(),
				MiddleName:  gofakeit.Name(),
				DateOfBirth: "12-12-2012",
				Classname:   "2Г",
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PUT", "http://127.0.0.1:1234/profile/edit", buf)
			req.Header.Add("Authorization", token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Change user with status 200", func() {
			user := &user.ChangeUserStatusReq{
				UserID:      1,
				InitiatorID: 512,
				Active:      true,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/admin/profiles", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Change user with status 404", func() {
			user := &user.ChangeUserStatusReq{
				UserID:      5,
				InitiatorID: 512,
				Active:      true,
			}
			data, _ := json.Marshal(user)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/admin/profiles", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Get students by classname with status 200", func() {
			req, _ := http.NewRequest("GET", "http://127.0.0.1:1234/teacher/lessons/set-grades?classname=2Г", nil)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
			body := strings.TrimSpace(w.Body.String())
			So(body, ShouldContainSubstring, "[")
		})
		Convey("Get students by classname with status 404", func() {
			req, _ := http.NewRequest("GET", "http://127.0.0.1:1234/teacher/lessons/set-grades?classname=22А", nil)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Delete user with status 404", func() {
			req, _ := http.NewRequest("DELETE", "http://127.0.0.1:1234/admin/profiles?id=15", nil)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Delete user with status 200", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid - 2
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("http://127.0.0.1:1234/admin/profiles?id=%d", id), nil)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Create lesson with status 201", func() {
			lesson := &lesson_service.SetLessonReq{
				TeacherID:  int64(id),
				ClassName:  "11Б",
				LessonName: gofakeit.JobTitle(),
			}
			data, _ := json.Marshal(lesson)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/admin/lessons", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusCreated)
		})
		Convey("Create lesson with status 409", func() {
			lesson := &lesson_service.SetLessonReq{
				TeacherID:  1,
				ClassName:  "11Б",
				LessonName: "Физкультура",
			}
			data, _ := json.Marshal(lesson)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/admin/lessons", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusConflict)
		})
		Convey("Set homework with status 404", func() {
			schedule := &schedule.SetHomeworkReq{
				Date:       "27-06-2024",
				Classname:  "11Б",
				Homework:   "Собрать гербарий",
				LessonName: "Биология",
			}
			data, _ := json.Marshal(schedule)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/teacher/schedule/set-schedule/lessons", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Set homework with status 200", func() {
			schedule := &schedule.SetHomeworkReq{
				Date:       "27-06-2024",
				Classname:  "11Б",
				Homework:   "Отжимания 10 раз",
				LessonName: "Физкультура",
			}
			data, _ := json.Marshal(schedule)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/teacher/schedule/set-schedule/lessons", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Get week schedule with status 404", func() {
			req, _ := http.NewRequest("GET", "http://127.0.0.1:1234/schedule?classname=12Г", nil)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Get week schedule with status 200", func() {
			req, _ := http.NewRequest("GET", "http://127.0.0.1:1234/schedule?classname=11Г", nil)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Set grade with status 201", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid
			grade := &grade.SetGradeReq{
				StudentID:  id,
				DeviceID:   int(gofakeit.Int64()),
				Grade:      5,
				LessonName: "Биология",
			}
			data, _ := json.Marshal(grade)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/teacher/lessons/grades", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusCreated)
		})
		Convey("Set grade with status 409", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid
			grade := &grade.SetGradeReq{
				StudentID:  id,
				DeviceID:   int(gofakeit.Int64()),
				Grade:      5,
				LessonName: "Биология",
			}
			data, _ := json.Marshal(grade)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/teacher/lessons/grades", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusConflict)
		})
		Convey("Get grades with status 200", func() {
			req, _ := http.NewRequest("GET", "http://127.0.0.1:1234/lessons/grades/all", nil)
			req.Header.Add("Authorization", token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Get lesson grades with status 200", func() {
			req, _ := http.NewRequest("GET", "http://127.0.0.1:1234/lessons/grades?lesson=Биология", nil)
			req.Header.Add("Authorization", token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Change grade with status 404", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid
			grade := &grade.ChangeGradeReq{
				StudentID:  id,
				LessonName: gofakeit.JobTitle(),
				Grade:      4,
				Date:       time.Now().Format(time.DateOnly),
			}
			data, _ := json.Marshal(grade)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/teacher/lessons/grades", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Change grade with status 200", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid
			grade := &grade.ChangeGradeReq{
				StudentID:  id,
				LessonName: "Биология",
				Grade:      4,
				Date:       time.Now().Format(time.DateOnly),
			}
			data, _ := json.Marshal(grade)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("PATCH", "http://127.0.0.1:1234/teacher/lessons/grades", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Set term grade with status 201", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid
			grade := &grade.SetTermGradeReq{
				UserID:     id,
				LessonName: "Биология",
			}
			data, _ := json.Marshal(grade)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/teacher/lessons/term-grades", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusCreated)
		})
		Convey("Set term grade with status 406", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid
			grade := &grade.SetTermGradeReq{
				UserID:     id,
				LessonName: "Химия",
			}
			data, _ := json.Marshal(grade)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/teacher/lessons/term-grades", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusNotAcceptable)
		})
		Convey("Set term grade with status 409", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid
			grade := &grade.SetTermGradeReq{
				UserID:     id,
				LessonName: "Биология",
			}
			data, _ := json.Marshal(grade)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/teacher/lessons/term-grades", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusConflict)
		})
		Convey("Delete grade with status 404", func() {
			grade := &grade.DeleteGradeReq{
				StudentID:  666666,
				LessonName: "Биология",
				Date:       time.Now().Format(time.DateOnly),
			}
			data, _ := json.Marshal(grade)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("DELETE", "http://127.0.0.1:1234/teacher/lessons/grades", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("Delete grade with status 200", func() {
			claims, _ := jwt.ValidateToken(reqToken)
			id := claims.Uid
			grade := &grade.DeleteGradeReq{
				StudentID:  id,
				LessonName: "Биология",
				Date:       time.Now().Format(time.DateOnly),
			}
			data, _ := json.Marshal(grade)
			buf := bytes.NewBuffer(data)
			req, _ := http.NewRequest("DELETE", "http://127.0.0.1:1234/teacher/lessons/grades", buf)
			req.Header.Add("Authorization", GetAdminToken())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}
func SetToken() {
	user := &user.LoginReq{
		Email:    email,
		Password: pass,
	}
	data, _ := json.Marshal(user)
	buf := bytes.NewBuffer(data)
	req, _ := http.NewRequest("POST", "http://127.0.0.1:1234/login", buf)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	token = w.Result().Header.Get("Authorization")
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, 10)
}
